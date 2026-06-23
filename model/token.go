package model

import (
	"time"
)

// AiToken Token管理表
type AiToken struct {
	ID            int64      `gorm:"primarykey;autoIncrement" json:"id"`
	Token         string     `gorm:"type:varchar(128);not null;uniqueIndex:uk_token" json:"token"`
	Name          string     `gorm:"type:varchar(255);not null;index:idx_name" json:"name"`
	Type          int8       `gorm:"type:tinyint;not null;default:1;comment:Token类型：1-企业 2-个人" json:"type"`
	TokenLimit    int64      `gorm:"type:bigint;not null;default:0;comment:Token数量限制（0=无限制）" json:"token_limit"`
	UsedTokens    int64      `gorm:"type:bigint;not null;default:0;comment:已使用Token数量" json:"used_tokens"`
	RequestLimit  int        `gorm:"type:int;not null;default:0;comment:请求频率限制（次/分钟，0=无限制）" json:"request_limit"`
	ExpireAt      *time.Time `gorm:"type:datetime;comment:过期时间" json:"expire_at"`
	Status        int8       `gorm:"type:tinyint;not null;default:1;index:idx_status;comment:状态：1-启用 2-禁用" json:"status"`
	AllowedModels string     `gorm:"type:json;comment:允许的模型列表" json:"allowed_models"`
	IPWhitelist   string     `gorm:"type:json;comment:IP白名单" json:"ip_whitelist"`
	Creator       string     `gorm:"type:varchar(64)" json:"creator"`
	CreatedAt     time.Time  `gorm:"type:datetime;not null;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"type:datetime;not null;autoUpdateTime" json:"updated_at"`
}

func (AiToken) TableName() string {
	return "dim_ai_token"
}

// IsValid 检查Token是否有效
func (t *AiToken) IsValid() bool {
	if t.Status != 1 {
		return false
	}
	if t.ExpireAt != nil && t.ExpireAt.Before(time.Now()) {
		return false
	}
	if t.TokenLimit > 0 && t.UsedTokens >= t.TokenLimit {
		return false
	}
	return true
}
