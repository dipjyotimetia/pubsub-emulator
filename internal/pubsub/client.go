package pubsub

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"github.com/dipjyotimetia/pubsub-emulator/pkg/logger"
)

// Client wraps the Google Cloud Pub/Sub client with additional functionality
type Client struct {
	client    *pubsub.Client
	projectID string
	log       *logger.Logger
}

// NewClient creates a new Pub/Sub client wrapper
func NewClient(ctx context.Context, projectID string, log *logger.Logger) (*Client, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}

	return &Client{
		client:    client,
		projectID: projectID,
		log:       log,
	}, nil
}

// Close closes the underlying Pub/Sub client
func (c *Client) Close() error {
	return c.client.Close()
}

// GetClient returns the underlying pubsub.Client
func (c *Client) GetClient() *pubsub.Client {
	return c.client
}

// ProjectID returns the project ID
func (c *Client) ProjectID() string {
	return c.projectID
}

// CreateTopic creates a new Pub/Sub topic
func (c *Client) CreateTopic(ctx context.Context, topicID string) (*pubsubpb.Topic, error) {
	topic, err := c.client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{
		Name: fmt.Sprintf("projects/%s/topics/%s", c.projectID, topicID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create topic %s: %w", topicID, err)
	}
	c.log.Info("Created topic: %s", topic.Name)
	return topic, nil
}

// CreateSubscription creates a new Pub/Sub subscription
func (c *Client) CreateSubscription(ctx context.Context, subscriptionID, topicID string, ackDeadlineSeconds int32) (*pubsubpb.Subscription, error) {
	sub, err := c.client.SubscriptionAdminClient.CreateSubscription(ctx, &pubsubpb.Subscription{
		Name:               fmt.Sprintf("projects/%s/subscriptions/%s", c.projectID, subscriptionID),
		Topic:              fmt.Sprintf("projects/%s/topics/%s", c.projectID, topicID),
		AckDeadlineSeconds: ackDeadlineSeconds,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription %s: %w", subscriptionID, err)
	}
	c.log.Info("Created subscription: %s", sub.Name)
	return sub, nil
}

// CreateTopicsAndSubscriptions creates multiple topics and their corresponding subscriptions
func (c *Client) CreateTopicsAndSubscriptions(ctx context.Context, topicIDs, subscriptionIDs []string) error {
	if len(topicIDs) != len(subscriptionIDs) {
		return fmt.Errorf("number of topics and subscriptions must match")
	}

	for i := range topicIDs {
		topic, err := c.CreateTopic(ctx, topicIDs[i])
		if err != nil {
			c.log.Warn("Failed to create topic %s: %v", topicIDs[i], err)
			continue
		}

		_, err = c.CreateSubscription(ctx, subscriptionIDs[i], topicIDs[i], 20)
		if err != nil {
			c.log.Warn("Failed to create subscription %s: %v", subscriptionIDs[i], err)
			continue
		}

		c.log.Info("Successfully created topic/subscription pair: %s/%s", topic.Name, subscriptionIDs[i])
	}

	return nil
}
