// Package llmfactory provides factory functions for creating LLM instances
package llmfactory

import (
	"github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/openai"
)

// CreateLLM creates a new LLM instance based on the configuration
func CreateLLM(cfg llm.LLMConfig, tools map[string]llm.LLMTool) (llm.LLM, error) {
	switch cfg.Type {
	case llm.LLMTypeOpenAI:
		return openai.NewOpenAILLM(
			openai.WithAPIKey(cfg.APIKey),
			openai.WithModel(cfg.Model),
			openai.WithTemperature(cfg.Temperature),
			openai.WithTools(toSlice(tools)),
		), nil
	default:
		return nil, llm.ErrUnsupportedLLMType
	}
}

func toSlice(tools map[string]llm.LLMTool) []llm.LLMTool {
	if len(tools) == 0 {
		return nil
	}

	slice := make([]llm.LLMTool, 0, len(tools))
	for _, tool := range tools {
		slice = append(slice, tool)
	}

	return slice
}
