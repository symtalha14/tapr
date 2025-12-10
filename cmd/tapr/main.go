// Tapr is a command-line tool for checking API endpoint health,
// measuring latency, and detecting failures.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal" // Add this
	"strings"
	"sync"
	"syscall" // Add this
	"time"

	"github.com/spf13/cobra"
	"github.com/symtalha14/tapr/internal/config"
	"github.com/symtalha14/tapr/internal/output"
	"github.com/symtalha14/tapr/internal/request"
	"github.com/symtalha14/tapr/internal/stats"
)

// Version
var Version = "dev"

// ASCII logo for the application
const logo = `

********************************************************************	
   _	                   
  | |_ __ _ _ __  _ __ 
  | __/ _\ | \_ \| \__|    V0.1.0
  | || (_| | |_) | |   
  \___\__,_| .__/|_|   
           |_|

  LIGHTWEIGHT API PERFORMANCE MONITORING BUILT INTO CLI

********************************************************************
`

// Command-line flags
var (
	timeout          time.Duration // Request timeout duration
	method           string        // HTTP method (GET, POST, etc.)
	headersFile      string        // Path to YAML file containing headers
	inlineHeaders    []string      // Individual headers from command line
	verbose          bool          // Enable verbose output
	retries          int           // Number of retry attempts on failure
	watchInterval    time.Duration // Time between requests in watch mode
	watchCount       int           // Number of requests (0 = infinite)
	batchConcurrency int           // Number of concurrent requests in batch mode
	quiet            bool          // Only show errors
	silent           bool          // No output at all
	failFast         bool          // Stop on first failure
	maxTime          time.Duration // Maximum time for batch
	outputFormat     string        // Output format: pretty, json, csv
)

// Latency thresholds for color-coding responses
const (
	fastThreshold = 200 * time.Millisecond // Green: fast response
	slowThreshold = 500 * time.Millisecond // Red: slow response
)

// Exit codes for CI/CD integration
const (
	ExitSuccess = 0 // All tests passed
	ExitFailure = 1 // Some tests failed
	ExitError   = 2 // Configuration error, invalid arguments, etc.
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tapr [url]",
	Short: "A fast API health checker",
	Long: output.Green(logo) + `
 ⚡ Fast API Health Checker

A tiny CLI tool to ping your API endpoints, measure latency, 
and detect failures without opening Postman or CloudWatch.

Perfect for:
  • Quick API health checks
  • Measuring endpoint latency  
  • Pre-deployment smoke tests
  • Debugging slow APIs`,
	Example: `  tapr https://api.example.com/health
  tapr https://api.example.com/users -t 5s -v
  tapr https://api.example.com/orders -X POST -r 3
  tapr https://api.example.com -H "Authorization: Bearer token123"`,
	Args:    cobra.ExactArgs(1), // Require exactly one URL argument
	Run:     runPing,            // Execute the ping command
	Version: Version,
}

// watchCmd represents the watch command for continuous monitoring
var watchCmd = &cobra.Command{
	Use:   "watch [url]",
	Short: "Continuously monitor an endpoint",
	Long: `Watch mode continuously pings an endpoint at regular intervals,
displaying live statistics and recent request history.

Perfect for:
  • Monitoring API health over time
  • Detecting intermittent failures
  • Observing performance patterns
  • Real-time latency tracking`,
	Example: `  tapr watch https://api.example.com/health
  tapr watch https://api.example.com/health --interval 5s
  tapr watch https://api.example.com/health --count 20 -v`,
	Args: cobra.ExactArgs(1),
	Run:  runWatch,
}

// batchCmd represents the batch command for testing multiple endpoints
var batchCmd = &cobra.Command{
	Use:   "batch [config-file]",
	Short: "Test multiple endpoints from a config file",
	Long: `Batch mode tests multiple API endpoints concurrently from a YAML configuration file.
Results are displayed in a summary table showing the health of all endpoints.

Perfect for:
  • Smoke testing after deployment
  • Health checks across multiple services
  • Comparing performance of different endpoints
  • Pre-deployment validation`,
	Example: `  tapr batch endpoints.yml
  tapr batch endpoints.yml --concurrency 10
  tapr batch endpoints.yml -v`,
	Args: cobra.ExactArgs(1),
	Run:  runBatch,
}

// traceCmd represents the trace command for detailed timing analysis
var traceCmd = &cobra.Command{
	Use:   "trace [url]",
	Short: "Show detailed timing breakdown of a request",
	Long: `Trace mode performs a single request and shows detailed timing information
for each phase: DNS lookup, TCP connection, TLS handshake, server processing,
and content transfer.

Perfect for:
  • Identifying bottlenecks in request flow
  • Debugging slow APIs
  • Understanding where latency comes from
  • Optimizing API performance`,
	Example: `  tapr trace https://api.example.com/health
  tapr trace https://api.example.com/users -v
  tapr trace https://api.example.com/data -H "Authorization: Bearer token"`,
	Args: cobra.ExactArgs(1),
	Run:  runTrace,
}

// versionCmd outputs the current tapr version installed
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Tapr",
	Long:  "Print the version number of Tapr",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Tapr version %s\n", Version)
	},
}

