package middleware

import (
	"bytes"
	"github.com/cngamesdk/model-transfer-hub/global"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"time"
)

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 读取请求体
		var body []byte
		if c.Request.Body != nil {
			body, _ = io.ReadAll(c.Request.Body)
			// 重新设置body，因为读取后会被消费
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		// 处理请求
		c.Next()

		// 计算耗时
		cost := time.Since(start)

		// 获取TraceID
		traceID, _ := c.Get("trace_id")

		// 记录日志
		global.MTH_LOG.Info("请求日志",
			zap.String("trace_id", traceID.(string)),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", cost),
			zap.String("body", string(body)),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)
	}
}
