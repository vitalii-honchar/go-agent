package agent_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/vitalii-honchar/go-agent/pkg/goagent/agent"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
)

func TestDynamicAgentRunsReActLoopWithDynamicSchemas(t *testing.T) {
	t.Parallel()

	provider := &fakeDynamicLLM{
		messages: []llm.DynamicMessage{
			{
				Role: llm.DynamicRoleAssistant,
				ToolCalls: []llm.DynamicToolCall{{
					ID:       "call-1",
					ToolName: "lookup",
					ArgsJSON: json.RawMessage(`{"topic":"agentd"}`),
				}},
			},
			{
				Role:    llm.DynamicRoleAssistant,
				Content: "ready",
				End:     true,
			},
		},
		structuredOutput: json.RawMessage(`{"summary":"agentd uses contracts"}`),
	}
	toolCalls := 0
	runner, err := agent.NewDynamicAgent(agent.DynamicAgentConfig{
		Name:            "dynamic_agent",
		Behavior:        "Use tools when needed, then finish.",
		LLM:             provider,
		OutputSchemaRaw: json.RawMessage(`{"type":"object","required":["summary"],"properties":{"summary":{"type":"string"}}}`),
		Tools: []llm.DynamicTool{{
			Name:                "lookup",
			Description:         "Lookup topic evidence",
			ParametersSchemaRaw: json.RawMessage(`{"type":"object","required":["topic"],"properties":{"topic":{"type":"string"}}}`),
			Callback: func(_ context.Context, call llm.DynamicToolCall) (llm.DynamicToolResult, error) {
				toolCalls++
				if string(call.ArgsJSON) != `{"topic":"agentd"}` {
					t.Fatalf("tool args: %s", call.ArgsJSON)
				}

				return llm.DynamicToolResult{
					ToolCallID:  call.ID,
					ContentJSON: json.RawMessage(`{"evidence":"contracts"}`),
				}, nil
			},
		}},
		MaxSteps: 4,
	})
	if err != nil {
		t.Fatalf("NewDynamicAgent: %v", err)
	}

	result, err := runner.Run(context.Background(), json.RawMessage(`{"topic":"agentd"}`))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if string(result.OutputJSON) != `{"summary":"agentd uses contracts"}` {
		t.Fatalf("output: %s", result.OutputJSON)
	}
	if toolCalls != 1 {
		t.Fatalf("tool calls: got %d want 1", toolCalls)
	}
	if provider.callCount != 2 {
		t.Fatalf("provider calls: got %d want 2", provider.callCount)
	}
	if len(result.Messages) < 4 {
		t.Fatalf("messages should include system, user, assistant, and tool observation: %#v", result.Messages)
	}
}

func TestDynamicAgentRunsOneProviderStepWhenProviderFinishes(t *testing.T) {
	t.Parallel()

	provider := &fakeDynamicLLM{
		messages: []llm.DynamicMessage{{
			Role:    llm.DynamicRoleAssistant,
			Content: "done",
			End:     true,
		}},
		structuredOutput: json.RawMessage(`{"summary":"done"}`),
	}
	runner, err := agent.NewDynamicAgent(agent.DynamicAgentConfig{
		Name:            "one_step_agent",
		Behavior:        "Answer directly when no tool is needed.",
		LLM:             provider,
		OutputSchemaRaw: json.RawMessage(`{"type":"object","required":["summary"],"properties":{"summary":{"type":"string"}}}`),
		MaxSteps:        3,
	})
	if err != nil {
		t.Fatalf("NewDynamicAgent: %v", err)
	}

	result, err := runner.Run(context.Background(), json.RawMessage(`{"topic":"agentd"}`))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if string(result.OutputJSON) != `{"summary":"done"}` {
		t.Fatalf("output: %s", result.OutputJSON)
	}
	if provider.callCount != 1 {
		t.Fatalf("provider calls: got %d want 1", provider.callCount)
	}
}

type fakeDynamicLLM struct {
	messages         []llm.DynamicMessage
	structuredOutput json.RawMessage
	callCount        int
}

func (f *fakeDynamicLLM) CallDynamic(
	_ context.Context,
	_ []llm.DynamicMessage,
	_ []llm.DynamicTool,
) (llm.DynamicMessage, error) {
	if len(f.messages) == 0 {
		return llm.DynamicMessage{}, nil
	}
	f.callCount++
	message := f.messages[0]
	f.messages = f.messages[1:]

	return message, nil
}

func (f *fakeDynamicLLM) CallWithDynamicStructuredOutput(
	_ context.Context,
	_ []llm.DynamicMessage,
	_ json.RawMessage,
) (json.RawMessage, error) {
	return append(json.RawMessage(nil), f.structuredOutput...), nil
}
