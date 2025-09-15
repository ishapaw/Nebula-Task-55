# ---------- Build Stage ----------
FROM golang:1.24-bullseye AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build query service binary
RUN go build -o bookings-view-service

# ---------- Runtime Stage ----------
FROM ubuntu:22.04
WORKDIR /root/
COPY --from=builder /app/bookings-view-service .
EXPOSE 8081
CMD ["./bookings-view-service"]
