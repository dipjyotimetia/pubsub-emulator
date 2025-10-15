package main

import (
	"context"
	"fmt"
	"time"

	gcppubsub "cloud.google.com/go/pubsub/v2"
	"github.com/dipjyotimetia/pubsub-emulator/internal/config"
	"github.com/dipjyotimetia/pubsub-emulator/internal/dashboard"
	"github.com/dipjyotimetia/pubsub-emulator/internal/pubsub"
	"github.com/dipjyotimetia/pubsub-emulator/internal/server"
	"github.com/dipjyotimetia/pubsub-emulator/pkg/logger"
)

func main() {
	// Initialize logger
	log := logger.New()
	log.Info("Starting Pub/Sub Emulator with refactored architecture")

	// Load configuration
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatal("Failed to load configuration: %v", err)
	}

	log.Info("Configuration loaded: Project=%s, Topics=%v, Subscriptions=%v",
		cfg.ProjectID, cfg.TopicIDs, cfg.SubscriptionIDs)

	// Create context
	ctx := context.Background()

	// Initialize Pub/Sub client wrapper (creates GCP client internally)
	psClient, err := pubsub.NewClient(ctx, cfg.ProjectID, log)
	if err != nil {
		log.Fatal("Failed to create Pub/Sub client: %v", err)
	}
	defer psClient.Close()

	// Get underlying GCP client for dashboard
	pubsubClient := psClient.GetClient()

	// Create topics and subscriptions
	if err := setupTopicsAndSubscriptions(ctx, psClient, cfg, log); err != nil {
		log.Fatal("Failed to setup topics and subscriptions: %v", err)
	}

	// Initialize dashboard
	dash := dashboard.New(pubsubClient, cfg.ProjectID, log)

	// Initialize publisher
	pub := pubsub.NewPublisher(psClient, log)

	// Publish initial messages to topics
	if err := publishInitialMessages(ctx, pub, cfg, dash, log); err != nil {
		log.Error("Failed to publish initial messages: %v", err)
	}

	// Initialize subscriber
	sub := pubsub.NewSubscriber(psClient, log)

	// Start receiving messages from subscriptions
	startSubscriptions(ctx, sub, cfg, dash, log)

	// Initialize and start HTTP server with graceful shutdown
	srv := server.New(&server.Config{
		Port:      cfg.DashboardPort,
		Dashboard: dash,
		Logger:    log,
	})

	log.Info("Dashboard will be available at http://localhost:%s", cfg.DashboardPort)

	// Start server (blocks until shutdown signal)
	if err := srv.Start(); err != nil {
		log.Fatal("Server error: %v", err)
	}

	log.Info("Application shutdown complete")
}

// setupTopicsAndSubscriptions creates topics and subscriptions
func setupTopicsAndSubscriptions(ctx context.Context, psClient *pubsub.Client, cfg *config.Config, log *logger.Logger) error {
	if len(cfg.TopicIDs) != len(cfg.SubscriptionIDs) {
		return fmt.Errorf("number of topics (%d) and subscriptions (%d) must match",
			len(cfg.TopicIDs), len(cfg.SubscriptionIDs))
	}

	for i, topicID := range cfg.TopicIDs {
		// Create topic
		topic, err := psClient.CreateTopic(ctx, topicID)
		if err != nil {
			log.Warn("Failed to create topic %s: %v (may already exist)", topicID, err)
		} else {
			log.Info("Created topic: %s", topic.Name)
		}

		// Create subscription
		subID := cfg.SubscriptionIDs[i]
		subscription, err := psClient.CreateSubscription(ctx, subID, topicID, 20)
		if err != nil {
			log.Warn("Failed to create subscription %s: %v (may already exist)", subID, err)
		} else {
			log.Info("Created subscription: %s", subscription.Name)
		}
	}

	return nil
}

// publishInitialMessages publishes messages to all configured topics
func publishInitialMessages(ctx context.Context, pub *pubsub.Publisher, cfg *config.Config, dash *dashboard.Dashboard, log *logger.Logger) error {
	message := cfg.MessageToPublish
	if message == "" {
		message = "Hello from Pub/Sub Emulator!"
	}

	attributes := map[string]string{
		"source": "emulator",
		"time":   time.Now().Format(time.RFC3339),
	}

	log.Info("Publishing initial message to %d topics", len(cfg.TopicIDs))

	for _, topicID := range cfg.TopicIDs {
		msgID, err := pub.PublishMessage(ctx, topicID, message, attributes)
		if err != nil {
			log.Error("Failed to publish to topic %s: %v", topicID, err)
			continue
		}

		log.Info("Published message to %s with ID: %s", topicID, msgID)

		// Add to dashboard
		msg := &gcppubsub.Message{
			ID:          msgID,
			Data:        []byte(message),
			Attributes:  attributes,
			PublishTime: time.Now(),
		}
		dash.AddMessage(msg, topicID)
	}

	return nil
}

// startSubscriptions starts message receivers for all subscriptions
func startSubscriptions(ctx context.Context, sub *pubsub.Subscriber, cfg *config.Config, dash *dashboard.Dashboard, log *logger.Logger) {
	log.Info("Starting message receivers for %d subscriptions", len(cfg.SubscriptionIDs))

	// Create handler function that adds messages to dashboard
	handler := func(ctx context.Context, msg *gcppubsub.Message) {
		// Find the topic for this subscription
		topicID := "" // Would need mapping

		// For now, use the first topic as default
		if len(cfg.TopicIDs) > 0 {
			topicID = cfg.TopicIDs[0]
		}

		log.Debug("Received message: %s from topic: %s", msg.ID, topicID)
		dash.AddMessage(msg, topicID)
	}

	// Subscribe to all subscriptions
	sub.SubscribeToAll(ctx, cfg.SubscriptionIDs, cfg.TopicIDs, handler)
}
