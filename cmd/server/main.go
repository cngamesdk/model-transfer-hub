package main

import (
	"flag"
	"fmt"
	"github.com/cngamesdk/model-transfer-hub/core"
	"github.com/cngamesdk/model-transfer-hub/global"
	"github.com/cngamesdk/model-transfer-hub/initialize"
	"go.uber.org/zap"
	"os"
)

var (
	configFile  string
	showHelp    bool
	showVersion bool
)

const Version = "1.0.0"

func init() {
	// 定义命令行参数
	flag.StringVar(&configFile, "c", "config.yaml", "配置文件路径")
	flag.StringVar(&configFile, "config", "config.yaml", "配置文件路径")
	flag.BoolVar(&showHelp, "h", false, "显示帮助信息")
	flag.BoolVar(&showHelp, "help", false, "显示帮助信息")
	flag.BoolVar(&showVersion, "v", false, "显示版本信息")
	flag.BoolVar(&showVersion, "version", false, "显示版本信息")
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 显示帮助信息
	if showHelp {
		printHelp()
		return
	}

	// 显示版本信息
	if showVersion {
		fmt.Printf("AI Model Transfer Hub v%s\n", Version)
		return
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("错误: 配置文件 '%s' 不存在\n", configFile)
		fmt.Println("请使用 -c 或 --config 参数指定正确的配置文件路径")
		fmt.Println("使用 -h 或 --help 查看帮助信息")
		os.Exit(1)
	}

	fmt.Printf("使用配置文件: %s\n", configFile)

	// 初始化系统
	initializeSystem()

	// 运行服务器
	core.RunServer()
}

// initializeSystem 初始化系统所有组件
func initializeSystem() {
	// 初始化配置（传入配置文件路径）
	global.MTH_VP = core.Viper(configFile)

	// 初始化日志
	global.MTH_LOG = core.Zap()
	zap.ReplaceGlobals(global.MTH_LOG)

	global.MTH_LOG.Info("配置加载完成", zap.String("config_file", configFile))

	// 初始化数据库
	global.MTH_DB = initialize.Gorm()

	global.MTH_LOG.Info("系统初始化完成")
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println("AI Model Transfer Hub - AI模型中转服务")
	fmt.Printf("版本: v%s\n\n", Version)
	fmt.Println("用法:")
	fmt.Println("  server [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -c, --config <file>    指定配置文件路径 (默认: config.yaml)")
	fmt.Println("  -h, --help             显示帮助信息")
	fmt.Println("  -v, --version          显示版本信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  server                           # 使用默认配置文件 config.yaml")
	fmt.Println("  server -c config.prod.yaml       # 使用生产环境配置")
	fmt.Println("  server --config config.dev.yaml  # 使用开发环境配置")
	fmt.Println()
	fmt.Println("环境变量:")
	fmt.Println("  MTH_CONFIG    配置文件路径（优先级低于命令行参数）")
	fmt.Println()
	fmt.Println("文档:")
	fmt.Println("  README.md       - 项目介绍")
	fmt.Println("  DEPLOYMENT.md   - 部署指南")
	fmt.Println()
}
