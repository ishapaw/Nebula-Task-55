package main

import (
	"bookings_consumer/consumer"
	"bookings_consumer/kafka"
	"bookings_consumer/models"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	kafkaBrokers := mustGetEnv("KAFKA_BROKERS")
	topic := mustGetEnv("TOPIC_BOOKINGS_REQUESTS")
	group := mustGetEnv("BOOKINGS_CONSUMER_GROUP")

	redisReq := newRedisClient(mustGetEnv("REDIS_REQUESTS_HOST"), mustGetEnv("REDIS_REQUESTS_PORT"), mustGetEnv("REDIS_REQUESTS_PASSWORD"))

	redisSeats := newRedisClient(mustGetEnv("REDIS_SEATS_HOST"), mustGetEnv("REDIS_SEATS_PORT"), mustGetEnv("REDIS_SEATS_PASSWORD"))

	redisPrice := newRedisClient(
		mustGetEnv("REDIS_PRICE_HOST"),
		mustGetEnv("REDIS_PRICE_PORT"),
		mustGetEnv("REDIS_PRICE_PASSWORD"),
	)

	dbHost := mustGetEnv("POSTGRES_BOOKINGS_HOST")
	dbPort := mustGetEnv("POSTGRES_BOOKINGS_PORT")
	dbUser := mustGetEnv("POSTGRES_BOOKINGS_USER")
	dbPass := mustGetEnv("POSTGRES_BOOKINGS_PASSWORD")
	dbName := mustGetEnv("POSTGRES_BOOKINGS_DB")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPass, dbName, dbPort)

	var db *gorm.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Println("Waiting for Bookings Postgres to be ready...")
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatal("Failed to connect to bookings database:", err)
	}

	log.Println("Starting Bookings Consumer...")

	producer := kafka.NewProducer(kafkaBrokers)

	deps := models.ProcessorDeps{
		RedisReq: redisReq,
		RedisPrice: redisPrice,
		RedisSeats: redisSeats,
		DB: db,
		Producer: producer,
	}


	consumer.StartBookingConsumer(kafkaBrokers, topic, group, &deps)
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

func mustGetEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		log.Fatalf("Environment variable %s is required but not set", key)
	}
	return value
}
