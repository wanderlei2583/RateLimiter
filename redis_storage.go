package main

import (
	"context"
	"time"

	redis "github.com/go-redis/redis/v8"
)

type RedisConfig struct {
	Host string
	Port string
}

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(config *RedisConfig) *RedisStorage {
	client := redis.NewClient(&redis.Options{
		Addr: config.Host + ":" + config.Port,
	})

	return &RedisStorage{client: client}
}

func (rs *RedisStorage) Increment(
	key string,
	expiry time.Duration,
) (int, error) {
	ctx := context.Background()

	pipe := rs.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiry)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return int(incr.Val()), nil
}
