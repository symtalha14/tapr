package stats

import (
	"testing"
	"time"

	"github.com/symtalha14/tapr/internal/request"
)

func TestNewHistory(t *testing.T) {
	history := NewHistory(10)

	if history.Size() != 0 {
		t.Errorf("Size() = %d, want 0", history.Size())
	}
	if history.maxSize != 10 {
		t.Errorf("maxSize = %d, want 10", history.maxSize)
	}
}

func TestHistory_Add(t *testing.T) {
	history := NewHistory(5)

	result := request.Result{
		URL:        "https://example.com",
		StatusCode: 200,
		Latency:    100 * time.Millisecond,
	}

	history.Add(result)

	if history.Size() != 1 {
		t.Errorf("Size() = %d, want 1", history.Size())
	}
}

func TestHistory_RollingWindow(t *testing.T) {
	history := NewHistory(3) // Max 3 entries

	// Add 5 results
	for i := 1; i <= 5; i++ {
		result := request.Result{
			URL:        "https://example.com",
			StatusCode: 200 + i,
			Latency:    time.Duration(i*100) * time.Millisecond,
		}
		history.Add(result)
	}

	// Should only keep last 3
	if history.Size() != 3 {
		t.Errorf("Size() = %d, want 3", history.Size())
	}

	// Check we have the latest entries (3, 4, 5)
	recent := history.GetRecent(3)
	if len(recent) != 3 {
		t.Fatalf("GetRecent() returned %d entries, want 3", len(recent))
	}

	// Latest entries should have status codes 203, 204, 205
	expectedCodes := []int{203, 204, 205}
	for i, entry := range recent {
		if entry.Result.StatusCode != expectedCodes[i] {
			t.Errorf("Entry %d: StatusCode = %d, want %d", i, entry.Result.StatusCode, expectedCodes[i])
		}
	}
}

func TestHistory_GetRecent(t *testing.T) {
	history := NewHistory(10)

	// Add 5 results
	for i := 1; i <= 5; i++ {
		result := request.Result{
			URL:        "https://example.com",
			StatusCode: 200,
			Latency:    time.Duration(i*100) * time.Millisecond,
		}
		history.Add(result)
	}

	tests := []struct {
		name string
		n    int
		want int
	}{
		{"get 3 recent", 3, 3},
		{"get all", 5, 5},
		{"get more than available", 10, 5},
		{"get zero", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recent := history.GetRecent(tt.n)
			if len(recent) != tt.want {
				t.Errorf("GetRecent(%d) returned %d entries, want %d", tt.n, len(recent), tt.want)
			}
		})
	}
}

func TestHistory_GetRecent_Order(t *testing.T) {
	history := NewHistory(10)

	// Add results with increasing latencies
	for i := 1; i <= 3; i++ {
		result := request.Result{
			URL:        "https://example.com",
			StatusCode: 200,
			Latency:    time.Duration(i*100) * time.Millisecond,
		}
		history.Add(result)
	}

	recent := history.GetRecent(3)

	// Should be in order: oldest to newest
	expectedLatencies := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		300 * time.Millisecond,
	}

	for i, entry := range recent {
		if entry.Result.Latency != expectedLatencies[i] {
			t.Errorf("Entry %d: Latency = %v, want %v", i, entry.Result.Latency, expectedLatencies[i])
		}
	}
}
