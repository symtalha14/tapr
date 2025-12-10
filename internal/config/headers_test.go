package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseInlineHeaders(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "single header",
			input: []string{"Authorization: Bearer token"},
			want:  map[string]string{"Authorization": "Bearer token"},
		},
		{
			name:  "multiple headers",
			input: []string{"Auth: token", "Content-Type: application/json"},
			want: map[string]string{
				"Auth":         "token",
				"Content-Type": "application/json",
			},
		},
		{
			name:  "header with spaces",
			input: []string{"X-API-Key:  my-key-123  "},
			want:  map[string]string{"X-API-Key": "my-key-123"},
		},
		{
			name:  "header with colon in value",
			input: []string{"URL: https://example.com:8080"},
			want:  map[string]string{"URL": "https://example.com:8080"},
		},
		{
			name:    "invalid format - no colon",
			input:   []string{"InvalidHeader"},
			wantErr: true,
		},
		{
			name:    "invalid format - empty key",
			input:   []string{": value"},
			wantErr: true,
		},
		{
			name:  "empty value is valid",
			input: []string{"X-Empty:"},
			want:  map[string]string{"X-Empty": ""},
		},
		{
			name:  "empty input",
			input: []string{},
			want:  map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInlineHeaders(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInlineHeaders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !mapsEqual(got, tt.want) {
				t.Errorf("ParseInlineHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadHeaders(t *testing.T) {
	// Create temp directory for test files
	tmpDir := t.TempDir()

	t.Run("valid YAML file", func(t *testing.T) {
		yamlFile := filepath.Join(tmpDir, "headers.yml")
		content := `Authorization: Bearer test123
Content-Type: application/json
X-API-Key: secret-key`

		err := os.WriteFile(yamlFile, []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}

		headers, err := LoadHeaders(yamlFile)
		if err != nil {
			t.Fatalf("LoadHeaders() error = %v", err)
		}

		want := map[string]string{
			"Authorization": "Bearer test123",
			"Content-Type":  "application/json",
			"X-API-Key":     "secret-key",
		}

		if !mapsEqual(headers, want) {
			t.Errorf("LoadHeaders() = %v, want %v", headers, want)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := LoadHeaders("nonexistent.yml")
		if err == nil {
			t.Error("LoadHeaders() expected error for nonexistent file")
		}
	})

	t.Run("invalid YAML", func(t *testing.T) {
		yamlFile := filepath.Join(tmpDir, "invalid.yml")
		content := `this is not: valid: yaml: content`

		err := os.WriteFile(yamlFile, []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}

		_, err = LoadHeaders(yamlFile)
		if err == nil {
			t.Error("LoadHeaders() expected error for invalid YAML")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		yamlFile := filepath.Join(tmpDir, "empty.yml")
		err := os.WriteFile(yamlFile, []byte(""), 0644)
		if err != nil {
			t.Fatal(err)
		}

		headers, err := LoadHeaders(yamlFile)
		if err != nil {
			t.Fatalf("LoadHeaders() error = %v", err)
		}

		if len(headers) != 0 {
			t.Errorf("LoadHeaders() expected empty map, got %v", headers)
		}
	})
}

func TestMergeHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers []Headers
		want    map[string]string
	}{
		{
			name: "merge two header sets",
			headers: []Headers{
				{"Auth": "token1", "Content-Type": "json"},
				{"Auth": "token2", "Accept": "json"},
			},
			want: map[string]string{
				"Auth":         "token2", // Last wins
				"Content-Type": "json",
				"Accept":       "json",
			},
		},
		{
			name: "merge with empty",
			headers: []Headers{
				{"Auth": "token"},
				{},
			},
			want: map[string]string{"Auth": "token"},
		},
		{
			name:    "merge no headers",
			headers: []Headers{},
			want:    map[string]string{},
		},
		{
			name: "merge three sets",
			headers: []Headers{
				{"A": "1"},
				{"B": "2"},
				{"C": "3"},
			},
			want: map[string]string{"A": "1", "B": "2", "C": "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeHeaders(tt.headers...)
			if !mapsEqual(got, tt.want) {
				t.Errorf("MergeHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to compare maps
func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
