# Stage 1: Build
FROM golang:1.24-alpine AS builder

# Set environment
ENV CGO_ENABLED=0
WORKDIR /app

# Copy go.mod and go.sum for caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the Go binary
RUN go build -o combined-service main.go

# Stage 2: Minimal runtime
FROM alpine:latest
WORKDIR /app

# Copy the built binary
COPY --from=builder /app/combined-service .

# Expose port (adjust if your service runs on a different port)
EXPOSE 8080

# Command to run
CMD ["./combined-service"]
