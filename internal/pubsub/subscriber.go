package pubsub

import (
	"context"
	"fmt"

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

	if err != nil && ctx.Err() == nil {
		return fmt.Errorf("error receiving from subscription %s: %w", subscriptionID, err)
	}

	return nil
}

// SubscribeToAll starts receiving messages from multiple subscriptions
func (s *Subscriber) SubscribeToAll(ctx context.Context, subscriptionIDs, topicIDs []string, handler MessageHandler) {
	// Create subscription to topic mapping
	subTopicMap := make(map[string]string)
	for i, subID := range subscriptionIDs {
		if i < len(topicIDs) {
			subTopicMap[subID] = topicIDs[i]
		}
	}

	// Start a goroutine for each subscription
	for _, subID := range subscriptionIDs {
		topicID := subTopicMap[subID]
		go func(subscriptionID, topic string) {
			sub := s.client.client.Subscriber(subscriptionID)
			err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
				s.log.Info("Received message from %s: %s", subscriptionID, string(msg.Data))
				handler(ctx, msg)
				msg.Ack()
			})
			if err != nil && ctx.Err() == nil {
				s.log.Error("Error receiving from subscription %s: %v", subscriptionID, err)
			}
		}(subID, topicID)
	}
}
