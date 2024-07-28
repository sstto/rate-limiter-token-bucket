package middleware

import "net/http"

type RateLimiter interface {
	Key(r *http.Request) string
	Take(key string) bool
}

type RateLimitMiddleware struct {
	RateLimiter
}

func NewRateLimitMiddleware(rateLimiter RateLimiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{rateLimiter}
}

func (rm *RateLimitMiddleware) ServeNext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rm.Take(rm.Key(r)) {
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Rate limit exceeded"))
		}
	})
}
