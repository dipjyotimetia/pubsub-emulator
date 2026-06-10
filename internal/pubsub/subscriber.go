package pubsub

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub/v2"
	"github.com/dipjyotimetia/pubsub-emulator/pkg/logger"
)

// MessageHandler is a function that processes received messages
type MessageHandler func(ctx context.Context, msg *pubsub.Message)

// Subscriber handles message subscriptions
type Subscriber struct {
	client *Client
	log    *logger.Logger
}

// NewSubscriber creates a new subscriber
func NewSubscriber(client *Client, log *logger.Logger) *Subscriber {
	return &Subscriber{
		client: client,
		log:    log,
	}
}

// Subscribe starts receiving messages from a subscription
func (s *Subscriber) Subscribe(ctx context.Context, subscriptionID string, handler MessageHandler) error {
	sub := s.client.client.Subscriber(subscriptionID)

	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		s.log.Info("Received message from %s: %s", subscriptionID, string(msg.Data))
		handler(ctx, msg)
		msg.Ack()
	})

	if err != nil && !errors.Is(ctx.Err(), context.Canceled) {
		return fmt.Errorf("error receiving from subscription %s: %w", subscriptionID, err)
	}

	return nil
}

// SubscribeToAll starts receiving messages from multiple subscriptions, one
// goroutine per subscription. The returned WaitGroup completes once every
// receiver has stopped (after ctx is cancelled), letting callers drain
// in-flight messages during shutdown.
func (s *Subscriber) SubscribeToAll(ctx context.Context, subscriptionIDs, topicIDs []string, handler MessageHandler) *sync.WaitGroup {
	subTopicMap := make(map[string]string)
	for i, subID := range subscriptionIDs {
		if i < len(topicIDs) {
			subTopicMap[subID] = topicIDs[i]
		}
	}

	var wg sync.WaitGroup
	for _, subID := range subscriptionIDs {
		topicID := subTopicMap[subID]
		wg.Add(1)
		go func(subscriptionID, topic string) {
			defer wg.Done()
			sub := s.client.client.Subscriber(subscriptionID)
			err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
				s.log.Info("Received message from %s: %s", subscriptionID, string(msg.Data))
				handler(ctx, msg)
				msg.Ack()
			})
			if err != nil && !errors.Is(ctx.Err(), context.Canceled) {
				s.log.Error("Error receiving from subscription %s: %v", subscriptionID, err)
			}
		}(subID, topicID)
	}
	return &wg
}
