# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Lint
```bash
make lint
# or
go vet ./...
```

### Build & Run Example
```bash
go build -o out/redditanalyzer ./cmd/redditanalyzer
./out/redditanalyzer
```

### Go Module Management
```bash
go mod init
go get <package>
go mod tidy
```

## Project Structure

This is a Go AI Agent library project structure:
- `go.mod` - Go module definition
- `Makefile` - Development commands (lint only)
- `cmd/redditanalyzer/` - Example AI Agent implementation
- `out/` - Build output directory

## Architecture

This is a Go library for developing AI Agents, designed to simplify AI Agent development. The project follows standard Go conventions with example applications in `cmd/` and build output in `out/`. The `redditanalyzer` serves as an example implementation using the AI Agent library.

## MCP Integration

The project is configured to work with MCP (Model Context Protocol) tools:
- GitHub integration for repository operations
- Reddit integration for fetching hot threads
- Configured permissions allow Go build operations and Reddit analyzer execution

## Development Notes

- Go version: 1.24.4
- Primary focus: AI Agent library development
- Example implementation: `redditanalyzer` executable
- Early-stage project with foundational structure in place