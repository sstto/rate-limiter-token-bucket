package middleware

import (
	"log"
	"net"
	"net/http"
	"project/bucket"
	"strings"
	"sync"
	"time"
)

type rateLimiter struct {
	bucketMap    map[string]*bucket.Bucket
	lock         sync.Mutex
	capacity     int
	refillPeriod time.Duration
	refillTokens int
}

func (r *rateLimiter) allow(key string) bool {
	b, exists := r.bucketMap[key]
	if exists {
		log.Printf("bucket tryConsume key: %v\n", key)
		return b.TryConsume()
	} else {
		r.lock.Lock()
		defer r.lock.Unlock()
		newBucket, _ := bucket.NewBuilder().
			SetCapacity(r.capacity).
			SetRefillPeriod(r.refillPeriod).
			SetRefillTokens(r.refillTokens).
			Build()
		r.bucketMap[key] = newBucket
		log.Printf("make new bucket key:%v\n", key)
		return newBucket.TryConsume()
	}
}

type RateLimiterBuilder struct {
	next         http.Handler
	capacity     int
	refillPeriod time.Duration
	refillTokens int
}

func NewRateLimiter(next http.Handler) *RateLimiterBuilder {
	return &RateLimiterBuilder{ // default value
		next:         next,
		capacity:     1000,
		refillPeriod: 10 * time.Second,
		refillTokens: 100,
	}
}

func (b *RateLimiterBuilder) SetCapacity(c int) *RateLimiterBuilder {
	b.capacity = c
	return b
}

func (b *RateLimiterBuilder) SetRefillPeriod(p time.Duration) *RateLimiterBuilder {
	b.refillPeriod = p
	return b
}

func (b *RateLimiterBuilder) SetRefillTokens(t int) *RateLimiterBuilder {
	b.refillTokens = t
	return b
}

func (b *RateLimiterBuilder) LimitByIp() http.Handler {
	rateLimiter := &rateLimiter{
		bucketMap:    make(map[string]*bucket.Bucket),
		capacity:     b.capacity,
		refillPeriod: b.refillPeriod,
		refillTokens: b.refillTokens,
	}
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if !rateLimiter.allow(getClientIp(r)) {
				http.Error(w, "IP Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			b.next.ServeHTTP(w, r)
		},
	)
}

func (b *RateLimiterBuilder) LimitByPath() http.Handler {
	rateLimiter := &rateLimiter{
		bucketMap:    make(map[string]*bucket.Bucket),
		capacity:     b.capacity,
		refillPeriod: b.refillPeriod,
		refillTokens: b.refillTokens,
	}
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if !rateLimiter.allow(r.URL.Path) {
				http.Error(w, "Path Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			b.next.ServeHTTP(w, r)
		},
	)
}

func (b *RateLimiterBuilder) LimitByRequest() http.Handler {
	rateLimiter := &rateLimiter{
		bucketMap:    make(map[string]*bucket.Bucket),
		capacity:     b.capacity,
		refillPeriod: b.refillPeriod,
		refillTokens: b.refillTokens,
	}
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if !rateLimiter.allow("key") {
				http.Error(w, "Server is Busy..", http.StatusTooManyRequests)
				return
			}
			b.next.ServeHTTP(w, r)
		},
	)
}

func getClientIp(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // IP 주소가 포트 없이 제공된 경우
	}

	if ip == "::1" { // IPv6 루프백 주소 필터링
		return "127.0.0.1"
	}

	return ip
}
