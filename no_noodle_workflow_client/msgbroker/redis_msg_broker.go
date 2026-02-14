package msgbroker

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisMessageBroker struct {
	client *redis.Client
}

// NewRedisMessageBroker creates a new Redis message broker instance
func NewRedisMessageBroker(addr string, password string, db int) (*RedisMessageBroker, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr, // e.g., "localhost:6379"
		// Username: "default",
		// Password: password,
		DB: db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RedisMessageBroker{
		client: client,
	}, nil
}

// Enqueue pushes a message onto a Redis list (queue)
func (rb *RedisMessageBroker) Enqueue(ctx context.Context, queue string, message []byte) error {
	return rb.client.RPush(ctx, queue, message).Err()
}

// Dequeue blocks until a message is available on the queue or context is cancelled
func (rb *RedisMessageBroker) Dequeue(ctx context.Context, queue string) ([]byte, error) {
	// 0 timeout means block indefinitely until a message arrives or ctx is cancelled
	res, err := rb.client.BLPop(ctx, 0, queue).Result()
	if err != nil {
		return nil, err
	}
	if len(res) < 2 {
		return nil, fmt.Errorf("unexpected BLPop response length: %d", len(res))
	}
	// res[0] is the queue name, res[1] is the payload
	return []byte(res[1]), nil
}

// Close closes the Redis connection
func (rb *RedisMessageBroker) Close() error {
	return rb.client.Close()
}
