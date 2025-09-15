package main

import (
	"bookings_view/auth"
	"bookings_view/controllers"
	"bookings_view/repository"
	"bookings_view/service"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
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
		log.Println("Waiting for Postgres to be ready...")
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	bookingRepo := repository.NewBookingsViewRepository(db)
	bookingService := service.NewBookingsViewService(bookingRepo)
	bookingController := controllers.NewBookingsViewController(bookingService)

	r := gin.Default()

	api := r.Group("/api/v1")
	{
		api.GET("/bookings/:id", bookingController.GetBookingByID)
		api.GET("/bookings/user/:user_id", bookingController.GetBookingsByUserID)
		api.GET("/bookings/event/:event_id", bookingController.GetBookingsByEventID)
		api.GET("/bookings/request/:request_id", bookingController.GetBookingByRequestID)

		admin := api.Group("/bookings/analytics")
		admin.Use(auth.AdminOnly())
		{
			admin.GET("/total-bookings", bookingController.GetTotalBookings)
			admin.GET("/dailyStats", bookingController.GetDailyBookingStats)
		}
	}

	port := mustGetEnv("PORT_BOOKINGS_VIEW")

	log.Println("Bookings View Service running on port " + port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func mustGetEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		log.Fatalf("Environment variable %s is required but not set", key)
	}
	return value
}
