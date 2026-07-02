package middleware

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cngamesdk/model-transfer-hub/global"
	"github.com/cngamesdk/model-transfer-hub/model"
	"github.com/gin-gonic/gin"
)

// Md5Sign MD5签名验证中间件
func Md5Sign() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := global.MTH_CONFIG.Md5Sign

		// 未启用时直接放行
		if !cfg.Enabled {
			c.Next()
			return
		}

		// 读取请求体
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusUnauthorized, model.ErrorResponse{
				Error: model.ErrorDetail{
					Message: "读取请求体失败",
					Type:    "signature_error",
				},
			})
			c.Abort()
			return
		}
		// 重新设置请求体，供后续handler读取
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// 提取签名相关header
		source := c.GetHeader(cfg.HeaderSource)
		timestampStr := c.GetHeader(cfg.HeaderTimestamp)
		sign := c.GetHeader(cfg.HeaderSign)

		// 验证必要header是否存在
		if source == "" || timestampStr == "" || sign == "" {
			c.JSON(http.StatusUnauthorized, model.ErrorResponse{
				Error: model.ErrorDetail{
					Message: fmt.Sprintf("缺少签名头: %s / %s / %s", cfg.HeaderSource, cfg.HeaderTimestamp, cfg.HeaderSign),
					Type:    "signature_missing",
				},
			})
			c.Abort()
			return
		}

		// 查找来源对应的密钥
		secretKey, ok := cfg.Keys[source]
		if !ok {
			c.JSON(http.StatusUnauthorized, model.ErrorResponse{
				Error: model.ErrorDetail{
					Message: fmt.Sprintf("未知的签名来源: %s", source),
					Type:    "invalid_source",
				},
			})
			c.Abort()
			return
		}

		// 解析时间戳
		timestamp, err := time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			// 兼容Unix时间戳（秒级）
			var ts int64
			_, err = fmt.Sscanf(timestampStr, "%d", &ts)
			if err != nil {
				c.JSON(http.StatusUnauthorized, model.ErrorResponse{
					Error: model.ErrorDetail{
						Message: "时间戳格式错误，请使用Unix时间戳或RFC3339格式",
						Type:    "invalid_timestamp",
					},
				})
				c.Abort()
				return
			}
			timestamp = time.Unix(ts, 0)
		}

		// 检查签名是否过期
		if cfg.TimeoutSeconds > 0 {
			now := time.Now()
			diff := now.Sub(timestamp)
			if diff < 0 {
				diff = -diff
			}
			if diff > time.Duration(cfg.TimeoutSeconds)*time.Second {
				c.JSON(http.StatusUnauthorized, model.ErrorResponse{
					Error: model.ErrorDetail{
						Message: "签名已过期",
						Type:    "signature_expired",
					},
				})
				c.Abort()
				return
			}
		}

		// 计算期望签名: md5(body + timestamp + secretKey)
		expected := fmt.Sprintf("%x", md5.Sum([]byte(string(bodyBytes)+timestampStr+secretKey)))

		// 比较签名（大小写不敏感）
		if !strings.EqualFold(expected, sign) {
			c.JSON(http.StatusUnauthorized, model.ErrorResponse{
				Error: model.ErrorDetail{
					Message: "签名验证失败",
					Type:    "signature_invalid",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
