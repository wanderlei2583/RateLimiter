package main

import "time"

type RateLimitStorage interface {
	Increment(
		key string,
		expiry time.Duration,
	) (int, error)
}
