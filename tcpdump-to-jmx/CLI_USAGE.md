# CLI Usage Guide for tcpdump-to-jmx

## Overview

The `tcpdump-to-jmx` tool provides a command-line interface for converting PCAP (packet capture) files to HAR (HTTP Archive) and JMX (JMeter Test Plan) formats. This is useful for creating performance tests from captured network traffic.

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/your-repo/tcpdump-to-jmx.git
cd tcpdump-to-jmx

# Install dependencies (required for PCAP processing)
sudo apt-get install libpcap-dev  # On Ubuntu/Debian
# or
brew install libpcap  # On macOS

# Build the binary
make build
# or
go build -o tcpdump-to-jmx

# Optional: Install globally
make install
# or
sudo cp tcpdump-to-jmx /usr/local/bin/
```

## Commands

### Main Command

```bash
tcpdump-to-jmx --help
```

Available commands:
- `convert` - Convert PCAP files to HAR/JMX
- `server` - Start the API server
- `help` - Show help information

### Convert Command

The `convert` command is the main CLI functionality for converting PCAP files.

#### Basic Usage

```bash
# Convert PCAP to both HAR and JMX
tcpdump-to-jmx convert -i capture.pcap -o ./output

# Convert to HAR only
tcpdump-to-jmx convert -i capture.pcap -o ./output -t har

# Convert to JMX only
tcpdump-to-jmx convert -i capture.pcap -o ./output -t jmx
```

#### Command Options

| Flag | Description | Default |
|------|-------------|---------|
| `-i, --input` | Input PCAP file path (required) | - |
| `-o, --output` | Output directory for generated files | `.` |
| `-t, --type` | Output type: `har`, `jmx`, or `both` | `both` |
| `--port` | Filter traffic by port | 0 (all) |
| `--host` | Filter traffic by host | - |
| `--threads` | Number of threads in JMX test plan | 1 |
| `--ramp-up` | Ramp-up time in seconds for JMX | 1 |
| `--loops` | Number of loops in JMX test plan | 1 |
| `--correlation` | Enable automatic correlation detection | false |
| `--parameterization` | Enable automatic parameterization | false |
| `--pretty` | Pretty print JSON output | true |
| `-v, --verbose` | Enable verbose output | false |
| `--log-level` | Set log level (debug, info, warn, error) | info |

## Examples

### 1. Basic Conversion

```bash
# Convert a PCAP file to both HAR and JMX
tcpdump-to-jmx convert -i network_capture.pcap -o ./results
```

This creates:
- `./results/network_capture.har`
- `./results/network_capture.jmx`

### 2. Filter by Port

```bash
# Extract only HTTP traffic on port 8080
tcpdump-to-jmx convert -i capture.pcap -o ./output --port 8080
```

### 3. Filter by Host

```bash
# Extract traffic for specific host
tcpdump-to-jmx convert -i capture.pcap -o ./output --host api.example.com
```

### 4. Advanced JMX Configuration

```bash
# Create JMX with load test parameters
tcpdump-to-jmx convert \
  -i capture.pcap \
  -o ./output \
  -t jmx \
  --threads 50 \
  --ramp-up 10 \
  --loops 100 \
  --correlation \
  --parameterization \
  --verbose
```

### 5. HAR Only with Verbose Output

```bash
# Convert to HAR with detailed logging
tcpdump-to-jmx convert \
  -i capture.pcap \
  -o ./output \
  -t har \
  --verbose \
  --log-level debug
```

## Capturing Network Traffic

Before converting, you need to capture network traffic. Here are some methods:

### Using tcpdump

```bash
# Capture all HTTP traffic
sudo tcpdump -i any -w capture.pcap port 80 or port 8080

# Capture with packet count limit
sudo tcpdump -i any -w capture.pcap -c 1000 port 80

# Capture for specific host
sudo tcpdump -i any -w capture.pcap host api.example.com

# Capture with time limit (60 seconds)
sudo timeout 60 tcpdump -i any -w capture.pcap port 80
```

### Using the Test Script

```bash
# Generate test traffic and capture it
sudo python3 examples/generate_test_traffic.py

# This creates examples/test.pcap with sample HTTP traffic
```

## Complete Workflow Example

```bash
# Step 1: Capture network traffic
sudo tcpdump -i any -w api_traffic.pcap -c 500 port 8080

# Step 2: Convert to HAR and JMX with filters and options
tcpdump-to-jmx convert \
  -i api_traffic.pcap \
  -o ./test_plans \
  --port 8080 \
  --host api.myapp.com \
  --threads 100 \
  --ramp-up 30 \
  --loops 10 \
  --correlation \
  --parameterization \
  --verbose

# Step 3: View results
ls -la ./test_plans/
# api_traffic.har  - HTTP Archive for analysis
# api_traffic.jmx  - JMeter test plan ready to run

# Step 4: Open JMX in JMeter
jmeter -t ./test_plans/api_traffic.jmx
```

## Output Files

### HAR File
The HAR (HTTP Archive) file contains:
- All HTTP requests and responses
- Headers, cookies, and parameters
- Timing information
- Response bodies (if captured)

Use cases:
- Analyze API calls
- Debug network issues
- Import into browser dev tools
- Convert to other formats

### JMX File
The JMX (JMeter Test Plan) file contains:
- Thread Group with configured users
- HTTP Request samplers for each captured request
- Headers and parameters
- Correlation rules (if enabled)
- Parameterized values (if enabled)

Use cases:
- Load testing with JMeter
- Performance testing
- API testing
- Stress testing

## Tips and Best Practices

1. **Filter Early**: Use `--port` and `--host` filters to reduce noise
2. **Enable Correlation**: Use `--correlation` for dynamic values like tokens
3. **Parameterize Data**: Use `--parameterization` for test data variation
4. **Verbose Mode**: Use `-v` flag to see detailed conversion progress
5. **Check HAR First**: Review the HAR file before generating JMX
6. **Test Small**: Start with small PCAP files to verify conversion

## Troubleshooting

### No HTTP Traffic Found
```bash
# Check if PCAP contains HTTP traffic
tcpdump -r capture.pcap -nn 'tcp port 80 or tcp port 8080' | head
```

### Permission Denied
```bash
# Run with sudo for packet capture
sudo tcpdump -i any -w capture.pcap port 80

# Or change file ownership after capture
sudo chown $USER:$USER capture.pcap
```

### Missing libpcap
```bash
# Install on Ubuntu/Debian
sudo apt-get install libpcap-dev

# Install on macOS
brew install libpcap

# Install on CentOS/RHEL
sudo yum install libpcap-devel
```

## Server Mode

To run as an API server instead of CLI:

```bash
# Start server on default port (8080)
tcpdump-to-jmx server

# Start on custom port
tcpdump-to-jmx server --port 9090

# With verbose logging
tcpdump-to-jmx server --verbose --log-level debug
```

## Support

For issues, feature requests, or questions:
- GitHub Issues: https://github.com/your-repo/tcpdump-to-jmx/issues
- Documentation: https://github.com/your-repo/tcpdump-to-jmx/wiki