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
		// 聊天完成接口（OpenAI格式）
		chatHandler := handler.NewChatCompletionsHandler()
		v1.POST("/chat/completions", chatHandler.Handle)

		// Claude Messages接口（Anthropic原生格式）
		messagesHandler := handler.NewMessagesHandler()
		v1.POST("/messages", messagesHandler.HandleMessages)

		// 文本完成接口（OpenAI格式）
		completionsHandler := handler.NewCompletionsHandler()
		v1.POST("/completions", completionsHandler.Handle)
	}

	// v1 Token管理API组（MD5签名验证）
	tokenManageHandler := &handler.TokenManageHandler{}
	v1TokenManage := router.Group("/v1/token/manage")
	v1TokenManage.Use(middleware.Md5Sign())
	{
		v1TokenManage.POST("/create", tokenManageHandler.Create)
		v1TokenManage.POST("/update", tokenManageHandler.Update)
		v1TokenManage.POST("/detail", tokenManageHandler.Detail)
		v1TokenManage.POST("/usage", tokenManageHandler.Usage)
	}

	return router
}
