package agent_test

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vitalii-honchar/go-agent/pkg/goagent/agent"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/llm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSumAgent(t *testing.T) {
	t.Parallel()
	// given
	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	toolCallCounter, addTool := createAddTool(t)
	calculatorAgent, err := agent.NewAgent(
		agent.WithName[AddNumbersResult]("calculator_agent"),
		agent.WithLLMConfig[AddNumbersResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      apiKey,
			Model:       "gpt-4.1",
			Temperature: 0.0,
		}),
		agent.WithBehavior[AddNumbersResult](
			"You are a calculator agent. You MUST use the add tool to calculate the sum of the two provided numbers. "+
				"Do NOT calculate manually. Return the result in the specified JSON format."),
		agent.WithTool[AddNumbersResult]("add", addTool),
		agent.WithToolLimit[AddNumbersResult]("add", 1),
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
	t.Parallel()
	// given
	toolCallCounter, hashTool := createHashTool(t)

	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	hashAgent, err := agent.NewAgent(
		agent.WithName[HashResult]("hash_agent"),
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
	assert.Positive(t, finalCount, "Hash tool should have been called")

	t.Logf("🔧 Hash tool called %d times", finalCount)
	t.Logf("✅ Hash result: %s", result.Data.Hash)
}

func TestSequentialToolCalls(t *testing.T) {
	t.Parallel()
	// given
	toolCallCounter, addTool := createAddTool(t)

	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	incrementAgent, err := agent.NewAgent(
		agent.WithName[IncrementResult]("increment_agent"),
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

func TestToolLimitReached(t *testing.T) {
	t.Parallel()
	t.Skip()

	toolCallCounter, addTool := createAddTool(t)
	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	behavior := `You are a calculation agent that can ONLY perform arithmetic using the add tool. ` +
		`You have no ability to calculate numbers manually.

Your task: You need to increment the start_number by adding complex floating point numbers that you cannot ` +
		`calculate yourself.

Process:
1. Use add tool to add start_number + 0.12345
2. Use add tool to add result + 0.23456  
3. Use add tool to add result + 0.34567
4. Use add tool to add result + 0.45678
5. Use add tool to add result + 0.56789

You MUST use the add tool for each step because these floating point calculations are too complex for you to do ` +
		`manually. 
Do not try to calculate yourself - you will get wrong results. Always use the add tool.

Continue making these precise floating point additions until you have made at least 3 tool calls.`

	limitTestAgent, err := agent.NewAgent(
		agent.WithName[IncrementResult]("limit_test_agent"),
		agent.WithLLMConfig[IncrementResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      apiKey,
			Model:       "gpt-4.1",
			Temperature: 0.0,
		}),
		agent.WithBehavior[IncrementResult](behavior),
		agent.WithTool[IncrementResult]("add", addTool),
		agent.WithToolLimit[IncrementResult]("add", 1),
	)
	require.NoError(t, err)

	input := IncrementInput{
		StartNumber: 100,
		Steps:       3,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Logf("🚀 Starting tool limit test: %d + floating point numbers (x%d times) with limit of 1",
		input.StartNumber, input.Steps)

	result, err := limitTestAgent.Run(ctx, input)

	verifyLimitReached(t, result, err, toolCallCounter)
	t.Logf("📝 Messages count: %d", len(result.Messages))
}

func verifyLimitReached(t *testing.T, result *agent.AgentResult[IncrementResult], err error, toolCallCounter *int64) {
	require.ErrorIs(t, err, agent.ErrLimitReached, "Agent should return ErrLimitReached")
	require.NotNil(t, result, "Result should not be nil even when limit reached")
	require.Nil(t, result.Data, "Result data should be nil when limit reached")
	require.NotEmpty(t, result.Messages, "Result should contain conversation messages")

	finalCount := atomic.LoadInt64(toolCallCounter)
	assert.Equal(t, int64(1), finalCount, "Add tool should have been called exactly 1 time before hitting limit")

	t.Logf("🔧 Tool called %d times (limit: 1)", finalCount)
	t.Logf("✅ Limit reached as expected with error: %v", err)
}

func TestMultiToolLimitReached(t *testing.T) {
	t.Parallel()
	// given
	addToolCallCounter, addTool := createAddTool(t)
	hashToolCallCounter, hashTool := createHashTool(t)

	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	multiToolAgent, err := agent.NewAgent(
		agent.WithName[IncrementResult]("multi_tool_agent"),
		agent.WithLLMConfig[IncrementResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      apiKey,
			Model:       "gpt-4.1",
			Temperature: 0.0,
		}),
		agent.WithBehavior[IncrementResult](`You are a multi-tool testing agent. You must:
1. Start with the provided start_number (10)
2. Use the add tool exactly 3 times to add 2 each time: add(10,2), add(12,2), add(14,2) 
3. Then use the hash tool to compute hash of "test"
4. Track each step and return final number

You have add tool limit of 3 and hash tool limit of 1. Use add tool 3 times, then hash tool 1 time.`),
		agent.WithTool[IncrementResult]("add", addTool),
		agent.WithTool[IncrementResult]("hash", hashTool),
		agent.WithToolLimit[IncrementResult]("add", 3),
		agent.WithToolLimit[IncrementResult]("hash", 1),
	)
	require.NoError(t, err)

	input := IncrementInput{
		StartNumber: 10,
		Steps:       3,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Logf("🚀 Starting multi-tool test: add limit=3, hash limit=1")

	// when
	result, err := multiToolAgent.Run(ctx, input)

	// then
	require.NoError(t, err, "Agent should complete successfully using both tools within limits")
	require.NotNil(t, result, "Result should not be nil")
	require.NotNil(t, result.Data, "Result data should not be nil")
	require.NotEmpty(t, result.Messages, "Result should contain conversation messages")

	addCallCount := atomic.LoadInt64(addToolCallCounter)
	hashCallCount := atomic.LoadInt64(hashToolCallCounter)

	assert.Equal(t, int64(3), addCallCount, "Add tool should have been called exactly 3 times")
	assert.Equal(t, int64(1), hashCallCount, "Hash tool should have been called exactly 1 time")
	assert.Equal(t, 16, result.Data.FinalNumber, "Final number should be 16 (10+2+2+2)")

	t.Logf("🔧 Add tool called %d times (limit: 3)", addCallCount)
	t.Logf("🔧 Hash tool called %d times (limit: 1)", hashCallCount)
	t.Logf("✅ Final result: %d", result.Data.FinalNumber)
	t.Logf("📝 Messages count: %d", len(result.Messages))
}

func TestDefaultToolLimit(t *testing.T) {
	t.Parallel()
	// given
	toolCallCounter, addTool := createAddTool(t)

	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	// Agent with no explicit tool limits - should use default limit of 3
	defaultLimitAgent, err := agent.NewAgent(
		agent.WithName[IncrementResult]("default_limit_agent"),
		agent.WithLLMConfig[IncrementResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      apiKey,
			Model:       "gpt-4.1",
			Temperature: 0.0,
		}),
		agent.WithBehavior[IncrementResult](`You are a default limit testing agent. You must:
1. Start with the provided start_number (5)
2. Use the add tool to add 1 three times: add(5,1), add(6,1), add(7,1)
3. Track each step and return final number

Use the add tool exactly 3 times.`),
		agent.WithTool[IncrementResult]("add", addTool),
		// Note: No explicit tool limit set - should use default of 3
	)
	require.NoError(t, err)

	input := IncrementInput{
		StartNumber: 5,
		Steps:       3,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Logf("🚀 Starting default limit test: no explicit limit (should default to 3)")

	// when
	result, err := defaultLimitAgent.Run(ctx, input)

	// then
	require.NoError(t, err, "Agent should complete successfully with default limit")
	require.NotNil(t, result, "Result should not be nil")
	require.NotNil(t, result.Data, "Result data should not be nil")

	finalCount := atomic.LoadInt64(toolCallCounter)
	assert.Equal(t, int64(3), finalCount, "Add tool should have been called exactly 3 times (default limit)")
	assert.Equal(t, 8, result.Data.FinalNumber, "Final number should be 8 (5+1+1+1)")

	t.Logf("🔧 Tool called %d times (default limit: 3)", finalCount)
	t.Logf("✅ Final result: %d", result.Data.FinalNumber)
}

func TestCustomDefaultToolLimit(t *testing.T) {
	t.Parallel()
	// given
	toolCallCounter, addTool := createAddTool(t)

	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	// Agent with custom default tool limit of 2
	customDefaultAgent, err := agent.NewAgent(
		agent.WithName[IncrementResult]("custom_default_agent"),
		agent.WithLLMConfig[IncrementResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      apiKey,
			Model:       "gpt-4.1",
			Temperature: 0.0,
		}),
		agent.WithBehavior[IncrementResult](`You are a custom default limit testing agent. You must:
1. Start with the provided start_number (10)
2. Use the add tool to add 5 exactly 2 times: add(10,5), add(15,5)
3. Track each step and return final number

Use the add tool exactly 2 times.`),
		agent.WithTool[IncrementResult]("add", addTool),
		agent.WithDefaultToolLimit[IncrementResult](2), // Custom default limit
		// Note: No explicit tool limit set - should use custom default of 2
	)
	require.NoError(t, err)

	input := IncrementInput{
		StartNumber: 10,
		Steps:       2,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Logf("🚀 Starting custom default limit test: custom default limit of 2")

	// when
	result, err := customDefaultAgent.Run(ctx, input)

	// then
	require.NoError(t, err, "Agent should complete successfully with custom default limit")
	require.NotNil(t, result, "Result should not be nil")
	require.NotNil(t, result.Data, "Result data should not be nil")

	finalCount := atomic.LoadInt64(toolCallCounter)
	assert.Equal(t, int64(2), finalCount, "Add tool should have been called exactly 2 times (custom default limit)")
	assert.Equal(t, 20, result.Data.FinalNumber, "Final number should be 20 (10+5+5)")

	t.Logf("🔧 Tool called %d times (custom default limit: 2)", finalCount)
	t.Logf("✅ Final result: %d", result.Data.FinalNumber)
}

type (
	AddNumbers struct {
		Num1 int `json:"num1" jsonschema_description:"First number to add"`
		Num2 int `json:"num2" jsonschema_description:"Second number to add"`
	}

	AddNumbersResult struct {
		Sum int `json:"sum" jsonschema_description:"Sum of the two numbers"`
	}

	AddToolResult struct {
		llm.BaseLLMToolResult

		Sum float64 `json:"sum" jsonschema_description:"Sum of the two numbers"`
	}

	IncrementInput struct {
		StartNumber int `json:"start_number" jsonschema_description:"Starting number for increment"`
		Steps       int `json:"steps"        jsonschema_description:"Number of steps to increment"`
	}

	IncrementResult struct {
		FinalNumber int      `json:"final_number" jsonschema_description:"Final result after all increments"`
		Steps       []string `json:"steps"        jsonschema_description:"List of steps taken to reach the final number"`
	}

	HashToolResult struct {
		llm.BaseLLMToolResult

		Hash string `json:"hash" jsonschema_description:"SHA256 hash of the input text"`
	}

	HashInput struct {
		Text string `json:"text" jsonschema_description:"Text to hash"`
	}

	HashResult struct {
		Hash string `json:"hash" jsonschema_description:"Hex hash of the input text"`
	}
)

type AddToolParams struct {
	Num1 float64 `json:"num1"`
	Num2 float64 `json:"num2"`
}

func createAddTool(t *testing.T) (*int64, llm.LLMTool) {
	counter := new(int64)

	tool, err := llm.NewLLMTool(
		llm.WithLLMToolName("add"),
		llm.WithLLMToolDescription("Adds two numbers together"),
		llm.WithLLMToolParametersSchema[AddToolParams](),
		llm.WithLLMToolCall(func(callID string, params AddToolParams) (AddToolResult, error) {
			callCount := atomic.AddInt64(counter, 1)
			t.Logf("🔧 TOOL CALL #%d: add(num1=%v, num2=%v)", callCount, params.Num1, params.Num2)

			result := AddToolResult{
				BaseLLMToolResult: llm.BaseLLMToolResult{
					ID: callID,
				},
				Sum: params.Num1 + params.Num2,
			}
			t.Logf("🔧 TOOL RESULT #%d: add(num1=%v, num2=%v) = %v", callCount, params.Num1, params.Num2, result.Sum)

			return result, nil
		}),
	)
	require.NoError(t, err, "Failed to create add tool")

	return counter, tool
}

type HashToolParams struct {
	Input string `json:"input"`
}

func createHashTool(t *testing.T) (*int64, llm.LLMTool) {
	counter := new(int64)

	tool, err := llm.NewLLMTool(
		llm.WithLLMToolName("hash"),
		llm.WithLLMToolDescription("Computes SHA256 hash of input string"),
		llm.WithLLMToolParametersSchema[HashToolParams](),
		llm.WithLLMToolCall(
			func(callID string, params HashToolParams) (HashToolResult, error) {
				callCount := atomic.AddInt64(counter, 1)
				t.Logf("🔧 TOOL CALL #%d: hash(input='%s')", callCount, params.Input)

				// This would be impossible for LLM to compute manually
				hash := fmt.Sprintf("%x", []byte(params.Input)) // Simple hex encoding as demo
				t.Logf("🔧 TOOL RESULT #%d: hash(input='%s') = %s", callCount, params.Input, hash)

				return HashToolResult{
					BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
					Hash:              hash,
				}, nil
			}),
	)
	require.NoError(t, err, "Failed to create hash tool")

	return counter, tool
}

func TestWithSystemPrompt(t *testing.T) {
	t.Parallel()

	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	customPrompt := agent.NewPrompt("You are a custom agent with special behavior: {{.Behavior}}")

	testAgent, err := agent.NewAgent(
		agent.WithName[AddNumbersResult]("custom_prompt_agent"),
		agent.WithLLMConfig[AddNumbersResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      apiKey,
			Model:       "gpt-4o-mini",
			Temperature: 0.0,
		}),
		agent.WithBehavior[AddNumbersResult]("Add numbers using the add tool"),
		agent.WithSystemPrompt[AddNumbersResult](customPrompt),
		agent.WithTool[AddNumbersResult]("add", createTestAddTool()),
	)
	require.NoError(t, err)

	input := AddNumbers{Num1: 1, Num2: 2}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := testAgent.Run(ctx, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Data)
	assert.Equal(t, 3, result.Data.Sum)
}

func TestNewAgentResult(t *testing.T) {
	t.Parallel()
	data := &AddNumbersResult{Sum: 10}
	messages := []llm.LLMMessage{
		llm.NewLLMMessage(llm.LLMMessageTypeUser, "Test message"),
	}

	result, err := agent.NewAgentResult(data, messages)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, data, result.Data)
	assert.Equal(t, messages, result.Messages)
}

