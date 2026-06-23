package internal

import (
	"github.com/cngamesdk/model-transfer-hub/global"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"time"
)

// Zap 获取zap.Core列表（按级别分离）
func Zap() []zapcore.Core {
	cores := make([]zapcore.Core, 0, 7)
	config := global.MTH_CONFIG.Zap

	// 获取所有日志级别
	levels := []zapcore.Level{
		zapcore.DebugLevel,
		zapcore.InfoLevel,
		zapcore.WarnLevel,
		zapcore.ErrorLevel,
		zapcore.DPanicLevel,
		zapcore.PanicLevel,
		zapcore.FatalLevel,
	}

	for _, level := range levels {
		if level >= config.TransportLevel() {
			cores = append(cores, NewZapCore(level))
		}
	}

	return cores
}

// NewZapCore 创建zapcore.Core
func NewZapCore(level zapcore.Level) zapcore.Core {
	config := global.MTH_CONFIG.Zap

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: config.StacktraceKey,
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   config.ZapEncodeLevel(),
		EncodeTime: func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(config.Prefix + t.Format("2006-01-02 15:04:05.000"))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 选择编码器
	var encoder zapcore.Encoder
	if config.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 日志输出
	writer := GetWriteSyncer(level)

	// 日志级别过滤
	levelEnabler := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l == level
	})

	return zapcore.NewCore(encoder, writer, levelEnabler)
}

// GetWriteSyncer 获取日志写入器
func GetWriteSyncer(level zapcore.Level) zapcore.WriteSyncer {
	config := global.MTH_CONFIG.Zap

	// 按天分文件夹
	dateDir := time.Now().Format("2006-01-02")
	logDir := filepath.Join(config.Director, dateDir)

	// 确保目录存在
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		_ = os.MkdirAll(logDir, os.ModePerm)
	}

	// 创建Cutter（日志切割器）
	cutter := NewCutter(
		logDir,
		level.String(),
		config.RetentionDay,
		WithMaxSize(int64(config.MaxSize)*1024*1024), // MB转Byte
	)

	// 如果启用控制台输出
	if config.LogInConsole {
		return zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout),
			zapcore.AddSync(cutter),
		)
	}

	return zapcore.AddSync(cutter)
}
