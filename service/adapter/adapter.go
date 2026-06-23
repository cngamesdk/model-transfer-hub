package adapter

import (
	"context"
	"github.com/cngamesdk/model-transfer-hub/model"
	"io"
)

// Adapter AI提供商适配器接口
type Adapter interface {
	// GetProviderName 获取提供商名称
	GetProviderName() string

	// ChatCompletion 聊天完成（非流式）
	ChatCompletion(ctx context.Context, req *model.ChatCompletionRequest) (*model.ChatCompletionResponse, error)

	// ChatCompletionStream 聊天完成（流式）
	ChatCompletionStream(ctx context.Context, req *model.ChatCompletionRequest) (io.ReadCloser, error)

	// Completion 文本完成（非流式）
	Completion(ctx context.Context, req *model.CompletionRequest) (*model.CompletionResponse, error)

	// CompletionStream 文本完成（流式）
	CompletionStream(ctx context.Context, req *model.CompletionRequest) (io.ReadCloser, error)
}

// BaseAdapter 基础适配器（提供公共功能）
type BaseAdapter struct {
	ProviderName string
	BaseURL      string
	APIKey       string
	Timeout      int
}

// GetProviderName 获取提供商名称
func (b *BaseAdapter) GetProviderName() string {
	return b.ProviderName
}
