// Package agent provides a powerful, production-ready Go library for building AI agents
// with configurable behavior, custom tools, and type-safe output schemas.
//
// This library is perfect for building intelligent automation, data analysis tools,
// web scrapers, and AI-powered applications with robust error handling and clean architecture.
//
// Key Features:
//
//   - Type-safe agents with custom output schemas using Go generics
//   - Extensible tool system with automatic limit enforcement
//   - Configurable agent behavior using natural language
//   - Multiple LLM support (OpenAI, extensible to others)
//   - Structured JSON output with schema validation
//   - Production-ready with comprehensive error handling
//
// Quick Start:
//
// The easiest way to get started is to create a simple agent:
//
//	package main
//
//	import (
//		"context"
//		"fmt"
//		"log"
//
//		"github.com/vitalii-honchar/go-agent/pkg/goagent/agent"
//		"github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
//	)
//
//	type Result struct {
//		Answer string `json:"answer" jsonschema_description:"The answer"`
//	}
//
//	func main() {
//		agent, err := agent.NewAgent(
//			agent.WithName[Result]("my-agent"),
//			agent.WithLLMConfig[Result](llm.LLMConfig{
//				Type:        llm.LLMTypeOpenAI,
//				APIKey:      "your-openai-api-key",
//				Model:       "gpt-4",
//				Temperature: 0.0,
//			}),
//			agent.WithBehavior[Result]("You are a helpful assistant."),
//		)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		result, err := agent.Run(context.Background(), "What is 2+2?")
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		fmt.Println("Answer:", result.Data.Answer)
//	}
//
// Tool Integration:
//
// Agents can use custom tools to extend their capabilities:
//
//	type MyToolParams struct {
//		Input string `json:"input" jsonschema_description:"Input to process"`
//	}
//
//	type MyToolResult struct {
//		llm.BaseLLMToolResult
//		Output string `json:"output" jsonschema_description:"Processed output"`
//	}
//
//	tool := llm.NewLLMTool(
//		llm.WithLLMToolName("my-tool"),
//		llm.WithLLMToolDescription("Processes input and returns output"),
//		llm.WithLLMToolParametersSchema[MyToolParams](),
//		llm.WithLLMToolCall(func(callID string, params MyToolParams) (MyToolResult, error) {
//			// Process params.Input
//			return MyToolResult{
//				BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
//				Output:            processedOutput,
//			}, nil
//		}),
//	)
//
//	agent, err := agent.NewAgent(
//		// ... other options
//		agent.WithTool[MyResult]("my-tool", tool),
//		agent.WithToolLimit[MyResult]("my-tool", 5), // Limit tool usage
//	)
//
// Package Structure:
//
//   - agent: Core agent functionality and orchestration
//   - llm: LLM abstractions and tool system
//   - config: Environment-based configuration management
//
// For comprehensive examples and advanced usage, see:
// https://github.com/vitalii-honchar/go-agent/tree/main/examples
//
// Documentation and tutorials:
// https://vitaliihonchar.com/insights/go-ai-agent-library
//
// The package supports tool usage limits to prevent runaway execution and provides
// comprehensive error handling with typed errors for different failure scenarios.
package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/llmfactory"
)

var (
	// ErrLimitReached is returned when a tool usage limit is exceeded
	ErrLimitReached = errors.New("tool limit reached")
	// ErrToolError is returned when a tool call fails
	ErrToolError = errors.New("tool error occurred")
	// ErrLLMCall is returned when an LLM call fails
	ErrLLMCall = errors.New("LLM call error occurred")
	// ErrFinish is returned when LLM execution is finished
	ErrFinish = errors.New("LLM finished execution")
	// ErrToolNotFound is returned when a requested tool is not found
	ErrToolNotFound = errors.New("tool not found")
	// ErrInvalidResultSchema is returned when result validation fails
	ErrInvalidResultSchema = errors.New("invalid result schema")
	// ErrCannotCreateSchema is returned when schema creation fails
	ErrCannotCreateSchema = errors.New("cannot create schema from output type")
	// ErrEmptySystemPrompt is returned when system prompt is empty
	ErrEmptySystemPrompt = errors.New("system prompt cannot be empty")
)

