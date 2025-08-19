#!/bin/bash

# Test script for TCP Dump to JMX API

echo "🚀 Testing TCP Dump to JMX Converter API"
echo "========================================="

# Check health endpoint
echo -e "\n📍 Checking health endpoint..."
curl -s http://localhost:8080/api/v1/health | jq '.' || echo "Server not running"

# Create a sample PCAP file for testing (if needed)
echo -e "\n📦 Creating test PCAP file..."
# This would normally be a real PCAP file
echo "Test PCAP data" > test.pcap

# Upload file for conversion
echo -e "\n📤 Uploading PCAP file for conversion..."
response=$(curl -s -X POST http://localhost:8080/api/v1/convert \
  -F "file=@test.pcap" \
  -F "correlation=true" \
  -F "parameterization=true" \
  -F "threads=10" \
  -F "rampup=10")

echo "$response" | jq '.'

# Extract job ID
job_id=$(echo "$response" | jq -r '.job_id')

if [ "$job_id" != "null" ] && [ -n "$job_id" ]; then
  echo -e "\n📊 Job ID: $job_id"
  
  # Check job status
  echo -e "\n📈 Checking job status..."
  sleep 2
  curl -s "http://localhost:8080/api/v1/status/$job_id" | jq '.'
  
  # Try to connect to WebSocket (just show the URL)
  echo -e "\n🔌 WebSocket URL for real-time updates:"
  echo "ws://localhost:8080/api/v1/ws/$job_id"
else
  echo "❌ Failed to get job ID"
fi

echo -e "\n✅ Test completed!"