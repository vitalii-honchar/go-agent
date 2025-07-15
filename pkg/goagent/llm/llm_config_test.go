package llm_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitalii-honchar/go-agent/internal/validation"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
)

func TestLLMConfig_Validate_ValidConfig(t *testing.T) {
	t.Parallel()

	config := llm.LLMConfig{
		Type:        llm.LLMTypeOpenAI,
		APIKey:      "test-api-key",
		Model:       "gpt-4",
		Temperature: 0.0,
	}

	err := config.Validate()

	require.NoError(t, err)
}

func TestLLMConfig_Validate_EmptyType(t *testing.T) {
	t.Parallel()

	config := llm.LLMConfig{
		Type:        "",
		APIKey:      "test-api-key",
		Model:       "gpt-4",
		Temperature: 0.0,
	}

	err := config.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "type")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestLLMConfig_Validate_EmptyAPIKey(t *testing.T) {
	t.Parallel()

	config := llm.LLMConfig{
		Type:        llm.LLMTypeOpenAI,
		APIKey:      "",
		Model:       "gpt-4",
		Temperature: 0.0,
	}

	err := config.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "api key")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestLLMConfig_Validate_EmptyModel(t *testing.T) {
	t.Parallel()

	config := llm.LLMConfig{
		Type:        llm.LLMTypeOpenAI,
		APIKey:      "test-api-key",
		Model:       "",
		Temperature: 0.0,
	}

	err := config.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "model")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestLLMConfig_Validate_AllFieldsEmpty(t *testing.T) {
	t.Parallel()

	config := llm.LLMConfig{
		Type:        "",
		APIKey:      "",
		Model:       "",
		Temperature: 0.0,
	}

	err := config.Validate()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "type")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}
