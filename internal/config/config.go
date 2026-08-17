package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr     string
	DatabaseURL  string
	RedisURL     string
	KafkaBrokers []string
	KafkaTopic   string
	CacheTTL     time.Duration
}

func Load() Config {
	return Config{
		HTTPAddr:     env("HTTP_ADDR", ":8080"),
		DatabaseURL:  env("DATABASE_URL", "postgres://flags:flags@localhost:5432/flags?sslmode=disable"),
		RedisURL:     env("REDIS_URL", "redis://localhost:6379/0"),
		KafkaBrokers: strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ","),
		KafkaTopic:   env("KAFKA_TOPIC", "flag-events"),
		CacheTTL:     durationEnv("CACHE_TTL", 30*time.Second),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return fallback
}