// init initializes command-line flags and their default values.
// This function runs automatically before main().
func init() {

	// add version command to root
	rootCmd.AddCommand(versionCmd)

	// add watch command to root
	rootCmd.AddCommand(watchCmd)

	// add trace command to root
	rootCmd.AddCommand(traceCmd)

	// Watch-specific flags
	watchCmd.Flags().DurationVarP(
		&watchInterval,
		"interval",
		"i",
		2*time.Second,
		"Time between requests",
	)

	watchCmd.Flags().IntVarP(
		&watchCount,
		"count",
		"n",
		0,
		"Number of requests (0 = infinite)",
	)

	// Timeout flag: -t or --timeout
	rootCmd.Flags().DurationVarP(
		&timeout,
		"timeout",
		"t",
		10*time.Second,
		"Maximum time to wait for response",
	)

	// Method flag: -X or --method
	rootCmd.Flags().StringVarP(
		&method,
		"method",
		"X",
		"GET",
		"HTTP method (GET, POST, PUT, PATCH, DELETE)",
	)

	// Headers file flag: --headers
	rootCmd.Flags().StringVar(
		&headersFile,
		"headers",
		"",
		"Path to YAML file containing request headers",
	)

	// Inline header flag: -H or --header (repeatable)
	rootCmd.Flags().StringSliceVarP(
		&inlineHeaders,
		"header",
		"H",
		[]string{},
		"Add a header (format: 'Key: Value'), repeatable",
	)

	// Verbose flag: -v or --verbose
	rootCmd.Flags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"Show detailed request and response information",
	)

	// Retries flag: -r or --retries
	rootCmd.Flags().IntVarP(
		&retries,
		"retries",
		"r",
		0,
		"Number of retry attempts on failure",
	)

	// Add batch command
	rootCmd.AddCommand(batchCmd)

	// Batch-specific flags
	batchCmd.Flags().IntVarP(
		&batchConcurrency,
		"concurrency",
		"c",
		0,
		"Number of concurrent requests (0 = use config default)",
	)

	// Batch-specific CI/CD flags
	batchCmd.Flags().BoolVar(
		&failFast,
		"fail-fast",
		false,
		"Stop testing on first failure",
	)

	batchCmd.Flags().DurationVar(
		&maxTime,
		"max-time",
		0,
		"Maximum time for entire batch (e.g., 5m, 30s)",
	)

	// CI/CD flags (persistent - available on all commands)
	rootCmd.PersistentFlags().BoolVarP(
		&quiet,
		"quiet",
		"q",
		false,
		"Only show errors (no output on success)",
	)

	rootCmd.PersistentFlags().BoolVar(
		&silent,
		"silent",
		false,
		"No output at all (only exit code)",
	)

	rootCmd.PersistentFlags().StringVarP(
		&outputFormat,
		"output",
		"o",
		"pretty",
		"Output format: pretty, json, csv",
	)
}

// main is the entry point of the application.
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// runPing executes the ping command with the provided URL and flags.
func runPing(cmd *cobra.Command, args []string) {
	url := args[0]

	// Validate that URL has proper HTTP/HTTPS scheme
	if !isValidURL(url) {
		fmt.Fprintln(os.Stderr, output.Red("Error: URL must start with http:// or https://"))
		os.Exit(1)
	}

	// Load headers from file if specified
	var fileHeaders map[string]string
	if headersFile != "" {
		loadedHeaders, err := config.LoadHeaders(headersFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, output.Red(fmt.Sprintf("Error loading headers: %v", err)))
			os.Exit(1)
		}
		fileHeaders = loadedHeaders
	}

	// Parse inline headers if provided
	var parsedInlineHeaders map[string]string
	if len(inlineHeaders) > 0 {
		parsed, err := config.ParseInlineHeaders(inlineHeaders)
		if err != nil {
			fmt.Fprintln(os.Stderr, output.Red(fmt.Sprintf("Error parsing headers: %v", err)))
			os.Exit(1)
		}
		parsedInlineHeaders = parsed
	}

	// Merge file headers and inline headers (inline headers take precedence)
	headers := config.MergeHeaders(fileHeaders, parsedInlineHeaders)

	// Show request details in verbose mode
	if verbose {
		printRequestDetails(url, headers)
	}

	// Configure and execute the ping
	opts := request.PingOptions{
		Method:  strings.ToUpper(method),
		Timeout: timeout,
		Retries: retries,
		Headers: headers,
	}

	result := request.Ping(url, opts)

	// Handle request failure
	if result.Error != nil {
		printError(url, result.Error)
		os.Exit(1)
	}

	// Print successful result
	printSuccess(result)
}

// runWatch executes the watch command for continuous monitoring.
// runWatch executes the watch command for continuous monitoring.
func runWatch(cmd *cobra.Command, args []string) {
	url := args[0]

	// Validate URL
	if !isValidURL(url) {
		fmt.Fprintln(os.Stderr, output.Red("Error: URL must start with http:// or https://"))
		os.Exit(1)
	}

	// Load headers (same as ping command)
	var fileHeaders map[string]string
	if headersFile != "" {
		loadedHeaders, err := config.LoadHeaders(headersFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, output.Red(fmt.Sprintf("Error loading headers: %v", err)))
			os.Exit(1)
		}
		fileHeaders = loadedHeaders
	}

	var parsedInlineHeaders map[string]string
	if len(inlineHeaders) > 0 {
		parsed, err := config.ParseInlineHeaders(inlineHeaders)
		if err != nil {
			fmt.Fprintln(os.Stderr, output.Red(fmt.Sprintf("Error parsing headers: %v", err)))
			os.Exit(1)
		}
		parsedInlineHeaders = parsed
	}

	headers := config.MergeHeaders(fileHeaders, parsedInlineHeaders)

	// Print header
	fmt.Printf("\n┌─────────────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ Watching: %s%s│\n", output.Blue(url), strings.Repeat(" ", 70-len(url)-11))
	fmt.Printf("│ Interval: %v, ", watchInterval)
	if watchCount > 0 {
		fmt.Printf("Count: %d%s│\n", watchCount, strings.Repeat(" ", 48-len(fmt.Sprintf("%d", watchCount))))
	} else {
		fmt.Printf("Count: infinite%s│\n", strings.Repeat(" ", 43))
	}
	fmt.Printf("└─────────────────────────────────────────────────────────────────────┘\n")

	// Initialize trackers
	tracker := stats.NewTracker()
	history := stats.NewHistory(10) // Keep last 10 requests
	startTime := time.Now()

	// Configure request options
	opts := request.PingOptions{
		Method:  strings.ToUpper(method),
		Timeout: timeout,
		Retries: retries,
		Headers: headers,
	}

	// Setup signal handling for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Request counter
	requestCount := 0

	// Create ticker for periodic requests
	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()

	// Make first request immediately
	makeWatchRequest(url, opts, tracker, history)
	requestCount++
	displayWatchStats(tracker, history)

	// Channel to signal when to stop
	done := make(chan bool)

	// Goroutine to handle watch loop
	go func() {
		for {
			select {
			case <-ticker.C:
				makeWatchRequest(url, opts, tracker, history)
				requestCount++
				displayWatchStats(tracker, history)

				// Stop if we've reached the count limit
				if watchCount > 0 && requestCount >= watchCount {
					done <- true
					return
				}
			case <-sigChan:
				// Ctrl+C pressed
				done <- true
				return
			}
		}
	}()

	// Wait for completion
	<-done

	// Calculate total duration
	totalDuration := time.Since(startTime)

	// Display final summary
	displayWatchSummary(url, tracker, history, totalDuration, requestCount)
}

