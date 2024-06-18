package redis

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	v9 "github.com/redis/go-redis/v9"
)

var (
	once     sync.Once
	instance *v9.Client
	ctx      = context.Background()
)

// GetRedisClient returns the singleton instance of the Redis client
func GetRedisClient() *v9.Client {
	once.Do(func() {
		redisURL := os.Getenv("REDIS_URL")
		if redisURL == "" {
			log.Fatal("REDIS_URL environment variable not set")
		}
		log.Println("Redis URL: ", redisURL)

		instance = v9.NewClient(&v9.Options{
			Addr:     redisURL,
			Password: "", // Use the appropriate password if set
			DB:       0,  // Use default DB
		})

		// Test the connection
		_, err := instance.Ping(ctx).Result()
		if err != nil {
			log.Fatalf("Could not connect to Redis: %v", err)
		}
	})
	return instance
}

// SaveMessage saves a message to Redis
func SaveMessage(key string, message string) error {
	rdb := GetRedisClient()
	err := rdb.Set(ctx, key, message, 0).Err()
	if err != nil {
		return fmt.Errorf("could not save message to Redis: %w", err)
	}
	return nil
}

// GetMessage retrieves a message from Redis
func GetMessage(key string) (string, error) {
	rdb := GetRedisClient()
	message, err := rdb.Get(ctx, key).Result()
	if err == v9.Nil {
		return "", fmt.Errorf("message not found")
	} else if err != nil {
		return "", fmt.Errorf("could not get message from Redis: %w", err)
	}
	return message, nil
}
