package llm_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
)

func TestDynamicToolCallsExternalCallbackWithRawJSON(t *testing.T) {
	t.Parallel()

	called := false
	tool, err := llm.NewDynamicTool(llm.DynamicToolConfig{
		Name:                "lookup",
		Description:         "Lookup topic evidence",
		ParametersSchemaRaw: json.RawMessage(`{"type":"object","required":["topic"],"properties":{"topic":{"type":"string"}}}`),
		Callback: func(_ context.Context, call llm.DynamicToolCall) (llm.DynamicToolResult, error) {
			called = true
			if call.ID != "call-1" || call.ToolName != "lookup" {
				t.Fatalf("call metadata: %#v", call)
			}
			if string(call.ArgsJSON) != `{"topic":"agentd"}` {
				t.Fatalf("args: %s", call.ArgsJSON)
			}

			return llm.DynamicToolResult{
				ToolCallID:  call.ID,
				ContentJSON: json.RawMessage(`{"evidence":"contracts"}`),
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("NewDynamicTool: %v", err)
	}

	result, err := tool.Call(context.Background(), llm.DynamicToolCall{
		ID:       "call-1",
		ToolName: "lookup",
		ArgsJSON: json.RawMessage(`{"topic":"agentd"}`),
	})
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if !called {
		t.Fatal("callback was not called")
	}
	if result.ToolCallID != "call-1" || string(result.ContentJSON) != `{"evidence":"contracts"}` {
		t.Fatalf("result: %#v", result)
	}
}

func TestDynamicToolReturnsCallbackError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("denied")
	tool, err := llm.NewDynamicTool(llm.DynamicToolConfig{
		Name:                "lookup",
		Description:         "Lookup topic evidence",
		ParametersSchemaRaw: json.RawMessage(`{"type":"object"}`),
		Callback: func(context.Context, llm.DynamicToolCall) (llm.DynamicToolResult, error) {
			return llm.DynamicToolResult{}, wantErr
		},
	})
	if err != nil {
		t.Fatalf("NewDynamicTool: %v", err)
	}

	_, err = tool.Call(context.Background(), llm.DynamicToolCall{
		ID:       "call-1",
		ToolName: "lookup",
		ArgsJSON: json.RawMessage(`{}`),
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Call error: got %v want %v", err, wantErr)
	}
}

func TestDynamicToolRejectsInvalidConfiguration(t *testing.T) {
	t.Parallel()

	_, err := llm.NewDynamicTool(llm.DynamicToolConfig{
		Name:                "",
		Description:         "Lookup topic evidence",
		ParametersSchemaRaw: json.RawMessage(`{"type":"object"}`),
		Callback: func(context.Context, llm.DynamicToolCall) (llm.DynamicToolResult, error) {
			return llm.DynamicToolResult{}, nil
		},
	})
	if err == nil {
		t.Fatal("NewDynamicTool error is nil")
	}
}
