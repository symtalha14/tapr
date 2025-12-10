package stats

import (
	"testing"
	"time"
)

func TestNewTracker(t *testing.T) {
	tracker := NewTracker()

	if tracker.Total != 0 {
		t.Errorf("Total = %d, want 0", tracker.Total)
	}
	if tracker.Successful != 0 {
		t.Errorf("Successful = %d, want 0", tracker.Successful)
	}
	if tracker.Failed != 0 {
		t.Errorf("Failed = %d, want 0", tracker.Failed)
	}
	if len(tracker.Latencies) != 0 {
		t.Errorf("Latencies = %v, want empty slice", tracker.Latencies)
	}
}

func TestTracker_Record(t *testing.T) {
	tracker := NewTracker()

	// Record successful request
	tracker.Record(100*time.Millisecond, true)

	if tracker.Total != 1 {
		t.Errorf("Total = %d, want 1", tracker.Total)
	}
	if tracker.Successful != 1 {
		t.Errorf("Successful = %d, want 1", tracker.Successful)
	}
	if tracker.Failed != 0 {
		t.Errorf("Failed = %d, want 0", tracker.Failed)
	}
	if tracker.MinLatency != 100*time.Millisecond {
		t.Errorf("MinLatency = %v, want 100ms", tracker.MinLatency)
	}
	if tracker.MaxLatency != 100*time.Millisecond {
		t.Errorf("MaxLatency = %v, want 100ms", tracker.MaxLatency)
	}

	// Record failed request
	tracker.Record(200*time.Millisecond, false)

	if tracker.Total != 2 {
		t.Errorf("Total = %d, want 2", tracker.Total)
	}
	if tracker.Successful != 1 {
		t.Errorf("Successful = %d, want 1", tracker.Successful)
	}
	if tracker.Failed != 1 {
		t.Errorf("Failed = %d, want 1", tracker.Failed)
	}
	if tracker.MaxLatency != 200*time.Millisecond {
		t.Errorf("MaxLatency = %v, want 200ms", tracker.MaxLatency)
	}
}

func TestTracker_AvgLatency(t *testing.T) {
	tests := []struct {
		name      string
		latencies []time.Duration
		want      time.Duration
	}{
		{
			name:      "single latency",
			latencies: []time.Duration{100 * time.Millisecond},
			want:      100 * time.Millisecond,
		},
		{
			name:      "multiple latencies",
			latencies: []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 300 * time.Millisecond},
			want:      200 * time.Millisecond,
		},
		{
			name:      "empty",
			latencies: []time.Duration{},
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTracker()
			for _, latency := range tt.latencies {
				tracker.Record(latency, true)
			}

			got := tracker.AvgLatency()
			if got != tt.want {
				t.Errorf("AvgLatency() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTracker_Percentile(t *testing.T) {
	tracker := NewTracker()

	// Add 100 requests from 1ms to 100ms
	for i := 1; i <= 100; i++ {
		tracker.Record(time.Duration(i)*time.Millisecond, true)
	}

	tests := []struct {
		name       string
		percentile float64
		want       time.Duration
	}{
		{"P50", 0.50, 50 * time.Millisecond},
		{"P95", 0.95, 95 * time.Millisecond},
		{"P99", 0.99, 99 * time.Millisecond},
		{"P100", 1.00, 100 * time.Millisecond},
		{"P0", 0.00, 1 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tracker.Percentile(tt.percentile)
			if got != tt.want {
				t.Errorf("Percentile(%v) = %v, want %v", tt.percentile, got, tt.want)
			}
		})
	}
}

func TestTracker_Percentile_Empty(t *testing.T) {
	tracker := NewTracker()
	got := tracker.Percentile(0.95)
	if got != 0 {
		t.Errorf("Percentile() on empty tracker = %v, want 0", got)
	}
}

func TestTracker_SuccessRate(t *testing.T) {
	tests := []struct {
		name       string
		successful int
		failed     int
		want       float64
	}{
		{"all successful", 10, 0, 100.0},
		{"all failed", 0, 10, 0.0},
		{"half and half", 5, 5, 50.0},
		{"mostly successful", 9, 1, 90.0},
		{"empty", 0, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTracker()

			for i := 0; i < tt.successful; i++ {
				tracker.Record(100*time.Millisecond, true)
			}
			for i := 0; i < tt.failed; i++ {
				tracker.Record(100*time.Millisecond, false)
			}

			got := tracker.SuccessRate()
			if got != tt.want {
				t.Errorf("SuccessRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTracker_MinMax(t *testing.T) {
	tracker := NewTracker()

	latencies := []time.Duration{
		500 * time.Millisecond,
		100 * time.Millisecond,
		300 * time.Millisecond,
		50 * time.Millisecond,
		1000 * time.Millisecond,
	}

	for _, latency := range latencies {
		tracker.Record(latency, true)
	}

	if tracker.MinLatency != 50*time.Millisecond {
		t.Errorf("MinLatency = %v, want 50ms", tracker.MinLatency)
	}
	if tracker.MaxLatency != 1000*time.Millisecond {
		t.Errorf("MaxLatency = %v, want 1000ms", tracker.MaxLatency)
	}
}
