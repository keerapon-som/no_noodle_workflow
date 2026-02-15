package api

import (
	"context"
	"fmt"
	"log"
	"no-noodle-workflow-core/msgbroker"
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

// SendToMsgChannal enqueues a message into a Redis list (channal) for the given topic.
// The channal key is prefix + topic (e.g. "workflow:task0").
func (ps *RedisMessageService) SendToMsgChannal(ctx context.Context, channal string, payload []byte) error {

	return ps.broker.Enqueue(ctx, channal, payload)
}

// SubscribeChannal continuously dequeues messages from the topic channal and processes them.
// This blocks until the context is cancelled or an unrecoverable error occurs.
func (ps *RedisMessageService) SubscribeChannal(ctx context.Context, callbackURL string, channal string, handler func(callbackURL string, payload []byte) error) {

	fmt.Println("Consuming messages from channal:", channal)

	for {
		// Exit if context is cancelled
		if ctx.Err() != nil {
			fmt.Println("Context cancelled, stopping subscription to channal:", channal, "error:", ctx.Err())
			return
		}

		payload, err := ps.broker.Dequeue(ctx, channal)
		if err != nil {
			// If the context was cancelled, just exit
			if ctx.Err() != nil {
				fmt.Println("Context cancelled, stopping subscription to channal:", channal, "error:", ctx.Err())
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

		err = handler(callbackURL, payload)
		if err != nil {
			log.Println("error handling message from channal:", err)
		}
	}
}
