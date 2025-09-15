package consumer

import (
	"bookings_consumer/models"
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"
	"os"

	"bookings_consumer/kafka"

	"gorm.io/gorm/clause"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var decrSeatsScript = redis.NewScript(`
local available = redis.call("GET", KEYS[1])
if not available then
    return -1
end
available = tonumber(available)
local required = tonumber(ARGV[1])
if available >= required then
    redis.call("DECRBY", KEYS[1], required)
    return 1
else
    return 0
end
`)



var casStateScript = redis.NewScript(`
local current = redis.call("GET", KEYS[1])
if current == "cancelled" then
    return "cancelled"
else
    redis.call("SET", KEYS[1], ARGV[1], "EX", ARGV[2])
    return current
end
`)

const stateTTL = 5 * time.Minute


func compareAndSetState(ctx context.Context, rdb *redis.Client, key string, nextState string) (string, error) {
    res, err := casStateScript.Run(ctx, rdb, []string{key}, nextState, int(stateTTL)).Result()
    if err != nil {
        return "", err
    }
    return res.(string), nil
}


func processBookingMessage(ctx context.Context, value []byte, deps *models.ProcessorDeps){
	var req models.KafkaEvent
	if err := json.Unmarshal(value, &req); err != nil {
		log.Printf("Invalid booking message: %v", err)
		return
	}

	reqKey := "bookingRequest:" + req.RequestID
	state, _ := deps.RedisReq.Get(ctx, reqKey).Result()

	if state == "" {
		state = "state1"
		deps.RedisReq.Set(ctx, reqKey, state, stateTTL)
	} 

	req.State = state

	switch req.State {

		case "state1":
			stateHandlerFunc1(ctx, req, deps)

		case "state2":
			stateHandlerFunc2(ctx, req, deps)

		case "state3":
			stateHandlerFunc3(ctx,req, deps)

		case "failed":
			log.Printf("Request %s already failed", req.RequestID)

		case "success":
			log.Printf("Request %s already succeeded", req.RequestID)

		case "cancelled":
			log.Printf("Request %s is already cancelled", req.RequestID)
	}
}

func stateHandlerFunc1(ctx context.Context, req models.KafkaEvent, deps *models.ProcessorDeps) {
	reqKey := "bookingRequest:" + req.RequestID
	seatsKey := "seatsLeft:" + req.EventID

	if isCancelled(ctx, deps.RedisReq, reqKey) {
		insertBooking(deps.DB, req, deps.RedisPrice, "cancelled")
		log.Printf("Request %s was cancelled before seat allocation", req.RequestID)
		return
	}


	result, err := decrSeatsScript.Run(ctx, deps.RedisSeats, []string{seatsKey}, req.Seats).Int()
	if err != nil {
		log.Printf("Redis error: %v", err)
		return
	}

	if result <= 0 {
		insertBooking(deps.DB, req, deps.RedisPrice, "failed")
		if result == -1 {
			log.Printf("Request %s failed: event not found", req.RequestID)
		} else {
			log.Printf("Request %s failed: not enough seats", req.RequestID)
		}
		deps.RedisReq.Set(ctx, reqKey, "failed",stateTTL)
		return
	}

	prev, err := compareAndSetState(ctx, deps.RedisReq, reqKey, "state2")
	if err != nil {
		log.Printf("CAS error: %v", err)
		return
	}
	if prev == "cancelled" {
		insertBooking(deps.DB, req, deps.RedisPrice, "cancelled")
		deps.RedisSeats.IncrBy(ctx, seatsKey, int64(req.Seats))
		log.Printf("Request %s cancelled before moving to state2", req.RequestID)
		return
	}

	req.State = "state2"
	stateHandlerFunc2(ctx, req, deps)
}


func stateHandlerFunc2(ctx context.Context, req models.KafkaEvent, deps *models.ProcessorDeps) {
	reqKey := "bookingRequest:" + req.RequestID
	seatsKey := "seatsLeft:" + req.EventID

	if isCancelled(ctx, deps.RedisReq, reqKey) {
		deps.RedisSeats.IncrBy(ctx, seatsKey, int64(req.Seats))
		deps.RedisReq.Set(ctx, reqKey, "cancelled", stateTTL)
		log.Printf("Request %s cancelled during processing, seats reverted", req.RequestID)
		return
	}

	if insertBooking(deps.DB, req, deps.RedisPrice, "confirmed") {
		prev, err := compareAndSetState(ctx, deps.RedisReq, reqKey, "state3")
		if err != nil {
			log.Printf("CAS error: %v", err)
			return
		}
		if prev == "cancelled" {
			if err := deps.DB.Model(&models.Booking{}).
				Where("request_id = ?", req.RequestID).
				Update("status", "cancelled").Error; err != nil {
				log.Printf("DB error cancelling booking %s: %v", req.RequestID, err)
			}
			
			deps.RedisSeats.IncrBy(ctx, seatsKey, int64(req.Seats))
			log.Printf("Request %s cancelled before moving to state3", req.RequestID)
			return
		}

		req.State = "state3"
		stateHandlerFunc3(ctx, req, deps)
	}

}

func stateHandlerFunc3(ctx context.Context, req models.KafkaEvent, deps *models.ProcessorDeps) {
	reqKey := "bookingRequest:" + req.RequestID
	seatsKey := "seatsLeft:" + req.EventID

	if isCancelled(ctx, deps.RedisReq, reqKey) {
		if err := deps.DB.Model(&models.Booking{}).
			Where("request_id = ?", req.RequestID).
			Update("status", "cancelled").Error; err != nil {
			log.Printf("DB error cancelling booking %s: %v", req.RequestID, err)
			return
		}
		deps.RedisSeats.IncrBy(ctx, seatsKey, int64(req.Seats))
		deps.RedisReq.Set(ctx, reqKey, "cancelled", stateTTL)
		log.Printf("Request %s cancelled after DB insert, seats reverted", req.RequestID)
		return
	}

	err := publishSeatsUpdate(deps.Producer, req)
	if err != nil {
		log.Printf("Kafka error: %v", err)
		return
	}

	req.State = "success"
	deps.RedisReq.Set(ctx, reqKey, "success", stateTTL)
	log.Printf("Request %s processed successfully", req.RequestID)
}


func publishSeatsUpdate(producer *kafka.Producer, req models.KafkaEvent) error {
	event := models.KafkaUpdateEvent{
		EventId:   req.EventID,
		Seats:     req.Seats,
		Operation: "subtract",
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	topic, _ := os.LookupEnv("UPDATE_SEATS_REQUESTS")
	return producer.Publish(topic, []byte(req.RequestID), payload)
}


func isCancelled(ctx context.Context, rdb *redis.Client, key string) bool {
	state, _ := rdb.Get(ctx, key).Result()
	return state == "cancelled"
}

func insertBooking(db *gorm.DB, req models.KafkaEvent, redisPrice *redis.Client, status string) bool {
	priceKey := "price:" + req.EventID
	priceStr, _ := redisPrice.Get(context.Background(), priceKey).Result()
	price, _ := strconv.ParseFloat(priceStr, 64)

	err := db.Transaction(func(tx *gorm.DB) error {
		booking := models.Booking{
			RequestID: req.RequestID,
			EventID:   req.EventID,
			UserID:    req.UserID,
			Price:     price * float64(req.Seats),
			Seats:     req.Seats,
			Status:    status,
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&booking).Error
	})

	if err != nil {
		log.Printf("DB error: %v", err)
		return false
	}
	return true
}