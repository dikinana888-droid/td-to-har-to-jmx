#!/usr/bin/env python3
"""
Script to generate test HTTP traffic and capture it to a PCAP file.
This creates a simple HTTP server and client for testing purposes.
"""

import subprocess
import threading
import time
import http.server
import socketserver
import requests
import os
import signal
import sys

PORT = 8888
PCAP_FILE = "examples/test.pcap"

class TestHTTPHandler(http.server.SimpleHTTPRequestHandler):
    """Simple HTTP handler for testing"""
    
    def do_GET(self):
        """Handle GET requests"""
        if self.path == "/api/users":
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(b'{"users": [{"id": 1, "name": "John"}, {"id": 2, "name": "Jane"}]}')
        elif self.path == "/api/products":
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(b'{"products": [{"id": 101, "name": "Laptop"}, {"id": 102, "name": "Phone"}]}')
        else:
            self.send_response(200)
            self.send_header("Content-Type", "text/html")
            self.end_headers()
            self.wfile.write(b"<html><body><h1>Test Server</h1></body></html>")
    
    def do_POST(self):
        """Handle POST requests"""
        content_length = int(self.headers.get('Content-Length', 0))
        post_data = self.rfile.read(content_length)
        
        self.send_response(201)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(b'{"status": "created", "id": 123}')
    
    def log_message(self, format, *args):
        """Suppress log messages"""
        pass

def start_server():
    """Start the test HTTP server"""
    with socketserver.TCPServer(("", PORT), TestHTTPHandler) as httpd:
        print(f"Test server running on port {PORT}")
        httpd.serve_forever()

def generate_traffic():
    """Generate test HTTP traffic"""
    time.sleep(2)  # Wait for server and tcpdump to start
    
    base_url = f"http://localhost:{PORT}"
    
    try:
        # Generate various HTTP requests
        print("Generating test traffic...")
        
        # GET requests
        requests.get(f"{base_url}/")
        requests.get(f"{base_url}/api/users")
        requests.get(f"{base_url}/api/products")
        requests.get(f"{base_url}/api/users?page=1&limit=10")
        
        # POST requests
        requests.post(f"{base_url}/api/users", json={"name": "Alice", "email": "alice@example.com"})
        requests.post(f"{base_url}/api/login", json={"username": "admin", "password": "secret"})
        
        # PUT request
        requests.put(f"{base_url}/api/users/1", json={"name": "John Updated"})
        
        # DELETE request
        requests.delete(f"{base_url}/api/users/2")
        
        # Headers and cookies
        requests.get(f"{base_url}/api/profile", headers={"Authorization": "Bearer token123"})
        requests.get(f"{base_url}/api/session", cookies={"session_id": "abc123xyz"})
        
        print("Test traffic generated successfully!")
        
    except Exception as e:
        print(f"Error generating traffic: {e}")

def main():
    """Main function"""
    print("Starting test environment...")
    
    # Create examples directory
    os.makedirs("examples", exist_ok=True)
    
    # Start HTTP server in a thread
    server_thread = threading.Thread(target=start_server, daemon=True)
    server_thread.start()
    
    # Start tcpdump to capture traffic
    print(f"Starting packet capture to {PCAP_FILE}...")
    tcpdump_process = subprocess.Popen([
        "tcpdump", "-i", "lo", "-w", PCAP_FILE, 
        f"port {PORT}", "-q"
    ], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    
    try:
        # Generate traffic
        generate_traffic()
        
        # Wait a bit for all packets to be captured
        time.sleep(2)
        
    finally:
        # Stop tcpdump
        tcpdump_process.terminate()
        tcpdump_process.wait()
        
        print(f"\nPCAP file created: {PCAP_FILE}")
        
        # Show file info
        if os.path.exists(PCAP_FILE):
            size = os.path.getsize(PCAP_FILE)
            print(f"File size: {size} bytes")
            print(f"\nYou can now convert this file using:")
            print(f"  ./tcpdump-to-jmx convert -i {PCAP_FILE} -o ./output")
        else:
            print("Error: PCAP file was not created")

if __name__ == "__main__":
    # Check if running as root (required for tcpdump)
    if os.geteuid() != 0:
        print("This script requires root privileges to run tcpdump.")
        print("Please run with: sudo python3 generate_test_traffic.py")
        sys.exit(1)
    
    main()