// Package output provides utilities for formatted terminal output,
// including JSON serialization for CI/CD integration.
package output

import (
	"encoding/json"

	"github.com/symtalha14/tapr/internal/stats"
)

// JSONBatchResult represents a batch result in JSON format.
type JSONBatchResult struct {
	Total       int            `json:"total"`
	Successful  int            `json:"successful"`
	Failed      int            `json:"failed"`
	Slow        int            `json:"slow"`
	SuccessRate float64        `json:"success_rate"`
	AvgLatency  int64          `json:"avg_latency_ms"`
	TotalTime   int64          `json:"total_time_ms"`
	Results     []JSONEndpoint `json:"results"`
}

// JSONEndpoint represents a single endpoint result in JSON format.
type JSONEndpoint struct {
	Name           string `json:"name"`
	URL            string `json:"url"`
	Method         string `json:"method"`
	Status         int    `json:"status"`
	ExpectedStatus int    `json:"expected_status"`
	Latency        int64  `json:"latency_ms"`
	Size           int64  `json:"size_bytes"`
	Success        bool   `json:"success"`
	Error          string `json:"error,omitempty"`
}

// FormatBatchResultJSON converts a batch summary to JSON format.
func FormatBatchResultJSON(summary *stats.BatchSummary) (string, error) {
	jsonResult := JSONBatchResult{
		Total:       summary.Total,
		Successful:  summary.Successful,
		Failed:      summary.Failed,
		Slow:        summary.Slow,
		SuccessRate: summary.SuccessRate(),
		AvgLatency:  summary.AvgLatency.Milliseconds(),
		TotalTime:   summary.TotalTime.Milliseconds(),
		Results:     make([]JSONEndpoint, len(summary.Results)),
	}

	for i, result := range summary.Results {
		endpoint := JSONEndpoint{
			Name:           result.Name,
			URL:            result.URL,
			Method:         result.Method,
			Status:         result.Result.StatusCode,
			ExpectedStatus: result.ExpectedStatus,
			Latency:        result.Result.Latency.Milliseconds(),
			Size:           result.Result.Size,
			Success:        result.Success,
		}

		if result.Result.Error != nil {
			endpoint.Error = result.Result.Error.Error()
		} else if !result.Success {
			endpoint.Error = result.Message
		}

		jsonResult.Results[i] = endpoint
	}

	data, err := json.MarshalIndent(jsonResult, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}
