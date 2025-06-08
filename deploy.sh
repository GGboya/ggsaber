#!/bin/bash

# Go Saber 系统简化部署脚本 - MySQL版本
# 直接构建和运行，不使用Docker

export PATH=$PATH:/usr/local/go/bin:/usr/bin/go/bin
set -e

echo "🚀 开始简化部署 Go Saber 系统 (MySQL版本)..."

# 检查必要的工具
echo "🔍 检查系统环境..."
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go 1.21+"
    exit 1
fi

if ! command -v mysql &> /dev/null; then
    echo "❌ MySQL 未安装，正在安装..."
    sudo apt-get update
    sudo apt-get install -y mysql-server mysql-client
    
    # 启动MySQL并设置开机自启
    sudo systemctl start mysql
    sudo systemctl enable mysql
    
    echo "✅ MySQL 安装完成"
fi

if ! command -v redis-server &> /dev/null; then
    echo "❌ Redis 未安装，正在安装..."
    sudo apt-get install -y redis-server
fi

# 生成配置
echo "🔧 生成配置文件..."
DB_PASSWORD="123456"  # 使用固定密码，与配置文件一致
JWT_SECRET=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-32)

cat > .env << EOF
DB_PASSWORD=$DB_PASSWORD
JWT_SECRET=$JWT_SECRET
ENVIRONMENT=production
PORT=8080
EOF

echo "✅ 配置文件已生成"

# 启动数据库服务
echo "🗄️ 启动数据库服务..."
sudo systemctl start mysql
sudo systemctl enable mysql
sudo systemctl start redis-server
sudo systemctl enable redis-server

# 创建数据库和用户
echo "💾 创建数据库..."

# 检查MySQL root用户是否有密码
if sudo mysql -e "SELECT 1;" 2>/dev/null; then
    # 无密码的root用户
    echo "🔧 使用无密码root用户配置MySQL..."
    sudo mysql -e "CREATE DATABASE IF NOT EXISTS saber CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" || echo "数据库可能已存在"
    sudo mysql -e "DROP USER IF EXISTS 'saber'@'localhost';" 2>/dev/null || true
    sudo mysql -e "CREATE USER 'saber'@'localhost' IDENTIFIED BY '$DB_PASSWORD';"
    sudo mysql -e "GRANT ALL PRIVILEGES ON saber.* TO 'saber'@'localhost';"
    sudo mysql -e "FLUSH PRIVILEGES;"
else
    # 需要密码的root用户
    echo "🔐 MySQL root用户需要密码，请输入root密码："
    read -s -p "MySQL root密码 (留空表示无密码): " MYSQL_ROOT_PASSWORD
    echo ""
    
    if [ -z "$MYSQL_ROOT_PASSWORD" ]; then
        # 尝试无密码
        mysql -u root -e "CREATE DATABASE IF NOT EXISTS saber CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" || echo "数据库可能已存在"
        mysql -u root -e "DROP USER IF EXISTS 'saber'@'localhost';" 2>/dev/null || true
        mysql -u root -e "CREATE USER 'saber'@'localhost' IDENTIFIED BY '$DB_PASSWORD';"
        mysql -u root -e "GRANT ALL PRIVILEGES ON saber.* TO 'saber'@'localhost';"
        mysql -u root -e "FLUSH PRIVILEGES;"
    else
        # 使用提供的密码
        mysql -u root -p"$MYSQL_ROOT_PASSWORD" -e "CREATE DATABASE IF NOT EXISTS saber CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" || echo "数据库可能已存在"
        mysql -u root -p"$MYSQL_ROOT_PASSWORD" -e "DROP USER IF EXISTS 'saber'@'localhost';" 2>/dev/null || true
        mysql -u root -p"$MYSQL_ROOT_PASSWORD" -e "CREATE USER 'saber'@'localhost' IDENTIFIED BY '$DB_PASSWORD';"
        mysql -u root -p"$MYSQL_ROOT_PASSWORD" -e "GRANT ALL PRIVILEGES ON saber.* TO 'saber'@'localhost';"
        mysql -u root -p"$MYSQL_ROOT_PASSWORD" -e "FLUSH PRIVILEGES;"
    fi
fi

echo "✅ 数据库配置完成"

# 设置环境变量
echo "🔄 设置环境变量..."
export DATABASE_DSN="saber:$DB_PASSWORD@tcp(localhost:3306)/saber?charset=utf8mb4&parseTime=True&loc=Local"
export SERVER_PORT="8080"
export REDIS_ADDR="localhost:6379"

