package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/schema"
)

const (
	openAIFinishReasonStop   = "stop"
	openAIFinishReasonLength = "length"
)

var (
	// ErrFailedToMarshalToolResult is returned when tool result marshaling fails
	ErrFailedToMarshalToolResult = errors.New("failed to marshal tool result")
	// ErrFailedToMarshalToolCall is returned when tool call marshaling fails
	ErrFailedToMarshalToolCall = errors.New("failed to marshal tool call")
	// ErrNoResponseFromOpenAI is returned when OpenAI returns no response
	ErrNoResponseFromOpenAI = errors.New("no response from OpenAI")
)

type OpenAILLM struct {
	client      openai.Client
	apiKey      string
	temperature float64
	model       openai.ChatModel
	tools       []llm.LLMTool
}

type OpenAILLMOption func(o *OpenAILLM)

func WithTemperature(temperature float64) OpenAILLMOption {
	return func(o *OpenAILLM) {
		o.temperature = temperature
	}
}

func WithModel(model string) OpenAILLMOption {
	return func(o *OpenAILLM) {
		o.model = model
	}
}

func WithAPIKey(apiKey string) OpenAILLMOption {
	return func(o *OpenAILLM) {
		o.apiKey = apiKey
		o.client = openai.NewClient(option.WithAPIKey(apiKey))
	}
}

func WithTools(tools []llm.LLMTool) OpenAILLMOption {
	return func(o *OpenAILLM) {
		o.tools = tools
	}
}

func NewOpenAILLM(options ...OpenAILLMOption) *OpenAILLM {
	llm := &OpenAILLM{}
	for _, opt := range options {
		opt(llm)
	}

	return llm
}

func (o *OpenAILLM) Call(ctx context.Context, msgs []llm.LLMMessage) (llm.LLMMessage, error) {
	choice, err := o.callLLM(ctx, msgs, nil)
	if err != nil {
		return llm.LLMMessage{}, err
	}

	return o.newLLMMessage(choice)
}

func (o *OpenAILLM) CallWithStructuredOutput(ctx context.Context, msgs []llm.LLMMessage, schemaT any) (string, error) {
	choice, err := o.callLLM(ctx, msgs, schemaT)
	if err != nil {
		return "", err
	}

	return choice.Message.Content, nil
}

func (o *OpenAILLM) callLLM(
	ctx context.Context, msgs []llm.LLMMessage, schemaT any,
) (openai.ChatCompletionChoice, error) {
	params, err := o.createParameters(msgs, schemaT)
	if err != nil {
		return openai.ChatCompletionChoice{}, fmt.Errorf("failed to create OpenAI parameters: %w", err)
	}

	completion, err := o.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return openai.ChatCompletionChoice{}, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(completion.Choices) == 0 {
		return openai.ChatCompletionChoice{}, ErrNoResponseFromOpenAI
	}

	return completion.Choices[0], nil
}

func (o *OpenAILLM) newLLMMessage(choice openai.ChatCompletionChoice) (llm.LLMMessage, error) {
	toolCalls, err := o.createLLMToolCalls(choice)
	if err != nil {
		return llm.LLMMessage{}, fmt.Errorf("failed to create tool calls: %w", err)
	}

	return llm.LLMMessage{
		Type:      llm.LLMMessageTypeAssistant,
		Content:   choice.Message.Content,
		ToolCalls: toolCalls,
		End:       choice.FinishReason == openAIFinishReasonStop || choice.FinishReason == openAIFinishReasonLength,
	}, nil
}

func (o *OpenAILLM) createLLMToolCalls(choice openai.ChatCompletionChoice) ([]llm.LLMToolCall, error) {
	res := make([]llm.LLMToolCall, 0, len(choice.Message.ToolCalls))
	for _, toolCall := range choice.Message.ToolCalls {
		toolCallObj, err := llm.NewLLMToolCall(toolCall.ID, toolCall.Function.Name, toolCall.Function.Arguments)
		if err != nil {
			return nil, fmt.Errorf("failed to create tool call: %w", err)
		}
		res = append(res, toolCallObj)
	}

	return res, nil
}

