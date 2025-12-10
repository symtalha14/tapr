package stats

import (
	"time"

	"github.com/symtalha14/tapr/internal/request"
)

// HistoryEntry represents a single request in the history.
type HistoryEntry struct {
	Timestamp time.Time      // When the request was made
	Result    request.Result // The request result
}

// History keeps a rolling window of recent requests.
type History struct {
	entries []HistoryEntry
	maxSize int
}

// NewHistory creates a new history tracker with a maximum size.
func NewHistory(maxSize int) *History {
	return &History{
		entries: make([]HistoryEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add records a new request result in the history.
// If history is full, removes the oldest entry.
func (h *History) Add(result request.Result) {
	entry := HistoryEntry{
		Timestamp: time.Now(),
		Result:    result,
	}

	h.entries = append(h.entries, entry)

	// Keep only the last maxSize entries
	if len(h.entries) > h.maxSize {
		h.entries = h.entries[1:] // Remove first element
	}
}

// GetRecent returns the N most recent entries.
func (h *History) GetRecent(n int) []HistoryEntry {
	if n > len(h.entries) {
		n = len(h.entries)
	}

	// Return last N entries
	start := len(h.entries) - n
	return h.entries[start:]
}

// Size returns the current number of entries in history.
func (h *History) Size() int {
	return len(h.entries)
}
