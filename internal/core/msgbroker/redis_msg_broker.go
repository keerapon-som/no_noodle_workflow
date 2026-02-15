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

// Dequeue reserves a message with a visibility timeout.
//
// It moves a message from the main queue to a processing queue and records a
// deadline in a sorted set. Until the message is Ack'ed or the deadline
// passes and it is re-queued, no other consumer will see it.
func (rb *RedisMessageBroker) Dequeue(ctx context.Context, queue string, visibilityTimeout time.Duration) ([]byte, error) {
	processingQueue := queue + ":processing"
	reservedSet := queue + ":reserved"

	// 0 timeout means block indefinitely until a message arrives or ctx is cancelled
	msg, err := rb.client.BRPopLPush(ctx, queue, processingQueue, 0).Result()
	if err != nil {
		return nil, err
	}

	// Record visibility timeout deadline
	deadline := time.Now().Add(visibilityTimeout).Unix()
	err = rb.client.ZAdd(ctx, reservedSet, redis.Z{
		Score:  float64(deadline),
		Member: msg,
	}).Err()
	if err != nil {
		return nil, err
	}

	return []byte(msg), nil
}

// Ack confirms successful processing of a message, removing it from the
// processing queue and the reserved set so it will not be re-delivered.
func (rb *RedisMessageBroker) Ack(ctx context.Context, queue string, message []byte) error {
	processingQueue := queue + ":processing"
	reservedSet := queue + ":reserved"
	msg := string(message)

	pipe := rb.client.TxPipeline()
	pipe.LRem(ctx, processingQueue, 1, msg)
	pipe.ZRem(ctx, reservedSet, msg)
	_, err := pipe.Exec(ctx)
	return err
}

// RequeueExpired scans for messages whose visibility timeout has expired and
// moves them back to the main queue so they can be processed again.
func (rb *RedisMessageBroker) RequeueExpired(ctx context.Context, queue string) error {
	processingQueue := queue + ":processing"
	reservedSet := queue + ":reserved"
	now := time.Now().Unix()

	expired, err := rb.client.ZRangeByScore(ctx, reservedSet, &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprint(now),
	}).Result()
	if err != nil {
		return err
	}
	if len(expired) == 0 {
		return nil
	}

	pipe := rb.client.TxPipeline()
	for _, msg := range expired {
		pipe.LRem(ctx, processingQueue, 1, msg)
		pipe.RPush(ctx, queue, msg)
		pipe.ZRem(ctx, reservedSet, msg)
	}
	_, err = pipe.Exec(ctx)
	return err
}

// Close closes the Redis connection
func (rb *RedisMessageBroker) Close() error {
	return rb.client.Close()
}
