#!/usr/bin/env python3
"""
Example Python client for TCPDump to JMX Converter API
"""

import requests
import json
import time
import sys
from pathlib import Path

class TCPDumpToJMXClient:
    def __init__(self, base_url="http://localhost:8080"):
        self.base_url = base_url
        self.api_base = f"{base_url}/api/v1"
        
    def check_health(self):
        """Check API health status"""
        try:
            response = requests.get(f"{self.api_base}/health")
            if response.status_code == 200:
                return response.json()
            return None
        except Exception as e:
            print(f"Error checking health: {e}")
            return None
    
    def convert_file(self, pcap_file_path, options=None):
        """
        Convert PCAP file to HAR and JMX formats
        
        Args:
            pcap_file_path: Path to PCAP file
            options: Dictionary with conversion options
        
        Returns:
            Job ID if successful, None otherwise
        """
        if not Path(pcap_file_path).exists():
            print(f"File not found: {pcap_file_path}")
            return None
        
        # Default options
        default_options = {
            'correlation': 'true',
            'parameterization': 'true',
            'threads': '10',
            'rampup': '10',
            'loops': '1'
        }
        
        if options:
            default_options.update(options)
        
        try:
            with open(pcap_file_path, 'rb') as f:
                files = {'file': f}
                response = requests.post(
                    f"{self.api_base}/convert",
                    files=files,
                    data=default_options
                )
            
            if response.status_code == 202:  # Accepted
                result = response.json()
                return result.get('job_id')
            else:
                print(f"Error: {response.status_code} - {response.text}")
                return None
                
        except Exception as e:
            print(f"Error converting file: {e}")
            return None
    
    def get_job_status(self, job_id):
        """Get the status of a conversion job"""
        try:
            response = requests.get(f"{self.api_base}/status/{job_id}")
            if response.status_code == 200:
                return response.json()
            return None
        except Exception as e:
            print(f"Error getting job status: {e}")
            return None
    
    def wait_for_completion(self, job_id, timeout=300, poll_interval=2):
        """
        Wait for job to complete
        
        Args:
            job_id: Job ID to monitor
            timeout: Maximum time to wait in seconds
            poll_interval: Time between status checks in seconds
        
        Returns:
            Final job status or None if timeout
        """
        start_time = time.time()
        
        while time.time() - start_time < timeout:
            status = self.get_job_status(job_id)
            
            if not status:
                print("Failed to get job status")
                return None
            
            current_status = status.get('status')
            progress = status.get('progress', 0)
            
            print(f"Status: {current_status} - Progress: {progress}%")
            
            if current_status == 'completed':
                return status
            elif current_status == 'failed':
                print(f"Job failed: {status.get('error', 'Unknown error')}")
                return status
            
            time.sleep(poll_interval)
        
        print("Timeout waiting for job completion")
        return None
    
    def download_file(self, job_id, file_type='jmx', output_path=None):
        """
        Download converted file
        
        Args:
            job_id: Job ID
            file_type: 'har' or 'jmx'
            output_path: Where to save the file
        
        Returns:
            Path to downloaded file or None
        """
        if file_type not in ['har', 'jmx']:
            print(f"Invalid file type: {file_type}")
            return None
        
        if not output_path:
            output_path = f"output_{job_id}.{file_type}"
        
        try:
            response = requests.get(
                f"{self.api_base}/download/{job_id}/{file_type}",
                stream=True
            )
            
            if response.status_code == 200:
                with open(output_path, 'wb') as f:
                    for chunk in response.iter_content(chunk_size=8192):
                        f.write(chunk)
                print(f"File saved to: {output_path}")
                return output_path
            else:
                print(f"Error downloading file: {response.status_code}")
                return None
                
        except Exception as e:
            print(f"Error downloading file: {e}")
            return None
    
    def list_conversions(self, limit=20, offset=0):
        """List recent conversions"""
        try:
            response = requests.get(
                f"{self.api_base}/conversions",
                params={'limit': limit, 'offset': offset}
            )
            if response.status_code == 200:
                return response.json()
            return None
        except Exception as e:
            print(f"Error listing conversions: {e}")
            return None


def main():
    """Example usage"""
    
    # Initialize client
    client = TCPDumpToJMXClient("http://localhost:8080")
    
    # Check health
    print("🏥 Checking API health...")
    health = client.check_health()
    if health:
        print(f"✅ API is {health['status']}")
    else:
        print("❌ API is not responding")
        sys.exit(1)
    
    # Convert a file
    pcap_file = "sample.pcap"  # Replace with your PCAP file
    
    if not Path(pcap_file).exists():
        print(f"⚠️  Sample file '{pcap_file}' not found")
        print("Creating a dummy file for demonstration...")
        with open(pcap_file, 'wb') as f:
            f.write(b"dummy pcap data")
    
    print(f"\n📤 Converting {pcap_file}...")
    job_id = client.convert_file(pcap_file, {
        'correlation': 'true',
        'parameterization': 'true',
        'threads': '10',
        'rampup': '5'
    })
    
    if not job_id:
        print("❌ Failed to start conversion")
        sys.exit(1)
    
    print(f"✅ Job started with ID: {job_id}")
    
    # Wait for completion
    print("\n⏳ Waiting for conversion to complete...")
    final_status = client.wait_for_completion(job_id, timeout=60)
    
    if final_status and final_status['status'] == 'completed':
        print("✅ Conversion completed successfully!")
        
        # Download HAR file
        print("\n📥 Downloading HAR file...")
        har_file = client.download_file(job_id, 'har')
        
        # Download JMX file
        print("\n📥 Downloading JMX file...")
        jmx_file = client.download_file(job_id, 'jmx')
        
        if har_file and jmx_file:
            print(f"\n🎉 Success! Files downloaded:")
            print(f"  - HAR: {har_file}")
            print(f"  - JMX: {jmx_file}")
    else:
        print("❌ Conversion failed or timed out")
    
    # List recent conversions
    print("\n📋 Recent conversions:")
    conversions = client.list_conversions(limit=5)
    if conversions and conversions.get('jobs'):
        for job in conversions['jobs']:
            print(f"  - {job['id']}: {job['status']} ({job['file_name']})")


if __name__ == "__main__":
    main()