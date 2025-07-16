package llm

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/vitalii-honchar/go-agent/internal/validation"
)

var (
	// ErrInvalidArguments is returned when tool arguments are invalid
	ErrInvalidArguments = errors.New("invalid arguments")
)

// LLMTool represents a tool that can be called by an LLM
type LLMTool struct {
	Name             string                                              `json:"name"`
	ParametersSchema any                                                 `json:"parameters_schema"`
	Description      string                                              `json:"description"`
	Call             func(id string, args string) (LLMToolResult, error) `json:"-"`
}

// LLMToolOption is a function that configures an LLMTool
type LLMToolOption func(tool *LLMTool)

// NewLLMTool creates a new LLM tool with the given options
func NewLLMTool(options ...LLMToolOption) (LLMTool, error) {
	tool := &LLMTool{}
	for _, opt := range options {
		opt(tool)
	}

	err := tool.validate()
	if err != nil {
		return LLMTool{}, fmt.Errorf("failed to create LLM tool: %w", err)
	}

	return *tool, nil
}

func (t *LLMTool) validate() error {
	if err := validation.NameIsValid(t.Name); err != nil {
		return fmt.Errorf("tool: %w", err)
	}
	if err := validation.DescriptionIsValid(t.Description); err != nil {
		return fmt.Errorf("description: %w", err)
	}
	if err := validation.NotNil(t.ParametersSchema); err != nil {
		return fmt.Errorf("parameters schema: %w", err)
	}
	if t.Call == nil {
		return fmt.Errorf("call: %w: value cannot be nil", validation.ErrValidationFailed)
	}

	return nil
}

// WithLLMToolName sets the name of the tool
func WithLLMToolName(name string) LLMToolOption {
	return func(tool *LLMTool) {
		tool.Name = name
	}
}

// WithLLMToolDescription sets the description of the tool
func WithLLMToolDescription(description string) LLMToolOption {
	return func(tool *LLMTool) {
		tool.Description = description
	}
}

func WithLLMToolParametersSchema[T any]() LLMToolOption {
	return func(tool *LLMTool) {
		tool.ParametersSchema = new(T)
	}
}

// WithLLMToolCall sets the call function for the tool
func WithLLMToolCall[P any, T LLMToolResult](callFunc func(callID string, args P) (T, error)) LLMToolOption {
	return func(tool *LLMTool) {
		tool.Call = func(callID string, args string) (LLMToolResult, error) {
			var typedArgs P
			if err := json.Unmarshal([]byte(args), &typedArgs); err != nil {
				return nil, fmt.Errorf("%w: failed to unmarshal arguments: %v", ErrInvalidArguments, err)
			}

			result, err := callFunc(callID, typedArgs)
			if err != nil {
				return nil, err
			}

			return result, nil
		}
	}
}

// LLMToolResult represents the result of a tool call
type LLMToolResult interface {
	GetID() string
}

// BaseLLMToolResult provides a base implementation for tool results
type BaseLLMToolResult struct {
	ID string `json:"id"`
}

// GetID returns the ID of the tool result
func (r BaseLLMToolResult) GetID() string {
	return r.ID
}

type ErrorLLMToolResult struct {
	BaseLLMToolResult
	Error string `json:"error"`
}

// LLMToolCall represents a call to an LLM tool
type LLMToolCall struct {
	ID       string `json:"id"`
	ToolName string `json:"tool_name"`
	Args     string `json:"args"`
}

// NewLLMToolCall creates a new LLM tool call with validation
func NewLLMToolCall(id string, toolName string, args string) (LLMToolCall, error) {
	toolCall := LLMToolCall{
		ID:       id,
		ToolName: toolName,
		Args:     args,
	}

	err := toolCall.validate()
	if err != nil {
		return LLMToolCall{}, fmt.Errorf("failed to create LLM tool call: %w", err)
	}

	return toolCall, nil
}

func (tc *LLMToolCall) validate() error {
	if err := validation.StringIsNotEmpty(tc.ID); err != nil {
		return fmt.Errorf("id: %w", err)
	}
	if err := validation.NameIsValid(tc.ToolName); err != nil {
		return fmt.Errorf("tool name: %w", err)
	}
	if err := validation.StringIsNotEmpty(tc.Args); err != nil {
		return fmt.Errorf("args: %w", err)
	}

	return nil
}
