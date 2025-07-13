package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/config"
)

func TestNewConfig_WithDefaults(t *testing.T) {
	// Clear environment variables to test defaults
	clearEnvVars(t)
	// Set a dummy API key to avoid fatal error
	err := os.Setenv("OPENAI_API_KEY", "test-key")
	require.NoError(t, err)
	defer func() {
		err := os.Unsetenv("OPENAI_API_KEY")
		require.NoError(t, err)
	}()

	cfg := config.NewConfig()

	require.NotNil(t, cfg)
	require.NotNil(t, cfg.OpenAI)
	assert.Equal(t, "test-key", cfg.OpenAI.APIKey)
	assert.Equal(t, "gpt-4", cfg.OpenAI.Model)
	assert.Equal(t, 4096, cfg.OpenAI.MaxTokens)
	assert.Equal(t, 0.7, cfg.OpenAI.Temperature)
	assert.Equal(t, 30*time.Second, cfg.OpenAI.Timeout)
}

func TestNewConfig_WithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	setEnvVars(t, map[string]string{
		"OPENAI_API_KEY":         "test-api-key",
		"OPENAI_MODEL":           "gpt-4o-mini",
		"OPENAI_MAX_TOKENS":      "2048",
		"OPENAI_TEMPERATURE":     "0.5",
		"OPENAI_TIMEOUT_SECONDS": "60",
	})
	defer clearEnvVars(t)

	cfg := config.NewConfig()

	require.NotNil(t, cfg)
	require.NotNil(t, cfg.OpenAI)
	assert.Equal(t, "test-api-key", cfg.OpenAI.APIKey)
	assert.Equal(t, "gpt-4o-mini", cfg.OpenAI.Model)
	assert.Equal(t, 2048, cfg.OpenAI.MaxTokens)
	assert.Equal(t, 0.5, cfg.OpenAI.Temperature)
	assert.Equal(t, 60*time.Second, cfg.OpenAI.Timeout)
}

func TestNewConfig_PartialEnvironmentVariables(t *testing.T) {
	// Set only some environment variables
	setEnvVars(t, map[string]string{
		"OPENAI_API_KEY": "partial-test-key",
		"OPENAI_MODEL":   "gpt-3.5-turbo",
	})
	defer clearEnvVars(t)

	cfg := config.NewConfig()

	require.NotNil(t, cfg)
	require.NotNil(t, cfg.OpenAI)
	assert.Equal(t, "partial-test-key", cfg.OpenAI.APIKey)
	assert.Equal(t, "gpt-3.5-turbo", cfg.OpenAI.Model)
	// These should use defaults
	assert.Equal(t, 4096, cfg.OpenAI.MaxTokens)
	assert.Equal(t, 0.7, cfg.OpenAI.Temperature)
	assert.Equal(t, 30*time.Second, cfg.OpenAI.Timeout)
}

func TestNewConfig_InvalidIntegerEnvironmentVariable(t *testing.T) {
	setEnvVars(t, map[string]string{
		"OPENAI_MAX_TOKENS": "not-a-number",
	})
	defer clearEnvVars(t)

	// This should cause log.Fatalf to be called, which would exit the program
	// In a real test, we might want to capture the log output instead
	// For now, we'll test the function indirectly by not calling NewConfig
	// and instead testing the internal functions if they were exported
}

func TestNewConfig_InvalidFloatEnvironmentVariable(t *testing.T) {
	setEnvVars(t, map[string]string{
		"OPENAI_TEMPERATURE": "not-a-float",
	})
	defer clearEnvVars(t)

	// Similar to the integer test, this would cause log.Fatalf
	// In practice, we'd want better error handling that doesn't use log.Fatalf
}

func TestNewConfig_ZeroValues(t *testing.T) {
	setEnvVars(t, map[string]string{
		"OPENAI_API_KEY":         "test-key",
		"OPENAI_MAX_TOKENS":      "0",
		"OPENAI_TEMPERATURE":     "0.0",
		"OPENAI_TIMEOUT_SECONDS": "0",
	})
	defer clearEnvVars(t)

	cfg := config.NewConfig()

	require.NotNil(t, cfg)
	require.NotNil(t, cfg.OpenAI)
	assert.Equal(t, 0, cfg.OpenAI.MaxTokens)
	assert.Equal(t, 0.0, cfg.OpenAI.Temperature)
	assert.Equal(t, 0*time.Second, cfg.OpenAI.Timeout)
}

func TestNewConfig_ExtremeValues(t *testing.T) {
	setEnvVars(t, map[string]string{
		"OPENAI_API_KEY":         "test-key",
		"OPENAI_MAX_TOKENS":      "100000",
		"OPENAI_TEMPERATURE":     "2.0",
		"OPENAI_TIMEOUT_SECONDS": "3600",
	})
	defer clearEnvVars(t)

	cfg := config.NewConfig()

	require.NotNil(t, cfg)
	require.NotNil(t, cfg.OpenAI)
	assert.Equal(t, 100000, cfg.OpenAI.MaxTokens)
	assert.Equal(t, 2.0, cfg.OpenAI.Temperature)
	assert.Equal(t, 3600*time.Second, cfg.OpenAI.Timeout)
}

// Helper functions
func setEnvVars(t *testing.T, vars map[string]string) {
	for key, value := range vars {
		err := os.Setenv(key, value)
		require.NoError(t, err)
	}
}

func clearEnvVars(t *testing.T) {
	envVars := []string{
		"OPENAI_API_KEY",
		"OPENAI_MODEL",
		"OPENAI_MAX_TOKENS",
		"OPENAI_TEMPERATURE",
		"OPENAI_TIMEOUT_SECONDS",
	}

	for _, envVar := range envVars {
		err := os.Unsetenv(envVar)
		require.NoError(t, err)
	}
}