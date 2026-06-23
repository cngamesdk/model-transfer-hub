package adapter

import (
	"fmt"
	"github.com/cngamesdk/model-transfer-hub/global"
)

// Factory 适配器工厂
type Factory struct{}

// NewFactory 创建适配器工厂
func NewFactory() *Factory {
	return &Factory{}
}

// GetAdapter 根据模型名称获取适配器
func (f *Factory) GetAdapter(model string) (Adapter, error) {
	// 从配置中查找对应的提供商
	provider := global.MTH_CONFIG.GetProviderByModel(model)
	if provider == nil {
		return nil, fmt.Errorf("未找到模型 %s 对应的提供商", model)
	}

	// 根据提供商名称创建适配器
	switch provider.Name {
	case "openai":
		return NewOpenAIAdapter(provider), nil
	case "anthropic":
		return NewAnthropicAdapter(provider), nil
	case "google":
		return NewGoogleAdapter(provider), nil
	default:
		return nil, fmt.Errorf("不支持的提供商: %s", provider.Name)
	}
}

// GetAdapterByProvider 根据提供商名称获取适配器
func (f *Factory) GetAdapterByProvider(providerName string) (Adapter, error) {
	provider := global.MTH_CONFIG.GetProviderByName(providerName)
	if provider == nil {
		return nil, fmt.Errorf("未找到提供商: %s", providerName)
	}

	switch providerName {
	case "openai":
		return NewOpenAIAdapter(provider), nil
	case "anthropic":
		return NewAnthropicAdapter(provider), nil
	case "google":
		return NewGoogleAdapter(provider), nil
	default:
		return nil, fmt.Errorf("不支持的提供商: %s", providerName)
	}
}
