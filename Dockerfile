FROM golang:1.23.4 AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download

RUN apk update && \
    apk add --no-cache ca-certificates && \
    update-ca-certificates

COPY . .
# 生成静态二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o shua .

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/shua /shua

ENTRYPOINT ["/shua"]