package llm

type LLMMessageType string

const (
	LLMMessageTypeUser      LLMMessageType = "user"
	LLMMessageTypeAssistant LLMMessageType = "assistant"
	LLMMessageTypeSystem    LLMMessageType = "system"
)

type LLMMessage struct {
	Type        LLMMessageType  `json:"type"`
	Content     string          `json:"content"`
	ToolCalls   []LLMToolCall   `json:"tool_call,omitempty"`
	ToolResults []LLMToolResult `json:"tool_result,omitempty"`
	End         bool            `json:"end,omitempty"`
}

func NewLLMMessage(msgType LLMMessageType, content string) LLMMessage {
	return LLMMessage{
		Type:    msgType,
		Content: content,
	}
}
