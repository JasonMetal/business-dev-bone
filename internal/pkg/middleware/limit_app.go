package middleware

import (
	"errors"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func LimitByAppId(maxEventsPerSec float64, maxBurstSize int) gin.HandlerFunc {
	al := NewAppLimiter(maxEventsPerSec, maxBurstSize)
	return func(c *gin.Context) {
		appID := c.GetHeader("mg-appid")
		if appID == "" {
			_ = c.Error(errors.New("mg-appid is required"))
			c.AbortWithStatusJSON(400, gin.H{"error": "mg-appid is required"})
			return
		}

		limiter := al.getLimiter(appID)
		if limiter.Allow() {
			c.Next()
			return
		}

		_ = c.Error(ErrLimitExceeded)
		c.AbortWithStatusJSON(429, gin.H{"error": "rate limit exceeded"})
	}
}

// AppLimiter manage limiter based on appId
type AppLimiter struct {
	limiters map[string]*rate.Limiter
	lastUsed map[string]time.Time // record last used time of limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

func NewAppLimiter(maxEventsPerSec float64, maxBurstSize int) *AppLimiter {
	al := &AppLimiter{
		limiters: make(map[string]*rate.Limiter),
		lastUsed: make(map[string]time.Time),
		rate:     rate.Limit(maxEventsPerSec),
		burst:    maxBurstSize,
	}

	// clean up limiter that not used for a long time
	go al.cleanupLoop(time.Minute, 1*time.Hour)

	return al
}

// getLimiter get or create a limiter for appId
func (al *AppLimiter) getLimiter(appId string) *rate.Limiter {
	al.mu.RLock()
	limiter, exists := al.limiters[appId]
	al.mu.RUnlock()

	if exists {
		// update last used time
		al.mu.Lock()
		al.lastUsed[appId] = time.Now()
		al.mu.Unlock()
		return limiter
	}

	al.mu.Lock()
	defer al.mu.Unlock()

	// check again, prevent concurrent creation
	if limiter, exists := al.limiters[appId]; exists {
		al.lastUsed[appId] = time.Now()
		return limiter
	}

	limiter = rate.NewLimiter(al.rate, al.burst)
	al.limiters[appId] = limiter
	al.lastUsed[appId] = time.Now()
	return limiter
}

// cleanupLoop clean up limiter that not used for a long time
func (al *AppLimiter) cleanupLoop(interval, ttl time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		al.mu.Lock()
		now := time.Now()
		for appId, lastUsed := range al.lastUsed {
			if now.Sub(lastUsed) > ttl {
				delete(al.limiters, appId)
				delete(al.lastUsed, appId)
			}
		}
		al.mu.Unlock()
	}
}
