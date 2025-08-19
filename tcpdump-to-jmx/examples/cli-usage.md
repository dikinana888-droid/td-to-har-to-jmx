# CLI Usage Examples

## Installation

```bash
# Build the binary
go build -o tcpdump-to-jmx

# Or install globally
go install github.com/tcpdump-to-jmx@latest
```

## Basic Usage

### Convert PCAP to both HAR and JMX

```bash
tcpdump-to-jmx convert -i capture.pcap -o ./output
```

This will create:
- `./output/capture.har` - HTTP Archive file
- `./output/capture.jmx` - JMeter test plan

### Convert PCAP to HAR only

```bash
tcpdump-to-jmx convert -i capture.pcap -o ./output -t har
```

### Convert PCAP to JMX only

```bash
tcpdump-to-jmx convert -i capture.pcap -o ./output -t jmx
```

## Advanced Options

### Filter by Port

Extract only HTTP traffic on specific port:

```bash
tcpdump-to-jmx convert -i capture.pcap -o ./output --port 8080
```

### Filter by Host

Extract traffic for specific host:

```bash
tcpdump-to-jmx convert -i capture.pcap -o ./output --host api.example.com
```

### JMX Configuration

Configure JMeter test plan parameters:

```bash
tcpdump-to-jmx convert -i capture.pcap -o ./output -t jmx \
  --threads 10 \
  --ramp-up 5 \
  --loops 100
```

### Enable Correlation and Parameterization

Automatically detect and handle dynamic values:

```bash
tcpdump-to-jmx convert -i capture.pcap -o ./output -t jmx \
  --correlation \
  --parameterization
```

## Capturing Network Traffic

### Using tcpdump

```bash
# Capture all HTTP traffic on port 80
sudo tcpdump -i any -w capture.pcap port 80

# Capture traffic for specific host
sudo tcpdump -i any -w capture.pcap host api.example.com

# Capture with size limit
sudo tcpdump -i any -w capture.pcap -C 100 port 80
```

### Using Wireshark

1. Start Wireshark
2. Select network interface
3. Apply filter: `http or http2`
4. Start capture
5. Save as PCAP file

## Complete Example

```bash
# 1. Capture traffic
sudo tcpdump -i any -w api-traffic.pcap -C 100 port 8080

# 2. Convert to HAR and JMX with filters
tcpdump-to-jmx convert \
  -i api-traffic.pcap \
  -o ./test-plans \
  --port 8080 \
  --host api.myapp.com \
  --threads 50 \
  --ramp-up 10 \
  --loops 1000 \
  --correlation \
  --parameterization \
  --verbose

# 3. View generated files
ls -la ./test-plans/
# api-traffic.har
# api-traffic.jmx
```

## Server Mode

Start as web service:

```bash
# Start with default settings
tcpdump-to-jmx server

# Start on custom port
tcpdump-to-jmx server --port 9090

# With verbose logging
tcpdump-to-jmx server --verbose --log-level debug
```

## Help

```bash
# General help
tcpdump-to-jmx --help

# Command-specific help
tcpdump-to-jmx convert --help
tcpdump-to-jmx server --help
```