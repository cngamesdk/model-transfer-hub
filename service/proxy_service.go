package service

import (
	"context"
	"github.com/cngamesdk/model-transfer-hub/global"
	"github.com/cngamesdk/model-transfer-hub/model"
	"github.com/cngamesdk/model-transfer-hub/pkg/logger"
	"github.com/cngamesdk/model-transfer-hub/service/adapter"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"io"
	"time"
)

// ProxyService 代理服务
type ProxyService struct {
	factory *adapter.Factory
}

// NewProxyService 创建代理服务
func NewProxyService() *ProxyService {
	return &ProxyService{
		factory: adapter.NewFactory(),
	}
}

// ChatCompletion 处理聊天完成请求
func (s *ProxyService) ChatCompletion(ctx context.Context, req *model.ChatCompletionRequest, tokenID int64, tokenName string) (*model.ChatCompletionResponse, any, error) {
	startTime := time.Now()

	// 获取适配器
	adp, err := s.factory.GetAdapter(req.Model)
	if err != nil {
		logger.Error(ctx, global.MTH_LOG, "获取适配器失败", zap.Error(err), zap.String("model", req.Model))
		return nil, nil, err
	}

	providerName := adp.GetProviderName()
	logger.Info(ctx, global.MTH_LOG, "开始请求AI服务",
		zap.String("provider", providerName),
		zap.String("model", req.Model),
		zap.Bool("stream", req.Stream),
	)

	// 调用适配器
	resp, selfData, err := adp.ChatCompletion(ctx, req)
	if err != nil {
		logger.Error(ctx, global.MTH_LOG, "AI服务请求失败", zap.Error(err))

		// 记录失败的使用日志
		s.recordUsage(ctx, tokenID, tokenName, providerName, req.Model, 0, 0, 0, startTime, 500, err.Error())

		return nil, nil, err
	}

	// 记录使用日志
	duration := time.Since(startTime)
	s.recordUsage(ctx, tokenID, tokenName, providerName, req.Model, resp.Usage.PromptTokens, resp.Usage.CompletionTokens, int(duration.Milliseconds()), startTime, 200, "")

	// 更新Token使用量
	s.updateTokenUsage(tokenID, resp.Usage.TotalTokens)

	logger.Info(ctx, global.MTH_LOG, "AI服务请求成功",
		zap.String("provider", providerName),
		zap.String("model", req.Model),
		zap.Int("total_tokens", resp.Usage.TotalTokens),
		zap.Duration("duration", duration),
	)

	return resp, selfData, nil
}

// ChatCompletionStream 处理流式聊天完成请求
func (s *ProxyService) ChatCompletionStream(ctx context.Context, req *model.ChatCompletionRequest, tokenID int64, tokenName string) (io.ReadCloser, string, error) {
	// 获取适配器
	adp, err := s.factory.GetAdapter(req.Model)
	if err != nil {
		logger.Error(ctx, global.MTH_LOG, "获取适配器失败", zap.Error(err), zap.String("model", req.Model))
		return nil, "", err
	}

	providerName := adp.GetProviderName()
	logger.Info(ctx, global.MTH_LOG, "开始流式请求AI服务",
		zap.String("provider", providerName),
		zap.String("model", req.Model),
	)

	// 调用适配器
	stream, err := adp.ChatCompletionStream(ctx, req)
	if err != nil {
		logger.Error(ctx, global.MTH_LOG, "流式AI服务请求失败", zap.Error(err))
		return nil, "", err
	}

	return stream, providerName, nil
}

// Completion 处理文本完成请求
func (s *ProxyService) Completion(ctx context.Context, req *model.CompletionRequest, tokenID int64, tokenName string) (*model.CompletionResponse, error) {
	startTime := time.Now()

	// 获取适配器
	adp, err := s.factory.GetAdapter(req.Model)
	if err != nil {
		logger.Error(ctx, global.MTH_LOG, "获取适配器失败", zap.Error(err), zap.String("model", req.Model))
		return nil, err
	}

	providerName := adp.GetProviderName()
	logger.Info(ctx, global.MTH_LOG, "开始请求AI服务",
		zap.String("provider", providerName),
		zap.String("model", req.Model),
	)

	// 调用适配器
	resp, err := adp.Completion(ctx, req)
	if err != nil {
		logger.Error(ctx, global.MTH_LOG, "AI服务请求失败", zap.Error(err))

		// 记录失败的使用日志
		s.recordUsage(ctx, tokenID, tokenName, providerName, req.Model, 0, 0, 0, startTime, 500, err.Error())

		return nil, err
	}

	// 记录使用日志
	duration := time.Since(startTime)
	s.recordUsage(ctx, tokenID, tokenName, providerName, req.Model, resp.Usage.PromptTokens, resp.Usage.CompletionTokens, int(duration.Milliseconds()), startTime, 200, "")

	// 更新Token使用量
	s.updateTokenUsage(tokenID, resp.Usage.TotalTokens)

	logger.Info(ctx, global.MTH_LOG, "AI服务请求成功",
		zap.String("provider", providerName),
		zap.String("model", req.Model),
		zap.Int("total_tokens", resp.Usage.TotalTokens),
		zap.Duration("duration", duration),
	)

	return resp, nil
}

// recordUsage 记录使用日志（异步）
func (s *ProxyService) recordUsage(ctx context.Context, tokenID int64, tokenName, provider, modelName string, promptTokens, completionTokens, durationMs int, requestTime time.Time, statusCode int, errorMsg string) {
	traceID := logger.GetTraceID(ctx)

	usageLog := &model.AiUsageLog{
		TraceID:        traceID,
		TokenID:        tokenID,
		TokenName:      tokenName,
		Provider:       provider,
		Model:          modelName,
		RequestTokens:  promptTokens,
		ResponseTokens: completionTokens,
		TotalTokens:    promptTokens + completionTokens,
		RequestTime:    requestTime,
		ResponseTime:   time.Now(),
		DurationMs:     durationMs,
		StatusCode:     statusCode,
		ErrorMsg:       errorMsg,
	}

	// 异步写入数据库
	go func() {
		if err := global.MTH_DB.Create(usageLog).Error; err != nil {
			global.MTH_LOG.Error("记录使用日志失败",
				zap.Error(err),
				zap.Any("data", usageLog),
				zap.String("trace_id", traceID),
			)
		}
	}()
}

// RecordStreamUsage 记录流式请求的使用日志（公开方法供Handler调用）
func (s *ProxyService) RecordStreamUsage(ctx context.Context, tokenID int64, tokenName, provider, modelName string, promptTokens, completionTokens int, requestTime time.Time, durationMs, statusCode int, errorMsg string) {
	// 记录使用日志
	s.recordUsage(ctx, tokenID, tokenName, provider, modelName, promptTokens, completionTokens, durationMs, requestTime, statusCode, errorMsg)

	// 更新Token使用量
	totalTokens := promptTokens + completionTokens
	if totalTokens > 0 && statusCode == 200 {
		s.updateTokenUsage(tokenID, totalTokens)
	}
}

// updateTokenUsage 更新Token使用量
func (s *ProxyService) updateTokenUsage(tokenID int64, tokens int) {
	go func() {
		if err := global.MTH_DB.Model(&model.AiToken{}).
			Where("id = ?", tokenID).
			Update("used_tokens", gorm.Expr("used_tokens + ?", tokens)).
			Error; err != nil {
			global.MTH_LOG.Error("更新Token使用量失败",
				zap.Error(err),
				zap.Int64("token_id", tokenID),
				zap.Int("tokens", tokens),
			)
		}
	}()
}
