package main

import (
	"log"
	"os"
	"time"
	"net"
	"context"
	"users/controllers"
	"users/repository"
	"users/service"

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

	port := mustGetEnv("PORT_USERS")
	log.Println("Users service running on port " + port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start users service:", err)
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
