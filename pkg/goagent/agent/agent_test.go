package agent_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"go-agent/pkg/goagent/agent"
	"go-agent/pkg/goagent/llm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type AddNumbers struct {
	Num1 int `json:"num1"`
	Num2 int `json:"num2"`
}

type Result struct {
	Sum int `json:"sum"`
}

type AddToolResult struct {
	llm.BaseLLMToolResult
	Sum float64 `json:"sum"`
}

func createAddTool() llm.LLMTool {
	return llm.NewLLMTool(
		llm.WithLLMToolName("add"),
		llm.WithLLMToolDescription("Adds two numbers together"),
		llm.WithLLMToolParametersSchema(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"num1": map[string]any{"type": "number"},
				"num2": map[string]any{"type": "number"},
			},
			"required": []string{"num1", "num2"},
		}),
		llm.WithLLMToolCall(func(id string, args map[string]any) (AddToolResult, error) {
			num1, ok1 := args["num1"].(float64)
			num2, ok2 := args["num2"].(float64)

			if !ok1 || !ok2 {
				return AddToolResult{}, fmt.Errorf("%w: num1 = '%v', num2 = '%v'", llm.ErrInvalidArguments, num1, num2)
			}
			return AddToolResult{
				BaseLLMToolResult: llm.BaseLLMToolResult{
					ID: id,
				},
				Sum: num1 + num2,
			}, nil
		}),
	)
}

func TestSumAgent(t *testing.T) {
	// given
	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	addTool := createAddTool()
	calculatorAgent, err := agent.NewAgent(
		agent.WithName[Result]("calculator"),
		agent.WithLLMConfig[Result](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      apiKey,
			Model:       "gpt-4.1",
			Temperature: 0.0,
		}),
		agent.WithBehavior[Result]("You are a calculator agent. Use the add tool to calculate the sum of the two provided numbers. Return the result in the specified JSON format."),
		agent.WithTool[Result]("add", addTool),
		agent.WithToolLimit[Result]("add", 1),
		agent.WithOutputSchema(&Result{}),
	)
	require.NoError(t, err, "Failed to create agent")

	input := AddNumbers{
		Num1: 3,
		Num2: 5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// when
	result, err := calculatorAgent.Run(ctx, input)

	// then
	require.NoError(t, err, "Agent run should not fail")
	require.NotNil(t, result, "Result should not be nil")
	require.NotNil(t, result.Data, "Result data should not be nil")

	assert.Equal(t, 8, result.Data.Sum, "Sum should be 8 (3 + 5)")
	assert.NotEmpty(t, result.Messages, "Result should contain conversation messages")

	t.Logf("Test passed! Agent successfully calculated: %d + %d = %d",
		input.Num1, input.Num2, result.Data.Sum)
}
