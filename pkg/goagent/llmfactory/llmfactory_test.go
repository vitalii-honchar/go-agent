package llmfactory_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/llmfactory"
)

func TestCreateLLM_OpenAI(t *testing.T) {
	t.Parallel()
	cfg := llm.LLMConfig{
		Type:        llm.LLMTypeOpenAI,
		APIKey:      "test-key",
		Model:       "gpt-4o-mini",
		Temperature: 0.5,
	}

	tools := map[string]llm.LLMTool{
		"test": createTestTool(),
	}

	result, err := llmfactory.CreateLLM(cfg, tools)

	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCreateLLM_OpenAI_NoTools(t *testing.T) {
	t.Parallel()
	cfg := llm.LLMConfig{
		Type:        llm.LLMTypeOpenAI,
		APIKey:      "test-key",
		Model:       "gpt-4o-mini",
		Temperature: 0.0,
	}

	result, err := llmfactory.CreateLLM(cfg, nil)

	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCreateLLM_OpenAI_EmptyTools(t *testing.T) {
	t.Parallel()
	cfg := llm.LLMConfig{
		Type:        llm.LLMTypeOpenAI,
		APIKey:      "test-key",
		Model:       "gpt-4o-mini",
		Temperature: 0.0,
	}

	tools := map[string]llm.LLMTool{}

	result, err := llmfactory.CreateLLM(cfg, tools)

	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCreateLLM_UnsupportedType(t *testing.T) {
	t.Parallel()
	cfg := llm.LLMConfig{
		Type:        "unsupported",
		APIKey:      "test-key",
		Model:       "test-model",
		Temperature: 0.0,
	}

	result, err := llmfactory.CreateLLM(cfg, nil)

	require.Error(t, err)
	require.ErrorIs(t, err, llm.ErrUnsupportedLLMType)
	assert.Nil(t, result)
}

func TestCreateLLM_MultipleTools(t *testing.T) {
	t.Parallel()
	cfg := llm.LLMConfig{
		Type:        llm.LLMTypeOpenAI,
		APIKey:      "test-key",
		Model:       "gpt-4o-mini",
		Temperature: 0.1,
	}

	tools := map[string]llm.LLMTool{
		"tool1": createTestTool(),
		"tool2": createTestTool(),
		"tool3": createTestTool(),
	}

	result, err := llmfactory.CreateLLM(cfg, tools)

	require.NoError(t, err)
	assert.NotNil(t, result)
}

type TestToolParams struct {
	Input string `json:"input"`
}

type TestToolResult struct {
	llm.BaseLLMToolResult
	Output string `json:"output"`
}

func createTestTool() llm.LLMTool {
	return llm.NewLLMTool(
		llm.WithLLMToolName("test"),
		llm.WithLLMToolDescription("Test tool"),
		llm.WithLLMToolParametersSchema[TestToolParams](),
		llm.WithLLMToolCall(func(callID string, params TestToolParams) (TestToolResult, error) {
			return TestToolResult{
				BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
				Output:            "test-output",
			}, nil
		}),
	)
}
