// Package llm provides LLM abstractions and implementations for the Go Agent library
package llm

import (
	"context"
	"fmt"
)

// ErrUnsupportedLLMType is returned when an unsupported LLM type is specified
var ErrUnsupportedLLMType = fmt.Errorf("unsupported LLM type")

// LLM represents a language model interface
type LLM interface {
	Call(ctx context.Context, msgs []LLMMessage) (LLMMessage, error)
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
