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
	ctx := context.Background()
	createTopicSubscription(ctx)
}

func createTopicSubscription(ctx context.Context) {
	topics := strings.Split(topicIDS, ",")
	subscriptions := strings.Split(subIDS, ",")

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		_ = fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	if len(topics) == len(subscriptions) {
		for i := 0; i < len(topics); i++ {
			t, err := client.CreateTopic(ctx, topics[i])
			if err != nil {
				_ = fmt.Errorf("failed to create topic: %v", err)
			}
			log.Printf("topic created: %v\n", t)
			sub, err := client.CreateSubscription(ctx, subscriptions[i], pubsub.SubscriptionConfig{
				Topic:       client.Topic(topics[i]),
				AckDeadline: 20 * time.Second,
			})
			if err != nil {
				_ = fmt.Errorf("failed to create subscription: %v", err)
			}
			log.Printf("created subscription: %v\n", sub)
		}
	} else {
		_ = fmt.Errorf("number of topic and subscription are not same")
	}
}
