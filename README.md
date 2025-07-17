<div align="center">
  <img src="logo.png" alt="Obsidian Search MCP Logo" width="200">
  
  # obsidian-search-mcp

  A Model Context Protocol (MCP) server that provides fast full-text search capabilities for Obsidian vaults using Tantivy-Go.
</div>

## Features

- 🔍 **Fast Full-Text Search**: Powered by Tantivy, a high-performance search engine
- 📁 **Incremental Indexing**: Only re-indexes modified files
- 🔄 **Real-time Updates**: Automatically updates the index when files change
- 🚀 **Concurrent Processing**: Uses multiple CPU cores for fast indexing
- 📍 **Context Snippets**: Shows surrounding context for search matches
- 🔗 **MCP Integration**: Works seamlessly with Claude Desktop

## Installation

### Prerequisites

- Obsidian vault with markdown files
- Either Go 1.20+ OR Docker

### Option 1: Docker (Recommended)

The easiest way to run the server is using Docker, which handles all CGO dependencies:

```bash
# Clone the repository
git clone https://github.com/Atomzwieback/obsidian-search-mcp.git
cd obsidian-search-mcp

# Build the Docker image
./scripts/docker-build.sh

# Or use docker-compose
docker-compose build
```

### Option 2: Build from Source

⚠️ **Note**: This requires CGO and Tantivy C libraries to be available.

```bash
git clone https://github.com/Atomzwieback/obsidian-search-mcp.git
cd obsidian-search-mcp
CGO_ENABLED=1 go build -o obsidian-search-mcp cmd/server/main.go
```

## Configuration

### Claude Desktop Setup

#### Docker Setup

For Docker deployment, you'll need to use a wrapper script:

```json
{
  "mcpServers": {
    "obsidian-search": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-e", "OBSIDIAN_VAULT_PATH=/vault",
        "-v", "/path/to/your/obsidian/vault:/vault:ro",
        "-v", "obsidian-mcp-index:/data",
        "obsidian-search-mcp:latest"
      ]
    }
  }
}
```

#### Native Binary Setup

If running the binary directly:

```json
{
  "mcpServers": {
    "obsidian-search": {
      "command": "/path/to/obsidian-search-mcp",
      "env": {
        "OBSIDIAN_VAULT_PATH": "/path/to/your/obsidian/vault",
        "MCP_INDEX_PATH": "/path/to/index/storage"
      }
    }
  }
}
```

### Environment Variables

- `OBSIDIAN_VAULT_PATH` (required): Path to your Obsidian vault directory
- `MCP_INDEX_PATH` (optional): Path to store the search index (defaults to `~/.obsidian-mcp/index`)

## Usage

Once configured, you can use the following commands in Claude:

- "Search for 'meeting notes' in my Obsidian vault"
- "Find all files containing 'project timeline'"
- "Show me notes about 'golang performance'"
- "Reindex my Obsidian vault"

### Available Tools

1. **search_vault**: Search for content in Obsidian vault markdown files
   - Parameters:
     - `query` (required): Search query text
     - `limit` (optional): Maximum number of results (default: 10)

2. **reindex_vault**: Force reindex of the entire Obsidian vault

### Resources

- **index_status**: Shows current index status and statistics

## Architecture

The server is built with:
- **Tantivy-Go**: High-performance full-text search engine
- **MCP-Go**: Model Context Protocol server implementation
- **FSNotify**: File system monitoring for real-time updates
- **Godirwalk**: Fast directory traversal

## Performance

- Incremental indexing ensures only changed files are processed
- Concurrent workers utilize multiple CPU cores
- Memory-mapped I/O for efficient file handling
- Debounced file watching prevents excessive updates

## License

MIT License - see LICENSE file for details
