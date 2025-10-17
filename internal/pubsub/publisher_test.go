package pubsub

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/pstest"
	"github.com/dipjyotimetia/pubsub-emulator/pkg/logger"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestCreateMessageInfo(t *testing.T) {
	now := time.Now()
	msg := &pubsub.Message{
		ID:   "test-msg-123",
		Data: []byte("test data"),
		Attributes: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		PublishTime: now,
	}
	topicID := "test-topic"

	msgInfo := CreateMessageInfo(msg, topicID)

	if msgInfo.ID != "test-msg-123" {
		t.Errorf("Expected ID 'test-msg-123', got '%s'", msgInfo.ID)
	}

	if msgInfo.Data != "test data" {
		t.Errorf("Expected Data 'test data', got '%s'", msgInfo.Data)
	}

	if msgInfo.Topic != "test-topic" {
		t.Errorf("Expected Topic 'test-topic', got '%s'", msgInfo.Topic)
	}

	if len(msgInfo.Attributes) != 2 {
		t.Errorf("Expected 2 attributes, got %d", len(msgInfo.Attributes))
	}

	if msgInfo.Attributes["key1"] != "value1" {
		t.Errorf("Expected attribute key1='value1', got '%s'", msgInfo.Attributes["key1"])
	}

	if msgInfo.PublishTime != now {
		t.Errorf("Expected PublishTime to match, got different time")
	}

	if msgInfo.Received.IsZero() {
		t.Error("Expected Received time to be set")
	}
}

func TestCreateMessageInfo_EmptyData(t *testing.T) {
	msg := &pubsub.Message{
		ID:          "test-msg-456",
		Data:        []byte{},
		Attributes:  map[string]string{},
		PublishTime: time.Now(),
	}
	topicID := "empty-topic"

	msgInfo := CreateMessageInfo(msg, topicID)

	if msgInfo.Data != "" {
		t.Errorf("Expected empty Data, got '%s'", msgInfo.Data)
	}

	if len(msgInfo.Attributes) != 0 {
		t.Errorf("Expected 0 attributes, got %d", len(msgInfo.Attributes))
	}
}

func TestCreateMessageInfo_NilAttributes(t *testing.T) {
	msg := &pubsub.Message{
		ID:          "test-msg-789",
		Data:        []byte("test"),
		Attributes:  nil,
		PublishTime: time.Now(),
	}
	topicID := "nil-attr-topic"

	msgInfo := CreateMessageInfo(msg, topicID)

	if msgInfo.Attributes != nil {
		t.Errorf("Expected nil Attributes, got %v", msgInfo.Attributes)
	}
}

func TestMessageInfo_Fields(t *testing.T) {
	msgInfo := MessageInfo{
		ID:   "test-123",
		Data: "test data",
		Attributes: map[string]string{
			"attr1": "val1",
		},
		PublishTime: time.Now(),
		Topic:       "test-topic",
		Received:    time.Now(),
	}

	if msgInfo.ID != "test-123" {
		t.Errorf("Expected ID 'test-123', got '%s'", msgInfo.ID)
	}

	if msgInfo.Data != "test data" {
		t.Errorf("Expected Data 'test data', got '%s'", msgInfo.Data)
	}

	if msgInfo.Topic != "test-topic" {
		t.Errorf("Expected Topic 'test-topic', got '%s'", msgInfo.Topic)
	}

	if msgInfo.Attributes["attr1"] != "val1" {
		t.Errorf("Expected attribute attr1='val1', got '%s'", msgInfo.Attributes["attr1"])
	}

	if msgInfo.PublishTime.IsZero() {
		t.Error("Expected PublishTime to be set")
	}

	if msgInfo.Received.IsZero() {
		t.Error("Expected Received time to be set")
	}
}

