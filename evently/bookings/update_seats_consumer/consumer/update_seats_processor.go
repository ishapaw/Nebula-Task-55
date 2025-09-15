package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"update_seats_consumer/kafka"

	"github.com/redis/go-redis/v9"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type KafkaUpdateEvent struct {
	EventId   string `json:"event_id"`
	Seats     int64  `json:"seats"`
	Operation string `json:"operation"`
}

func processUpdateSeatsMessage(ctx context.Context, key []byte, value []byte, redis *redis.Client, mongoClient *mongo.Client, producer *kafka.Producer) error {
	var msg KafkaUpdateEvent

	if err := json.Unmarshal(value, &msg); err != nil {
		log.Printf("Failed to parse update seats message: %v", err)
		return err
	}

	updatedSeatsKey := "updatedSeats:" + string(key)

	exists, err := redis.Exists(ctx, updatedSeatsKey).Result()
	if err != nil {
		log.Printf("Redis error: %v", err)
		return err
	}

	if exists > 0 {
		log.Printf("Skipping request %s: already processed", string(key))
		return nil
	}

	oid, err := primitive.ObjectIDFromHex(msg.EventId)

	collection := mongoClient.Database("eventsdb").Collection("events")
	filter := bson.M{"_id": oid}

	var update bson.M

	if msg.Operation == "add" {
		update = bson.M{"$inc": bson.M{"available_seats": msg.Seats}}
	} else if msg.Operation == "subtract" {
		update = bson.M{"$inc": bson.M{"available_seats": -msg.Seats}}
	} else {
		return fmt.Errorf("Invalid operation type: %s for requestId %s", msg.Operation, string(key))
	}

	res, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("Failed to update MongoDB: %v", err)
		return err
	}

	if res.MatchedCount == 0 {
		log.Printf("No event found with ID %s", msg.EventId)
		return nil
	}

	if err := redis.Set(ctx, updatedSeatsKey, "processed", 5*time.Minute).Err(); err != nil {
		log.Printf("Failed to mark request in Redis: %v", err)
		return err
	}

	log.Printf("Updated seats for event %v", msg)

	return nil
}
