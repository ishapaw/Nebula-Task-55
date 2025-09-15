# 1️⃣ Build stage
FROM golang:1.24-bullseye AS builder

RUN apt-get update && apt-get install -y curl

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o gateway .

# 2️⃣ Runtime stage
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/gateway .

EXPOSE 8080
CMD ["./gateway"]


