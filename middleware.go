package main

import (
	"net/http"
	"strings"
)

type RateLimitMiddleware struct {
	limiter *RateLimiter
}

func NewRateLimitMiddleware(limiter *RateLimiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{limiter: limiter}
}

func (m *RateLimitMiddleware) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var allowed bool
		var err error

		token := r.Header.Get("API_KEY")
		if token != "" {
			allowed, err = m.limiter.IsAllowed(token, "token")
		} else {
			ip := getClientIP(r)
			allowed, err = m.limiter.IsAllowed(ip, "ip")
		}

		if err != nil {
			http.Error(
				w,
				"Internal Server Error",
				http.StatusInternalServerError,
			)
			return
		}

		if !allowed {
			w.Header().Set("Retry-After", "10")
			http.Error(
				w,
				"you have reached the maximum number of requests or actions allowed within a certain time frame",
				http.StatusTooManyRequests,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	ip = r.Header.Get("X-Forwarded-For")
	if ip != "" {
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	ip = r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}

	return ip
}
