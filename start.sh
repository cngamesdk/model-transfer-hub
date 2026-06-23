#!/bin/bash

# AI模型中转站启动脚本（支持参数）

# 默认配置文件
CONFIG_FILE="config.yaml"

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        -h|--help)
            echo "AI Model Transfer Hub - 启动脚本"
            echo ""
            echo "用法:"
            echo "  $0 [选项]"
            echo ""
            echo "选项:"
            echo "  -c, --config <file>    指定配置文件路径 (默认: config.yaml)"
            echo "  -h, --help             显示帮助信息"
            echo ""
            echo "示例:"
            echo "  $0                          # 使用默认配置"
            echo "  $0 -c config.prod.yaml      # 使用生产配置"
            echo "  $0 --config config.dev.yaml # 使用开发配置"
            exit 0
            ;;
        *)
            echo "未知参数: $1"
            echo "使用 -h 或 --help 查看帮助"
            exit 1
            ;;
    esac
done

echo "======================================"
echo "AI模型中转站 - 快速启动"
echo "配置文件: $CONFIG_FILE"
echo "======================================"
echo ""

# 检查配置文件
if [ ! -f "$CONFIG_FILE" ]; then
    echo "❌ 错误: 配置文件 '$CONFIG_FILE' 不存在"
    echo "请确保配置文件存在，或使用 -c 参数指定正确的路径"
    exit 1
fi

# 检查数据库连接（可选）
echo "检查配置文件..."
echo ""

# 编译项目
echo "编译项目..."
go build -o bin/server cmd/server/main.go
if [ $? -ne 0 ]; then
    echo "❌ 编译失败"
    exit 1
fi
echo "✅ 编译成功"
echo ""

# 启动服务
echo "======================================"
echo "启动服务..."
echo "======================================"
echo ""
./bin/server -c "$CONFIG_FILE"
