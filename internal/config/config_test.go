package config

import (
	"os"
	"testing"
)

func TestLoadFromEnv_Success(t *testing.T) {
	// Set up environment variables
	_ = os.Setenv("PUBSUB_PROJECT", "test-project")
	_ = os.Setenv("PUBSUB_TOPIC", "topic1,topic2")
	_ = os.Setenv("PUBSUB_SUBSCRIPTION", "sub1,sub2")
	_ = os.Setenv("DASHBOARD_PORT", "8080")
	_ = os.Setenv("PUBSUB_PORT", "8085")
	defer cleanupEnv()

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.ProjectID != "test-project" {
		t.Errorf("Expected ProjectID 'test-project', got '%s'", cfg.ProjectID)
	}

	if len(cfg.TopicIDs) != 2 {
		t.Errorf("Expected 2 topics, got %d", len(cfg.TopicIDs))
	}

	if cfg.TopicIDs[0] != "topic1" || cfg.TopicIDs[1] != "topic2" {
		t.Errorf("Expected topics [topic1, topic2], got %v", cfg.TopicIDs)
	}

	if len(cfg.SubscriptionIDs) != 2 {
		t.Errorf("Expected 2 subscriptions, got %d", len(cfg.SubscriptionIDs))
	}

	if cfg.SubscriptionIDs[0] != "sub1" || cfg.SubscriptionIDs[1] != "sub2" {
		t.Errorf("Expected subscriptions [sub1, sub2], got %v", cfg.SubscriptionIDs)
	}

	if cfg.DashboardPort != "8080" {
		t.Errorf("Expected DashboardPort '8080', got '%s'", cfg.DashboardPort)
	}

	if cfg.PubSubPort != "8085" {
		t.Errorf("Expected PubSubPort '8085', got '%s'", cfg.PubSubPort)
	}
}

func TestLoadFromEnv_MissingProjectID(t *testing.T) {
	os.Setenv("PUBSUB_TOPIC", "topic1")
	os.Setenv("PUBSUB_SUBSCRIPTION", "sub1")
	defer cleanupEnv()

	_, err := LoadFromEnv()
	if err == nil {
		t.Fatal("Expected error for missing PUBSUB_PROJECT, got nil")
	}
}

func TestLoadFromEnv_MissingTopic(t *testing.T) {
	os.Setenv("PUBSUB_PROJECT", "test-project")
	os.Setenv("PUBSUB_SUBSCRIPTION", "sub1")
	defer cleanupEnv()

	_, err := LoadFromEnv()
	if err == nil {
		t.Fatal("Expected error for missing PUBSUB_TOPIC, got nil")
	}
}

func TestLoadFromEnv_MissingSubscription(t *testing.T) {
	os.Setenv("PUBSUB_PROJECT", "test-project")
	os.Setenv("PUBSUB_TOPIC", "topic1")
	defer cleanupEnv()

	_, err := LoadFromEnv()
	if err == nil {
		t.Fatal("Expected error for missing PUBSUB_SUBSCRIPTION, got nil")
	}
}

func TestLoadFromEnv_MismatchedTopicsAndSubscriptions(t *testing.T) {
	os.Setenv("PUBSUB_PROJECT", "test-project")
	os.Setenv("PUBSUB_TOPIC", "topic1,topic2")
	os.Setenv("PUBSUB_SUBSCRIPTION", "sub1")
	defer cleanupEnv()

	_, err := LoadFromEnv()
	if err == nil {
		t.Fatal("Expected error for mismatched topics and subscriptions, got nil")
	}
}

func TestLoadFromEnv_DefaultValues(t *testing.T) {
	os.Setenv("PUBSUB_PROJECT", "test-project")
	os.Setenv("PUBSUB_TOPIC", "topic1")
	os.Setenv("PUBSUB_SUBSCRIPTION", "sub1")
	defer cleanupEnv()

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.DashboardPort != "" {
		t.Errorf("Expected default DashboardPort '', got '%s'", cfg.DashboardPort)
	}

	if cfg.PubSubPort != "8085" {
		t.Errorf("Expected default PubSubPort '8085', got '%s'", cfg.PubSubPort)
	}

	if cfg.MessageToPublish != "Hello, Pub/Sub emulator!" {
		t.Errorf("Expected default MessageToPublish, got '%s'", cfg.MessageToPublish)
	}
}

