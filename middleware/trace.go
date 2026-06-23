package middleware

import (
	"github.com/cngamesdk/model-transfer-hub/global"
	"github.com/cngamesdk/model-transfer-hub/pkg/logger"
	"github.com/cngamesdk/model-transfer-hub/pkg/trace"
	"github.com/gin-gonic/gin"
)

// Trace 链路追踪中间件
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取TraceID，如果不存在则生成
		traceID := c.GetHeader(global.MTH_CONFIG.Trace.HeaderName)
		if traceID == "" && global.MTH_CONFIG.Trace.GenerateIfMissing {
			traceID = trace.GenerateTraceID()
		}

		// 存入context
		ctx := logger.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		// 设置到gin.Context，方便后续使用
		c.Set("trace_id", traceID)

		// 在响应Header中返回TraceID
		c.Header(global.MTH_CONFIG.Trace.HeaderName, traceID)

		c.Next()
	}
}
