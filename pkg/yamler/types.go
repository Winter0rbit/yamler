package yamler

import "gopkg.in/yaml.v3"

// Document represents a YAML document with preserved formatting
type Document struct {
	root *yaml.Node
	raw  string
}

// SchemaType represents the type of a YAML value
type SchemaType string

const (
	TypeString SchemaType = "string"
	TypeInt    SchemaType = "int"
	TypeFloat  SchemaType = "float"
	TypeBool   SchemaType = "bool"
	TypeArray  SchemaType = "array"
	TypeMap    SchemaType = "map"
	TypeAny    SchemaType = "any"
)

// ValidationRule represents a validation rule for a YAML value
type ValidationRule struct {
	Type SchemaType `yaml:"type"`
	// String validation
	MinLength *int    `yaml:"minLength,omitempty"`
	MaxLength *int    `yaml:"maxLength,omitempty"`
	Pattern   *string `yaml:"pattern,omitempty"`
	// Number validation
	Minimum          *float64 `yaml:"minimum,omitempty"`
	Maximum          *float64 `yaml:"maximum,omitempty"`
	ExclusiveMinimum *float64 `yaml:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *float64 `yaml:"exclusiveMaximum,omitempty"`
	// Array validation
	MinItems    *int            `yaml:"minItems,omitempty"`
	MaxItems    *int            `yaml:"maxItems,omitempty"`
	UniqueItems bool            `yaml:"uniqueItems,omitempty"`
	Items       *ValidationRule `yaml:"items,omitempty"`
	// Map validation
	Required             []string                   `yaml:"required,omitempty"`
	Properties           map[string]*ValidationRule `yaml:"properties,omitempty"`
	AdditionalProperties *bool                      `yaml:"additionalProperties,omitempty"`
	// Common
	Enum     []interface{} `yaml:"enum,omitempty"`
	Nullable bool          `yaml:"nullable,omitempty"`
}
