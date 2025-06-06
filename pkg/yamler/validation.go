package yamler

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
)

// LoadSchemaFromString loads a validation schema from a YAML string
func LoadSchemaFromString(content string) (*ValidationRule, error) {
	var schema ValidationRule
	if err := yaml.Unmarshal([]byte(content), &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema: %v", err)
	}
	return &schema, nil
}

// LoadSchemaFromFile loads a validation schema from a YAML file
func LoadSchemaFromFile(filename string) (*ValidationRule, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %v", err)
	}
	return LoadSchemaFromString(string(content))
}

// Validate validates the YAML document against a schema
func (d *Document) Validate(schema *ValidationRule) error {
	if schema == nil {
		return fmt.Errorf("schema is nil")
	}

	return validateNode(d.root.Content[0], schema, "")
}

// validateNode validates a YAML node against a schema
func validateNode(node *yaml.Node, schema *ValidationRule, path string) error {
	if schema.Nullable && node.Tag == "!!null" {
		return nil
	}

	switch schema.Type {
	case TypeString:
		return validateString(node, schema, path)
	case TypeInt:
		return validateInt(node, schema, path)
	case TypeFloat:
		return validateFloat(node, schema, path)
	case TypeBool:
		return validateBool(node, schema, path)
	case TypeArray:
		return validateArray(node, schema, path)
	case TypeMap:
		return validateMap(node, schema, path)
	case TypeAny:
		return nil
	default:
		return fmt.Errorf("path %s: unsupported type: %s", path, schema.Type)
	}
}

