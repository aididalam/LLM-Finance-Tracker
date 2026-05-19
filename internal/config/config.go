package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	AppEnv  string
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	TelegramBotToken string
	TelegramChatID   string

	LLMProvider     string
	AnthropicAPIKey string
	AnthropicModel  string
	OpenAIAPIKey    string
	OpenAIModel     string

	JWTSecret string

	LogLevel string
}

var required = []string{
	"APP_ENV",
	"APP_PORT",
	"DB_HOST",
	"DB_PORT",
	"DB_USER",
	"DB_PASSWORD",
	"DB_NAME",
	"TELEGRAM_BOT_TOKEN",
	"TELEGRAM_CHAT_ID",
	"LLM_PROVIDER",
}

func Load() *Config {
	_ = godotenv.Load()

	for _, key := range required {
		if os.Getenv(key) == "" {
			log.Fatal().Msgf("required environment variable %s is missing", key)
		}
	}

	return &Config{
		AppEnv:                getenv("APP_ENV", "development"),
		AppPort:               getenv("APP_PORT", "8080"),
		DBHost:                os.Getenv("DB_HOST"),
		DBPort:                getenv("DB_PORT", "3306"),
		DBUser:                os.Getenv("DB_USER"),
		DBPassword:            os.Getenv("DB_PASSWORD"),
		DBName:                os.Getenv("DB_NAME"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
		LLMProvider:           os.Getenv("LLM_PROVIDER"),
		AnthropicAPIKey:       os.Getenv("ANTHROPIC_API_KEY"),
		AnthropicModel:        getenv("ANTHROPIC_MODEL", "claude-haiku-4-5-20251001"),
		OpenAIAPIKey:          os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:           getenv("OPENAI_MODEL", "gpt-4o-mini"), // same tier as claude-haiku
		JWTSecret:             os.Getenv("JWT_SECRET"),
		LogLevel:              getenv("LOG_LEVEL", "info"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
