package http

import (
	"net/http"
	"strings"

	"github.com/electrum/dynamic-pricing-engine/internal/infrastructure/auth"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware extracts and validates a Bearer JWT from the Authorization header.
// On success, it sets "username" and "role" in the gin context.
func AuthMiddleware(jwtSvc *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			Error(c, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			Error(c, http.StatusUnauthorized, "invalid authorization format")
			return
		}

		claims, err := jwtSvc.ValidateToken(parts[1])
		if err != nil {
			Error(c, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		username, _ := (*claims)["username"].(string)
		role, _ := (*claims)["role"].(string)

		c.Set("username", username)
		c.Set("role", role)
		c.Next()
	}
}

// AdminMiddleware ensures the authenticated user has the "admin" role.
// Must be placed after AuthMiddleware.
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			Error(c, http.StatusForbidden, "admin access required")
			return
		}
		c.Next()
	}
}
