package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"github.com/dipjyotimetia/pubsub-emulator/pkg/logger"
	"github.com/gorilla/websocket"
)

// Dashboard manages the dashboard state and operations
type Dashboard struct {
	client        *pubsub.Client
	projectID     string
	messages      []MessageInfo
	messagesMutex sync.RWMutex
	maxMessages   int
	wsClients     map[*websocket.Conn]bool
	wsClientsMux  sync.RWMutex
	broadcast     chan []byte
	log           *logger.Logger
}

// New creates a new Dashboard instance
func New(client *pubsub.Client, projectID string, log *logger.Logger) *Dashboard {
	d := &Dashboard{
		client:      client,
		projectID:   projectID,
		messages:    make([]MessageInfo, 0),
		maxMessages: 1000,
		wsClients:   make(map[*websocket.Conn]bool),
		broadcast:   make(chan []byte, 256),
		log:         log,
	}

	// Start WebSocket broadcast handler
	go d.handleBroadcast()

	return d
}

// AddMessage adds a message to the dashboard
func (d *Dashboard) AddMessage(msg *pubsub.Message, topic string) {
	d.messagesMutex.Lock()
	defer d.messagesMutex.Unlock()

	msgInfo := MessageInfo{
		ID:          msg.ID,
		Data:        string(msg.Data),
		Attributes:  msg.Attributes,
		PublishTime: msg.PublishTime,
		Topic:       topic,
		Received:    msg.PublishTime,
	}

	d.messages = append(d.messages, msgInfo)

	// Keep only the last maxMessages
	if len(d.messages) > d.maxMessages {
		d.messages = d.messages[len(d.messages)-d.maxMessages:]
	}

	// Broadcast new message to WebSocket clients
	msgJSON, _ := json.Marshal(map[string]any{
		"type":    "new_message",
		"message": msgInfo,
	})
	select {
	case d.broadcast <- msgJSON:
	default:
		// Channel full, skip broadcast
	}
}

// GetStats retrieves dashboard statistics
func (d *Dashboard) GetStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{
		Topics:        make([]TopicInfo, 0),
		Subscriptions: make([]SubscriptionInfo, 0),
		TopicList:     make([]string, 0),
		SubscriptionList: make([]string, 0),
	}

	// List topics
	it := d.client.TopicAdminClient.ListTopics(ctx, &pubsubpb.ListTopicsRequest{
		Project: fmt.Sprintf("projects/%s", d.projectID),
	})

	for {
		topic, err := it.Next()
		if err != nil {
			break
		}
		topicID := extractID(topic.Name)
		stats.Topics = append(stats.Topics, TopicInfo{
			Name: topic.Name,
			ID:   topicID,
		})
		stats.TopicList = append(stats.TopicList, topicID)
	}

	// List subscriptions
	subIt := d.client.SubscriptionAdminClient.ListSubscriptions(ctx, &pubsubpb.ListSubscriptionsRequest{
		Project: fmt.Sprintf("projects/%s", d.projectID),
	})

	for {
		sub, err := subIt.Next()
		if err != nil {
			break
		}
		subID := extractID(sub.Name)
		stats.Subscriptions = append(stats.Subscriptions, SubscriptionInfo{
			Name:               sub.Name,
			ID:                 subID,
			Topic:              sub.Topic,
			AckDeadlineSeconds: sub.AckDeadlineSeconds,
		})
		stats.SubscriptionList = append(stats.SubscriptionList, subID)
	}

	// Get recent messages
	d.messagesMutex.RLock()
	stats.MessageCount = len(d.messages)
	stats.TotalMessages = len(d.messages)
	stats.TopicCount = len(stats.Topics)
	stats.SubCount = len(stats.Subscriptions)

	// Get last message time
	if len(d.messages) > 0 {
		lastMsg := d.messages[len(d.messages)-1]
		stats.LastMessageTime = &lastMsg.Received
	}

	// Return last 20 messages
	start := max(0, len(d.messages)-20)
	stats.RecentMessages = d.messages[start:]
	d.messagesMutex.RUnlock()

	return stats, nil
}

// GetMessages returns all messages
func (d *Dashboard) GetMessages() []MessageInfo {
	d.messagesMutex.RLock()
	defer d.messagesMutex.RUnlock()

	messages := make([]MessageInfo, len(d.messages))
	copy(messages, d.messages)
	return messages
}

// GetMessageByID finds a message by its ID
func (d *Dashboard) GetMessageByID(id string) *MessageInfo {
	d.messagesMutex.RLock()
	defer d.messagesMutex.RUnlock()

	for i := range d.messages {
		if d.messages[i].ID == id {
			return &d.messages[i]
		}
	}
	return nil
}

// RegisterWebSocketClient registers a new WebSocket client
func (d *Dashboard) RegisterWebSocketClient(conn *websocket.Conn) {
	d.wsClientsMux.Lock()
	d.wsClients[conn] = true
	d.wsClientsMux.Unlock()
}

// UnregisterWebSocketClient removes a WebSocket client
func (d *Dashboard) UnregisterWebSocketClient(conn *websocket.Conn) {
	d.wsClientsMux.Lock()
	delete(d.wsClients, conn)
	d.wsClientsMux.Unlock()
	conn.Close()
}

// handleBroadcast broadcasts messages to all connected WebSocket clients
func (d *Dashboard) handleBroadcast() {
	for message := range d.broadcast {
		d.wsClientsMux.RLock()
		for client := range d.wsClients {
			err := client.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				d.log.Error("WebSocket error: %v", err)
				client.Close()
				delete(d.wsClients, client)
			}
		}
		d.wsClientsMux.RUnlock()
	}
}

// extractID extracts the ID from a full resource name
func extractID(fullName string) string {
	for i := len(fullName) - 1; i >= 0; i-- {
		if fullName[i] == '/' {
			return fullName[i+1:]
		}
	}
	return fullName
}
