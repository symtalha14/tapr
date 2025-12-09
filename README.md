```
*************************************************
  _                   
 | |_ __ _ _ __  _ __ 
 | __/ _` | '_ \| '__|
 | || (_| | |_) | |   
  \__\__,_| .__/|_|   
          |_|
-------------------------------------------------
TAP INTO API REQUESTS STARIGHT FROM YOUR TERMINAL
-------------------------------------------------
```

# Tapr

**Tapr** is a fast, minimal CLI tool to tap into API endpoints, inspect headers, measure latency, compare responses, and track changes over time — all without leaving your terminal. Perfect for developers, DevOps engineers, and backend teams.

---

## Features

- Inspect API responses quickly
- Measure request latency in real-time
- Compare responses between endpoints
- Send custom headers from `headers.yml`
- Continuous monitoring with intervals
- Export results to JSON or CSV for further analysis
- Lightweight, single binary (no heavy dependencies)

---

## Installation

### From Source (Go Required)

```bash
git clone https://github.com/yourusername/tapr.git
cd tapr
go build -o tapr ./cmd/tapr
```

# Usage Examples

## 1. Basic latency check
```tapr https://api.example.com/health --interval 10s```

## 2. Compare two endpoints
```tapr diff https://v1.api.com/users https://v2.api.com/users```

## 3. Continuous monitoring
```tapr watch https://api.example.com/login```

## 4. Use custom headers
```tapr https://api.example.com/orders --headers headers.yml```

## 5. Show CLI help
```tapr --help```


## Configuration

### headers.yml example:

```
Authorization: Bearer <TOKEN>
User-Agent: TaprCLI/1.0
X-Custom-Header: MyValue
```


### CLI Flags

| Flag | Alias | Type | Description |
|------|-------|------|-------------|
| `--interval` | `-i` | string | Set the polling interval for requests (e.g., `10s`, `1m`) |
| `--headers` | `-H` | file | Path to YAML file containing headers to send with requests |
| `--diff` | `-d` | string | Compare responses between two endpoints |
| `--watch` | `-w` | boolean | Continuously monitor endpoint |
| `--output` | `-o` | string | Export results in `json` or `csv` format |
| `--help` | `-h` | boolean | Show help and available commands |
| `--version` | `-v` | boolean | Display Tapr version |



## Contributing

Tapr is an open-source project built for developers by developers. Whether you want to debug APIs faster, improve CLI tooling, or just play with Go networking, your contributions are **welcome and celebrated**. Check out [CONTRIBUTING.MD](Contributing.MD) for the complete contribution guide.



## Fork the repository

Create a new branch: ```git checkout -b feature/your-feature```

Commit your changes: ```git commit -m "Add your feature"```

Push to the branch: ```git push origin feature/your-feature```



## Open a Pull Request

Please check CONTRIBUTING.md
 for detailed guidelines.
 


## Code of Conduct

This project adheres to a Code of Conduct
. By participating, you agree to follow its guidelines.