// makeWatchRequest makes a single request and updates trackers.
func makeWatchRequest(url string, opts request.PingOptions, tracker *stats.Tracker, history *stats.History) {
	result := request.Ping(url, opts)

	success := result.Error == nil
	tracker.Record(result.Latency, success)
	history.Add(result)
}

// displayWatchSummary shows a comprehensive summary when watch mode ends.
func displayWatchSummary(url string, tracker *stats.Tracker, history *stats.History, duration time.Duration, requestCount int) {
	// Clear screen one last time
	fmt.Print("\033[H\033[2J")

	fmt.Printf("\n")
	fmt.Printf("┌─────────────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ %s Watch Summary%s │\n", output.Blue("📋"), strings.Repeat(" ", 52))
	fmt.Printf("└─────────────────────────────────────────────────────────────────────┘\n")

	// Endpoint info
	fmt.Printf("🎯 Endpoint\n")
	fmt.Printf("   URL:      %s\n", url)
	fmt.Printf("   Method:   %s\n", method)
	fmt.Printf("   Duration: %s\n", duration.Round(time.Second))
	fmt.Printf("   Requests: %d\n", requestCount)

	// Success/Failure stats
	fmt.Printf("📊 Results\n")
	successRate := tracker.SuccessRate()

	var rateColor func(string) string
	var rateEmoji string
	if successRate == 100 {
		rateColor = output.Green
		rateEmoji = "✓"
	} else if successRate >= 80 {
		rateColor = output.Yellow
		rateEmoji = "⚠️"
	} else {
		rateColor = output.Red
		rateEmoji = "✗"
	}

	fmt.Printf("   Success Rate:  %s %s (%d/%d)\n",
		rateEmoji,
		rateColor(fmt.Sprintf("%.1f%%", successRate)),
		tracker.Successful,
		tracker.Total)
	fmt.Printf("   Successful:    %s\n", output.Green(fmt.Sprintf("%d", tracker.Successful)))
	fmt.Printf("   Failed:        %s\n", output.Red(fmt.Sprintf("%d", tracker.Failed)))
	fmt.Println()

	// Latency statistics
	if tracker.Total > 0 {
		fmt.Printf("⚡ Performance\n")
		fmt.Printf("   Min Latency:   %s\n", output.Cyan(tracker.MinLatency.String()))
		fmt.Printf("   Max Latency:   %s\n", output.Red(tracker.MaxLatency.String()))
		fmt.Printf("   Avg Latency:   %s\n", formatLatency(tracker.AvgLatency()))

		if tracker.Total >= 2 {
			fmt.Printf("   P50 Latency:   %s\n", tracker.Percentile(0.50).String())
			fmt.Printf("   P95 Latency:   %s\n", tracker.Percentile(0.95).String())
			fmt.Printf("   P99 Latency:   %s\n", tracker.Percentile(0.99).String())
		}

		// Calculate standard deviation for consistency
		stdDev := calculateStdDev(tracker.Latencies, tracker.AvgLatency())
		fmt.Printf("   Std Dev:       %s", stdDev.String())

		if stdDev < 50*time.Millisecond {
			fmt.Printf(" %s\n", output.Green("(very consistent)"))
		} else if stdDev < 200*time.Millisecond {
			fmt.Printf(" %s\n", output.Yellow("(moderate variance)"))
		} else {
			fmt.Printf(" %s\n", output.Red("(high variance)"))
		}
		fmt.Println()
	}

	// Insights section
	fmt.Printf("💡 Insights\n")
	insights := generateInsights(tracker, duration, requestCount)
	for _, insight := range insights {
		fmt.Printf("   %s\n", insight)
	}
	fmt.Println()

	// Final message
	if successRate == 100 {
		fmt.Printf("%s\n", output.Green("✓ All requests successful! API is healthy."))
	} else if successRate >= 80 {
		fmt.Printf("%s\n", output.Yellow("⚠️  Some failures detected. API may be unstable."))
	} else {
		fmt.Printf("%s\n", output.Red("✗ High failure rate. API needs attention!"))
	}
}

