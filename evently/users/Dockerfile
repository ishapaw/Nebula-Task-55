# 1️⃣ Build stage
FROM golang:1.24-bullseye AS builder

# Set working directory
WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build Go binary
RUN go build -o users-service .

# 2️⃣ Runtime stage
FROM ubuntu:22.04

# Install CA certificates for HTTPS
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/users-service .

# Expose service port
EXPOSE 8081

# Run service
CMD ["./users-service"]

