package logger

import (
	"context"
	"go.uber.org/zap"
)

type ctxKey string

const (
	TraceIDKey ctxKey = "trace_id"
	TokenIDKey ctxKey = "token_id"
)

// WithTraceID 将TraceID存入context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// GetTraceID 从context获取TraceID
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// WithTokenID 将TokenID存入context
func WithTokenID(ctx context.Context, tokenID int64) context.Context {
	return context.WithValue(ctx, TokenIDKey, tokenID)
}

// GetTokenID 从context获取TokenID
func GetTokenID(ctx context.Context) int64 {
	if tokenID, ok := ctx.Value(TokenIDKey).(int64); ok {
		return tokenID
	}
	return 0
}

// Info 记录带TraceID的info日志
func Info(ctx context.Context, logger *zap.Logger, msg string, fields ...zap.Field) {
	traceID := GetTraceID(ctx)
	if traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}
	logger.Info(msg, fields...)
}

// Error 记录带TraceID的error日志
func Error(ctx context.Context, logger *zap.Logger, msg string, fields ...zap.Field) {
	traceID := GetTraceID(ctx)
	if traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}
	logger.Error(msg, fields...)
}

// Warn 记录带TraceID的warn日志
func Warn(ctx context.Context, logger *zap.Logger, msg string, fields ...zap.Field) {
	traceID := GetTraceID(ctx)
	if traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}
	logger.Warn(msg, fields...)
}
