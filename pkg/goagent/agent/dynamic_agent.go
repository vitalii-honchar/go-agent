package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vitalii-honchar/go-agent/internal/validation"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/schema"
)

const defaultDynamicMaxSteps = 8

type DynamicAgentConfig struct {
	Name            string
	Behavior        string
	LLM             llm.DynamicLLM
	Tools           []llm.DynamicTool
	OutputSchemaRaw json.RawMessage
	MaxSteps        int
}

type DynamicAgent struct {
	name            string
	behavior        string
	provider        llm.DynamicLLM
	tools           []llm.DynamicTool
	toolsByName     map[string]llm.DynamicTool
	outputSchemaRaw json.RawMessage
	maxSteps        int
}

type DynamicAgentResult struct {
	OutputJSON json.RawMessage
	Messages   []llm.DynamicMessage
	Steps      int
}

func NewDynamicAgent(config DynamicAgentConfig) (*DynamicAgent, error) {
	if err := validation.NameIsValid(config.Name); err != nil {
		return nil, fmt.Errorf("name: %w", err)
	}
	if err := validation.StringIsNotEmpty(config.Behavior); err != nil {
		return nil, fmt.Errorf("behavior: %w", err)
	}
	if config.LLM == nil {
		return nil, fmt.Errorf("llm: %w: value cannot be nil", validation.ErrValidationFailed)
	}
	outputSchemaRaw, err := schema.NormalizeDynamicSchema(config.OutputSchemaRaw)
	if err != nil {
		return nil, fmt.Errorf("output schema: %w", err)
	}
	maxSteps := config.MaxSteps
	if maxSteps <= 0 {
		maxSteps = defaultDynamicMaxSteps
	}
	tools := append([]llm.DynamicTool(nil), config.Tools...)
	toolsByName := make(map[string]llm.DynamicTool, len(tools))
	for _, tool := range tools {
		toolsByName[tool.Name] = tool
	}

	return &DynamicAgent{
		name:            config.Name,
		behavior:        config.Behavior,
		provider:        config.LLM,
		tools:           tools,
		toolsByName:     toolsByName,
		outputSchemaRaw: outputSchemaRaw,
		maxSteps:        maxSteps,
	}, nil
}

func (a *DynamicAgent) Run(ctx context.Context, input json.RawMessage) (*DynamicAgentResult, error) {
	if len(input) == 0 {
		input = json.RawMessage(`{}`)
	}
	if !json.Valid(input) {
		return nil, fmt.Errorf("input must be valid JSON")
	}
	messages := []llm.DynamicMessage{
		{Role: llm.DynamicRoleSystem, Content: a.behavior},
		{Role: llm.DynamicRoleUser, Content: string(input)},
	}
	for step := 1; step <= a.maxSteps; step++ {
		message, err := a.provider.CallDynamic(ctx, append([]llm.DynamicMessage(nil), messages...), a.tools)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrLLMCall, err)
		}
		messages = append(messages, message)
		if len(message.ToolCalls) > 0 {
			toolResults := make([]llm.DynamicToolResult, 0, len(message.ToolCalls))
			for _, toolCall := range message.ToolCalls {
				tool, ok := a.toolsByName[toolCall.ToolName]
				if !ok {
					toolResults = append(toolResults, llm.DynamicToolResult{
						ToolCallID: toolCall.ID,
						Error:      fmt.Sprintf("%s: %s", ErrToolNotFound, toolCall.ToolName),
					})

					continue
				}
				result, err := tool.Call(ctx, toolCall)
				if err != nil {
					toolResults = append(toolResults, llm.DynamicToolResult{
						ToolCallID: toolCall.ID,
						Error:      fmt.Sprintf("%s: %v", ErrToolError, err),
					})

					continue
				}
				toolResults = append(toolResults, result)
			}
			messages = append(messages, llm.DynamicMessage{
				Role:        llm.DynamicRoleTool,
				ToolResults: toolResults,
			})
		}
		if message.End {
			output, err := a.provider.CallWithDynamicStructuredOutput(ctx, messages, a.outputSchemaRaw)
			if err != nil {
				return nil, fmt.Errorf("%w: %v", ErrLLMCall, err)
			}

			return &DynamicAgentResult{
				OutputJSON: append(json.RawMessage(nil), output...),
				Messages:   append([]llm.DynamicMessage(nil), messages...),
				Steps:      step,
			}, nil
		}
	}

	return &DynamicAgentResult{
		OutputJSON: nil,
		Messages:   append([]llm.DynamicMessage(nil), messages...),
		Steps:      a.maxSteps,
	}, ErrLimitReached
}
