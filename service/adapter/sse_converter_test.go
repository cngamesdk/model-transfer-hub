package adapter

import (
	"io"
	"strings"
	"testing"
)

func TestAnthropicToOpenAIStreamConverter(t *testing.T) {
	input := `event: message_start
data: {"type":"message_start","message":{"id":"msg_001","type":"message","role":"assistant","model":"claude-opus-4-6","usage":{"input_tokens":10}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" world"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":5}}

event: message_stop
data: {"type":"message_stop"}
`

	converter := newAnthropicToOpenAIStreamConverter(io.NopCloser(strings.NewReader(input)), "claude-opus-4-6")
	defer converter.Close()

	output, err := io.ReadAll(converter)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	outputStr := string(output)

	// Verify role chunk present
	if !strings.Contains(outputStr, `"object":"chat.completion.chunk"`) {
		t.Error("missing chat.completion.chunk object")
	}
	if !strings.Contains(outputStr, `"role":"assistant"`) {
		t.Error("missing role delta")
	}
	if !strings.Contains(outputStr, `"content":"Hello"`) {
		t.Error("missing content delta 'Hello'")
	}
	if !strings.Contains(outputStr, `"content":" world"`) {
		t.Error("missing content delta ' world'")
	}
	if !strings.Contains(outputStr, `"finish_reason":"end_turn"`) {
		t.Error("missing finish_reason")
	}
	if !strings.Contains(outputStr, `[DONE]`) {
		t.Error("missing [DONE] marker")
	}
	if !strings.HasSuffix(strings.TrimSpace(outputStr), "[DONE]") {
		t.Error("output should end with [DONE]")
	}
}

func TestAnthropicToOpenAIStreamConverterIgnoresPing(t *testing.T) {
	input := `event: message_start
data: {"type":"message_start","message":{"id":"msg_002","type":"message","role":"assistant","model":"claude-3","usage":{"input_tokens":2}}}

event: ping
data: {}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"ping test"}}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"stop"},"usage":{"output_tokens":2}}

event: message_stop
data: {"type":"message_stop"}
`

	converter := newAnthropicToOpenAIStreamConverter(io.NopCloser(strings.NewReader(input)), "claude-3")
	defer converter.Close()

	output, err := io.ReadAll(converter)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	outputStr := string(output)

	if !strings.Contains(outputStr, `"content":"ping test"`) {
		t.Error("missing content delta after ping")
	}
	if !strings.Contains(outputStr, `[DONE]`) {
		t.Error("missing [DONE] marker after ping test")
	}
}

func TestAnthropicToOpenAIStreamConverterEmptyStream(t *testing.T) {
	input := `event: message_start
data: {"type":"message_start","message":{"id":"msg_003","type":"message","role":"assistant","model":"claude-3"}}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":0}}

event: message_stop
data: {"type":"message_stop"}
`

	converter := newAnthropicToOpenAIStreamConverter(io.NopCloser(strings.NewReader(input)), "claude-3")
	defer converter.Close()

	output, err := io.ReadAll(converter)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	outputStr := string(output)

	if !strings.Contains(outputStr, "[DONE]") {
		t.Error("empty stream should still end with [DONE]")
	}
}
