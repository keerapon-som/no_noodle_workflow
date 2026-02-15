package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"no-noodle-workflow-core/entitites"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisTaskSubscriberRegistry struct {
	client            *redis.Client
	defaultExpiration time.Duration
}

// NewRedisTaskSubscriberRegistry creates a new Redis message broker instance
func NewRedisTaskSubscriberRegistry(addr string, password string, db int, defaultExpiration time.Duration) (*RedisTaskSubscriberRegistry, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr, // e.g., "localhost:6379"
		// Username: "default",
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RedisTaskSubscriberRegistry{
		client:            client,
		defaultExpiration: defaultExpiration,
	}, nil
}

func (r *RedisTaskSubscriberRegistry) AddSubscriber(sessionKey string, channal, registerURL string, expiration time.Duration) error {

	ctx := context.Background()

	channelInfo := entitites.ChannelInfo{
		Channel:     channal,
		RegisterURL: registerURL,
	}

	channelInfoBytes, err := json.Marshal(channelInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal channel info: %v", err)
	}

	// Store the channel info as a Redis hash with key "subscriber:{sessionKey}"
	hashKey := fmt.Sprintf("subscriber:%s", sessionKey)
	_, err = r.client.HSet(ctx, hashKey, "data", channelInfoBytes).Result()
	if err != nil {
		return fmt.Errorf("failed to add subscriber: %v", err)
	}

	// Set expiration for the subscriber
	if expiration > 0 {
		_, err = r.client.Expire(ctx, hashKey, expiration).Result()
		if err != nil {
			return fmt.Errorf("failed to set expiration for subscriber: %v", err)
		}
	} else if r.defaultExpiration > 0 {
		_, err = r.client.Expire(ctx, hashKey, r.defaultExpiration).Result()
		if err != nil {
			return fmt.Errorf("failed to set default expiration for subscriber: %v", err)
		}
	}

	return nil
}

func (r *RedisTaskSubscriberRegistry) GetChannelInfoBySessionKey(sessionKey string) (entitites.ChannelInfo, error) {
	ctx := context.Background()

	// Retrieve the channel info from Redis
	hashKey := fmt.Sprintf("subscriber:%s", sessionKey)
	channelInfoBytes, err := r.client.HGet(ctx, hashKey, "data").Bytes()
	if err != nil {
		return entitites.ChannelInfo{}, fmt.Errorf("failed to get channel info: %v", err)
	}

	var channelInfo entitites.ChannelInfo
	if err := json.Unmarshal(channelInfoBytes, &channelInfo); err != nil {
		return entitites.ChannelInfo{}, fmt.Errorf("failed to unmarshal channel info: %v", err)
	}

	return channelInfo, nil
}

func (r *RedisTaskSubscriberRegistry) RemoveSubscriber(sessionKey string) error {
	ctx := context.Background()
	hashKey := fmt.Sprintf("subscriber:%s", sessionKey)
	_, err := r.client.Del(ctx, hashKey).Result()
	return err
}
