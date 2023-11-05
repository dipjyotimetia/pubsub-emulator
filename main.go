package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
)

var (
	projectID        = os.Getenv("PUBSUB_PROJECT")
	topicIDS         = os.Getenv("PUBSUB_TOPIC")
	subIDS           = os.Getenv("PUBSUB_SUBSCRIPTION")
	messageToPublish = "Hello, Pub/Sub emulator!"
)

func main() {
	if projectID == "" || topicIDS == "" || subIDS == "" {
		log.Fatal("Environment variables PUBSUB_PROJECT, PUBSUB_TOPIC, or PUBSUB_SUBSCRIPTION are not set")
	}

	ctx := context.Background()
	// Create topics and subscriptions
	createTopicSubscription(ctx)

	// Publish a message to the topic
	publishMessage(ctx)

	// Subscribe and receive messages from the subscription
	subscribeAndReceiveMessages(ctx)
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

func publishMessage(ctx context.Context) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create pubsub client: %v", err)
	}
	defer client.Close()

	topic := client.Topic(strings.Split(topicIDS, ",")[0]) // Assuming there's only one topic

	// Publish a message to the topic
	result := topic.Publish(ctx, &pubsub.Message{
		Data: []byte(messageToPublish),
	})

	// Get the message ID to confirm successful publishing
	msgID, err := result.Get(ctx)
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}

	fmt.Printf("Published message with ID: %s\n", msgID)
}

func subscribeAndReceiveMessages(ctx context.Context) {
	subscriptions := strings.Split(subIDS, ",")

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create pubsub client: %v", err)
	}
	defer client.Close()

	for _, subName := range subscriptions {
		sub := client.Subscription(subName)

		// Receive messages from the subscription
		err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			fmt.Printf("Received message: %s\n", string(msg.Data))

			// Acknowledge the message to mark it as processed
			msg.Ack()
		})
		if err != nil {
			log.Printf("Error receiving messages from subscription %s: %v", subName, err)
		}
	}
}
