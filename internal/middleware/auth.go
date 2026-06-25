package middleware

import (
	"net/http"
	"strings"

	"llm_relay/internal/auth"

	"github.com/gin-gonic/gin"
)

const authSubjectKey = "auth_subject"

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or malformed authorization header"})
			return
		}

		claims, err := auth.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token", "detail": err.Error()})
			return
		}

		c.Set(authSubjectKey, claims.Sub)
		c.Next()
	}
}

func extractBearerToken(authHeader string) string {
	parts := strings.Fields(authHeader)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
