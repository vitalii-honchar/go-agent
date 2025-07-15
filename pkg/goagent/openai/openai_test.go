package openai_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/openai"
)

func TestOpenAILLM_Call(t *testing.T) {
	t.Parallel()

	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	openaiLLM := openai.NewOpenAILLM(
		openai.WithAPIKey(apiKey),
		openai.WithModel("gpt-4o-mini"),
		openai.WithTemperature(0.0),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages := []llm.LLMMessage{
		{
			Type:    llm.LLMMessageTypeSystem,
			Content: "You are a helpful assistant.",
		},
		{
			Type:    llm.LLMMessageTypeUser,
			Content: "Say hello in exactly 3 words.",
		},
	}

	result, err := openaiLLM.Call(ctx, messages)

	require.NoError(t, err)
	assert.Equal(t, llm.LLMMessageTypeAssistant, result.Type)
	assert.NotEmpty(t, result.Content)
	assert.True(t, result.End)
	assert.Empty(t, result.ToolCalls)
}

func TestOpenAILLM_CallWithStructuredOutput(t *testing.T) {
	t.Parallel()

	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	openaiLLM := openai.NewOpenAILLM(
		openai.WithAPIKey(apiKey),
		openai.WithModel("gpt-4o-mini"),
		openai.WithTemperature(0.0),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages := []llm.LLMMessage{
		{
			Type:    llm.LLMMessageTypeSystem,
			Content: "You are a helpful assistant that responds in JSON format.",
		},
		{
			Type:    llm.LLMMessageTypeUser,
			Content: "Generate a person with name and age.",
		},
	}

	type Person struct {
		Name string `json:"name" jsonschema_description:"Person's name"`
		Age  int    `json:"age"  jsonschema_description:"Person's age"`
	}

	result, err := openaiLLM.CallWithStructuredOutput(ctx, messages, Person{})

	require.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "name")
	assert.Contains(t, result, "age")
}

func TestOpenAILLM_CallWithTools(t *testing.T) {
	t.Parallel()

	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	// Create a simple add tool
	addTool, err := llm.NewLLMTool(
		llm.WithLLMToolName("add"),
		llm.WithLLMToolDescription("Adds two numbers together"),
		llm.WithLLMToolParametersSchema[AddToolParams](),
		llm.WithLLMToolCall(func(callID string, params AddToolParams) (AddToolResult, error) {
			return AddToolResult{
				BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
				Sum:               params.Num1 + params.Num2,
			}, nil
		}),
	)
	require.NoError(t, err, "Failed to create add tool")

	openaiLLM := openai.NewOpenAILLM(
		openai.WithAPIKey(apiKey),
		openai.WithModel("gpt-4o-mini"),
		openai.WithTemperature(0.0),
		openai.WithTools([]llm.LLMTool{addTool}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages := []llm.LLMMessage{
		{
			Type:    llm.LLMMessageTypeSystem,
			Content: "You are a calculator. Use the add tool to calculate 5 + 3.",
		},
		{
			Type:    llm.LLMMessageTypeUser,
			Content: "What is 5 + 3?",
		},
	}

	result, err := openaiLLM.Call(ctx, messages)

	require.NoError(t, err)
	assert.Equal(t, llm.LLMMessageTypeAssistant, result.Type)
	assert.NotEmpty(t, result.ToolCalls)
	assert.Equal(t, "add", result.ToolCalls[0].ToolName)
}

func TestOpenAILLM_Options(t *testing.T) {
	t.Parallel()

	apiKey := "test-key"
	temperature := 0.5
	model := "gpt-4o-mini"

	openaiLLM := openai.NewOpenAILLM(
		openai.WithAPIKey(apiKey),
		openai.WithModel(model),
		openai.WithTemperature(temperature),
	)

	assert.NotNil(t, openaiLLM)
}

func TestOpenAILLM_CallWithAssistantMessages(t *testing.T) {
	t.Parallel()

	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	addTool := createTestAddTool()
	openaiLLM := openai.NewOpenAILLM(
		openai.WithAPIKey(apiKey),
		openai.WithModel("gpt-4o-mini"),
		openai.WithTemperature(0.0),
		openai.WithTools([]llm.LLMTool{addTool}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages := createTestMessages()
	result, err := openaiLLM.Call(ctx, messages)

	require.NoError(t, err)
	assert.Equal(t, llm.LLMMessageTypeAssistant, result.Type)
	assert.NotNil(t, result)
}

func createTestAddTool() llm.LLMTool {
	tool, err := llm.NewLLMTool(
		llm.WithLLMToolName("add"),
		llm.WithLLMToolDescription("Adds two numbers together"),
		llm.WithLLMToolParametersSchema[AddToolParams](),
		llm.WithLLMToolCall(func(callID string, params AddToolParams) (AddToolResult, error) {
			return AddToolResult{
				BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
				Sum:               params.Num1 + params.Num2,
			}, nil
		}),
	)
	if err != nil {
		panic("Failed to create test add tool: " + err.Error())
	}

	return tool
}

func createTestMessages() []llm.LLMMessage {
	return []llm.LLMMessage{
		{
			Type:    llm.LLMMessageTypeSystem,
			Content: "You are a calculator.",
		},
		{
			Type:    llm.LLMMessageTypeUser,
			Content: "What is 2 + 3?",
		},
		{
			Type:    llm.LLMMessageTypeAssistant,
			Content: "I'll calculate that for you.",
			ToolCalls: []llm.LLMToolCall{
				{
					ID:       "call_123",
					ToolName: "add",
					Args:     `{"num1": 2, "num2": 3}`,
				},
			},
			ToolResults: []llm.LLMToolResult{
				AddToolResult{
					BaseLLMToolResult: llm.BaseLLMToolResult{ID: "call_123"},
					Sum:               5,
				},
			},
		},
		{
			Type:    llm.LLMMessageTypeUser,
			Content: "Thank you! Now what is 10 + 15?",
		},
	}
}

func TestOpenAILLM_CallWithAssistantMessagesNoToolResults(t *testing.T) {
	t.Parallel()

	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	openaiLLM := openai.NewOpenAILLM(
		openai.WithAPIKey(apiKey),
		openai.WithModel("gpt-4o-mini"),
		openai.WithTemperature(0.0),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test assistant message without tool calls or results
	messages := []llm.LLMMessage{
		{
			Type:    llm.LLMMessageTypeSystem,
			Content: "You are a helpful assistant.",
		},
		{
			Type:    llm.LLMMessageTypeUser,
			Content: "Hello",
		},
		{
			Type:    llm.LLMMessageTypeAssistant,
			Content: "Hi there! How can I help you?",
		},
		{
			Type:    llm.LLMMessageTypeUser,
			Content: "Tell me a joke.",
		},
	}

	result, err := openaiLLM.Call(ctx, messages)

	require.NoError(t, err)
	assert.Equal(t, llm.LLMMessageTypeAssistant, result.Type)
	assert.NotEmpty(t, result.Content)
}

type AddToolParams struct {
	Num1 float64 `json:"num1"`
	Num2 float64 `json:"num2"`
}

type AddToolResult struct {
	llm.BaseLLMToolResult
	Sum float64 `json:"sum"`
}
