package handler

import (
	"bytes"
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

// handleStream 处理流式请求（高性能版本）
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

	// ✅ 初始化异步Token计数器
	counter := newAsyncTokenCounter(provider, req.Model)

	// 确保counter在流结束后停止
	defer counter.stop()

	// 用于异步收集统计结果
	type streamStats struct {
		totalTokens int
		errorMsg    string
		statusCode  int
	}

	statsCh := make(chan streamStats, 1)

	// 流式转发
	c.Stream(func(w io.Writer) bool {
		streamErr := h.streamForward(w, stream, counter)

		// 流结束后，异步收集统计信息
		go func() {
			stats := streamStats{
				statusCode:  http.StatusOK,
				totalTokens: counter.getTotalTokens(),
			}

			if streamErr {
				stats.statusCode = http.StatusInternalServerError
				stats.errorMsg = "流式传输中断"
			}

			select {
			case statsCh <- stats:
			default:
			}
		}()

		return streamErr
	})

	// 等待统计结果（带超时）
	var stats streamStats
	select {
	case stats = <-statsCh:
		// 获取到统计数据
	case <-time.After(60000 * time.Millisecond):
		// 超时，使用默认值
		stats = streamStats{
			statusCode:  http.StatusOK,
			totalTokens: counter.getTotalTokens(),
		}
	}

	// 计算耗时
	duration := time.Since(startTime)

	// 异步记录使用日志
	h.proxyService.RecordStreamUsage(
		c.Request.Context(),
		tokenID,
		tokenName,
		provider,
		req.Model,
		stats.totalTokens,
		startTime,
		int(duration.Milliseconds()),
		stats.statusCode,
		stats.errorMsg,
	)

	logger.Info(c.Request.Context(), global.MTH_LOG, "流式响应完成",
		zap.String("provider", provider),
		zap.String("model", req.Model),
		zap.Int("estimated_tokens", stats.totalTokens),
		zap.Duration("duration", duration),
	)
}

// streamForward 高性能流转发
func (h *ChatCompletionsHandler) streamForward(w io.Writer, stream io.Reader, counter *asyncTokenCounter) bool {
	// 重用buffer，减少内存分配
	buf := make([]byte, 4096)

	// 类型断言一次，避免重复判断
	flusher, canFlush := w.(http.Flusher)

	for {
		n, err := stream.Read(buf)
		if n > 0 {
			// 写入响应
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				return false
			}

			// 异步处理token统计
			counter.processAsync(buf[:n])

			// 立即刷新
			if canFlush {
				flusher.Flush()
			}
		}

		if err != nil {
			if err == io.EOF {
				return false
			}
			return false
		}
	}
}

// asyncTokenCounter 异步Token计数器
type asyncTokenCounter struct {
	provider string
	model    string

	// 使用atomic操作，避免锁竞争
	inputTokens  int64
	outputTokens int64
	totalTokens  int64

	// 用于批量处理的buffer
	buffer chan []byte
	done   chan struct{}

	// 处理器池
	processor *sseProcessor
}

