package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/errors"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const (
	openAIFinishReasonStop   = "stop"
	openAIFinishReasonLength = "length"
)

var (
	ErrFailedToMarshalToolResult = errors.New("failed to marshal tool result")
	ErrFailedToMarshalToolCall   = errors.New("failed to marshal tool call")
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

				messages, err := o.addToolResults(openAIMessages, msg)
				if err != nil {
					return nil, err
				}
				openAIMessages = messages
			} else {
				openAIMessages = append(openAIMessages, openai.AssistantMessage(msg.Content))
			}
		}
	}

	return openAIMessages, nil
}

func (o *openAILLM) addToolCalls(openAIMessages []openai.ChatCompletionMessageParamUnion, msg LLMMessage) ([]openai.ChatCompletionMessageParamUnion, error) {
	for _, toolCall := range msg.ToolCalls {
		toolCallJson, err := json.Marshal(toolCall)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedToMarshalToolCall, err)
		}
		openAIMessages = append(openAIMessages, openai.ToolMessage(string(toolCallJson), toolCall.GetID()))
	}
	return openAIMessages, nil
}

func (o *openAILLM) addToolResults(openAIMessages []openai.ChatCompletionMessageParamUnion, msg LLMMessage) ([]openai.ChatCompletionMessageParamUnion, error) {
	for _, toolRes := range msg.ToolResults {
		toolResJson, err := json.Marshal(toolRes)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedToMarshalToolResult, err)
		}
		openAIMessages = append(openAIMessages, openai.ToolMessage(string(toolResJson), toolRes.GetID()))
	}
	return openAIMessages, nil
}
