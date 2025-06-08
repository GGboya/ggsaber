#!/bin/bash

# Go Saber 系统部署脚本
# 使用方法: ./deploy.sh [production|staging]

set -e

# 配置
ENVIRONMENT=${1:-production}
PROJECT_NAME="go-saber-system"
DOMAIN="your-domain.com"
EMAIL="your-email@example.com"

echo "🚀 开始部署 Go Saber 系统到 $ENVIRONMENT 环境..."

# 检查 Docker 和 Docker Compose
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装，请先安装 Docker"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose 未安装，请先安装 Docker Compose"
    exit 1
fi

# 创建必要的目录
echo "📁 创建必要的目录..."
mkdir -p logs ssl static

# 设置环境变量
echo "🔧 设置环境变量..."
if [ ! -f .env ]; then
    cat > .env << EOF
# 数据库配置
DB_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 64)

# 域名配置
DOMAIN=$DOMAIN
EMAIL=$EMAIL

# 环境
ENVIRONMENT=$ENVIRONMENT
EOF
    echo "✅ 创建了 .env 文件，请根据需要修改配置"
fi

source .env

# 更新配置文件中的占位符
echo "🔄 更新配置文件..."
sed -i "s/your-domain.com/$DOMAIN/g" nginx.conf
sed -i "s/your_password_here/$DB_PASSWORD/g" docker-compose.yml
sed -i "s/your_jwt_secret_here/$JWT_SECRET/g" docker-compose.yml
sed -i "s/your_password_here/$DB_PASSWORD/g" config/production.json
sed -i "s/your_jwt_secret_here/$JWT_SECRET/g" config/production.json
sed -i "s/your-domain.com/$DOMAIN/g" config/production.json

# 停止现有服务
echo "🛑 停止现有服务..."
docker-compose down || true

# 构建并启动服务
echo "🏗️ 构建并启动服务..."
docker-compose up -d --build

# 等待服务启动
echo "⏰ 等待服务启动..."
sleep 30

# 检查服务状态
echo "🔍 检查服务状态..."
docker-compose ps

# 运行数据库迁移和初始化
echo "💾 初始化数据库..."
docker-compose exec -T go-saber ./main migrate || echo "Migration skipped or failed"

# 设置 SSL 证书 (Let's Encrypt)
if [ "$ENVIRONMENT" = "production" ]; then
    echo "🔒 设置 SSL 证书..."
    
    # 安装 certbot (如果未安装)
    if ! command -v certbot &> /dev/null; then
        echo "Installing certbot..."
        if command -v apt-get &> /dev/null; then
            sudo apt-get update
            sudo apt-get install -y certbot
        elif command -v yum &> /dev/null; then
            sudo yum install -y certbot
        else
            echo "请手动安装 certbot"
        fi
    fi
    
    # 获取 SSL 证书
    sudo certbot certonly --webroot \
        -w ./ssl \
        -d $DOMAIN \
        -d www.$DOMAIN \
        --email $EMAIL \
        --agree-tos \
        --no-eff-email || echo "SSL setup failed, using self-signed certificates"
    
    # 如果 Let's Encrypt 失败，创建自签名证书
    if [ ! -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ]; then
        echo "创建自签名证书..."
        mkdir -p ssl
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout ssl/privkey.pem \
            -out ssl/fullchain.pem \
            -subj "/C=US/ST=State/L=City/O=Organization/CN=$DOMAIN"
    else
        # 复制 Let's Encrypt 证书
        sudo cp "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ssl/
        sudo cp "/etc/letsencrypt/live/$DOMAIN/privkey.pem" ssl/
        sudo chown $(whoami):$(whoami) ssl/*.pem
    fi
    
    # 重启 Nginx
    docker-compose restart nginx
fi

# 设置定时任务进行自动备份
echo "📅 设置自动备份..."
cat > backup.sh << 'EOF'
#!/bin/bash
# 数据库备份脚本
DATE=$(date +%Y%m%d_%H%M%S)
docker-compose exec -T postgres pg_dump -U postgres go_saber > "backup/db_backup_$DATE.sql"
# 保留最近30天的备份
find backup/ -name "db_backup_*.sql" -mtime +30 -delete
EOF

chmod +x backup.sh
mkdir -p backup

# 添加到 crontab (每天凌晨2点备份)
(crontab -l 2>/dev/null || true; echo "0 2 * * * cd $(pwd) && ./backup.sh") | crontab -

# 显示部署结果
echo ""
echo "🎉 部署完成！"
echo ""
echo "📊 服务状态:"
docker-compose ps
echo ""
echo "🌐 访问地址:"
if [ "$ENVIRONMENT" = "production" ]; then
    echo "  - HTTPS: https://$DOMAIN"
    echo "  - HTTP: http://$DOMAIN (将重定向到 HTTPS)"
else
    echo "  - HTTP: http://localhost"
fi
echo ""
echo "📋 管理命令:"
echo "  - 查看日志: docker-compose logs -f"
echo "  - 重启服务: docker-compose restart"
echo "  - 停止服务: docker-compose down"
echo "  - 更新应用: ./deploy.sh $ENVIRONMENT"
echo ""
echo "📂 重要文件:"
echo "  - 应用日志: ./logs/"
echo "  - 数据库备份: ./backup/"
echo "  - SSL 证书: ./ssl/"
echo ""

# 检查健康状态
echo "🔍 最终健康检查..."
sleep 10
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ 应用健康检查通过"
else
    echo "❌ 应用健康检查失败，请检查日志"
    docker-compose logs go-saber
fi

echo "✨ 部署完成！请访问您的域名查看效果。" 