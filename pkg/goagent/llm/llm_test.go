package llm_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
)

func TestNewLLMMessage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		msgType  llm.LLMMessageType
		content  string
		expected llm.LLMMessage
	}{
		{
			name:    "user message",
			msgType: llm.LLMMessageTypeUser,
			content: "Hello",
			expected: llm.LLMMessage{
				Type:    llm.LLMMessageTypeUser,
				Content: "Hello",
			},
		},
		{
			name:    "assistant message",
			msgType: llm.LLMMessageTypeAssistant,
			content: "Hi there!",
			expected: llm.LLMMessage{
				Type:    llm.LLMMessageTypeAssistant,
				Content: "Hi there!",
			},
		},
		{
			name:    "system message",
			msgType: llm.LLMMessageTypeSystem,
			content: "You are a helpful assistant",
			expected: llm.LLMMessage{
				Type:    llm.LLMMessageTypeSystem,
				Content: "You are a helpful assistant",
			},
		},
		{
			name:    "empty content",
			msgType: llm.LLMMessageTypeUser,
			content: "",
			expected: llm.LLMMessage{
				Type:    llm.LLMMessageTypeUser,
				Content: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := llm.NewLLMMessage(tt.msgType, tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewLLMTool(t *testing.T) {
	t.Parallel()
	tool := llm.NewLLMTool()

	// Default tool should have empty values
	assert.Empty(t, tool.Name)
	assert.Empty(t, tool.Description)
	assert.Nil(t, tool.ParametersSchema)
	assert.Nil(t, tool.Call)
}

func TestWithLLMToolName(t *testing.T) {
	t.Parallel()
	name := "test-tool"
	tool := llm.NewLLMTool(
		llm.WithLLMToolName(name),
	)

	assert.Equal(t, name, tool.Name)
}

func TestWithLLMToolDescription(t *testing.T) {
	t.Parallel()
	description := "A test tool for testing"
	tool := llm.NewLLMTool(
		llm.WithLLMToolDescription(description),
	)

	assert.Equal(t, description, tool.Description)
}

func TestWithLLMToolParametersSchema(t *testing.T) {
	t.Parallel()
	type TestParams struct {
		Value string `json:"value"`
	}

	tool := llm.NewLLMTool(
		llm.WithLLMToolParametersSchema[TestParams](),
	)

	assert.NotNil(t, tool.ParametersSchema)
	_, ok := tool.ParametersSchema.(*TestParams)
	assert.True(t, ok)
}

func TestWithLLMToolCall(t *testing.T) {
	t.Parallel()
	type TestParams struct {
		Input string `json:"input"`
	}

	type TestResult struct {
		llm.BaseLLMToolResult
		Output string `json:"output"`
	}

	callFunc := func(callID string, params TestParams) (TestResult, error) {
		return TestResult{
			BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
			Output:            "processed: " + params.Input,
		}, nil
	}

	tool := llm.NewLLMTool(
		llm.WithLLMToolCall[TestParams, TestResult](callFunc),
	)

	assert.NotNil(t, tool.Call)

	// Test the call function
	result, err := tool.Call("test-id", `{"input": "hello"}`)
	require.NoError(t, err)
	assert.Equal(t, "test-id", result.GetID())

	testResult, ok := result.(TestResult)
	require.True(t, ok)
	assert.Equal(t, "processed: hello", testResult.Output)
}

func TestWithLLMToolCall_InvalidJSON(t *testing.T) {
	t.Parallel()
	type TestParams struct {
		Input string `json:"input"`
	}

	type TestResult struct {
		llm.BaseLLMToolResult
		Output string `json:"output"`
	}

	callFunc := func(callID string, params TestParams) (TestResult, error) {
		return TestResult{
			BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
			Output:            params.Input,
		}, nil
	}

	tool := llm.NewLLMTool(
		llm.WithLLMToolCall[TestParams, TestResult](callFunc),
	)

	// Test with invalid JSON
	_, err := tool.Call("test-id", `invalid json`)
	require.Error(t, err)
	assert.ErrorIs(t, err, llm.ErrInvalidArguments)
}

func TestLLMTool_CompleteConfiguration(t *testing.T) {
	t.Parallel()
	type TestParams struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	type TestResult struct {
		llm.BaseLLMToolResult
		Sum int `json:"sum"`
	}

	tool := llm.NewLLMTool(
		llm.WithLLMToolName("add"),
		llm.WithLLMToolDescription("Adds two numbers"),
		llm.WithLLMToolParametersSchema[TestParams](),
		llm.WithLLMToolCall[TestParams, TestResult](func(callID string, params TestParams) (TestResult, error) {
			return TestResult{
				BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
				Sum:               params.X + params.Y,
			}, nil
		}),
	)

	assert.Equal(t, "add", tool.Name)
	assert.Equal(t, "Adds two numbers", tool.Description)
	assert.NotNil(t, tool.ParametersSchema)
	assert.NotNil(t, tool.Call)

	// Test the complete tool
	result, err := tool.Call("call-123", `{"x": 5, "y": 3}`)
	require.NoError(t, err)
	assert.Equal(t, "call-123", result.GetID())

	testResult, ok := result.(TestResult)
	require.True(t, ok)
	assert.Equal(t, 8, testResult.Sum)
}

func TestBaseLLMToolResult_GetID(t *testing.T) {
	t.Parallel()
	result := llm.BaseLLMToolResult{ID: "test-id-123"}
	assert.Equal(t, "test-id-123", result.GetID())
}

func TestNewLLMToolCall(t *testing.T) {
	t.Parallel()
	callID := "call-123"
	toolName := "test-tool"
	args := `{"param": "value"}`

	toolCall := llm.NewLLMToolCall(callID, toolName, args)

	assert.Equal(t, callID, toolCall.ID)
	assert.Equal(t, toolName, toolCall.ToolName)
	assert.Equal(t, args, toolCall.Args)
}

func TestLLMMessageTypes(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "user", string(llm.LLMMessageTypeUser))
	assert.Equal(t, "assistant", string(llm.LLMMessageTypeAssistant))
	assert.Equal(t, "system", string(llm.LLMMessageTypeSystem))
}
