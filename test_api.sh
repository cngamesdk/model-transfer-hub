#!/bin/bash

# AI模型中转站测试脚本

BASE_URL="http://localhost:8080"
TEST_TOKEN="test-token-12345678"

# 检查jq是否安装
HAS_JQ=false
if command -v jq &> /dev/null; then
    HAS_JQ=true
fi

# 格式化JSON输出
format_json() {
    if [ "$HAS_JQ" = true ]; then
        jq .
    else
        # 如果没有jq，使用python格式化（通常系统都有python）
        if command -v python3 &> /dev/null; then
            python3 -m json.tool
        elif command -v python &> /dev/null; then
            python -m json.tool
        else
            # 如果都没有，直接输出
            cat
        fi
    fi
}

echo "=========================================="
echo "AI模型中转站 API测试"
echo "=========================================="
echo ""

# 1. 健康检查
echo "1. 测试健康检查接口"
echo "GET /health"
curl -s -X GET "$BASE_URL/health" | format_json
echo ""
echo ""

# 2. 测试无Token访问（应该失败）
echo "2. 测试无Token访问（预期失败）"
echo "POST /v1/chat/completions (without token)"
curl -s -X POST "$BASE_URL/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"Hello"}]}' | format_json
echo ""
echo ""

# 3. 测试无效Token（应该失败）
echo "3. 测试无效Token（预期失败）"
echo "POST /v1/chat/completions (invalid token)"
curl -s -X POST "$BASE_URL/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer invalid-token-123" \
  -d '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"Hello"}]}' | format_json
echo ""
echo ""

# 4. 测试有效Token（正常请求）
echo "4. 测试有效Token的聊天完成请求"
echo "POST /v1/chat/completions (with valid token)"
curl -s -X POST "$BASE_URL/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TEST_TOKEN" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {"role": "user", "content": "你好，请用一句话介绍自己"}
    ],
    "max_tokens": 100
  }' | format_json
echo ""
echo ""

# 5. 测试流式响应
echo "5. 测试流式响应"
echo "POST /v1/chat/completions (stream=true)"
curl -s -N -X POST "$BASE_URL/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TEST_TOKEN" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {"role": "user", "content": "数到3"}
    ],
    "stream": true
  }' | head -20
echo ""
echo ""

# 6. 测试文本完成接口
echo "6. 测试文本完成接口"
echo "POST /v1/completions"
curl -s -X POST "$BASE_URL/v1/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TEST_TOKEN" \
  -d '{
    "model": "gpt-3.5-turbo",
    "prompt": "Once upon a time",
    "max_tokens": 50
  }' | format_json
echo ""
echo ""

echo "=========================================="
echo "测试完成"
echo "=========================================="
echo ""

# 提示
if [ "$HAS_JQ" = false ]; then
    echo "提示：安装 jq 可以获得更好的JSON格式化输出"
    echo "  macOS: brew install jq"
    echo "  Ubuntu: sudo apt-get install jq"
    echo "  CentOS: sudo yum install jq"
    echo ""
fi
