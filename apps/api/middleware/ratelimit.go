package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, timestamps := range rl.requests {
			var active []time.Time
			for _, t := range timestamps {
				if now.Sub(t) < rl.window {
					active = append(active, t)
				}
			}
			if len(active) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = active
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	timestamps := rl.requests[key]

	var active []time.Time
	for _, t := range timestamps {
		if now.Sub(t) < rl.window {
			active = append(active, t)
		}
	}

	if len(active) >= rl.limit {
		rl.requests[key] = active
		return false
	}

	active = append(active, now)
	rl.requests[key] = active
	return true
}

var globalRateLimiter = NewRateLimiter(60, 1*time.Minute)

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()

		if userID, exists := c.Get("userID"); exists {
			key = userID.(string)
		}

		if !globalRateLimiter.allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "muitas requisições, tente novamente em breve"})
			c.Abort()
			return
		}
		c.Next()
	}
}
