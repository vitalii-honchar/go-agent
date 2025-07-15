package llm_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitalii-honchar/go-agent/internal/validation"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
)

type TestParams struct {
	Input string `json:"input"`
}

type TestResult struct {
	llm.BaseLLMToolResult
	Output string `json:"output"`
}

func TestNewLLMTool_ValidTool(t *testing.T) {
	t.Parallel()

	tool, err := llm.NewLLMTool(
		llm.WithLLMToolName("test_tool"),
		llm.WithLLMToolDescription("A test tool"),
		llm.WithLLMToolParametersSchema[TestParams](),
		llm.WithLLMToolCall(func(callID string, params TestParams) (TestResult, error) {
			return TestResult{
				BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
				Output:            params.Input,
			}, nil
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, "test_tool", tool.Name)
	assert.Equal(t, "A test tool", tool.Description)
	assert.NotNil(t, tool.ParametersSchema)
	assert.NotNil(t, tool.Call)
}

func TestNewLLMTool_EmptyName(t *testing.T) {
	t.Parallel()

	_, err := llm.NewLLMTool(
		llm.WithLLMToolName(""),
		llm.WithLLMToolDescription("A test tool"),
		llm.WithLLMToolParametersSchema[TestParams](),
		llm.WithLLMToolCall(func(callID string, params TestParams) (TestResult, error) {
			return TestResult{
				BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
				Output:            params.Input,
			}, nil
		}),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create LLM tool")
	assert.Contains(t, err.Error(), "tool:")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestNewLLMTool_InvalidNamePattern(t *testing.T) {
	t.Parallel()

	_, err := llm.NewLLMTool(
		llm.WithLLMToolName("Invalid-Name-With-Dashes"),
		llm.WithLLMToolDescription("A test tool"),
		llm.WithLLMToolParametersSchema[TestParams](),
		llm.WithLLMToolCall(func(callID string, params TestParams) (TestResult, error) {
			return TestResult{
				BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
				Output:            params.Input,
			}, nil
		}),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create LLM tool")
	assert.Contains(t, err.Error(), "tool:")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestNewLLMTool_NameTooLong(t *testing.T) {
	t.Parallel()

	longName := strings.Repeat("a", 65) // > 64 chars

	_, err := llm.NewLLMTool(
		llm.WithLLMToolName(longName),
		llm.WithLLMToolDescription("A test tool"),
		llm.WithLLMToolParametersSchema[TestParams](),
		llm.WithLLMToolCall(func(callID string, params TestParams) (TestResult, error) {
			return TestResult{
				BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
				Output:            params.Input,
			}, nil
		}),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create LLM tool")
	assert.Contains(t, err.Error(), "tool:")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestNewLLMTool_EmptyDescription(t *testing.T) {
	t.Parallel()

	_, err := llm.NewLLMTool(
		llm.WithLLMToolName("test_tool"),
		llm.WithLLMToolDescription(""),
		llm.WithLLMToolParametersSchema[TestParams](),
		llm.WithLLMToolCall(func(callID string, params TestParams) (TestResult, error) {
			return TestResult{
				BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
				Output:            params.Input,
			}, nil
		}),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create LLM tool")
	assert.Contains(t, err.Error(), "description:")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestNewLLMTool_DescriptionTooLong(t *testing.T) {
	t.Parallel()

	longDescription := strings.Repeat("a", 1025) // > 1024 chars

	_, err := llm.NewLLMTool(
		llm.WithLLMToolName("test_tool"),
		llm.WithLLMToolDescription(longDescription),
		llm.WithLLMToolParametersSchema[TestParams](),
		llm.WithLLMToolCall(func(callID string, params TestParams) (TestResult, error) {
			return TestResult{
				BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
				Output:            params.Input,
			}, nil
		}),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create LLM tool")
	assert.Contains(t, err.Error(), "description:")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestNewLLMTool_MissingParametersSchema(t *testing.T) {
	t.Parallel()

	_, err := llm.NewLLMTool(
		llm.WithLLMToolName("test_tool"),
		llm.WithLLMToolDescription("A test tool"),
		// Don't set parameters schema
		llm.WithLLMToolCall(func(callID string, params TestParams) (TestResult, error) {
			return TestResult{
				BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
				Output:            params.Input,
			}, nil
		}),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create LLM tool")
	assert.Contains(t, err.Error(), "parameters schema:")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestNewLLMTool_MissingCallFunction(t *testing.T) {
	t.Parallel()

	_, err := llm.NewLLMTool(
		llm.WithLLMToolName("test_tool"),
		llm.WithLLMToolDescription("A test tool"),
		llm.WithLLMToolParametersSchema[TestParams](),
		// Don't set call function
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create LLM tool")
	assert.Contains(t, err.Error(), "call:")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestLLMTool_CallWithValidArgs(t *testing.T) {
	t.Parallel()

	tool, err := llm.NewLLMTool(
		llm.WithLLMToolName("test_tool"),
		llm.WithLLMToolDescription("A test tool"),
		llm.WithLLMToolParametersSchema[TestParams](),
		llm.WithLLMToolCall(func(callID string, params TestParams) (TestResult, error) {
			return TestResult{
				BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
				Output:            "processed: " + params.Input,
			}, nil
		}),
	)
	require.NoError(t, err)

	result, err := tool.Call("test-id", `{"input": "hello"}`)

	require.NoError(t, err)
	assert.Equal(t, "test-id", result.GetID())

	testResult, ok := result.(TestResult)
	require.True(t, ok)
	assert.Equal(t, "processed: hello", testResult.Output)
}

func TestLLMTool_CallWithInvalidArgs(t *testing.T) {
	t.Parallel()

	tool, err := llm.NewLLMTool(
		llm.WithLLMToolName("test_tool"),
		llm.WithLLMToolDescription("A test tool"),
		llm.WithLLMToolParametersSchema[TestParams](),
		llm.WithLLMToolCall(func(callID string, params TestParams) (TestResult, error) {
			return TestResult{
				BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
				Output:            params.Input,
			}, nil
		}),
	)
	require.NoError(t, err)

	_, err = tool.Call("test-id", `{invalid json`)

	require.Error(t, err)
	require.ErrorIs(t, err, llm.ErrInvalidArguments)
	assert.Contains(t, err.Error(), "failed to unmarshal arguments")
}

func TestNewLLMToolCall_ValidCall(t *testing.T) {
	t.Parallel()

	toolCall, err := llm.NewLLMToolCall("call-123", "test_tool", `{"param": "value"}`)

	require.NoError(t, err)
	assert.Equal(t, "call-123", toolCall.ID)
	assert.Equal(t, "test_tool", toolCall.ToolName)
	assert.JSONEq(t, `{"param": "value"}`, toolCall.Args)
}

func TestNewLLMToolCall_EmptyID(t *testing.T) {
	t.Parallel()

	_, err := llm.NewLLMToolCall("", "test_tool", `{"param": "value"}`)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create LLM tool call")
	assert.Contains(t, err.Error(), "id:")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestNewLLMToolCall_EmptyToolName(t *testing.T) {
	t.Parallel()

	_, err := llm.NewLLMToolCall("call-123", "", `{"param": "value"}`)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create LLM tool call")
	assert.Contains(t, err.Error(), "tool name:")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestNewLLMToolCall_InvalidToolName(t *testing.T) {
	t.Parallel()

	_, err := llm.NewLLMToolCall("call-123", "Invalid-Tool-Name", `{"param": "value"}`)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create LLM tool call")
	assert.Contains(t, err.Error(), "tool name:")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}

func TestNewLLMToolCall_EmptyArgs(t *testing.T) {
	t.Parallel()

	_, err := llm.NewLLMToolCall("call-123", "test_tool", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create LLM tool call")
	assert.Contains(t, err.Error(), "args:")
	assert.ErrorIs(t, err, validation.ErrValidationFailed)
}
