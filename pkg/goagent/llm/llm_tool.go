package llm

import "errors"

var ErrInvalidArguments = errors.New("invalid arguments")

type LLMTool struct {
	Name             string                                                      `json:"name"`
	ParametersSchema map[string]any                                              `json:"parameters_schema"`
	Description      string                                                      `json:"description"`
	Call             func(id string, args map[string]any) (LLMToolResult, error) `json:"-"`
}

type LLMToolOption func(tool *LLMTool)

func NewLLMTool(options ...LLMToolOption) LLMTool {
	tool := &LLMTool{}
	for _, opt := range options {
		opt(tool)
	}
	return *tool
}

func WithLLMToolName(name string) LLMToolOption {
	return func(tool *LLMTool) {
		tool.Name = name
	}
}

func WithLLMToolDescription(description string) LLMToolOption {
	return func(tool *LLMTool) {
		tool.Description = description
	}
}

func WithLLMToolParametersSchema(schema map[string]any) LLMToolOption {
	return func(tool *LLMTool) {
		tool.ParametersSchema = schema
	}
}

func WithLLMToolCall[T LLMToolResult](callFunc func(id string, args map[string]any) (T, error)) LLMToolOption {
	return func(tool *LLMTool) {
		tool.Call = func(id string, args map[string]any) (LLMToolResult, error) {
			result, err := callFunc(id, args)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
	}
}

type LLMToolResult interface {
	GetID() string
}

type BaseLLMToolResult struct {
	ID string `json:"id"`
}

func (r BaseLLMToolResult) GetID() string {
	return r.ID
}

type LLMToolCall struct {
	ID       string         `json:"id"`
	ToolName string         `json:"tool_name"`
	Args     map[string]any `json:"args"`
}

func NewLLMToolCall(id string, toolName string, args map[string]any) LLMToolCall {
	return LLMToolCall{
		ID:       id,
		ToolName: toolName,
		Args:     args,
	}
}