func TestNewAgentResult_NilData(t *testing.T) {
	t.Parallel()
	messages := []llm.LLMMessage{
		llm.NewLLMMessage(llm.LLMMessageTypeUser, "Test message"),
	}

	result, err := agent.NewAgentResult[AddNumbersResult](nil, messages)

	require.Error(t, err)
	require.ErrorIs(t, err, agent.ErrInvalidResultSchema)
	assert.Nil(t, result)
}

func TestNewAgentResult_EmptyMessages(t *testing.T) {
	t.Parallel()
	data := &AddNumbersResult{Sum: 10}

	result, err := agent.NewAgentResult(data, []llm.LLMMessage{})

	require.Error(t, err)
	require.ErrorIs(t, err, agent.ErrInvalidResultSchema)
	assert.Nil(t, result)
}

func createTestAddTool() llm.LLMTool {
	tool, err := llm.NewLLMTool(
		llm.WithLLMToolName("add"),
		llm.WithLLMToolDescription("Adds two numbers"),
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

func TestMiddlewareLogging(t *testing.T) {
	t.Parallel()

	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	// Create a logging middleware that tracks LLM calls
	var logEntries []string
	loggingMiddleware := func(ctx context.Context, state *agent.AgentState, llmMessage llm.LLMMessage) (llm.LLMMessage, error) {
		logEntries = append(logEntries, fmt.Sprintf("LLM Response: content_length=%d, tool_calls=%d",
			len(llmMessage.Content), len(llmMessage.ToolCalls)))

		return llmMessage, nil
	}

	toolCallCounter, addTool := createAddTool(t)
	loggingAgent, err := agent.NewAgent(
		agent.WithName[AddNumbersResult]("logging_agent"),
		agent.WithLLMConfig[AddNumbersResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      apiKey,
			Model:       "gpt-4o-mini",
			Temperature: 0.0,
		}),
		agent.WithBehavior[AddNumbersResult]("You are a calculator. Use the add tool to calculate 5+3."),
		agent.WithTool[AddNumbersResult]("add", addTool),
		agent.WithMiddleware[AddNumbersResult](loggingMiddleware),
	)
	require.NoError(t, err)

	input := AddNumbers{Num1: 5, Num2: 3}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Logf("🚀 Starting logging middleware test")

	result, err := loggingAgent.Run(ctx, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Data)
	assert.Equal(t, 8, result.Data.Sum)

	// Verify logging middleware was called
	assert.NotEmpty(t, logEntries, "Logging middleware should have recorded entries")
	assert.NotEmpty(t, logEntries, "Should have at least one log entry")

	finalCount := atomic.LoadInt64(toolCallCounter)
	assert.Equal(t, int64(1), finalCount, "Add tool should have been called exactly once")

	t.Logf("🔧 Tool called %d times", finalCount)
	t.Logf("📝 Log entries: %v", logEntries)
	t.Logf("✅ Final result: %d", result.Data.Sum)
}

func TestMiddlewareRBAC(t *testing.T) {
	t.Parallel()

	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	// Create an RBAC middleware that blocks certain operations
	rbacMiddleware := func(ctx context.Context, state *agent.AgentState, llmMessage llm.LLMMessage) (llm.LLMMessage, error) {
		// Simulate checking for forbidden operations
		if len(llmMessage.ToolCalls) > 0 {
			for _, toolCall := range llmMessage.ToolCalls {
				if toolCall.ToolName == "add" {
					// Simulate RBAC rejection
					return llm.LLMMessage{}, agent.ErrAccessDenied
				}
			}
		}

		return llmMessage, nil
	}

	toolCallCounter, addTool := createAddTool(t)
	rbacAgent, err := agent.NewAgent(
		agent.WithName[AddNumbersResult]("rbac_agent"),
		agent.WithLLMConfig[AddNumbersResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      apiKey,
			Model:       "gpt-4o-mini",
			Temperature: 0.0,
		}),
		agent.WithBehavior[AddNumbersResult]("You are a calculator. Use the add tool to calculate 2+2."),
		agent.WithTool[AddNumbersResult]("add", addTool),
		agent.WithMiddleware[AddNumbersResult](rbacMiddleware),
	)
	require.NoError(t, err)

	input := AddNumbers{Num1: 2, Num2: 2}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Logf("🚀 Starting RBAC middleware test (should fail with access denied)")

	result, err := rbacAgent.Run(ctx, input)

	// Should fail due to RBAC middleware blocking access
	require.Error(t, err)
	require.ErrorIs(t, err, agent.ErrMiddlewareError)
	assert.Contains(t, err.Error(), "access denied")
	assert.Nil(t, result)

	// Tool should not have been called due to middleware blocking
	finalCount := atomic.LoadInt64(toolCallCounter)
	assert.Equal(t, int64(0), finalCount, "Add tool should not have been called due to RBAC block")

	t.Logf("🔧 Tool called %d times (blocked by RBAC)", finalCount)
	t.Logf("✅ RBAC correctly blocked access: %v", err)
}

