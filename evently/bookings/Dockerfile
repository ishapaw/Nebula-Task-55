# ---- Build stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

# cancel_consumer
WORKDIR /app/cancel_consumer
COPY cancel_consumer/go.mod cancel_consumer/go.sum ./
RUN go mod download
COPY cancel_consumer/ .
RUN go build -o /bin/cancel_consumer

# bookings_consumer
WORKDIR /app/bookings_consumer
COPY bookings_consumer/go.mod bookings_consumer/go.sum ./
RUN go mod download
COPY bookings_consumer/ .
RUN go build -o /bin/bookings_consumer

# update_seats_consumer
WORKDIR /app/update_seats_consumer
COPY update_seats_consumer/go.mod update_seats_consumer/go.sum ./
RUN go mod download
COPY update_seats_consumer/ .
RUN go build -o /bin/update_seats_consumer

# ---- Runtime ----
FROM alpine:latest
WORKDIR /app

COPY --from=builder /bin/cancel_consumer .
COPY --from=builder /bin/bookings_consumer .
COPY --from=builder /bin/update_seats_consumer .

# Add Procfile
COPY Procfile .

# Install goreman (Procfile runner)
RUN apk add --no-cache curl bash && \
    curl -sL https://github.com/mattn/goreman/releases/download/v0.3.11/goreman_linux_amd64.tar.gz \
    | tar -xz -C /usr/local/bin && \
    chmod +x /usr/local/bin/goreman

# Start with Procfile
CMD ["goreman", "start"]
