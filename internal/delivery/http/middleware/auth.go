package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/sotalk/internal/delivery/http/response"
	"github.com/yourusername/sotalk/pkg/logger"
	"github.com/yourusername/sotalk/pkg/middleware"
	"go.uber.org/zap"
)

// AuthMiddleware creates a middleware for JWT authentication
func AuthMiddleware(jwtManager *middleware.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Missing authorization header",
				zap.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusUnauthorized, response.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authorization header is required",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("Invalid authorization header format",
				zap.String("header", authHeader),
			)
			c.JSON(http.StatusUnauthorized, response.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authorization header must be in format: Bearer <token>",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		userID, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			logger.Warn("Invalid or expired token",
				zap.Error(err),
				zap.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusUnauthorized, response.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid or expired token",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set("user_id", userID)

		// Continue to next handler
		c.Next()
	}
}

// OptionalAuthMiddleware is similar to AuthMiddleware but doesn't abort if token is missing
func OptionalAuthMiddleware(jwtManager *middleware.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token := parts[1]
		userID, err := jwtManager.ValidateAccessToken(token)
		if err == nil {
			c.Set("user_id", userID)
		}

		c.Next()
	}
}

// WebSocketAuthMiddleware handles authentication for WebSocket connections
// Unlike regular auth, it accepts tokens from query parameters (needed for WebSocket handshake)
func WebSocketAuthMiddleware(jwtManager *middleware.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// Try to get token from query parameter first (WebSocket standard)
		token = c.Query("token")

		// Fallback to Authorization header if query param not present
		if token == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					token = parts[1]
				}
			}
		}

		// No token found
		if token == "" {
			logger.Warn("Missing authentication token for WebSocket",
				zap.String("path", c.Request.URL.Path),
				zap.String("remote_addr", c.Request.RemoteAddr),
			)
			c.JSON(http.StatusUnauthorized, response.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authentication token is required (pass as ?token=... query parameter)",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Validate token
		userID, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			logger.Warn("Invalid or expired WebSocket token",
				zap.Error(err),
				zap.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusUnauthorized, response.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid or expired token",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		logger.Info("WebSocket authentication successful",
			zap.String("user_id", userID),
			zap.String("remote_addr", c.Request.RemoteAddr),
		)

		// Set user ID in context
		c.Set("user_id", userID)

		// Continue to next handler
		c.Next()
	}
}
