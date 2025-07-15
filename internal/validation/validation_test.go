package validation_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitalii-honchar/go-agent/internal/validation"
)

func TestNameIsValid_ValidNames(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"valid_tool_name",
		"tool_v1_name",
		"tool",
		"my_great_tool_123",
		strings.Repeat("a", 64), // exactly at limit
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := validation.NameIsValid(name)
			require.NoError(t, err)
		})
	}
}

func TestNameIsValid_InvalidNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"empty name", ""},
		{"name with dashes", "invalid-name"},
		{"name with uppercase", "InvalidName"},
		{"name with spaces", "invalid name"},
		{"name with special characters", "invalid@name"},
		{"name too long", strings.Repeat("a", 65)},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			err := validation.NameIsValid(testCase.input)
			require.Error(t, err)
			require.ErrorIs(t, err, validation.ErrValidationFailed)
		})
	}
}

func TestDescriptionIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{
			name:      "valid description",
			input:     "This is a valid description",
			wantError: false,
		},
		{
			name:      "valid description with special characters",
			input:     "This is a valid description with @#$%^&*()!",
			wantError: false,
		},
		{
			name:      "description at max length",
			input:     strings.Repeat("a", 1024),
			wantError: false,
		},
		{
			name:      "empty description",
			input:     "",
			wantError: true,
		},
		{
			name:      "description too long",
			input:     strings.Repeat("a", 1025),
			wantError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validation.DescriptionIsValid(testCase.input)

			if testCase.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, validation.ErrValidationFailed)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNotNil(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     any
		wantError bool
	}{
		{
			name:      "non-nil string",
			input:     "hello",
			wantError: false,
		},
		{
			name:      "non-nil struct",
			input:     struct{}{},
			wantError: false,
		},
		{
			name:      "non-nil function",
			input:     func() {},
			wantError: false,
		},
		{
			name:      "non-nil pointer",
			input:     &struct{}{},
			wantError: false,
		},
		{
			name:      "nil value",
			input:     nil,
			wantError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validation.NotNil(testCase.input)

			if testCase.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, validation.ErrValidationFailed)
				assert.Contains(t, err.Error(), "value cannot be nil")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStringIsNotEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{
			name:      "non-empty string",
			input:     "hello",
			wantError: false,
		},
		{
			name:      "whitespace string",
			input:     "   ",
			wantError: false, // whitespace is not considered empty
		},
		{
			name:      "single character",
			input:     "a",
			wantError: false,
		},
		{
			name:      "empty string",
			input:     "",
			wantError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validation.StringIsNotEmpty(testCase.input)

			if testCase.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, validation.ErrValidationFailed)
				assert.Contains(t, err.Error(), "string cannot be empty")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStringIsMaxLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		maxLen    int
		wantError bool
	}{
		{
			name:      "string within limit",
			input:     "hello",
			maxLen:    10,
			wantError: false,
		},
		{
			name:      "string at exact limit",
			input:     "hello",
			maxLen:    5,
			wantError: false,
		},
		{
			name:      "empty string with any limit",
			input:     "",
			maxLen:    5,
			wantError: false,
		},
		{
			name:      "string exceeds limit",
			input:     "hello world",
			maxLen:    5,
			wantError: true,
		},
		{
			name:      "zero limit with non-empty string",
			input:     "a",
			maxLen:    0,
			wantError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validation.StringIsMaxLength(testCase.input, testCase.maxLen)

			if testCase.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, validation.ErrValidationFailed)
				assert.Contains(t, err.Error(), fmt.Sprintf("string cannot be longer than %d characters", testCase.maxLen))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStringMatchesPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		pattern   string
		wantError bool
	}{
		{
			name:      "valid snake_case pattern",
			input:     "valid_name",
			pattern:   `^[a-z0-9_]+$`,
			wantError: false,
		},
		{
			name:      "valid email pattern",
			input:     "test@example.com",
			pattern:   `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			wantError: false,
		},
		{
			name:      "pattern doesn't match",
			input:     "Invalid-Name",
			pattern:   `^[a-z0-9_]+$`,
			wantError: true,
		},
		{
			name:      "empty string with pattern",
			input:     "",
			pattern:   `^[a-z0-9_]+$`,
			wantError: true,
		},
		{
			name:      "invalid regex pattern",
			input:     "test",
			pattern:   `[unclosed`,
			wantError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := validation.StringMatchesPattern(testCase.input, testCase.pattern)

			if testCase.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, validation.ErrValidationFailed)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
