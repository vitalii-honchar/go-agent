package agent

import (
	"fmt"
	"reddit-analyzer/internal/agent/llm"
)

type AgentResult[T any] struct {
	Data     *T               `json:"data"`
	Messages []llm.LLMMessage `json:"messages"`
}

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
