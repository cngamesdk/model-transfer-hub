package core

import (
	"context"
	"fmt"
	"github.com/cngamesdk/model-transfer-hub/global"
	"github.com/cngamesdk/model-transfer-hub/initialize"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// RunServer 运行HTTP服务器
func RunServer() {
	// 设置Gin模式
	if global.MTH_CONFIG.System.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化路由
	router := initialize.Router()

	// 服务器配置
	addr := fmt.Sprintf(":%s", global.MTH_CONFIG.System.Addr)
	server := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   0, // Disabled for SSE streaming support
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	global.MTH_LOG.Info("服务器启动", zap.String("addr", addr))

	// 启动服务器
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			global.MTH_LOG.Fatal("服务器启动失败", zap.Error(err))
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	global.MTH_LOG.Info("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		global.MTH_LOG.Fatal("服务器强制关闭", zap.Error(err))
	}

	global.MTH_LOG.Info("服务器已关闭")
}
