# Build stage for tantivy-go library
FROM rust:1.75-alpine AS rust-builder

# Install build dependencies
RUN apk add --no-cache musl-dev gcc g++ make git

# Clone and build tantivy-go
WORKDIR /build
RUN git clone https://github.com/anyproto/tantivy-go.git
WORKDIR /build/tantivy-go/rust

# Build the C library
RUN cargo build --release
RUN cp target/release/libtantivy_go.a /usr/local/lib/
RUN cp target/release/libtantivy_go.so /usr/local/lib/ || true

# Go build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc g++ musl-dev git

# Copy the tantivy library from rust builder
COPY --from=rust-builder /usr/local/lib/libtantivy_go.* /usr/local/lib/

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
ENV CGO_LDFLAGS="-L/usr/local/lib"

RUN go build -a -installsuffix cgo -o obsidian-search-mcp ./cmd/server

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates libgcc libstdc++

# Copy the tantivy library
COPY --from=rust-builder /usr/local/lib/libtantivy_go.* /usr/local/lib/

# Create non-root user
RUN addgroup -g 1000 -S obsidian && \
    adduser -u 1000 -S obsidian -G obsidian

# Create directories
RUN mkdir -p /home/obsidian/.obsidian-mcp/index && \
    chown -R obsidian:obsidian /home/obsidian

WORKDIR /home/obsidian

# Copy binary from builder
COPY --from=builder /build/obsidian-search-mcp /usr/local/bin/obsidian-search-mcp

# Update library cache
RUN ldconfig /usr/local/lib || true

# Switch to non-root user
USER obsidian

# Environment variables (can be overridden)
ENV MCP_INDEX_PATH=/home/obsidian/.obsidian-mcp/index
ENV LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH

# Run the server
ENTRYPOINT ["obsidian-search-mcp"]