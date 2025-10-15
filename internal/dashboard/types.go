package dashboard

import (
	"time"
)

// MessageInfo represents a Pub/Sub message in the dashboard
type MessageInfo struct {
	ID          string            `json:"id"`
	Data        string            `json:"data"`
	Attributes  map[string]string `json:"attributes"`
	PublishTime time.Time         `json:"publish_time"`
	Topic       string            `json:"topic"`
	Received    time.Time         `json:"received"`
}

// TopicInfo represents topic information
type TopicInfo struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// SubscriptionInfo represents subscription information
type SubscriptionInfo struct {
	Name               string `json:"name"`
	ID                 string `json:"id"`
	Topic              string `json:"topic"`
	AckDeadlineSeconds int32  `json:"ackDeadlineSeconds"`
}

// DashboardStats contains statistics for the dashboard
type DashboardStats struct {
	Topics          []TopicInfo        `json:"topics"`
	Subscriptions   []SubscriptionInfo `json:"subscriptions"`
	MessageCount    int                `json:"messageCount"`
	RecentMessages  []MessageInfo      `json:"recentMessages"`
	TopicCount      int                `json:"topics"`
	SubCount        int                `json:"subscriptions"`
	TotalMessages   int                `json:"total_messages"`
	LastMessageTime *time.Time         `json:"last_message_time,omitempty"`
	TopicList       []string           `json:"topic_list"`
	SubscriptionList []string          `json:"subscription_list"`
}

// PublishRequest represents a request to publish a message
type PublishRequest struct {
	TopicID    string            `json:"topic_id"`
	Data       string            `json:"data"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// CreateTopicRequest represents a request to create a topic
type CreateTopicRequest struct {
	TopicID string `json:"topic_id"`
}

// CreateSubscriptionRequest represents a request to create a subscription
type CreateSubscriptionRequest struct {
	SubscriptionID     string `json:"subscription_id"`
	TopicID            string `json:"topic_id"`
	AckDeadlineSeconds int32  `json:"ack_deadline_seconds"`
}
