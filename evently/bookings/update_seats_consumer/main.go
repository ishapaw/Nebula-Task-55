package main

import (
	"context"
	"log"
	"os"
	"time"

	"update_seats_consumer/consumer"
	"update_seats_consumer/kafka"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	kafkaBrokers := mustGetEnv("KAFKA_BROKERS")
	topic := mustGetEnv("UPDATE_SEATS_REQUESTS")
	group := mustGetEnv("UPDATE_SEATS_CONSUMER_GROUP")

	redisUpdatedSeats := newRedisClient(mustGetEnv("REDIS_UPDATED_SEATS_HOST"), mustGetEnv("REDIS_UPDATED_SEATS_PORT"), mustGetEnv("REDIS_UPDATED_SEATS_PASSWORD"))

	dbHost := mustGetEnv("DB_EVENTS_HOST")
	dbPort := mustGetEnv("DB_EVENTS_PORT")
	dbUser := mustGetEnv("DB_EVENTS_USER")
	dbPass := mustGetEnv("DB_EVENTS_PASSWORD")
	// dbName := mustGetEnv("DB_EVENTS_NAME")

	uri := "mongodb://" + dbUser + ":" + dbPass + "@" + dbHost + ":" + dbPort

	var client *mongo.Client
	var err error

	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
		cancel()

		if err == nil {
			ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
			err = client.Ping(ctx, nil)
			cancel()
			if err == nil {
				log.Println("Connected to MongoDB")
				break
			}
		}

		log.Println("Waiting for MongoDB to be ready...")
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	producer := kafka.NewProducer(kafkaBrokers)


	log.Println("Starting UpdateSeats Consumer...")
	consumer.StartUpdateSeatsConsumer(kafkaBrokers, topic, group, redisUpdatedSeats, client,producer)

}

func mustGetEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		log.Fatalf("Environment variable %s is required but not set", key)
	}
	return value
}

func newRedisClient(host, port, password string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0,
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", host+":"+port, err)
	}
	log.Println("Connected to Redis at", host+":"+port)
	return rdb
}