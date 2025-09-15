# Start from Go image
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first (for better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the consumer binary
RUN go build -o update-seats-consumer main.go

# Use minimal image
FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/update-seats-consumer .

# Run binary
CMD ["./update-seats-consumer"]

