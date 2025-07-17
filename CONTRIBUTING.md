# Contributing to Obsidian Search MCP

First off, thank you for considering contributing to Obsidian Search MCP! 🎉

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples**
- **Include your environment details** (OS, Go version, Docker version)
- **Include logs if applicable**

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion:

- **Use a clear and descriptive title**
- **Provide a detailed description of the suggested enhancement**
- **Explain why this enhancement would be useful**
- **List any alternative solutions you've considered**

### Pull Requests

1. Fork the repo and create your branch from `main`
2. If you've added code that should be tested, add tests
3. Ensure the test suite passes
4. Make sure your code follows the existing style
5. Issue that pull request!

## Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/obsidian-search-mcp
cd obsidian-search-mcp

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
docker build -t obsidian-search-mcp:dev .
```

## Style Guidelines

### Go Code Style

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable names
- Add comments for exported functions
- Keep functions small and focused

### Commit Messages

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally after the first line

Example:
```
feat: Add incremental indexing support

- Only re-index modified files
- Store timestamps in .timestamps file
- Improve indexing performance by 80%

Fixes #123
```

### Documentation

- Use Markdown for all documentation
- Keep README.md up to date
- Document new features with examples
- Update the changelog

## Questions?

Feel free to contact the maintainer at contact at atomzwieback.dev

Thank you for contributing! 🚀