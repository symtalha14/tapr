package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/symtalha14/tapr/internal/config"
)

func TestPingHTTPStatus(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "tapr")
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build tapr: %v\n%s", err, output)
	}

	tests := []struct {
		name     string
		statuses []int
		flags    []string
		exitCode int
		requests int32
		wantText string
	}{
		{"ok", []int{200}, nil, 0, 1, "Success"},
		{"no content", []int{204}, nil, 0, 1, "Success"},
		{"redirect", []int{302}, nil, 0, 1, "Success"},
		{"not found", []int{404}, nil, 1, 1, "404"},
		{"server error", []int{500}, nil, 1, 1, "500"},
		{"unavailable", []int{503}, nil, 1, 1, "503"},
		{"expected not found", []int{404}, []string{"--expect", "404"}, 0, 1, "Success"},
		{"expected status mismatch", []int{200}, []string{"--expect", "204"}, 1, 1, "expected 204"},
		{"expected error is not retried", []int{503}, []string{"--expect", "503", "--retries", "1"}, 0, 1, "Success"},
		{"expected rate limit is not retried", []int{429}, []string{"--expect", "429", "--retries", "1"}, 0, 1, "Success"},
		{"retry server error", []int{500, 200}, []string{"--retries", "1"}, 0, 2, "Success"},
		{"retry unavailable", []int{503, 200}, []string{"--retries", "1"}, 0, 2, "Success"},
		{"retry rate limit", []int{429, 200}, []string{"--retries", "1"}, 0, 2, "Success"},
		{"retries exhausted", []int{503}, []string{"--retries", "1"}, 1, 2, "503"},
		{"not found is not retried", []int{404}, []string{"--retries", "1"}, 1, 1, "404"},
		{"retry transport error", []int{0, 200}, []string{"--retries", "1"}, 0, 2, "Success"},
		{"retry until expected error", []int{503, 500}, []string{"--expect", "500", "--retries", "1"}, 0, 2, "Success"},
		{"invalid zero expectation", []int{200}, []string{"--expect", "0"}, 2, 0, "--expect"},
		{"invalid low expectation", []int{200}, []string{"--expect", "99"}, 2, 0, "--expect"},
		{"invalid high expectation", []int{200}, []string{"--expect", "600"}, 2, 0, "--expect"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				index := int(requests.Add(1)) - 1
				if index >= len(test.statuses) {
					index = len(test.statuses) - 1
				}
				status := test.statuses[index]
				if status == 0 {
					connection, _, err := w.(http.Hijacker).Hijack()
					if err != nil {
						t.Errorf("hijack connection: %v", err)
						return
					}
					connection.Close()
					return
				}
				w.WriteHeader(status)
			}))
			defer server.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			args := append([]string{server.URL}, test.flags...)
			output, err := exec.CommandContext(ctx, binary, args...).CombinedOutput()
			exitCode := 0
			if err != nil {
				var exitError *exec.ExitError
				if !errors.As(err, &exitError) {
					t.Fatalf("run tapr: %v\n%s", err, output)
				}
				exitCode = exitError.ExitCode()
			}
			if exitCode != test.exitCode {
				t.Errorf("exit code = %d, want %d\n%s", exitCode, test.exitCode, output)
			}
			if !strings.Contains(string(output), test.wantText) {
				t.Errorf("output does not contain %q:\n%s", test.wantText, output)
			}
			if test.exitCode == 1 && !strings.Contains(string(output), "✗") {
				t.Errorf("failed health check has no failure marker:\n%s", output)
			}
			if got := requests.Load(); got != test.requests {
				t.Errorf("requests = %d, want %d", got, test.requests)
			}
		})
	}
}

func TestBatchAllowsExpectedErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	result := testEndpoint(config.Endpoint{
		Name:           "missing resource",
		URL:            server.URL,
		Method:         http.MethodGet,
		ExpectedStatus: http.StatusNotFound,
	}, time.Second)
	if !result.Success {
		t.Fatalf("expected 404 should remain successful in batch mode: %+v", result)
	}
}

func TestBoxLine(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"x", "x  "},
		{"abc", "abc"},
		{"over width", "over width"},
		{"é", "é  "},
	} {
		if got := boxLine(tc.input, 3); got != tc.want {
			t.Errorf("boxLine(%q, 3) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