echo "DATABASE_DSN=$DATABASE_DSN"

# 构建应用
echo "🏗️ 构建应用..."
export GOPROXY=https://goproxy.cn,direct
go mod tidy
go build -o saber-server cmd/server/main.go

# 运行数据库初始化
echo "🔧 初始化数据库..."
if [ -f "scripts/create_tables.sql" ]; then
    echo "正在执行 create_tables.sql..."
    mysql -u saber -p$DB_PASSWORD saber < scripts/create_tables.sql || echo "create_tables.sql 执行失败，跳过"
fi

if [ -f "scripts/insert_data.sql" ]; then
    echo "正在执行 insert_data.sql..."
    mysql -u saber -p$DB_PASSWORD saber < scripts/insert_data.sql || echo "insert_data.sql 执行失败，跳过"
fi

# 添加反转链表问题
echo "📝 添加示例问题..."
if [ -f "scripts/add_reverse_linked_list.go" ]; then
    go run scripts/add_reverse_linked_list.go
fi

# 创建systemd服务
echo "📋 创建系统服务..."
sudo tee /etc/systemd/system/saber-server.service > /dev/null << EOF
[Unit]
Description=Go Saber Server
After=network.target mysql.service redis.service

[Service]
Type=simple
User=$USER
WorkingDirectory=$(pwd)
ExecStart=$(pwd)/saber-server
Restart=always
RestartSec=5
Environment=DATABASE_DSN=saber:$DB_PASSWORD@tcp(localhost:3306)/saber?charset=utf8mb4&parseTime=True&loc=Local
Environment=SERVER_PORT=8080
Environment=REDIS_ADDR=localhost:6379

[Install]
WantedBy=multi-user.target
EOF

# 启动服务
echo "🚀 启动服务..."
sudo systemctl daemon-reload
sudo systemctl enable saber-server
sudo systemctl start saber-server

# 配置Nginx (可选)
echo "🌐 配置Nginx..."
if command -v nginx &> /dev/null; then
    sudo tee /etc/nginx/sites-available/saber-server > /dev/null << 'EOF'
# Go Saber 服务器配置
server {
    listen 80;
    server_name cnggboy.com www.cnggboy.com;

    # 日志配置
    access_log /var/log/nginx/saber_access.log;
    error_log /var/log/nginx/saber_error.log;

    # 客户端上传大小限制
    client_max_body_size 10M;

    # 静态文件配置 - 代理到Go服务器
    location /static/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 缓存配置
        expires 1h;
        add_header Cache-Control "public";
    }

    # WebSocket 支持
    location /ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 86400;
    }

    # API 路由
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 超时配置
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }

    # 健康检查
    location /health {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        access_log off;
    }

    # 主页面和其他所有请求
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

    sudo ln -sf /etc/nginx/sites-available/saber-server /etc/nginx/sites-enabled/
    sudo nginx -t && sudo systemctl reload nginx

    echo "✅ Nginx 配置完成，域名: saber.cnggboy.com"
else
    echo "⚠️ Nginx 未安装，跳过反向代理配置"
fi

# 检查服务状态
echo "🔍 检查服务状态..."
sleep 5
sudo systemctl status saber-server --no-pager

# 显示结果
echo ""
echo "🎉 部署完成！"
echo ""
echo "📊 服务信息:"
echo "  - 应用端口: 8080"
echo "  - 数据库: MySQL (saber)"
echo "  - 缓存: Redis"
echo ""
echo "🌐 访问地址:"
if command -v nginx &> /dev/null; then
    echo "  - HTTP: http://$(hostname -I | awk '{print $1}')"
    echo "  - 域名: http://saber.cnggboy.com"
    echo "  - 直接访问: http://$(hostname -I | awk '{print $1}'):8080"
else
    echo "  - HTTP: http://$(hostname -I | awk '{print $1}'):8080"
fi
echo ""
echo "📋 管理命令:"
echo "  - 查看日志: sudo journalctl -u saber-server -f"
echo "  - 重启服务: sudo systemctl restart saber-server"
echo "  - 停止服务: sudo systemctl stop saber-server"
echo "  - 检查状态: sudo systemctl status saber-server"
echo "  - 查看数据库: mysql -u saber -p saber"
echo ""

# 健康检查
echo "🔍 健康检查..."
sleep 3
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ 应用运行正常"
else
    echo "❌ 应用可能未正常启动，请检查日志"
    echo "日志命令: sudo journalctl -u saber-server -f"
fi

echo "✨ MySQL版本部署完成！" 