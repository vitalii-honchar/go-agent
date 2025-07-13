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
		return result, fmt.Errorf("%w: %w", ErrStructuredOutput, err)
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return result, fmt.Errorf("%w: %w", ErrStructuredOutput, err)
	}

	return result, nil
}