// validateString validates a string value
func validateString(node *yaml.Node, schema *ValidationRule, path string) error {
	if node.Kind != yaml.ScalarNode || node.Tag != "!!str" {
		return fmt.Errorf("path %s: expected string, got %s", path, node.Tag)
	}

	value := node.Value

	if schema.MinLength != nil && len(value) < *schema.MinLength {
		return fmt.Errorf("path %s: string length %d is less than minimum %d", path, len(value), *schema.MinLength)
	}

	if schema.MaxLength != nil && len(value) > *schema.MaxLength {
		return fmt.Errorf("path %s: string length %d is greater than maximum %d", path, len(value), *schema.MaxLength)
	}

	if schema.Pattern != nil {
		re, err := regexp.Compile(*schema.Pattern)
		if err != nil {
			return fmt.Errorf("path %s: invalid pattern: %v", path, err)
		}
		if !re.MatchString(value) {
			return fmt.Errorf("path %s: string does not match pattern %s", path, *schema.Pattern)
		}
	}

	if schema.Enum != nil {
		valid := false
		for _, enum := range schema.Enum {
			if str, ok := enum.(string); ok && str == value {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("path %s: value %s is not in enum", path, value)
		}
	}

	return nil
}

// validateInt validates an integer value
func validateInt(node *yaml.Node, schema *ValidationRule, path string) error {
	if node.Kind != yaml.ScalarNode || node.Tag != "!!int" {
		return fmt.Errorf("path %s: expected integer, got %s", path, node.Tag)
	}

	value, err := strconv.ParseInt(node.Value, 10, 64)
	if err != nil {
		return fmt.Errorf("path %s: invalid integer: %v", path, err)
	}

	return validateNumericConstraints(float64(value), schema, path, "integer")
}

// validateFloat validates a float value
func validateFloat(node *yaml.Node, schema *ValidationRule, path string) error {
	if node.Kind != yaml.ScalarNode || (node.Tag != "!!float" && node.Tag != "!!int") {
		return fmt.Errorf("path %s: expected float, got %s", path, node.Tag)
	}

	value, err := strconv.ParseFloat(node.Value, 64)
	if err != nil {
		return fmt.Errorf("path %s: invalid float: %v", path, err)
	}

	return validateNumericConstraints(value, schema, path, "float")
}

// validateNumericConstraints validates numeric constraints for both int and float
func validateNumericConstraints(value float64, schema *ValidationRule, path, valueType string) error {
	if err := validateMinimumConstraints(value, schema, path, valueType); err != nil {
		return err
	}

	if err := validateMaximumConstraints(value, schema, path, valueType); err != nil {
		return err
	}

	return validateNumericEnum(value, schema, path, valueType)
}

// validateMinimumConstraints validates minimum and exclusive minimum constraints
func validateMinimumConstraints(value float64, schema *ValidationRule, path, valueType string) error {
	if schema.Minimum != nil && value < *schema.Minimum {
		if valueType == "integer" {
			return fmt.Errorf("path %s: value %d is less than minimum %f", path, int64(value), *schema.Minimum)
		}
		return fmt.Errorf("path %s: value %f is less than minimum %f", path, value, *schema.Minimum)
	}

	if schema.ExclusiveMinimum != nil && value <= *schema.ExclusiveMinimum {
		if valueType == "integer" {
			return fmt.Errorf("path %s: value %d is not greater than exclusive minimum %f", path, int64(value), *schema.ExclusiveMinimum)
		}
		return fmt.Errorf("path %s: value %f is not greater than exclusive minimum %f", path, value, *schema.ExclusiveMinimum)
	}

	return nil
}

// validateMaximumConstraints validates maximum and exclusive maximum constraints
func validateMaximumConstraints(value float64, schema *ValidationRule, path, valueType string) error {
	if schema.Maximum != nil && value > *schema.Maximum {
		if valueType == "integer" {
			return fmt.Errorf("path %s: value %d is greater than maximum %f", path, int64(value), *schema.Maximum)
		}
		return fmt.Errorf("path %s: value %f is greater than maximum %f", path, value, *schema.Maximum)
	}

	if schema.ExclusiveMaximum != nil && value >= *schema.ExclusiveMaximum {
		if valueType == "integer" {
			return fmt.Errorf("path %s: value %d is not less than exclusive maximum %f", path, int64(value), *schema.ExclusiveMaximum)
		}
		return fmt.Errorf("path %s: value %f is not less than exclusive maximum %f", path, value, *schema.ExclusiveMaximum)
	}

	return nil
}

// validateNumericEnum validates enum constraints for numeric values
func validateNumericEnum(value float64, schema *ValidationRule, path, valueType string) error {
	if schema.Enum == nil {
		return nil
	}

	for _, enum := range schema.Enum {
		switch valueType {
		case "integer":
			if num, ok := enum.(int64); ok && num == int64(value) {
				return nil
			}
		case "float":
			if num, ok := enum.(float64); ok && num == value {
				return nil
			}
		}
	}

	if valueType == "integer" {
		return fmt.Errorf("path %s: value %d is not in enum", path, int64(value))
	}
	return fmt.Errorf("path %s: value %f is not in enum", path, value)
}

// validateBool validates a boolean value
func validateBool(node *yaml.Node, schema *ValidationRule, path string) error {
	if node.Kind != yaml.ScalarNode || node.Tag != "!!bool" {
		return fmt.Errorf("path %s: expected boolean, got %s", path, node.Tag)
	}

	value, err := strconv.ParseBool(node.Value)
	if err != nil {
		return fmt.Errorf("path %s: invalid boolean: %v", path, err)
	}

	if schema.Enum != nil {
		valid := false
		for _, enum := range schema.Enum {
			if b, ok := enum.(bool); ok && b == value {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("path %s: value %t is not in enum", path, value)
		}
	}

	return nil
}

// validateArray validates an array value
func validateArray(node *yaml.Node, schema *ValidationRule, path string) error {
	if node.Kind != yaml.SequenceNode {
		return fmt.Errorf("path %s: expected array, got %s", path, node.Tag)
	}

	if schema.MinItems != nil && len(node.Content) < *schema.MinItems {
		return fmt.Errorf("path %s: array length %d is less than minimum %d", path, len(node.Content), *schema.MinItems)
	}

	if schema.MaxItems != nil && len(node.Content) > *schema.MaxItems {
		return fmt.Errorf("path %s: array length %d is greater than maximum %d", path, len(node.Content), *schema.MaxItems)
	}

	if schema.UniqueItems {
		seen := make(map[string]bool)
		for i, item := range node.Content {
			key := fmt.Sprintf("%v", item.Value)
			if seen[key] {
				return fmt.Errorf("path %s: duplicate item at index %d", path, i)
			}
			seen[key] = true
		}
	}

	if schema.Items != nil {
		for i, item := range node.Content {
			itemPath := fmt.Sprintf("%s[%d]", path, i)
			if err := validateNode(item, schema.Items, itemPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateMap validates a map value
func validateMap(node *yaml.Node, schema *ValidationRule, path string) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("path %s: expected map, got %s", path, node.Tag)
	}

	// Check required fields
	if schema.Required != nil {
		for _, required := range schema.Required {
			found := false
			for i := 0; i < len(node.Content); i += 2 {
				if node.Content[i].Value == required {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("path %s: required field %s is missing", path, required)
			}
		}
	}

	// Validate properties
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		value := node.Content[i+1]
		fieldPath := path
		if fieldPath == "" {
			fieldPath = key
		} else {
			fieldPath = fmt.Sprintf("%s.%s", path, key)
		}

		// Check if property is defined in schema
		if propertySchema, ok := schema.Properties[key]; ok {
			if err := validateNode(value, propertySchema, fieldPath); err != nil {
				return err
			}
		} else if schema.AdditionalProperties != nil && !*schema.AdditionalProperties {
			return fmt.Errorf("path %s: additional property %s is not allowed", path, key)
		}
	}

	return nil
}
