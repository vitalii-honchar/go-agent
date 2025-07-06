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

func createAddToolWithCounter(counter *int64) llm.LLMTool {
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
			callCount := atomic.AddInt64(counter, 1) // Thread-safe increment
			fmt.Printf("🔧 TOOL CALLED #%d: add(num1=%v, num2=%v)\n", callCount, args["num1"], args["num2"])
			
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
			fmt.Printf("🔧 TOOL RESULT #%d: %v\n", callCount, result.Sum)
			return result, nil
		}),
	)
}

func TestSumAgent(t *testing.T) {
	// Atomic counter for tool calls
	var toolCallCount int64
	
	// given
	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	addTool := createAddToolWithCounter(&toolCallCount)
	calculatorAgent, err := agent.NewAgent(
		agent.WithName[Result]("calculator"),
		agent.WithLLMConfig[Result](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      apiKey,
			Model:       "gpt-4.1",
			Temperature: 0.0,
		}),
		agent.WithBehavior[Result]("You are a calculator agent. You MUST use the add tool to calculate the sum of the two provided numbers. Do NOT calculate manually. Return the result in the specified JSON format."),
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
	
	// 🚨 CRITICAL: Verify tool was actually called using atomic counter
	finalCount := atomic.LoadInt64(&toolCallCount)
	assert.Greater(t, finalCount, int64(0), "Tool should have been called at least once")
	assert.Equal(t, int64(1), finalCount, "Tool should have been called exactly once")
	
	// Print message flow for debugging
	t.Logf("📝 Message flow (%d messages):", len(result.Messages))
	for i, msg := range result.Messages {
		t.Logf("  %d. %s: %s", i+1, msg.Type, msg.Content[:min(100, len(msg.Content))])
		if len(msg.ToolCalls) > 0 {
			t.Logf("     Tool calls: %d", len(msg.ToolCalls))
		}
		if len(msg.ToolResults) > 0 {
			t.Logf("     Tool results: %d", len(msg.ToolResults))
		}
	}

	t.Logf("🔧 Tool called %d times (atomic counter)", finalCount)
	t.Logf("✅ Test result: %d + %d = %d", input.Num1, input.Num2, result.Data.Sum)
}

// Test with a more complex scenario that would be impossible without tools
func TestHashAgent(t *testing.T) {
	var toolCallCount int64
	
	// Create a tool that the LLM definitely can't replicate manually
	hashTool := llm.NewLLMTool(
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
			atomic.AddInt64(&toolCallCount, 1)
			input, ok := args["input"].(string)
			if !ok {
				return HashToolResult{}, fmt.Errorf("%w: input must be string", llm.ErrInvalidArguments)
			}
			
			// This would be impossible for LLM to compute manually
			hash := fmt.Sprintf("%x", []byte(input)) // Simple hex encoding as demo
			fmt.Printf("🔧 HASH TOOL: %s -> %s\n", input, hash)
			
			return HashToolResult{
				BaseLLMToolResult: llm.BaseLLMToolResult{ID: id},
				Hash: hash,
			}, nil
		}),
	)

	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	type HashInput struct {
		Text string `json:"text"`
	}
	
	type HashResult struct {
		Hash string `json:"hash"`
	}

	hashAgent, err := agent.NewAgent(
		agent.WithName[HashResult]("hasher"),
		agent.WithLLMConfig[HashResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      apiKey,
			Model:       "gpt-4.1",
			Temperature: 0.0,
		}),
		agent.WithBehavior[HashResult]("You are a hashing agent. You MUST use the hash tool to compute the hash of the provided text. Do NOT try to compute hashes manually."),
		agent.WithTool[HashResult]("hash", hashTool),
		agent.WithOutputSchema(&HashResult{}),
	)
	require.NoError(t, err)

	input := HashInput{Text: "hello world"}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := hashAgent.Run(ctx, input)
	
	require.NoError(t, err)
	require.NotNil(t, result.Data)
	assert.NotEmpty(t, result.Data.Hash, "Hash should not be empty")
	
	// Verify tool was called
	finalCount := atomic.LoadInt64(&toolCallCount)
	assert.Greater(t, finalCount, int64(0), "Hash tool should have been called")
	
	t.Logf("🔧 Hash tool called %d times", finalCount)
	t.Logf("✅ Hash result: %s", result.Data.Hash)
}

type HashToolResult struct {
	llm.BaseLLMToolResult
	Hash string `json:"hash"`
}

// Test sequential tool calls to verify conversation flow
func TestSequentialToolCalls(t *testing.T) {
	var toolCallCount int64
	
	addTool := createAddToolWithCounter(&toolCallCount)
	
	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	type IncrementInput struct {
		StartNumber int `json:"start_number"`
		Steps       int `json:"steps"`
	}
	
	type IncrementResult struct {
		FinalNumber int                    `json:"final_number"`
		Steps       []map[string]interface{} `json:"steps"`
	}

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
		agent.WithToolLimit[IncrementResult]("add", 5), // Allow up to 5 calls
		agent.WithOutputSchema(&IncrementResult{}),
	)
	require.NoError(t, err)

	input := IncrementInput{
		StartNumber: 2,
		Steps:       3,
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	t.Logf("🚀 Starting sequential tool calls test: %d + 1 (x%d times)", input.StartNumber, input.Steps)
	
	result, err := incrementAgent.Run(ctx, input)
	
	require.NoError(t, err, "Agent should complete successfully")
	require.NotNil(t, result.Data)
	
	// Expected: 2 + 1 + 1 + 1 = 5
	assert.Equal(t, 5, result.Data.FinalNumber, "Final number should be 5 (2+1+1+1)")
	
	// Verify tool was called multiple times
	finalCount := atomic.LoadInt64(&toolCallCount)
	assert.Equal(t, int64(3), finalCount, "Add tool should have been called exactly 3 times")
	
	t.Logf("🔧 Tool called %d times", finalCount)
	t.Logf("✅ Final result: %d", result.Data.FinalNumber)
	t.Logf("📝 Steps taken: %+v", result.Data.Steps)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
