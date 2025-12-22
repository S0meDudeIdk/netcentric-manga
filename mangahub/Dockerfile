# Multi-stage build for Go backend services
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc g++ musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files (go.sum will be generated if missing)
COPY go.mod ./
COPY go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build all services
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/tcp-server ./cmd/tcp-server
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/udp-server ./cmd/udp-server
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/grpc-server ./cmd/grpc-server
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/api-server ./cmd/api-server
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/fetch-manga-server ./cmd/fetch-manga-server

# Final stage - minimal runtime image
FROM alpine:latest

# Install runtime dependencies including healthcheck tools
RUN apk add --no-cache ca-certificates sqlite-libs curl netcat-openbsd

# Create app directory
WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/bin/* /app/

# Copy necessary files
COPY --from=builder /app/data /app/data
COPY --from=builder /app/proto /app/proto

# Create data directory if it doesn't exist
RUN mkdir -p /app/data

# Default command (will be overridden by docker-compose)
CMD ["/app/api-server"]
