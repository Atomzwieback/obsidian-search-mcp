# obsidian-mcp-search

A Model Context Protocol (MCP) server that provides fast full-text search capabilities for Obsidian vaults using Tantivy-Go.

## Features

- 🔍 **Fast Full-Text Search**: Powered by Tantivy, a high-performance search engine
- 📁 **Incremental Indexing**: Only re-indexes modified files
- 🔄 **Real-time Updates**: Automatically updates the index when files change
- 🚀 **Concurrent Processing**: Uses multiple CPU cores for fast indexing
- 📍 **Context Snippets**: Shows surrounding context for search matches
- 🔗 **MCP Integration**: Works seamlessly with Claude Desktop

## Installation

### Prerequisites

- Go 1.20 or higher
- Obsidian vault with markdown files

### Build from Source

```bash
git clone https://github.com/Atomzwieback/obsidian-mcp-search.git
cd obsidian-mcp-search
go build -o obsidian-mcp-search cmd/server/main.go
```

### Install with Go

```bash
go install github.com/Atomzwieback/obsidian-mcp-search/cmd/server@latest
```

## Configuration

### Claude Desktop Setup

Add the following to your Claude Desktop `config.json`:

```json
{
  "mcpServers": {
    "obsidian-search": {
      "command": "/path/to/obsidian-mcp-search",
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

- "Search for 'meeting notes' in my Obsidian notes"
- "Find all files containing 'project timeline'"
- "Show me notes about 'golang performance'"
- "Reindex my Obsidian vault"

### Available Tools

1. **search_obsidian**: Search for content in Obsidian markdown files
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
