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
	projectID = os.Getenv("PUBSUB_PROJECT")
	topicIDS  = os.Getenv("PUBSUB_TOPIC")
	subIDS    = os.Getenv("PUBSUB_SUBSCRIPTION")
)

func main() {
	if projectID == "" || topicIDS == "" || subIDS == "" {
		log.Fatal("Environment variables PUBSUB_PROJECT, PUBSUB_TOPIC, or PUBSUB_SUBSCRIPTION are not set")
	}

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create pubsub client: %v", err)
	}
	defer client.Close()

	// Create topics and subscriptions
	err = createTopicSubscription(ctx, client)
	if err != nil {
		log.Fatalf("Failed to create topics and subscriptions: %v", err)
	}

	// Publish a message to the topic
	err = publishMessage(ctx, client)
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}

	// Subscribe and receive messages from the subscription
	err = subscribeAndReceiveMessages(ctx, client)
	if err != nil {
		log.Fatalf("Failed to subscribe and receive messages: %v", err)
	}
}

func createTopicSubscription(ctx context.Context, client *pubsub.Client) error {
	topics := strings.Split(topicIDS, ",")
	subscriptions := strings.Split(subIDS, ",")

	if len(topics) != len(subscriptions) {
		return fmt.Errorf("number of topics and subscriptions are not the same")
	}

	for i, topic := range topics {
		t, err := client.CreateTopic(ctx, topic)
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

	return nil
}

func publishMessage(ctx context.Context, client *pubsub.Client) error {
	// Publish a message to the topic
	topic := client.Topic(strings.Split(topicIDS, ",")[0])
	result := topic.Publish(ctx, &pubsub.Message{
		Data: []byte("Hello, Pub/Sub"),
	})
	_, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	return nil
}

func subscribeAndReceiveMessages(ctx context.Context, client *pubsub.Client) error {
	// Subscribe and receive messages from the subscription
	sub := client.Subscription(subIDS)
	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		log.Printf("Received message: %s\n", string(msg.Data))
		msg.Ack()
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe and receive messages: %v", err)
	}

	return nil
}
