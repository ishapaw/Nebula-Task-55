package consumer

import (
	"cancel_consumer/models"
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func (p *CancelProcessor) ProcessCancelBookingMessage(ctx context.Context, key, value []byte) error {
	var msg models.KafkaCancelEvent
	if err := json.Unmarshal(value, &msg); err != nil {
		log.Printf("Invalid cancel message: %v", err)
		return err
	}

	if msg.BookingRequestId != "" {
		reqKey := "bookingRequest:" + msg.BookingRequestId

		state, err := p.redisReq.Get(ctx, reqKey).Result()
		if err == redis.Nil {

			err = p.redisReq.Set(ctx, reqKey, "cancelled", 0).Err()

			if err != nil {
				log.Printf("Failed to mark request %s cancelled: %v", msg.BookingRequestId, err)
				return err
			}

			log.Printf("Request %s not in Redis, cancelling at DB level", msg.BookingRequestId)
			return p.cancelAtDB(ctx, msg, key)

		} else if err != nil {

			log.Printf("Redis error while fetching state for %s: %v", msg.BookingRequestId, err)
			return err
		}

		switch state {
		case "state1", "state2", "state3":
			// still inflight, mark cancelled

			err = p.redisReq.Set(ctx, reqKey, "cancelled", 0).Err()

			if err != nil {
				log.Printf("Failed to mark request %s cancelled: %v", msg.BookingRequestId, err)
				return err
			}

			log.Printf("Marked request %s as cancelled", msg.BookingRequestId)

		case "success":
			// already success -> cancel at DB

			err = p.redisReq.Set(ctx, reqKey, "cancelled", 0).Err()

			if err != nil {
				log.Printf("Failed to mark request %s cancelled: %v", msg.BookingRequestId, err)
				return err
			}

			log.Printf("Request %s already success, deleting booking", msg.BookingRequestId)

			return p.cancelAtDB(ctx, msg, key)

		case "failed", "cancelled":
			log.Printf("Request %s already in terminal state: %s", msg.BookingRequestId, state)
		}

		return nil
	}

	if msg.BookingId != "" {

		log.Printf("Processing cancel by BookingID: %s", msg.BookingId)
		return p.cancelAtDB(ctx, msg, key)
	}

	log.Printf("Cancel message missing bookingRequestId and bookingId")
	return nil
}

func (p *CancelProcessor) publishSeatsUpdate(requestId string, msg models.KafkaCancelEvent) error {
	updateEvent := models.KafkaUpdateEvent{
		EventId:   msg.EventId,
		Seats:     msg.Seats,
		Operation: "add",
	}

	payload, err := json.Marshal(updateEvent)
	if err != nil {
		log.Printf("Failed to marshal seats update event: %v", err)
		return err
	}

	topic, _ := os.LookupEnv("UPDATE_SEATS_REQUESTS")

	err = p.producer.Publish(
		topic,
		[]byte(requestId),
		payload,
	)

	if err != nil {
		log.Printf("Failed to publish seats update event: %v", err)
	} else {
		log.Printf("Published seats update event %v", msg)
	}

	return err
}

func (p *CancelProcessor) cancelAtDB(ctx context.Context, msg models.KafkaCancelEvent, key []byte) error {
	seatsKey := "seatsLeft:" + msg.EventId

	if msg.BookingId != "" {
		if err := p.db.Model(&models.Booking{}).Where("id = ?", msg.BookingId).Update("status", "cancelled").Error; err != nil {
			log.Printf("Error marking booking %s as cancelled: %v", msg.BookingId, err)
			return err
		}
		log.Printf("Deleted booking %s from DB", msg.BookingId)

	} else if msg.BookingRequestId != "" {
		if err := p.db.Model(&models.Booking{}).Where("request_id = ?", msg.BookingRequestId).Update("status", "cancelled").Error; err != nil {
			log.Printf("Error marking booking with reqId %s as cancelled : %v", msg.BookingRequestId, err)
			return err
		}
		log.Printf("Deleted booking for requestID %s", msg.BookingRequestId)
	}

	if msg.Seats > 0 {
		p.publishSeatsUpdate(string(key), msg)

		if err := p.redisSeats.IncrBy(ctx, seatsKey, int64(msg.Seats)).Err(); err != nil {
			log.Printf("Error incrementing seats for requestID %s: %v", msg.BookingRequestId, err)
			return err
		}
		log.Printf("Restored %d seats for eventId %s", msg.Seats, msg.EventId)
	}

	return nil
}
