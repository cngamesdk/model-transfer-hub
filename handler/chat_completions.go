package handler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/cngamesdk/model-transfer-hub/global"
	"github.com/cngamesdk/model-transfer-hub/model"
	"github.com/cngamesdk/model-transfer-hub/pkg/logger"
	"github.com/cngamesdk/model-transfer-hub/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type ChatCompletionsHandler struct {
	proxyService *service.ProxyService
}

func NewChatCompletionsHandler() *ChatCompletionsHandler {
	return &ChatCompletionsHandler{
		proxyService: service.NewProxyService(),
	}
}

// Handle 处理聊天完成请求
func (h *ChatCompletionsHandler) Handle(c *gin.Context) {
	var req model.ChatCompletionRequest
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

	// 判断是否流式请求
	if req.Stream {
		h.handleStream(c, &req, tokenID, tokenName)
	} else {
		h.handleNormal(c, &req, tokenID, tokenName)
	}
}

// handleNormal 处理非流式请求
func (h *ChatCompletionsHandler) handleNormal(c *gin.Context, req *model.ChatCompletionRequest, tokenID int64, tokenName string) {
	resp, err := h.proxyService.ChatCompletion(c.Request.Context(), req, tokenID, tokenName)
	if err != nil {
		logger.Error(c.Request.Context(), global.MTH_LOG, "处理聊天完成请求失败", zap.Error(err))
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

// handleStream 处理流式请求
func (h *ChatCompletionsHandler) handleStream(c *gin.Context, req *model.ChatCompletionRequest, tokenID int64, tokenName string) {
	startTime := time.Now()

	stream, provider, err := h.proxyService.ChatCompletionStream(c.Request.Context(), req, tokenID, tokenName)
	if err != nil {
		logger.Error(c.Request.Context(), global.MTH_LOG, "处理流式聊天完成请求失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{
				Message: "请求处理失败: " + err.Error(),
				Type:    "api_error",
			},
		})
		return
	}
	defer stream.Close()

	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// 用于统计Token使用量
	var totalTokens int
	var errorMsg string
	statusCode := http.StatusOK

	// 流式转发
	c.Stream(func(w io.Writer) bool {
		reader := bufio.NewReader(stream)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					logger.Error(c.Request.Context(), global.MTH_LOG, "读取流式响应失败", zap.Error(err))
					errorMsg = err.Error()
					statusCode = http.StatusInternalServerError
				}
				return false
			}

			// 解析SSE数据，提取Token使用量
			if bytes.HasPrefix(line, []byte("data: ")) {
				data := bytes.TrimPrefix(line, []byte("data: "))
				data = bytes.TrimSpace(data)

				// 跳过 [DONE] 标记
				if string(data) == "[DONE]" {
					// 写入响应
					if _, err := w.Write(line); err != nil {
						logger.Error(c.Request.Context(), global.MTH_LOG, "写入流式响应失败", zap.Error(err))
						return false
					}
					if f, ok := w.(http.Flusher); ok {
						f.Flush()
					}
					continue
				}

				// 方法1：尝试提取OpenAI格式的usage信息（最准确）
				var streamResp model.StreamResponse
				if err := json.Unmarshal(data, &streamResp); err == nil {
					// OpenAI在最后一个chunk包含完整的usage
					// 检查是否包含usage信息
					if streamResp.Usage.TotalTokens > 0 {
						totalTokens = streamResp.Usage.TotalTokens
						logger.Info(c.Request.Context(), global.MTH_LOG, "从usage字段提取Token数量",
							zap.Int("total_tokens", totalTokens))
					}
				}

				// 方法2：Anthropic的内容长度估算（降级方案）
				// 只有在没有usage信息时才累计估算
				if totalTokens == 0 {
					var chunk struct {
						Delta struct {
							Content string `json:"content"`
							Text    string `json:"text"` // Anthropic使用text字段
						} `json:"delta"`
					}
					if err := json.Unmarshal(data, &chunk); err == nil {
						content := chunk.Delta.Content
						if content == "" {
							content = chunk.Delta.Text
						}
						if content != "" {
							// 粗略估算：每4个字符约等于1个token
							totalTokens += len(content) / 4
						}
					}
				}
			}

			// 写入响应
			if _, err := w.Write(line); err != nil {
				logger.Error(c.Request.Context(), global.MTH_LOG, "写入流式响应失败", zap.Error(err))
				errorMsg = err.Error()
				statusCode = http.StatusInternalServerError
				return false
			}

			// 刷新缓冲区
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	})

	// 计算耗时
	duration := time.Since(startTime)

	// 记录流式请求的使用日志
	h.proxyService.RecordStreamUsage(
		c.Request.Context(),
		tokenID,
		tokenName,
		provider,
		req.Model,
		totalTokens,
		startTime,
		int(duration.Milliseconds()),
		statusCode,
		errorMsg,
	)

	logger.Info(c.Request.Context(), global.MTH_LOG, "流式响应完成",
		zap.String("provider", provider),
		zap.String("model", req.Model),
		zap.Int("estimated_tokens", totalTokens),
		zap.Duration("duration", duration),
	)
}
