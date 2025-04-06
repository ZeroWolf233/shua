FROM golang:1.20 AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build
