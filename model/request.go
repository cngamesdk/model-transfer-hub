package model

import "encoding/json"

// ContentPart 内容部分（支持文本和图片）
type ContentPart struct {
	Type   string       `json:"type"` // "text" 或 "image"
	Text   string       `json:"text,omitempty"`
	Source *ImageSource `json:"source,omitempty"`
}

// ImageSource 图片源
type ImageSource struct {
	Type      string `json:"type"`       // "base64" 或 "url"
	MediaType string `json:"media_type"` // "image/jpeg", "image/png" 等
	Data      string `json:"data"`       // base64数据或URL
}

// MessageContent 消息内容（支持字符串或数组）
type MessageContent struct {
	isString bool
	text     string
	parts    []ContentPart
}

// NewMessageContentFromString 从字符串创建MessageContent
func NewMessageContentFromString(s string) MessageContent {
	return MessageContent{
		isString: true,
		text:     s,
	}
}

// NewMessageContentFromParts 从内容部分数组创建MessageContent
func NewMessageContentFromParts(parts []ContentPart) MessageContent {
	return MessageContent{
		isString: false,
		parts:    parts,
	}
}

// UnmarshalJSON 自定义JSON解析
func (mc *MessageContent) UnmarshalJSON(data []byte) error {
	// 先尝试解析为字符串
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		mc.isString = true
		mc.text = s
		return nil
	}

	// 再尝试解析为数组
	var parts []ContentPart
	if err := json.Unmarshal(data, &parts); err != nil {
		return err
	}
	mc.isString = false
	mc.parts = parts
	return nil
}

// MarshalJSON 自定义JSON序列化
func (mc MessageContent) MarshalJSON() ([]byte, error) {
	if mc.isString {
		return json.Marshal(mc.text)
	}
	return json.Marshal(mc.parts)
}

// String 获取纯文本内容
func (mc MessageContent) String() string {
	if mc.isString {
		return mc.text
	}
	// 从数组中提取文本
	var text string
	for _, part := range mc.parts {
		if part.Type == "text" {
			text += part.Text
		}
	}
	return text
}

// IsString 判断是否为字符串类型
func (mc MessageContent) IsString() bool {
	return mc.isString
}

// Parts 获取内容部分数组
func (mc MessageContent) Parts() []ContentPart {
	return mc.parts
}

// Message 消息结构（兼容OpenAI和Claude格式）
type Message struct {
	Role    string         `json:"role"`    // system/user/assistant
	Content MessageContent `json:"content"` // 支持字符串或数组
}

// ChatCompletionRequest 聊天完成请求（OpenAI格式）
type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
}

// CompletionRequest 文本完成请求（OpenAI格式）
type CompletionRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	Stream      bool    `json:"stream,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
}
