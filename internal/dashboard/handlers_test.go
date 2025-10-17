package dashboard

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func setupHandlerTest(t *testing.T) (*Dashboard, func()) {
	t.Helper()

	srv := pstest.NewServer()
	ctx := context.Background()
	conn, err := grpc.NewClient(srv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}

	gcpClient, err := pubsub.NewClient(ctx, "test-project", option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	log := logger.New()
	dash := New(gcpClient, "test-project", log)

	cleanup := func() {
		_ = gcpClient.Close()
		_ = conn.Close()
		_ = srv.Close()
	}

	return dash, cleanup
}

func TestHandleHealth(t *testing.T) {
	dash, cleanup := setupHandlerTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	dash.handleHealth(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", result["status"])
	}

	if result["time"] == "" {
		t.Error("Expected time to be set")
	}
}

func TestHandleHealth_MethodNotAllowed(t *testing.T) {
	dash, cleanup := setupHandlerTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/health", nil)
	w := httptest.NewRecorder()

	dash.handleHealth(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

func TestHandleMessages(t *testing.T) {
	dash, cleanup := setupHandlerTest(t)
	defer cleanup()

	// Add some test messages
	msg1 := &pubsub.Message{
		ID:          "msg-1",
		Data:        []byte("test message 1"),
		PublishTime: time.Now(),
	}
	msg2 := &pubsub.Message{
		ID:          "msg-2",
		Data:        []byte("test message 2"),
		PublishTime: time.Now(),
	}
	dash.AddMessage(msg1, "topic1")
	dash.AddMessage(msg2, "topic2")

	req := httptest.NewRequest(http.MethodGet, "/api/messages", nil)
	w := httptest.NewRecorder()

	dash.handleMessages(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var messages []MessageInfo
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}
}

func TestHandleMessages_MethodNotAllowed(t *testing.T) {
	dash, cleanup := setupHandlerTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/messages", nil)
	w := httptest.NewRecorder()

	dash.handleMessages(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

func TestHandleSearchMessages(t *testing.T) {
	dash, cleanup := setupHandlerTest(t)
	defer cleanup()

	// Add test messages
	msg1 := &pubsub.Message{
		ID:          "msg-1",
		Data:        []byte("hello world"),
		PublishTime: time.Now(),
	}
	msg2 := &pubsub.Message{
		ID:          "msg-2",
		Data:        []byte("goodbye world"),
		PublishTime: time.Now(),
	}
	msg3 := &pubsub.Message{
		ID:          "msg-3",
		Data:        []byte("hello universe"),
		PublishTime: time.Now(),
	}
	dash.AddMessage(msg1, "topic1")
	dash.AddMessage(msg2, "topic1")
	dash.AddMessage(msg3, "topic2")

	tests := []struct {
		name          string
		query         string
		expectedCount int
	}{
		{"Search by term 'hello'", "/api/messages/search?q=hello", 2},
		{"Search by term 'goodbye'", "/api/messages/search?q=goodbye", 1},
		{"Filter by topic", "/api/messages/search?topic=topic2", 1},
		{"Search with term and topic", "/api/messages/search?q=hello&topic=topic1", 1},
		{"No results", "/api/messages/search?q=nonexistent", 0},
		{"All messages", "/api/messages/search", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.query, nil)
			w := httptest.NewRecorder()

			dash.handleSearchMessages(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}

			var messages []MessageInfo
			if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if len(messages) != tt.expectedCount {
				t.Errorf("Expected %d messages, got %d", tt.expectedCount, len(messages))
			}
		})
	}
}

func TestHandleCreateTopic(t *testing.T) {
	dash, cleanup := setupHandlerTest(t)
	defer cleanup()

	reqBody := CreateTopicRequest{
		TopicID: "new-topic",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/topics", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	dash.handleCreateTopic(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", result["status"])
	}
}

func TestHandleCreateTopic_InvalidRequest(t *testing.T) {
	dash, cleanup := setupHandlerTest(t)
	defer cleanup()

	tests := []struct {
		name           string
		requestBody    any
		expectedStatus int
	}{
		{
			name:           "Empty topic ID",
			requestBody:    CreateTopicRequest{TopicID: ""},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Topic ID too long",
			requestBody:    CreateTopicRequest{TopicID: string(make([]byte, 300))},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/topics", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			dash.handleCreateTopic(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestHandleCreateTopic_MethodNotAllowed(t *testing.T) {
	dash, cleanup := setupHandlerTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/topics", nil)
	w := httptest.NewRecorder()

	dash.handleCreateTopic(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

func TestHandleCreateTopic_InvalidContentType(t *testing.T) {
	dash, cleanup := setupHandlerTest(t)
	defer cleanup()

	reqBody := CreateTopicRequest{TopicID: "test-topic"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/topics", bytes.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	dash.handleCreateTopic(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("Expected status 415, got %d", resp.StatusCode)
	}
}

func TestHandleCreateSubscription(t *testing.T) {
	dash, cleanup := setupHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create topic first
	_, _ = dash.client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{
		Name: "projects/test-project/topics/test-topic",
	})

	reqBody := CreateSubscriptionRequest{
		SubscriptionID:     "new-sub",
		TopicID:            "test-topic",
		AckDeadlineSeconds: 30,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	dash.handleCreateSubscription(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", result["status"])
	}
}

func TestHandleCreateSubscription_InvalidRequest(t *testing.T) {
	dash, cleanup := setupHandlerTest(t)
	defer cleanup()

	tests := []struct {
		name           string
		requestBody    CreateSubscriptionRequest
		expectedStatus int
	}{
		{
			name: "Empty subscription ID",
			requestBody: CreateSubscriptionRequest{
				SubscriptionID:     "",
				TopicID:            "test-topic",
				AckDeadlineSeconds: 10,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Empty topic ID",
			requestBody: CreateSubscriptionRequest{
				SubscriptionID:     "test-sub",
				TopicID:            "",
				AckDeadlineSeconds: 10,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Subscription ID too long",
			requestBody: CreateSubscriptionRequest{
				SubscriptionID:     string(make([]byte, 300)),
				TopicID:            "test-topic",
				AckDeadlineSeconds: 10,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Ack deadline too long",
			requestBody: CreateSubscriptionRequest{
				SubscriptionID:     "test-sub",
				TopicID:            "test-topic",
				AckDeadlineSeconds: 700,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)

			req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			dash.handleCreateSubscription(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestHandlePublish(t *testing.T) {
	dash, cleanup := setupHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create topic first
	_, _ = dash.client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{
		Name: "projects/test-project/topics/test-topic",
	})

	reqBody := PublishRequest{
		TopicID: "test-topic",
		Data:    "test message",
		Attributes: map[string]string{
			"key": "value",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/publish", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	dash.handlePublish(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", result["status"])
	}

	if result["messageId"] == "" {
		t.Error("Expected messageId to be set")
	}
}

func TestHandlePublish_InvalidRequest(t *testing.T) {
	dash, cleanup := setupHandlerTest(t)
	defer cleanup()

	tests := []struct {
		name           string
		requestBody    PublishRequest
		expectedStatus int
	}{
		{
			name: "Empty topic ID",
			requestBody: PublishRequest{
				TopicID: "",
				Data:    "test",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Empty data",
			requestBody: PublishRequest{
				TopicID: "test-topic",
				Data:    "",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Data too large",
			requestBody: PublishRequest{
				TopicID: "test-topic",
				Data:    string(make([]byte, 11*1024*1024)), // > 10MB
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)

			req := httptest.NewRequest(http.MethodPost, "/api/publish", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			dash.handlePublish(w, req)

			resp := w.Result()
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestHandleReplay(t *testing.T) {
	dash, cleanup := setupHandlerTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create topic
	_, _ = dash.client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{
		Name: "projects/test-project/topics/test-topic",
	})

	// Add a message to replay
	msg := &pubsub.Message{
		ID:   "msg-to-replay",
		Data: []byte("replay me"),
		Attributes: map[string]string{
			"original": "true",
		},
		PublishTime: time.Now(),
	}
	dash.AddMessage(msg, "test-topic")

	req := httptest.NewRequest(http.MethodPost, "/api/replay?id=msg-to-replay", nil)
	w := httptest.NewRecorder()

	dash.handleReplay(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "success" {
		t.Errorf("Expected status 'success', got '%s'", result["status"])
	}

	if result["originalId"] != "msg-to-replay" {
		t.Errorf("Expected originalId 'msg-to-replay', got '%s'", result["originalId"])
	}
}

func TestHandleReplay_MessageNotFound(t *testing.T) {
	dash, cleanup := setupHandlerTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/replay?id=nonexistent", nil)
	w := httptest.NewRecorder()

	dash.handleReplay(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestHandleReplay_MissingID(t *testing.T) {
	dash, cleanup := setupHandlerTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/replay", nil)
	w := httptest.NewRecorder()

	dash.handleReplay(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestHandleStats(t *testing.T) {
	dash, cleanup := setupHandlerTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	w := httptest.NewRecorder()

	dash.handleStats(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var stats DashboardStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Stats should be initialized even if empty
	if stats.Topics == nil {
		t.Error("Expected topics to be initialized")
	}

	if stats.Subscriptions == nil {
		t.Error("Expected subscriptions to be initialized")
	}
}

func TestRegisterRoutes(t *testing.T) {
	dash, cleanup := setupHandlerTest(t)
	defer cleanup()

	mux := http.NewServeMux()
	dash.RegisterRoutes(mux)

	// Test that routes are registered
	routes := []string{
		"/api/stats",
		"/api/messages",
		"/api/messages/search",
		"/api/topics",
		"/api/subscriptions",
		"/api/publish",
		"/api/replay",
		"/api/health",
		"/",
	}

	for _, route := range routes {
		req := httptest.NewRequest(http.MethodGet, route, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		// Just verify that the route exists (not 404)
		if w.Code == http.StatusNotFound {
			t.Errorf("Route %s not registered", route)
		}
	}
}
