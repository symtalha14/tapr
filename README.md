```
************************************************************************************************************************
  _                   
 | |_ __ _ _ __  _ __ 
 | __/ _` | '_ \| '__|
 | || (_| | |_) | |   
  \__\__,_| .__/|_|   
          |_|         
************************************************************************************************************************
```

[![Tests](https://github.com/symtalha14/tapr/actions/workflows/quality.yml/badge.svg)](https://github.com/symtalha14/tapr/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/symtalha14/tapr)](https://goreportcard.com/report/github.com/symtalha14/tapr)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)


# Tapr

> **Tap into API performance, straight from your terminal**

**Tapr** is a fast, lightweight CLI tool for API health checking, performance monitoring, and debugging. Built in Go for speed and reliability, it's perfect for developers, DevOps engineers, and SREs who need quick insights into API behavior.

**Why Tapr?**
- 🚀 **Fast**: Single binary, zero dependencies, instant startup
- 🎯 **Focused**: Does one thing well - API health monitoring
- 🔧 **CI/CD Ready**: JSON/CSV output, exit codes, fail-fast mode
- 🎨 **Beautiful**: Color-coded output with visual latency bars
- ⚡ **Real-time**: Watch mode with live statistics
- 🔍 **Detailed**: Trace mode shows DNS, TCP, TLS, and server timing

---

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage Examples](#usage-examples)
- [Configuration](#configuration)
- [Command Reference](#command-reference)
- [CI/CD Integration](#cicd-integration)
- [Output Formats](#output-formats)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

---

## Features

### Core Capabilities

- ✅ **Single Endpoint Testing** - Quick health checks with latency measurement
- ✅ **Batch Testing** - Test multiple endpoints concurrently from a config file
- ✅ **Continuous Monitoring** - Watch mode with live updating statistics
- ✅ **Request Tracing** - Detailed breakdown of DNS, TCP, TLS, and server processing time
- ✅ **Custom Headers** - Support for inline headers and YAML header files
- ✅ **Retry Logic** - Configurable retries with exponential backoff
- ✅ **Multiple Output Formats** - Pretty (terminal), JSON, or CSV
- ✅ **CI/CD Integration** - Exit codes, quiet mode, fail-fast for automation

### Performance Metrics

Tapr measures and reports:
- Total request latency
- Success/failure rates
- P50, P95, P99 percentiles
- Min/max/average response times
- Standard deviation for consistency analysis
- Response size and HTTP protocol version

---

## Installation

### Option 1: Download Binary (Recommended)

Pre-built binaries are available in [Releases](https://github.com/symtalha14/tapr/releases).
```bash
# Linux
curl -L https://github.com/symtalha14/tapr/releases/latest/download/tapr-linux-amd64 -o tapr
chmod +x tapr
sudo mv tapr /usr/local/bin/

# macOS
curl -L https://github.com/symtalha14/tapr/releases/latest/download/tapr-darwin-amd64 -o tapr
chmod +x tapr
sudo mv tapr /usr/local/bin/

# Windows
# Download tapr-windows-amd64.exe from releases
```

### Option 2: Install with Go
```bash
go install github.com/symtalha14/tapr/cmd/tapr@latest
```

### Option 3: Build from Source
```bash
git clone https://github.com/symtalha14/tapr.git
cd tapr
go build -o tapr ./cmd/tapr/
sudo mv tapr /usr/local/bin/
```

### Verify Installation
```bash
tapr --help
```

---

## Quick Start

### Basic Health Check
```bash
# Simple GET request
tapr https://api.example.com/health

# Output:
# ✓ 200 OK
# Latency: 142ms
# Size: 312 bytes
```

### With Custom Headers
```bash
# Inline headers
tapr https://api.example.com/users \
  -H "Authorization: Bearer token123" \
  -H "Content-Type: application/json"

# From file
tapr https://api.example.com/users --headers headers.yml
```

### Continuous Monitoring
```bash
# Watch with 5-second intervals
tapr watch https://api.example.com/health --interval 5s

# Live output shows:
# - Success rate
# - Average/min/max latency
# - P95 percentile
# - Recent request history with visual bars
```

### Batch Testing
```bash
# Test multiple endpoints from config
tapr batch endpoints.yml

# With CI/CD options
tapr batch endpoints.yml --quiet --fail-fast --output json
```

### Request Tracing
```bash
# See detailed timing breakdown
tapr trace https://api.example.com/data

