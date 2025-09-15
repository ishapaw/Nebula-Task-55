package consumer

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"update_seats_consumer/kafka"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/redis/go-redis/v9"
)


func StartUpdateSeatsConsumer(
	broker string,
	topic string,
	groupID string,
	redis *redis.Client,
	mongoClient *mongo.Client,
	producer *kafka.Producer,
) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan
		log.Println("Shutdown signal received")
		cancel()
	}()

	reader := kafka.NewReader(broker, topic, groupID)
	defer reader.Close()

	log.Printf("UpdateSeats Kafka consumer started: topic=%s, groupID=%s", topic, groupID)

	reader.Start(ctx, func(key, value []byte) error {
		return processUpdateSeatsMessage(ctx, key, value, redis, mongoClient, producer)
	})
}
