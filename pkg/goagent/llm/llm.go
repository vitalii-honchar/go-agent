// Package llm provides LLM abstractions and implementations for the Go Agent library.
//
// This package defines interfaces and types for working with Large Language Models (LLMs)
// and their tools. It supports multiple LLM providers through a unified interface.
//
// Core Types:
//
//	type LLM interface {
//		Call(ctx context.Context, messages []LLMMessage, tools []LLMTool) (*LLMResponse, error)
//	}
//
// Supported LLM providers:
//   - OpenAI (GPT-3.5, GPT-4, GPT-4 Turbo)
//   - Extensible to other providers
//
// Tool System:
//
// Tools allow LLMs to perform external actions and retrieve information:
//
//	type MyToolParams struct {
//		Input string `json:"input" jsonschema_description:"Input to process"`
//	}
//
//	type MyToolResult struct {
//		BaseLLMToolResult
//		Output string `json:"output" jsonschema_description:"Processed output"`
//	}
//
//	tool := NewLLMTool(
//		WithLLMToolName("my-tool"),
//		WithLLMToolDescription("Processes input and returns output"),
//		WithLLMToolParametersSchema[MyToolParams](),
//		WithLLMToolCall(func(callID string, params MyToolParams) (MyToolResult, error) {
//			// Process params.Input
//			return MyToolResult{
//				BaseLLMToolResult: BaseLLMToolResult{ID: callID},
//				Output:            processedOutput,
//			}, nil
//		}),
//	)
//
// Configuration:
//
//	config := LLMConfig{
//		Type:        LLMTypeOpenAI,
//		APIKey:      "your-api-key",
//		Model:       "gpt-4",
//		Temperature: 0.1,
//	}
//
// The package handles JSON schema generation for tool parameters automatically,
// ensuring type-safe communication between the LLM and your tool implementations.
package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// ErrUnsupportedLLMType is returned when an unsupported LLM type is specified
var ErrUnsupportedLLMType = errors.New("unsupported LLM type")
var ErrStructuredOutput = errors.New("failed to call LLM with structured output")

// LLM represents a language model interface
type LLM interface {
	Call(ctx context.Context, msgs []LLMMessage) (LLMMessage, error)
	CallWithStructuredOutput(ctx context.Context, msgs []LLMMessage, schemaT any) (string, error)
}

// Call the LLM with structured output
func CallWithStructuredOutput[T any](ctx context.Context, llm LLM, msgs []LLMMessage) (T, error) {
	var result T

	output, err := llm.CallWithStructuredOutput(ctx, msgs, new(T))
	if err != nil {
		return result, fmt.Errorf("%w: %w", ErrStructuredOutput, err)
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return result, fmt.Errorf("%w: %w", ErrStructuredOutput, err)
	}

	return result, nil
}
