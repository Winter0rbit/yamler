package yamler

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// GetAll returns all values that match the wildcard pattern
// Supported patterns:
//   - config.*.name - matches any key at that level
//   - config.**.name - matches any nested key (recursive)
//   - config.db.* - matches all keys under config.db
func (d *Document) GetAll(pattern string) (map[string]interface{}, error) {
	root, err := d.mappingRoot()
	if err != nil {
		return nil, err
	}

	results := make(map[string]interface{})
	err = findMatchingPaths(root, pattern, "", results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// SetAll sets a value for all paths that match the wildcard pattern
// Note: This only works with existing paths, it won't create new ones
func (d *Document) SetAll(pattern string, value interface{}) error {
	// First, get all matching paths
	matches, err := d.GetAll(pattern)
	if err != nil {
		return err
	}

	// Set value for each matching path
	for path := range matches {
		err = d.Set(path, value)
		if err != nil {
			return fmt.Errorf("failed to set value at path %s: %w", path, err)
		}
	}

	return nil
}

// GetKeys returns all keys that match the wildcard pattern (without values)
func (d *Document) GetKeys(pattern string) ([]string, error) {
	matches, err := d.GetAll(pattern)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(matches))
	for key := range matches {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys, nil
}

// findMatchingPaths recursively finds paths that match the pattern
func findMatchingPaths(node *yaml.Node, pattern, currentPath string, results map[string]interface{}) error {
	if node == nil {
		return nil
	}

	// Check if current path matches the pattern
	if pathMatches(currentPath, pattern) {
		value, err := nodeToInterface(node)
		if err != nil {
			return err
		}
		results[currentPath] = value
		return nil
	}

	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i].Value
			childNode := node.Content[i+1]

			var childPath string
			if currentPath == "" {
				childPath = key
			} else {
				childPath = currentPath + "." + key
			}

			// Check if we should continue exploring this path
			if couldMatch(childPath, pattern) {
				err := findMatchingPaths(childNode, pattern, childPath, results)
				if err != nil {
					return err
				}
			}
		}

	case yaml.SequenceNode:
		for idx, childNode := range node.Content {
			childPath := fmt.Sprintf("%s[%d]", currentPath, idx)

			// Check if we should continue exploring this path
			if couldMatch(childPath, pattern) {
				err := findMatchingPaths(childNode, pattern, childPath, results)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// pathMatches checks if a path matches a wildcard pattern
func pathMatches(path, pattern string) bool {
	if pattern == path {
		return true
	}

	// Convert wildcard pattern to regex
	regexPattern := wildcardToRegex(pattern)
	matched, err := regexp.MatchString(regexPattern, path)
	if err != nil {
		return false
	}

	return matched
}

// couldMatch checks if a path could potentially match the pattern
// (used for optimization to avoid exploring irrelevant branches)
func couldMatch(path, pattern string) bool {
	// If path is longer than pattern and pattern doesn't have ** or ends with *, probably won't match
	pathParts := splitPath(path)
	patternParts := strings.Split(pattern, ".")

	// Handle ** (recursive wildcard) - always could match
	for _, part := range patternParts {
		if part == "**" {
			return true
		}
	}

	// If path has more parts than pattern and no ** or ending *, won't match
	if len(pathParts) > len(patternParts) {
		lastPatternPart := patternParts[len(patternParts)-1]
		if lastPatternPart != "*" && lastPatternPart != "**" {
			return false
		}
	}

	// Check if current path could lead to a match
	for i := 0; i < len(pathParts) && i < len(patternParts); i++ {
		pathPart := pathParts[i]
		patternPart := patternParts[i]

		if patternPart == "*" || patternPart == "**" {
			continue // Wildcard matches anything
		}

		// Handle array indices in pattern
		if strings.Contains(patternPart, "[") {
			patternBase := strings.Split(patternPart, "[")[0]
			pathBase := strings.Split(pathPart, "[")[0]
			if patternBase != pathBase {
				return false
			}
			continue
		}

		// Handle array indices in path vs pattern
		if strings.Contains(pathPart, "[") && !strings.Contains(patternPart, "[") {
			pathBase := strings.Split(pathPart, "[")[0]
			if pathBase != patternPart {
				return false
			}
			continue
		}

		if pathPart != patternPart {
			return false
		}
	}

	return true
}

// wildcardToRegex converts a wildcard pattern to a regex pattern
func wildcardToRegex(pattern string) string {
	// Escape special regex characters except * and **
	escaped := regexp.QuoteMeta(pattern)

	// Replace escaped wildcard patterns with regex equivalents
	escaped = strings.ReplaceAll(escaped, `\*\*`, `.*`) // ** matches any path

	// Handle [*] for array index wildcards
	escaped = strings.ReplaceAll(escaped, `\[\*\]`, `\[[0-9]+\]`) // [*] matches any array index

	escaped = strings.ReplaceAll(escaped, `\*`, `[^.\[\]]*`) // * matches any single path segment (but not array indices)

	// Handle array indices - allow them to match * wildcard
	// This makes servers.* match servers[0], servers[1], etc.
	escaped = regexp.MustCompile(`([^\\])\\\*`).ReplaceAllString(escaped, `$1[^.]*(\[[0-9]+\])?`)
	// Handle start of string
	escaped = regexp.MustCompile(`^\\\*`).ReplaceAllString(escaped, `[^.]*(\[[0-9]+\])?`)

	// Handle array indices
	escaped = strings.ReplaceAll(escaped, `\[`, `\[`)
	escaped = strings.ReplaceAll(escaped, `\]`, `\]`)

	// Anchor the pattern
	return "^" + escaped + "$"
}

// FilterByPattern returns only the items from a map that match the pattern
func FilterByPattern(data map[string]interface{}, pattern string) map[string]interface{} {
	filtered := make(map[string]interface{})

	for path, value := range data {
		if pathMatches(path, pattern) {
			filtered[path] = value
		}
	}

	return filtered
}

// GetPathsRecursive returns all paths in the document recursively
func (d *Document) GetPathsRecursive() ([]string, error) {
	root, err := d.mappingRoot()
	if err != nil {
		return nil, err
	}

	var paths []string
	err = collectPaths(root, "", &paths)
	if err != nil {
		return nil, err
	}

	sort.Strings(paths)
	return paths, nil
}

// collectPaths recursively collects all paths in the node
func collectPaths(node *yaml.Node, currentPath string, paths *[]string) error {
	if node == nil {
		return nil
	}

	// Add current path if it's not empty
	if currentPath != "" {
		*paths = append(*paths, currentPath)
	}

	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i].Value
			childNode := node.Content[i+1]

			var childPath string
			if currentPath == "" {
				childPath = key
			} else {
				childPath = currentPath + "." + key
			}

			err := collectPaths(childNode, childPath, paths)
			if err != nil {
				return err
			}
		}

	case yaml.SequenceNode:
		for idx, childNode := range node.Content {
			childPath := fmt.Sprintf("%s[%d]", currentPath, idx)
			err := collectPaths(childNode, childPath, paths)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
