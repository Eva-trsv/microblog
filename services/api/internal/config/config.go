package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL              string
	KafkaBrokers             []string
	KafkaTopicUserRegistered string
	KafkaTopicPostCreated    string
	KafkaTopicPostLiked      string
	KafkaWriteTimeout        time.Duration
	KafkaRetries             int
}

func Load() *Config {
	return &Config{
		DatabaseURL:              os.Getenv("DATABASE_URL"),
		KafkaBrokers:             splitEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopicUserRegistered: getenv("KAFKA_TOPIC_USER_REGISTERED", "user_registered"),
		KafkaTopicPostCreated:    getenv("KAFKA_TOPIC_POST_CREATED", "post_created"),
		KafkaTopicPostLiked:      getenv("KAFKA_TOPIC_POST_LIKED", "post_liked"),
		KafkaWriteTimeout:        time.Duration(intEnv("KAFKA_WRITE_TIMEOUT_SECONDS", 5)) * time.Second,
		KafkaRetries:             intEnv("KAFKA_RETRIES", 3),
	}
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func splitEnv(key, fallback string) []string {
	raw := getenv(key, fallback)
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func intEnv(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
