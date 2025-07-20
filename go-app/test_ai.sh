#!/bin/bash

echo "🧪 测试 AI 服务..."
echo "=================="

# 测试 AI 服务
echo "1. 测试 AI 服务..."
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"query":"test","user_id":"00000000-0000-0000-0000-000000000000"}' \
  http://localhost:8080/api/v1/interpret

echo -e "\n"

echo "✅ AI 服务测试完成！" 