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
		Brokers:  []string{broker},
		GroupID:  groupID,
		Topic:    topic,
		Dialer:   dialer,
		MinBytes: 1,
		MaxBytes: 10e6,
		// remove CommitInterval to avoid auto commit
	})

	return &Reader{reader: reader}
}

// Start begins consuming messages and calls the handler for each one
// Commits offsets ONLY after handler succeeds
func (r *Reader) Start(ctx context.Context, handler func(key, value []byte) error) {
	for {
		m, err := r.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("Kafka reader stopped:", ctx.Err())
				return
			}
			log.Println("Kafka read error:", err)
			time.Sleep(time.Second)
			continue
		}

		log.Printf("Received message: topic=%s partition=%d offset=%d key=%s value=%s\n",
			m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))

		if handler != nil {
			if err := handler(m.Key, m.Value); err != nil {
				log.Println("Handler error, offset not committed:", err)
				// optionally: retry here
				continue
			}

			// Commit offset AFTER successful processing
			if err := r.reader.CommitMessages(ctx, m); err != nil {
				log.Println("Failed to commit offset:", err)
			} else {
				log.Printf("Offset committed: partition=%d offset=%d\n", m.Partition, m.Offset)
			}
		}
	}
}

// Close closes the Kafka reader
func (r *Reader) Close() error {
	return r.reader.Close()
}