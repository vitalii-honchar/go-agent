// Package config provides configuration management for the Go Agent library
package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config represents the main configuration structure
type Config struct {
	OpenAI *OpenAIConfig
}

// OpenAIConfig contains OpenAI-specific configuration
type OpenAIConfig struct {
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float64
	Timeout     time.Duration
}

// NewConfig creates a new Config instance with values from environment variables
func NewConfig() *Config {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
		// Continue execution - not finding .env is not a fatal error
		// as environment variables might be set through other means
	}

	return &Config{
		OpenAI: &OpenAIConfig{
			APIKey:      getEnvStr("OPENAI_API_KEY", ""),
			Model:       getEnvStr("OPENAI_MODEL", "gpt-4"),
			MaxTokens:   getEnvInt("OPENAI_MAX_TOKENS", 4096),
			Temperature: getEnvFloat("OPENAI_TEMPERATURE", 0.7),
			Timeout:     time.Duration(getEnvInt("OPENAI_TIMEOUT_SECONDS", 30)) * time.Second,
		},
	}
}

// getEnvStr returns the value of an environment variable or a default value if it's not set
func getEnvStr(envVar string, defaultValue string) string {
	value := os.Getenv(envVar)
	if value == "" {
		if defaultValue == "" {
			log.Fatalf("environment variable cannot be empty and no default provided: %s", envVar)
		}

		return defaultValue
	}

	return value
}

// getEnvInt returns the value of an environment variable as an integer or a default value if it's not set
func getEnvInt(envVar string, defaultValue int) int {
	value := os.Getenv(envVar)
	if value == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		log.Fatalf("environment variable must be an integer: name = %s, value = %s", envVar, value)
	}

	return i
}

// getEnvFloat returns the value of an environment variable as a float64 or a default value if it's not set
func getEnvFloat(envVar string, defaultValue float64) float64 {
	value := os.Getenv(envVar)
	if value == "" {
		return defaultValue
	}

	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Fatalf("environment variable must be a float: name = %s, value = %s", envVar, value)
	}

	return f
}
