package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cngamesdk/model-transfer-hub/config"
	"github.com/cngamesdk/model-transfer-hub/model"
	"io"
	"net/http"
	"time"
)

// AnthropicAdapter Anthropic适配器
type AnthropicAdapter struct {
	BaseAdapter
}

// NewAnthropicAdapter 创建Anthropic适配器
func NewAnthropicAdapter(provider *config.Provider) *AnthropicAdapter {
	return &AnthropicAdapter{
		BaseAdapter: BaseAdapter{
			ProviderName: provider.Name,
			BaseURL:      provider.BaseURL,
			APIKey:       provider.ApiKey,
			Timeout:      provider.Timeout,
		},
	}
}

// AnthropicRequest Anthropic请求格式
type AnthropicRequest struct {
	Model       string          `json:"model"`
	Messages    []model.Message `json:"messages"`
	System      any             `json:"system,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	TopP        float64         `json:"top_p,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

// AnthropicResponse Anthropic响应格式
type AnthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// ChatCompletion 聊天完成（非流式）
func (a *AnthropicAdapter) ChatCompletion(ctx context.Context, req *model.ChatCompletionRequest) (*model.ChatCompletionResponse, any, error) {
	// 转换请求格式
	anthropicReq := AnthropicRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      false,
	}

	if anthropicReq.MaxTokens == 0 {
		anthropicReq.MaxTokens = 4096 // Anthropic要求必须设置max_tokens
	}

	body, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/messages", a.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{
		Timeout: time.Duration(a.Timeout) * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var anthropicResp AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, nil, fmt.Errorf("decode response: %w", err)
	}

	// 转换为OpenAI格式
	return a.convertToOpenAIFormat(&anthropicResp), &anthropicResp, nil
}

// ChatCompletionStream 聊天完成（流式）
func (a *AnthropicAdapter) ChatCompletionStream(ctx context.Context, req *model.ChatCompletionRequest) (io.ReadCloser, error) {
	anthropicReq := AnthropicRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      true,
	}

	if anthropicReq.MaxTokens == 0 {
		anthropicReq.MaxTokens = 4096
	}

	body, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/messages", a.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("x-api-key", a.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// 流式请求不设置超时，避免长时间流式响应被中断
	client := &http.Client{
		Timeout: 0,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	// Convert Anthropic SSE format to OpenAI SSE format
	return newAnthropicToOpenAIStreamConverter(resp.Body, req.Model), nil
}

// MessagesRaw 处理Anthropic原生messages请求（非流式）— 纯透传raw bytes
func (a *AnthropicAdapter) MessagesRaw(ctx context.Context, rawBody []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/messages", a.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(rawBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{
		Timeout: time.Duration(a.Timeout) * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// MessagesStream 处理Anthropic原生messages流式请求 — 纯透传raw bytes
func (a *AnthropicAdapter) MessagesStream(ctx context.Context, rawBody []byte) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/messages", a.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(rawBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("x-api-key", a.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{
		Timeout: 0, // No timeout for streaming
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	// Return raw Anthropic SSE stream — no conversion
	return resp.Body, nil
}

// Completion 文本完成（非流式）
func (a *AnthropicAdapter) Completion(ctx context.Context, req *model.CompletionRequest) (*model.CompletionResponse, error) {
	// Anthropic不直接支持completion接口，转换为chat格式
	chatReq := &model.ChatCompletionRequest{
		Model: req.Model,
		Messages: []model.Message{
			{
				Role:    "user",
				Content: model.NewMessageContentFromString(req.Prompt),
			},
		},
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      false,
	}

	chatResp, _, err := a.ChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	// 转换为Completion格式
	return &model.CompletionResponse{
		ID:      chatResp.ID,
		Object:  "text_completion",
		Created: chatResp.Created,
		Model:   chatResp.Model,
		Choices: []model.Choice{
			{
				Index:        0,
				Text:         chatResp.Choices[0].Message.Content.String(),
				FinishReason: chatResp.Choices[0].FinishReason,
			},
		},
		Usage: chatResp.Usage,
	}, nil
}

// CompletionStream 文本完成（流式）
func (a *AnthropicAdapter) CompletionStream(ctx context.Context, req *model.CompletionRequest) (io.ReadCloser, error) {
	chatReq := &model.ChatCompletionRequest{
		Model: req.Model,
		Messages: []model.Message{
			{
				Role:    "user",
				Content: model.NewMessageContentFromString(req.Prompt),
			},
		},
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      true,
	}

	return a.ChatCompletionStream(ctx, chatReq)
}

// convertToOpenAIFormat 转换为OpenAI格式
func (a *AnthropicAdapter) convertToOpenAIFormat(resp *AnthropicResponse) *model.ChatCompletionResponse {
	content := ""
	if len(resp.Content) > 0 {
		content = resp.Content[0].Text
	}

	return &model.ChatCompletionResponse{
		ID:      resp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   resp.Model,
		Choices: []model.Choice{
			{
				Index: 0,
				Message: model.Message{
					Role:    "assistant",
					Content: model.NewMessageContentFromString(content),
				},
				FinishReason: resp.StopReason,
			},
		},
		Usage: model.Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}
}