var systemPromptTemplate = NewPrompt(`You are an agent that implements the ReAct ` +
	`(Reasoning-Action-Observation) pattern to solve tasks through systematic thinking and tool usage.

## REASONING PROTOCOL

Before EVERY action:
1. **THINK**: State your reasoning for the next step
2. **ACT**: Execute the appropriate tool with complete parameters
3. **OBSERVE**: Analyze the results and their implications

Always maintain explicit reasoning chains. Your thoughts should be visible and logical.

## EXECUTION CONTEXT

TOOLS AVAILABLE TO USE:
{{.tools}}

CURRENT TOOLS USAGE:
{{.tools_usage}}

TOOLS USAGE LIMITS:
{{.calling_limits}}

## AGENT BEHAVIOR

<BEHAVIOR>
{{.behavior}}
</BEHAVIOR>
`)

var outputPromptTemplate = NewPrompt(`Based on the entire conversation above, provide your final output.

Requirements:
- Synthesize all findings from your reasoning and observations
- Structure the output according to the required schema
- Include only factual information gathered during your analysis
- Ensure all required fields are populated with relevant data
- Output ONLY the JSON object with no additional text`)

// Agent represents a configurable AI agent with tools and behavior
type Agent[T any] struct {
	name             string
	llm              llm.LLM
	llmConfig        llm.LLMConfig
	tools            map[string]llm.LLMTool
	limits           map[string]int
	defaultToolLimit int
	outputSchema     *T
	systemPrompt     Prompt
	behavior         string
}

// AgentOption is a function that configures an Agent
type AgentOption[T any] func(*Agent[T])

// NewAgent creates a new Agent with the given options.
//
// The agent is configured using the options pattern for maximum flexibility.
// Required options include WithName and WithLLMConfig. Optional options include
// WithBehavior, WithTool, WithToolLimit, and WithSystemPrompt.
//
// Example:
//
//	type MyResult struct {
//		Answer   string `json:"answer" jsonschema_description:"The answer"`
//		Thoughts string `json:"thoughts" jsonschema_description:"Reasoning process"`
//	}
//
//	agent, err := NewAgent(
//		WithName[MyResult]("calculator-agent"),
//		WithLLMConfig[MyResult](llm.LLMConfig{
//			Type:        llm.LLMTypeOpenAI,
//			APIKey:      os.Getenv("OPENAI_API_KEY"),
//			Model:       "gpt-4",
//			Temperature: 0.1,
//		}),
//		WithBehavior[MyResult]("You are a precise calculator. Show your work."),
//		WithTool[MyResult]("add", addTool),
//		WithToolLimit[MyResult]("add", 5),
//	)
//
// The type parameter T specifies the expected output schema. The agent will
// validate that the LLM's response matches this schema before returning it.
//
// Returns an error if required options are missing or if the agent configuration
// is invalid (e.g., empty behavior, missing LLM config).
func NewAgent[T any](options ...AgentOption[T]) (*Agent[T], error) {
	agent := &Agent[T]{
		tools:            make(map[string]llm.LLMTool),
		limits:           make(map[string]int),
		defaultToolLimit: 3,
		systemPrompt:     systemPromptTemplate,
	}
	for _, opt := range options {
		opt(agent)
	}

	agentLLM, err := llmfactory.CreateLLM(agent.llmConfig, agent.tools)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM: %w", err)
	}

	agent.llm = agentLLM
	agent.outputSchema = new(T)

	return agent, nil
}

// WithName sets the agent's name
func WithName[T any](name string) AgentOption[T] {
	return func(a *Agent[T]) {
		a.name = name
	}
}

// WithLLMConfig sets the LLM configuration for the agent
func WithLLMConfig[T any](config llm.LLMConfig) AgentOption[T] {
	return func(a *Agent[T]) {
		a.llmConfig = config
	}
}

// WithBehavior sets the agent's behavior description
func WithBehavior[T any](behavior string) AgentOption[T] {
	return func(a *Agent[T]) {
		a.behavior = strings.TrimSpace(behavior)
	}
}

// WithSystemPrompt sets a custom system prompt template
func WithSystemPrompt[T any](prompt Prompt) AgentOption[T] {
	return func(a *Agent[T]) {
		a.systemPrompt = prompt
	}
}

// WithTool adds a tool to the agent
func WithTool[T any](name string, tool llm.LLMTool) AgentOption[T] {
	return func(a *Agent[T]) {
		a.tools[name] = tool
	}
}

// WithToolLimit sets a usage limit for a specific tool
func WithToolLimit[T any](name string, limit int) AgentOption[T] {
	return func(a *Agent[T]) {
		a.limits[name] = limit
	}
}

