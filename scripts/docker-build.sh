#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building Obsidian Search MCP Docker image...${NC}"

# Build the Docker image
docker build -t obsidian-search-mcp:latest .

echo -e "${GREEN}Build complete!${NC}"
echo -e "${YELLOW}To run the container:${NC}"
echo "docker run -it --rm \\"
echo "  -e OBSIDIAN_VAULT_PATH=/vault \\"
echo "  -v /path/to/your/vault:/vault:ro \\"
echo "  -v obsidian-index:/data \\"
echo "  obsidian-search-mcp:latest"