// Package request provides HTTP client functionality for making API requests
// and measuring their performance characteristics.
package request

import (
	"net/http"
	"time"
)

// Result represents the outcome of an HTTP request, including timing
// information, response status, and any errors encountered.
type Result struct {
	URL        string        // The URL that was requested
	StatusCode int           // HTTP status code (e.g., 200, 404, 500)
	Status     string        // HTTP status text (e.g., "200 OK")
	Latency    time.Duration // Total time taken for the request
	Size       int64         // Response body size in bytes (-1 if unknown)
	Protocol   string        // HTTP protocol version (e.g., "HTTP/2.0")
	Error      error         // Any error that occurred during the request
}

// PingOptions contains configuration options for making HTTP requests.
type PingOptions struct {
	Method  string            // HTTP method (GET, POST, PUT, etc.)
	Timeout time.Duration     // Maximum time to wait for response
	Retries int               // Number of retry attempts on failure
	Headers map[string]string // HTTP headers to include in the request
}

// Ping makes an HTTP request to the specified URL and returns detailed
// timing and response information. It will retry the request if it fails,
// up to the number of times specified in options.Retries.
//
// Example:
//
//	opts := request.PingOptions{
//	    Method:  "GET",
//	    Timeout: 10 * time.Second,
//	    Retries: 3,
//	    Headers: map[string]string{
//	        "Authorization": "Bearer token123",
//	        "Content-Type": "application/json",
//	    },
//	}
//	result := request.Ping("https://api.example.com/health", opts)
func Ping(url string, opts PingOptions) Result {
	// Create HTTP client with custom timeout
	client := &http.Client{
		Timeout: opts.Timeout,
	}

	var lastResult Result
	maxAttempts := opts.Retries + 1 // Initial attempt + retries

	// Attempt the request, with retries if needed
	for attempt := 0; attempt < maxAttempts; attempt++ {
		lastResult = makeRequest(client, url, opts.Method, opts.Headers)

		// If successful, return immediately
		if lastResult.Error == nil {
			return lastResult
		}

		// If this wasn't the last attempt, wait before retrying
		if attempt < maxAttempts-1 {
			// Exponential backoff: 1s, 2s, 4s, 8s...
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(backoff)
		}
	}

	// Return the last result (which contains the error)
	return lastResult
}

// makeRequest performs a single HTTP request and measures its timing.
// This is an internal helper function used by Ping.
func makeRequest(client *http.Client, url, method string, headers map[string]string) Result {
	// Record the start time for latency measurement
	start := time.Now()

	// Create the HTTP request
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return Result{
			URL:     url,
			Latency: time.Since(start),
			Error:   err,
		}
	}

	// Add headers to the request
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute the request
	resp, err := client.Do(req)

	// Calculate total latency
	latency := time.Since(start)

	// Handle request errors (network issues, timeout, etc.)
	if err != nil {
		return Result{
			URL:     url,
			Latency: latency,
			Error:   err,
		}
	}

	// Always close the response body to prevent connection leaks
	// defer ensures this runs even if we return early
	defer resp.Body.Close()

	// Return successful result with all response metadata
	return Result{
		URL:        url,
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Latency:    latency,
		Size:       resp.ContentLength,
		Protocol:   resp.Proto,
		Error:      nil,
	}
}