func TestValidate_Success(t *testing.T) {
	cfg := &Config{
		ProjectID:       "test-project",
		TopicIDs:        []string{"topic1"},
		SubscriptionIDs: []string{"sub1"},
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestValidate_EmptyProjectID(t *testing.T) {
	cfg := &Config{
		ProjectID:       "",
		TopicIDs:        []string{"topic1"},
		SubscriptionIDs: []string{"sub1"},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for empty ProjectID, got nil")
	}
}

func TestValidate_NoTopics(t *testing.T) {
	cfg := &Config{
		ProjectID:       "test-project",
		TopicIDs:        []string{},
		SubscriptionIDs: []string{"sub1"},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for no topics, got nil")
	}
}

func TestValidate_NoSubscriptions(t *testing.T) {
	cfg := &Config{
		ProjectID:       "test-project",
		TopicIDs:        []string{"topic1"},
		SubscriptionIDs: []string{},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for no subscriptions, got nil")
	}
}

func TestValidate_MismatchedCounts(t *testing.T) {
	cfg := &Config{
		ProjectID:       "test-project",
		TopicIDs:        []string{"topic1", "topic2"},
		SubscriptionIDs: []string{"sub1"},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for mismatched counts, got nil")
	}
}

func TestIsDashboardEnabled_True(t *testing.T) {
	cfg := &Config{
		DashboardPort: "8080",
	}

	if !cfg.IsDashboardEnabled() {
		t.Error("Expected IsDashboardEnabled to be true")
	}
}

func TestIsDashboardEnabled_False(t *testing.T) {
	cfg := &Config{
		DashboardPort: "",
	}

	if cfg.IsDashboardEnabled() {
		t.Error("Expected IsDashboardEnabled to be false")
	}
}

func TestParseCommaSeparated_MultipleValues(t *testing.T) {
	result := parseCommaSeparated("topic1,topic2,topic3")
	expected := []string{"topic1", "topic2", "topic3"}

	if len(result) != len(expected) {
		t.Fatalf("Expected %d items, got %d", len(expected), len(result))
	}

	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("Expected item %d to be '%s', got '%s'", i, expected[i], result[i])
		}
	}
}

func TestParseCommaSeparated_WithSpaces(t *testing.T) {
	result := parseCommaSeparated(" topic1 , topic2 , topic3 ")
	expected := []string{"topic1", "topic2", "topic3"}

	if len(result) != len(expected) {
		t.Fatalf("Expected %d items, got %d", len(expected), len(result))
	}

	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("Expected item %d to be '%s', got '%s'", i, expected[i], result[i])
		}
	}
}

func TestParseCommaSeparated_EmptyString(t *testing.T) {
	result := parseCommaSeparated("")
	if result != nil {
		t.Errorf("Expected nil for empty string, got %v", result)
	}
}

func TestParseCommaSeparated_SingleValue(t *testing.T) {
	result := parseCommaSeparated("topic1")
	expected := []string{"topic1"}

	if len(result) != len(expected) {
		t.Fatalf("Expected %d items, got %d", len(expected), len(result))
	}

	if result[0] != expected[0] {
		t.Errorf("Expected '%s', got '%s'", expected[0], result[0])
	}
}

func TestGetEnvOrDefault_WithValue(t *testing.T) {
	_ = os.Setenv("TEST_KEY", "test-value")
	defer func() { _ = os.Unsetenv("TEST_KEY") }()

	result := getEnvOrDefault("TEST_KEY", "default")
	if result != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", result)
	}
}

func TestGetEnvOrDefault_WithoutValue(t *testing.T) {
	result := getEnvOrDefault("NONEXISTENT_KEY", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}
}

// Helper function to clean up environment variables after tests
func cleanupEnv() {
	_ = os.Unsetenv("PUBSUB_PROJECT")
	_ = os.Unsetenv("PUBSUB_TOPIC")
	_ = os.Unsetenv("PUBSUB_SUBSCRIPTION")
	_ = os.Unsetenv("DASHBOARD_PORT")
	_ = os.Unsetenv("PUBSUB_PORT")
}
