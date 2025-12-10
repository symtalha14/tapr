package stats

import (
	"time"

	"github.com/symtalha14/tapr/internal/request"
)

// BatchResult represents the result of testing a single endpoint in batch mode.
type BatchResult struct {
	Name           string         // Endpoint name
	URL            string         // Endpoint URL
	Method         string         // HTTP method
	Result         request.Result // The actual request result
	ExpectedStatus int            // What status code we expected
	Success        bool           // Whether the test passed
	Message        string         // Optional message (e.g., "Status mismatch")
}

// BatchSummary aggregates results from multiple endpoint tests.
type BatchSummary struct {
	Total      int           // Total endpoints tested
	Successful int           // Number of successful tests
	Failed     int           // Number of failed tests
	Slow       int           // Number of slow responses (> 500ms)
	TotalTime  time.Duration // Total time for all tests
	AvgLatency time.Duration // Average latency across all tests
	Results    []BatchResult // Individual results
}

// NewBatchSummary creates a new batch summary.
func NewBatchSummary() *BatchSummary {
	return &BatchSummary{
		Results: make([]BatchResult, 0),
	}
}

// AddResult adds a result to the summary and updates statistics.
func (bs *BatchSummary) AddResult(result BatchResult) {
	bs.Results = append(bs.Results, result)
	bs.Total++

	if result.Success {
		bs.Successful++
	} else {
		bs.Failed++
	}

	// Count slow responses
	if result.Result.Error == nil && result.Result.Latency > 500*time.Millisecond {
		bs.Slow++
	}

	// Update average latency
	if result.Result.Error == nil {
		bs.AvgLatency = (bs.AvgLatency*time.Duration(bs.Total-1) + result.Result.Latency) / time.Duration(bs.Total)
	}
}

// SuccessRate returns the success rate as a percentage.
func (bs *BatchSummary) SuccessRate() float64 {
	if bs.Total == 0 {
		return 0
	}
	return float64(bs.Successful) / float64(bs.Total) * 100
}
