package llm

import (
	"fmt"

	"github.com/vitalii-honchar/go-agent/internal/validation"
)

// LLMType represents the type of LLM provider
type LLMType string

const (
	// LLMTypeOpenAI represents the OpenAI LLM provider
	LLMTypeOpenAI LLMType = "openai"
)

// LLMConfig contains configuration for LLM providers
type LLMConfig struct {
	Type        LLMType `json:"type"`
	APIKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
}

func (c *LLMConfig) Validate() error {
	if err := validation.StringIsNotEmpty(string(c.Type)); err != nil {
		return fmt.Errorf("type: %w", err)
	}
	if err := validation.StringIsNotEmpty(c.APIKey); err != nil {
		return fmt.Errorf("api key: %w", err)
	}
	if err := validation.StringIsNotEmpty(c.Model); err != nil {
		return fmt.Errorf("model: %w", err)
	}

	return nil
}
