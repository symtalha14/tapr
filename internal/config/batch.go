// Package config handles configuration file parsing and validation.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Endpoint represents a single API endpoint to test in batch mode.
type Endpoint struct {
	Name           string            `yaml:"name"`            // Friendly name for the endpoint
	URL            string            `yaml:"url"`             // Full URL to test
	Method         string            `yaml:"method"`          // HTTP method (GET, POST, etc.)
	Headers        map[string]string `yaml:"headers"`         // Optional headers for this endpoint
	Body           string            `yaml:"body"`            // Optional request body
	ExpectedStatus int               `yaml:"expected_status"` // Expected HTTP status code
	Timeout        time.Duration     `yaml:"timeout"`         // Optional timeout override
}

// BatchConfig represents the entire batch configuration file.
type BatchConfig struct {
	Endpoints   []Endpoint    `yaml:"endpoints"`   // List of endpoints to test
	Concurrency int           `yaml:"concurrency"` // Number of concurrent requests
	Timeout     time.Duration `yaml:"timeout"`     // Global timeout
}

// LoadBatchConfig reads and parses a batch configuration YAML file.
func LoadBatchConfig(filepath string) (*BatchConfig, error) {
	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil, fmt.Errorf("batch config file not found: %s", filepath)
	}

	// Read file contents
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read batch config: %w", err)
	}

	// Parse YAML
	var config BatchConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse batch config YAML: %w", err)
	}

	// Validate
	if len(config.Endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints defined in batch config")
	}

	// Set defaults
	for i := range config.Endpoints {
		endpoint := &config.Endpoints[i]

		// Default method to GET
		if endpoint.Method == "" {
			endpoint.Method = "GET"
		}

		// Default expected status to 200
		if endpoint.ExpectedStatus == 0 {
			endpoint.ExpectedStatus = 200
		}

		// Validate URL
		if endpoint.URL == "" {
			return nil, fmt.Errorf("endpoint '%s' has no URL", endpoint.Name)
		}
	}

	// Default concurrency
	if config.Concurrency == 0 {
		config.Concurrency = 5
	}

	// Default timeout
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	return &config, nil
}
