package kafka

import (
	"context"
	"crypto/tls"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(broker string) *Producer {
	key, _ := os.LookupEnv("KAFKA_KEY")
	secret, _ := os.LookupEnv("KAFKA_SECRET")

	mechanism := plain.Mechanism{
		Username: key,
		Password: secret,
	}

	return &Producer{
		writer: &kafka.Writer{
			Addr: kafka.TCP(broker),
			Transport: &kafka.Transport{
				SASL: mechanism,
				TLS:  &tls.Config{},
			},
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) Publish(topic string, key, value []byte) error {
	p.writer.Topic = topic

	log.Printf("Publishing message to Kafka: topic=%s, key=%s, value=%s\n", topic, string(key), string(value))

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
