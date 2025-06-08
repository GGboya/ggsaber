#!/bin/bash

# API测试脚本
SERVER_URL="http://localhost:8080"

echo "🚀 开始测试Saber系统API..."

# 测试健康检查
echo "📊 测试健康检查..."
curl -s "$SERVER_URL/health" | jq .
echo ""

# 测试用户注册
echo "👤 测试用户注册..."
curl -s -X POST "$SERVER_URL/api/v1/users/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "hello"
  }' | jq .
echo ""

# 测试用户登录
echo "🔐 测试用户登录..."
curl -s -X POST "$SERVER_URL/api/v1/users/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser", 
    "password": "hello"
  }' | jq .
echo ""

# 测试获取排行榜
echo "🏆 测试获取排行榜..."
curl -s "$SERVER_URL/api/v1/users/leaderboard?limit=5" | jq .
echo ""

# 测试获取题目列表
echo "📝 测试获取题目列表..."
curl -s "$SERVER_URL/api/v1/problems?page=1&page_size=3" | jq .
echo ""

# 测试加入匹配队列
echo "🎯 测试加入匹配队列..."
curl -s -X POST "$SERVER_URL/api/v1/match/join" \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1}' | jq .
echo ""

# 测试获取匹配状态
echo "⏳ 测试获取匹配状态..."
curl -s "$SERVER_URL/api/v1/match/status?user_id=1" | jq .
echo ""

# 测试离开匹配队列
echo "🚪 测试离开匹配队列..."
curl -s -X DELETE "$SERVER_URL/api/v1/match/leave" \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1}' | jq .
echo ""

# 测试语法检查
echo "🔍 测试语法检查..."
curl -s -X POST "$SERVER_URL/api/v1/judge/check" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "package main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}",
    "language": "go"
  }' | jq .
echo ""

# 测试获取支持的编程语言
echo "💻 测试获取支持的编程语言..."
curl -s "$SERVER_URL/api/v1/judge/languages" | jq .
echo ""

echo "✅ API测试完成！" 