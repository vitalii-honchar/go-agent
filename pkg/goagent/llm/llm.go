// Package llm provides LLM abstractions and implementations for the Go Agent library
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
		return result, fmt.Errorf("%w: %s", ErrStructuredOutput, err)
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return result, fmt.Errorf("%w: %s", ErrStructuredOutput, err)
	}

	return result, nil
}

// CreateLLM creates a new LLM instance based on the configuration
func CreateLLM(cfg LLMConfig, tools map[string]LLMTool) (LLM, error) {
	switch cfg.Type {
	case LLMTypeOpenAI:
		return newOpenAILLM(
			withOpenAIAPIKey(cfg.APIKey),
			withOpenAILLMModel(cfg.Model),
			withOpenAILLMTemperature(cfg.Temperature),
			withOpenAITools(toSlice(tools)),
		), nil
	default:
		return nil, ErrUnsupportedLLMType
	}
}

func toSlice(tools map[string]LLMTool) []LLMTool {
	if len(tools) == 0 {
		return nil
	}
	slice := make([]LLMTool, 0, len(tools))
	for _, tool := range tools {
		slice = append(slice, tool)
	}
	return slice
}
