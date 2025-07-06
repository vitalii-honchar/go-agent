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

### Test
```bash
make test
# or
go test ./...
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
- `Makefile` - Development commands (lint, test)
- `pkg/goagent/` - Core library packages
  - `agent/` - Agent implementation and interfaces
  - `llm/` - LLM integrations and abstractions
  - `config/` - Configuration management
- `cmd/redditanalyzer/` - Example AI Agent implementation
- `out/` - Build output directory

## Architecture

This is a Go library for developing AI Agents, designed to simplify AI Agent development. The architecture follows clean code principles and Go best practices:

### Core Components
- **Agent**: Generic agent with configurable behavior, tools, and output schemas
- **LLM**: Abstraction layer for different LLM providers (OpenAI, etc.)
- **Config**: Environment-based configuration management
- **Tools**: Extensible tool system for agent capabilities

### Design Principles
- **Interface Segregation**: Small, focused interfaces (LLM, LLMTool)
- **Dependency Injection**: Options pattern for flexible configuration
- **Generic Types**: Type-safe agents with custom output schemas
- **Error Handling**: Explicit error types and proper error wrapping
- **Single Responsibility**: Each package has a clear, single purpose

## MCP Integration

The project is configured to work with MCP (Model Context Protocol) tools:
- GitHub integration for repository operations
- Reddit integration for fetching hot threads
- Configured permissions allow Go build operations and Reddit analyzer execution

## Clean Code Principles

### Go-Specific Guidelines
- **Naming**: Use clear, descriptive names (Agent, LLMConfig, NewAgent)
- **Package Structure**: Organize by domain, not by layer
- **Interfaces**: Define at point of use, keep them small
- **Error Handling**: Use typed errors, wrap with context
- **Testing**: Write integration tests for public APIs

### Architecture Patterns
- **Options Pattern**: Flexible configuration (WithName, WithBehavior)
- **Factory Pattern**: LLM creation with type safety
- **Strategy Pattern**: Different LLM implementations
- **Builder Pattern**: Agent construction with validation

### Code Organization
- **pkg/**: Library code, no main packages
- **cmd/**: Example applications and tools
- **Avoid**: Deep nesting, god objects, global state
- **Prefer**: Composition over inheritance, explicit dependencies

## Development Notes

- Go version: 1.24.4
- Primary focus: AI Agent library development
- Example implementation: `redditanalyzer` executable
- Testing: Integration tests with real LLM calls
- Dependencies: Minimal, well-maintained packages only