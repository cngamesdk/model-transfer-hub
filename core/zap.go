package core

import (
	"fmt"
	"github.com/cngamesdk/model-transfer-hub/core/internal"
	"github.com/cngamesdk/model-transfer-hub/global"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

// Zap 初始化日志系统
func Zap() *zap.Logger {
	// 确保日志目录存在
	if ok, _ := PathExists(global.MTH_CONFIG.Zap.Director); !ok {
		fmt.Printf("创建日志目录 %v\n", global.MTH_CONFIG.Zap.Director)
		_ = os.Mkdir(global.MTH_CONFIG.Zap.Director, os.ModePerm)
	}

	// 创建不同级别的Core
	cores := internal.Zap()

	logger := zap.New(zapcore.NewTee(cores...), zap.AddCallerSkip(1))

	if global.MTH_CONFIG.Zap.ShowLine {
		logger = logger.WithOptions(zap.AddCaller())
	}

	return logger
}

// PathExists 判断路径是否存在
func PathExists(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil {
		if fi.IsDir() {
			return true, nil
		}
		return false, fmt.Errorf("存在同名文件")
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
