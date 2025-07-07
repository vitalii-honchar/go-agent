# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Lint
```bash
make lint
# or
go vet ./...
```

### Build & Test
```bash
go build ./pkg/goagent/...
# For tests, first export .env variables:
export $$(cat .env | xargs) && go test ./pkg/goagent/... -v
```

### Test
```bash
make test
# or
# First export environment variables from .env file:
export $$(cat .env | xargs) && go test ./...
# Note: Tests require OPENAI_API_KEY environment variable
```

### Go Module Management
```bash
go mod init
go get <package>
go mod tidy
```

## Project Structure

This is a Go AI Agent library project structure:
- `go.mod` - Go module definition and dependencies
- `go.sum` - Go module checksums
- `Makefile` - Development commands (lint, test)
- `.env` - Environment variables (not committed, contains API keys)
- `pkg/goagent/` - Core library packages
  - `agent/` - Agent implementation, interfaces, and tests
    - `agent.go` - Main agent implementation
    - `agent_result.go` - Agent result types
    - `agent_test.go` - Integration tests
    - `prompt.go` - System prompt management
  - `llm/` - LLM integrations and abstractions
    - `llm.go` - LLM interfaces and types
    - `llm_config.go` - LLM configuration
    - `llm_message.go` - Message types
    - `llm_tool.go` - Tool definitions
    - `openai.go` - OpenAI implementation
  - `config/` - Configuration management
    - `config.go` - Environment-based config
- `.github/workflows/` - CI/CD pipeline
  - `ci.yml` - GitHub Actions workflow
- `.golangci.yml` - Linter configuration
- `README.md` - Project documentation
- `CLAUDE.md` - Claude Code instructions

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

## Current Status

The project is in active development with the following features implemented:
- ✅ Core Agent functionality with generic types
- ✅ OpenAI LLM integration with configurable models
- ✅ Tool system with usage limits and validation
- ✅ Comprehensive integration test suite
- ✅ CI/CD pipeline with GitHub Actions
- ✅ Configuration management with environment variables
- ✅ JSON schema validation for structured outputs
- ✅ System prompt templating and customization
- ✅ Type-safe agent construction with options pattern
- ✅ Error handling with proper context and wrapping

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
- **tests/**: Integration tests with real LLM calls
- **Avoid**: Deep nesting, god objects, global state
- **Prefer**: Composition over inheritance, explicit dependencies

## Development Notes

- Go version: 1.24.4
- Primary focus: AI Agent library development
- Repository: https://github.com/vitalii-honchar/go-agent
- Testing: Integration tests with real LLM calls (requires OPENAI_API_KEY in .env)
- Dependencies: Key packages include:
  - `github.com/openai/openai-go` - OpenAI API client
  - `github.com/joho/godotenv` - Environment variable loading
  - `github.com/invopop/jsonschema` - JSON schema generation
  - `github.com/xeipuuv/gojsonschema` - JSON schema validation
  - `github.com/stretchr/testify` - Testing utilities
- CI/CD: GitHub Actions with automated testing and linting

## Environment Setup

**Required for testing:**
1. Create `.env` file with required environment variables:
   ```
   OPENAI_API_KEY=your_openai_api_key_here
   ```
2. Export environment variables before running tests:
   ```bash
   export $$(cat .env | xargs)
   ```
3. Run tests:
   ```bash
   go test ./...
   ```

**Note:** Tests will fail without proper environment variables as they make real API calls to OpenAI.