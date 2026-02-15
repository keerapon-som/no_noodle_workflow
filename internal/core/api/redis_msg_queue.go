package api

import (
	"context"
	"fmt"
	"log"

	"time"

	"github.com/keerapon-som/no_noodle_workflow/internal/core/msgbroker"
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

	// Background requeue loop: periodically move expired messages back to main queue
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := ps.broker.RequeueExpired(ctx, channal); err != nil {
					log.Println("error requeueing expired messages:", err)
				}
			}
		}
	}()

	for {
		// Exit if context is cancelled
		if ctx.Err() != nil {
			fmt.Println("Context cancelled, stopping subscription to channal:", channal, "error:", ctx.Err())
			return
		}

		// Reserve a message with a visibility timeout
		payload, err := ps.broker.Dequeue(ctx, channal, 20*time.Second)
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

		// Handle the message payload here

		if err := handler(callbackURL, payload); err != nil {
			log.Println("error handling message from channal:", err)
			// Do not Ack; message will be re-delivered after visibility timeout
			continue
		}

		// On successful handling, acknowledge the message so it is not re-delivered
		if err := ps.broker.Ack(context.Background(), channal, payload); err != nil {
			log.Println("error acking message from channal:", err)
		}
	}
}