func TestMiddlewareMessageModification(t *testing.T) {
	t.Parallel()

	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	// Create a middleware that modifies LLM responses
	var originalContent string
	var modifiedContent string
	modificationMiddleware := func(ctx context.Context, state *agent.AgentState, llmMessage llm.LLMMessage) (llm.LLMMessage, error) {
		originalContent = llmMessage.Content
		// Add a prefix to the LLM response content
		llmMessage.Content = "[MODIFIED] " + llmMessage.Content
		modifiedContent = llmMessage.Content

		return llmMessage, nil
	}

	toolCallCounter, addTool := createAddTool(t)
	modificationAgent, err := agent.NewAgent(
		agent.WithName[AddNumbersResult]("modification_agent"),
		agent.WithLLMConfig[AddNumbersResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      apiKey,
			Model:       "gpt-4o-mini",
			Temperature: 0.0,
		}),
		agent.WithBehavior[AddNumbersResult]("You are a calculator. Use the add tool to calculate 1+1."),
		agent.WithTool[AddNumbersResult]("add", addTool),
		agent.WithMiddleware[AddNumbersResult](modificationMiddleware),
	)
	require.NoError(t, err)

	input := AddNumbers{Num1: 1, Num2: 1}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Logf("🚀 Starting message modification middleware test")

	result, err := modificationAgent.Run(ctx, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Data)
	assert.Equal(t, 2, result.Data.Sum)

	// Verify message was modified
	assert.NotEmpty(t, originalContent, "Should have captured original content")
	assert.NotEmpty(t, modifiedContent, "Should have captured modified content")
	assert.Contains(t, modifiedContent, "[MODIFIED]", "Modified content should contain prefix")
	assert.NotEqual(t, originalContent, modifiedContent, "Content should have been modified")

	finalCount := atomic.LoadInt64(toolCallCounter)
	assert.Equal(t, int64(1), finalCount, "Add tool should have been called exactly once")

	t.Logf("🔧 Tool called %d times", finalCount)
	t.Logf("📝 Original content length: %d", len(originalContent))
	t.Logf("📝 Modified content length: %d", len(modifiedContent))
	t.Logf("✅ Final result: %d", result.Data.Sum)
}

