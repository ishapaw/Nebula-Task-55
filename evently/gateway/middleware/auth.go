package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "supersecret"
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			log.Println("JWT Claims:", claims)

			c.Request.Header.Set("X-User-Role", claims["role"].(string))
			if userIDVal, ok := claims["user_id"]; ok {
				switch v := userIDVal.(type) {
				case float64:
					c.Request.Header.Set("X-User-Id", fmt.Sprintf("%.0f", v))
				case int:
					c.Request.Header.Set("X-User-Id", fmt.Sprintf("%d", v))
				case string:
					c.Request.Header.Set("X-User-Id", v)
				default:
					log.Println("user_id claim has unexpected type")
					c.Request.Header.Set("X-User-Id", "")
				}
			}
		}

		c.Next()
	}
}
