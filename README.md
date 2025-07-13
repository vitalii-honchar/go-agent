# Go Agent

> 📖 **Featured Blog Post**: [Building AI Agents in Go: A Comprehensive Guide](https://vitaliihonchar.com/insights/go-ai-agent-library) - Learn about the design principles and real-world applications of this library.

A powerful, production-ready Go library for building AI agents with configurable behavior, custom tools, and type-safe output schemas. Perfect for building intelligent automation, data analysis tools, web scrapers, and AI-powered applications.

[![CI](https://github.com/vitalii-honchar/go-agent/actions/workflows/ci.yml/badge.svg)](https://github.com/vitalii-honchar/go-agent/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/vitalii-honchar/go-agent)](https://goreportcard.com/report/github.com/vitalii-honchar/go-agent)
[![GoDoc](https://godoc.org/github.com/vitalii-honchar/go-agent?status.svg)](https://godoc.org/github.com/vitalii-honchar/go-agent)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 🌟 Why Choose Go Agent?

- **Production Ready** - Used in real-world applications with robust error handling
- **Developer Friendly** - Intuitive API design with comprehensive examples
- **Extensible** - Easy to add custom tools and integrate with external services
- **Type Safe** - Leverage Go's type system for reliable agent development
- **Well Tested** - Comprehensive test suite with high code coverage

## ✨ Features

- 🤖 **Generic AI Agents** - Type-safe agents with custom output schemas
- 🔧 **Extensible Tools** - Add custom tools with automatic limit enforcement
- 🎯 **Configurable Behavior** - Define agent behavior with natural language
- 🚀 **Multiple LLM Support** - Currently supports OpenAI (extensible to others)
- ⚡ **Tool Limits** - Prevent runaway execution with per-tool usage limits
- 📝 **Structured Output** - JSON schema validation for reliable results
- 🛡️ **Type Safety** - Full Go generics support for compile-time safety
- 🏗️ **Clean Architecture** - Interface-based design following Go best practices

## 🚀 Quick Start

### Installation

```bash
go get github.com/vitalii-honchar/go-agent
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "github.com/vitalii-honchar/go-agent/pkg/goagent/agent"
    "github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
)

// Define your output schema
type MathResult struct {
    Answer int    `json:"answer" jsonschema_description:"The calculated answer"`
    Steps  string `json:"steps" jsonschema_description:"The calculation steps taken"`
}

func main() {
    // Create an agent
    mathAgent, err := agent.NewAgent(
        agent.WithName[MathResult]("math-solver"),
        agent.WithLLMConfig[MathResult](llm.LLMConfig{
            Type:        llm.LLMTypeOpenAI,
            APIKey:      "your-openai-api-key",
            Model:       "gpt-4",
            Temperature: 0.0,
        }),
        agent.WithBehavior[MathResult]("You are a math solver. Calculate the given expression and show your work."),
    )
    if err != nil {
        panic(err)
    }

    // Run the agent
    result, err := mathAgent.Run(context.Background(), "What is 15 + 27?")
    if err != nil {
        panic(err)
    }

    fmt.Printf("Answer: %d\n", result.Data.Answer)
    fmt.Printf("Steps: %s\n", result.Data.Steps)
}
```

## 📚 Real-World Examples

Explore our comprehensive examples in the [`examples/`](./examples/) directory, showcasing practical implementations:

### 1. Increment Agent - Sequential Tool Calls

The [increment-agent](./examples/increment-agent/main.go) demonstrates how to build an agent that makes sequential tool calls to perform calculations:

```go
// Agent that starts with a number and increments it using tool calls
type IncrementResult struct {
    FinalNumber int      `json:"final_number" jsonschema_description:"Final result after all increments"`
    Steps       []string `json:"steps"        jsonschema_description:"List of steps taken"`
}

// Creates an agent that uses an "add" tool 3 times to increment by 2 each time
incrementAgent, err := agent.NewAgent(
    agent.WithName[IncrementResult]("increment-agent"),
    agent.WithLLMConfig[IncrementResult](llmConfig),
    agent.WithBehavior[IncrementResult]("You must use the add tool exactly 3 times to add 2 each time"),
    agent.WithTool[IncrementResult]("add", addTool),
    agent.WithToolLimit[IncrementResult]("add", 3),
)
```

**Run it:**
```bash
cd examples/increment-agent
export OPENAI_API_KEY="your-key"
go run main.go
```

### 2. Site Analyzer Agent - HTTP Tool Integration

The [site-analyzer-agent](./examples/site-analyzer-agent/main.go) shows how to build agents that interact with external APIs and perform comprehensive analysis:

```go
// Agent that analyzes websites by fetching and examining their content
type AgentResult struct {
    Title       string   `json:"title"        jsonschema_description:"Title of the site"`
    Purpose     string   `json:"purpose"      jsonschema_description:"Purpose of the site"`
    KeyInsights []string `json:"key_insights" jsonschema_description:"Key insights about the site"`
}

// Creates an agent with HTTP tool to fetch and analyze web content
analyzerAgent, err := agent.NewAgent(
    agent.WithName[AgentResult]("analyzer-agent"),
    agent.WithLLMConfig[AgentResult](llmConfig),
    agent.WithBehavior[AgentResult](websiteAnalysisBehavior),
    agent.WithTool[AgentResult]("http", httpTool),
    agent.WithToolLimit[AgentResult]("http", 10),
)
```

**Run it:**
```bash
cd examples/site-analyzer-agent
export OPENAI_API_KEY="your-key"
go run main.go
```

## 🛠️ Development

### Prerequisites

- Go 1.24.4+
- OpenAI API key for testing

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Linting

```bash
make lint
```

## 🏗️ Architecture

### Core Components

- **Agent** - Main orchestrator with configurable behavior and tools
- **LLM** - Abstraction layer for different language model providers  
- **Tools** - Extensible functions that agents can call
- **Config** - Environment-based configuration management

### Design Principles

- **Type Safety** - Full generic type support for compile-time guarantees
- **Interface Segregation** - Small, focused interfaces
- **Dependency Injection** - Flexible configuration with options pattern
- **Error Handling** - Explicit error types and proper error wrapping

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`make test`)
6. Ensure linting passes (`make lint`)
7. Commit your changes (`git commit -m 'Add amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

### Code Standards

- All public APIs must have documentation comments
- Tests required for new functionality
- Follow Go best practices and idioms
- Use the provided linter configuration

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📬 Stay Connected

- 📖 [Blog](https://vitaliihonchar.com) - Technical articles and insights
- 📧 [Newsletter](https://vitaliihonchar.substack.com/) - Subscribe for updates on Go, AI, and software engineering
- 🐛 [Issue Tracker](https://github.com/vitalii-honchar/go-agent/issues) - Report bugs and feature requests
- 💬 [Discussions](https://github.com/vitalii-honchar/go-agent/discussions) - Community discussions

---

**Built with ❤️ for the Go community**