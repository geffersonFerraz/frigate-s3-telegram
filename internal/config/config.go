package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	BUCKET_SERVER       string
	BUCKET_NAME         string
	BUCKET_REGION       string
	KEY_PAIR_ID         string
	KEY_PAIR_SECRET     string
	TelegramBotToken    string
	FrigateURL          string
	FrigateEventLimit   int
	TelegramChatID      int64
	TelegramErrorChatID int64
	RabbitURL           string
	RabbitExchange      string
	RabbitQueue         string
	RabbitRoutingKey    string
	RedisAddr           string
	RedisPassword       string
	RedisDB             int
	RedisProtocol       int
	RedisTTL            int
}

// New returns a new Config struct
func New() *Config {
	return &Config{
		BUCKET_SERVER:       getEnv("BUCKET_SERVER", "play.min.io"),
		BUCKET_NAME:         getEnv("BUCKET_NAME", "mybucket"),
		KEY_PAIR_ID:         getEnv("KEY_PAIR_ID", "Q3AM3UQ867SPQQA43P2F"),
		KEY_PAIR_SECRET:     getEnv("KEY_PAIR_SECRET", "Q3AM3UQ867SPQQA43P2F"),
		BUCKET_REGION:       getEnv("BUCKET_REGION", "us-east-1"),
		TelegramBotToken:    getEnv("TELEGRAM_BOT_TOKEN", ""),
		FrigateURL:          getEnv("FRIGATE_URL", "http://localhost:5000"),
		FrigateEventLimit:   getEnvAsInt("FRIGATE_EVENT_LIMIT", 20),
		TelegramChatID:      getEnvAsInt64("TELEGRAM_CHAT_ID", 0),
		TelegramErrorChatID: getEnvAsInt64("TELEGRAM_ERROR_CHAT_ID", getEnvAsInt64("TELEGRAM_CHAT_ID", 0)),
		RabbitURL:           getEnv("RABBIT_URL", "amqp://guest:guest@localhost:5672/"),
		RabbitExchange:      getEnv("RABBIT_EXCHANGE", "frigate"),
		RabbitQueue:         getEnv("RABBIT_QUEUE", "frigate"),
		RabbitRoutingKey:    getEnv("RABBIT_ROUTING_KEY", "frigate"),
		RedisAddr:           getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:       getEnv("REDIS_PASSWORD", ""),
		RedisDB:             getEnvAsInt("REDIS_DB", 0),
		RedisProtocol:       getEnvAsInt("REDIS_PROTOCOL", 3),
		RedisTTL:            getEnvAsInt("REDIS_TTL", 1209600), // 7 days
	}
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// Simple helper function to read an environment variable into integer or return a default value
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

// Simple helper function to read an environment variable into integer or return a default value
func getEnvAsInt64(name string, defaultVal int64) int64 {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return int64(value)
	}

	return defaultVal
}

// Helper to read an environment variable into a bool or return default value
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}

// Helper to read an environment variable into a string slice or return default value
func getEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := getEnv(name, "")

	if valStr == "" {
		return defaultVal
	}

	val := strings.Split(valStr, sep)

	return val
}
