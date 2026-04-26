package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"franchise-saas-backend/internal/cache"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	requests map[string]*clientRequests
	mu       sync.Mutex
}

type clientRequests struct {
	count     int
	resetTime time.Time
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*clientRequests),
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, req := range rl.requests {
			if now.After(req.resetTime) {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(key string, limit int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	req, exists := rl.requests[key]

	if !exists || now.After(req.resetTime) {
		rl.requests[key] = &clientRequests{
			count:     1,
			resetTime: now.Add(window),
		}
		return true
	}

	if req.count >= limit {
		return false
	}

	req.count++
	return true
}

var globalRateLimiter = NewRateLimiter()

func RateLimit(requests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" || c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		key := c.ClientIP()

		if !globalRateLimiter.Allow(key, requests, window) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RateLimitByUser(requests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" || c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			c.Next()
			return
		}

		key := "ratelimit:" + userID.(string)
		ctx := context.Background()

		count, err := cache.SetWithIncrement(ctx, key, window, requests)
		if err != nil {
			c.Next()
			return
		}

		if count > requests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}