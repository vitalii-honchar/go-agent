package schema

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/invopop/jsonschema"
)

var ErrCannotCreateSchema = errors.New("cannot create schema from output type")

// Generate a JSON schema from the Go type T
func GenerateSchema(schemaT any) (map[string]any, error) {
	schema, err := GenerateSchemaStr(schemaT)

	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(schema), &result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCannotCreateSchema, err)
	}

	return result, nil
}

func GenerateSchemaStr(schemaT any) (string, error) {
	schema := jsonschema.Reflect(schemaT)
	schemaMap, err := json.Marshal(schema)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrCannotCreateSchema, err)
	}
	return string(schemaMap), nil
}