# Shows:
# - DNS Lookup: 12ms
# - TCP Connection: 24ms
# - TLS Handshake: 89ms
# - Server Processing: 156ms
# - Content Transfer: 8ms
```

---

## Usage Examples

### 1. Basic Request with Retry
```bash
tapr https://api.example.com/flaky-endpoint --retries 3 --timeout 5s
```

### 2. Continuous Monitoring (10 Requests)
```bash
tapr watch https://api.example.com/health --interval 2s --count 10
```

**Sample Output:**
```
📈 Live Stats (10 requests)
   Success Rate:  100% (10/10) ✓
   Avg Latency:   234ms
   Min Latency:   145ms
   Max Latency:   356ms
   P95 Latency:   340ms

📊 Recent Checks
   14:30:15  ✓  200 OK    145ms  [██░░░░░░░░]  30%
   14:30:17  ✓  200 OK    234ms  [█████░░░░░]  50%
   14:30:19  ✓  200 OK    356ms  [█████████░]  95%
```

### 3. Batch Testing for Microservices

**endpoints.yml:**
```yaml
concurrency: 5
timeout: 10s

endpoints:
  - name: "Auth Service"
    url: https://auth.example.com/health
    method: GET
    expected_status: 200

  - name: "User Service"
    url: https://users.example.com/health
    method: GET
    expected_status: 200

  - name: "Payment Gateway"
    url: https://payments.example.com/health
    method: GET
    expected_status: 200
    headers:
      X-API-Key: secret-key
```

**Run:**
```bash
tapr batch endpoints.yml
```

**Sample Output:**
```
ENDPOINT          METHOD  STATUS  LATENCY  SIZE    RESULT
Auth Service      GET     200     142ms    1.2KB   ✓
User Service      GET     200     234ms    892B    ✓
Payment Gateway   POST    200     189ms    3.4KB   ✓

📊 Summary
   Total:        3 endpoints
   Successful:   3 (100%)
   Failed:       0 (0%)
   Avg Latency:  188ms
   Total Time:   1.2s

✓ All endpoints healthy!
```

### 4. CI/CD Integration

**GitHub Actions example:**
```yaml
- name: Check API Health
  run: tapr batch production-endpoints.yml --quiet --fail-fast
```

**Jenkins example:**
```groovy
sh 'tapr batch endpoints.yml --output json > health-report.json'
```

### 5. Export to JSON for Analysis
```bash
tapr batch endpoints.yml --output json > results.json

# Parse with jq
cat results.json | jq '.results[] | select(.success == false)'
```

### 6. Performance Tracing
```bash
tapr trace https://api.example.com/slow-endpoint
```

**Sample Output:**
```
📊 Request Timeline

   DNS Lookup         ██░░░░░░░░░░░░░░  12ms   ( 4.1%)
   TCP Connection     ████░░░░░░░░░░░░  24ms   ( 8.3%)
   TLS Handshake      ████████████░░░░  89ms   (30.7%)
   Server Processing  ████████████████  156ms  (53.8%)
   Content Transfer   ██░░░░░░░░░░░░░░  8ms    ( 2.8%)

💡 Insights
   ⚠️  TLS handshake is slow (30.7% of total)
   ⚠️  Server processing is the main bottleneck
```

---

## Configuration

### Headers File (YAML)

**headers.yml:**
```yaml
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json
User-Agent: Tapr/0.1.0
X-Request-ID: abc-123-def
X-API-Key: your-api-key-here
```

**Usage:**
```bash
tapr https://api.example.com/users --headers headers.yml
```

### Batch Configuration (YAML)

**batch-config.yml:**
```yaml
# Global settings
timeout: 30s
concurrency: 10

# Endpoints to test
endpoints:
  - name: "Production API"
    url: https://api.example.com/health
    method: GET
    expected_status: 200
    
  - name: "Database Health"
    url: https://api.example.com/db/ping
    method: GET
    expected_status: 200
    timeout: 5s  # Override global timeout
    
  - name: "Create User Endpoint"
    url: https://api.example.com/users
    method: POST
    expected_status: 201
    headers:
      Content-Type: application/json
      X-Test-Mode: "true"
