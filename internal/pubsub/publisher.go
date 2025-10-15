package pubsub

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"github.com/dipjyotimetia/pubsub-emulator/pkg/logger"
)

// Publisher handles message publishing
type Publisher struct {
	client *Client
	log    *logger.Logger
}

// NewPublisher creates a new publisher
func NewPublisher(client *Client, log *logger.Logger) *Publisher {
	return &Publisher{
		client: client,
		log:    log,
	}
}

// PublishMessage publishes a message to a specific topic
func (p *Publisher) PublishMessage(ctx context.Context, topicID, data string, attributes map[string]string) (string, error) {
	publisher := p.client.client.Publisher(topicID)
	defer publisher.Stop()

	msg := &pubsub.Message{
		Data:       []byte(data),
		Attributes: attributes,
	}

	result := publisher.Publish(ctx, msg)
	msgID, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to publish message to topic %s: %w", topicID, err)
	}

	p.log.Info("Published message to %s with ID: %s", topicID, msgID)
	return msgID, nil
}

// PublishToTopics publishes the same message to multiple topics
func (p *Publisher) PublishToTopics(ctx context.Context, topicIDs []string, data string) (map[string]string, error) {
	messageIDs := make(map[string]string)

	for _, topicID := range topicIDs {
		msgID, err := p.PublishMessage(ctx, topicID, data, nil)
		if err != nil {
			p.log.Error("Failed to publish to topic %s: %v", topicID, err)
			continue
		}
		messageIDs[topicID] = msgID
	}

	if len(messageIDs) == 0 {
		return nil, fmt.Errorf("failed to publish to any topics")
	}

	return messageIDs, nil
}

// CreateMessageInfo creates a MessageInfo from a pubsub.Message
func CreateMessageInfo(msg *pubsub.Message, topicID string) MessageInfo {
	return MessageInfo{
		ID:          msg.ID,
		Data:        string(msg.Data),
		Attributes:  msg.Attributes,
		PublishTime: msg.PublishTime,
		Topic:       topicID,
		Received:    time.Now(),
	}
}
