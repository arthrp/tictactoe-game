package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	validKeys   = make(map[string]bool)
	ctx         = context.Background()
)

// InitRedis initializes the Redis client and loads API keys.
func InitRedis(addr string) error {
	redisClient = redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Check connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %v", err)
	}

	return LoadAPIKeys()
}

// LoadAPIKeys loads API keys from Redis into memory.
func LoadAPIKeys() error {
	keys, err := redisClient.SMembers(ctx, "api_keys").Result()
	if err != nil {
		return fmt.Errorf("failed to load api keys from redis: %v", err)
	}

	// Clear existing keys and reload
	newKeys := make(map[string]bool)
	for _, key := range keys {
		if isValidUUID(key) {
			newKeys[key] = true
		} else {
			fmt.Printf("Warning: Skipping invalid UUID key from Redis: %s\n", key)
		}
	}
	validKeys = newKeys
	fmt.Printf("Loaded %d API keys from Redis\n", len(validKeys))
	return nil
}

func isValidUUID(u string) bool {
	parsed, err := uuid.Parse(u)
	return err == nil && parsed.Version() == 4
}

// AuthMiddleware checks for a valid API key in the x-api-key header.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("x-api-key")

		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key required in x-api-key header"})
			return
		}

		if !isValidUUID(apiKey) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key format. Must be UUIDv4"})
			return
		}

		if !validKeys[apiKey] {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or unauthorized API key"})
			return
		}

		c.Set("apiKey", apiKey)
		c.Next()
	}
}