// displayWatchStats displays current statistics and recent history.
func displayWatchStats(tracker *stats.Tracker, history *stats.History) {
	// Clear previous output (move cursor up)
	// We'll implement this simply for now
	fmt.Print("\033[H\033[2J") // Clear screen

	// Display stats header
	fmt.Printf("\n📈 Live Stats (%d requests)\n", tracker.Total)

	// Success rate with color
	successRate := tracker.SuccessRate()
	var rateColor func(string) string
	if successRate == 100 {
		rateColor = output.Green
	} else if successRate >= 80 {
		rateColor = output.Yellow
	} else {
		rateColor = output.Red
	}

	fmt.Printf("   Success Rate:  %s (%d/%d)\n",
		rateColor(fmt.Sprintf("%.1f%%", successRate)),
		tracker.Successful,
		tracker.Total)

	// Latency stats
	if tracker.Total > 0 {
		fmt.Printf("   Avg Latency:   %s\n", formatLatency(tracker.AvgLatency()))
		fmt.Printf("   Min Latency:   %s\n", output.Green(tracker.MinLatency.String()))
		fmt.Printf("   Max Latency:   %s\n", output.Red(tracker.MaxLatency.String()))

		if tracker.Total >= 2 {
			fmt.Printf("   P95 Latency:   %s\n", tracker.Percentile(0.95).String())
		}
	}

	// Recent history with better formatting
	fmt.Printf("\n📊 Recent Checks\n")
	fmt.Printf("   %-8s  %-3s  %-10s  %-10s  %-25s\n", "TIME", "✓/✗", "STATUS", "LATENCY", "PERFORMANCE")
	fmt.Printf("   %s\n", strings.Repeat("─", 65))

	recent := history.GetRecent(5)

	for _, entry := range recent {
		timestamp := entry.Timestamp.Format("15:04:05")

		if entry.Result.Error != nil {
			fmt.Printf("   %-8s  %s  %-10s  %-10s  %s\n",
				timestamp,
				output.Red("✗"),
				"Error",
				entry.Result.Latency.String(),
				makeColoredLatencyBar(entry.Result.Latency, tracker.MaxLatency))
		} else {
			statusStr := fmt.Sprintf("%d", entry.Result.StatusCode)
			latencyStr := entry.Result.Latency.String()

			fmt.Printf("   %-8s  %s  %-10s  %-10s  %s\n",
				timestamp,
				output.Green("✓"),
				statusStr,
				latencyStr,
				makeColoredLatencyBar(entry.Result.Latency, tracker.MaxLatency))
		}
	}

	fmt.Printf("\n%s\n", output.Blue("Press Ctrl+C to stop..."))
}

// calculateStdDev calculates the standard deviation of latencies.
func calculateStdDev(latencies []time.Duration, avg time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	var sumSquares float64
	for _, latency := range latencies {
		diff := float64(latency - avg)
		sumSquares += diff * diff
	}

	variance := sumSquares / float64(len(latencies))
	stdDev := time.Duration(int64(variance))

	// Take square root approximation
	if stdDev > 0 {
		// Simple Newton's method for square root
		x := float64(stdDev)
		for i := 0; i < 10; i++ {
			x = (x + variance/x) / 2
		}
		stdDev = time.Duration(int64(x))
	}

	return stdDev
}

// generateInsights creates helpful observations about the API behavior.
func generateInsights(tracker *stats.Tracker, duration time.Duration, requestCount int) []string {
	insights := make([]string, 0)

	// Success rate insights
	successRate := tracker.SuccessRate()
	if successRate == 100 {
		insights = append(insights, output.Green("✓ Perfect reliability - no failures detected"))
	} else if tracker.Failed > 0 {
		failureRate := float64(tracker.Failed) / float64(tracker.Total) * 100
		insights = append(insights, output.Red(fmt.Sprintf("⚠️  %.1f%% failure rate - investigate error patterns", failureRate)))
	}

	// Latency insights
	if tracker.Total > 0 {
		avgLatency := tracker.AvgLatency()

		if avgLatency < 50*time.Millisecond {
			insights = append(insights, output.Cyan("⚡ Exceptional response times (< 50ms average)"))
		} else if avgLatency < 200*time.Millisecond {
			insights = append(insights, output.Green("✓ Fast response times (< 200ms average)"))
		} else if avgLatency < 500*time.Millisecond {
			insights = append(insights, output.Yellow("⚠️  Moderate response times (200-500ms average)"))
		} else if avgLatency < 1*time.Second {
			insights = append(insights, output.Yellow("⚠️  Slow response times (500ms-1s average)"))
		} else {
			insights = append(insights, output.Red("⚠️  Very slow response times (> 1s average)"))
		}

		// Variance insights
		stdDev := calculateStdDev(tracker.Latencies, avgLatency)
		varianceRatio := float64(stdDev) / float64(avgLatency)

		if varianceRatio < 0.2 {
			insights = append(insights, output.Green("✓ Highly consistent performance (low variance)"))
		} else if varianceRatio > 0.5 {
			insights = append(insights, output.Yellow("⚠️  Inconsistent performance (high variance)"))
		}

		// Range insights
		latencyRange := tracker.MaxLatency - tracker.MinLatency
		if latencyRange > 1*time.Second {
			insights = append(insights, output.Yellow(fmt.Sprintf("⚠️  Large latency spread: %s (min) to %s (max)",
				tracker.MinLatency, tracker.MaxLatency)))
		}

		// Throughput
		requestsPerSec := float64(requestCount) / duration.Seconds()
		insights = append(insights, fmt.Sprintf("📈 Throughput: %.2f requests/second", requestsPerSec))
	}

	// Duration insights
	if duration > 5*time.Minute {
		insights = append(insights, fmt.Sprintf("⏱️  Long monitoring session: %s", duration.Round(time.Second)))
	}

	return insights
}

