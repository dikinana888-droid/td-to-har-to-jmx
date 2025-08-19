#!/bin/bash

# Test CLI functionality for tcpdump-to-jmx

set -e

echo "==================================="
echo "Testing tcpdump-to-jmx CLI"
echo "==================================="

# Build the binary
echo "Building binary..."
go build -o tcpdump-to-jmx

# Create test directories
echo "Creating test directories..."
mkdir -p test_output
mkdir -p test_pcap

# Create a sample PCAP file (or use existing one)
echo "Note: You need to provide a PCAP file for testing"
echo "You can create one using:"
echo "  sudo tcpdump -i any -w test_pcap/sample.pcap -c 100 port 80"
echo ""

# Test help command
echo "Testing help command..."
./tcpdump-to-jmx --help
echo ""

echo "Testing convert help..."
./tcpdump-to-jmx convert --help
echo ""

# Check if sample PCAP exists
if [ -f "test_pcap/sample.pcap" ]; then
    echo "Found sample.pcap, running conversion tests..."
    
    # Test basic conversion
    echo "Test 1: Basic conversion (both HAR and JMX)..."
    ./tcpdump-to-jmx convert -i test_pcap/sample.pcap -o test_output
    
    # Check if files were created
    if [ -f "test_output/sample.har" ]; then
        echo "✓ HAR file created successfully"
    else
        echo "✗ HAR file not created"
    fi
    
    if [ -f "test_output/sample.jmx" ]; then
        echo "✓ JMX file created successfully"
    else
        echo "✗ JMX file not created"
    fi
    
    # Test HAR only conversion
    echo ""
    echo "Test 2: HAR only conversion..."
    ./tcpdump-to-jmx convert -i test_pcap/sample.pcap -o test_output -t har
    
    # Test JMX only conversion with options
    echo ""
    echo "Test 3: JMX conversion with options..."
    ./tcpdump-to-jmx convert \
        -i test_pcap/sample.pcap \
        -o test_output \
        -t jmx \
        --threads 10 \
        --ramp-up 5 \
        --loops 100 \
        --correlation \
        --parameterization \
        --verbose
    
    # Test with filters
    echo ""
    echo "Test 4: Conversion with filters..."
    ./tcpdump-to-jmx convert \
        -i test_pcap/sample.pcap \
        -o test_output \
        --port 80 \
        --verbose
    
    echo ""
    echo "==================================="
    echo "All tests completed!"
    echo "Check test_output/ for generated files"
    echo "==================================="
else
    echo "No sample.pcap found in test_pcap/"
    echo "Please create a PCAP file first:"
    echo "  sudo tcpdump -i any -w test_pcap/sample.pcap -c 100 port 80"
    echo "Or copy an existing PCAP file to test_pcap/sample.pcap"
fi

# Test server command (just check if it starts)
echo ""
echo "Testing server command (will timeout in 2 seconds)..."
timeout 2 ./tcpdump-to-jmx server --port 9090 || true
echo "Server command tested"

echo ""
echo "CLI testing complete!"