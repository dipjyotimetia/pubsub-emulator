package pubsub

import (
	"context"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/pstest"
	"github.com/dipjyotimetia/pubsub-emulator/pkg/logger"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func setupSubscriberTest(t *testing.T) (*pstest.Server, *Subscriber, *Publisher, func()) {
	t.Helper()

	// Create fake Pub/Sub server
	srv := pstest.NewServer()

	// Create connection to the fake server
	ctx := context.Background()
	conn, err := grpc.Dial(srv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial test server: %v", err)
	}

	// Create client connected to the fake server
	gcpClient, err := pubsub.NewClient(ctx, "test-project", option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("Failed to create pubsub client: %v", err)
	}

	log := logger.New()
	client := &Client{
		client:    gcpClient,
		projectID: "test-project",
		log:       log,
	}

	subscriber := NewSubscriber(client, log)
	publisher := NewPublisher(client, log)

	cleanup := func() {
		gcpClient.Close()
		conn.Close()
		srv.Close()
	}

	return srv, subscriber, publisher, cleanup
}

func TestNewSubscriber(t *testing.T) {
	_, sub, _, cleanup := setupSubscriberTest(t)
	defer cleanup()

	if sub == nil {
		t.Fatal("Expected subscriber to be created, got nil")
	}

	if sub.client == nil {
		t.Error("Expected client to be set")
	}

	if sub.log == nil {
		t.Error("Expected logger to be set")
	}
}

func TestSubscriber_Subscribe(t *testing.T) {
	_, sub, pub, cleanup := setupSubscriberTest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create topic and subscription
	_, err := sub.client.CreateTopic(ctx, "test-topic")
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	_, err = sub.client.CreateSubscription(ctx, "test-sub", "test-topic", 20)
	if err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	// Publish a message
	_, err = pub.PublishMessage(ctx, "test-topic", "test message", nil)
	if err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}

	// Subscribe and receive message
	var receivedMsg *pubsub.Message
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(ctx context.Context, msg *pubsub.Message) {
		receivedMsg = msg
		wg.Done()
		cancel() // Cancel context to stop receiving
	}

	go func() {
		sub.Subscribe(ctx, "test-sub", handler)
	}()

	// Wait for message or timeout
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		if receivedMsg == nil {
			t.Error("Expected to receive a message, got nil")
		} else if string(receivedMsg.Data) != "test message" {
			t.Errorf("Expected message data 'test message', got '%s'", string(receivedMsg.Data))
		}
	case <-time.After(3 * time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestSubscriber_Subscribe_MultipleMessages(t *testing.T) {
	_, sub, pub, cleanup := setupSubscriberTest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create topic and subscription
	_, err := sub.client.CreateTopic(ctx, "test-topic")
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	_, err = sub.client.CreateSubscription(ctx, "test-sub", "test-topic", 20)
	if err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	// Publish multiple messages
	numMessages := 3
	for i := 0; i < numMessages; i++ {
		_, err = pub.PublishMessage(ctx, "test-topic", "test message", nil)
		if err != nil {
			t.Fatalf("Failed to publish message %d: %v", i, err)
		}
	}

	// Subscribe and count received messages
	receivedCount := 0
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(numMessages)

	handler := func(ctx context.Context, msg *pubsub.Message) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
		wg.Done()
		if receivedCount >= numMessages {
			cancel() // Cancel after receiving all messages
		}
	}

	go func() {
		sub.Subscribe(ctx, "test-sub", handler)
	}()

	// Wait for all messages or timeout
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		mu.Lock()
		count := receivedCount
		mu.Unlock()
		if count != numMessages {
			t.Errorf("Expected to receive %d messages, got %d", numMessages, count)
		}
	case <-time.After(3 * time.Second):
		mu.Lock()
		count := receivedCount
		mu.Unlock()
		t.Errorf("Timeout: only received %d/%d messages", count, numMessages)
	}
}

func TestSubscriber_SubscribeToAll(t *testing.T) {
	_, sub, pub, cleanup := setupSubscriberTest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create topics and subscriptions
	topicIDs := []string{"topic1", "topic2", "topic3"}
	subIDs := []string{"sub1", "sub2", "sub3"}

	for i := range topicIDs {
		_, err := sub.client.CreateTopic(ctx, topicIDs[i])
		if err != nil {
			t.Fatalf("Failed to create topic %s: %v", topicIDs[i], err)
		}

		_, err = sub.client.CreateSubscription(ctx, subIDs[i], topicIDs[i], 20)
		if err != nil {
			t.Fatalf("Failed to create subscription %s: %v", subIDs[i], err)
		}
	}

	// Publish one message to each topic
	for _, topicID := range topicIDs {
		_, err := pub.PublishMessage(ctx, topicID, "test message", nil)
		if err != nil {
			t.Fatalf("Failed to publish to %s: %v", topicID, err)
		}
	}

	// Subscribe to all subscriptions
	receivedCount := 0
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(subIDs))

	handler := func(ctx context.Context, msg *pubsub.Message) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
		wg.Done()
		if receivedCount >= len(subIDs) {
			cancel()
		}
	}

	sub.SubscribeToAll(ctx, subIDs, topicIDs, handler)

	// Wait for all messages or timeout
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		mu.Lock()
		count := receivedCount
		mu.Unlock()
		if count != len(subIDs) {
			t.Errorf("Expected to receive %d messages, got %d", len(subIDs), count)
		}
	case <-time.After(3 * time.Second):
		mu.Lock()
		count := receivedCount
		mu.Unlock()
		t.Errorf("Timeout: only received %d/%d messages", count, len(subIDs))
	}
}

func TestSubscriber_Subscribe_WithAttributes(t *testing.T) {
	_, sub, pub, cleanup := setupSubscriberTest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create topic and subscription
	_, err := sub.client.CreateTopic(ctx, "test-topic")
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	_, err = sub.client.CreateSubscription(ctx, "test-sub", "test-topic", 20)
	if err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	// Publish message with attributes
	expectedAttrs := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	_, err = pub.PublishMessage(ctx, "test-topic", "test message", expectedAttrs)
	if err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}

	// Subscribe and verify attributes
	var receivedMsg *pubsub.Message
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(ctx context.Context, msg *pubsub.Message) {
		receivedMsg = msg
		wg.Done()
		cancel()
	}

	go func() {
		sub.Subscribe(ctx, "test-sub", handler)
	}()

	// Wait for message or timeout
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		if receivedMsg == nil {
			t.Fatal("Expected to receive a message, got nil")
		}
		for key, expectedValue := range expectedAttrs {
			if receivedMsg.Attributes[key] != expectedValue {
				t.Errorf("Expected attribute %s='%s', got '%s'", key, expectedValue, receivedMsg.Attributes[key])
			}
		}
	case <-time.After(3 * time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestSubscriber_Subscribe_NonexistentSubscription(t *testing.T) {
	_, sub, _, cleanup := setupSubscriberTest(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	handler := func(ctx context.Context, msg *pubsub.Message) {
		t.Error("Should not receive any messages from non-existent subscription")
	}

	// Try to subscribe to non-existent subscription
	err := sub.Subscribe(ctx, "nonexistent-sub", handler)

	// Should return an error or context cancellation
	if err == nil && ctx.Err() == nil {
		t.Error("Expected error when subscribing to non-existent subscription")
	}
}
