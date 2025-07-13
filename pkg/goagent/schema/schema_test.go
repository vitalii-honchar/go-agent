package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vitalii-honchar/go-agent/pkg/goagent/schema"
)

func TestGenerateSchema_SimpleStruct(t *testing.T) {
	type Person struct {
		Name string `json:"name" jsonschema_description:"Person's name"`
		Age  int    `json:"age" jsonschema_description:"Person's age"`
	}

	result, err := schema.GenerateSchema(Person{})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "type")
	assert.Equal(t, "object", result["type"])
	assert.Contains(t, result, "properties")

	properties, ok := result["properties"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, properties, "name")
	assert.Contains(t, properties, "age")
}

func TestGenerateSchema_NestedStruct(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type Person struct {
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	result, err := schema.GenerateSchema(Person{})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "properties")

	properties, ok := result["properties"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, properties, "name")
	assert.Contains(t, properties, "address")

	address, ok := properties["address"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, address, "properties")
}

func TestGenerateSchema_WithSlice(t *testing.T) {
	type Person struct {
		Name  string   `json:"name"`
		Hobbies []string `json:"hobbies"`
	}

	result, err := schema.GenerateSchema(Person{})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "properties")

	properties, ok := result["properties"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, properties, "hobbies")

	hobbies, ok := properties["hobbies"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "array", hobbies["type"])
}

func TestGenerateSchema_WithPointer(t *testing.T) {
	type Person struct {
		Name *string `json:"name,omitempty"`
		Age  int     `json:"age"`
	}

	result, err := schema.GenerateSchema(Person{})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "properties")
}

func TestGenerateSchemaStr_SimpleStruct(t *testing.T) {
	type Simple struct {
		Value string `json:"value"`
	}

	result, err := schema.GenerateSchemaStr(Simple{})

	require.NoError(t, err)
	assert.NotEmpty(t, result)

	// Verify it's valid JSON
	var jsonMap map[string]interface{}
	err = json.Unmarshal([]byte(result), &jsonMap)
	require.NoError(t, err)
	assert.Contains(t, jsonMap, "type")
	assert.Equal(t, "object", jsonMap["type"])
}

func TestGenerateSchemaStr_EmptyStruct(t *testing.T) {
	type Empty struct{}

	result, err := schema.GenerateSchemaStr(Empty{})

	require.NoError(t, err)
	assert.NotEmpty(t, result)

	// Verify it's valid JSON
	var jsonMap map[string]interface{}
	err = json.Unmarshal([]byte(result), &jsonMap)
	require.NoError(t, err)
	assert.Contains(t, jsonMap, "type")
}

func TestGenerateSchema_ComplexTypes(t *testing.T) {
	type Config struct {
		Enabled    bool               `json:"enabled"`
		Count      int                `json:"count"`
		Rate       float64            `json:"rate"`
		Tags       []string           `json:"tags"`
		Properties map[string]string  `json:"properties"`
	}

	result, err := schema.GenerateSchema(Config{})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "properties")

	properties, ok := result["properties"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, properties, "enabled")
	assert.Contains(t, properties, "count")
	assert.Contains(t, properties, "rate")
	assert.Contains(t, properties, "tags")
	assert.Contains(t, properties, "properties")
}

func TestGenerateSchema_WithDescription(t *testing.T) {
	type Documented struct {
		Name string `json:"name" jsonschema_description:"The user's full name"`
		Age  int    `json:"age" jsonschema_description:"The user's age in years"`
	}

	result, err := schema.GenerateSchema(Documented{})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Contains(t, result, "properties")

	properties, ok := result["properties"].(map[string]interface{})
	require.True(t, ok)
	
	name, ok := properties["name"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, name, "description")
	assert.Equal(t, "The user's full name", name["description"])
}