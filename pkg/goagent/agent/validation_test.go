package agent_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitalii-honchar/go-agent/internal/validation"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/agent"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
)

type TestResult struct {
	Answer string `json:"answer"`
}

func TestNewAgent_ValidAgent(t *testing.T) {
	t.Parallel()

	validAgent, err := agent.NewAgent(
		agent.WithName[TestResult]("test_agent"),
		agent.WithLLMConfig[TestResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      "test-api-key",
			Model:       "gpt-4",
			Temperature: 0.0,
		}),
		agent.WithBehavior[TestResult]("You are a helpful assistant."),
	)

	require.NoError(t, err)
	require.NotNil(t, validAgent)
}

func TestNewAgent_EmptyName(t *testing.T) {
	t.Parallel()

	_, err := agent.NewAgent(
		agent.WithName[TestResult](""),
		agent.WithLLMConfig[TestResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      "test-api-key",
			Model:       "gpt-4",
			Temperature: 0.0,
		}),
		agent.WithBehavior[TestResult]("You are a helpful assistant."),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create agent")
	assert.Contains(t, err.Error(), "name")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestNewAgent_InvalidName(t *testing.T) {
	t.Parallel()

	_, err := agent.NewAgent(
		agent.WithName[TestResult]("Invalid Name"),
		agent.WithLLMConfig[TestResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      "test-api-key",
			Model:       "gpt-4",
			Temperature: 0.0,
		}),
		agent.WithBehavior[TestResult]("You are a helpful assistant."),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create agent")
	assert.Contains(t, err.Error(), "name")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestNewAgent_EmptyBehavior(t *testing.T) {
	t.Parallel()

	_, err := agent.NewAgent(
		agent.WithName[TestResult]("test_agent"),
		agent.WithLLMConfig[TestResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      "test-api-key",
			Model:       "gpt-4",
			Temperature: 0.0,
		}),
		agent.WithBehavior[TestResult](""),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create agent")
	assert.Contains(t, err.Error(), "behavior")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestNewAgent_EmptyLLMType(t *testing.T) {
	t.Parallel()

	_, err := agent.NewAgent(
		agent.WithName[TestResult]("test_agent"),
		agent.WithLLMConfig[TestResult](llm.LLMConfig{
			Type:        "",
			APIKey:      "test-api-key",
			Model:       "gpt-4",
			Temperature: 0.0,
		}),
		agent.WithBehavior[TestResult]("You are a helpful assistant."),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create agent")
	assert.Contains(t, err.Error(), "llm config")
	assert.Contains(t, err.Error(), "type")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestNewAgent_EmptyAPIKey(t *testing.T) {
	t.Parallel()

	_, err := agent.NewAgent(
		agent.WithName[TestResult]("test_agent"),
		agent.WithLLMConfig[TestResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      "",
			Model:       "gpt-4",
			Temperature: 0.0,
		}),
		agent.WithBehavior[TestResult]("You are a helpful assistant."),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create agent")
	assert.Contains(t, err.Error(), "llm config")
	assert.Contains(t, err.Error(), "api key")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestNewAgent_EmptyModel(t *testing.T) {
	t.Parallel()

	_, err := agent.NewAgent(
		agent.WithName[TestResult]("test_agent"),
		agent.WithLLMConfig[TestResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      "test-api-key",
			Model:       "",
			Temperature: 0.0,
		}),
		agent.WithBehavior[TestResult]("You are a helpful assistant."),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create agent")
	assert.Contains(t, err.Error(), "llm config")
	assert.Contains(t, err.Error(), "model")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestNewAgent_WhitespaceBehavior(t *testing.T) {
	t.Parallel()

	_, err := agent.NewAgent(
		agent.WithName[TestResult]("test_agent"),
		agent.WithLLMConfig[TestResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      "test-api-key",
			Model:       "gpt-4",
			Temperature: 0.0,
		}),
		agent.WithBehavior[TestResult]("   \t\n   "),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create agent")
	assert.Contains(t, err.Error(), "behavior")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestNewAgent_TooLongName(t *testing.T) {
	t.Parallel()

	longName := "a_very_long_name_that_exceeds_the_maximum_allowed_length_for_names_in_the_system"

	_, err := agent.NewAgent(
		agent.WithName[TestResult](longName),
		agent.WithLLMConfig[TestResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      "test-api-key",
			Model:       "gpt-4",
			Temperature: 0.0,
		}),
		agent.WithBehavior[TestResult]("You are a helpful assistant."),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create agent")
	assert.Contains(t, err.Error(), "name")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}
