package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetHeader("X-User-Role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied, only admins can perform this action"})
			c.Abort()
			return
		}
		c.Next()
	}
}