// WithDefaultToolLimit sets the default tool usage limit for all tools
func WithDefaultToolLimit[T any](limit int) AgentOption[T] {
	return func(a *Agent[T]) {
		a.defaultToolLimit = limit
	}
}

// AgentState represents the current state of agent execution
type AgentState struct {
	Messages []llm.LLMMessage
}

// AddMessage adds a message to the agent's conversation history
func (a *AgentState) AddMessage(msg llm.LLMMessage) {
	a.Messages = append(a.Messages, msg)
}

// GetToolLimit returns the usage limit for a specific tool
func (a *Agent[T]) GetToolLimit(name string) int {
	if limit, exists := a.limits[name]; exists {
		return limit
	}

	return a.defaultToolLimit
}

// Run executes the agent with the given input and returns the result
func (a *Agent[T]) Run(ctx context.Context, input any) (*AgentResult[T], error) {
	state, err := a.createInitState(input)
	if err != nil {
		return nil, err
	}

	usage := make(map[string]int)

	for {
		llmMessage, err := a.llm.Call(ctx, state.Messages)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrLLMCall, err)
		}

		if llmMessage.ToolCalls != nil {
			results, err := a.callTools(llmMessage, usage)
			if err != nil {
				if errors.Is(err, ErrLimitReached) {
					state.AddMessage(llmMessage)

					return &AgentResult[T]{
						Data:     nil,
						Messages: state.Messages,
					}, ErrLimitReached
				}

				return nil, fmt.Errorf("%w: %s", ErrToolError, err)
			}

			llmMessage.ToolResults = results
		}

		state.AddMessage(llmMessage)

		if llmMessage.End {
			return a.createResult(ctx, state)
		}

		newSystemPrompt, err := a.createSystemPrompt(usage)
		if err != nil {
			return nil, fmt.Errorf("failed to update system prompt: %w", err)
		}

		state.Messages[0].Content = newSystemPrompt
	}
}

func (a *Agent[T]) createInitState(input any) (*AgentState, error) {
	systemPrompt, err := a.createSystemPrompt(make(map[string]int))
	if err != nil {
		return nil, fmt.Errorf("failed to create system prompt: %w", err)
	}

	if systemPrompt == "" {
		return nil, ErrEmptySystemPrompt
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	return &AgentState{
		Messages: []llm.LLMMessage{
			llm.NewLLMMessage(llm.LLMMessageTypeSystem, systemPrompt),
			llm.NewLLMMessage(llm.LLMMessageTypeUser, string(inputJSON)),
		},
	}, nil
}

func (a *Agent[T]) createSystemPrompt(usage map[string]int) (string, error) {
	tools, err := json.Marshal(a.tools)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tools: %w", err)
	}

	toolsUsage, err := json.Marshal(usage)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tools usage: %w", err)
	}

	callingLimits, err := json.Marshal(a.limits)
	if err != nil {
		return "", fmt.Errorf("failed to marshal calling limits: %w", err)
	}

	return a.systemPrompt.Render(map[string]any{
		"tools":          string(tools),
		"tools_usage":    string(toolsUsage),
		"calling_limits": string(callingLimits),
		"behavior":       a.behavior,
	})
}

func (a *Agent[T]) callTools(llmMessage llm.LLMMessage, usage map[string]int) ([]llm.LLMToolResult, error) {
	results := make([]llm.LLMToolResult, 0, len(llmMessage.ToolCalls))

	for _, toolCall := range llmMessage.ToolCalls {
		tool, ok := a.tools[toolCall.ToolName]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrToolNotFound, toolCall.ToolName)
		}

		limit := a.GetToolLimit(toolCall.ToolName)
		if usage[toolCall.ToolName] >= limit {
			return nil, ErrLimitReached
		}

		toolRes, err := tool.Call(toolCall.ID, toolCall.Args)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrToolError, err)
		}

		usage[toolCall.ToolName]++

		results = append(results, toolRes)
	}

	return results, nil
}

func (a *Agent[T]) createResult(ctx context.Context, state *AgentState) (*AgentResult[T], error) {
	// Create output prompt with schema
	outputPrompt, err := outputPromptTemplate.Render(map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("failed to render output prompt: %w", err)
	}

	state.Messages = append(state.Messages, llm.NewLLMMessage(llm.LLMMessageTypeUser, outputPrompt))

	// Call LLM with structured output
	result, err := llm.CallWithStructuredOutput[T](ctx, a.llm, state.Messages)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrLLMCall, err)
	}

	return &AgentResult[T]{
		Data:     &result,
		Messages: state.Messages,
	}, nil
}
