package dashboard

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"cloud.google.com/go/pubsub/v2/pstest"
	"github.com/dipjyotimetia/pubsub-emulator/pkg/logger"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestNew(t *testing.T) {
	log := logger.New()

	dash := New(nil, "test-project", log)

	if dash == nil {
		t.Fatal("Expected dashboard to be created, got nil")
	}

	if dash.projectID != "test-project" {
		t.Errorf("Expected projectID 'test-project', got '%s'", dash.projectID)
	}

	if dash.maxMessages != 1000 {
		t.Errorf("Expected maxMessages 1000, got %d", dash.maxMessages)
	}

	if len(dash.messages) != 0 {
		t.Errorf("Expected empty messages slice, got %d messages", len(dash.messages))
	}
}

func TestAddMessage(t *testing.T) {
	log := logger.New()
	dash := New(nil, "test-project", log)

	msg := &pubsub.Message{
		ID:   "msg-123",
		Data: []byte("test data"),
		Attributes: map[string]string{
			"key": "value",
		},
		PublishTime: time.Now(),
	}

	dash.AddMessage(msg, "test-topic")

	if len(dash.messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(dash.messages))
	}

	storedMsg := dash.messages[0]
	if storedMsg.ID != "msg-123" {
		t.Errorf("Expected ID 'msg-123', got '%s'", storedMsg.ID)
	}

	if storedMsg.Data != "test data" {
		t.Errorf("Expected Data 'test data', got '%s'", storedMsg.Data)
	}

	if storedMsg.Topic != "test-topic" {
		t.Errorf("Expected Topic 'test-topic', got '%s'", storedMsg.Topic)
	}

	if storedMsg.Attributes["key"] != "value" {
		t.Errorf("Expected Attribute key='value', got '%s'", storedMsg.Attributes["key"])
	}
}

func TestAddMessage_MaxMessagesLimit(t *testing.T) {
	log := logger.New()
	dash := New(nil, "test-project", log)
	dash.maxMessages = 10

	// Add more messages than the limit
	for i := 0; i < 15; i++ {
		msg := &pubsub.Message{
			ID:          string(rune(i)),
			Data:        []byte("test"),
			PublishTime: time.Now(),
		}
		dash.AddMessage(msg, "test-topic")
	}

	if len(dash.messages) != 10 {
		t.Errorf("Expected messages to be capped at 10, got %d", len(dash.messages))
	}
}

func TestGetMessages(t *testing.T) {
	log := logger.New()
	dash := New(nil, "test-project", log)

	msg1 := &pubsub.Message{
		ID:          "msg-1",
		Data:        []byte("data1"),
		PublishTime: time.Now(),
	}
	msg2 := &pubsub.Message{
		ID:          "msg-2",
		Data:        []byte("data2"),
		PublishTime: time.Now(),
	}

	dash.AddMessage(msg1, "topic1")
	dash.AddMessage(msg2, "topic2")

	messages := dash.GetMessages()

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	if messages[0].ID != "msg-1" {
		t.Errorf("Expected first message ID 'msg-1', got '%s'", messages[0].ID)
	}

	if messages[1].ID != "msg-2" {
		t.Errorf("Expected second message ID 'msg-2', got '%s'", messages[1].ID)
	}
}

func TestGetMessageByID_Found(t *testing.T) {
	log := logger.New()
	dash := New(nil, "test-project", log)

	msg := &pubsub.Message{
		ID:          "msg-123",
		Data:        []byte("test data"),
		PublishTime: time.Now(),
	}

	dash.AddMessage(msg, "test-topic")

	foundMsg := dash.GetMessageByID("msg-123")

	if foundMsg == nil {
		t.Fatal("Expected to find message, got nil")
	}

	if foundMsg.ID != "msg-123" {
		t.Errorf("Expected ID 'msg-123', got '%s'", foundMsg.ID)
	}

	if foundMsg.Data != "test data" {
		t.Errorf("Expected Data 'test data', got '%s'", foundMsg.Data)
	}
}

func TestGetMessageByID_NotFound(t *testing.T) {
	log := logger.New()
	dash := New(nil, "test-project", log)

	msg := &pubsub.Message{
		ID:          "msg-123",
		Data:        []byte("test data"),
		PublishTime: time.Now(),
	}

	dash.AddMessage(msg, "test-topic")

	foundMsg := dash.GetMessageByID("nonexistent")

	if foundMsg != nil {
		t.Errorf("Expected nil for nonexistent message, got %v", foundMsg)
	}
}

