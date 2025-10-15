package pubsub

import "time"

// MessageInfo represents a Pub/Sub message with metadata
type MessageInfo struct {
	ID          string            `json:"id"`
	Data        string            `json:"data"`
	Attributes  map[string]string `json:"attributes"`
	PublishTime time.Time         `json:"publishTime"`
	Topic       string            `json:"topic"`
	Received    time.Time         `json:"received"`
}
