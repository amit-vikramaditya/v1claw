package providers

import "context"

type ToolCall struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type,omitempty"`
	Function  *FunctionCall          `json:"function,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type LLMResponse struct {
	Content      string     `json:"content"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	FinishReason string     `json:"finish_reason"`
	Usage        *UsageInfo `json:"usage,omitempty"`
}

type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// StreamChunk carries a single streamed token or the final done signal.
type StreamChunk struct {
	Text string // Token text; empty when Done is true
	Done bool   // True when the stream has finished
}

// LLMProvider is the core blocking chat interface.
type LLMProvider interface {
	Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error)
	GetDefaultModel() string
}

// StreamingProvider is an optional extension of LLMProvider that supports
// token-by-token SSE streaming. Callers should type-assert to check support.
type StreamingProvider interface {
	LLMProvider
	// Stream sends messages and yields StreamChunks over the returned channel.
	// The channel is closed after the chunk with Done=true is sent.
	Stream(ctx context.Context, messages []Message, model string, options map[string]interface{}) (<-chan StreamChunk, error)
}

type ToolDefinition struct {
	Type     string                 `json:"type"`
	Function ToolFunctionDefinition `json:"function"`
}

type ToolFunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}
