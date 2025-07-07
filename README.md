# Go Agent

A Go library for building AI agents with configurable behavior, tools, and output schemas.

[![CI](https://github.com/vitalii-honchar/go-agent/actions/workflows/ci.yml/badge.svg)](https://github.com/vitalii-honchar/go-agent/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/vitalii-honchar/go-agent)](https://goreportcard.com/report/github.com/vitalii-honchar/go-agent)
[![GoDoc](https://godoc.org/github.com/vitalii-honchar/go-agent?status.svg)](https://godoc.org/github.com/vitalii-honchar/go-agent)

## Features

- 🤖 **Generic AI Agents** - Type-safe agents with custom output schemas
- 🔧 **Extensible Tools** - Add custom tools with automatic limit enforcement
- 🎯 **Configurable Behavior** - Define agent behavior with natural language
- 🚀 **Multiple LLM Support** - Currently supports OpenAI (extensible to others)
- ⚡ **Tool Limits** - Prevent runaway execution with per-tool usage limits
- 📝 **Structured Output** - JSON schema validation for reliable results
- 🛡️ **Type Safety** - Full Go generics support for compile-time safety

## Quick Start

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
    Answer int    `json:"answer"`
    Steps  string `json:"steps"`
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
        agent.WithOutputSchema(&MathResult{}),
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

## Advanced Usage

### Adding Custom Tools

```go
// Create a custom tool
addTool := llm.NewLLMTool(
    llm.WithLLMToolName("add"),
    llm.WithLLMToolDescription("Adds two numbers together"),
    llm.WithLLMToolParametersSchema(map[string]any{
        "type": "object",
        "properties": map[string]any{
            "a": map[string]any{"type": "number"},
            "b": map[string]any{"type": "number"},
        },
        "required": []string{"a", "b"},
    }),
    llm.WithLLMToolCall(func(id string, args map[string]any) (AddResult, error) {
        a := args["a"].(float64)
        b := args["b"].(float64)
        return AddResult{
            BaseLLMToolResult: llm.BaseLLMToolResult{ID: id},
            Sum:              a + b,
        }, nil
    }),
)

// Add tool to agent with usage limit
calculatorAgent, err := agent.NewAgent(
    agent.WithName[CalculatorResult]("calculator"),
    agent.WithLLMConfig[CalculatorResult](llmConfig),
    agent.WithBehavior[CalculatorResult]("Use the add tool to calculate sums. Do not calculate manually."),
    agent.WithTool[CalculatorResult]("add", addTool),
    agent.WithToolLimit[CalculatorResult]("add", 5), // Max 5 calls
    agent.WithOutputSchema(&CalculatorResult{}),
)
```

### Tool Limits

Control tool usage to prevent runaway execution:

```go
agent, err := agent.NewAgent(
    // ... other options
    agent.WithDefaultToolLimit[MyResult](3),           // Default limit for all tools
    agent.WithToolLimit[MyResult]("expensive_tool", 1), // Specific limit for one tool
)
```

### Custom System Prompts

```go
customPrompt := agent.NewPrompt(`
You are {{.agent_name}} with the following tools: {{.tools}}.
Your behavior: {{.behavior}}
Output format: {{.output_schema}}
`)

agent, err := agent.NewAgent(
    // ... other options
    agent.WithSystemPrompt[MyResult](customPrompt),
)
```

## Configuration

### Environment Variables

Set your API keys and configuration:

```bash
export OPENAI_API_KEY="your-openai-api-key"
export OPENAI_MODEL="gpt-4"                    # Optional, defaults to gpt-4
export OPENAI_TEMPERATURE="0.0"               # Optional, defaults to 0.7
export OPENAI_MAX_TOKENS="4096"              # Optional, defaults to 4096
export OPENAI_TIMEOUT_SECONDS="30"           # Optional, defaults to 30
```

Or use a `.env` file:

```
OPENAI_API_KEY=your-openai-api-key
OPENAI_MODEL=gpt-4
OPENAI_TEMPERATURE=0.0
```

### LLM Configuration

```go
config := llm.LLMConfig{
    Type:        llm.LLMTypeOpenAI,
    APIKey:      "your-api-key",
    Model:       "gpt-4",
    Temperature: 0.1,
}
```

## Examples

### Complete Working Example

Here's a complete example that demonstrates the library's key features:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/vitalii-honchar/go-agent/pkg/goagent/agent"
    "github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
)

type WeatherResult struct {
    Location    string  `json:"location"`
    Temperature float64 `json:"temperature"`
    Condition   string  `json:"condition"`
    Humidity    int     `json:"humidity"`
}

type WeatherToolResult struct {
    llm.BaseLLMToolResult
    Data WeatherResult `json:"data"`
}

func main() {
    // Create a mock weather tool
    weatherTool := llm.NewLLMTool(
        llm.WithLLMToolName("get_weather"),
        llm.WithLLMToolDescription("Get current weather for a location"),
        llm.WithLLMToolParametersSchema(map[string]any{
            "type": "object",
            "properties": map[string]any{
                "location": map[string]any{
                    "type": "string",
                    "description": "City name",
                },
            },
            "required": []string{"location"},
        }),
        llm.WithLLMToolCall(func(id string, args map[string]any) (WeatherToolResult, error) {
            location := args["location"].(string)
            // Mock weather data
            return WeatherToolResult{
                BaseLLMToolResult: llm.BaseLLMToolResult{ID: id},
                Data: WeatherResult{
                    Location:    location,
                    Temperature: 22.5,
                    Condition:   "Sunny",
                    Humidity:    65,
                },
            }, nil
        }),
    )

    // Create weather agent
    weatherAgent, err := agent.NewAgent(
        agent.WithName[WeatherResult]("weather-assistant"),
        agent.WithLLMConfig[WeatherResult](llm.LLMConfig{
            Type:        llm.LLMTypeOpenAI,
            APIKey:      os.Getenv("OPENAI_API_KEY"),
            Model:       "gpt-4",
            Temperature: 0.1,
        }),
        agent.WithBehavior[WeatherResult](
            "You are a weather assistant. Use the get_weather tool to fetch current weather data for any location requested. Always provide the temperature in Celsius.",
        ),
        agent.WithTool[WeatherResult]("get_weather", weatherTool),
        agent.WithToolLimit[WeatherResult]("get_weather", 3),
        agent.WithOutputSchema(&WeatherResult{}),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Run the agent
    result, err := weatherAgent.Run(context.Background(), "What's the weather like in New York?")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Weather in %s:\n", result.Data.Location)
    fmt.Printf("Temperature: %.1f°C\n", result.Data.Temperature)
    fmt.Printf("Condition: %s\n", result.Data.Condition)
    fmt.Printf("Humidity: %d%%\n", result.Data.Humidity)
}
```

### More Examples

For additional usage patterns, check the comprehensive test suite in `pkg/goagent/agent/agent_test.go`:

- **Basic Agent** - Simple question answering
- **Calculator Agent** - Math operations with tools
- **Multi-Tool Agent** - Using multiple tools together
- **Tool Limits** - Usage limits and error handling

## Architecture

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

## Development

### Prerequisites

- Go 1.24.4+
- OpenAI API key for testing

### Building

```bash
go build ./pkg/goagent/...
```

### Testing

```bash
make test
# or
go test ./... -v
```

### Linting

```bash
make lint
# or
go vet ./...
```

### Running Examples

Examples are coming soon! Check the test files for usage patterns:

```bash
go test ./pkg/goagent/agent -v
```

## Contributing

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

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and breaking changes.

## Support

- 📖 [Documentation](https://godoc.org/github.com/vitalii-honchar/go-agent)
- 🐛 [Issue Tracker](https://github.com/vitalii-honchar/go-agent/issues)
- 💬 [Discussions](https://github.com/vitalii-honchar/go-agent/discussions)

---

Built with ❤️ for the Go community