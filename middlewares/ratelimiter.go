package middlewares

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type client struct {
	Requests int
	Expiry   time.Time
}

var clients = make(map[string]*client)
var mu sync.Mutex

func RateLimitMiddleware(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				http.Error(w, "Unable to determine IP", http.StatusInternalServerError)
				return
			}
			mu.Lock()
			c, exists := clients[ip]
			now := time.Now()
			if !exists || now.After(c.Expiry) {
				clients[ip] = &client{Requests: 1, Expiry: now.Add(window)}
			} else {
				if c.Requests >= maxRequests {
					mu.Unlock()
					http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
					return
				}
				c.Requests++
			}
			mu.Unlock()
			next.ServeHTTP(w, r)
		})
	}
}