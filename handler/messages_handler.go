package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/cngamesdk/model-transfer-hub/global"
	"github.com/cngamesdk/model-transfer-hub/model"
	"github.com/cngamesdk/model-transfer-hub/pkg/logger"
	"github.com/cngamesdk/model-transfer-hub/service"
	"github.com/cngamesdk/model-transfer-hub/service/adapter"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MessagesHandler handles Anthropic-native /v1/messages requests — pure passthrough.
type MessagesHandler struct {
	proxyService *service.ProxyService
	factory      *adapter.Factory
}

func NewMessagesHandler() *MessagesHandler {
	return &MessagesHandler{
		proxyService: service.NewProxyService(),
		factory:      adapter.NewFactory(),
	}
}

// HandleMessages handles /v1/messages with raw bytes passthrough.
func (h *MessagesHandler) HandleMessages(c *gin.Context) {
	// Read raw body for passthrough
	rawBody, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorDetail{
				Message: "读取请求体失败: " + err.Error(),
				Type:    "invalid_request_error",
			},
		})
		return
	}

	// Extract just model and stream for routing
	var meta struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}
	if err := json.Unmarshal(rawBody, &meta); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorDetail{
				Message: "解析请求参数失败: " + err.Error(),
				Type:    "invalid_request_error",
			},
		})
		return
	}

	tokenID := c.GetInt64("token_id")
	tokenName := c.GetString("token_name")

	if meta.Stream {
		h.handleStream(c, meta.Model, rawBody, tokenID, tokenName)
	} else {
		h.handleNormal(c, meta.Model, rawBody, tokenID, tokenName)
	}
}

func (h *MessagesHandler) handleNormal(c *gin.Context, modelName string, rawBody []byte, tokenID int64, tokenName string) {
	startTime := time.Now()

	adp, err := h.factory.GetAdapterByProvider("anthropic")
	if err != nil {
		logger.Error(c.Request.Context(), global.MTH_LOG, "获取适配器失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Message: "请求处理失败: " + err.Error(), Type: "api_error"},
		})
		return
	}

	anthropicAdapter, ok := adp.(*adapter.AnthropicAdapter)
	if !ok {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Message: "不支持的适配器类型", Type: "api_error"},
		})
		return
	}

	respBytes, err := anthropicAdapter.MessagesRaw(c.Request.Context(), rawBody)
	if err != nil {
		logger.Error(c.Request.Context(), global.MTH_LOG, "Messages请求失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Message: "请求处理失败: " + err.Error(), Type: "api_error"},
		})
		return
	}

	// Log usage using raw bytes
	duration := time.Since(startTime)
	inputTokens, outputTokens := extractUsageFromResponse(respBytes)
	h.proxyService.RecordStreamUsage(
		c.Request.Context(), tokenID, tokenName,
		"anthropic", modelName, inputTokens, outputTokens,
		startTime, int(duration.Milliseconds()), http.StatusOK, "",
	)

	logger.Info(c.Request.Context(), global.MTH_LOG, "Messages请求完成",
		zap.String("model", modelName),
		zap.Int("input_tokens", inputTokens),
		zap.Int("output_tokens", outputTokens),
		zap.Duration("duration", duration),
	)

	c.Data(http.StatusOK, "application/json", respBytes)
}

func (h *MessagesHandler) handleStream(c *gin.Context, modelName string, rawBody []byte, tokenID int64, tokenName string) {
	startTime := time.Now()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	go func() {
		<-c.Request.Context().Done()
		cancel()
	}()

	adp, err := h.factory.GetAdapterByProvider("anthropic")
	if err != nil {
		logger.Error(ctx, global.MTH_LOG, "获取适配器失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Message: "请求处理失败: " + err.Error(), Type: "api_error"},
		})
		return
	}

	anthropicAdapter, ok := adp.(*adapter.AnthropicAdapter)
	if !ok {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Message: "不支持的适配器类型", Type: "api_error"},
		})
		return
	}

	stream, err := anthropicAdapter.MessagesStream(ctx, rawBody)
	if err != nil {
		logger.Error(ctx, global.MTH_LOG, "Messages流式请求失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{Message: "请求处理失败: " + err.Error(), Type: "api_error"},
		})
		return
	}
	defer stream.Close()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}

	buf := make([]byte, 4096)
	streamSuccess := true
	statusCode := http.StatusOK
	errorMsg := ""
	var tokenCounter streamTokenCounter

	for {
		n, err := stream.Read(buf)
		if n > 0 {
			tokenCounter.scan(buf[:n])
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				streamSuccess = false
				statusCode = http.StatusInternalServerError
				errorMsg = "写流异常"
				break
			}
			flusher.Flush()
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			streamSuccess = false
			statusCode = http.StatusInternalServerError
			errorMsg = err.Error()
			break
		}
	}

	duration := time.Since(startTime)
	h.proxyService.RecordStreamUsage(
		ctx, tokenID, tokenName,
		"anthropic", modelName,
		int(tokenCounter.inputTokens.Load()),
		int(tokenCounter.outputTokens.Load()),
		startTime, int(duration.Milliseconds()), statusCode, errorMsg,
	)

	logger.Info(ctx, global.MTH_LOG, "Messages流式响应完成",
		zap.String("model", modelName),
		zap.Int64("input_tokens", tokenCounter.inputTokens.Load()),
		zap.Int64("output_tokens", tokenCounter.outputTokens.Load()),
		zap.Duration("duration", duration),
		zap.Bool("success", streamSuccess),
	)
}

// streamTokenCounter tracks input/output tokens from Anthropic SSE stream data.
type streamTokenCounter struct {
	inputTokens  atomic.Int64
	outputTokens atomic.Int64
}

func (c *streamTokenCounter) scan(chunk []byte) {
	if idx := bytes.Index(chunk, []byte(`"input_tokens":`)); idx != -1 {
		c.inputTokens.Store(int64(jsonIntAt(chunk, idx+len(`"input_tokens":`))))
	}
	if idx := bytes.Index(chunk, []byte(`"output_tokens":`)); idx != -1 {
		c.outputTokens.Store(int64(jsonIntAt(chunk, idx+len(`"output_tokens":`))))
	}
}

// extractUsageFromResponse extracts input/output tokens from raw Anthropic response bytes.
func extractUsageFromResponse(data []byte) (int, int) {
	var inputTokens, outputTokens int
	if idx := jsonIndex(data, `"input_tokens":`); idx != -1 {
		inputTokens = jsonIntAt(data, idx)
	}
	if idx := jsonIndex(data, `"output_tokens":`); idx != -1 {
		outputTokens = jsonIntAt(data, idx)
	}
	return inputTokens, outputTokens
}

func jsonIndex(data []byte, key string) int {
	for i := 0; i < len(data)-len(key); i++ {
		if string(data[i:i+len(key)]) == key {
			return i + len(key)
		}
	}
	return -1
}

func jsonIntAt(data []byte, start int) int {
	var n int
	for start < len(data) && data[start] >= '0' && data[start] <= '9' {
		n = n*10 + int(data[start]-'0')
		start++
	}
	return n
}
