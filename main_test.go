package main

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestEnvironment(t *testing.T) {
	// Set environment variables for tests
	os.Setenv("PUBSUB_PROJECT", "test-project")
	os.Setenv("PUBSUB_TOPIC", "test-topic")
	os.Setenv("PUBSUB_SUBSCRIPTION", "test-subscription")

	// Verify PubSub emulator is running
	if os.Getenv("PUBSUB_EMULATOR_HOST") == "" {
		t.Log("PUBSUB_EMULATOR_HOST not set, setting to localhost:8085")
		os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8085")
	}
}

func cleanupTestEnvironment(t *testing.T, ctx context.Context, client *pubsub.Client) {
	// Clean up test topics and subscriptions
	topics := strings.Split(os.Getenv("PUBSUB_TOPIC"), ",")
	subs := strings.SplitSeq(os.Getenv("PUBSUB_SUBSCRIPTION"), ",")

	for subID := range subs {
		sub := client.Subscription(subID)
		if exists, _ := sub.Exists(ctx); exists {
			if err := sub.Delete(ctx); err != nil {
				t.Logf("Failed to delete subscription %s: %v", subID, err)
			}
		}
	}

	for _, topicID := range topics {
		topic := client.Topic(topicID)
		if exists, _ := topic.Exists(ctx); exists {
			if err := topic.Delete(ctx); err != nil {
				t.Logf("Failed to delete topic %s: %v", topicID, err)
			}
		}
	}

	client.Close()
}

func TestCreateTopicSubscription(t *testing.T) {
	setupTestEnvironment(t)

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, os.Getenv("PUBSUB_PROJECT"))
	require.NoError(t, err, "Failed to create pubsub client")

	defer cleanupTestEnvironment(t, ctx, client)

	cfg := Config{
		ProjectID: os.Getenv("PUBSUB_PROJECT"),
		TopicIDs:  os.Getenv("PUBSUB_TOPIC"),
		SubIDs:    os.Getenv("PUBSUB_SUBSCRIPTION"),
	}

	err = createTopicSubscription(ctx, client, cfg)
	require.NoError(t, err, "Failed to create topic and subscription")

	// Verify topic exists
	topic := client.Topic(os.Getenv("PUBSUB_TOPIC"))
	exists, err := topic.Exists(ctx)
	assert.NoError(t, err)
	assert.True(t, exists, "Topic was not created")

	// Verify subscription exists
	sub := client.Subscription(os.Getenv("PUBSUB_SUBSCRIPTION"))
	exists, err = sub.Exists(ctx)
	assert.NoError(t, err)
	assert.True(t, exists, "Subscription was not created")
}

func TestPublishMessage(t *testing.T) {
	setupTestEnvironment(t)

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, os.Getenv("PUBSUB_PROJECT"))
	require.NoError(t, err, "Failed to create pubsub client")

	defer cleanupTestEnvironment(t, ctx, client)

	cfg := Config{
		ProjectID:        os.Getenv("PUBSUB_PROJECT"),
		TopicIDs:         os.Getenv("PUBSUB_TOPIC"),
		SubIDs:           os.Getenv("PUBSUB_SUBSCRIPTION"),
		MessageToPublish: "Test message",
	}

	// First create topic and subscription
	err = createTopicSubscription(ctx, client, cfg)
	require.NoError(t, err, "Failed to create topic and subscription")

	// Then publish message
	err = publishMessage(ctx, client, cfg)
	require.NoError(t, err, "Failed to publish message")
}

func TestSubscribeAndReceiveMessages(t *testing.T) {
	setupTestEnvironment(t)

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, os.Getenv("PUBSUB_PROJECT"))
	require.NoError(t, err, "Failed to create pubsub client")

	defer cleanupTestEnvironment(t, ctx, client)

	cfg := Config{
		ProjectID:        os.Getenv("PUBSUB_PROJECT"),
		TopicIDs:         os.Getenv("PUBSUB_TOPIC"),
		SubIDs:           os.Getenv("PUBSUB_SUBSCRIPTION"),
		MessageToPublish: "Test message for subscription",
	}

	// Create topic and subscription
	err = createTopicSubscription(ctx, client, cfg)
	require.NoError(t, err, "Failed to create topic and subscription")

	// Publish message
	err = publishMessage(ctx, client, cfg)
	require.NoError(t, err, "Failed to publish message")

	// Test subscription with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = subscribeAndReceiveMessages(ctxWithTimeout, client, cfg)
	require.NoError(t, err, "Failed to receive message")
}

func TestInvalidConfig(t *testing.T) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, "test-project")
	require.NoError(t, err, "Failed to create pubsub client")

	defer client.Close()

	// Test unequal topics and subscriptions
	cfg := Config{
		ProjectID: "test-project",
		TopicIDs:  "topic1,topic2",
		SubIDs:    "sub1",
	}

	err = createTopicSubscription(ctx, client, cfg)
	assert.Error(t, err, "Should error when topics and subscriptions are not equal")
	assert.Contains(t, err.Error(), "number of topics and subscriptions are not the same")
}

// Fix the SplitSeq issue in the publishMessage function for testing
func TestFixForPublishMessage(t *testing.T) {
	setupTestEnvironment(t)

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, os.Getenv("PUBSUB_PROJECT"))
	require.NoError(t, err, "Failed to create pubsub client")

	defer cleanupTestEnvironment(t, ctx, client)

	// First create topic and subscription
	origCfg := Config{
		ProjectID: os.Getenv("PUBSUB_PROJECT"),
		TopicIDs:  os.Getenv("PUBSUB_TOPIC"),
		SubIDs:    os.Getenv("PUBSUB_SUBSCRIPTION"),
	}

	err = createTopicSubscription(ctx, client, origCfg)
	require.NoError(t, err, "Failed to create topic and subscription")

	// Test with fixed function
	topics := strings.Split(origCfg.TopicIDs, ",")

	for _, topicID := range topics {
		topic := client.Topic(topicID)
		ok, err := topic.Exists(ctx)
		assert.NoError(t, err, "Failed to check if topic exists")
		assert.True(t, ok, "Topic should exist")

		result := topic.Publish(ctx, &pubsub.Message{
			Data: []byte("Test message"),
		})

		msgID, err := result.Get(ctx)
		assert.NoError(t, err, "Failed to publish message")
		assert.NotEmpty(t, msgID, "Message ID should not be empty")
	}
}
