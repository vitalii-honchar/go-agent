package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reddit-analyzer/internal/agent/llm"

	"github.com/invopop/jsonschema"
	"github.com/xeipuuv/gojsonschema"
)

var (
	ErrLimitReached        = errors.New("tool limit reached")
	ErrToolError           = errors.New("tool error occurred")
	ErrLLMCall             = errors.New("LLM call error occurred")
	ErrFinish              = errors.New("LLM finished execution")
	ErrToolNotFound        = errors.New("tool not found")
	ErrInvalidResultSchema = errors.New("invalid result schema")
	ErrCannotCreateSchema  = errors.New("cannot create schema from output type")
	ErrEmptySystemPrompt   = errors.New("system prompt cannot be empty")
)

var systemPromptTemplate = NewPrompt(`You are an agent that should act as specified in escaped content <BEHAVIOR></BEHAVIOR>.
At the end of execution when you will be read to finish, you should return a JSON object that matches the output schema.

TOOLS AVAILABLE TO USE:
{{.tools}}

CURRENT TOOLS USAGE:
{{.tools_usage}}

TOOLS USAGE LIMITS:
{{.calling_limits}}

OUTPUT SCHEMA:
{{.output_schema}}

<BEHAVIOR>
{{.behavior}}
</BEHAVIOR>
`)

type Agent[T any] struct {
	name         string
	llm          llm.LLM
	llmConfig    llm.LLMConfig
	tools        map[string]llm.LLMTool
	limits       map[string]int
	outputSchema *T
	systemPrompt Prompt
	behavior     string
	schemaLoader gojsonschema.JSONLoader
}

type AgentOption[T any] func(*Agent[T])

func NewAgent[T any](options ...AgentOption[T]) (*Agent[T], error) {
	agent := &Agent[T]{
		tools:        make(map[string]llm.LLMTool),
		limits:       make(map[string]int),
		systemPrompt: systemPromptTemplate,
	}
	for _, opt := range options {
		opt(agent)
	}

	agentLLM, err := llm.CreateLLM(agent.llmConfig, agent.tools)
	if err != nil {
		return nil, err
	}
	agent.llm = agentLLM

	return agent, nil
}

func WithName[T any](name string) AgentOption[T] {
	return func(a *Agent[T]) {
		a.name = name
	}
}

func WithLLMConfig[T any](config llm.LLMConfig) AgentOption[T] {
	return func(a *Agent[T]) {
		a.llmConfig = config
	}
}

func WithBehavior[T any](behavior string) AgentOption[T] {
	return func(a *Agent[T]) {
		a.behavior = behavior
	}
}

func WithOutputSchema[T any](schema *T) AgentOption[T] {
	return func(a *Agent[T]) {
		a.outputSchema = schema
		// Pre-compile schema for validation
		reflectedSchema := jsonschema.Reflect(schema)
		schemaBytes, err := json.Marshal(reflectedSchema)
		if err == nil {
			a.schemaLoader = gojsonschema.NewStringLoader(string(schemaBytes))
		}
	}
}

func WithSystemPrompt[T any](prompt Prompt) AgentOption[T] {
	return func(a *Agent[T]) {
		a.systemPrompt = prompt
	}
}

func WithTool[T any](name string, tool llm.LLMTool) AgentOption[T] {
	return func(a *Agent[T]) {
		a.tools[name] = tool
	}
}

func WithToolLimit[T any](name string, limit int) AgentOption[T] {
	return func(a *Agent[T]) {
		a.limits[name] = limit
	}
}

type AgentState struct {
	Messages []llm.LLMMessage
}

func (a *AgentState) AddMessage(msg llm.LLMMessage) {
	a.Messages = append(a.Messages, msg)
}

func (a *Agent[T]) Run(ctx context.Context, input any) (*AgentResult[T], error) {
	state, err := a.createInitState(input)
	if err != nil {
		return nil, err
	}
	usage := make(map[string]int)

	for {
		// if a.isLimitReached(usage) {
		// 	res, err := a.createResult(state)
		// 	if err != nil {
		// 		return nil, fmt.Errorf("%w: %s", ErrLimitReached, err)
		// 	}
		// 	return res, ErrLimitReached
		// }

		llmMessage, err := a.llm.Call(ctx, state.Messages)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrLLMCall, err)
		}

		if llmMessage.ToolCalls != nil {
			results, err := a.callTools(llmMessage, usage)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", ErrToolError, err)
			}
			llmMessage.ToolResults = results
		}

		state.AddMessage(llmMessage)

		if llmMessage.End {
			return a.createResult(state)
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

	inputJson, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	return &AgentState{
		Messages: []llm.LLMMessage{
			llm.NewLLMMessage(llm.LLMMessageTypeSystem, systemPrompt),
			llm.NewLLMMessage(llm.LLMMessageTypeUser, string(inputJson)),
		},
	}, nil
}

func (a *Agent[T]) createSystemPrompt(usage map[string]int) (string, error) {
	schema := jsonschema.Reflect(a.outputSchema)
	outputSchema, err := json.Marshal(schema)
	if err != nil {
		return "", err
	}

	tools, err := json.Marshal(a.tools)
	if err != nil {
		return "", err
	}

	toolsUsage, err := json.Marshal(usage)
	if err != nil {
		return "", err
	}

	callingLimits, err := json.Marshal(a.limits)
	if err != nil {
		return "", err
	}

	return a.systemPrompt.Render(map[string]any{
		"tools":          string(tools),
		"tools_usage":    string(toolsUsage),
		"calling_limits": string(callingLimits),
		"output_schema":  string(outputSchema),
		"behavior":       a.behavior,
	})
}

func (a *Agent[T]) callTools(llmMessage llm.LLMMessage, usage map[string]int) ([]llm.LLMToolResult, error) {
	var results []llm.LLMToolResult
	for _, toolCall := range llmMessage.ToolCalls {
		tool, ok := a.tools[toolCall.ToolName]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrToolNotFound, toolCall.ToolName)
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

func (a *Agent[T]) createResult(state *AgentState) (*AgentResult[T], error) {
	dataLoader := gojsonschema.NewStringLoader(state.Messages[len(state.Messages)-1].Content)
	validationRes, err := gojsonschema.Validate(a.schemaLoader, dataLoader)

	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidResultSchema, err)
	}
	if !validationRes.Valid() {
		return nil, fmt.Errorf("%w: %s", ErrInvalidResultSchema, validationRes.Errors())
	}

	var data T

	if err := json.Unmarshal([]byte(state.Messages[len(state.Messages)-1].Content), &data); err != nil {
		return nil, err
	}

	return &AgentResult[T]{
		Data:     &data,
		Messages: state.Messages,
	}, nil
}

func (a *Agent[T]) isLimitReached(usage map[string]int) bool {
	limitReached := make(map[string]bool)
	for toolName, limit := range a.limits {
		if usage, exists := usage[toolName]; exists && usage >= limit {
			limitReached[toolName] = true
		}
	}
	return len(limitReached) == len(a.limits)
}
