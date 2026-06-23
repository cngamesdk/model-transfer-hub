package middleware

import (
	"github.com/cngamesdk/model-transfer-hub/global"
	"github.com/cngamesdk/model-transfer-hub/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"
)

// RateLimiter 限流器
type RateLimiter struct {
	tokens map[int64]*TokenBucket
	mutex  sync.RWMutex
}

// TokenBucket 令牌桶
type TokenBucket struct {
	rpm       int       // 每分钟请求数
	rph       int       // 每小时请求数
	minCount  int       // 当前分钟计数
	hourCount int       // 当前小时计数
	minTime   time.Time // 当前分钟开始时间
	hourTime  time.Time // 当前小时开始时间
	mutex     sync.Mutex
}

var limiter = &RateLimiter{
	tokens: make(map[int64]*TokenBucket),
}

// RateLimit 限流中间件
func RateLimit() gin.HandlerFunc {
	if !global.MTH_CONFIG.RateLimit.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		tokenID, exists := c.Get("token_id")
		if !exists {
			c.Next()
			return
		}

		tokenIDInt64 := tokenID.(int64)

		// 获取Token信息
		aiToken, _ := c.Get("token")
		token := aiToken.(*model.AiToken)

		// 获取限流配置
		rpm := global.MTH_CONFIG.RateLimit.DefaultRPM
		rph := global.MTH_CONFIG.RateLimit.DefaultRPH
		if token.RequestLimit > 0 {
			rpm = token.RequestLimit
		}

		// 检查限流
		if !limiter.Allow(tokenIDInt64, rpm, rph) {
			c.JSON(http.StatusTooManyRequests, model.ErrorResponse{
				Error: model.ErrorDetail{
					Message: "请求过于频繁，请稍后再试",
					Type:    "rate_limit_exceeded",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(tokenID int64, rpm, rph int) bool {
	rl.mutex.Lock()
	bucket, exists := rl.tokens[tokenID]
	if !exists {
		bucket = &TokenBucket{
			rpm:      rpm,
			rph:      rph,
			minTime:  time.Now(),
			hourTime: time.Now(),
		}
		rl.tokens[tokenID] = bucket
	}
	rl.mutex.Unlock()

	return bucket.Take()
}

// Take 消费一个令牌
func (tb *TokenBucket) Take() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	now := time.Now()

	// 重置分钟计数
	if now.Sub(tb.minTime) >= time.Minute {
		tb.minCount = 0
		tb.minTime = now
	}

	// 重置小时计数
	if now.Sub(tb.hourTime) >= time.Hour {
		tb.hourCount = 0
		tb.hourTime = now
	}

	// 检查限流
	if tb.rpm > 0 && tb.minCount >= tb.rpm {
		return false
	}
	if tb.rph > 0 && tb.hourCount >= tb.rph {
		return false
	}

	// 消费令牌
	tb.minCount++
	tb.hourCount++

	return true
}
