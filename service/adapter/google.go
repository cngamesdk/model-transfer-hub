package adapter

import (
	"context"
	"fmt"
	"github.com/cngamesdk/model-transfer-hub/config"
	"github.com/cngamesdk/model-transfer-hub/model"
	"io"
)

// GoogleAdapter Google适配器
type GoogleAdapter struct {
	BaseAdapter
}

// NewGoogleAdapter 创建Google适配器
func NewGoogleAdapter(provider *config.Provider) *GoogleAdapter {
	return &GoogleAdapter{
		BaseAdapter: BaseAdapter{
			ProviderName: provider.Name,
			BaseURL:      provider.BaseURL,
			APIKey:       provider.ApiKey,
			Timeout:      provider.Timeout,
		},
	}
}

// ChatCompletion 聊天完成（非流式）
func (a *GoogleAdapter) ChatCompletion(ctx context.Context, req *model.ChatCompletionRequest) (*model.ChatCompletionResponse, error) {
	// TODO: 实现Google Gemini API转换
	// Google使用不同的API格式，需要转换
	// 这里先返回未实现错误
	return nil, ErrNotImplemented
}

// ChatCompletionStream 聊天完成（流式）
func (a *GoogleAdapter) ChatCompletionStream(ctx context.Context, req *model.ChatCompletionRequest) (io.ReadCloser, error) {
	return nil, ErrNotImplemented
}

// Completion 文本完成（非流式）
func (a *GoogleAdapter) Completion(ctx context.Context, req *model.CompletionRequest) (*model.CompletionResponse, error) {
	return nil, ErrNotImplemented
}

// CompletionStream 文本完成（流式）
func (a *GoogleAdapter) CompletionStream(ctx context.Context, req *model.CompletionRequest) (io.ReadCloser, error) {
	return nil, ErrNotImplemented
}

var ErrNotImplemented = fmt.Errorf("Google适配器尚未实现")
