package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
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

type openAILLM struct {
	client      openai.Client
	apiKey      string
	temperature float64
	model       openai.ChatModel
	tools       []LLMTool
}

type openAILLMOption func(o *openAILLM)

func withOpenAILLMTemperature(temperature float64) openAILLMOption {
	return func(o *openAILLM) {
		o.temperature = temperature
	}
}
func withOpenAILLMModel(model string) openAILLMOption {
	return func(o *openAILLM) {
		o.model = model
	}
}

func withOpenAIAPIKey(apiKey string) openAILLMOption {
	return func(o *openAILLM) {
		o.apiKey = apiKey
		o.client = openai.NewClient(option.WithAPIKey(apiKey))
	}
}

func withOpenAITools(tools []LLMTool) openAILLMOption {
	return func(o *openAILLM) {
		o.tools = tools
	}
}

func newOpenAILLM(options ...openAILLMOption) *openAILLM {
	llm := &openAILLM{}
	for _, opt := range options {
		opt(llm)
	}

	return llm
}

func (o *openAILLM) Call(ctx context.Context, msgs []LLMMessage) (LLMMessage, error) {
	choice, err := o.callLLM(ctx, msgs, nil)
	if err != nil {
		return LLMMessage{}, err
	}

	return o.newLLMMessage(choice), nil
}

func (o *openAILLM) CallWithStructuredOutput(ctx context.Context, msgs []LLMMessage, schemaT any) (string, error) {
	choice, err := o.callLLM(ctx, msgs, schemaT)
	if err != nil {
		return "", err
	}

	return choice.Message.Content, nil
}

func (o *openAILLM) callLLM(ctx context.Context, msgs []LLMMessage, schemaT any) (openai.ChatCompletionChoice, error) {
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

func (o *openAILLM) newLLMMessage(choice openai.ChatCompletionChoice) LLMMessage {
	return LLMMessage{
		Type:      LLMMessageTypeAssistant,
		Content:   choice.Message.Content,
		ToolCalls: o.createLLMToolCalls(choice),
		End:       choice.FinishReason == openAIFinishReasonStop || choice.FinishReason == openAIFinishReasonLength,
	}
}

func (o *openAILLM) createLLMToolCalls(choice openai.ChatCompletionChoice) []LLMToolCall {
	res := make([]LLMToolCall, 0, len(choice.Message.ToolCalls))
	for _, toolCall := range choice.Message.ToolCalls {
		res = append(res, NewLLMToolCall(toolCall.ID, toolCall.Function.Name, toolCall.Function.Arguments))
	}

	return res
}

func (o *openAILLM) createParameters(messages []LLMMessage, schemaT any) (openai.ChatCompletionNewParams, error) {
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

func (o *openAILLM) createToolParams() ([]openai.ChatCompletionToolParam, error) {
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

func (o *openAILLM) createMessages(msgs []LLMMessage) ([]openai.ChatCompletionMessageParamUnion, error) {
	openAIMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(msgs))

	for _, msg := range msgs {
		switch msg.Type {
		case LLMMessageTypeSystem:
			openAIMessages = append(openAIMessages, openai.SystemMessage(msg.Content))
		case LLMMessageTypeUser:
			openAIMessages = append(openAIMessages, openai.UserMessage(msg.Content))
		case LLMMessageTypeAssistant:
			messages, err := o.handleAssistantMessage(openAIMessages, msg)
			if err != nil {
				return nil, err
			}

			openAIMessages = messages
		}
	}

	return openAIMessages, nil
}

func (o *openAILLM) handleAssistantMessage(openAIMessages []openai.ChatCompletionMessageParamUnion,
	msg LLMMessage) ([]openai.ChatCompletionMessageParamUnion, error) {
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

func (o *openAILLM) addToolCalls(openAIMessages []openai.ChatCompletionMessageParamUnion,
	msg LLMMessage) ([]openai.ChatCompletionMessageParamUnion, error) {
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

func (o *openAILLM) addToolResults(openAIMessages []openai.ChatCompletionMessageParamUnion,
	msg LLMMessage) ([]openai.ChatCompletionMessageParamUnion, error) {
	for _, toolRes := range msg.ToolResults {
		toolResJSON, err := json.Marshal(toolRes)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedToMarshalToolResult, err)
		}

		openAIMessages = append(openAIMessages, openai.ToolMessage(string(toolResJSON), toolRes.GetID()))
	}

	return openAIMessages, nil
}
