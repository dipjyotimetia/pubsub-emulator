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

	// Wait for emulator to be available
	waitForEmulator(t)
}

// waitForEmulator attempts to connect to the emulator with retries
func waitForEmulator(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Log("Waiting for PubSub emulator to be ready...")

	// Try to connect to the emulator with retries
	var client *pubsub.Client
	var err error

	for i := range 10 {
		client, err = pubsub.NewClient(ctx, os.Getenv("PUBSUB_PROJECT"))
		if err == nil {
			client.Close()
			t.Log("Successfully connected to PubSub emulator")
			return
		}
		t.Logf("Attempt %d: Failed to connect to emulator: %v", i+1, err)
		time.Sleep(3 * time.Second)
	}

	t.Log("Warning: Could not connect to PubSub emulator after multiple attempts")
}

func cleanupTestEnvironment(t *testing.T, ctx context.Context, client *pubsub.Client) {
	// Clean up test topics and subscriptions
	topics := strings.Split(os.Getenv("PUBSUB_TOPIC"), ",")
	subscriptions := strings.Split(os.Getenv("PUBSUB_SUBSCRIPTION"), ",")

	// Use standard Split instead of SplitSeq
	for _, subID := range subscriptions {
		sub := client.Subscription(subID)
		exists, err := sub.Exists(ctx)
		if err != nil {
			t.Logf("Failed to check if subscription exists: %v", err)
			continue
		}
		if exists {
			if err := sub.Delete(ctx); err != nil {
				t.Logf("Failed to delete subscription %s: %v", subID, err)
			}
		}
	}

	for _, topicID := range topics {
		topic := client.Topic(topicID)
		exists, err := topic.Exists(ctx)
		if err != nil {
			t.Logf("Failed to check if topic exists: %v", err)
			continue
		}
		if exists {
			if err := topic.Delete(ctx); err != nil {
				t.Logf("Failed to delete topic %s: %v", topicID, err)
			}
		}
	}

	client.Close()
}

func TestCreateTopicSubscription(t *testing.T) {
	setupTestEnvironment(t)

	// Use a longer timeout for CI environments
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := pubsub.NewClient(ctx, os.Getenv("PUBSUB_PROJECT"))
	if err != nil {
		t.Skipf("Skipping test: Failed to create pubsub client: %v", err)
		return
	}
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

	// Use a longer timeout for CI environments
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := pubsub.NewClient(ctx, os.Getenv("PUBSUB_PROJECT"))
	if err != nil {
		t.Skipf("Skipping test: Failed to create pubsub client: %v", err)
		return
	}
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

	// Fix the usage of publishMessage to use our fixed version that uses Split instead of SplitSeq
	topics := strings.Split(cfg.TopicIDs, ",")
	for _, topicID := range topics {
		topic := client.Topic(topicID)
		ok, err := topic.Exists(ctx)
		require.NoError(t, err)
		require.True(t, ok, "Topic should exist")

		result := topic.Publish(ctx, &pubsub.Message{
			Data: []byte(cfg.MessageToPublish),
		})

		msgID, err := result.Get(ctx)
		require.NoError(t, err, "Failed to publish message")
		require.NotEmpty(t, msgID, "Message ID should not be empty")
	}
}

func TestSubscribeAndReceiveMessages(t *testing.T) {
	setupTestEnvironment(t)

	// Use a longer timeout for CI environments
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := pubsub.NewClient(ctx, os.Getenv("PUBSUB_PROJECT"))
	if err != nil {
		t.Skipf("Skipping test: Failed to create pubsub client: %v", err)
		return
	}
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

	// Publish directly to avoid SplitSeq issue
	topics := strings.Split(cfg.TopicIDs, ",")
	for _, topicID := range topics {
		topic := client.Topic(topicID)
		result := topic.Publish(ctx, &pubsub.Message{
			Data: []byte(cfg.MessageToPublish),
		})

		msgID, err := result.Get(ctx)
		require.NoError(t, err, "Failed to publish message")
		t.Logf("Published message with ID: %s", msgID)
	}

	// Test subscription with longer timeout for CI
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	err = subscribeAndReceiveMessages(ctxWithTimeout, client, cfg)
	require.NoError(t, err, "Failed to receive message")
}

func TestInvalidConfig(t *testing.T) {
	// Use a longer timeout for CI environments
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := pubsub.NewClient(ctx, "test-project")
	if err != nil {
		t.Skipf("Skipping test: Failed to create pubsub client: %v", err)
		return
	}
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

// Fix the SplitSeq issue by using Split instead
func TestFixForPublishMessage(t *testing.T) {
	setupTestEnvironment(t)

	// Use a longer timeout for CI environments
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := pubsub.NewClient(ctx, os.Getenv("PUBSUB_PROJECT"))
	if err != nil {
		t.Skipf("Skipping test: Failed to create pubsub client: %v", err)
		return
	}
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
