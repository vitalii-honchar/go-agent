package llm

import (
	"context"
	"fmt"
)

var ErrUnsupportedLLMType = fmt.Errorf("unsupported LLM type")

type LLM interface {
	Call(ctx context.Context, msgs []LLMMessage) (LLMMessage, error)
}

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
