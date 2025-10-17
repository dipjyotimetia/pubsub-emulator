package pubsub

import (
	"context"
	"testing"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/pstest"
	"github.com/dipjyotimetia/pubsub-emulator/pkg/logger"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func setupTestServer(t *testing.T) (*pstest.Server, *Client, func()) {
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

	cleanup := func() {
		gcpClient.Close()
		conn.Close()
		srv.Close()
	}

	return srv, client, cleanup
}

func TestNewClient(t *testing.T) {
	srv := pstest.NewServer()
	defer srv.Close()

	ctx := context.Background()
	conn, err := grpc.Dial(srv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	log := logger.New()

	// Create client using the test server
	gcpClient, err := pubsub.NewClient(ctx, "test-project", option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer gcpClient.Close()

	client := &Client{
		client:    gcpClient,
		projectID: "test-project",
		log:       log,
	}

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	if client.projectID != "test-project" {
		t.Errorf("Expected projectID 'test-project', got '%s'", client.projectID)
	}
}

func TestClient_ProjectID(t *testing.T) {
	_, client, cleanup := setupTestServer(t)
	defer cleanup()

	if client.ProjectID() != "test-project" {
		t.Errorf("Expected ProjectID 'test-project', got '%s'", client.ProjectID())
	}
}

func TestClient_GetClient(t *testing.T) {
	_, client, cleanup := setupTestServer(t)
	defer cleanup()

	gcpClient := client.GetClient()
	if gcpClient == nil {
		t.Error("Expected GetClient to return non-nil client")
	}
}

func TestClient_Close(t *testing.T) {
	_, client, cleanup := setupTestServer(t)
	defer cleanup()

	err := client.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got %v", err)
	}
}

func TestClient_CreateTopic(t *testing.T) {
	_, client, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	topic, err := client.CreateTopic(ctx, "test-topic")

	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	if topic == nil {
		t.Fatal("Expected topic to be created, got nil")
	}

	expectedName := "projects/test-project/topics/test-topic"
	if topic.Name != expectedName {
		t.Errorf("Expected topic name '%s', got '%s'", expectedName, topic.Name)
	}
}

func TestClient_CreateTopic_AlreadyExists(t *testing.T) {
	_, client, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create topic first time
	_, err := client.CreateTopic(ctx, "test-topic")
	if err != nil {
		t.Fatalf("Failed to create topic first time: %v", err)
	}

	// Try to create same topic again
	_, err = client.CreateTopic(ctx, "test-topic")
	if err == nil {
		t.Error("Expected error when creating duplicate topic, got nil")
	}
}

func TestClient_CreateSubscription(t *testing.T) {
	_, client, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create topic first
	_, err := client.CreateTopic(ctx, "test-topic")
	if err != nil {
		t.Fatalf("Failed to create topic: %v", err)
	}

	// Create subscription
	sub, err := client.CreateSubscription(ctx, "test-sub", "test-topic", 20)
	if err != nil {
		t.Fatalf("Failed to create subscription: %v", err)
	}

	if sub == nil {
		t.Fatal("Expected subscription to be created, got nil")
	}

	expectedName := "projects/test-project/subscriptions/test-sub"
	if sub.Name != expectedName {
		t.Errorf("Expected subscription name '%s', got '%s'", expectedName, sub.Name)
	}

	if sub.AckDeadlineSeconds != 20 {
		t.Errorf("Expected AckDeadlineSeconds 20, got %d", sub.AckDeadlineSeconds)
	}
}

func TestClient_CreateSubscription_TopicNotFound(t *testing.T) {
	_, client, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Try to create subscription for non-existent topic
	_, err := client.CreateSubscription(ctx, "test-sub", "nonexistent-topic", 20)
	if err == nil {
		t.Error("Expected error when creating subscription for non-existent topic, got nil")
	}
}

func TestClient_CreateTopicsAndSubscriptions(t *testing.T) {
	_, client, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	topicIDs := []string{"topic1", "topic2", "topic3"}
	subIDs := []string{"sub1", "sub2", "sub3"}

	err := client.CreateTopicsAndSubscriptions(ctx, topicIDs, subIDs)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestClient_CreateTopicsAndSubscriptions_MismatchedCounts(t *testing.T) {
	_, client, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	topicIDs := []string{"topic1", "topic2"}
	subIDs := []string{"sub1"}

	err := client.CreateTopicsAndSubscriptions(ctx, topicIDs, subIDs)
	if err == nil {
		t.Error("Expected error for mismatched topic/subscription counts, got nil")
	}
}

func TestClient_CreateTopicsAndSubscriptions_EmptySlices(t *testing.T) {
	_, client, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	err := client.CreateTopicsAndSubscriptions(ctx, []string{}, []string{})
	if err != nil {
		t.Errorf("Expected no error for empty slices, got %v", err)
	}
}
