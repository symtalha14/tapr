package output

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/symtalha14/tapr/internal/request"
	"github.com/symtalha14/tapr/internal/stats"
)

func TestFormatBatchResultJSON(t *testing.T) {
	summary := stats.NewBatchSummary()

	// Add successful result
	summary.AddResult(stats.BatchResult{
		Name:           "Test API",
		URL:            "https://example.com",
		Method:         "GET",
		ExpectedStatus: 200,
		Success:        true,
		Result: request.Result{
			StatusCode: 200,
			Latency:    150 * time.Millisecond,
			Size:       1024,
		},
	})

	// Add failed result
	summary.AddResult(stats.BatchResult{
		Name:           "Broken API",
		URL:            "https://broken.com",
		Method:         "POST",
		ExpectedStatus: 200,
		Success:        false,
		Message:        "Expected 200, got 500",
		Result: request.Result{
			StatusCode: 500,
			Latency:    250 * time.Millisecond,
			Size:       512,
		},
	})

	summary.TotalTime = 500 * time.Millisecond

	jsonStr, err := FormatBatchResultJSON(summary)
	if err != nil {
		t.Fatalf("FormatBatchResultJSON() error = %v", err)
	}

	// Parse JSON to verify it's valid
	var result JSONBatchResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Verify fields
	if result.Total != 2 {
		t.Errorf("Total = %d, want 2", result.Total)
	}
	if result.Successful != 1 {
		t.Errorf("Successful = %d, want 1", result.Successful)
	}
	if result.Failed != 1 {
		t.Errorf("Failed = %d, want 1", result.Failed)
	}
	if result.SuccessRate != 50.0 {
		t.Errorf("SuccessRate = %v, want 50.0", result.SuccessRate)
	}
	if result.TotalTime != 500 {
		t.Errorf("TotalTime = %d, want 500", result.TotalTime)
	}

	// Verify results array
	if len(result.Results) != 2 {
		t.Fatalf("Results length = %d, want 2", len(result.Results))
	}

	// Check first result (successful)
	if result.Results[0].Name != "Test API" {
		t.Errorf("Results[0].Name = %s, want 'Test API'", result.Results[0].Name)
	}
	if result.Results[0].Success != true {
		t.Errorf("Results[0].Success = %v, want true", result.Results[0].Success)
	}
	if result.Results[0].Status != 200 {
		t.Errorf("Results[0].Status = %d, want 200", result.Results[0].Status)
	}

	// Check second result (failed)
	if result.Results[1].Name != "Broken API" {
		t.Errorf("Results[1].Name = %s, want 'Broken API'", result.Results[1].Name)
	}
	if result.Results[1].Success != false {
		t.Errorf("Results[1].Success = %v, want false", result.Results[1].Success)
	}
	if result.Results[1].Error != "Expected 200, got 500" {
		t.Errorf("Results[1].Error = %s, want 'Expected 200, got 500'", result.Results[1].Error)
	}
}

func TestFormatBatchResultJSON_Empty(t *testing.T) {
	summary := stats.NewBatchSummary()

	jsonStr, err := FormatBatchResultJSON(summary)
	if err != nil {
		t.Fatalf("FormatBatchResultJSON() error = %v", err)
	}

	var result JSONBatchResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if result.Total != 0 {
		t.Errorf("Total = %d, want 0", result.Total)
	}
	if result.SuccessRate != 0 {
		t.Errorf("SuccessRate = %v, want 0", result.SuccessRate)
	}
	if len(result.Results) != 0 {
		t.Errorf("Results length = %d, want 0", len(result.Results))
	}
}
