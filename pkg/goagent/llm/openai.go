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
		o.model = openai.ChatModel(model)
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
	params, err := o.createParameters(msgs)
	if err != nil {
		return LLMMessage{}, fmt.Errorf("failed to create OpenAI parameters: %w", err)
	}

	completion, err := o.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return LLMMessage{}, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(completion.Choices) == 0 {
		return LLMMessage{}, fmt.Errorf("no response from OpenAI")
	}

	return o.newLLMMessage(completion.Choices[0]), nil
}

func (o *openAILLM) CallWithStructuredOutput(ctx context.Context, msgs []LLMMessage, schemaT any) (string, error) {
	params, err := o.createParametersWithStructuredOutput(msgs, schemaT)
	if err != nil {
		return "", fmt.Errorf("failed to create OpenAI parameters: %w", err)
	}

	completion, err := o.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return completion.Choices[0].Message.Content, nil
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
	var res []LLMToolCall
	for _, toolCall := range choice.Message.ToolCalls {
		args := make(map[string]any)
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			continue
		}

		res = append(res, NewLLMToolCall(toolCall.ID, toolCall.Function.Name, args))
	}
	return res
}

func (o *openAILLM) createParameters(messages []LLMMessage) (openai.ChatCompletionNewParams, error) {
	openAIMessages, err := o.createMessages(messages)
	if err != nil {
		return openai.ChatCompletionNewParams{}, err
	}

	return openai.ChatCompletionNewParams{
		Messages:    openAIMessages,
		Model:       o.model,
		Temperature: openai.Float(o.temperature),
		Tools:       o.createToolParams(),
	}, nil
}

func (o *openAILLM) createParametersWithStructuredOutput(messages []LLMMessage, schemaT any) (openai.ChatCompletionNewParams, error) {
	openAIMessages, err := o.createMessages(messages)
	if err != nil {
		return openai.ChatCompletionNewParams{}, err
	}

	// Convert jsonschema.Schema to map[string]any directly
	schemaMap, err := schema.GenerateSchema(schemaT)
	if err != nil {
		return openai.ChatCompletionNewParams{}, fmt.Errorf("failed to convert schema to map: %w", err)
	}

	responseFormat := openai.ChatCompletionNewParamsResponseFormatUnion{}
	responseFormat.OfJSONSchema = &shared.ResponseFormatJSONSchemaParam{
		JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:   "response_schema",
			Schema: schemaMap,
			Strict: openai.Bool(true),
		},
	}

	return openai.ChatCompletionNewParams{
		Messages:       openAIMessages,
		Model:          o.model,
		Temperature:    openai.Float(o.temperature),
		ResponseFormat: responseFormat,
	}, nil
}

func (o *openAILLM) createToolParams() []openai.ChatCompletionToolParam {
	toolParams := make([]openai.ChatCompletionToolParam, 0, len(o.tools))

	for _, tool := range o.tools {
		toolParams = append(toolParams, openai.ChatCompletionToolParam{
			Function: openai.FunctionDefinitionParam{
				Name:        string(tool.Name),
				Description: openai.String(tool.Description),
				Parameters:  tool.ParametersSchema,
			},
		})
	}

	return toolParams
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
			if len(msg.ToolCalls) > 0 {
				// First add the assistant message with tool calls
				messages, err := o.addToolCalls(openAIMessages, msg)
				if err != nil {
					return nil, err
				}
				openAIMessages = messages

				// Then add tool results if any
				if len(msg.ToolResults) > 0 {
					messages, err = o.addToolResults(openAIMessages, msg)
					if err != nil {
						return nil, err
					}
					openAIMessages = messages
				}
			} else {
				openAIMessages = append(openAIMessages, openai.AssistantMessage(msg.Content))
			}
		}
	}

	return openAIMessages, nil
}

func (o *openAILLM) addToolCalls(openAIMessages []openai.ChatCompletionMessageParamUnion, msg LLMMessage) ([]openai.ChatCompletionMessageParamUnion, error) {
	var toolCalls []openai.ChatCompletionMessageToolCallParam

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

func (o *openAILLM) addToolResults(openAIMessages []openai.ChatCompletionMessageParamUnion, msg LLMMessage) ([]openai.ChatCompletionMessageParamUnion, error) {
	for _, toolRes := range msg.ToolResults {
		toolResJSON, err := json.Marshal(toolRes)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedToMarshalToolResult, err)
		}
		openAIMessages = append(openAIMessages, openai.ToolMessage(string(toolResJSON), toolRes.GetID()))
	}
	return openAIMessages, nil
}
