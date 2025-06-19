# 构建阶段
FROM golang:1.22 AS builder

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

# 复制项目依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制全部源代码
COPY . .

# 构建 go-zero 可执行文件（假设入口为 main.go）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app dronestats.go

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 部署时可注释掉(仅调试用)
RUN mkdir -p /app/log && chmod -R 777 /app/log

# 拷贝配置文件和二进制文件
COPY --from=builder /app/app .
COPY --from=builder /app/etc ./etc

# 设置国内 apk 源
RUN echo "https://mirrors.aliyun.com/alpine/v3.22/main" > /etc/apk/repositories && \
    echo "https://mirrors.aliyun.com/alpine/v3.22/community" >> /etc/apk/repositories

RUN apk add --no-cache tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

# 替换为实际的端口
EXPOSE 8088

CMD ["sh", "-c", "./app -f etc/dronestats.yaml > /app/log/app.log 2>&1"]