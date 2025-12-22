package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestValidator adds comprehensive request validation
func RequestValidator() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Check Content-Type for POST/PUT requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			contentType := c.GetHeader("Content-Type")
			if contentType != "application/json" && contentType != "" {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"error": "Content-Type must be application/json",
				})
				c.Abort()
				return
			}
		}

		// Add request validation context
		c.Set("validation_start_time", time.Now())

		c.Next()
	})
}

// ResponseValidator adds response validation and formatting
func ResponseValidator() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()

		// Add response headers for better API consistency
		c.Header("X-API-Version", "1.0")
		c.Header("X-Response-Time", time.Since(c.GetTime("validation_start_time")).String())
	})
}

// SecurityHeaders adds security headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")

		c.Next()
	}
}

// RequestSizeLimit limits request body size
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":    "Request body too large",
				"max_size": maxSize,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
