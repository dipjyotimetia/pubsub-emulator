package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
)

var (
	projectID = os.Getenv("PUBSUB_PROJECT")
	topicIDS  = os.Getenv("PUBSUB_TOPIC")
	subIDS    = os.Getenv("PUBSUB_SUBSCRIPTION")
)

func main() {
	if projectID == "" || topicIDS == "" || subIDS == "" {
		log.Fatal("Environment variables PUBSUB_PROJECT, PUBSUB_TOPIC, or PUBSUB_SUBSCRIPTION are not set")
	}

	ctx := context.Background()
	createTopicSubscription(ctx)
}

func createTopicSubscription(ctx context.Context) {
	topics := strings.Split(topicIDS, ",")
	subscriptions := strings.Split(subIDS, ",")

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create pubsub client: %v", err)
	}
	defer client.Close()

	if len(topics) != len(subscriptions) {
		log.Fatalf("Number of topics and subscriptions are not the same")
	}

	for i := 0; i < len(topics); i++ {
		t, err := client.CreateTopic(ctx, topics[i])
		if err != nil {
			log.Printf("Failed to create topic: %v", err)
			continue
		}
		log.Printf("Topic created: %v\n", t)

		sub, err := client.CreateSubscription(ctx, subscriptions[i], pubsub.SubscriptionConfig{
			Topic:       t,
			AckDeadline: 20 * time.Second,
		})
		if err != nil {
			log.Printf("Failed to create subscription: %v", err)
			continue
		}
		log.Printf("Created subscription: %v\n", sub)
	}
}
