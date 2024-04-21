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

type Config struct {
	ProjectID        string
	TopicIDs         string
	SubIDs           string
	MessageToPublish string
}

func main() {
	cfg := Config{
		ProjectID:        os.Getenv("PUBSUB_PROJECT"),
		TopicIDs:         os.Getenv("PUBSUB_TOPIC"),
		SubIDs:           os.Getenv("PUBSUB_SUBSCRIPTION"),
		MessageToPublish: "Hello, Pub/Sub emulator!",
	}

	if cfg.ProjectID == "" || cfg.TopicIDs == "" || cfg.SubIDs == "" {
		log.Fatal("Environment variables PUBSUB_PROJECT, PUBSUB_TOPIC, or PUBSUB_SUBSCRIPTION are not set")
	}

	ctx := context.Background()
	client, err := createClient(ctx, cfg.ProjectID)
	if err != nil {
		log.Fatalf("Failed to create pubsub client: %v", err)
	}
	defer client.Close()

	// Create topics and subscriptions
	if err := createTopicSubscription(ctx, client, cfg); err != nil {
		log.Fatalf("Failed to create topics and subscriptions: %v", err)
	}

	// Publish a message to the topic
	if err := publishMessage(ctx, client, cfg); err != nil {
		log.Fatalf("Failed to publish messages: %v", err)
	}

	// Subscribe and receive messages from the subscription
	if err := subscribeAndReceiveMessages(ctx, client, cfg); err != nil {
		log.Fatalf("Failed to subscribe and receive messages: %v", err)
	}
}

func createClient(ctx context.Context, projectID string) (*pubsub.Client, error) {
	return pubsub.NewClient(ctx, projectID)
}

func createTopicSubscription(ctx context.Context, client *pubsub.Client, cfg Config) error {
	topics := strings.Split(cfg.TopicIDs, ",")
	subscriptions := strings.Split(cfg.SubIDs, ",")

	if len(topics) != len(subscriptions) {
		return fmt.Errorf("number of topics and subscriptions are not the same")
	}

	client, err := pubsub.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		log.Fatalf("Failed to create pubsub client: %v", err)
		return fmt.Errorf("Failed to create pubsub client: %v", err)
	}
	defer client.Close()

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
	return nil
}

func publishMessage(ctx context.Context, client *pubsub.Client, cfg Config) error {
	topics := strings.Split(cfg.TopicIDs, ",")
	client, err := pubsub.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		log.Fatalf("Failed to create pubsub client: %v", err)
		return err
	}
	defer client.Close()

	for _, topicId := range topics {
		topic := client.Topic(topicId)
		ok, err := topic.Exists(ctx)
		if err != nil {
			log.Fatalf("Failed to check if topic exists: %v", err)
			return err
		}
		if !ok {
			log.Fatalf("Topic %v does not exist", topicId)
			return err
		}
		// Publish a message to the topic
		result := topic.Publish(ctx, &pubsub.Message{
			Data: []byte(cfg.MessageToPublish),
		})

		// Get the message ID to confirm successful publishing
		msgID, err := result.Get(ctx)
		if err != nil {
			log.Fatalf("Failed to publish message: %v", err)
			return err
		}
		fmt.Printf("Published message with ID: %s\n", msgID)
	}
	return nil
}

func subscribeAndReceiveMessages(ctx context.Context, client *pubsub.Client, cfg Config) error {
	subscriptions := strings.Split(cfg.SubIDs, ",")

	client, err := pubsub.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		log.Fatalf("Failed to create pubsub client: %v", err)
		return err
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
	return nil
}
