package main

import (
	"context"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/vitalii-honchar/go-agent/pkg/goagent/agent"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
)

type (
	AddToolParams struct {
		Num1 float64 `json:"num1" jsonschema_description:"First number to add"`
		Num2 float64 `json:"num2" jsonschema_description:"Second number to add"`
	}

	AddToolResult struct {
		llm.BaseLLMToolResult

		Sum float64 `json:"sum" jsonschema_description:"Sum of the two numbers"`
	}

	IncrementResult struct {
		FinalNumber int      `json:"final_number" jsonschema_description:"Final result after all increments"`
		Steps       []string `json:"steps"        jsonschema_description:"List of steps taken to reach the final number"`
	}

	IncrementInput struct {
		StartNumber int `json:"start_number" jsonschema_description:"Starting number for increment"`
		Steps       int `json:"steps"        jsonschema_description:"Number of steps to increment"`
	}
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	addTool := createAddTool()
	loggingMiddleware := createLoggingMiddleware()

	incrementAgent, err := agent.NewAgent(
		agent.WithName[IncrementResult]("increment_agent"),
		agent.WithLLMConfig[IncrementResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      apiKey,
			Model:       "gpt-4.1",
			Temperature: 0.0,
		}),
		agent.WithBehavior[IncrementResult](`You are a increment agent. You must:
1. Start with the provided start_number
2. Use the add tool exactly 3 times to add 2 each time
3. Track each step and return final number`,
		),
		agent.WithTool[IncrementResult]("add", addTool),
		agent.WithToolLimit[IncrementResult]("add", 3),
		agent.WithMiddleware[IncrementResult](loggingMiddleware),
	)

	if err != nil {
		log.Fatalf("Failed to create increment-agent: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	input := IncrementInput{
		StartNumber: 100,
		Steps:       3,
	}

	log.Printf("Starting increment-agent...")
	result, err := incrementAgent.Run(ctx, input)
	if err != nil {
		log.Fatalf("Failed to run increment-agent: %v", err) //nolint:gocritic
	}
	log.Printf("Increment result: %+v", result.Data.FinalNumber) // 106
	log.Printf("Steps: %v", result.Data.Steps)
}

func createAddTool() llm.LLMTool {
	counter := new(int64)

	tool, err := llm.NewLLMTool(
		llm.WithLLMToolName("add"),
		llm.WithLLMToolDescription("Adds two numbers together"),
		llm.WithLLMToolParametersSchema[AddToolParams](),
		llm.WithLLMToolCall(func(callID string, params AddToolParams) (AddToolResult, error) {
			callCount := atomic.AddInt64(counter, 1)
			log.Printf("🔧 TOOL CALL #%d: add(num1=%v, num2=%v)", callCount, params.Num1, params.Num2)

			result := AddToolResult{
				BaseLLMToolResult: llm.BaseLLMToolResult{
					ID: callID,
				},
				Sum: params.Num1 + params.Num2,
			}
			log.Printf("🔧 TOOL RESULT #%d: add(num1=%v, num2=%v) = %v", callCount, params.Num1, params.Num2, result.Sum)

			return result, nil
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create add tool: %v", err)
	}

	return tool
}

func createLoggingMiddleware() agent.AgentMiddleware {
	return func(ctx context.Context, state *agent.AgentState, llmMessage llm.LLMMessage) (llm.LLMMessage, error) {
		log.Printf("🤖 LLM Message [%d chars]: %s", len(llmMessage.Content), truncateContent(llmMessage.Content, 200))

		if len(llmMessage.ToolCalls) > 0 {
			log.Printf("🔧 Tool Calls: %d", len(llmMessage.ToolCalls))
			for i, toolCall := range llmMessage.ToolCalls {
				log.Printf("   #%d: %s(%s)", i+1, toolCall.ToolName, truncateContent(toolCall.Args, 100))
			}
		}

		if len(llmMessage.ToolResults) > 0 {
			log.Printf("📋 Tool Results: %d", len(llmMessage.ToolResults))
		}

		if llmMessage.End {
			log.Printf("🏁 LLM Execution Complete")
		}

		return llmMessage, nil
	}
}

func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}

	return content[:maxLen] + "..."
}
