package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"kirimaja-go/internal/common/response"
)

// RateLimit returns a per-client-IP token-bucket limiter. The Midtrans
// webhook is unauthenticated (protected only by signature); without a limit
// a flood of requests — valid replays or forged probes — has no ceiling.
//
// r is the sustained requests/sec allowed per IP, burst the bucket size.
func RateLimit(r rate.Limit, burst int) gin.HandlerFunc {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Evict idle clients so the map can't grow unbounded.
	go func() {
		for range time.Tick(3 * time.Minute) {
			mu.Lock()
			for ip, c := range clients {
				if time.Since(c.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		cl, ok := clients[ip]
		if !ok {
			cl = &client{limiter: rate.NewLimiter(r, burst)}
			clients[ip] = cl
		}
		cl.lastSeen = time.Now()
		allow := cl.limiter.Allow()
		mu.Unlock()

		if !allow {
			response.Error(c, http.StatusTooManyRequests, "Too many requests", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}
