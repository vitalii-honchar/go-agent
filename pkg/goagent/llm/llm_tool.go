package llm

import (
	"encoding/json"
	"errors"
	"fmt"
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
func NewLLMTool(options ...LLMToolOption) LLMTool {
	tool := &LLMTool{}
	for _, opt := range options {
		opt(tool)
	}

	return *tool
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

// LLMToolCall represents a call to an LLM tool
type LLMToolCall struct {
	ID       string `json:"id"`
	ToolName string `json:"tool_name"`
	Args     string `json:"args"`
}

// NewLLMToolCall creates a new LLM tool call
func NewLLMToolCall(id string, toolName string, args string) LLMToolCall {
	return LLMToolCall{
		ID:       id,
		ToolName: toolName,
		Args:     args,
	}
}
