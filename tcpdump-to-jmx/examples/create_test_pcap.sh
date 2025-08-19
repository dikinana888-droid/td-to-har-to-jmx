#!/bin/bash

# Script to create a test PCAP file for demonstration

echo "Creating test PCAP file..."
echo "Note: This requires sudo privileges to capture network traffic"

# Create examples directory if it doesn't exist
mkdir -p examples

# Capture 100 packets on any interface, filtering HTTP traffic
echo "Starting packet capture (will capture 100 packets)..."
echo "You may need to generate some HTTP traffic (browse websites, make API calls, etc.)"

sudo tcpdump -i any -w examples/test.pcap -c 100 \
    '(tcp port 80 or tcp port 8080 or tcp port 443) and (((ip[2:2] - ((ip[0]&0xf)<<2)) - ((tcp[12]&0xf0)>>2)) != 0)'

echo "PCAP file created: examples/test.pcap"

# Show file info
echo ""
echo "File information:"
ls -lh examples/test.pcap

echo ""
echo "You can now convert this file using:"
echo "  ./tcpdump-to-jmx convert -i examples/test.pcap -o ./output"