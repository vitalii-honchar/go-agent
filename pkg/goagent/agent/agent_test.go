package agent_test

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"go-agent/pkg/goagent/agent"
	"go-agent/pkg/goagent/llm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSumAgent(t *testing.T) {
	// given
	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	toolCallCounter, addTool := createAddTool(t)
	calculatorAgent, err := agent.NewAgent(
		agent.WithName[AddNumbersResult]("calculator"),
		agent.WithLLMConfig[AddNumbersResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      apiKey,
			Model:       "gpt-4.1",
			Temperature: 0.0,
		}),
		agent.WithBehavior[AddNumbersResult]("You are a calculator agent. You MUST use the add tool to calculate the sum of the two provided numbers. Do NOT calculate manually. Return the result in the specified JSON format."),
		agent.WithTool[AddNumbersResult]("add", addTool),
		agent.WithToolLimit[AddNumbersResult]("add", 1),
		agent.WithOutputSchema(&AddNumbersResult{}),
	)
	require.NoError(t, err, "Failed to create agent")

	input := AddNumbers{
		Num1: 3,
		Num2: 5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// when
	result, err := calculatorAgent.Run(ctx, input)

	// then
	require.NoError(t, err, "Agent run should not fail")
	require.NotNil(t, result, "Result should not be nil")
	require.NotNil(t, result.Data, "Result data should not be nil")

	assert.Equal(t, 8, result.Data.Sum, "Sum should be 8 (3 + 5)")
	assert.NotEmpty(t, result.Messages, "Result should contain conversation messages")

	finalCount := atomic.LoadInt64(toolCallCounter)
	assert.Equal(t, int64(1), finalCount, "Tool should have been called exactly once")
}

func TestHashAgent(t *testing.T) {
	// given
	toolCallCounter, hashTool := createHashTool(t)

	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	hashAgent, err := agent.NewAgent(
		agent.WithName[HashResult]("hasher"),
		agent.WithLLMConfig[HashResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      apiKey,
			Model:       "gpt-4.1",
			Temperature: 0.0,
		}),
		agent.WithBehavior[HashResult](`You are a hashing agent. 
		You MUST use the hash tool to compute the hash of the provided text. 
		Do NOT try to compute hashes manually.`),
		agent.WithTool[HashResult]("hash", hashTool),
		agent.WithOutputSchema(&HashResult{}),
	)
	require.NoError(t, err)

	input := HashInput{Text: "hello world"}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// when
	result, err := hashAgent.Run(ctx, input)

	// then
	require.NoError(t, err)
	require.NotNil(t, result.Data)
	assert.NotEmpty(t, result.Data.Hash, "Hash should not be empty")

	// Verify tool was called
	finalCount := atomic.LoadInt64(toolCallCounter)
	assert.Greater(t, finalCount, int64(0), "Hash tool should have been called")

	t.Logf("🔧 Hash tool called %d times", finalCount)
	t.Logf("✅ Hash result: %s", result.Data.Hash)
}

func TestSequentialToolCalls(t *testing.T) {
	// given
	toolCallCounter, addTool := createAddTool(t)

	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	incrementAgent, err := agent.NewAgent(
		agent.WithName[IncrementResult]("incrementer"),
		agent.WithLLMConfig[IncrementResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      apiKey,
			Model:       "gpt-4.1",
			Temperature: 0.0,
		}),
		agent.WithBehavior[IncrementResult](`You are an incrementer agent. You must:
1. Start with the provided start_number
2. Use the add tool to add 1 to it (steps times)
3. For 3 steps: call add(start_number, 1), then add(result, 1), then add(result, 1)
4. Track each step showing the calculation
5. Return the final number and all steps taken

You MUST use the add tool for each increment. Do NOT calculate manually.`),
		agent.WithTool[IncrementResult]("add", addTool),
		agent.WithToolLimit[IncrementResult]("add", 5),
		agent.WithOutputSchema(&IncrementResult{}),
	)
	require.NoError(t, err)

	input := IncrementInput{
		StartNumber: 2,
		Steps:       3,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Logf("🚀 Starting sequential tool calls test: %d + 1 (x%d times)", input.StartNumber, input.Steps)

	// when
	result, err := incrementAgent.Run(ctx, input)

	// then
	require.NoError(t, err, "Agent should complete successfully")
	require.NotNil(t, result.Data)

	assert.Equal(t, 5, result.Data.FinalNumber, "Final number should be 5 (2+1+1+1)")

	finalCount := atomic.LoadInt64(toolCallCounter)
	assert.Equal(t, int64(3), finalCount, "Add tool should have been called exactly 3 times")

	t.Logf("🔧 Tool called %d times", finalCount)
	t.Logf("✅ Final result: %d", result.Data.FinalNumber)
	t.Logf("📝 Steps taken: %+v", result.Data.Steps)
}

