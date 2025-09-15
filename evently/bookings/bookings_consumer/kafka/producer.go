package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(broker string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(broker),
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) Publish(topic string, key, value []byte) error {
	p.writer.Topic = topic

	log.Printf("Publishing message to Kafka: topic=%s, key=%s, value=%s\n",
		topic, string(key), string(value))

	err := p.writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   key,
			Value: value,
		},
	)

	if err != nil {
		log.Println("Kafka publish error:", err)
	} else {
		log.Println("Message published successfully")
	}

	return err
}

