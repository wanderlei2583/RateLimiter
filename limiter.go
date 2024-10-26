package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type RateLimiter struct {
	storage RateLimiterStorage
}

func NewRateLimiter(storage RateLimiterStorage) *RateLimiter {
	return &RateLimiter{storage: storage}
}

func (rl *RateLimiter) IsAllowed(
	identifier string,
	limitType string,
) (bool, error) {
	var limit, window int
	var err error
	switch limitType {
	case "ip":
		limit, err = strconv.Atoi(os.Getenv("IP_LIMIT"))
		if err != nil {
			return false, err
		}
		window, err = strconv.Atoi(os.Getenv("IP_WINDOW_SECONDS"))
	case "token":
		limit, err = strconv.Atoi(os.Getenv("TOKEN_LIMIT"))
		if err != nil {
			return false, err
		}
		window, err = strconv.Atoi(os.Getenv("TOKEN_WINDOW_SECONDS"))
	}
	if err != nil {
		return false, err
	}

	key := fmt.Sprintf("%s:%s", limitType, identifier)
	count, err := rl.storage.Increment(
		key,
		time.Duration(window)*time.Second,
	)
	if err != nil {
		return false, err
	}

	return count <= limit, nil
}
