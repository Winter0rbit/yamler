package yamler

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Get returns a value from the YAML document by its path
func (d *Document) Get(path string) (interface{}, error) {
	root, err := d.mappingRoot()
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nodeToInterface(root)
	}

	parts := strings.Split(path, ".")
	node := root
	for _, part := range parts {
		// Check if part is an array index
		if strings.HasSuffix(part, "]") {
			// Extract array name and index
			idx := strings.LastIndex(part, "[")
			if idx == -1 {
				return nil, fmt.Errorf("path %s: invalid array index format", path)
			}
			arrayName := part[:idx]
			indexStr := part[idx+1 : len(part)-1]
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, fmt.Errorf("path %s: invalid array index: %s", path, indexStr)
			}

			// Get array node
			if node.Kind != yaml.MappingNode {
				return nil, fmt.Errorf("path %s: expected mapping node", path)
			}
			found := false
			for i := 0; i < len(node.Content); i += 2 {
				if node.Content[i].Value == arrayName {
					arrayNode := node.Content[i+1]
					if arrayNode.Kind != yaml.SequenceNode {
						return nil, fmt.Errorf("path %s: expected sequence node", path)
					}
					if index < 0 || index >= len(arrayNode.Content) {
						return nil, fmt.Errorf("path %s: array index out of bounds", path)
					}
					node = arrayNode.Content[index]
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("path %s: key %s not found", path, arrayName)
			}
		} else {
			// Handle regular key
			if node.Kind != yaml.MappingNode {
				return nil, fmt.Errorf("path %s: expected mapping node", path)
			}
			found := false
			for i := 0; i < len(node.Content); i += 2 {
				if node.Content[i].Value == part {
					node = node.Content[i+1]
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("path %s: key %s not found", path, part)
			}
		}
	}
	return nodeToInterface(node)
}

// GetString returns a string value from the YAML document
func (d *Document) GetString(path string) (string, error) {
	value, err := d.Get(path)
	if err != nil {
		return "", err
	}

	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("path %s: expected string, got %T", path, value)
	}

	return str, nil
}

// GetInt returns an integer value from the YAML document
func (d *Document) GetInt(path string) (int64, error) {
	value, err := d.Get(path)
	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case int64:
		return v, nil
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("path %s: invalid integer value: %v", path, err)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("path %s: expected integer, got %T", path, value)
	}
}

// GetFloat returns a float value from the YAML document
func (d *Document) GetFloat(path string) (float64, error) {
	value, err := d.Get(path)
	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case int64:
		return float64(v), nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("path %s: invalid float value: %v", path, err)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("path %s: expected float, got %T", path, value)
	}
}

// GetBool returns a boolean value from the YAML document
func (d *Document) GetBool(path string) (bool, error) {
	value, err := d.Get(path)
	if err != nil {
		return false, err
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		v = strings.ToLower(v)
		switch v {
		case "true", "yes", "1", "on":
			return true, nil
		case "false", "no", "0", "off":
			return false, nil
		default:
			return false, fmt.Errorf("path %s: invalid boolean value: %s", path, v)
		}
	default:
		return false, fmt.Errorf("path %s: expected boolean, got %T", path, value)
	}
}

// GetSlice returns a slice value from the YAML document
func (d *Document) GetSlice(path string) ([]interface{}, error) {
	value, err := d.Get(path)
	if err != nil {
		return nil, err
	}

	slice, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("path %s: expected slice, got %T", path, value)
	}

	return slice, nil
}

// GetMap returns a map value from the YAML document
func (d *Document) GetMap(path string) (map[string]interface{}, error) {
	value, err := d.Get(path)
	if err != nil {
		return nil, err
	}

	m, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("path %s: expected map, got %T", path, value)
	}

	return m, nil
}

// GetStringSlice returns a string slice from the YAML document
func (d *Document) GetStringSlice(path string) ([]string, error) {
	slice, err := d.GetSlice(path)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(slice))
	for i, v := range slice {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("path %s: element %d is not a string", path, i)
		}
		result[i] = str
	}

	return result, nil
}

// GetIntSlice returns an integer slice from the YAML document
func (d *Document) GetIntSlice(path string) ([]int64, error) {
	slice, err := d.GetSlice(path)
	if err != nil {
		return nil, err
	}

	result := make([]int64, len(slice))
	for i, v := range slice {
		switch val := v.(type) {
		case int64:
			result[i] = val
		case string:
			n, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("path %s: element %d is not a valid integer", path, i)
			}
			result[i] = n
		default:
			return nil, fmt.Errorf("path %s: element %d is not an integer", path, i)
		}
	}

	return result, nil
}

// GetFloatSlice returns a float slice from the YAML document
func (d *Document) GetFloatSlice(path string) ([]float64, error) {
	slice, err := d.GetSlice(path)
	if err != nil {
		return nil, err
	}

	result := make([]float64, len(slice))
	for i, v := range slice {
		switch val := v.(type) {
		case float64:
			result[i] = val
		case int64:
			result[i] = float64(val)
		case string:
			n, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return nil, fmt.Errorf("path %s: element %d is not a valid float", path, i)
			}
			result[i] = n
		default:
			return nil, fmt.Errorf("path %s: element %d is not a float", path, i)
		}
	}

	return result, nil
}

// GetBoolSlice returns a boolean slice from the YAML document
func (d *Document) GetBoolSlice(path string) ([]bool, error) {
	slice, err := d.GetSlice(path)
	if err != nil {
		return nil, err
	}

	result := make([]bool, len(slice))
	for i, v := range slice {
		switch val := v.(type) {
		case bool:
			result[i] = val
		case string:
			val = strings.ToLower(val)
			switch val {
			case "true", "yes", "1", "on":
				result[i] = true
			case "false", "no", "0", "off":
				result[i] = false
			default:
				return nil, fmt.Errorf("path %s: element %d is not a valid boolean", path, i)
			}
		default:
			return nil, fmt.Errorf("path %s: element %d is not a boolean", path, i)
		}
	}

	return result, nil
}

// GetMapSlice returns a slice of maps from the YAML document
func (d *Document) GetMapSlice(path string) ([]map[string]interface{}, error) {
	slice, err := d.GetSlice(path)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(slice))
	for i, v := range slice {
		m, ok := v.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("path %s: element %d is not a map", path, i)
		}
		result[i] = m
	}

	return result, nil
}