```

---

## Command Reference

### Global Flags

Available on all commands:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--timeout` | `-t` | duration | `10s` | Maximum time to wait for response |
| `--method` | `-X` | string | `GET` | HTTP method (GET, POST, PUT, PATCH, DELETE) |
| `--headers` | | string | | Path to YAML file with headers |
| `--header` | `-H` | string[] | | Inline header (repeatable): `"Key: Value"` |
| `--verbose` | `-v` | bool | `false` | Show detailed request/response info |
| `--retries` | `-r` | int | `0` | Number of retry attempts on failure |
| `--quiet` | `-q` | bool | `false` | Only show errors (for CI/CD) |
| `--silent` | | bool | `false` | No output at all, only exit code |
| `--output` | `-o` | string | `pretty` | Output format: `pretty`, `json`, `csv` |

### Commands

#### `tapr [URL]`

Test a single endpoint.

**Examples:**
```bash
tapr https://api.example.com
tapr https://api.example.com -X POST -H "Auth: token"
tapr https://api.example.com --timeout 30s --retries 3
```

---

#### `tapr watch [URL]`

Continuously monitor an endpoint with live statistics.

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--interval` | `-i` | duration | `2s` | Time between requests |
| `--count` | `-n` | int | `0` | Number of requests (0 = infinite) |

**Examples:**
```bash
# Monitor every 5 seconds
tapr watch https://api.example.com --interval 5s

# Monitor 20 times
tapr watch https://api.example.com --count 20

# Infinite monitoring with custom headers
tapr watch https://api.example.com -i 3s -H "Auth: token"
```

**Press Ctrl+C to stop and see summary.**

---

#### `tapr trace [URL]`

Show detailed timing breakdown for each request phase.

**Examples:**
```bash
tapr trace https://api.example.com
tapr trace https://api.example.com -H "Authorization: Bearer token"
```

---

#### `tapr batch [CONFIG]`

Test multiple endpoints from a YAML configuration file.

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--concurrency` | `-c` | int | `5` | Number of concurrent requests |
| `--fail-fast` | | bool | `false` | Stop on first failure |
| `--max-time` | | duration | `0` | Maximum time for entire batch |

**Examples:**
```bash
# Basic batch test
tapr batch endpoints.yml

# CI/CD mode
tapr batch endpoints.yml --quiet --fail-fast --output json

# High concurrency
tapr batch endpoints.yml --concurrency 20

# Time-limited
tapr batch endpoints.yml --max-time 2m
```

---

## CI/CD Integration

### Exit Codes

Tapr uses standard exit codes for automation:

- `0` - Success (all tests passed)
- `1` - Failure (some tests failed)
- `2` - Error (configuration error, invalid arguments)

### GitHub Actions

**.github/workflows/api-health.yml:**
```yaml
name: API Health Check

on:
  push:
    branches: [main]
  schedule:
    - cron: '*/15 * * * *'  # Every 15 minutes

jobs:
  health-check:
    runs-on: ubuntu-latest
    steps:
      - name: Install Tapr
        run: |
          curl -L https://github.com/symtalha14/tapr/releases/latest/download/tapr-linux-amd64 -o tapr
          chmod +x tapr
          sudo mv tapr /usr/local/bin/

      - name: Check Production APIs
        run: tapr batch production-endpoints.yml --quiet --fail-fast

      - name: Notify on Failure
        if: failure()
        run: |
          curl -X POST ${{ secrets.SLACK_WEBHOOK }} \
            -d '{"text":"🚨 Production APIs are unhealthy!"}'
```

### GitLab CI

**.gitlab-ci.yml:**
```yaml
api-health-check:
  stage: test
  script:
    - tapr batch staging-endpoints.yml --quiet --fail-fast
  only:
    - merge_requests
```

### Jenkins

**Jenkinsfile:**
```groovy
pipeline {
    agent any
    stages {
        stage('API Health Check') {
            steps {
                sh 'tapr batch endpoints.yml --quiet --output json > results.json'
                archiveArtifacts artifacts: 'results.json'
            }
        }
    }
}
```

### Bash Script Integration
```bash
#!/bin/bash

# Pre-deployment health check
if tapr batch production-endpoints.yml --quiet --fail-fast; then
    echo "✓ APIs healthy, proceeding with deployment"
    ./deploy.sh
else
    echo "✗ APIs unhealthy, aborting deployment"
    exit 1
fi
```

---

## Output Formats

### Pretty (Default)

Beautiful, color-coded terminal output with visual indicators.
```bash
tapr batch endpoints.yml
```

### JSON

Machine-readable format for parsing and automation.
```bash
tapr batch endpoints.yml --output json
```

**Sample Output:**
```json
{
  "total": 3,
  "successful": 2,
  "failed": 1,
  "success_rate": 66.67,
  "avg_latency_ms": 234,
  "total_time_ms": 1543,
  "results": [
    {
      "name": "Auth API",
      "url": "https://api.example.com/auth",
      "method": "GET",
      "status": 200,
      "latency_ms": 142,
      "success": true
    }
  ]
}
```

