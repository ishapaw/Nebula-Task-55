package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"gateway/middleware"
	"gateway/kafka"
	"gateway/proxy"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var producer *kafka.Producer

func InitProducer(broker string) {
	producer = kafka.NewProducer(broker)
	log.Printf("Kafka producer initialized for broker %s\n", broker)
}

func HandleBookingRequest(c *gin.Context) {
	log.Println("HandleBookingRequest called")

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Println("Invalid JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	log.Println("Request body:", body)

	if _, ok := body["request_id"]; !ok {

		body["request_id"] = uuid.New().String()
		log.Println("Generated new request_id:", body["request_id"])
	
	}
		

	userID := c.GetHeader("X-User-Id")
	if userID == "" {
		log.Println("Missing X-User-Id header")
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing X-User-Id header"})
		return
	}

	body["user_id"] = userID
	log.Println("User ID from header:", userID)

	newBody, err := json.Marshal(body)
	if err != nil {
		log.Println("Failed to marshal body:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process request"})
		return
	}

	topic := selectTopic(c.Request.Method)
	log.Printf("Publishing to topic: %s, key: %s\n", topic, body["request_id"])

	if err := producer.Publish(topic, []byte(body["request_id"].(string)), newBody); err != nil {
		log.Println("Failed to publish to Kafka:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish"})
		return
	}

	log.Println("request successfully queued")
	c.JSON(http.StatusAccepted, gin.H{"status": "request queued", "request_id" : body["request_id"]})
}

func RegisterRoutes(r *gin.Engine, prod *kafka.Producer, redis *redis.Client) {
	producer = prod
	log.Println("Registering routes")

	api := r.Group("/api")

	api.Any("/users/*path", proxy.ReverseProxy("http://users-service:8081"))

	protected := api.Group("/v1")
	protected.Use(middleware.AuthMiddleware())

	protected.Use(middleware.RateLimitMiddleware(redis))
	{
		
		protected.Any("/events/*path", proxy.ReverseProxy("http://events-service:8082"))
		protected.Any("/bookings/*path", func(c *gin.Context) {
			method := c.Request.Method
			log.Println("Received /bookings request, method:", method)

			if method == http.MethodGet {
				proxy.ReverseProxy("http://bookings-view-service:8084")(c)
			} else if method == http.MethodPost || method == http.MethodDelete {
				HandleBookingRequest(c)
			} else {
				log.Println("Method not allowed:", method)
				c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed for /bookings"})
			}
		})
	}
}

func selectTopic(method string) string {
	switch method {
	case http.MethodPost:
		return "bookings.requests"
	case http.MethodDelete:
		return "cancel.requests"
	default:
		return "bookings.requests"
	}
}
