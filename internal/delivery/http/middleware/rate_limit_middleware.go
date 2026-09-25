package middleware

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

/*
clientWindow tracks how many requests one client has made in its
current window, and when that window ends.
*/
type clientWindow struct {
	count   int
	resetAt time.Time
}

/*
rateLimiter is a fixed-window, in-memory limiter keyed by client IP.
State lives in this process only, so the limit applies per instance;
running several instances behind a load balancer would need a shared
store such as Redis instead.
*/
type rateLimiter struct {
	mu        sync.Mutex
	limit     int
	window    time.Duration
	clients   map[string]*clientWindow
	lastSweep time.Time
}

/*
allow records one request for key and reports whether it is within
the limit. When it is not, it also returns how long until the client's
window resets. Expired windows are swept at most once per window so
the map cannot grow without bound.
*/
func (rl *rateLimiter) allow(key string, now time.Time) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if now.Sub(rl.lastSweep) >= rl.window {
		for k, w := range rl.clients {
			if !now.Before(w.resetAt) {
				delete(rl.clients, k)
			}
		}
		rl.lastSweep = now
	}

	w, ok := rl.clients[key]
	if !ok || !now.Before(w.resetAt) {
		rl.clients[key] = &clientWindow{count: 1, resetAt: now.Add(rl.window)}
		return true, 0
	}

	if w.count >= rl.limit {
		return false, w.resetAt.Sub(now)
	}

	w.count++
	return true, 0
}

/*
RateLimit returns a Gin middleware that allows at most limit requests
per client IP within each window. Requests over the limit are aborted
with 429 Too Many Requests and a Retry-After header. Each call builds
its own limiter, so routes that use separate RateLimit calls are
counted independently.

The client IP comes from c.ClientIP(), which only trusts
X-Forwarded-For from proxies configured via SetTrustedProxies —
otherwise a client could dodge the limit by spoofing that header.
*/
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	return rateLimit(limit, window, func(c *gin.Context) string {
		return c.ClientIP()
	})
}

/*
RateLimitPerUser works like RateLimit but counts requests per
authenticated user instead of per IP, so users sharing one IP (an
office NAT, mobile carrier) do not use up each other's quota. It must
run after AuthMiddleware, which sets the user ID it keys on.
*/
func RateLimitPerUser(limit int, window time.Duration) gin.HandlerFunc {
	return rateLimit(limit, window, func(c *gin.Context) string {
		return c.GetString(UserIDKey)
	})
}

/*
rateLimit builds the limiter middleware shared by RateLimit and
RateLimitPerUser, counting requests under whatever key keyOf returns.
*/
func rateLimit(limit int, window time.Duration, keyOf func(c *gin.Context) string) gin.HandlerFunc {
	rl := &rateLimiter{
		limit:   limit,
		window:  window,
		clients: make(map[string]*clientWindow),
	}

	return func(c *gin.Context) {
		allowed, retryAfter := rl.allow(keyOf(c), time.Now())
		if !allowed {
			c.Header("Retry-After", strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests, please try again later"})
			return
		}

		c.Next()
	}
}
