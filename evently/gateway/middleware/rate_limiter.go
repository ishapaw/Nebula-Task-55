package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimitMiddleware(redis *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var key string
		var limit int
		var ttl time.Duration

		if userID := c.GetHeader("X-User-Id"); userID != "" &&
			(c.Request.Method == http.MethodPost || c.Request.Method == http.MethodDelete) {
			key = "ratelimit:user:" + userID
			limit = 1
			ttl = 2 * time.Second
		} else {
			key = "ratelimit:ip:" + c.ClientIP()
			limit = 5
			ttl = 10 * time.Second
		}

		count, err := redis.Incr(c.Request.Context(), key).Result()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "redis error"})
			return
		}

		if count == 1 {
			redis.Expire(c.Request.Context(), key, ttl)
		}

		if int(count) > limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		c.Next()
	}
}
