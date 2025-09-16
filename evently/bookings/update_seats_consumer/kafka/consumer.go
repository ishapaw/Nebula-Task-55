package kafka

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type Reader struct {
	reader *kafka.Reader
}

// NewReader creates a Kafka reader compatible with Confluent Cloud
func NewReader(broker, topic, groupID string) *Reader {
	key, ok := os.LookupEnv("KAFKA_KEY")
	if !ok || key == "" {
		log.Fatal("KAFKA_KEY env variable not set")
	}
	secret, ok := os.LookupEnv("KAFKA_SECRET")
	if !ok || secret == "" {
		log.Fatal("KAFKA_SECRET env variable not set")
	}

	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		TLS:           &tls.Config{},
		SASLMechanism: plain.Mechanism{Username: key, Password: secret},
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker},
		GroupID:     groupID,
		Topic:       topic,
		Dialer:      dialer,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset, 
	})
	

	return &Reader{reader: reader}
}

// Start begins consuming messages and calls the handler for each one
func (r *Reader) Start(ctx context.Context, handler func(key, value []byte) error) {
	for {
		m, err := r.reader.ReadMessage(ctx)
		if err != nil {
			// Exit on context cancellation
			if ctx.Err() != nil {
				log.Println("Kafka reader stopped:", ctx.Err())
				return
			}

			// Log and retry on temporary errors
			log.Println("Kafka read error:", err)
			time.Sleep(time.Second)
			continue
		}

		// Log received message (remove or adjust for high throughput)
		log.Printf("Received message: topic=%s partition=%d offset=%d key=%s value=%s\n",
			m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))

		// Call user handler
		if handler != nil {
			if err := handler(m.Key, m.Value); err != nil {
				log.Println("Handler error:", err)
			}
		}
	}
}

// Close closes the Kafka reader
func (r *Reader) Close() error {
	return r.reader.Close()
}
