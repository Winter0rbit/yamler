package yamler

import (
	"testing"

	"github.com/Winter0rbit/yamler/internal/testutil"
)

func TestValidation(t *testing.T) {
	t.Run("Basic Types", func(t *testing.T) {
		yamlContent := `
string: hello
number: 42
float: 3.14
boolean: true
array:
  - item1
  - item2
map:
  key1: value1
  key2: value2
`
		doc, err := Load(yamlContent)
		if err != nil {
			t.Fatalf("Failed to load YAML: %v", err)
		}

		// String validation
		stringSchema := &ValidationRule{
			Type:      TypeString,
			MinLength: testutil.IntPtr(3),
			MaxLength: testutil.IntPtr(10),
			Pattern:   testutil.StrPtr("^[a-z]+$"),
		}
		stringNode := doc.root.Content[0].Content[1] // Get the "string" field value
		if err := validateNode(stringNode, stringSchema, "string"); err != nil {
			t.Errorf("Unexpected validation error for string schema: %v", err)
		}

		// Number validation
		numberSchema := &ValidationRule{
			Type:    TypeInt,
			Minimum: testutil.Float64Ptr(0),
			Maximum: testutil.Float64Ptr(100),
		}
		numberNode := doc.root.Content[0].Content[3] // Get the "number" field value
		if err := validateNode(numberNode, numberSchema, "number"); err != nil {
			t.Errorf("Unexpected validation error for number schema: %v", err)
		}

		// Array validation
		arraySchema := &ValidationRule{
			Type:        TypeArray,
			MinItems:    testutil.IntPtr(1),
			MaxItems:    testutil.IntPtr(5),
			UniqueItems: true,
			Items: &ValidationRule{
				Type: TypeString,
			},
		}
		arrayNode := doc.root.Content[0].Content[9] // Get the "array" field value
		if err := validateNode(arrayNode, arraySchema, "array"); err != nil {
			t.Errorf("Unexpected validation error for array schema: %v", err)
		}

		// Map validation
		mapSchema := &ValidationRule{
			Type:     TypeMap,
			Required: []string{"key1", "key2"},
			Properties: map[string]*ValidationRule{
				"key1": {Type: TypeString},
				"key2": {Type: TypeString},
			},
		}
		mapNode := doc.root.Content[0].Content[11] // Get the "map" field value
		if err := validateNode(mapNode, mapSchema, "map"); err != nil {
			t.Errorf("Unexpected validation error for map schema: %v", err)
		}

		// Test invalid values
		invalidStringSchema := &ValidationRule{
			Type:      TypeString,
			MinLength: testutil.IntPtr(10), // String "hello" is too short
		}
		if err := validateNode(stringNode, invalidStringSchema, "string"); err == nil {
			t.Error("Expected validation error for invalid string")
		}

		invalidNumberSchema := &ValidationRule{
			Type:    TypeInt,
			Maximum: testutil.Float64Ptr(10), // Number 42 is too large
		}
		if err := validateNode(numberNode, invalidNumberSchema, "number"); err == nil {
			t.Error("Expected validation error for invalid number")
		}
	})

	t.Run("Complex Schema", func(t *testing.T) {
		yamlContent := `
database:
  host: localhost
  port: 5432
  credentials:
    username: admin
    password: secret
  options:
    - name: timeout
      value: 30
    - name: maxConnections
      value: 100
settings:
  debug: true
  features:
    - enabled: true
      name: feature1
    - enabled: false
      name: feature2
`
		doc, err := Load(yamlContent)
		if err != nil {
			t.Fatalf("Failed to load YAML: %v", err)
		}

		schema := &ValidationRule{
			Type: TypeMap,
			Properties: map[string]*ValidationRule{
				"database": {
					Type: TypeMap,
					Required: []string{
						"host",
						"port",
						"credentials",
					},
					Properties: map[string]*ValidationRule{
						"host": {
							Type:    TypeString,
							Pattern: testutil.StrPtr("^[a-zA-Z0-9.-]+$"),
						},
						"port": {
							Type:    TypeInt,
							Minimum: testutil.Float64Ptr(1),
							Maximum: testutil.Float64Ptr(65535),
						},
						"credentials": {
							Type: TypeMap,
							Required: []string{
								"username",
								"password",
							},
							Properties: map[string]*ValidationRule{
								"username": {Type: TypeString},
								"password": {Type: TypeString},
							},
						},
						"options": {
							Type: TypeArray,
							Items: &ValidationRule{
								Type: TypeMap,
								Required: []string{
									"name",
									"value",
								},
								Properties: map[string]*ValidationRule{
									"name":  {Type: TypeString},
									"value": {Type: TypeAny},
								},
							},
						},
					},
				},
				"settings": {
					Type: TypeMap,
					Properties: map[string]*ValidationRule{
						"debug": {Type: TypeBool},
						"features": {
							Type: TypeArray,
							Items: &ValidationRule{
								Type: TypeMap,
								Required: []string{
									"enabled",
									"name",
								},
								Properties: map[string]*ValidationRule{
									"enabled": {Type: TypeBool},
									"name":    {Type: TypeString},
								},
							},
						},
					},
				},
			},
		}

		if err := doc.Validate(schema); err != nil {
			t.Errorf("Validation failed: %v", err)
		}

		// Test invalid document
		invalidContent := `
database:
  host: ""  # Invalid host
  port: 999999  # Port out of range
  credentials:
    username: admin
    # Missing required password
settings:
  debug: not-a-bool  # Invalid boolean
  features:
    - enabled: true
      # Missing required name
`
		invalidDoc, err := Load(invalidContent)
		if err != nil {
			t.Fatalf("Failed to load invalid YAML: %v", err)
		}

		if err := invalidDoc.Validate(schema); err == nil {
			t.Error("Expected validation error for invalid document")
		}
	})
}

func TestLoadSchema(t *testing.T) {
	t.Run("Load Schema from String", func(t *testing.T) {
		schemaContent := `
type: map
properties:
  name:
    type: string
    minLength: 3
    maxLength: 50
  age:
    type: int
    minimum: 0
    maximum: 120
  email:
    type: string
    pattern: ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$
required:
  - name
  - email
`
		schema, err := LoadSchemaFromString(schemaContent)
		if err != nil {
			t.Fatalf("Failed to load schema: %v", err)
		}

		if schema.Type != TypeMap {
			t.Errorf("Expected schema type to be map, got %s", schema.Type)
		}

		if len(schema.Required) != 2 {
			t.Errorf("Expected 2 required fields, got %d", len(schema.Required))
		}

		if schema.Properties["name"].Type != TypeString {
			t.Errorf("Expected name type to be string, got %s", schema.Properties["name"].Type)
		}

		if *schema.Properties["age"].Maximum != 120 {
			t.Errorf("Expected age maximum to be 120, got %f", *schema.Properties["age"].Maximum)
		}
	})
}

func TestValidationRules(t *testing.T) {
	// Tests will be moved here from yamler_test.go
}