// makeColoredLatencyBar creates a color-coded, well-formatted progress bar.
func makeColoredLatencyBar(latency, maxLatency time.Duration) string {
	if maxLatency == 0 {
		return "[···············]   0%"
	}

	barWidth := 15

	// Thresholds
	const blazingFastThreshold = 50 * time.Millisecond

	// Calculate filled blocks
	percentage := int(float64(latency) / float64(maxLatency) * 100)
	if percentage > 100 {
		percentage = 100
	}

	filled := int(float64(latency) / float64(maxLatency) * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}

	// For very fast responses, ensure at least 1 block is visible
	if latency < blazingFastThreshold && filled == 0 {
		filled = 1
	}

	var coloredBar string
	var badge string

	if latency < blazingFastThreshold {
		// Blazing fast - use stars instead of blocks
		filledBar := strings.Repeat("★", filled)
		emptyBar := strings.Repeat("·", barWidth-filled)
		coloredBar = output.Green(filledBar) + emptyBar
		badge = " ⚡"
	} else if latency < fastThreshold {
		// Fast - green blocks
		filledBar := strings.Repeat("█", filled)
		emptyBar := strings.Repeat("·", barWidth-filled)
		coloredBar = output.Green(filledBar) + emptyBar
		badge = ""
	} else if latency < slowThreshold {
		// Medium - yellow blocks
		filledBar := strings.Repeat("█", filled)
		emptyBar := strings.Repeat("·", barWidth-filled)
		coloredBar = output.Yellow(filledBar) + emptyBar
		badge = ""
	} else {
		// Slow - red blocks
		filledBar := strings.Repeat("█", filled)
		emptyBar := strings.Repeat("·", barWidth-filled)
		coloredBar = output.Red(filledBar) + emptyBar
		badge = ""
	}

	return fmt.Sprintf("[%s] %3d%%%s", coloredBar, percentage, badge)
}

// runBatch executes the batch command to test multiple endpoints.
func runBatch(cmd *cobra.Command, args []string) {
	configFile := args[0]

	// Load batch configuration
	batchConfig, err := config.LoadBatchConfig(configFile)
	if err != nil {
		if !silent {
			fmt.Fprintln(os.Stderr, output.Red(fmt.Sprintf("Error loading batch config: %v", err)))
		}
		os.Exit(ExitError)
	}

	// Override concurrency if specified via flag
	if batchConcurrency > 0 {
		batchConfig.Concurrency = batchConcurrency
	}

	// Print header (only in normal mode)
	if !quiet && !silent && outputFormat == "pretty" {
		fmt.Printf("\n┌─────────────────────────────────────────────────────────────────────┐\n")
		fmt.Printf("│ Running batch: %d endpoints (concurrency: %d)%s│\n",
			len(batchConfig.Endpoints),
			batchConfig.Concurrency,
			strings.Repeat(" ", 44-len(fmt.Sprintf("%d", len(batchConfig.Endpoints)))-len(fmt.Sprintf("%d", batchConfig.Concurrency))))
		fmt.Printf("└─────────────────────────────────────────────────────────────────────┘\n")

		fmt.Println("Testing endpoints... ⚡")
	}

	// Run batch tests
	startTime := time.Now()
	summary := runBatchTests(batchConfig)
	summary.TotalTime = time.Since(startTime)

	// Display results
	displayBatchResults(summary)
}

// runBatchTests executes all endpoint tests concurrently with CI/CD features.
func runBatchTests(batchConfig *config.BatchConfig) *stats.BatchSummary {
	summary := stats.NewBatchSummary()

	// Channel to collect results
	resultsChan := make(chan stats.BatchResult, len(batchConfig.Endpoints))

	// Channel to signal stopping (for fail-fast)
	stopChan := make(chan struct{})
	stopped := false

	// Semaphore to limit concurrency
	semaphore := make(chan struct{}, batchConfig.Concurrency)

	// WaitGroup to wait for all goroutines
	var wg sync.WaitGroup

	// Context with timeout (for max-time)
	ctx := context.Background()
	var cancel context.CancelFunc

	if maxTime > 0 {
		ctx, cancel = context.WithTimeout(ctx, maxTime)
		defer cancel()
	}

	// Launch goroutine for each endpoint
	for _, endpoint := range batchConfig.Endpoints {
		wg.Add(1)

		go func(ep config.Endpoint) {
			defer wg.Done()

			// Check if we should stop (fail-fast triggered)
			select {
			case <-stopChan:
				return
			case <-ctx.Done():
				return
			default:
			}

			// Acquire semaphore
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-stopChan:
				return
			case <-ctx.Done():
				return
			}

			// Test the endpoint
			result := testEndpoint(ep, batchConfig.Timeout)

			// Send result
			select {
			case resultsChan <- result:
				// If fail-fast is enabled and this test failed, signal stop
				if failFast && !result.Success && !stopped {
					stopped = true
					close(stopChan)
				}
			case <-stopChan:
				return
			case <-ctx.Done():
				return
			}
		}(endpoint)
	}

	// Close results channel when all goroutines finish
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for result := range resultsChan {
		summary.AddResult(result)

		// In quiet mode, print failures immediately
		if quiet && !silent && !result.Success {
			if result.Result.Error != nil {
				fmt.Fprintf(os.Stderr, "%s %s: %v\n",
					output.Red("✗"),
					result.Name,
					result.Result.Error)
			} else {
				fmt.Fprintf(os.Stderr, "%s %s: Expected %d, got %d\n",
					output.Red("✗"),
					result.Name,
					result.ExpectedStatus,
					result.Result.StatusCode)
			}
		}
	}

	// Check if we hit timeout
	if ctx.Err() == context.DeadlineExceeded {
		if !silent {
			fmt.Fprintf(os.Stderr, "%s Batch exceeded max-time limit (%v)\n",
				output.Yellow("⏱️"), maxTime)
		}
	}

	return summary
}

