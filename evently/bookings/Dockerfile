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

# Install goreman (Procfile manager)
RUN go install github.com/mattn/goreman@latest


# ---- Runtime stage ----
FROM alpine:latest
WORKDIR /app

# Copy binaries
COPY --from=builder /bin/cancel_consumer .
COPY --from=builder /bin/bookings_consumer .
COPY --from=builder /bin/update_seats_consumer .

# Copy goreman binary
COPY --from=builder /go/bin/goreman /usr/local/bin/goreman

# Copy Procfile
COPY Procfile .

# Run all consumers with goreman
CMD ["goreman", "start", "-f", "Procfile"]
