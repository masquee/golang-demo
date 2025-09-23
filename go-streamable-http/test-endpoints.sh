#!/bin/bash

echo "🧪 测试 Go Streamable HTTP API 端点..."
echo "========================================"

BASE_URL="http://localhost:8080"

echo ""
echo "1. 测试 JSON 流式响应..."
echo "curl $BASE_URL/stream/json?count=3"
echo "----------------------------------------"
curl -s "$BASE_URL/stream/json?count=3"
echo ""
echo ""

echo "2. 测试文本流式响应..."
echo "curl $BASE_URL/stream/text"
echo "----------------------------------------"
timeout 5s curl -s "$BASE_URL/stream/text" || echo "已超时停止"
echo ""
echo ""

echo "3. 测试 Server-Sent Events..."
echo "curl $BASE_URL/sse"
echo "----------------------------------------"
timeout 8s curl -s "$BASE_URL/sse" || echo "已超时停止"
echo ""
echo ""

echo "✅ 所有端点测试完成！"
echo "💡 提示：访问 $BASE_URL 查看交互式网页界面"
