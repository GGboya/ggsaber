# 多阶段构建
# 第一阶段：构建
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git gcc g++ musl-dev

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# 第二阶段：运行
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates python3 g++ gcc musl-dev

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/web ./web/
COPY --from=builder /app/config ./config/

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./main"] 