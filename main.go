package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
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
		t, err := client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{
			Name: fmt.Sprintf("projects/%s/topics/%s", cfg.ProjectID, topics[i]),
		})
		if err != nil {
			log.Printf("Failed to create topic: %v", err)
			continue
		}
		log.Printf("Topic created: %v\n", t.Name)

		sub, err := client.SubscriptionAdminClient.CreateSubscription(ctx, &pubsubpb.Subscription{
			Name:               fmt.Sprintf("projects/%s/subscriptions/%s", cfg.ProjectID, subscriptions[i]),
			Topic:              t.Name,
			AckDeadlineSeconds: 20,
		})
		if err != nil {
			log.Printf("Failed to create subscription: %v", err)
			continue
		}
		log.Printf("Created subscription: %v\n", sub.Name)
	}
	return nil
}

func publishMessage(ctx context.Context, client *pubsub.Client, cfg Config) error {
	topics := strings.SplitSeq(cfg.TopicIDs, ",")

	for topicID := range topics {
		publisher := client.Publisher(topicID)

		// Publish a message to the topic
		result := publisher.Publish(ctx, &pubsub.Message{
			Data: []byte(cfg.MessageToPublish),
		})

		// Get the message ID to confirm successful publishing
		msgID, err := result.Get(ctx)
		if err != nil {
			publisher.Stop()
			return fmt.Errorf("failed to publish message to topic %s: %v", topicID, err)
		}
		fmt.Printf("Published message to %s with ID: %s\n", topicID, msgID)
		publisher.Stop()
	}
	return nil
}

func subscribeAndReceiveMessages(ctx context.Context, client *pubsub.Client, cfg Config) error {
	subscriptions := strings.Split(cfg.SubIDs, ",")
	errChan := make(chan error, len(subscriptions))
	msgChan := make(chan *pubsub.Message, len(subscriptions))

	// Set up receivers for all subscriptions
	for _, subName := range subscriptions {
		sub := client.Subscriber(subName)

		// Make subscription non-blocking
		go func(subName string, sub *pubsub.Subscriber) {
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