func setupPublisherTest(t *testing.T) (*pstest.Server, *Publisher, func()) {
	t.Helper()

	// Create fake Pub/Sub server
	srv := pstest.NewServer()

	// Create connection to the fake server
	ctx := context.Background()
	conn, err := grpc.NewClient(srv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	publisher := NewPublisher(client, log)

	cleanup := func() {
		_ = gcpClient.Close()
		_ = conn.Close()
		_ = srv.Close()
	}

	return srv, publisher, cleanup
}

func TestNewPublisher(t *testing.T) {
	_, pub, cleanup := setupPublisherTest(t)
	defer cleanup()

	if pub == nil {
		t.Fatal("Expected publisher to be created, got nil")
	}

	if pub.client == nil {
		t.Error("Expected client to be set")
	}

	if pub.log == nil {
		t.Error("Expected logger to be set")
	}
}

func TestPublisher_PublishMessage(t *testing.T) {
	srv, pub, cleanup := setupPublisherTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create topic first
	_, err := pub.client.CreateTopic(ctx, "test-topic")
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	// Publish message
	msgID, err := pub.PublishMessage(ctx, "test-topic", "test data", map[string]string{
		"key": "value",
	})

	if err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}

	if msgID == "" {
		t.Error("Expected non-empty message ID")
	}

	// Verify message was published
	messages := srv.Messages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if string(msg.Data) != "test data" {
		t.Errorf("Expected data 'test data', got '%s'", string(msg.Data))
	}

	if msg.Attributes["key"] != "value" {
		t.Errorf("Expected attribute key='value', got '%s'", msg.Attributes["key"])
	}
}

func TestPublisher_PublishMessage_TopicNotFound(t *testing.T) {
	_, pub, cleanup := setupPublisherTest(t)
	defer cleanup()

	ctx := context.Background()

	// Try to publish to non-existent topic
	_, err := pub.PublishMessage(ctx, "nonexistent-topic", "test data", nil)
	if err == nil {
		t.Error("Expected error when publishing to non-existent topic, got nil")
	}
}

func TestPublisher_PublishToTopics(t *testing.T) {
	srv, pub, cleanup := setupPublisherTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create topics
	topicIDs := []string{"topic1", "topic2", "topic3"}
	for _, topicID := range topicIDs {
		_, err := pub.client.CreateTopic(ctx, topicID)
		if err != nil {
			t.Fatalf("Failed to create topic %s: %v", topicID, err)
		}
	}

	// Publish to all topics
	messageIDs, err := pub.PublishToTopics(ctx, topicIDs, "broadcast message")
	if err != nil {
		t.Fatalf("Failed to publish to topics: %v", err)
	}

	if len(messageIDs) != 3 {
		t.Errorf("Expected 3 message IDs, got %d", len(messageIDs))
	}

	// Verify all messages were published
	messages := srv.Messages()
	if len(messages) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(messages))
	}

	for _, msg := range messages {
		if string(msg.Data) != "broadcast message" {
			t.Errorf("Expected data 'broadcast message', got '%s'", string(msg.Data))
		}
	}
}

func TestPublisher_PublishToTopics_SomeTopicsNotFound(t *testing.T) {
	_, pub, cleanup := setupPublisherTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create only one topic
	_, err := pub.client.CreateTopic(ctx, "topic1")
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	// Try to publish to multiple topics (some don't exist)
	topicIDs := []string{"topic1", "nonexistent-topic", "another-nonexistent"}
	messageIDs, err := pub.PublishToTopics(ctx, topicIDs, "test message")

	// Should succeed for at least one topic
	if err != nil {
		t.Errorf("Expected partial success, got error: %v", err)
	}

	if len(messageIDs) != 1 {
		t.Errorf("Expected 1 successful publish, got %d", len(messageIDs))
	}
}

func TestPublisher_PublishToTopics_AllTopicsNotFound(t *testing.T) {
	_, pub, cleanup := setupPublisherTest(t)
	defer cleanup()

	ctx := context.Background()

	// Try to publish to non-existent topics
	topicIDs := []string{"nonexistent1", "nonexistent2"}
	_, err := pub.PublishToTopics(ctx, topicIDs, "test message")

	if err == nil {
		t.Error("Expected error when all topics are non-existent, got nil")
	}
}

func TestPublisher_PublishMessage_WithAttributes(t *testing.T) {
	srv, pub, cleanup := setupPublisherTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create topic
	_, err := pub.client.CreateTopic(ctx, "test-topic")
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	attributes := map[string]string{
		"source":    "test",
		"timestamp": time.Now().String(),
		"priority":  "high",
	}

	// Publish message with multiple attributes
	msgID, err := pub.PublishMessage(ctx, "test-topic", "test data", attributes)
	if err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}

	if msgID == "" {
		t.Error("Expected non-empty message ID")
	}

	// Verify attributes
	messages := srv.Messages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	for key, expectedValue := range attributes {
		if msg.Attributes[key] != expectedValue {
			t.Errorf("Expected attribute %s='%s', got '%s'", key, expectedValue, msg.Attributes[key])
		}
	}
}
