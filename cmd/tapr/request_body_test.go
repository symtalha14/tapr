package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestPingRequestBody(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "tapr")
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build tapr: %v\n%s", err, output)
	}

	dir := t.TempDir()
	writeFile := func(name string, data []byte) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	jsonBody := []byte("{\"message\":\"hello\"}")
	fileBody := append(append([]byte(nil), jsonBody...), '\n')
	jsonFile := writeFile("body.json", fileBody)
	binaryBody := []byte{'a', 0, 0xff, '\n'}
	binaryFile := writeFile("body.bin", binaryBody)
	emptyFile := writeFile("empty.txt", nil)
	headerFile := writeFile("headers.yaml", []byte("content-type: application/problem+json\n"))
	missingFile := filepath.Join(dir, "missing.json")

	tests := []struct {
		name           string
		method         string
		flags          []string
		body           []byte
		contentType    string
		retry          bool
		redirectStatus int
		exitCode       int
		requests       int32
		wantText       string
	}{
		{name: "no data", requests: 1},
		{name: "empty POST", method: "POST", requests: 1},
		{name: "inline JSON POST", method: "POST", flags: []string{"-d", string(jsonBody)}, body: jsonBody, contentType: "application/json", requests: 1},
		{name: "inline JSON PUT", method: "PUT", flags: []string{"--data", string(jsonBody)}, body: jsonBody, contentType: "application/json", requests: 1},
		{name: "inline JSON PATCH", method: "PATCH", flags: []string{"-d", string(jsonBody)}, body: jsonBody, contentType: "application/json", requests: 1},
		{name: "inline JSON DELETE", method: "DELETE", flags: []string{"-d", string(jsonBody)}, body: jsonBody, contentType: "application/json", requests: 1},
		{name: "default method stays GET", flags: []string{"-d", "plain text"}, body: []byte("plain text"), requests: 1},
		{name: "JSON array", method: "POST", flags: []string{"-d", "[1,2]"}, body: []byte("[1,2]"), contentType: "application/json", requests: 1},
		{name: "invalid JSON stays raw", method: "POST", flags: []string{"-d", "{invalid"}, body: []byte("{invalid"), requests: 1},
		{name: "explicit content type", method: "POST", flags: []string{"-d", string(jsonBody), "-H", "Content-Type: application/custom"}, body: jsonBody, contentType: "application/custom", requests: 1},
		{name: "lowercase content type", method: "POST", flags: []string{"-d", string(jsonBody), "-H", "content-type: text/plain"}, body: jsonBody, contentType: "text/plain", requests: 1},
		{name: "empty content type", method: "POST", flags: []string{"-d", string(jsonBody), "-H", "Content-Type:"}, body: jsonBody, requests: 1},
		{name: "content type from file", method: "POST", flags: []string{"-d", string(jsonBody), "--headers", headerFile}, body: jsonBody, contentType: "application/problem+json", requests: 1},
		{name: "JSON file", method: "POST", flags: []string{"--data", "@" + jsonFile}, body: fileBody, contentType: "application/json", requests: 1},
		{name: "binary file", method: "POST", flags: []string{"-d", "@" + binaryFile}, body: binaryBody, requests: 1},
		{name: "empty inline data", method: "POST", flags: []string{"-d", ""}, requests: 1},
		{name: "empty file", method: "POST", flags: []string{"-d", "@" + emptyFile}, requests: 1},
		{name: "missing file", method: "POST", flags: []string{"-d", "@" + missingFile}, exitCode: 2, wantText: "missing.json"},
		{name: "missing filename", method: "POST", flags: []string{"-d", "@"}, exitCode: 2, wantText: "Error loading request body"},
		{name: "retry inline data", method: "POST", flags: []string{"-d", string(jsonBody), "--retries", "1"}, body: jsonBody, contentType: "application/json", retry: true, requests: 2},
		{name: "retry file data", method: "POST", flags: []string{"-d", "@" + jsonFile, "--retries", "1"}, body: fileBody, contentType: "application/json", retry: true, requests: 2},
		{name: "temporary redirect", method: "POST", flags: []string{"-d", string(jsonBody)}, body: jsonBody, contentType: "application/json", redirectStatus: http.StatusTemporaryRedirect, requests: 2},
		{name: "permanent redirect", method: "POST", flags: []string{"-d", string(jsonBody)}, body: jsonBody, contentType: "application/json", redirectStatus: http.StatusPermanentRedirect, requests: 2},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			wantMethod := test.method
			if wantMethod == "" {
				wantMethod = http.MethodGet
			}
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("read request body: %v", err)
				}
				if !bytes.Equal(body, test.body) {
					t.Errorf("body = %q, want %q", body, test.body)
				}
				if r.Method != wantMethod {
					t.Errorf("method = %s, want %s", r.Method, wantMethod)
				}
				if got := r.Header.Get("Content-Type"); got != test.contentType {
					t.Errorf("Content-Type = %q, want %q", got, test.contentType)
				}
				if r.ContentLength != int64(len(test.body)) || len(r.TransferEncoding) != 0 {
					t.Errorf("ContentLength = %d, TransferEncoding = %v; want %d and no transfer encoding", r.ContentLength, r.TransferEncoding, len(test.body))
				}
				if requests.Add(1) == 1 {
					if test.retry {
						w.WriteHeader(http.StatusServiceUnavailable)
						return
					}
					if test.redirectStatus != 0 {
						w.Header().Set("Location", "/again")
						w.WriteHeader(test.redirectStatus)
						return
					}
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			args := []string{server.URL}
			if test.method != "" {
				args = append(args, "-X", test.method)
			}
			args = append(args, test.flags...)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
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
			if test.wantText != "" && !strings.Contains(string(output), test.wantText) {
				t.Errorf("output does not contain %q:\n%s", test.wantText, output)
			}
			if got := requests.Load(); got != test.requests {
				t.Errorf("requests = %d, want %d", got, test.requests)
			}
		})
	}
}