// testEndpoint tests a single endpoint and returns the result.
func testEndpoint(endpoint config.Endpoint, defaultTimeout time.Duration) stats.BatchResult {
	// Use endpoint-specific timeout or default
	timeout := endpoint.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	// Configure request
	opts := request.PingOptions{
		Method:  strings.ToUpper(endpoint.Method),
		Timeout: timeout,
		Retries: 0, // No retries in batch mode for speed
		Headers: endpoint.Headers,
	}

	// Make request
	result := request.Ping(endpoint.URL, opts)

	// Check if test passed
	success := result.Error == nil && result.StatusCode == endpoint.ExpectedStatus

	var message string
	if result.Error != nil {
		message = fmt.Sprintf("Error: %v", result.Error)
	} else if result.StatusCode != endpoint.ExpectedStatus {
		message = fmt.Sprintf("Expected %d, got %d", endpoint.ExpectedStatus, result.StatusCode)
	}

	return stats.BatchResult{
		Name:           endpoint.Name,
		URL:            endpoint.URL,
		Method:         endpoint.Method,
		Result:         result,
		ExpectedStatus: endpoint.ExpectedStatus,
		Success:        success,
		Message:        message,
	}
}

// displayBatchResults shows the batch test results based on output format.
func displayBatchResults(summary *stats.BatchSummary) {
	// Handle different output formats
	switch outputFormat {
	case "json":
		displayBatchResultsJSON(summary)
		return
	case "csv":
		displayBatchResultsCSV(summary)
		return
	case "pretty":
		// Continue with normal display
	default:
		fmt.Fprintf(os.Stderr, "Unknown output format: %s\n", outputFormat)
		os.Exit(ExitError)
	}

	// Silent mode: no output at all
	if silent {
		if summary.Failed > 0 {
			os.Exit(ExitFailure)
		}
		os.Exit(ExitSuccess)
	}

	// Quiet mode: errors already printed during execution
	if quiet {
		if summary.Failed > 0 {
			os.Exit(ExitFailure)
		}
		os.Exit(ExitSuccess)
	}

	// Normal mode: pretty output
	displayBatchResultsPretty(summary)
}

// displayBatchResultsJSON outputs results in JSON format.
func displayBatchResultsJSON(summary *stats.BatchSummary) {
	jsonOutput, err := output.FormatBatchResultJSON(summary)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
		os.Exit(ExitError)
	}

	fmt.Println(jsonOutput)

	if summary.Failed > 0 {
		os.Exit(ExitFailure)
	}
	os.Exit(ExitSuccess)
}

// displayBatchResultsCSV outputs results in CSV format.
func displayBatchResultsCSV(summary *stats.BatchSummary) {
	// CSV header
	fmt.Println("name,url,method,status,expected_status,latency_ms,size_bytes,success,error")

	// CSV rows
	for _, result := range summary.Results {
		errMsg := ""
		if result.Result.Error != nil {
			errMsg = result.Result.Error.Error()
		} else if !result.Success {
			errMsg = result.Message
		}

		fmt.Printf("%s,%s,%s,%d,%d,%d,%d,%t,%s\n",
			result.Name,
			result.URL,
			result.Method,
			result.Result.StatusCode,
			result.ExpectedStatus,
			result.Result.Latency.Milliseconds(),
			result.Result.Size,
			result.Success,
			errMsg,
		)
	}

	if summary.Failed > 0 {
		os.Exit(ExitFailure)
	}
	os.Exit(ExitSuccess)
}

// displayBatchResultsPretty shows the normal pretty output.
func displayBatchResultsPretty(summary *stats.BatchSummary) {
	// Table header
	fmt.Printf("%-20s %-7s %-7s %-10s %-8s %s\n",
		"ENDPOINT", "METHOD", "STATUS", "LATENCY", "SIZE", "RESULT")
	fmt.Printf("%s\n", strings.Repeat("─", 75))

	// Results rows
	for _, result := range summary.Results {
		// Format endpoint name (truncate if too long)
		name := result.Name
		if len(name) > 20 {
			name = name[:17] + "..."
		}

		// Format status
		statusStr := "-"
		if result.Result.Error == nil {
			statusStr = fmt.Sprintf("%d", result.Result.StatusCode)
		}

		// Format latency
		latencyStr := "-"
		if result.Result.Error == nil {
			latencyStr = result.Result.Latency.String()
		}

		// Format size
		sizeStr := "-"
		if result.Result.Size > 0 {
			sizeStr = formatBytes(result.Result.Size)
		}

		// Format result indicator
		var resultStr string
		if result.Success {
			if result.Result.Latency > 500*time.Millisecond {
				resultStr = output.Yellow("⚠️  SLOW")
			} else {
				resultStr = output.Green("✓")
			}
		} else {
			resultStr = output.Red(fmt.Sprintf("✗ %s", result.Message))
		}

		fmt.Printf("%-20s %-7s %-7s %-10s %-8s %s\n",
			name,
			result.Method,
			statusStr,
			latencyStr,
			sizeStr,
			resultStr)
	}

	// Summary section
	fmt.Printf("\n%s\n", strings.Repeat("─", 75))
	fmt.Printf("📊 Summary\n")
	fmt.Printf("   Total:        %d endpoints\n", summary.Total)

	successRate := summary.SuccessRate()
	var rateColor func(string) string
	if successRate == 100 {
		rateColor = output.Green
	} else if successRate >= 80 {
		rateColor = output.Yellow
	} else {
		rateColor = output.Red
	}

	fmt.Printf("   Successful:   %s (%.1f%%)\n",
		rateColor(fmt.Sprintf("%d", summary.Successful)),
		successRate)
	fmt.Printf("   Failed:       %s\n", output.Red(fmt.Sprintf("%d", summary.Failed)))

	if summary.Slow > 0 {
		fmt.Printf("   Slow:         %s (> 500ms)\n", output.Yellow(fmt.Sprintf("%d", summary.Slow)))
	}

	if summary.Total > 0 && summary.AvgLatency > 0 {
		fmt.Printf("   Avg Latency:  %s\n", formatLatency(summary.AvgLatency))
	}
	fmt.Printf("   Total Time:   %s\n", summary.TotalTime.Round(10*time.Millisecond))

	// Final message
	fmt.Println()
	if summary.Failed == 0 {
		fmt.Printf("%s\n", output.Green("✓ All endpoints healthy!"))
		os.Exit(ExitSuccess)
	} else {
		fmt.Printf("%s\n", output.Red(fmt.Sprintf("✗ %d endpoint(s) failed!", summary.Failed)))
		os.Exit(ExitFailure)
	}
}

