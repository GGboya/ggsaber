#!/bin/bash

echo "🚀 Go Saber 算法竞技平台 - 前端演示"
echo "========================================="
echo

# 检查服务器状态
echo "1. 检查服务器状态..."
SERVER_STATUS=$(curl -s http://localhost:8080/health 2>/dev/null)
if [[ $? -eq 0 ]]; then
    echo "✅ 服务器运行正常: $SERVER_STATUS"
else
    echo "❌ 服务器未启动，请先运行: ./bin/saber-server"
    exit 1
fi
echo

echo "2. 前端界面访问地址："
echo "🌐 主页面: http://localhost:8080"
echo "📱 适配移动端和桌面端"
echo

echo "3. 主要功能页面："
echo "👤 用户注册/登录 - 首次访问会显示登录界面"
echo "📊 控制台 - 显示个人信息和统计数据"
echo "⚔️  对战系统 - 实时1v1算法竞技"
echo "📝 题目列表 - 浏览和练习算法题目"
echo "🏆 排行榜 - 查看全球玩家积分排行"
echo

echo "4. 核心特性："
echo "✨ 现代化响应式设计 (Tailwind CSS)"
echo "💻 多语言代码编辑器 (CodeMirror)"
echo "🔄 实时对战通信 (WebSocket)"
echo "🎯 智能匹配系统 (ELO算法)"
echo "🛡️  安全代码判题"
echo

echo "5. 技术栈："
echo "🎨 前端: HTML5 + JavaScript + Tailwind CSS"
echo "⚙️  后端: Go + Gin框架"
echo "💾 数据库: MySQL + Redis"
echo "🔗 通信: REST API + WebSocket"
echo

echo "6. 使用方法："
echo "📖 1. 在浏览器中打开 http://localhost:8080"
echo "👥 2. 注册一个新账号或使用现有账号登录"
echo "🎮 3. 点击「开始对战」进入匹配队列"
echo "⏱️  4. 等待匹配对手，开始实时编程对决"
echo "🏅 5. 最快解决问题的玩家获胜，获得积分奖励"
echo

echo "7. 演示数据："
echo "已预装用户: alice, bob, charlie, david (密码: hello)"
echo "已预装题目: 两数之和, 反转链表"
echo

echo "🎉 现在请打开浏览器访问: http://localhost:8080"
echo "享受你的算法竞技之旅！"
echo

# 打开浏览器 (如果可能)
if command -v xdg-open > /dev/null; then
    echo "正在为您打开浏览器..."
    xdg-open http://localhost:8080 2>/dev/null &
elif command -v open > /dev/null; then
    echo "正在为您打开浏览器..."
    open http://localhost:8080 2>/dev/null &
else
    echo "请手动在浏览器中访问: http://localhost:8080"
fi 