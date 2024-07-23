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
	bucketMap map[string]*bucket.Bucket
	lock      sync.Mutex
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
			SetCapacity(10).
			SetRefillPeriod(60 * time.Second).
			SetRefillTokens(10).
			Build()
		r.bucketMap[key] = newBucket
		log.Printf("make new bucket key:%v\n", key)
		return newBucket.TryConsume()
	}
}

func LimitByIp(next http.Handler) http.Handler {
	rateLimiter := &rateLimiter{
		bucketMap: make(map[string]*bucket.Bucket),
	}
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if !rateLimiter.allow(getClientIp(r)) {
				http.Error(w, "IP Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		},
	)
}

func LimitByPath(next http.Handler) http.Handler {
	rateLimiter := &rateLimiter{
		bucketMap: make(map[string]*bucket.Bucket),
	}
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if !rateLimiter.allow(r.URL.Path) {
				http.Error(w, "Path Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		},
	)
}

func LimitByRequest(next http.Handler) http.Handler {
	rateLimiter := &rateLimiter{
		bucketMap: make(map[string]*bucket.Bucket),
	}
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if !rateLimiter.allow("key") {
				http.Error(w, "Server is Busy..", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
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