// isValidURL checks if the URL starts with http:// or https://
func isValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// printRequestDetails displays verbose information about the request being made.
func printRequestDetails(url string, headers map[string]string) {
	fmt.Printf("   Request\n")
	fmt.Printf("   URL:     %s\n", output.Blue(url))
	fmt.Printf("   Method:  %s\n", method)
	fmt.Printf("   Timeout: %v\n", timeout)
	if retries > 0 {
		fmt.Printf("   Retries: %d\n", retries)
	}
	if len(headers) > 0 {
		fmt.Printf("   Headers: %d total\n", len(headers))
		for key, value := range headers {
			// Mask sensitive headers for security
			displayValue := value
			if isSensitiveHeader(key) {
				displayValue = maskSensitiveValue(value)
			}
			fmt.Printf("     %s: %s\n", key, displayValue)
		}
	}
	fmt.Println()
}

// isSensitiveHeader checks if a header contains sensitive information
func isSensitiveHeader(header string) bool {
	sensitive := []string{"authorization", "api-key", "x-api-key", "token", "password"}
	headerLower := strings.ToLower(header)
	for _, s := range sensitive {
		if strings.Contains(headerLower, s) {
			return true
		}
	}
	return false
}

// maskSensitiveValue masks a sensitive header value, showing only the last 4 characters
func maskSensitiveValue(value string) string {
	if len(value) <= 4 {
		return "***"
	}
	return "***" + value[len(value)-4:]
}

// printError displays a formatted error message for failed requests.
func printError(url string, err error) {
	fmt.Printf("%s Failed to ping %s\n", output.Red("✗"), url)
	fmt.Printf("  Error: %v\n", err)
}

// printSuccess displays a formatted success message with response details.
func printSuccess(result request.Result) {
	// Format latency with color based on speed
	latencyDisplay := formatLatency(result.Latency)

	// Print main success message
	fmt.Printf("%s Success\n", output.Green("✓"))
	fmt.Printf("  Status:   %s\n", result.Status)
	fmt.Printf("  Latency:  %s\n", latencyDisplay)

	// Show protocol if available
	if result.Protocol != "" {
		fmt.Printf("  Protocol: %s\n", result.Protocol)
	}

	// Show size if known (ContentLength returns -1 if unknown)
	if result.Size > 0 {
		fmt.Printf("  Size:     %s\n", formatBytes(result.Size))
	}
}

// formatLatency returns a color-coded latency string based on performance thresholds.
// Fast responses (<200ms) are green, medium (200-500ms) are yellow, slow (>500ms) are red.
func formatLatency(latency time.Duration) string {
	latencyStr := latency.String()

	if latency < fastThreshold {
		return output.Green(latencyStr)
	} else if latency < slowThreshold {
		return output.Yellow(latencyStr)
	}
	return output.Red(latencyStr)
}

// formatBytes converts a byte count to a human-readable string (e.g., "1.2 KB").
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
	)

	switch {
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

// runTrace executes the trace command to show detailed timing breakdown.
func runTrace(cmd *cobra.Command, args []string) {
	url := args[0]

	// Validate URL
	if !isValidURL(url) {
		fmt.Fprintln(os.Stderr, output.Red("Error: URL must start with http:// or https://"))
		os.Exit(1)
	}

	// Load headers
	var fileHeaders map[string]string
	if headersFile != "" {
		loadedHeaders, err := config.LoadHeaders(headersFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, output.Red(fmt.Sprintf("Error loading headers: %v", err)))
			os.Exit(1)
		}
		fileHeaders = loadedHeaders
	}

	var parsedInlineHeaders map[string]string
	if len(inlineHeaders) > 0 {
		parsed, err := config.ParseInlineHeaders(inlineHeaders)
		if err != nil {
			fmt.Fprintln(os.Stderr, output.Red(fmt.Sprintf("Error parsing headers: %v", err)))
			os.Exit(1)
		}
		parsedInlineHeaders = parsed
	}

	headers := config.MergeHeaders(fileHeaders, parsedInlineHeaders)

	// Print header
	fmt.Printf("\n┌─────────────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ %s Trace: %s%s│\n",
		output.Blue("🔍"),
		url,
		strings.Repeat(" ", 57-len(url)))
	fmt.Printf("└─────────────────────────────────────────────────────────────────────┘\n")

	if verbose {
		fmt.Printf("⚡ Request\n")
		fmt.Printf("   Method:  %s\n", method)
		fmt.Printf("   Timeout: %v\n", timeout)
		if len(headers) > 0 {
			fmt.Printf("   Headers: %d total\n", len(headers))
		}
		fmt.Println()
	}

	// Configure request
	opts := request.PingOptions{
		Method:  strings.ToUpper(method),
		Timeout: timeout,
		Headers: headers,
	}

	// Execute trace
	fmt.Println("Tracing request...")
	result := request.TraceRequest(url, opts.Method, opts)

	// Display results
	if result.Error != nil {
		fmt.Printf("%s Failed to trace request\n", output.Red("✗"))
		fmt.Printf("  Error: %v\n", result.Error)
		os.Exit(1)
	}

	displayTraceResults(result)
}

