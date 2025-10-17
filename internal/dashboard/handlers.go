package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"github.com/dipjyotimetia/pubsub-emulator/internal/web"
)

// handleSearchMessages searches and filters messages based on query parameters
func (d *Dashboard) handleSearchMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	searchTerm := strings.ToLower(query.Get("q"))
	topicFilter := query.Get("topic")

	d.messagesMutex.RLock()
	defer d.messagesMutex.RUnlock()

	filtered := make([]MessageInfo, 0)
	for _, msg := range d.messages {
		if topicFilter != "" && msg.Topic != topicFilter {
			continue
		}

		if searchTerm != "" {
			dataLower := strings.ToLower(msg.Data)
			idLower := strings.ToLower(msg.ID)
			if !strings.Contains(dataLower, searchTerm) && !strings.Contains(idLower, searchTerm) {
				continue
			}
		}

		filtered = append(filtered, msg)
	}

	d.log.With("search_term", searchTerm, "topic_filter", topicFilter, "results_count", len(filtered)).
		Info("Message search completed")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(filtered); err != nil {
		d.log.Error("Failed to encode search results: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleCreateTopic creates a new Pub/Sub topic
func (d *Dashboard) handleCreateTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate Content-Type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" && contentType != "" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req CreateTopicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TopicID == "" {
		http.Error(w, "Topic ID is required", http.StatusBadRequest)
		return
	}
	if len(req.TopicID) > 255 {
		http.Error(w, "Topic ID too long (max 255 characters)", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	topic, err := d.client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{
		Name: fmt.Sprintf("projects/%s/topics/%s", d.projectID, req.TopicID),
	})
	if err != nil {
		d.log.With("topic_id", req.TopicID, "error", err.Error()).
			Error("Failed to create topic")
		http.Error(w, fmt.Sprintf("Failed to create topic: %v", err), http.StatusInternalServerError)
		return
	}

	d.log.With("topic_id", req.TopicID, "topic_name", topic.Name).
		Info("Topic created successfully")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"topic":  topic.Name,
	}); err != nil {
		d.log.Error("Failed to encode create topic response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleCreateSubscription creates a new Pub/Sub subscription
func (d *Dashboard) handleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate Content-Type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" && contentType != "" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate subscription ID
	if req.SubscriptionID == "" {
		http.Error(w, "Subscription ID is required", http.StatusBadRequest)
		return
	}
	if len(req.SubscriptionID) > 255 {
		http.Error(w, "Subscription ID too long (max 255 characters)", http.StatusBadRequest)
		return
	}

	if req.TopicID == "" {
		http.Error(w, "Topic ID is required", http.StatusBadRequest)
		return
	}

	if req.AckDeadlineSeconds <= 0 {
		req.AckDeadlineSeconds = 10 // Default 10 seconds
	}
	if req.AckDeadlineSeconds > 600 {
		http.Error(w, "Ack deadline too long (max 600 seconds)", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	sub, err := d.client.SubscriptionAdminClient.CreateSubscription(ctx, &pubsubpb.Subscription{
		Name:               fmt.Sprintf("projects/%s/subscriptions/%s", d.projectID, req.SubscriptionID),
		Topic:              fmt.Sprintf("projects/%s/topics/%s", d.projectID, req.TopicID),
		AckDeadlineSeconds: req.AckDeadlineSeconds,
	})
	if err != nil {
		d.log.With("subscription_id", req.SubscriptionID, "topic_id", req.TopicID, "error", err.Error()).
			Error("Failed to create subscription")
		http.Error(w, fmt.Sprintf("Failed to create subscription: %v", err), http.StatusInternalServerError)
		return
	}

	d.log.With("subscription_id", req.SubscriptionID, "subscription_name", sub.Name, "topic_id", req.TopicID, "ack_deadline", req.AckDeadlineSeconds).
		Info("Subscription created successfully")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":       "success",
		"subscription": sub.Name,
	}); err != nil {
		d.log.Error("Failed to encode create subscription response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handlePublish publishes a message to a topic via API
func (d *Dashboard) handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate Content-Type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" && contentType != "" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 11<<20)

	var req PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TopicID == "" {
		http.Error(w, "Topic ID is required", http.StatusBadRequest)
		return
	}

	if req.Data == "" {
		http.Error(w, "Message data is required", http.StatusBadRequest)
		return
	}
	if len(req.Data) > 10*1024*1024 { // 10MB limit
		http.Error(w, "Message data too large (max 10MB)", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	publisher := d.client.Publisher(req.TopicID)

	msg := &pubsub.Message{
		Data:       []byte(req.Data),
		Attributes: req.Attributes,
	}

	result := publisher.Publish(ctx, msg)
	msgID, err := result.Get(ctx)
	publisher.Stop()

	if err != nil {
		d.log.With("topic_id", req.TopicID, "data_size", len(req.Data), "error", err.Error()).
			Error("Failed to publish message")
		http.Error(w, fmt.Sprintf("Failed to publish message: %v", err), http.StatusInternalServerError)
		return
	}

	msg.ID = msgID
	msg.PublishTime = time.Now()
	d.AddMessage(msg, req.TopicID)

	d.log.With("topic_id", req.TopicID, "message_id", msgID, "data_size", len(req.Data)).
		Info("Message published successfully")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":    "success",
		"messageId": msgID,
	}); err != nil {
		d.log.Error("Failed to encode publish response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleReplay replays a historical message by publishing it again
func (d *Dashboard) handleReplay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	messageID := r.URL.Query().Get("id")
	if messageID == "" {
		http.Error(w, "Message ID required", http.StatusBadRequest)
		return
	}

	d.messagesMutex.RLock()
	var originalMsg MessageInfo
	var found bool
	for i := range d.messages {
		if d.messages[i].ID == messageID {
			originalMsg = d.messages[i] // Copy the value
			found = true
			break
		}
	}
	d.messagesMutex.RUnlock()

	if !found {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	ctx := r.Context()
	publisher := d.client.Publisher(originalMsg.Topic)

	msg := &pubsub.Message{
		Data:       []byte(originalMsg.Data),
		Attributes: originalMsg.Attributes,
	}

	result := publisher.Publish(ctx, msg)
	msgID, err := result.Get(ctx)
	publisher.Stop()

	if err != nil {
		d.log.With("original_message_id", messageID, "topic", originalMsg.Topic, "error", err.Error()).
			Error("Failed to replay message")
		http.Error(w, fmt.Sprintf("Failed to replay message: %v", err), http.StatusInternalServerError)
		return
	}

	msg.ID = msgID
	msg.PublishTime = time.Now()
	d.AddMessage(msg, originalMsg.Topic)

	d.log.With("original_message_id", messageID, "new_message_id", msgID, "topic", originalMsg.Topic).
		Info("Message replayed successfully")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":     "success",
		"messageId":  msgID,
		"originalId": messageID,
	}); err != nil {
		d.log.Error("Failed to encode replay response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleStats returns dashboard statistics including topics, subscriptions, and message counts
func (d *Dashboard) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	stats, err := d.GetStats(ctx)
	if err != nil {
		d.log.With("error", err.Error()).
			Error("Failed to get stats")
		http.Error(w, fmt.Sprintf("Error getting stats: %v", err), http.StatusInternalServerError)
		return
	}

	d.log.With("topic_count", stats.TopicCount, "subscription_count", stats.SubCount, "message_count", stats.MessageCount).
		Info("Stats retrieved successfully")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		d.log.Error("Failed to encode stats response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleMessages returns all stored messages
func (d *Dashboard) handleMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	d.messagesMutex.RLock()
	messages := make([]MessageInfo, len(d.messages))
	copy(messages, d.messages)
	d.messagesMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(messages); err != nil {
		d.log.Error("Failed to encode messages response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleHealth returns health check status
func (d *Dashboard) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}); err != nil {
		d.log.Error("Failed to encode health response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// handleIndex serves the main dashboard HTML page
func (d *Dashboard) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := web.DashboardTemplate.Execute(w, nil); err != nil {
		d.log.Error("Failed to render template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// RegisterRoutes registers all HTTP handlers to the provided mux
func (d *Dashboard) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/stats", d.handleStats)
	mux.HandleFunc("/api/messages", d.handleMessages)
	mux.HandleFunc("/api/messages/search", d.handleSearchMessages)
	mux.HandleFunc("/api/topics", d.handleCreateTopic)
	mux.HandleFunc("/api/subscriptions", d.handleCreateSubscription)
	mux.HandleFunc("/api/publish", d.handlePublish)
	mux.HandleFunc("/api/replay", d.handleReplay)
	mux.HandleFunc("/api/health", d.handleHealth)

	mux.HandleFunc("/static/css/dashboard.css", web.ServeDashboardCSS)
	mux.HandleFunc("/static/js/dashboard.js", web.ServeDashboardJS)

	mux.HandleFunc("/", d.handleIndex)
}

// StartHTTPServer starts the HTTP server with all routes registered
func (d *Dashboard) StartHTTPServer(port string) error {
	mux := http.NewServeMux()
	d.RegisterRoutes(mux)

	d.log.Info("Starting dashboard server on port %s", port)
	return http.ListenAndServe(":"+port, CORSMiddleware(mux))
}
