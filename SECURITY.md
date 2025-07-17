# Security Policy

## Supported Versions

Currently supported versions for security updates:

| Version | Supported          |
| ------- | ------------------ |
| Latest  | :white_check_mark: |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability within Obsidian Search MCP, please follow these steps:

1. **DO NOT** open a public issue
2. Email details to: contact at atomzwieback.dev
3. Include the following information:
   - Type of vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### What to expect:

- **Response Time**: We aim to respond within 48 hours
- **Acknowledgment**: You'll receive confirmation that we received your report
- **Updates**: We'll keep you informed about the progress
- **Credit**: We'll credit you for the discovery (unless you prefer to remain anonymous)

## Security Best Practices

When using Obsidian Search MCP:

1. **File System Access**: The server has read-only access to your vault
2. **Docker Isolation**: Use Docker for additional security isolation
3. **Network**: The MCP server only communicates via stdio (no network access)
4. **Updates**: Keep the server updated to the latest version

## Known Security Considerations

- The server requires read access to your Obsidian vault
- Index files are stored locally and contain searchable content from your vault
- No data is sent to external servers
- All processing happens locally

## Dependencies

We regularly update dependencies to patch known vulnerabilities:
- Go dependencies: Run `go mod tidy` and check with `go list -m all`
- Docker base images: Updated in Dockerfile

For questions about security, contact: contact at atomzwieback.dev