// displayTraceResults shows the detailed timing breakdown.
func displayTraceResults(result request.TraceResult) {
	fmt.Printf("📊 Request Timeline\n")

	// Calculate percentages
	total := float64(result.TotalTime)
	phases := []struct {
		name     string
		duration time.Duration
		color    func(string) string
	}{
		{"DNS Lookup", result.DNSLookup, output.Cyan},
		{"TCP Connection", result.TCPConnection, output.Green},
		{"TLS Handshake", result.TLSHandshake, output.Blue},
		{"Server Processing", result.ServerProcessing, output.Yellow},
		{"Content Transfer", result.ContentTransfer, output.Green},
	}

	// Find max duration for bar scaling
	maxDuration := result.DNSLookup
	for _, phase := range phases {
		if phase.duration > maxDuration {
			maxDuration = phase.duration
		}
	}

	// Display each phase
	for _, phase := range phases {
		if phase.duration == 0 {
			continue // Skip phases that didn't happen (e.g., TLS for HTTP)
		}

		percentage := float64(phase.duration) / total * 100
		barWidth := 20
		filled := int(float64(phase.duration) / float64(maxDuration) * float64(barWidth))
		if filled < 1 && phase.duration > 0 {
			filled = 1
		}

		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

		fmt.Printf("   %-18s %s  %-8s (%5.1f%%)\n",
			phase.name,
			phase.color(bar),
			phase.duration,
			percentage)
	}

	// Total
	fmt.Printf("   %s\n", strings.Repeat("─", 50))
	fmt.Printf("   %-18s %s  %s\n",
		"Total Time",
		strings.Repeat(" ", 20),
		output.Cyan(result.TotalTime.String()))

	// Response information
	fmt.Printf("📬 Response\n")
	fmt.Printf("   Status:   %s\n", formatStatusCode(result.StatusCode, result.Status))
	fmt.Printf("   Protocol: %s\n", result.Protocol)
	if result.Size > 0 {
		fmt.Printf("   Size:     %s\n", formatBytes(result.Size))
	}
	if result.RemoteAddr != "" {
		fmt.Printf("   Server:   %s\n", result.RemoteAddr)
	}
	fmt.Println()

	// Insights
	fmt.Printf("💡 Insights\n")
	insights := generateTraceInsights(result)
	for _, insight := range insights {
		fmt.Printf("   %s\n", insight)
	}
	fmt.Println()
}

// formatStatusCode formats the status code with color.
func formatStatusCode(code int, status string) string {
	if code >= 200 && code < 300 {
		return output.Green(status)
	} else if code >= 300 && code < 400 {
		return output.Blue(status)
	} else if code >= 400 && code < 500 {
		return output.Yellow(status)
	} else {
		return output.Red(status)
	}
}

// generateTraceInsights generates helpful observations about the trace.
func generateTraceInsights(result request.TraceResult) []string {
	insights := make([]string, 0)

	total := result.TotalTime

	// DNS insights
	if result.DNSLookup > 0 {
		dnsPercent := float64(result.DNSLookup) / float64(total) * 100
		if result.DNSLookup < 10*time.Millisecond {
			insights = append(insights, output.Green("✓ Fast DNS lookup (likely cached)"))
		} else if result.DNSLookup > 100*time.Millisecond {
			insights = append(insights, output.Yellow(fmt.Sprintf("⚠️  Slow DNS lookup (%v, %.1f%% of total)", result.DNSLookup, dnsPercent)))
		}
	}

	// TCP insights
	if result.TCPConnection > 0 {
		tcpPercent := float64(result.TCPConnection) / float64(total) * 100
		if result.TCPConnection < 20*time.Millisecond {
			insights = append(insights, output.Green("✓ Fast TCP connection (server nearby)"))
		} else if result.TCPConnection > 100*time.Millisecond {
			insights = append(insights, output.Yellow(fmt.Sprintf("⚠️  Slow TCP connection (%v, %.1f%% of total) - server may be far away", result.TCPConnection, tcpPercent)))
		}
	}

	// TLS insights
	if result.TLSHandshake > 0 {
		tlsPercent := float64(result.TLSHandshake) / float64(total) * 100
		if result.TLSHandshake < 50*time.Millisecond {
			insights = append(insights, output.Green("✓ Fast TLS handshake"))
		} else if result.TLSHandshake > 200*time.Millisecond {
			insights = append(insights, output.Yellow(fmt.Sprintf("⚠️  Slow TLS handshake (%v, %.1f%% of total) - consider connection reuse", result.TLSHandshake, tlsPercent)))
		}
	}

	// Server processing insights
	if result.ServerProcessing > 0 {
		serverPercent := float64(result.ServerProcessing) / float64(total) * 100
		if result.ServerProcessing < 100*time.Millisecond {
			insights = append(insights, output.Green("✓ Fast server processing"))
		} else if result.ServerProcessing > 500*time.Millisecond {
			insights = append(insights, output.Yellow(fmt.Sprintf("⚠️  Slow server processing (%v, %.1f%% of total) - backend optimization needed", result.ServerProcessing, serverPercent)))
		}

		// Check if server processing is the bottleneck
		if serverPercent > 50 {
			insights = append(insights, output.Yellow(fmt.Sprintf("⚠️  Server processing is %.1f%% of total time - main bottleneck", serverPercent)))
		}
	}

	// Content transfer insights
	if result.ContentTransfer > 0 && result.Size > 0 {
		transferPercent := float64(result.ContentTransfer) / float64(total) * 100
		if result.ContentTransfer < 50*time.Millisecond {
			insights = append(insights, output.Green("✓ Fast content transfer"))
		} else if transferPercent > 20 {
			insights = append(insights, output.Yellow(fmt.Sprintf("⚠️  Slow content transfer (%.1f%% of total) - consider compression or CDN", transferPercent)))
		}
	}

	// Overall assessment
	if total < 200*time.Millisecond {
		insights = append(insights, output.Cyan("⚡ Excellent overall performance (< 200ms)"))
	} else if total > 1*time.Second {
		insights = append(insights, output.Red("⚠️  Poor overall performance (> 1s) - multiple issues need attention"))
	}

	if len(insights) == 0 {
		insights = append(insights, "✓ No major issues detected")
	}

	return insights
}
