// Package goagent provides a powerful, production-ready Go library for building AI agents
// with configurable behavior, custom tools, and type-safe output schemas.
//
// This library is perfect for building intelligent automation, data analysis tools,
// web scrapers, and AI-powered applications with robust error handling and clean architecture.
//
// Key Features:
//
//   - Type-safe agents with custom output schemas using Go generics
//   - Extensible tool system with automatic limit enforcement
//   - Configurable agent behavior using natural language
//   - Multiple LLM support (OpenAI, extensible to others)
//   - Structured JSON output with schema validation
//   - Production-ready with comprehensive error handling
//
// Quick Start:
//
// The easiest way to get started is to create a simple agent:
//
//	package main
//
//	import (
//		"context"
//		"fmt"
//		"log"
//
//		"github.com/vitalii-honchar/go-agent/pkg/goagent/agent"
//		"github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
//	)
//
//	type Result struct {
//		Answer string `json:"answer" jsonschema_description:"The answer"`
//	}
//
//	func main() {
//		agent, err := agent.NewAgent(
//			agent.WithName[Result]("my-agent"),
//			agent.WithLLMConfig[Result](llm.LLMConfig{
//				Type:        llm.LLMTypeOpenAI,
//				APIKey:      "your-openai-api-key",
//				Model:       "gpt-4",
//				Temperature: 0.0,
//			}),
//			agent.WithBehavior[Result]("You are a helpful assistant."),
//		)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		result, err := agent.Run(context.Background(), "What is 2+2?")
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		fmt.Println("Answer:", result.Data.Answer)
//	}
//
// Package Structure:
//
//   - agent: Core agent functionality and orchestration
//   - llm: LLM abstractions and tool system
//   - config: Environment-based configuration management
//
// For comprehensive examples and advanced usage, see:
// https://github.com/vitalii-honchar/go-agent/tree/main/examples
//
// Documentation and tutorials:
// https://vitaliihonchar.com/insights/go-ai-agent-library
package goagent
