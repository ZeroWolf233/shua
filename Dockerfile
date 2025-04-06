# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o http-loop .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/shua .
COPY --from=builder /app/templates ./templates

RUN apk --no-cache add ca-certificates

ENTRYPOINT ["./shua"]