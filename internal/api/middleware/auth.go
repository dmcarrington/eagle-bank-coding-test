package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/davidcarrington/eagle-bank/internal/auth"
)

const userIDKey = "userID"

func RequireAuth(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "authorization header is required"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "authorization header must be Bearer <token>"})
			return
		}

		claims, err := auth.ParseToken(secret, parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid or expired token"})
			return
		}

		c.Set(userIDKey, claims.UserID)
		c.Next()
	}
}

func CallerID(c *gin.Context) string {
	return c.GetString(userIDKey)
}
