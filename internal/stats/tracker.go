// Package stats provides utilities for tracking and calculating
// request statistics over time.
package stats

import (
	"sort"
	"time"
)

// Tracker keeps track of request statistics for watch mode.
type Tracker struct {
	Total      int             // Total number of requests
	Successful int             // Number of successful requests
	Failed     int             // Number of failed requests
	Latencies  []time.Duration // All latency measurements
	MinLatency time.Duration   // Minimum latency observed
	MaxLatency time.Duration   // Maximum latency observed
}

// NewTracker creates a new statistics tracker.
func NewTracker() *Tracker {
	return &Tracker{
		Latencies: make([]time.Duration, 0),
	}
}

// Record adds a new request result to the tracker.
func (t *Tracker) Record(latency time.Duration, success bool) {
	t.Total++

	if success {
		t.Successful++
	} else {
		t.Failed++
	}

	// Record latency
	t.Latencies = append(t.Latencies, latency)

	// Update min/max
	if t.MinLatency == 0 || latency < t.MinLatency {
		t.MinLatency = latency
	}
	if latency > t.MaxLatency {
		t.MaxLatency = latency
	}
}

// AvgLatency calculates the average latency.
func (t *Tracker) AvgLatency() time.Duration {
	if len(t.Latencies) == 0 {
		return 0
	}

	var total time.Duration
	for _, latency := range t.Latencies {
		total += latency
	}

	return total / time.Duration(len(t.Latencies))
}

// Percentile calculates the Nth percentile of latencies.
// For example, P95 means 95% of requests were faster than this value.
func (t *Tracker) Percentile(p float64) time.Duration {
	if len(t.Latencies) == 0 {
		return 0
	}

	// Sort latencies
	sorted := make([]time.Duration, len(t.Latencies))
	copy(sorted, t.Latencies)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// Calculate index for percentile (0-based indexing)
	index := int(float64(len(sorted))*p) - 1

	// Handle edge cases
	if index < 0 {
		index = 0
	}
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}

// SuccessRate returns the success rate as a percentage.
func (t *Tracker) SuccessRate() float64 {
	if t.Total == 0 {
		return 0
	}
	return float64(t.Successful) / float64(t.Total) * 100
}
