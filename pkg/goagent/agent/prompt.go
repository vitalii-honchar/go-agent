package agent

import (
	"bytes"
	"fmt"
	"text/template"
)

// Prompt represents a template for generating system prompts
type Prompt struct {
	Template string `json:"template"`
}

// NewPrompt creates a new Prompt with the given template string
func NewPrompt(template string) Prompt {
	return Prompt{
		Template: template,
	}
}

// Render renders the prompt template with the given arguments
func (p Prompt) Render(args map[string]any) (string, error) {
	tmpl, err := template.New("prompt").Parse(p.Template)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, args); err != nil {
		return "", fmt.Errorf("failed to execute prompt template: %w", err)
	}

	return buf.String(), nil
}
