# Build Stage
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Copy dependency files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build static binary
COPY . .
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -ldflags="-s -w" -o main .

# Final Stage
FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/main .

# Default port (matches common defaults, though your app reads from env)
EXPOSE 8080

CMD ["./main"]