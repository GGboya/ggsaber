.PHONY: build run test clean deps docker-build docker-run

# 默认目标
all: deps build

# 安装依赖
deps:
	go mod tidy
	go mod download

# 构建项目
build:
	go build -o bin/saber-server cmd/server/main.go

# 运行项目
run:
	go run cmd/server/main.go

# 运行测试
test:
	go test -v ./...

# 清理构建文件
clean:
	rm -rf bin/
	rm -rf tmp/

# 代码格式化
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run

# 初始化数据库
init-db:
	mysql -u root -p < scripts/init_data.sql

# Docker构建
docker-build:
	docker build -t saber-system .

# Docker运行
docker-run:
	docker-compose up -d

# 开发环境启动
dev: deps
	go run cmd/server/main.go

# 生产环境构建
prod: deps
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/saber-server cmd/server/main.go

# 帮助信息
help:
	@echo "Available commands:"
	@echo "  deps      - Install dependencies"
	@echo "  build     - Build the application"
	@echo "  run       - Run the application in development mode"
	@echo "  test      - Run tests"
	@echo "  clean     - Clean build files"
	@echo "  fmt       - Format code"
	@echo "  lint      - Run linter"
	@echo "  init-db   - Initialize database"
	@echo "  dev       - Start development server"
	@echo "  prod      - Build for production"
	@echo "  help      - Show this help message" 