type (
	AddNumbers struct {
		Num1 int `json:"num1"`
		Num2 int `json:"num2"`
	}

	AddNumbersResult struct {
		Sum int `json:"sum"`
	}

	AddToolResult struct {
		llm.BaseLLMToolResult
		Sum float64 `json:"sum"`
	}

	IncrementInput struct {
		StartNumber int `json:"start_number"`
		Steps       int `json:"steps"`
	}

	IncrementResult struct {
		FinalNumber int                      `json:"final_number"`
		Steps       []map[string]interface{} `json:"steps"`
	}

	HashToolResult struct {
		llm.BaseLLMToolResult
		Hash string `json:"hash"`
	}

	HashInput struct {
		Text string `json:"text"`
	}

	HashResult struct {
		Hash string `json:"hash"`
	}
)

func createAddTool(t *testing.T) (*int64, llm.LLMTool) {
	counter := new(int64)

	return counter, llm.NewLLMTool(
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
			callCount := atomic.AddInt64(counter, 1)
			t.Logf("🔧 TOOL CALL #%d: add(num1=%v, num2=%v)", callCount, args["num1"], args["num2"])

			num1, ok1 := args["num1"].(float64)
			num2, ok2 := args["num2"].(float64)

			if !ok1 || !ok2 {
				return AddToolResult{}, fmt.Errorf("%w: num1 = '%v', num2 = '%v'", llm.ErrInvalidArguments, num1, num2)
			}
			result := AddToolResult{
				BaseLLMToolResult: llm.BaseLLMToolResult{
					ID: id,
				},
				Sum: num1 + num2,
			}
			t.Logf("🔧 TOOL RESULT #%d: add(num1=%v, num2=%v) = %v", callCount, num1, num2, result.Sum)
			return result, nil
		}),
	)
}

func createHashTool(t *testing.T) (*int64, llm.LLMTool) {
	counter := new(int64)

	return counter, llm.NewLLMTool(
		llm.WithLLMToolName("hash"),
		llm.WithLLMToolDescription("Computes SHA256 hash of input string"),
		llm.WithLLMToolParametersSchema(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"input": map[string]any{"type": "string"},
			},
			"required": []string{"input"},
		}),
		llm.WithLLMToolCall(func(id string, args map[string]any) (HashToolResult, error) {
			callCount := atomic.AddInt64(counter, 1)
			t.Logf("🔧 TOOL CALL #%d: hash(input='%s')", callCount, args["input"])

			input, ok := args["input"].(string)
			if !ok {
				return HashToolResult{}, fmt.Errorf("%w: input must be string", llm.ErrInvalidArguments)
			}

			// This would be impossible for LLM to compute manually
			hash := fmt.Sprintf("%x", []byte(input)) // Simple hex encoding as demo
			t.Logf("🔧 TOOL RESULT #%d: hash(input='%s') = %s", callCount, input, hash)

			return HashToolResult{
				BaseLLMToolResult: llm.BaseLLMToolResult{ID: id},
				Hash:              hash,
			}, nil
		}),
	)
}
