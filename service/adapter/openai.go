package adapter

import (
	"context"
	"github.com/cngamesdk/model-transfer-hub/config"
	"github.com/cngamesdk/model-transfer-hub/model"
	"io"
)

// OpenAIAdapter OpenAI适配器
type OpenAIAdapter struct {
	BaseAdapter
}

// NewOpenAIAdapter 创建OpenAI适配器
func NewOpenAIAdapter(provider *config.Provider) *OpenAIAdapter {
	return &OpenAIAdapter{
		BaseAdapter: BaseAdapter{
			ProviderName: provider.Name,
			BaseURL:      provider.BaseURL,
			APIKey:       provider.ApiKey,
			Timeout:      provider.Timeout,
		},
	}
}

// ChatCompletion 聊天完成（非流式）
func (a *OpenAIAdapter) ChatCompletion(ctx context.Context, req *model.ChatCompletionRequest) (*model.ChatCompletionResponse, error) {
	// OpenAI格式本身就是标准格式，直接转发
	return sendChatCompletionRequest(ctx, a.BaseURL, a.APIKey, a.Timeout, req)
}

// ChatCompletionStream 聊天完成（流式）
func (a *OpenAIAdapter) ChatCompletionStream(ctx context.Context, req *model.ChatCompletionRequest) (io.ReadCloser, error) {
	return sendChatCompletionStreamRequest(ctx, a.BaseURL, a.APIKey, a.Timeout, req)
}

// Completion 文本完成（非流式）
func (a *OpenAIAdapter) Completion(ctx context.Context, req *model.CompletionRequest) (*model.CompletionResponse, error) {
	return sendCompletionRequest(ctx, a.BaseURL, a.APIKey, a.Timeout, req)
}

// CompletionStream 文本完成（流式）
func (a *OpenAIAdapter) CompletionStream(ctx context.Context, req *model.CompletionRequest) (io.ReadCloser, error) {
	return sendCompletionStreamRequest(ctx, a.BaseURL, a.APIKey, a.Timeout, req)
}