func TestExtractID(t *testing.T) {
	tests := []struct {
		name     string
		fullName string
		expected string
	}{
		{
			name:     "Topic name",
			fullName: "projects/test-project/topics/my-topic",
			expected: "my-topic",
		},
		{
			name:     "Subscription name",
			fullName: "projects/test-project/subscriptions/my-sub",
			expected: "my-sub",
		},
		{
			name:     "Simple name",
			fullName: "simple",
			expected: "simple",
		},
		{
			name:     "Empty string",
			fullName: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractID(tt.fullName)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestMessageInfo_Structure(t *testing.T) {
	now := time.Now()
	msgInfo := MessageInfo{
		ID:   "test-id",
		Data: "test data",
		Attributes: map[string]string{
			"attr1": "val1",
		},
		PublishTime: now,
		Topic:       "test-topic",
		Received:    now,
	}

	if msgInfo.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", msgInfo.ID)
	}

	if msgInfo.Data != "test data" {
		t.Errorf("Expected Data 'test data', got '%s'", msgInfo.Data)
	}

	if msgInfo.Topic != "test-topic" {
		t.Errorf("Expected Topic 'test-topic', got '%s'", msgInfo.Topic)
	}

	if msgInfo.Attributes["attr1"] != "val1" {
		t.Errorf("Expected Attribute attr1='val1', got '%s'", msgInfo.Attributes["attr1"])
	}
}

func TestConcurrentAddMessage(t *testing.T) {
	log := logger.New()
	dash := New(nil, "test-project", log)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			msg := &pubsub.Message{
				ID:          string(rune(id)),
				Data:        []byte("concurrent test"),
				PublishTime: time.Now(),
			}
			dash.AddMessage(msg, "test-topic")
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	messages := dash.GetMessages()
	if len(messages) != 10 {
		t.Errorf("Expected 10 messages after concurrent adds, got %d", len(messages))
	}
}

func TestConcurrentGetMessages(t *testing.T) {
	log := logger.New()
	dash := New(nil, "test-project", log)

	// Add some messages first
	for i := 0; i < 5; i++ {
		msg := &pubsub.Message{
			ID:          string(rune(i)),
			Data:        []byte("test"),
			PublishTime: time.Now(),
		}
		dash.AddMessage(msg, "test-topic")
	}

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			messages := dash.GetMessages()
			if len(messages) < 5 {
				t.Errorf("Expected at least 5 messages, got %d", len(messages))
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func setupDashboardTest(t *testing.T) (*pstest.Server, *Dashboard, func()) {
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
	dash := New(gcpClient, "test-project", log)

	cleanup := func() {
		gcpClient.Close()
		conn.Close()
		srv.Close()
	}

	return srv, dash, cleanup
}

func TestDashboard_GetStats(t *testing.T) {
	_, dash, cleanup := setupDashboardTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create some topics and subscriptions
	topicIDs := []string{"topic1", "topic2"}
	for _, topicID := range topicIDs {
		dash.client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{
			Name: "projects/test-project/topics/" + topicID,
		})
	}

	for i, topicID := range topicIDs {
		dash.client.SubscriptionAdminClient.CreateSubscription(ctx, &pubsubpb.Subscription{
			Name:               "projects/test-project/subscriptions/sub" + string(rune(i)),
			Topic:              "projects/test-project/topics/" + topicID,
			AckDeadlineSeconds: 20,
		})
	}

	// Add some messages
	for i := 0; i < 5; i++ {
		msg := &pubsub.Message{
			ID:          string(rune(i)),
			Data:        []byte("test message"),
			PublishTime: time.Now(),
		}
		dash.AddMessage(msg, "topic1")
	}

	// Get stats
	stats, err := dash.GetStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.TopicCount != 2 {
		t.Errorf("Expected 2 topics, got %d", stats.TopicCount)
	}

	if stats.SubCount != 2 {
		t.Errorf("Expected 2 subscriptions, got %d", stats.SubCount)
	}

	if stats.MessageCount != 5 {
		t.Errorf("Expected 5 messages, got %d", stats.MessageCount)
	}

	if stats.TotalMessages != 5 {
		t.Errorf("Expected 5 total messages, got %d", stats.TotalMessages)
	}

	if stats.LastMessageTime == nil {
		t.Error("Expected LastMessageTime to be set")
	}
}

func TestDashboard_GetStats_EmptyDashboard(t *testing.T) {
	_, dash, cleanup := setupDashboardTest(t)
	defer cleanup()

	ctx := context.Background()

	stats, err := dash.GetStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.TopicCount != 0 {
		t.Errorf("Expected 0 topics, got %d", stats.TopicCount)
	}

	if stats.SubCount != 0 {
		t.Errorf("Expected 0 subscriptions, got %d", stats.SubCount)
	}

	if stats.MessageCount != 0 {
		t.Errorf("Expected 0 messages, got %d", stats.MessageCount)
	}

	if stats.LastMessageTime != nil {
		t.Error("Expected LastMessageTime to be nil for empty dashboard")
	}
}