### CSV

Spreadsheet-friendly format for analysis.
```bash
tapr batch endpoints.yml --output csv > results.csv
```

**Sample Output:**
```csv
name,url,method,status,expected_status,latency_ms,size_bytes,success,error
Auth API,https://api.example.com/auth,GET,200,200,142,1024,true,
User API,https://api.example.com/users,GET,200,200,234,2048,true,
```

---

## Troubleshooting

### Common Issues

**Problem:** `Error: URL must start with http:// or https://`
```bash
# Wrong:
tapr api.example.com

# Right:
tapr https://api.example.com
```

**Problem:** `Error: headers file not found`
```bash
# Check file path
ls headers.yml

# Use absolute path if needed
tapr https://api.example.com --headers /path/to/headers.yml
```

**Problem:** Tests passing but batch fails
```bash
# Check expected_status in config
# Make sure it matches actual response
```

**Problem:** Slow batch tests
```bash
# Increase concurrency
tapr batch endpoints.yml --concurrency 10

# Set max time limit
tapr batch endpoints.yml --max-time 30s
```

---

## Performance Tips

### 1. Adjust Concurrency
```bash
# For many endpoints, increase concurrency
tapr batch endpoints.yml --concurrency 20
```

### 2. Use Fail-Fast in CI
```bash
# Stop immediately on first failure (saves time)
tapr batch endpoints.yml --fail-fast
```

### 3. Set Timeouts
```bash
# Don't wait forever for slow endpoints
tapr watch https://slow-api.com --timeout 5s
```

### 4. Limit Watch Duration
```bash
# Stop after N requests
tapr watch https://api.com --count 100
```

---

## Roadmap

### v0.2.0 (Planned)
- [ ] Request body support (POST/PUT data)
- [ ] Response body display and comparison
- [ ] Save results to file
- [ ] HTML report generation

### v0.3.0 (Planned)
- [ ] Configuration file (~/.taprrc)
- [ ] Rate limiting (--rate 10/s)
- [ ] Comparison mode (diff between endpoints)
- [ ] Authentication helpers (OAuth, API keys)

### v1.0.0 (Future)
- [ ] Stable API
- [ ] Comprehensive documentation
- [ ] Plugin system
- [ ] Web UI for results

**Want a feature?** [Open an issue](https://github.com/symtalha14/tapr/issues)!

---

## Contributing

Tapr is built for developers by developers. Contributions are **welcome and celebrated**! 🎉

### Ways to Contribute

- 🐛 **Report bugs** - [Open an issue](https://github.com/symtalha14/tapr/issues)
- 💡 **Suggest features** - Share your ideas
- 📝 **Improve docs** - Fix typos, add examples
- 🧪 **Write tests** - Increase coverage
- 🔧 **Fix issues** - Check [good first issues](https://github.com/symtalha14/tapr/labels/good%20first%20issue)
- ⭐ **Star the repo** - Show your support!

### Quick Start for Contributors
```bash
# Fork and clone
git clone https://github.com/YOUR_USERNAME/tapr.git
cd tapr

# Create a branch
git checkout -b feature/your-feature

# Make changes and test
go test ./...
go build -o tapr ./cmd/tapr/

# Format code
gofmt -w .

# Commit and push
git commit -m "Add your feature"
git push origin feature/your-feature

# Open a Pull Request on GitHub
```

**See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.**

---

## Community

- 💬 [Discussions](https://github.com/symtalha14/tapr/discussions) - Ask questions, share ideas
- 🐛 [Issues](https://github.com/symtalha14/tapr/issues) - Report bugs, request features
- 📢 [Twitter](https://twitter.com/symtalha14) - Follow for updates

## License

Tapr is released under the [Apache License 2.0](LICENSE).
```
                                 Apache License
                           Version 2.0, January 2004
                        http://www.apache.org/licenses/

   Copyright 2025 Syed Muhammad Talha

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
```

## Acknowledgments

Built with ❤️ using:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Go](https://golang.org/) - Programming language
- [YAML](https://github.com/go-yaml/yaml) - Configuration parsing

---

<div align="center">

**[⬆ back to top](#tapr)**

Made with ❤️ by [Syed Muhammad Talha](https://github.com/symtalha14)

If you find Tapr useful, please ⭐ star the repo!

</div>