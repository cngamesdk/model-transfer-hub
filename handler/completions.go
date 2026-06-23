package handler

import (
	"github.com/cngamesdk/model-transfer-hub/global"
	"github.com/cngamesdk/model-transfer-hub/model"
	"github.com/cngamesdk/model-transfer-hub/pkg/logger"
	"github.com/cngamesdk/model-transfer-hub/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

type CompletionsHandler struct {
	proxyService *service.ProxyService
}

func NewCompletionsHandler() *CompletionsHandler {
	return &CompletionsHandler{
		proxyService: service.NewProxyService(),
	}
}

// Handle 处理文本完成请求
func (h *CompletionsHandler) Handle(c *gin.Context) {
	var req model.CompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorDetail{
				Message: "请求参数错误: " + err.Error(),
				Type:    "invalid_request_error",
			},
		})
		return
	}

	// 获取Token信息
	tokenID := c.GetInt64("token_id")
	tokenName := c.GetString("token_name")

	// 处理请求
	resp, err := h.proxyService.Completion(c.Request.Context(), &req, tokenID, tokenName)
	if err != nil {
		logger.Error(c.Request.Context(), global.MTH_LOG, "处理文本完成请求失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{
				Message: "请求处理失败: " + err.Error(),
				Type:    "api_error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
