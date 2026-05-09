package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func NormalizeDynamicSchema(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("%w: dynamic schema cannot be empty", ErrCannotCreateSchema)
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCannotCreateSchema, err)
	}
	if _, ok := value.(map[string]any); !ok {
		return nil, fmt.Errorf("%w: dynamic schema must be a JSON object", ErrCannotCreateSchema)
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCannotCreateSchema, err)
	}

	return append(json.RawMessage(nil), compact.Bytes()...), nil
}

func DynamicSchemaMap(raw json.RawMessage) (map[string]any, error) {
	normalized, err := NormalizeDynamicSchema(raw)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(normalized, &result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCannotCreateSchema, err)
	}

	return result, nil
}
