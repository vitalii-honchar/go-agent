package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/vitalii-honchar/go-agent/pkg/goagent/agent"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/llm"
)

type (
	AgentInput struct {
		URL string `json:"url" jsonschema_description:"URL of the site to analyze"`
	}

	AgentResult struct {
		Title       string   `json:"title"        jsonschema_description:"Title of the site"`
		Purpose     string   `json:"purpose"      jsonschema_description:"Purpose of the site"`
		KeyInsights []string `json:"key_insights" jsonschema_description:"Key insights about the site"`
	}

	HttpToolParams struct {
		URL string `json:"url" jsonschema_description:"URL to fetch"`
	}

	HttpToolResult struct {
		llm.BaseLLMToolResult
		StatusCode int               `json:"status_code" jsonschema_description:"HTTP status code"`
		Body       string            `json:"body"        jsonschema_description:"Response body"`
		Headers    map[string]string `json:"headers"     jsonschema_description:"Response headers"`
	}
)

func main() {
	analyzerAgent, err := createAnalyzerAgent()
	if err != nil {
		log.Fatalf("Failed to create analyzer-agent: %v", err)
	}

	input := AgentInput{URL: "https://vitaliihonchar.com/"}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	runAnalysis(ctx, analyzerAgent, input)
}

func createAnalyzerAgent() (*agent.Agent[AgentResult], error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	analyzerAgent, err := agent.NewAgent(
		agent.WithName[AgentResult]("analyzer-agent"),
		agent.WithLLMConfig[AgentResult](llm.LLMConfig{
			Type:        llm.LLMTypeOpenAI,
			APIKey:      apiKey,
			Model:       "gpt-4.1",
			Temperature: 0.0,
		}),
		agent.WithBehavior[AgentResult](getAnalysisBehavior()),
		agent.WithTool[AgentResult]("http", createHttpTool()),
		agent.WithToolLimit[AgentResult]("http", 10),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create analyzer agent: %w", err)
	}

	return analyzerAgent, nil
}

func getAnalysisBehavior() string {
	return `You are a website analysis expert specializing in ` +
		`comprehensive site evaluation and content extraction.

		ANALYSIS PROCEDURE:
		1. INITIAL FETCH: Use the http tool to fetch the main page content
		2. CONTENT ANALYSIS: Analyze HTML structure, meta tags, headings, and visible text
		3. DEEP EXPLORATION: Look for additional pages, contact info, about sections, or portfolio links
		4. STRUCTURE MAPPING: Identify navigation patterns, page hierarchy, and site organization
		5. PURPOSE IDENTIFICATION: Determine the primary function and target audience
		6. INSIGHT EXTRACTION: Extract key technical details, business model, and unique features

		FOCUS AREAS:
		- Site title, description, and branding elements
		- Main content themes and messaging
		- Technical stack indicators (frameworks, libraries)
		- Business/personal information and contact details
		- Key features, services, or products offered
		- Design patterns and user experience elements

		THOROUGHNESS REQUIREMENT:
		DO NOT BE LAZY! You must continue analyzing until you have exhausted all ` +
		`available information or reached the tool limit. Start with the main page, ` +
		`then explore additional pages like:
		- /about, /contact, /portfolio, /services, /products
		- Any links found in navigation menus or footer
		- Subpages that provide more context about the site owner or business
		- Continue fetching pages until you have a complete picture or hit the 3-request limit

		OUTPUT REQUIREMENTS:
		- Provide a clear, descriptive title based on actual content
		- Summarize the site's primary purpose in 1-2 sentences
		- List 5-10 key insights that reveal important aspects of the site`
}

func runAnalysis(ctx context.Context, analyzerAgent *agent.Agent[AgentResult], input AgentInput) {
	log.Printf("Starting analyzer-agent...")
	res, err := analyzerAgent.Run(ctx, input)
	if err != nil {
		log.Fatalf("Failed to run analyzer-agent: %v", err)
	}

	data, err := json.MarshalIndent(res.Data, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal result: %v", err)
	}
	log.Printf("Analyzer result: %s\n", data)
}

func createHttpTool() llm.LLMTool {
	tool, err := llm.NewLLMTool(
		llm.WithLLMToolName("http"),
		llm.WithLLMToolDescription("Fetches content from a URL via HTTP GET request"),
		llm.WithLLMToolParametersSchema[HttpToolParams](),
		llm.WithLLMToolCall(handleHttpRequest),
	)
	if err != nil {
		log.Fatalf("Failed to create http tool: %v", err)
	}

	return tool
}

func handleHttpRequest(callID string, params HttpToolParams) (HttpToolResult, error) {
	log.Printf("🌐 HTTP CALL: GET %s", params.URL)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, params.URL, nil)
	if err != nil {
		return createErrorResult(callID, 0, "Error creating request: "+err.Error()), nil
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return createErrorResult(callID, 0, "Error: "+err.Error()), nil
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Error closing response body: %v", closeErr)
		}
	}()

	headers := extractHeaders(resp.Header)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return HttpToolResult{
			BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
			StatusCode:        resp.StatusCode,
			Body:              "Error reading response body: " + err.Error(),
			Headers:           headers,
		}, nil
	}

	result := HttpToolResult{
		BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
		StatusCode:        resp.StatusCode,
		Body:              string(body),
		Headers:           headers,
	}

	log.Printf("🌐 HTTP RESULT: %d - %d bytes", resp.StatusCode, len(body))

	return result, nil
}

func createErrorResult(callID string, statusCode int, errorMsg string) HttpToolResult {
	return HttpToolResult{
		BaseLLMToolResult: llm.BaseLLMToolResult{ID: callID},
		StatusCode:        statusCode,
		Body:              errorMsg,
		Headers:           map[string]string{},
	}
}

func extractHeaders(header http.Header) map[string]string {
	headers := make(map[string]string)
	for name, values := range header {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}

	return headers
}