func (o *OpenAILLM) createParameters(messages []llm.LLMMessage, schemaT any) (openai.ChatCompletionNewParams, error) {
	openAIMessages, err := o.createMessages(messages)
	if err != nil {
		return openai.ChatCompletionNewParams{}, err
	}

	tools, err := o.createToolParams()
	if err != nil {
		return openai.ChatCompletionNewParams{}, fmt.Errorf("failed to create tool parameters: %w", err)
	}

	params := openai.ChatCompletionNewParams{
		Messages:    openAIMessages,
		Model:       o.model,
		Temperature: openai.Float(o.temperature),
		Tools:       tools,
	}

	if schemaT != nil {
		schemaMap, err := schema.GenerateSchema(schemaT)
		if err != nil {
			return openai.ChatCompletionNewParams{}, fmt.Errorf("failed to convert schema to map: %w", err)
		}

		params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
				JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        "response_schema",
					Description: openai.String("Response schema for structured output of a conversation"),
					Schema:      schemaMap,
					Strict:      openai.Bool(true),
				},
			},
		}
	}

	return params, nil
}

func (o *OpenAILLM) createToolParams() ([]openai.ChatCompletionToolParam, error) {
	toolParams := make([]openai.ChatCompletionToolParam, 0, len(o.tools))

	for _, tool := range o.tools {
		parameterSchema, err := schema.GenerateSchema(tool.ParametersSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to generate schema for tool %s: %w", tool.Name, err)
		}

		toolParams = append(toolParams, openai.ChatCompletionToolParam{
			Function: openai.FunctionDefinitionParam{
				Name:        tool.Name,
				Description: openai.String(tool.Description),
				Parameters:  parameterSchema,
			},
		})
	}

	return toolParams, nil
}

func (o *OpenAILLM) createMessages(msgs []llm.LLMMessage) ([]openai.ChatCompletionMessageParamUnion, error) {
	openAIMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(msgs))

	for _, msg := range msgs {
		switch msg.Type {
		case llm.LLMMessageTypeSystem:
			openAIMessages = append(openAIMessages, openai.SystemMessage(msg.Content))
		case llm.LLMMessageTypeUser:
			openAIMessages = append(openAIMessages, openai.UserMessage(msg.Content))
		case llm.LLMMessageTypeAssistant:
			messages, err := o.handleAssistantMessage(openAIMessages, msg)
			if err != nil {
				return nil, err
			}

			openAIMessages = messages
		}
	}

	return openAIMessages, nil
}

func (o *OpenAILLM) handleAssistantMessage(openAIMessages []openai.ChatCompletionMessageParamUnion,
	msg llm.LLMMessage) ([]openai.ChatCompletionMessageParamUnion, error) {
	if len(msg.ToolCalls) == 0 {
		return append(openAIMessages, openai.AssistantMessage(msg.Content)), nil
	}

	// First add the assistant message with tool calls
	messages, err := o.addToolCalls(openAIMessages, msg)
	if err != nil {
		return nil, err
	}

	// Then add tool results if any
	if len(msg.ToolResults) > 0 {
		messages, err = o.addToolResults(messages, msg)
		if err != nil {
			return nil, err
		}
	}

	return messages, nil
}

func (o *OpenAILLM) addToolCalls(openAIMessages []openai.ChatCompletionMessageParamUnion,
	msg llm.LLMMessage) ([]openai.ChatCompletionMessageParamUnion, error) {
	toolCalls := make([]openai.ChatCompletionMessageToolCallParam, 0, len(msg.ToolCalls))

	for _, toolCall := range msg.ToolCalls {
		argsJSON, err := json.Marshal(toolCall.Args)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedToMarshalToolCall, err)
		}

		toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallParam{
			ID: toolCall.ID,
			Function: openai.ChatCompletionMessageToolCallFunctionParam{
				Name:      toolCall.ToolName,
				Arguments: string(argsJSON),
			},
		})
	}

	assistantMsg := openai.AssistantMessage(msg.Content)
	assistantMsg.OfAssistant.ToolCalls = toolCalls

	openAIMessages = append(openAIMessages, assistantMsg)

	return openAIMessages, nil
}

func (o *OpenAILLM) addToolResults(openAIMessages []openai.ChatCompletionMessageParamUnion,
	msg llm.LLMMessage) ([]openai.ChatCompletionMessageParamUnion, error) {
	for _, toolRes := range msg.ToolResults {
		toolResJSON, err := json.Marshal(toolRes)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedToMarshalToolResult, err)
		}

		openAIMessages = append(openAIMessages, openai.ToolMessage(string(toolResJSON), toolRes.GetID()))
	}

	return openAIMessages, nil
}
