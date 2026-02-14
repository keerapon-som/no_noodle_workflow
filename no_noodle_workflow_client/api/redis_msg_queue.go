package api

import (
	"context"
	"fmt"
	"log"
	"no-noodle-workflow-client/msgbroker"
)

// RedisMessageService now uses Redis lists as a simple message channal instead of pub/sub.
// Messages are persisted in the list until a consumer dequeues them.
type RedisMessageService struct {
	broker *msgbroker.RedisMessageBroker
}

// NewRedisMessageService creates a new channal-based messaging service
func NewRedisMessageService(broker *msgbroker.RedisMessageBroker) *RedisMessageService {
	return &RedisMessageService{
		broker: broker,
	}
}

// This blocks until the context is cancelled or an unrecoverable error occurs.
func (ps *RedisMessageService) SubscribeChannal(ctx context.Context, channal string) {

	fmt.Println("Consuming messages from channal:", channal)

	for {
		// Exit if context is cancelled
		if ctx.Err() != nil {
			return
		}

		payload, err := ps.broker.Dequeue(ctx, channal)
		if err != nil {
			// If the context was cancelled, just exit
			if ctx.Err() != nil {
				return
			}
			log.Println("error dequeuing from Redis channal:", err)
			continue
		}

		if len(payload) == 0 {
			continue
		}

		log.Printf("Channal %s queued message: %s", channal, string(payload))
		// Handle the message payload here
	}
}
