package consumer

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cancel_consumer/kafka"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type CancelProcessor struct {
	redisReq   *redis.Client
	redisSeats *redis.Client
	db         *gorm.DB
	producer   *kafka.Producer
}

func NewCancelProcessor(redisReq, redisSeats *redis.Client, db *gorm.DB, producer *kafka.Producer) *CancelProcessor {
	return &CancelProcessor{redisReq: redisReq, redisSeats: redisSeats, db: db, producer: producer}
}

func StartCancelConsumer(
	broker string,
	topic string,
	groupID string,
	redisReq *redis.Client,
	redisSeats *redis.Client,
	db *gorm.DB,
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

	log.Printf("Kafka consumer started: topic=%s, groupID=%s", topic, groupID)

	processor := NewCancelProcessor(redisReq, redisSeats, db, producer)

	reader.Start(ctx, func(key, value []byte) error {
		processor.ProcessCancelBookingMessage(ctx, key, value)
		return nil
	})
}
