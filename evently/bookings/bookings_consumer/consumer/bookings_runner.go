package consumer

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bookings_consumer/kafka" 
	"bookings_consumer/models" 
)

func StartBookingConsumer(
	broker string,
	topic string,
	groupID string,
	deps *models.ProcessorDeps,
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

	// Start consuming with handler
	reader.Start(ctx, func(key, value []byte) error {
		processBookingMessage(ctx, value, deps)
		return nil
	})
}
