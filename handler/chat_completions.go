package handler

import (
	"bytes"
	"context"
	"github.com/cngamesdk/model-transfer-hub/global"
	"github.com/cngamesdk/model-transfer-hub/model"
	"github.com/cngamesdk/model-transfer-hub/pkg/logger"
	"github.com/cngamesdk/model-transfer-hub/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"sync/atomic"
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
	resp, _, err := h.proxyService.ChatCompletion(c.Request.Context(), req, tokenID, tokenName)
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

// handleStream 处理流式请求（直接使用 Gin Writer）
func (h *ChatCompletionsHandler) handleStream(c *gin.Context, req *model.ChatCompletionRequest, tokenID int64, tokenName string) {
	startTime := time.Now()

	// 创建独立的 context，不受 HTTP 请求生命周期影响
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听客户端断开，当客户端真正断开时取消上游请求
	go func() {
		<-c.Request.Context().Done()
		cancel()
	}()

	stream, provider, err := h.proxyService.ChatCompletionStream(ctx, req, tokenID, tokenName)
	if err != nil {
		logger.Error(ctx, global.MTH_LOG, "处理流式聊天完成请求失败", zap.Error(err))
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
	c.Status(http.StatusOK)

	// 初始化Token计数器
	counter := newAsyncTokenCounter(provider, req.Model)
	defer counter.stop()

	// 获取响应写入器
	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		logger.Error(ctx, global.MTH_LOG, "响应不支持流式传输")
		return
	}

	// 流式转发
	buf := make([]byte, 4096)
	streamSuccess := true
	statusCode := http.StatusOK
	errorMsg := ""

	for {
		n, err := stream.Read(buf)
		if n > 0 {
			// 写入数据
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				logger.Error(ctx, global.MTH_LOG, "写流异常", zap.Error(writeErr))
				streamSuccess = false
				statusCode = http.StatusInternalServerError
				errorMsg = "写流异常"
				break
			}

			// 异步处理token统计
			counter.processAsync(buf[:n])

			// 立即刷新
			flusher.Flush()
		}

		if err != nil {
			if err == io.EOF {
				// 正常结束
				break
			}
			// 检查是否是 context 取消（客户端主动断开）
			if ctx.Err() != nil {
				logger.Info(ctx, global.MTH_LOG, "流式传输中断（客户端断开）", zap.Error(err))
			} else {
				logger.Error(ctx, global.MTH_LOG, "读取流异常", zap.Error(err))
			}
			streamSuccess = false
			statusCode = http.StatusInternalServerError
			errorMsg = err.Error()
			break
		}
	}

	if !streamSuccess && statusCode == http.StatusOK {
		statusCode = http.StatusInternalServerError
		errorMsg = "流式传输中断"
	}

	// 计算耗时
	duration := time.Since(startTime)
	totalTokens := counter.getTotalTokens()

	// 异步记录使用日志
	h.proxyService.RecordStreamUsage(
		ctx,
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

	logger.Info(ctx, global.MTH_LOG, "流式响应完成",
		zap.String("provider", provider),
		zap.String("model", req.Model),
		zap.Int("estimated_tokens", totalTokens),
		zap.Duration("duration", duration),
		zap.Bool("success", streamSuccess),
	)
}

// asyncTokenCounter 异步Token计数器
type asyncTokenCounter struct {
	// 使用atomic操作，避免锁竞争
	totalTokens int64

	// 用于批量处理的buffer
	buffer    chan []byte
	done      chan struct{}
	processor *openAITokenProcessor
}

// ✅ 初始化函数
func newAsyncTokenCounter(_, _ string) *asyncTokenCounter {
	counter := &asyncTokenCounter{
		buffer:    make(chan []byte, 100), // 缓冲100个批次
		done:      make(chan struct{}),
		processor: newOpenAITokenProcessor(),
	}

	// 启动后台处理goroutine
	go counter.processLoop()

	return counter
}

// stop 停止计数器
func (c *asyncTokenCounter) stop() {
	close(c.buffer) // 关闭buffer channel，触发processLoop退出
	<-c.done        // 等待处理完成
}

// processAsync 异步处理数据
func (c *asyncTokenCounter) processAsync(data []byte) {
	// 必须复制数据，因为原始buffer会被重用
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	// 非阻塞发送到处理channel
	select {
	case c.buffer <- dataCopy:
	default:
		// 缓冲区满，丢弃（性能优先）
	}
}

// processLoop 后台处理循环
func (c *asyncTokenCounter) processLoop() {
	defer close(c.done)

	for data := range c.buffer {
		c.processBatch(data)
	}
}

// processBatch 批量处理token数据 — all providers now emit OpenAI SSE format
func (c *asyncTokenCounter) processBatch(data []byte) {
	c.processor.process(data, c)
}

// getTotalTokens 获取总token数
func (c *asyncTokenCounter) getTotalTokens() int {
	return int(atomic.LoadInt64(&c.totalTokens))
}

// addTokens 原子操作添加token
func (c *asyncTokenCounter) addTokens(tokens int) {
	atomic.AddInt64(&c.totalTokens, int64(tokens))
}

// genericProcess 通用token处理（fallback）
func (c *asyncTokenCounter) genericProcess(data []byte) {
	if len(data) > 0 {
		tokens := len(data) / 4
		if tokens > 0 {
			c.addTokens(tokens)
		}
	}
}

// openAITokenProcessor OpenAI Token处理器
type openAITokenProcessor struct {
	doneMarker []byte
	dataPrefix []byte
}

func newOpenAITokenProcessor() *openAITokenProcessor {
	return &openAITokenProcessor{
		doneMarker: []byte("[DONE]"),
		dataPrefix: []byte("data: "),
	}
}

// process 处理OpenAI格式 SSE chunk
func (p *openAITokenProcessor) process(data []byte, counter *asyncTokenCounter) {
	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		if !bytes.HasPrefix(line, p.dataPrefix) {
			continue
		}

		jsonData := bytes.TrimSpace(line[len(p.dataPrefix):])

		if bytes.Equal(jsonData, p.doneMarker) {
			return
		}

		p.extractUsage(jsonData, counter)
	}
}

// extractUsage 快速提取usage信息
func (p *openAITokenProcessor) extractUsage(data []byte, counter *asyncTokenCounter) {
	totalIdx := bytes.Index(data, []byte(`"total_tokens":`))
	if totalIdx != -1 {
		numStart := totalIdx + len(`"total_tokens":`)
		var tokens int
		for numStart < len(data) && data[numStart] >= '0' && data[numStart] <= '9' {
			tokens = tokens*10 + int(data[numStart]-'0')
			numStart++
		}

		if tokens > 0 {
			counter.addTokens(tokens)
			return
		}
	}

	p.estimateTokens(data, counter)
}

// estimateTokens 快速估算token数
func (p *openAITokenProcessor) estimateTokens(data []byte, counter *asyncTokenCounter) {
	contentIdx := bytes.Index(data, []byte(`"content":"`))
	if contentIdx == -1 {
		contentIdx = bytes.Index(data, []byte(`"content": "`))
		if contentIdx == -1 {
			return
		}
	}

	contentStart := contentIdx + bytes.Index(data[contentIdx:], []byte(`"`)) + 1
	contentStart = contentIdx + bytes.Index(data[contentIdx:], []byte(`"`)) + 1
	contentEnd := bytes.IndexByte(data[contentStart:], '"')
	if contentEnd == -1 {
		return
	}

	content := data[contentStart : contentStart+contentEnd]

	if len(content) > 0 {
		tokens := (len(content) + 3) / 4
		counter.addTokens(tokens)
	}
}
