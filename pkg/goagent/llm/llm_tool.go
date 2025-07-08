package llm

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/invopop/jsonschema"
)

var (
	// ErrInvalidArguments is returned when tool arguments are invalid
	ErrInvalidArguments = errors.New("invalid arguments")
)

// LLMTool represents a tool that can be called by an LLM
type LLMTool struct {
	Name             string                                                      `json:"name"`
	ParametersSchema map[string]any                                              `json:"parameters_schema"`
	Description      string                                                      `json:"description"`
	Call             func(id string, args map[string]any) (LLMToolResult, error) `json:"-"`
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

// WithLLMToolParametersSchema sets the parameters schema from a Go type
func WithLLMToolParametersSchema[T any](paramType *T) LLMToolOption {
	return func(tool *LLMTool) {
		// Generate schema from Go type
		reflectedSchema := jsonschema.Reflect(paramType)
		schemaBytes, err := json.Marshal(reflectedSchema)
		if err != nil {
			return
		}
		
		// Convert to map for processing
		var schemaMap map[string]any
		if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
			return
		}
		
		// Extract the actual object definition from $defs
		// The jsonschema.Reflect creates a schema with $ref and $defs
		// We need to extract the actual object schema for OpenAI
		if defs, ok := schemaMap["$defs"].(map[string]any); ok {
			for _, def := range defs {
				if defMap, ok := def.(map[string]any); ok {
					if defMap["type"] == "object" {
						// Use the actual object definition
						tool.ParametersSchema = defMap
						return
					}
				}
			}
		}
		
		// Fallback: use the schema as-is if no $defs found
		tool.ParametersSchema = schemaMap
	}
}

// WithLLMToolCall sets the call function for the tool
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

// WithLLMToolCallTyped sets the call function for the tool with typed parameters
func WithLLMToolCallTyped[P any, T LLMToolResult](callFunc func(id string, params P) (T, error)) LLMToolOption {
	return func(tool *LLMTool) {
		tool.Call = func(id string, args map[string]any) (LLMToolResult, error) {
			// Marshal args to JSON and unmarshal to typed params
			argsBytes, err := json.Marshal(args)
			if err != nil {
				return nil, fmt.Errorf("%w: failed to marshal arguments: %v", ErrInvalidArguments, err)
			}
			
			var params P
			if err := json.Unmarshal(argsBytes, &params); err != nil {
				return nil, fmt.Errorf("%w: failed to unmarshal arguments: %v", ErrInvalidArguments, err)
			}
			
			result, err := callFunc(id, params)
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
	ID       string         `json:"id"`
	ToolName string         `json:"tool_name"`
	Args     map[string]any `json:"args"`
}

// NewLLMToolCall creates a new LLM tool call
func NewLLMToolCall(id string, toolName string, args map[string]any) LLMToolCall {
	return LLMToolCall{
		ID:       id,
		ToolName: toolName,
		Args:     args,
	}
}

// NewLLMToolTyped creates a new LLM tool with typed parameters (recommended approach)
func NewLLMToolTyped[P any, T LLMToolResult](name, description string, paramType *P, callFunc func(id string, params P) (T, error)) LLMTool {
	return NewLLMTool(
		WithLLMToolName(name),
		WithLLMToolDescription(description),
		WithLLMToolParametersSchema(paramType),
		WithLLMToolCallTyped(callFunc),
	)
}
