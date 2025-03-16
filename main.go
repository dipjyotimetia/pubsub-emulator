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
	client, err := pubsub.NewClient(ctx, cfg.ProjectID)
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

	// Subscribe and receive messages from the subscription with a timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := subscribeAndReceiveMessages(ctxWithTimeout, client, cfg); err != nil {
		log.Fatalf("Failed to subscribe and receive messages: %v", err)
	}
}

func createTopicSubscription(ctx context.Context, client *pubsub.Client, cfg Config) error {
	topics := strings.Split(cfg.TopicIDs, ",")
	subscriptions := strings.Split(cfg.SubIDs, ",")

	if len(topics) != len(subscriptions) {
		return fmt.Errorf("number of topics and subscriptions are not the same")
	}

	for i := range topics {
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
	topics := strings.SplitSeq(cfg.TopicIDs, ",")

	for topicID := range topics {
		topic := client.Topic(topicID)
		ok, err := topic.Exists(ctx)
		if err != nil {
			return fmt.Errorf("failed to check if topic exists: %v", err)
		}
		if !ok {
			return fmt.Errorf("topic %v does not exist", topicID)
		}

		// Publish a message to the topic
		result := topic.Publish(ctx, &pubsub.Message{
			Data: []byte(cfg.MessageToPublish),
		})

		// Get the message ID to confirm successful publishing
		msgID, err := result.Get(ctx)
		if err != nil {
			return fmt.Errorf("failed to publish message: %v", err)
		}
		fmt.Printf("Published message with ID: %s\n", msgID)
	}
	return nil
}

func subscribeAndReceiveMessages(ctx context.Context, client *pubsub.Client, cfg Config) error {
	subscriptions := strings.Split(cfg.SubIDs, ",")
	errChan := make(chan error, len(subscriptions))
	msgChan := make(chan *pubsub.Message, len(subscriptions))

	// Set up receivers for all subscriptions
	for _, subName := range subscriptions {
		sub := client.Subscription(subName)

		// Make subscription non-blocking
		go func(subName string, sub *pubsub.Subscription) {
			err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
				msgChan <- msg
				msg.Ack()
			})
			if err != nil {
				errChan <- fmt.Errorf("error receiving from subscription %s: %v", subName, err)
			}
		}(subName, sub)
	}

	// Wait for messages or context cancellation
	select {
	case msg := <-msgChan:
		fmt.Printf("Received message: %s\n", string(msg.Data))
		return nil
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
