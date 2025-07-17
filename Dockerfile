# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc g++ musl-dev git

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

RUN go build -a -installsuffix cgo -o obsidian-search-mcp ./cmd/server

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates libgcc libstdc++

# Create non-root user
RUN addgroup -g 1000 -S obsidian && \
    adduser -u 1000 -S obsidian -G obsidian

# Create directories
RUN mkdir -p /home/obsidian/.obsidian-mcp/index && \
    chown -R obsidian:obsidian /home/obsidian

WORKDIR /home/obsidian

# Copy binary from builder
COPY --from=builder /build/obsidian-search-mcp /usr/local/bin/obsidian-search-mcp

# Switch to non-root user
USER obsidian

# Environment variables (can be overridden)
ENV MCP_INDEX_PATH=/home/obsidian/.obsidian-mcp/index

# Run the server
ENTRYPOINT ["obsidian-search-mcp"]