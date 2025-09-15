package main

import (
	"context"
	"log"
	"os"
	"time"

	"events/auth"
	"events/controllers"
	"events/repository"
	"events/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	
	dbHost := mustGetEnv("DB_EVENTS_HOST")
	dbPort := mustGetEnv("DB_EVENTS_PORT")
	dbUser := mustGetEnv("DB_EVENTS_USER")
	dbPass := mustGetEnv("DB_EVENTS_PASSWORD")
	dbName := mustGetEnv("DB_EVENTS_NAME")

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

	db := client.Database(dbName)
	repo := repository.NewEventRepository(db)

	redisClient := newRedisClient(mustGetEnv("REDIS_HOST"), mustGetEnv("REDIS_PORT"), mustGetEnv("REDIS_PASSWORD"))
	redisSeats := newRedisClient(mustGetEnv("REDIS_SEATS_HOST"), mustGetEnv("REDIS_SEATS_PORT"), mustGetEnv("REDIS_SEATS_PASSWORD"))
	redisPrice := newRedisClient(
		mustGetEnv("REDIS_PRICE_HOST"),
		mustGetEnv("REDIS_PRICE_PORT"),
		mustGetEnv("REDIS_PRICE_PASSWORD"),

	)

	eventService := service.NewEventService(repo, redisClient, redisSeats, redisPrice)
	eventController := controllers.NewEventController(eventService)

	r := gin.Default()
	api := r.Group("/api/v1")
	{
		api.GET("/events/all", eventController.GetAllEvents)
		api.GET("/events/upcoming", eventController.GetAllUpcomingEvents)
		api.GET("/events/:id", eventController.GetEventByID)

		admin := api.Group("/events")
		admin.Use(auth.AdminOnly())
		{
			admin.POST("/create", eventController.CreateEvent)
			admin.PUT("/:id", eventController.UpdateEvent)
			admin.GET("/analytics/capacityUtil", eventController.GetCapacityUtilization)
			admin.GET("/analytics/mostBooked", eventController.GetMostBookedEvents)
			admin.GET("/analytics/mostPopular", eventController.GetMostPopularEvents)

		}
	}

	port := mustGetEnv("PORT")
log.Println("Events service running on port " + port)
if err := r.Run("0.0.0.0:" + port); err != nil {
    log.Fatal("Failed to start Events service:", err)
}

}

func newRedisClient(host, port, pass string) *redis.Client {
	addr := host + ":" + port
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       0,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	log.Println("Connected to Redis at", addr)
	return rdb
}

func mustGetEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		log.Fatalf("Environment variable %s is required but not set", key)
	}
	return value
}
