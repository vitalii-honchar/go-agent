package agent

import (
	"fmt"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
)

// AgentResult represents the result of an agent execution
type AgentResult[T any] struct {
	Data     *T               `json:"data"`
	Messages []llm.LLMMessage `json:"messages"`
}

// NewAgentResult creates a new AgentResult with the given data and messages
func NewAgentResult[T any](data *T, messages []llm.LLMMessage) (*AgentResult[T], error) {
	if data == nil {
		return nil, fmt.Errorf("%w: data cannot be nil", ErrInvalidResultSchema)
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("%w: messages cannot be empty", ErrInvalidResultSchema)
	}

	return &AgentResult[T]{
		Data:     data,
		Messages: messages,
	}, nil
}
