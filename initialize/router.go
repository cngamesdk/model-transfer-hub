package initialize

import (
	"github.com/cngamesdk/model-transfer-hub/handler"
	"github.com/cngamesdk/model-transfer-hub/middleware"
	"github.com/gin-gonic/gin"
)

// Router 初始化路由
func Router() *gin.Engine {
	router := gin.New()

	// 全局中间件
	router.Use(gin.Recovery())
	router.Use(middleware.Trace())
	router.Use(middleware.Logger())

	// 健康检查（不需要认证）
	healthHandler := &handler.HealthHandler{}
	router.GET("/health", healthHandler.Check)

	// v1 API组（需要认证）
	v1 := router.Group("/v1")
	v1.Use(middleware.TokenAuth())
	v1.Use(middleware.RateLimit())
	{
		// 聊天完成接口
		chatHandler := handler.NewChatCompletionsHandler()
		v1.POST("/chat/completions", chatHandler.Handle)

		// Claude 聊天接口
		v1.POST("/messages", chatHandler.Handle)

		// 文本完成接口
		completionsHandler := handler.NewCompletionsHandler()
		v1.POST("/completions", completionsHandler.Handle)
	}

	return router
}
