package redis

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// VIOLATION: state outside State. Owner: t0-plredis; moves to PackageDBMng in phase 2.
var rdb *redis.Client

// VIOLATION: init chaining. Init 2 of the package loader tools: the redis
// client for localhost:6379, called from Go package init.
func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

// Add stores value under key.
func Add(key string, value interface{}) error {
	if err := rdb.Set(ctx, key, value, 0).Err(); err != nil {
		return fmt.Errorf("redis: set %q: %w", key, err)
	}
	return nil
}

// Get returns the value stored under key. A missing key is an error.
func Get(key string) (interface{}, error) {
	ret, err := rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("redis: no value for key %q", key)
	}
	if err != nil {
		return nil, fmt.Errorf("redis: get %q: %w", key, err)
	}
	return ret, nil
}

// Ping reports whether the redis server on localhost:6379 answers.
func Ping() error {
	return rdb.Ping(ctx).Err()
}
