package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all application configuration
type Config struct {
	ProjectID        string
	TopicIDs         []string
	SubscriptionIDs  []string
	MessageToPublish string
	DashboardPort    string
	PubSubPort       string
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	projectID := os.Getenv("PUBSUB_PROJECT")
	topicsStr := os.Getenv("PUBSUB_TOPIC")
	subsStr := os.Getenv("PUBSUB_SUBSCRIPTION")

	if projectID == "" || topicsStr == "" || subsStr == "" {
		return nil, fmt.Errorf("required environment variables PUBSUB_PROJECT, PUBSUB_TOPIC, or PUBSUB_SUBSCRIPTION are not set")
	}

	topics := parseCommaSeparated(topicsStr)
	subs := parseCommaSeparated(subsStr)

	if len(topics) != len(subs) {
		return nil, fmt.Errorf("number of topics (%d) and subscriptions (%d) must match", len(topics), len(subs))
	}

	return &Config{
		ProjectID:        projectID,
		TopicIDs:         topics,
		SubscriptionIDs:  subs,
		MessageToPublish: "Hello, Pub/Sub emulator!",
		DashboardPort:    getEnvOrDefault("DASHBOARD_PORT", ""),
		PubSubPort:       getEnvOrDefault("PUBSUB_PORT", "8085"),
	}, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.ProjectID == "" {
		return fmt.Errorf("project ID cannot be empty")
	}
	if len(c.TopicIDs) == 0 {
		return fmt.Errorf("at least one topic must be specified")
	}
	if len(c.SubscriptionIDs) == 0 {
		return fmt.Errorf("at least one subscription must be specified")
	}
	if len(c.TopicIDs) != len(c.SubscriptionIDs) {
		return fmt.Errorf("number of topics and subscriptions must match")
	}
	return nil
}

// IsDashboardEnabled returns true if dashboard should be started
func (c *Config) IsDashboardEnabled() bool {
	return c.DashboardPort != ""
}

// parseCommaSeparated splits a comma-separated string and trims whitespace
func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
