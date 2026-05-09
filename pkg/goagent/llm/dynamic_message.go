package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vitalii-honchar/go-agent/internal/validation"
)

type DynamicRole string

const (
	DynamicRoleSystem    DynamicRole = "system"
	DynamicRoleUser      DynamicRole = "user"
	DynamicRoleAssistant DynamicRole = "assistant"
	DynamicRoleTool      DynamicRole = "tool"
)

type DynamicLLM interface {
	CallDynamic(ctx context.Context, messages []DynamicMessage, tools []DynamicTool) (DynamicMessage, error)
	CallWithDynamicStructuredOutput(
		ctx context.Context,
		messages []DynamicMessage,
		outputSchemaRaw json.RawMessage,
	) (json.RawMessage, error)
}

type DynamicMessage struct {
	Role        DynamicRole         `json:"role"`
	Content     string              `json:"content,omitempty"`
	ToolCalls   []DynamicToolCall   `json:"tool_calls,omitempty"`
	ToolResults []DynamicToolResult `json:"tool_results,omitempty"`
	End         bool                `json:"end,omitempty"`
}

type DynamicToolCall struct {
	ID       string          `json:"id"`
	ToolName string          `json:"tool_name"`
	ArgsJSON json.RawMessage `json:"args_json"`
}

type DynamicToolResult struct {
	ToolCallID  string          `json:"tool_call_id"`
	ContentJSON json.RawMessage `json:"content_json,omitempty"`
	Error       string          `json:"error,omitempty"`
}

type DynamicToolCallback func(context.Context, DynamicToolCall) (DynamicToolResult, error)

type DynamicToolConfig struct {
	Name                string
	Description         string
	ParametersSchemaRaw json.RawMessage
	Callback            DynamicToolCallback
}

type DynamicTool struct {
	Name                string              `json:"name"`
	Description         string              `json:"description"`
	ParametersSchemaRaw json.RawMessage     `json:"parameters_schema"`
	Callback            DynamicToolCallback `json:"-"`
}

func NewDynamicTool(config DynamicToolConfig) (DynamicTool, error) {
	tool := DynamicTool{
		Name:                config.Name,
		Description:         config.Description,
		ParametersSchemaRaw: append(json.RawMessage(nil), config.ParametersSchemaRaw...),
		Callback:            config.Callback,
	}
	if err := tool.validate(); err != nil {
		return DynamicTool{}, fmt.Errorf("failed to create dynamic tool: %w", err)
	}

	return tool, nil
}

func (t DynamicTool) Call(ctx context.Context, call DynamicToolCall) (DynamicToolResult, error) {
	if t.Callback == nil {
		return DynamicToolResult{}, fmt.Errorf("callback: %w: value cannot be nil", validation.ErrValidationFailed)
	}

	return t.Callback(ctx, call)
}

func (t DynamicTool) validate() error {
	if err := validation.NameIsValid(t.Name); err != nil {
		return fmt.Errorf("tool: %w", err)
	}
	if err := validation.DescriptionIsValid(t.Description); err != nil {
		return fmt.Errorf("description: %w", err)
	}
	if len(t.ParametersSchemaRaw) == 0 || !json.Valid(t.ParametersSchemaRaw) {
		return fmt.Errorf("parameters schema: %w: value must be valid JSON", validation.ErrValidationFailed)
	}
	if t.Callback == nil {
		return fmt.Errorf("callback: %w: value cannot be nil", validation.ErrValidationFailed)
	}

	return nil
}
