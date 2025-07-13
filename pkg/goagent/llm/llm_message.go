package llm

// LLMMessageType represents the type of LLM message
type LLMMessageType string

const (
	// LLMMessageTypeUser represents a user message
	LLMMessageTypeUser LLMMessageType = "user"
	// LLMMessageTypeAssistant represents an assistant message
	LLMMessageTypeAssistant LLMMessageType = "assistant"
	// LLMMessageTypeSystem represents a system message
	LLMMessageTypeSystem LLMMessageType = "system"
)

// LLMMessage represents a message in an LLM conversation
type LLMMessage struct {
	Type        LLMMessageType  `json:"type"`
	Content     string          `json:"content"`
	ToolCalls   []LLMToolCall   `json:"tool_call,omitempty"`
	ToolResults []LLMToolResult `json:"tool_result,omitempty"`
	End         bool            `json:"end,omitempty"`
}

// NewLLMMessage creates a new LLM message with the given type and content
func NewLLMMessage(msgType LLMMessageType, content string) LLMMessage {
	return LLMMessage{
		Type:    msgType,
		Content: content,
	}
}
