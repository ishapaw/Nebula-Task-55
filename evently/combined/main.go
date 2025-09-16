package main

import (
	"log"
	"os"
	"time"

	"combined/repository"
	"combined/controllers"
	"combined/service"
	"combined/auth"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dbHost := mustGetEnv("DB_HOST")
	dbPort := mustGetEnv("DB_PORT")
	dbUser := mustGetEnv("DB_USER")
	dbPass := mustGetEnv("DB_PASSWORD")
	dbName := mustGetEnv("DB_NAME")

	dsn := "host=" + dbHost + " user=" + dbUser + " password=" + dbPass + " dbname=" + dbName + " port=" + dbPort + " sslmode=require"

	var db *gorm.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Println("Waiting for Postgres to be ready...")
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userController := controllers.NewUserController(userService)

	r := gin.Default()

	r.POST("/api/users/register", userController.Register)
	r.POST("/api/users/login", userController.Login)

	bookingRepo := repository.NewBookingsViewRepository(db)
	bookingService := service.NewBookingsViewService(bookingRepo)
	bookingController := controllers.NewBookingsViewController(bookingService)

	api := r.Group("/api/v1")
	{
		api.GET("/bookings/:id", bookingController.GetBookingByID)
		api.GET("/bookings/user/:user_id", bookingController.GetBookingsByUserID)
		api.GET("/bookings/request/:request_id", bookingController.GetBookingByRequestID)

		admin := api.Group("/bookings")
		admin.Use(auth.AdminOnly())
		{
			admin.GET("/all", bookingController.GetAllBookings)
			admin.GET("/event/:event_id", bookingController.GetBookingsByEventID)
			admin.GET("/analytics/total-bookings", bookingController.GetTotalBookings)
			admin.GET("/analytics/dailyStats", bookingController.GetDailyBookingStats)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local dev
	}
	log.Println("Combined service running on port " + port)
	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatal("Failed to start combined service:", err)
	}

}

func mustGetEnv(key string) string {
	value, ok := lookupEnv(key)
	if !ok || value == "" {
		log.Fatalf("Environment variable %s is required but not set", key)
	}
	return value
}

func lookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}
