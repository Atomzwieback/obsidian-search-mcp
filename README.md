<div align="center">
  <img src="logo.png" alt="Obsidian Search MCP Logo" width="400">

  [![Go Tests](https://github.com/Atomzwieback/obsidian-search-mcp/actions/workflows/go.yml/badge.svg)](https://github.com/Atomzwieback/obsidian-search-mcp/actions/workflows/go.yml)
  [![Docker](https://github.com/Atomzwieback/obsidian-search-mcp/actions/workflows/docker.yml/badge.svg)](https://github.com/Atomzwieback/obsidian-search-mcp/actions/workflows/docker.yml)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

  **Blazing fast search for your Obsidian knowledge base, powered by [Tantivy](https://github.com/quickwit-oss/tantivy) and accessible through any MCP client**
</div>

## Why Obsidian Search MCP?

Your Obsidian vault grows organically over years - scattered thoughts, meeting notes, project docs, research snippets. Eventually, even you can't remember where everything is. You need help organizing and connecting all this knowledge.

**But for an AI assistant to help organize your vault, it needs to see what's already there.**

This MCP server enables that by giving AI assistants instant search access to your entire vault. Now you can have conversations like:

- "Help me consolidate all my notes about the PYCK project"
- "Find patterns in my meeting notes from last quarter"
- "Which notes reference this API design but aren't properly linked?"
- "Create a MOC (Map of Content) for my Temporal workflow research"

The AI can quickly search through thousands of files, understand your knowledge structure, and help you build a better organized, more connected vault.

## Features

- 🔍 **Fast Full-Text Search**: Powered by [Tantivy](https://github.com/quickwit-oss/tantivy), a high-performance search engine written in Rust
- 📁 **Incremental Indexing**: Only re-indexes modified files
- 🔄 **Real-time Updates**: Automatically updates the index when files change
- 🚀 **Concurrent Processing**: Uses multiple CPU cores for fast indexing
- 📍 **Context Snippets**: Shows surrounding context for search matches
- 🔗 **MCP Integration**: Works seamlessly with Claude Code, Claude Desktop, and other MCP clients

## Installation

### Prerequisites

- Obsidian vault with markdown files
- Either Go 1.20+ OR Docker

### Option 1: Docker (Recommended)

[![Docker Hub](https://img.shields.io/docker/v/atomzwieback/obsidian-search-mcp?label=Docker%20Hub)](https://hub.docker.com/r/atomzwieback/obsidian-search-mcp)
[![Docker Pulls](https://img.shields.io/docker/pulls/atomzwieback/obsidian-search-mcp)](https://hub.docker.com/r/atomzwieback/obsidian-search-mcp)

The pre-built image will be automatically pulled from Docker Hub when you first run the MCP server.

To build from source instead:

```bash
# Clone the repository
git clone https://github.com/Atomzwieback/obsidian-search-mcp.git
cd obsidian-search-mcp

# Build the Docker image
docker build -t obsidian-search-mcp .
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
        "atomzwieback/obsidian-search-mcp:latest"
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
- **[Tantivy-Go](https://github.com/anyproto/tantivy-go)**: Go bindings for the Tantivy search engine
- **[Tantivy](https://github.com/quickwit-oss/tantivy)**: Lightning-fast full-text search engine written in Rust
- **[MCP-Go](https://github.com/mark3labs/mcp-go)**: Model Context Protocol server implementation
- **[FSNotify](https://github.com/fsnotify/fsnotify)**: File system monitoring for real-time updates
- **[Godirwalk](https://github.com/karrick/godirwalk)**: Fast directory traversal

## Performance

- Incremental indexing ensures only changed files are processed
- Concurrent workers utilize multiple CPU cores
- Memory-mapped I/O for efficient file handling
- Debounced file watching prevents excessive updates

## License

MIT License - see LICENSE file for details
