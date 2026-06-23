package config

// Provider AI提供商配置
type Provider struct {
	Name    string   `mapstructure:"name" json:"name" yaml:"name"`
	Enabled bool     `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	BaseURL string   `mapstructure:"base_url" json:"base_url" yaml:"base_url"`
	ApiKey  string   `mapstructure:"api_key" json:"api_key" yaml:"api_key"`
	Timeout int      `mapstructure:"timeout" json:"timeout" yaml:"timeout"` // 超时时间（秒）
	Models  []string `mapstructure:"models" json:"models" yaml:"models"`
}

// GetProviderByName 根据名称获取提供商配置
func (s *Server) GetProviderByName(name string) *Provider {
	for _, p := range s.Providers {
		if p.Name == name && p.Enabled {
			return &p
		}
	}
	return nil
}

// GetProviderByModel 根据模型名称获取提供商配置
func (s *Server) GetProviderByModel(model string) *Provider {
	for _, p := range s.Providers {
		if !p.Enabled {
			continue
		}
		for _, m := range p.Models {
			if m == model {
				return &p
			}
		}
	}
	return nil
}
