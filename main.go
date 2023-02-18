package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
)

var (
	projectID    = os.Getenv("PUBSUB_PROJECT")
	topicID      = os.Getenv("PUBSUB_TOPIC")
	subscription = os.Getenv("PUBSUB_SUBSCRIPTION")
)

func main() {
	ctx := context.Background()
	createTopic(ctx)
	createSubscription(ctx)
}

func createTopic(ctx context.Context) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		_ = fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	t, err := client.CreateTopic(ctx, topicID)
	if err != nil {
		_ = fmt.Errorf("CreateTopic: %v", err)
	}
	fmt.Printf("Topic created: %v\n", t)
}

func createSubscription(ctx context.Context) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		_ = fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	sub, err := client.CreateSubscription(ctx, subscription, pubsub.SubscriptionConfig{
		Topic:       client.Topic(topicID),
		AckDeadline: 20 * time.Second,
	})
	if err != nil {
		_ = fmt.Errorf("CreateSubscription: %v", err)
	}
	fmt.Printf("Created subscription: %v\n", sub)
}
