package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter holds rate limiting data
type RateLimiter struct {
	visitors map[string]*Visitor
	mutex    sync.RWMutex
	rate     int
	window   time.Duration
}

// Visitor represents a visitor with request tracking
type Visitor struct {
	requests int
	window   time.Time
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     rate,
		window:   window,
	}

	// Clean up old visitors every minute
	go rl.cleanupVisitors()

	return rl
}

// cleanupVisitors removes old visitor entries
func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		rl.mutex.Lock()
		for ip, visitor := range rl.visitors {
			if time.Since(visitor.lastSeen) > rl.window*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mutex.Unlock()
	}
}

// isAllowed checks if the request is allowed
func (rl *RateLimiter) isAllowed(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	visitor, exists := rl.visitors[ip]

	if !exists {
		rl.visitors[ip] = &Visitor{
			requests: 1,
			window:   now,
			lastSeen: now,
		}
		return true
	}

	// Reset window if expired
	if now.Sub(visitor.window) > rl.window {
		visitor.requests = 1
		visitor.window = now
		visitor.lastSeen = now
		return true
	}

	// Check if within rate limit
	if visitor.requests < rl.rate {
		visitor.requests++
		visitor.lastSeen = now
		return true
	}

	visitor.lastSeen = now
	return false
}

// RateLimit middleware function
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !rl.isAllowed(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": rl.window.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CreateRateLimiter creates a rate limiting middleware
func CreateRateLimiter(requestsPerMinute int) gin.HandlerFunc {
	limiter := NewRateLimiter(requestsPerMinute, time.Minute)
	return limiter.RateLimit()
}
