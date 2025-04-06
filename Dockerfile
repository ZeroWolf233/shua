FROM golang:1.23.4 AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build
