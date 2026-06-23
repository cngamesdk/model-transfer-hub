package middleware

import (
	"fmt"
	"github.com/cngamesdk/model-transfer-hub/global"
	"github.com/cngamesdk/model-transfer-hub/model"
	"github.com/cngamesdk/model-transfer-hub/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

// TokenAuth Token验证中间件
func TokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 提取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, model.ErrorResponse{
				Error: model.ErrorDetail{
					Message: "缺少Authorization header",
					Type:    "unauthorized",
				},
			})
			c.Abort()
			return
		}

		// 检查Bearer格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, model.ErrorResponse{
				Error: model.ErrorDetail{
					Message: "Authorization header格式错误，应为: Bearer {token}",
					Type:    "unauthorized",
				},
			})
			c.Abort()
			return
		}

		token := parts[1]

		// 查询数据库验证Token
		var aiToken model.AiToken
		if err := global.MTH_DB.Where("token = ?", token).First(&aiToken).Error; err != nil {
			c.JSON(http.StatusUnauthorized, model.ErrorResponse{
				Error: model.ErrorDetail{
					Message: "Token无效",
					Type:    "invalid_token",
				},
			})
			c.Abort()
			return
		}

		// 检查Token有效性
		if !aiToken.IsValid() {
			message := "Token已禁用"
			if aiToken.ExpireAt != nil && aiToken.ExpireAt.Before(time.Now()) {
				message = "Token已过期"
			} else if aiToken.TokenLimit > 0 && aiToken.UsedTokens >= aiToken.TokenLimit {
				message = fmt.Sprintf("Token配额已用完 (%d/%d)", aiToken.UsedTokens, aiToken.TokenLimit)
			}

			c.JSON(http.StatusForbidden, model.ErrorResponse{
				Error: model.ErrorDetail{
					Message: message,
					Type:    "token_expired",
				},
			})
			c.Abort()
			return
		}

		// 存入context
		ctx := logger.WithTokenID(c.Request.Context(), aiToken.ID)
		c.Request = c.Request.WithContext(ctx)

		c.Set("token_id", aiToken.ID)
		c.Set("token_name", aiToken.Name)
		c.Set("token", &aiToken)

		c.Next()
	}
}
