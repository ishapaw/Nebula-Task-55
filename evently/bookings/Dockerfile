# ---- Build stage ----
FROM golang:1.22 AS builder

WORKDIR /app

# Copy everything
COPY . .

# Build each consumer binary
WORKDIR /app/bookings/cancel_consumer
RUN go build -o /bin/cancel_consumer

WORKDIR /app/bookings/bookings_consumer
RUN go build -o /bin/bookings_consumer

WORKDIR /app/bookings/update_seats_consumer
RUN go build -o /bin/update_seats_consumer


# ---- Runtime stage ----
FROM debian:bookworm-slim

WORKDIR /app

# Copy built binaries
COPY --from=builder /bin/cancel_consumer /app/
COPY --from=builder /bin/bookings_consumer /app/
COPY --from=builder /bin/update_seats_consumer /app/

# Install forego (process manager)
RUN apt-get update && apt-get install -y --no-install-recommends \
    wget unzip ca-certificates && \
    wget -q https://github.com/ddollar/forego/releases/download/v0.17.2/forego-linux-amd64.tgz && \
    tar -xvzf forego-linux-amd64.tgz && \
    mv forego /usr/local/bin/ && \
    rm forego-linux-amd64.tgz && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Add Procfile
COPY Procfile /app/Procfile

# Start all consumers with forego
CMD ["forego", "start", "-r"]
