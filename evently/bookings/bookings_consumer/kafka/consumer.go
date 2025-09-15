package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)


type Reader struct {
	reader *kafka.Reader
}

func NewReader(broker, topic, groupID string) *Reader {
	return &Reader{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{broker},
			GroupID: groupID,
			Topic:   topic,
		}),
	}
}

func (r *Reader) Start(ctx context.Context, handler func(key, value []byte) error) {
	for {
		m, err := r.reader.ReadMessage(ctx)
		if err != nil {
			log.Println("Kafka read error:", err)
			return
		}

		log.Printf("Received message from Kafka: topic=%s, partition=%d, offset=%d, key=%s, value=%s\n",
			m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))

		if handler != nil {
			if err := handler(m.Key, m.Value); err != nil {
				log.Println("Handler error:", err)
			}
		}
	}
}

func (r *Reader) Close() error {
	return r.reader.Close()
}
