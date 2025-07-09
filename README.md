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

### Basic Calculator Agent

Here's a simple calculator agent that uses a tool to perform addition:

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

// Define input and output types
type AddNumbers struct {
    Num1 int `json:"num1"`
    Num2 int `json:"num2"`
}

type AddNumbersResult struct {
    Sum int `json:"sum"`
}

// Define tool parameters and result
type AddToolParams struct {
    Num1 float64 `json:"num1"`
    Num2 float64 `json:"num2"`
}

type AddToolResult struct {
    llm.BaseLLMToolResult
    Sum float64 `json:"sum"`
}

func main() {
    // Create an addition tool
    addTool := llm.NewLLMTool(
        llm.WithLLMToolName("add"),
        llm.WithLLMToolDescription("Adds two numbers together"),
        llm.WithLLMToolParametersSchema[AddToolParams](),
        llm.WithLLMToolCall(func(callID string, params AddToolParams) (AddToolResult, error) {
            return AddToolResult{
                BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
                Sum:              params.Num1 + params.Num2,
            }, nil
        }),
    )

    // Create calculator agent
    calculatorAgent, err := agent.NewAgent(
        agent.WithName[AddNumbersResult]("calculator"),
        agent.WithLLMConfig[AddNumbersResult](llm.LLMConfig{
            Type:        llm.LLMTypeOpenAI,
            APIKey:      os.Getenv("OPENAI_API_KEY"),
            Model:       "gpt-4",
            Temperature: 0.0,
        }),
        agent.WithBehavior[AddNumbersResult](
            "You are a calculator agent. You MUST use the add tool to calculate the sum of the two provided numbers. " +
                "Do NOT calculate manually. Return the result in the specified JSON format."),
        agent.WithTool[AddNumbersResult]("add", addTool),
        agent.WithToolLimit[AddNumbersResult]("add", 1),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Run the agent
    input := AddNumbers{Num1: 3, Num2: 5}
    result, err := calculatorAgent.Run(context.Background(), input)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Result: %d + %d = %d\n", input.Num1, input.Num2, result.Data.Sum)
}
```

### Sequential Tool Calls

This example shows an agent that makes multiple sequential tool calls:

```go
// Define types for incremental operations
type IncrementInput struct {
    StartNumber int `json:"start_number"`
    Steps       int `json:"steps"`
}

type IncrementResult struct {
    FinalNumber int      `json:"final_number"`
    Steps       []string `json:"steps"`
}

func createIncrementAgent() (*agent.Agent[IncrementResult], error) {
    // Create the same add tool as above
    addTool := llm.NewLLMTool(
        llm.WithLLMToolName("add"),
        llm.WithLLMToolDescription("Adds two numbers together"),
        llm.WithLLMToolParametersSchema[AddToolParams](),
        llm.WithLLMToolCall(func(callID string, params AddToolParams) (AddToolResult, error) {
            return AddToolResult{
                BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
                Sum:              params.Num1 + params.Num2,
            }, nil
        }),
    )

    return agent.NewAgent(
        agent.WithName[IncrementResult]("incrementer"),
        agent.WithLLMConfig[IncrementResult](llm.LLMConfig{
            Type:        llm.LLMTypeOpenAI,
            APIKey:      os.Getenv("OPENAI_API_KEY"),
            Model:       "gpt-4",
            Temperature: 0.0,
        }),
        agent.WithBehavior[IncrementResult](`You are an incrementer agent. You must:
1. Start with the provided start_number
2. Use the add tool to add 1 to it (steps times)
3. For 3 steps: call add(start_number, 1), then add(result, 1), then add(result, 1)
4. Track each step showing the calculation
5. Return the final number and all steps taken

You MUST use the add tool for each increment. Do NOT calculate manually.`),
        agent.WithTool[IncrementResult]("add", addTool),
        agent.WithToolLimit[IncrementResult]("add", 5),
    )
}

func main() {
    incrementAgent, err := createIncrementAgent()
    if err != nil {
        log.Fatal(err)
    }

    input := IncrementInput{StartNumber: 2, Steps: 3}
    result, err := incrementAgent.Run(context.Background(), input)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Started with: %d\n", input.StartNumber)
    fmt.Printf("Final result: %d\n", result.Data.FinalNumber)
    fmt.Printf("Steps taken: %v\n", result.Data.Steps)
}
```

### Multi-Tool Agent

This example demonstrates using multiple tools with different limits:

```go
// Hash tool types
type HashToolParams struct {
    Input string `json:"input"`
}

type HashToolResult struct {
    llm.BaseLLMToolResult
    Hash string `json:"hash"`
}

func createMultiToolAgent() (*agent.Agent[IncrementResult], error) {
    // Create add tool (same as above)
    addTool := llm.NewLLMTool(
        llm.WithLLMToolName("add"),
        llm.WithLLMToolDescription("Adds two numbers together"),
        llm.WithLLMToolParametersSchema[AddToolParams](),
        llm.WithLLMToolCall(func(callID string, params AddToolParams) (AddToolResult, error) {
            return AddToolResult{
                BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
                Sum:              params.Num1 + params.Num2,
            }, nil
        }),
    )

    // Create hash tool
    hashTool := llm.NewLLMTool(
        llm.WithLLMToolName("hash"),
        llm.WithLLMToolDescription("Computes SHA256 hash of input string"),
        llm.WithLLMToolParametersSchema[HashToolParams](),
        llm.WithLLMToolCall[HashToolParams, HashToolResult](
            func(callID string, params HashToolParams) (HashToolResult, error) {
                hash := fmt.Sprintf("%x", []byte(params.Input))
                return HashToolResult{
                    BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
                    Hash:              hash,
                }, nil
            }),
    )

    return agent.NewAgent(
        agent.WithName[IncrementResult]("multi-tool-tester"),
        agent.WithLLMConfig[IncrementResult](llm.LLMConfig{
            Type:        llm.LLMTypeOpenAI,
            APIKey:      os.Getenv("OPENAI_API_KEY"),
            Model:       "gpt-4",
            Temperature: 0.0,
        }),
        agent.WithBehavior[IncrementResult](`You are a multi-tool testing agent. You must:
1. Start with the provided start_number (10)
2. Use the add tool exactly 3 times to add 2 each time: add(10,2), add(12,2), add(14,2) 
3. Then use the hash tool to compute hash of "test"
4. Track each step and return final number

You have add tool limit of 3 and hash tool limit of 1. Use add tool 3 times, then hash tool 1 time.`),
        agent.WithTool[IncrementResult]("add", addTool),
        agent.WithTool[IncrementResult]("hash", hashTool),
        agent.WithToolLimit[IncrementResult]("add", 3),
        agent.WithToolLimit[IncrementResult]("hash", 1),
    )
}
```

### Tool Limits and Error Handling

This example shows how to handle tool limits:

```go
import (
    "context"
    "errors"
    "fmt"
    "log"
    "os"
    
    "github.com/vitalii-honchar/go-agent/pkg/goagent/agent"
    "github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
)

func createLimitedAgent() (*agent.Agent[IncrementResult], error) {
    addTool := llm.NewLLMTool(
        llm.WithLLMToolName("add"),
        llm.WithLLMToolDescription("Adds two numbers together"),
        llm.WithLLMToolParametersSchema[AddToolParams](),
        llm.WithLLMToolCall(func(callID string, params AddToolParams) (AddToolResult, error) {
            return AddToolResult{
                BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
                Sum:              params.Num1 + params.Num2,
            }, nil
        }),
    )

    return agent.NewAgent(
        agent.WithName[IncrementResult]("limit-tester"),
        agent.WithLLMConfig[IncrementResult](llm.LLMConfig{
            Type:        llm.LLMTypeOpenAI,
            APIKey:      os.Getenv("OPENAI_API_KEY"),
            Model:       "gpt-4",
            Temperature: 0.0,
        }),
        agent.WithBehavior[IncrementResult](`You are a calculation agent that can ONLY perform arithmetic using the add tool. 
You have no ability to calculate numbers manually.

Your task: You need to increment the start_number by adding complex floating point numbers that you cannot calculate yourself.

Process:
1. Use add tool to add start_number + 0.12345
2. Use add tool to add result + 0.23456  
3. Use add tool to add result + 0.34567
4. Use add tool to add result + 0.45678
5. Use add tool to add result + 0.56789

You MUST use the add tool for each step because these floating point calculations are too complex for you to do manually. 
Do not try to calculate yourself - you will get wrong results. Always use the add tool.`),
        agent.WithTool[IncrementResult]("add", addTool),
        agent.WithToolLimit[IncrementResult]("add", 1), // Very restrictive limit
    )
}

func main() {
    limitedAgent, err := createLimitedAgent()
    if err != nil {
        log.Fatal(err)
    }

    input := IncrementInput{StartNumber: 100, Steps: 3}
    result, err := limitedAgent.Run(context.Background(), input)
    
    // Check if limit was reached
    if err != nil {
        if errors.Is(err, agent.ErrLimitReached) {
            fmt.Printf("Tool limit reached as expected: %v\n", err)
            fmt.Printf("Partial result available: %v\n", result != nil)
            fmt.Printf("Messages: %d\n", len(result.Messages))
        } else {
            log.Fatal(err)
        }
    } else {
        fmt.Printf("Final result: %d\n", result.Data.FinalNumber)
    }
}
```

### Default Tool Limits

You can set default limits for all tools:

```go
agent, err := agent.NewAgent(
    agent.WithName[MyResult]("default-limit-agent"),
    agent.WithLLMConfig[MyResult](llmConfig),
    agent.WithBehavior[MyResult]("Your behavior here"),
    agent.WithTool[MyResult]("tool1", tool1),
    agent.WithTool[MyResult]("tool2", tool2),
    agent.WithDefaultToolLimit[MyResult](2),              // Default limit for all tools
    agent.WithToolLimit[MyResult]("tool1", 5),            // Override for specific tool
    // tool2 will use the default limit of 2
)
```

### More Examples

For additional usage patterns, check the comprehensive test suite in `pkg/goagent/agent/agent_test.go`:

- **TestSumAgent** - Basic calculator with tool usage
- **TestHashAgent** - Text hashing with custom tools
- **TestSequentialToolCalls** - Multiple sequential operations
- **TestMultiToolLimitReached** - Multiple tools with different limits
- **TestDefaultToolLimit** - Default tool limit configuration
- **TestCustomDefaultToolLimit** - Custom default limits

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