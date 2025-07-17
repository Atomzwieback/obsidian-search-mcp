# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc g++ musl-dev git curl tar

# Set tantivy-go version
ENV TANTIVY_VERSION=1.0.4

# Download pre-built tantivy libraries (musl version for Alpine)
WORKDIR /tantivy-libs
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then \
        TANTIVY_ARCH="amd64"; \
    elif [ "$ARCH" = "aarch64" ]; then \
        TANTIVY_ARCH="arm64"; \
    else \
        echo "Unsupported architecture: $ARCH" && exit 1; \
    fi && \
    echo "Downloading tantivy-go v${TANTIVY_VERSION} for linux-${TANTIVY_ARCH}-musl" && \
    curl -L -o tantivy.tar.gz https://github.com/anyproto/tantivy-go/releases/download/v${TANTIVY_VERSION}/linux-${TANTIVY_ARCH}-musl.tar.gz && \
    tar -xzf tantivy.tar.gz && \
    mkdir -p /usr/local/lib && \
    cp libtantivy_go.a /usr/local/lib/

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV CGO_LDFLAGS="-L/usr/local/lib"

# Build the application
RUN go build -a -o obsidian-search-mcp ./cmd/server

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates libgcc libstdc++

# Copy the tantivy library from builder
COPY --from=builder /usr/local/lib/libtantivy_go.a /usr/local/lib/

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

# Environment variables
ENV MCP_INDEX_PATH=/home/obsidian/.obsidian-mcp/index

# Run the server
ENTRYPOINT ["obsidian-search-mcp"]