// ✅ 初始化函数
func newAsyncTokenCounter(provider, model string) *asyncTokenCounter {
	counter := &asyncTokenCounter{
		provider: provider,
		model:    model,
		buffer:   make(chan []byte, 100), // 缓冲100个批次
		done:     make(chan struct{}),
		processor: &sseProcessor{
			openAIProcessor: newOpenAITokenProcessor(),
			claudeProcessor: newClaudeTokenProcessor(),
		},
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

// processBatch 批量处理token数据
func (c *asyncTokenCounter) processBatch(data []byte) {
	// 根据provider选择处理器
	switch {
	case c.provider == "openai" || c.provider == "azure":
		c.processor.openAIProcessor.process(data, c)
	case c.provider == "claude" || c.provider == "anthropic":
		c.processor.claudeProcessor.process(data, c)
	default:
		c.genericProcess(data)
	}
}

// getTotalTokens 获取总token数
func (c *asyncTokenCounter) getTotalTokens() int {
	return int(atomic.LoadInt64(&c.totalTokens))
}

// addTokens 原子操作添加token
func (c *asyncTokenCounter) addTokens(tokens int) {
	atomic.AddInt64(&c.totalTokens, int64(tokens))
}

// genericProcess 通用token处理
func (c *asyncTokenCounter) genericProcess(data []byte) {
	if len(data) > 0 {
		tokens := len(data) / 4
		if tokens > 0 {
			c.addTokens(tokens)
		}
	}
}

// sseProcessor SSE数据处理器
type sseProcessor struct {
	openAIProcessor *openAITokenProcessor
	claudeProcessor *claudeTokenProcessor
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

// process 处理OpenAI格式
func (p *openAITokenProcessor) process(data []byte, counter *asyncTokenCounter) {
	// 按行分割处理
	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		// 查找data:前缀
		if !bytes.HasPrefix(line, p.dataPrefix) {
			continue
		}

		// 提取data内容
		jsonData := bytes.TrimSpace(line[len(p.dataPrefix):])

		// 检查[DONE]标记
		if bytes.Equal(jsonData, p.doneMarker) {
			return
		}

		// 提取usage信息
		p.extractUsage(jsonData, counter)
	}
}

// extractUsage 快速提取usage信息
func (p *openAITokenProcessor) extractUsage(data []byte, counter *asyncTokenCounter) {
	// 查找total_tokens
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

	// 没有精确数据，估算token
	p.estimateTokens(data, counter)
}

// estimateTokens 快速估算token数
func (p *openAITokenProcessor) estimateTokens(data []byte, counter *asyncTokenCounter) {
	// 查找content字段
	contentIdx := bytes.Index(data, []byte(`"content":"`))
	if contentIdx == -1 {
		contentIdx = bytes.Index(data, []byte(`"content": "`))
		if contentIdx == -1 {
			return
		}
	}

	// 提取content内容
	contentStart := contentIdx + bytes.Index(data[contentIdx:], []byte(`"`)) + 1
	contentStart = contentIdx + bytes.Index(data[contentIdx:], []byte(`"`)) + 1
	contentEnd := bytes.IndexByte(data[contentStart:], '"')
	if contentEnd == -1 {
		return
	}

	content := data[contentStart : contentStart+contentEnd]

	// 快速估算（约4字符/token）
	if len(content) > 0 {
		tokens := (len(content) + 3) / 4
		counter.addTokens(tokens)
	}
}

// claudeTokenProcessor Claude Token处理器
type claudeTokenProcessor struct {
	messageStart []byte
	messageDelta []byte
	messageStop  []byte
}

func newClaudeTokenProcessor() *claudeTokenProcessor {
	return &claudeTokenProcessor{
		messageStart: []byte(`"message_start"`),
		messageDelta: []byte(`"message_delta"`),
		messageStop:  []byte(`"message_stop"`),
	}
}

// process 处理Claude格式
func (p *claudeTokenProcessor) process(data []byte, counter *asyncTokenCounter) {
	if bytes.Contains(data, p.messageStart) {
		p.processMessageStart(data, counter)
	} else if bytes.Contains(data, p.messageDelta) {
		p.processMessageDelta(data, counter)
	}
}

func (p *claudeTokenProcessor) processMessageStart(data []byte, counter *asyncTokenCounter) {
	inputIdx := bytes.Index(data, []byte(`"input_tokens":`))
	if inputIdx == -1 {
		return
	}

	numStart := inputIdx + len(`"input_tokens":`)
	var tokens int
	for numStart < len(data) && data[numStart] >= '0' && data[numStart] <= '9' {
		tokens = tokens*10 + int(data[numStart]-'0')
		numStart++
	}

	if tokens > 0 {
		atomic.StoreInt64(&counter.inputTokens, int64(tokens))
	}
}

func (p *claudeTokenProcessor) processMessageDelta(data []byte, counter *asyncTokenCounter) {
	outputIdx := bytes.Index(data, []byte(`"output_tokens":`))
	if outputIdx == -1 {
		return
	}

	numStart := outputIdx + len(`"output_tokens":`)
	var tokens int
	for numStart < len(data) && data[numStart] >= '0' && data[numStart] <= '9' {
		tokens = tokens*10 + int(data[numStart]-'0')
		numStart++
	}

	if tokens > 0 {
		atomic.StoreInt64(&counter.outputTokens, int64(tokens))
		totalTokens := int(atomic.LoadInt64(&counter.inputTokens)) + tokens
		atomic.StoreInt64(&counter.totalTokens, int64(totalTokens))
	}
}
