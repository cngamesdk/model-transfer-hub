package model

import (
	"time"
)

// AiUsageLog 使用记录表
type AiUsageLog struct {
	ID             int64     `gorm:"primarykey;autoIncrement" json:"id"`
	TraceID        string    `gorm:"type:varchar(64);not null;index:idx_trace_id" json:"trace_id"`
	TokenID        int64     `gorm:"type:bigint;not null;index:idx_token_id" json:"token_id"`
	TokenName      string    `gorm:"type:varchar(255);not null" json:"token_name"`
	Provider       string    `gorm:"type:varchar(32);not null;index:idx_provider_model" json:"provider"`
	Model          string    `gorm:"type:varchar(128);not null;index:idx_provider_model" json:"model"`
	RequestTokens  int       `gorm:"type:int;not null;default:0" json:"request_tokens"`
	ResponseTokens int       `gorm:"type:int;not null;default:0" json:"response_tokens"`
	TotalTokens    int       `gorm:"type:int;not null;default:0" json:"total_tokens"`
	RequestTime    time.Time `gorm:"type:datetime;not null;index:idx_request_time" json:"request_time"`
	ResponseTime   time.Time `gorm:"type:datetime;not null" json:"response_time"`
	DurationMs     int       `gorm:"type:int;not null" json:"duration_ms"`
	StatusCode     int       `gorm:"type:int;not null" json:"status_code"`
	ErrorMsg       string    `gorm:"type:text" json:"error_msg"`
	IP             string    `gorm:"type:varchar(64)" json:"ip"`
	UserAgent      string    `gorm:"type:varchar(512)" json:"user_agent"`
	CreatedAt      time.Time `gorm:"type:datetime;not null;autoCreateTime" json:"created_at"`
}

func (AiUsageLog) TableName() string {
	return "ods_ai_usage_log"
}
