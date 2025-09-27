package http

import (
	"net"
	stdhttp "net/http"
	"strings"
	"sync"
	"time"
)

type bucket struct {
	tokens float64
	last   time.Time
}

type RateLimiter struct {
	mu     sync.Mutex
	m      map[string]*bucket
	rate   float64
	burst  float64
	lastGC time.Time
}

func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		m:      make(map[string]*bucket),
		rate:   rps,
		burst:  float64(burst),
		lastGC: time.Now(),
	}
}

func (rl *RateLimiter) Middleware() func(stdhttp.Handler) stdhttp.Handler {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			if !rl.allow(clientIP(r)) {
				writeError(w, r, stdhttp.StatusTooManyRequests, "rate_limited", "too many requests")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) allow(key string) bool {
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b := rl.m[key]
	if b == nil {
		b = &bucket{tokens: rl.burst, last: now}
		rl.m[key] = b
	}

	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * rl.rate
	if b.tokens > rl.burst {
		b.tokens = rl.burst
	}
	if b.tokens < 1 {
		rl.gc(now)
		return false
	}

	b.tokens -= 1
	b.last = now
	rl.gc(now)
	return true
}

func (rl *RateLimiter) gc(now time.Time) {
	if now.Sub(rl.lastGC) < time.Minute {
		return
	}
	for k, b := range rl.m {
		if now.Sub(b.last) > 3*time.Minute {
			delete(rl.m, k)
		}
	}
	rl.lastGC = now
}

func clientIP(r *stdhttp.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		return strings.TrimSpace(strings.Split(xf, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
