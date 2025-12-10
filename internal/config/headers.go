// Package config handles configuration file parsing and validation.
package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Headers represents HTTP headers loaded from a YAML file.
// The structure is a simple map of header name to header value.
//
// Example YAML format:
//
//	Authorization: Bearer token123
//	Content-Type: application/json
//	X-API-Key: abc123
type Headers map[string]string

// LoadHeaders reads and parses a YAML file containing HTTP headers.
// Returns a map of header names to values, or an error if the file
// cannot be read or parsed.
//
// Example usage:
//
//	headers, err := config.LoadHeaders("headers.yml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for key, value := range headers {
//	    req.Header.Set(key, value)
//	}
func LoadHeaders(filepath string) (Headers, error) {
	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil, fmt.Errorf("headers file not found: %s", filepath)
	}

	// Read file contents
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read headers file: %w", err)
	}

	// If empty file, return empty headers (not an error)
	if len(data) == 0 {
		return make(Headers), nil // ← Changed: Return empty map, no error
	}

	// Parse YAML
	var headers Headers
	if err := yaml.Unmarshal(data, &headers); err != nil {
		return nil, fmt.Errorf("failed to parse headers YAML: %w", err)
	}

	return headers, nil
}

// ParseInlineHeaders converts a slice of "Key: Value" strings into a Headers map.
// Each string must be in the format "Key: Value" with a colon separator.
// Returns an error if any header is malformed.
//
// Example:
//
//	headers, err := config.ParseInlineHeaders([]string{
//	    "Authorization: Bearer token123",
//	    "Content-Type: application/json",
//	})
func ParseInlineHeaders(headerStrings []string) (Headers, error) {
	headers := make(Headers)

	for _, headerStr := range headerStrings {
		// Split on the first colon
		parts := strings.SplitN(headerStr, ":", 2)

		// Validate format
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header format: '%s' (expected 'Key: Value')", headerStr)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Validate that key is not empty
		if key == "" {
			return nil, fmt.Errorf("empty header key in: '%s'", headerStr)
		}

		headers[key] = value
	}

	return headers, nil
}

// MergeHeaders combines multiple header maps into one.
// If the same key exists in multiple maps, the last one wins.
// This is useful for combining file-based headers with inline headers.
//
// Example:
//
//	fileHeaders := Headers{"Authorization": "Bearer old"}
//	inlineHeaders := Headers{"Authorization": "Bearer new", "X-Custom": "value"}
//	merged := MergeHeaders(fileHeaders, inlineHeaders)
//	// Result: {"Authorization": "Bearer new", "X-Custom": "value"}
func MergeHeaders(headerMaps ...Headers) Headers {
	result := make(Headers)

	// Iterate through each map and add/overwrite keys
	for _, headers := range headerMaps {
		for key, value := range headers {
			result[key] = value
		}
	}

	return result
}
