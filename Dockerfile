# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc g++ musl-dev git make curl

# Install Rust (needed for tantivy-go setup)
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"

# Clone tantivy-go to get pre-built libraries
WORKDIR /tantivy-setup
RUN git clone https://github.com/anyproto/tantivy-go.git
WORKDIR /tantivy-setup/tantivy-go/rust

# Download pre-built libraries instead of building from source
RUN make download-tantivy-all

# Copy the downloaded libraries to system location
RUN mkdir -p /usr/local/lib && \
    cp lib/linux_amd64/*.a /usr/local/lib/ && \
    cp lib/linux_arm64/*.a /usr/local/lib/ || true

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled and correct library paths
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV CGO_LDFLAGS="-L/usr/local/lib"

# Build for the current architecture
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