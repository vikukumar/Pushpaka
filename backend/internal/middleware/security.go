package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// SecureHeaders sets production-grade security headers
func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Header("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' 'unsafe-eval' blob: https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; "+
				"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://fonts.googleapis.com; "+
				"font-src 'self' data: https://fonts.gstatic.com; "+
				"img-src 'self' data: blob: https://cdn.jsdelivr.net; "+
				"connect-src 'self' ws: wss: https://cdn.jsdelivr.net")
		c.Next()
	}
}

// RecoveryMiddleware handles panics and returns 500
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()
		c.Next()
	}
}

// ─── Rate Limiter ─────────────────────────────────────────────────────────────
// Simple sliding-window rate limiter keyed by client IP.
// Uses an in-memory token bucket with no external dependencies.

type rateBucket struct {
	tokens    int
	lastReset time.Time
}

type rateLimiter struct {
	mu        sync.Mutex
	buckets   map[string]*rateBucket
	maxReqs   int           // requests per window
	window    time.Duration // window duration
	stopClean chan struct{}
}

func newRateLimiter(maxReqs int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		buckets:   make(map[string]*rateBucket),
		maxReqs:   maxReqs,
		window:    window,
		stopClean: make(chan struct{}),
	}
	// Periodically evict stale buckets to avoid unbounded memory growth.
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				rl.mu.Lock()
				for k, b := range rl.buckets {
					if time.Since(b.lastReset) > rl.window*2 {
						delete(rl.buckets, k)
					}
				}
				rl.mu.Unlock()
			case <-rl.stopClean:
				return
			}
		}
	}()
	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[ip]
	if !ok || time.Since(b.lastReset) >= rl.window {
		rl.buckets[ip] = &rateBucket{tokens: rl.maxReqs - 1, lastReset: time.Now()}
		return true
	}
	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

// authLimiter is shared across all auth-route middleware instances.
var authLimiter = newRateLimiter(20, time.Minute)

// apiLimiter applies a generous limit to API routes.
var apiLimiter = newRateLimiter(300, time.Minute)

// RateLimit returns a Gin middleware that enforces the request limit.
// use "auth" for authentication endpoints, "api" for general API endpoints.
func RateLimit(kind string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		var allowed bool
		switch kind {
		case "auth":
			allowed = authLimiter.allow(ip)
		default:
			allowed = apiLimiter.allow(ip)
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests — please slow down",
			})
			return
		}
		c.Next()
	}
}
