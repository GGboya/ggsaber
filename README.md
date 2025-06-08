# Go Saber - 在线编程对战系统

一个功能完整的在线编程对战平台，支持实时匹配、代码提交、自动判题等功能。

## 🌟 功能特色

- **实时对战匹配**: 基于用户积分的智能匹配系统
- **多语言支持**: 支持 Go、Python、C++、Java 等主流编程语言
- **核心模式判题**: 类似 LeetCode 的函数实现模式
- **WebSocket 实时通信**: 实时对战状态更新
- **积分排行榜**: ELO 积分系统和排行榜
- **代码编辑器**: 简洁流畅的代码编辑界面

## 🚀 快速部署

### 服务器要求

- **操作系统**: Ubuntu 20.04+ / CentOS 7+ / Debian 10+
- **内存**: 最低 2GB，推荐 4GB+
- **存储**: 最低 20GB 可用空间
- **CPU**: 最低 2 核，推荐 4 核+
- **网络**: 公网 IP 和域名（可选）

### 1. 环境准备

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装 Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 重新登录以应用 Docker 组权限
```

### 2. 下载项目

```bash
# 下载项目代码
git clone <your-repository-url> go-saber-system
cd go-saber-system

# 或者直接上传项目文件到服务器
```

### 3. 配置部署

```bash
# 修改部署脚本中的域名和邮箱
vim deploy.sh

# 修改以下变量:
DOMAIN="your-domain.com"        # 替换为您的域名
EMAIL="your-email@example.com"  # 替换为您的邮箱
```

### 4. 执行部署

```bash
# 给脚本执行权限
chmod +x deploy.sh

# 生产环境部署
./deploy.sh production

# 或者测试环境部署
./deploy.sh staging
```

### 5. 域名解析

如果您有域名，请将域名解析到服务器 IP：

```
A记录: your-domain.com -> 您的服务器IP
A记录: www.your-domain.com -> 您的服务器IP
```

## 🔧 配置说明

### 环境变量配置

部署脚本会自动创建 `.env` 文件，您可以根据需要修改：

```bash
# 数据库密码
DB_PASSWORD=自动生成的随机密码

# JWT 密钥
JWT_SECRET=自动生成的随机密钥

# 域名配置
DOMAIN=your-domain.com
EMAIL=your-email@example.com

# 环境
ENVIRONMENT=production
```

### 应用配置

主要配置文件位于 `config/production.json`：

```json
{
  "server": {
    "port": ":8080",
    "mode": "release"
  },
  "database": {
    "host": "postgres",
    "port": 5432,
    "user": "postgres",
    "password": "your_password_here",
    "dbname": "go_saber"
  },
  "redis": {
    "addr": "redis:6379",
    "password": "",
    "db": 0
  }
  // ... 其他配置
}
```

## 📊 服务管理

### 查看服务状态

```bash
# 查看所有服务状态
docker-compose ps

# 查看实时日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f go-saber
```

### 服务操作

```bash
# 重启所有服务
docker-compose restart

# 重启特定服务
docker-compose restart go-saber

# 停止所有服务
docker-compose down

# 更新应用
./deploy.sh production
```

### 数据库管理

```bash
# 进入数据库容器
docker-compose exec postgres psql -U postgres -d go_saber

# 手动备份数据库
./backup.sh

# 恢复数据库
docker-compose exec -T postgres psql -U postgres -d go_saber < backup/db_backup_YYYYMMDD_HHMMSS.sql
```

## 🔒 安全配置

### SSL 证书

部署脚本会自动配置 Let's Encrypt SSL 证书。如果失败，会使用自签名证书。

手动更新证书：

```bash
# 更新 Let's Encrypt 证书
sudo certbot renew

# 重启 Nginx
docker-compose restart nginx
```

### 防火墙配置

```bash
# Ubuntu/Debian
sudo ufw allow 22    # SSH
sudo ufw allow 80    # HTTP
sudo ufw allow 443   # HTTPS
sudo ufw enable

# CentOS/RHEL
sudo firewall-cmd --permanent --add-port=22/tcp
sudo firewall-cmd --permanent --add-port=80/tcp
sudo firewall-cmd --permanent --add-port=443/tcp
sudo firewall-cmd --reload
```

## 📈 监控和维护

### 系统监控

```bash
# 查看系统资源使用情况
docker stats

# 查看磁盘使用情况
df -h

# 查看内存使用情况
free -h
```

### 日志管理

```bash
# 应用日志
tail -f logs/app.log

# Nginx 日志
docker-compose exec nginx tail -f /var/log/nginx/access.log

# 清理旧日志
find logs/ -name "*.log" -mtime +7 -delete
```

### 数据备份

系统会自动进行数据库备份（每天凌晨2点），备份文件保存在 `backup/` 目录。

手动备份：

```bash
# 数据库备份
./backup.sh

# 完整系统备份
tar -czf go-saber-backup-$(date +%Y%m%d).tar.gz \
  --exclude='logs' \
  --exclude='backup' \
  .
```

## 🛠 故障排除

### 常见问题

1. **服务无法启动**
   ```bash
   # 查看详细错误日志
   docker-compose logs go-saber
   
   # 检查端口占用
   sudo netstat -tlnp | grep :8080
   ```

2. **数据库连接失败**
   ```bash
   # 检查数据库状态
   docker-compose exec postgres pg_isready -U postgres
   
   # 重启数据库
   docker-compose restart postgres
   ```

3. **SSL 证书问题**
   ```bash
   # 手动生成自签名证书
   mkdir -p ssl
   openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
     -keyout ssl/privkey.pem -out ssl/fullchain.pem
   ```

### 性能优化

1. **数据库优化**
   - 定期清理过期数据
   - 添加适当的索引
   - 调整 PostgreSQL 配置

2. **应用优化**
   - 调整 Worker 数量
   - 配置连接池
   - 启用缓存

3. **系统优化**
   - 增加交换空间
   - 调整文件描述符限制
   - 配置系统监控

## 📞 技术支持

如果您在部署过程中遇到问题，可以：

1. 查看应用日志：`docker-compose logs go-saber`
2. 检查系统资源：`docker stats`
3. 查看网络连接：`docker-compose ps`

## 📝 更新日志

- **v1.0.0** - 初始版本发布
  - 基础对战功能
  - 核心模式判题
  - 用户系统和积分

---

**部署完成后，您的在线编程对战系统就可以正式运行了！** 🎉 