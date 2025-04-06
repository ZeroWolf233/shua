# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o http-loop .

# 最终镜像
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/http-loop .
COPY --from=builder /app/templates ./templates  # 如果有模板文件需要添加

# 安装CA证书（用于HTTPS请求）
RUN apk --no-cache add ca-certificates

ENTRYPOINT ["./http-loop"]