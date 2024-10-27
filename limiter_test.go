package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setupTest() {
	os.Setenv("IP_RATE_LIMIT", "5")
	os.Setenv("IP_WINDOW_SECONDS", "1")
	os.Setenv("TOKEN_RATE_LIMIT", "10")
	os.Setenv("TOKEN_WINDOW_SECONDS", "1")
}

func TestRateLimiterIP(t *testing.T) {
	setupTest()

	config := &RedisConfig{
		Host: "localhost",
		Port: "6379",
	}

	storage := NewRedisStorage(config)
	limiter := NewRateLimiter(storage)

	testIP := "192.168.1.1"

	ctx := storage.client.Context()
	storage.client.FlushDB(ctx)

	tests := []struct {
		name          string
		expectedAllow bool
	}{
		{"First Request", true},
		{"Second Request", true},
		{"Third Request", true},
		{"Fourth Request", true},
		{"Fifth Request", true},
		{"Sixth Request", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := limiter.IsAllowed(testIP, "ip")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedAllow, allowed)
		})
	}
}

func TestRateLimiterToken(t *testing.T) {
	setupTest()

	config := &RedisConfig{
		Host: "localhost",
		Port: "6379",
	}

	storage := NewRedisStorage(config)
	limiter := NewRateLimiter(storage)

	testToken := "test-token"

	ctx := storage.client.Context()
	storage.client.FlushDB(ctx)

	for i := 1; i <= 10; i++ {
		allowed, err := limiter.IsAllowed(testToken, "token")
		assert.NoError(t, err)
		assert.True(t, allowed, fmt.Sprintf("Request %d should be allowed", i))
	}

	allowed, err := limiter.IsAllowed(testToken, "token")
	assert.NoError(t, err)
	assert.False(t, allowed, "11th request should be blocked")
}

func TestMiddleware(t *testing.T) {
	setupTest()

	config := &RedisConfig{
		Host: "localhost",
		Port: "6379",
	}

	storage := NewRedisStorage(config)
	limiter := NewRateLimiter(storage)
	middleware := NewRateLimitMiddleware(limiter)

	ctx := storage.client.Context()
	storage.client.FlushDB(ctx)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name         string
		setupRequest func() *http.Request
		expectedCode int
	}{
		{
			name: "IP Request - Should Allow",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/", nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Token Request - Should Allow",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("API_KEY", "test-token")
				return req
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupRequest()
			rr := httptest.NewRecorder()

			middlewareHandler := middleware.Limit(handler)
			middlewareHandler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
		})
	}
}

func TestRateLimiterErrors(t *testing.T) {
	config := &RedisConfig{
		Host: "localhost",
		Port: "6379",
	}

	storage := NewRedisStorage(config)
	limiter := NewRateLimiter(storage)

	tests := []struct {
		name      string
		setupEnv  func()
		expectErr bool
	}{
		{
			name: "Invalid IP Rate Limit",
			setupEnv: func() {
				os.Setenv("IP_RATE_LIMIT", "invalid")
				os.Setenv("IP_WINDOW_SECONDS", "1")
			},
			expectErr: true,
		},
		{
			name: "Invalid Window Seconds",
			setupEnv: func() {
				os.Setenv("IP_RATE_LIMIT", "5")
				os.Setenv("IP_WINDOW_SECONDS", "invalid")
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()

			_, err := limiter.IsAllowed("test", "ip")

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIPRateLimitExceeded(t *testing.T) {
	setupTest()

	config := &RedisConfig{
		Host: "localhost",
		Port: "6379",
	}

	storage := NewRedisStorage(config)
	limiter := NewRateLimiter(storage)

	ctx := storage.client.Context()
	storage.client.FlushDB(ctx)

	testIP := "192.168.1.2"

	for i := 0; i < 5; i++ {
		allowed, err := limiter.IsAllowed(testIP, "ip")
		assert.NoError(t, err)
		assert.True(t, allowed)
	}

	allowed, err := limiter.IsAllowed(testIP, "ip")
	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestTokenRateLimitExceeded(t *testing.T) {
	setupTest()

	config := &RedisConfig{
		Host: "localhost",
		Port: "6379",
	}

	storage := NewRedisStorage(config)
	limiter := NewRateLimiter(storage)

	ctx := storage.client.Context()
	storage.client.FlushDB(ctx)

	testToken := "test-token-2"

	for i := 0; i < 10; i++ {
		allowed, err := limiter.IsAllowed(testToken, "token")
		assert.NoError(t, err)
		assert.True(t, allowed)
	}

	allowed, err := limiter.IsAllowed(testToken, "token")
	assert.NoError(t, err)
	assert.False(t, allowed)
}

func TestRateLimitReset(t *testing.T) {
	setupTest()

	config := &RedisConfig{
		Host: "localhost",
		Port: "6379",
	}

	storage := NewRedisStorage(config)
	limiter := NewRateLimiter(storage)

	ctx := storage.client.Context()
	storage.client.FlushDB(ctx)

	testIP := "192.168.1.3"

	for i := 0; i < 5; i++ {
		allowed, err := limiter.IsAllowed(testIP, "ip")
		assert.NoError(t, err)
		assert.True(t, allowed)
	}

	time.Sleep(2 * time.Second)

	allowed, err := limiter.IsAllowed(testIP, "ip")
	assert.NoError(t, err)
	assert.True(t, allowed)
}
