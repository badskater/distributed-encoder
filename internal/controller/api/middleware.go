package api

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

// corsMiddleware sets CORS headers based on the provided allowed origins and
// handles preflight OPTIONS requests.
func corsMiddleware(origins []string, next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		allowed[o] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if _, ok := allowed[origin]; ok {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// rateLimitMiddleware enforces per-IP token-bucket rate limiting.
// Limit: 200 req/s, burst: 400.
func rateLimitMiddleware(next http.Handler) http.Handler {
	var clients sync.Map // map[string]*rate.Limiter

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}

		val, _ := clients.LoadOrStore(ip, rate.NewLimiter(200, 400))
		limiter := val.(*rate.Limiter)

		if !limiter.Allow() {
			w.Header().Set("Content-Type", "application/problem+json")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type":   "https://distencoder.dev/errors/rate-limit",
				"title":  "Too Many Requests",
				"status": 429,
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// responseBuffer captures the status code and body written by downstream
// handlers so that etagMiddleware can compute a hash before flushing.
type responseBuffer struct {
	http.ResponseWriter
	status int
	body   []byte
}

func (rb *responseBuffer) WriteHeader(code int) {
	rb.status = code
}

func (rb *responseBuffer) Write(b []byte) (int, error) {
	rb.body = append(rb.body, b...)
	return len(b), nil
}

// etagMiddleware computes a SHA-256 ETag for GET /api/v1/* responses that
// return 200 OK and returns 304 Not Modified when the client already has the
// current version.
func etagMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, "/api/v1/") {
			next.ServeHTTP(w, r)
			return
		}

		buf := &responseBuffer{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(buf, r)

		if buf.status != http.StatusOK {
			w.WriteHeader(buf.status)
			_, _ = w.Write(buf.body)
			return
		}

		hash := sha256.Sum256(buf.body)
		etag := fmt.Sprintf(`"%x"`, hash)

		if r.Header.Get("If-None-Match") == etag {
			w.Header().Set("ETag", etag)
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("ETag", etag)
		w.WriteHeader(buf.status)
		_, _ = w.Write(buf.body)
	})
}
