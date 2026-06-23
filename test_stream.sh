#!/bin/bash

# 流式请求测试脚本

BASE_URL="http://localhost:8080"
TEST_TOKEN="test-token-12345678"

echo "=========================================="
echo "流式响应测试"
echo "=========================================="
echo ""

# 测试流式聊天
echo "测试流式聊天完成"
echo "POST /v1/chat/completions (stream=true)"
echo ""

curl -N -X POST "$BASE_URL/v1/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TEST_TOKEN" \
  -d '{
    "model": "claude-opus-4-6",
    "messages": [
      {"role": "user", "content": "请数到5"}
    ],
    "stream": true,
    "max_tokens": 50
  }'

echo ""
echo ""
echo "=========================================="
echo "测试完成"
echo "=========================================="