func TestMiddlewareChaining(t *testing.T) {
	t.Parallel()

	apiKey := os.Getenv("OPENAI_API_KEY")
	require.NotEmpty(t, apiKey, "OPENAI_API_KEY environment variable must be set")

	// Create multiple middleware that modify content in sequence
	var executionOrder []string

	firstMiddleware := func(ctx context.Context, state *agent.AgentState, llmMessage llm.LLMMessage) (llm.LLMMessage, error) {
		executionOrder = append(executionOrder, "first")
		llmMessage.Content = "[FIRST] " + llmMessage.Content
		return llmMessage, nil
	}

	secondMiddleware := func(ctx context.Context, state *agent.AgentState, llmMessage llm.LLMMessage) (llm.LLMMessage, error) {
		executionOrder = append(executionOrder, "second")
		llmMessage.Content = "[SECOND] " + llmMessage.Content
		return llmMessage, nil
	}

	thirdMiddleware := func(ctx context.Context, state *agent.AgentState, llmMessage llm.LLMMessage) (llm.LLMMessage, error) {
		executionOrder = append(executionOrder, "third")
		llmMessage.Content = "[THIRD] " + llmMessage.Content
		return llmMessage, nil
	}

	toolCallCounter, addTool := createAddTool(t)
	chainingAgent, err := agent.NewAgent(
		agent.WithName[AddNumbersResult]("chaining_agent"),
		agent.WithLLMConfig[AddNumbersResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      apiKey,
			Model:       "gpt-4o-mini",
			Temperature: 0.0,
		}),
		agent.WithBehavior[AddNumbersResult]("You are a calculator. Use the add tool to calculate 4+4."),
		agent.WithTool[AddNumbersResult]("add", addTool),
		agent.WithMiddleware[AddNumbersResult](firstMiddleware),
		agent.WithMiddleware[AddNumbersResult](secondMiddleware),
		agent.WithMiddleware[AddNumbersResult](thirdMiddleware),
	)
	require.NoError(t, err)

	input := AddNumbers{Num1: 4, Num2: 4}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Logf("🚀 Starting middleware chaining test")

	result, err := chainingAgent.Run(ctx, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Data)
	assert.Equal(t, 8, result.Data.Sum)

	// Verify middleware execution order
	assert.NotEmpty(t, executionOrder, "Should have recorded middleware execution order")
	// Middleware should execute in the order they were added
	assert.Contains(t, executionOrder, "first", "First middleware should have executed")
	assert.Contains(t, executionOrder, "second", "Second middleware should have executed")
	assert.Contains(t, executionOrder, "third", "Third middleware should have executed")

	finalCount := atomic.LoadInt64(toolCallCounter)
	assert.Equal(t, int64(1), finalCount, "Add tool should have been called exactly once")

	t.Logf("🔧 Tool called %d times", finalCount)
	t.Logf("📝 Middleware execution order: %v", executionOrder)
	t.Logf("✅ Final result: %d", result.Data.Sum)
}
