// Package request provides HTTP client functionality for making API requests
// and measuring their performance characteristics.
package request

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptrace"
	"time"
)

// TraceResult contains detailed timing information for each phase of an HTTP request.
type TraceResult struct {
	URL string // The URL that was traced

	// Timing for each phase
	DNSLookup        time.Duration // Time to resolve DNS
	TCPConnection    time.Duration // Time to establish TCP connection
	TLSHandshake     time.Duration // Time for TLS handshake (HTTPS only)
	ServerProcessing time.Duration // Time server took to process request
	ContentTransfer  time.Duration // Time to transfer response body
	TotalTime        time.Duration // Total end-to-end time

	// Additional metadata
	StatusCode int    // HTTP status code
	Status     string // HTTP status text
	Protocol   string // HTTP protocol version
	RemoteAddr string // Server IP address
	Size       int64  // Response size

	Error error // Any error that occurred
}

// TraceRequest performs an HTTP request with detailed timing information.
// It uses Go's httptrace package to capture timing at each phase.
func TraceRequest(url, method string, opts PingOptions) TraceResult {
	result := TraceResult{
		URL: url,
	}

	// Timing markers
	var (
		dnsStart     time.Time
		dnsDone      time.Time
		connectStart time.Time
		connectDone  time.Time
		tlsStart     time.Time
		tlsDone      time.Time
		gotConn      time.Time
		firstByte    time.Time
	)

	// Track the overall start time
	overallStart := time.Now()

	// Create trace hooks
	trace := &httptrace.ClientTrace{
		// DNS lookup
		DNSStart: func(_ httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			dnsDone = time.Now()
			result.DNSLookup = dnsDone.Sub(dnsStart)
		},

		// TCP connection
		ConnectStart: func(_, _ string) {
			connectStart = time.Now()
		},
		ConnectDone: func(_, _ string, err error) {
			if err == nil {
				connectDone = time.Now()
				result.TCPConnection = connectDone.Sub(connectStart)
			}
		},

		// TLS handshake (HTTPS only)
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, err error) {
			if err == nil {
				tlsDone = time.Now()
				result.TLSHandshake = tlsDone.Sub(tlsStart)
			}
		},

		// Connection obtained (reused or new)
		GotConn: func(_ httptrace.GotConnInfo) {
			gotConn = time.Now()
		},

		// First byte of response received
		GotFirstResponseByte: func() {
			firstByte = time.Now()
		},
	}

	// Create HTTP client with tracing and disabled keep-alives
	client := &http.Client{
		Timeout: opts.Timeout,
		Transport: &http.Transport{
			// CRITICAL: Disable connection pooling to force fresh connections
			DisableKeepAlives: true,
			// Disable compression to get accurate transfer times
			DisableCompression: false,
			// Force new connection for each request
			MaxIdleConns:        0,
			MaxIdleConnsPerHost: 0,
			IdleConnTimeout:     0,
		},
	}

	// Create request with trace context
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		result.Error = err
		return result
	}

	// Add headers
	for key, value := range opts.Headers {
		req.Header.Set(key, value)
	}

	// Attach trace to request context
	req = req.WithContext(httptrace.WithClientTrace(context.Background(), trace))

	// Execute request
	resp, err := client.Do(req)
	overallEnd := time.Now()

	if err != nil {
		result.Error = err
		result.TotalTime = overallEnd.Sub(overallStart)
		return result
	}
	defer resp.Body.Close()

	// Read the entire body to complete content transfer timing
	_, _ = io.ReadAll(resp.Body)
	transferEnd := time.Now()

	// Calculate server processing time
	// From when connection was ready to first byte
	if !gotConn.IsZero() && !firstByte.IsZero() {
		result.ServerProcessing = firstByte.Sub(gotConn)
	}

	// Calculate content transfer time
	// From first byte to end of body read
	if !firstByte.IsZero() {
		result.ContentTransfer = transferEnd.Sub(firstByte)
	}

	// Total time
	result.TotalTime = transferEnd.Sub(overallStart)

	// Capture response metadata
	result.StatusCode = resp.StatusCode
	result.Status = resp.Status
	result.Protocol = resp.Proto
	result.Size = resp.ContentLength

	// Get remote address if available
	if resp.Request != nil && resp.Request.RemoteAddr != "" {
		result.RemoteAddr = resp.Request.RemoteAddr
	}

	return result
}
