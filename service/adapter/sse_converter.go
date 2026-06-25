package adapter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cngamesdk/model-transfer-hub/model"
)

// anthropicSSEConverter converts Anthropic SSE stream to OpenAI SSE format using io.Pipe.
type anthropicSSEConverter struct {
	body  io.ReadCloser
	pr    *io.PipeReader
	pw    *io.PipeWriter
	model string
}

// Anthropic SSE event types (internal, private)
type anthropicSSEEvent struct {
	Type    string          `json:"type"`
	Message json.RawMessage `json:"message,omitempty"`
	Delta   json.RawMessage `json:"delta,omitempty"`
	Index   int             `json:"index"`
	Usage   struct {
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type anthropicMsgDelta struct {
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
}

type anthropicContentBlockDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// newAnthropicToOpenAIStreamConverter creates a converter wrapping the raw Anthropic response body.
func newAnthropicToOpenAIStreamConverter(body io.ReadCloser, model string) *anthropicSSEConverter {
	pr, pw := io.Pipe()
	c := &anthropicSSEConverter{
		body:  body,
		pr:    pr,
		pw:    pw,
		model: model,
	}
	go c.convert()
	return c
}

func (c *anthropicSSEConverter) Read(p []byte) (int, error) { return c.pr.Read(p) }
func (c *anthropicSSEConverter) Close() error               { return c.body.Close() }

func (c *anthropicSSEConverter) convert() {
	defer c.pw.Close()
	defer c.body.Close()

	scanner := bufio.NewScanner(c.body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var (
		chunkID      = fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
		created      = time.Now().Unix()
		modelName    = c.model
		eventType    string
		dataLines    []string
		finishReason *string
	)

	writeChunk := func(content string, fr *string, usage *model.Usage) {
		chunk := model.StreamResponse{
			ID:      chunkID,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   modelName,
			Choices: []model.StreamChoice{
				{
					Index: 0,
					Delta: model.MessageDelta{
						Content: content,
					},
					FinishReason: fr,
				},
			},
			Usage: usage,
		}
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(c.pw, "data: %s\n\n", string(data))
	}

	writeRoleChunk := func(role string) {
		chunk := model.StreamResponse{
			ID:      chunkID,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   modelName,
			Choices: []model.StreamChoice{
				{
					Index:        0,
					Delta:        model.MessageDelta{Role: role},
					FinishReason: nil,
				},
			},
		}
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(c.pw, "data: %s\n\n", string(data))
	}

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "event: ") {
			eventType = strings.TrimPrefix(line, "event: ")
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			dataLines = append(dataLines, strings.TrimPrefix(line, "data: "))
			continue
		}

		// Blank line terminates an SSE event
		if line == "" && len(dataLines) > 0 {
			dataStr := strings.Join(dataLines, "")
			dataLines = nil

			switch eventType {
			case "message_start":
				var evt anthropicSSEEvent
				if err := json.Unmarshal([]byte(dataStr), &evt); err == nil {
					if evt.Message != nil {
						var msg struct {
							ID    string `json:"id"`
							Role  string `json:"role"`
							Model string `json:"model"`
						}
						if err := json.Unmarshal(evt.Message, &msg); err == nil {
							chunkID = msg.ID
							if msg.Model != "" {
								modelName = msg.Model
							}
							if msg.Role != "" {
								writeRoleChunk(msg.Role)
							}
						}
					}
				}

			case "content_block_delta":
				var evt anthropicSSEEvent
				if err := json.Unmarshal([]byte(dataStr), &evt); err == nil {
					if evt.Delta != nil {
						var delta anthropicContentBlockDelta
						if err := json.Unmarshal(evt.Delta, &delta); err == nil {
							if delta.Type == "text_delta" && delta.Text != "" {
								writeChunk(delta.Text, nil, nil)
							}
						}
					}
				}

			case "message_delta":
				var evt anthropicSSEEvent
				if err := json.Unmarshal([]byte(dataStr), &evt); err == nil {
					var delta anthropicMsgDelta
					if evt.Delta != nil {
						json.Unmarshal(evt.Delta, &delta)
					}
					usage := &model.Usage{
						CompletionTokens: evt.Usage.OutputTokens,
						TotalTokens:      evt.Usage.OutputTokens,
					}
					if delta.StopReason != "" {
						sr := delta.StopReason
						finishReason = &sr
					}
					writeChunk("", finishReason, usage)
				}

			case "message_stop":
				// Will emit [DONE] at end

			case "ping":
				// Ignore ping events
			}

			eventType = ""
		}
	}

	// Scanner error (e.g. connection dropped) is non-fatal for the stream;
	// emit [DONE] regardless so the client sees a clean termination.
	fmt.Fprintf(c.pw, "data: [DONE]\n\n")
}
