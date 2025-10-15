package dashboard

import (
	"encoding/csv"
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

	var filtered []MessageInfo
	for _, msg := range d.messages {
		// Apply topic filter
		if topicFilter != "" && msg.Topic != topicFilter {
			continue
		}

		// Apply search term filter
		if searchTerm != "" {
			dataLower := strings.ToLower(msg.Data)
			idLower := strings.ToLower(msg.ID)
			if !strings.Contains(dataLower, searchTerm) && !strings.Contains(idLower, searchTerm) {
				continue
			}
		}

		filtered = append(filtered, msg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filtered)
}

// handleExportJSON exports all messages as JSON file
func (d *Dashboard) handleExportJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	d.messagesMutex.RLock()
	messages := make([]MessageInfo, len(d.messages))
	copy(messages, d.messages)
	d.messagesMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=messages.json")
	json.NewEncoder(w).Encode(messages)
}

// handleExportCSV exports all messages as CSV file
func (d *Dashboard) handleExportCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	d.messagesMutex.RLock()
	messages := make([]MessageInfo, len(d.messages))
	copy(messages, d.messages)
	d.messagesMutex.RUnlock()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=messages.csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"ID", "Data", "Topic", "PublishTime", "ReceivedTime"})

	// Write messages
	for _, msg := range messages {
		writer.Write([]string{
			msg.ID,
			msg.Data,
			msg.Topic,
			msg.PublishTime.Format(time.RFC3339),
			msg.Received.Format(time.RFC3339),
		})
	}
}

// handleCreateTopic creates a new Pub/Sub topic
func (d *Dashboard) handleCreateTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateTopicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	topic, err := d.client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{
		Name: fmt.Sprintf("projects/%s/topics/%s", d.projectID, req.TopicID),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create topic: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"topic":  topic.Name,
	})
}

// handleCreateSubscription creates a new Pub/Sub subscription
func (d *Dashboard) handleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AckDeadlineSeconds <= 0 {
		req.AckDeadlineSeconds = 10 // Default 10 seconds
	}

	ctx := r.Context()
	sub, err := d.client.SubscriptionAdminClient.CreateSubscription(ctx, &pubsubpb.Subscription{
		Name:               fmt.Sprintf("projects/%s/subscriptions/%s", d.projectID, req.SubscriptionID),
		Topic:              fmt.Sprintf("projects/%s/topics/%s", d.projectID, req.TopicID),
		AckDeadlineSeconds: req.AckDeadlineSeconds,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create subscription: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":       "success",
		"subscription": sub.Name,
	})
}

// handlePublish publishes a message to a topic via API
func (d *Dashboard) handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	publisher := d.client.Publisher(req.TopicID)
	defer publisher.Stop()

	msg := &pubsub.Message{
		Data:       []byte(req.Data),
		Attributes: req.Attributes,
	}

	result := publisher.Publish(ctx, msg)
	msgID, err := result.Get(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to publish message: %v", err), http.StatusInternalServerError)
		return
	}

	// Add to dashboard
	msg.ID = msgID
	msg.PublishTime = time.Now()
	d.AddMessage(msg, req.TopicID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "success",
		"messageId": msgID,
	})
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
	var originalMsg *MessageInfo
	for i := range d.messages {
		if d.messages[i].ID == messageID {
			originalMsg = &d.messages[i]
			break
		}
	}
	d.messagesMutex.RUnlock()

	if originalMsg == nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	ctx := r.Context()
	publisher := d.client.Publisher(originalMsg.Topic)
	defer publisher.Stop()

	msg := &pubsub.Message{
		Data:       []byte(originalMsg.Data),
		Attributes: originalMsg.Attributes,
	}

	result := publisher.Publish(ctx, msg)
	msgID, err := result.Get(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to replay message: %v", err), http.StatusInternalServerError)
		return
	}

	// Add to dashboard
	msg.ID = msgID
	msg.PublishTime = time.Now()
	d.AddMessage(msg, originalMsg.Topic)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":     "success",
		"messageId":  msgID,
		"originalId": messageID,
	})
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
		http.Error(w, fmt.Sprintf("Error getting stats: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
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
	json.NewEncoder(w).Encode(messages)
}

// handleHealth returns health check status
func (d *Dashboard) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
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
	// API endpoints
	mux.HandleFunc("/api/stats", d.handleStats)
	mux.HandleFunc("/api/messages", d.handleMessages)
	mux.HandleFunc("/api/messages/search", d.handleSearchMessages)
	mux.HandleFunc("/api/messages/export/json", d.handleExportJSON)
	mux.HandleFunc("/api/messages/export/csv", d.handleExportCSV)
	mux.HandleFunc("/api/topics", d.handleCreateTopic)
	mux.HandleFunc("/api/subscriptions", d.handleCreateSubscription)
	mux.HandleFunc("/api/publish", d.handlePublish)
	mux.HandleFunc("/api/replay", d.handleReplay)
	mux.HandleFunc("/api/health", d.handleHealth)
	mux.HandleFunc("/ws", d.HandleWebSocket)

	// Serve static files
	mux.HandleFunc("/static/css/dashboard.css", web.ServeDashboardCSS)
	mux.HandleFunc("/static/js/dashboard.js", web.ServeDashboardJS)

	// Serve index page
	mux.HandleFunc("/", d.handleIndex)
}

// StartHTTPServer starts the HTTP server with all routes registered
func (d *Dashboard) StartHTTPServer(port string) error {
	mux := http.NewServeMux()
	d.RegisterRoutes(mux)

	d.log.Info("Starting dashboard server on port %s", port)
	return http.ListenAndServe(":"+port, CORSMiddleware(mux))
}
