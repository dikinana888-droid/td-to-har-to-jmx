#!/usr/bin/env node

/**
 * Example Node.js client for TCPDump to JMX Converter API
 */

const fs = require('fs');
const path = require('path');
const FormData = require('form-data');
const axios = require('axios');
const WebSocket = require('ws');

class TCPDumpToJMXClient {
    constructor(baseUrl = 'http://localhost:8080') {
        this.baseUrl = baseUrl;
        this.apiBase = `${baseUrl}/api/v1`;
    }

    /**
     * Check API health status
     */
    async checkHealth() {
        try {
            const response = await axios.get(`${this.apiBase}/health`);
            return response.data;
        } catch (error) {
            console.error('Error checking health:', error.message);
            return null;
        }
    }

    /**
     * Convert PCAP file to HAR and JMX formats
     */
    async convertFile(pcapFilePath, options = {}) {
        if (!fs.existsSync(pcapFilePath)) {
            console.error(`File not found: ${pcapFilePath}`);
            return null;
        }

        const defaultOptions = {
            correlation: 'true',
            parameterization: 'true',
            threads: '10',
            rampup: '10',
            loops: '1',
            ...options
        };

        try {
            const form = new FormData();
            form.append('file', fs.createReadStream(pcapFilePath));
            
            Object.entries(defaultOptions).forEach(([key, value]) => {
                form.append(key, value);
            });

            const response = await axios.post(
                `${this.apiBase}/convert`,
                form,
                {
                    headers: form.getHeaders()
                }
            );

            if (response.status === 202) {
                return response.data.job_id;
            }
            
            console.error(`Error: ${response.status} - ${response.data}`);
            return null;
        } catch (error) {
            console.error('Error converting file:', error.message);
            return null;
        }
    }

    /**
     * Get the status of a conversion job
     */
    async getJobStatus(jobId) {
        try {
            const response = await axios.get(`${this.apiBase}/status/${jobId}`);
            return response.data;
        } catch (error) {
            console.error('Error getting job status:', error.message);
            return null;
        }
    }

    /**
     * Monitor job progress via WebSocket
     */
    monitorProgress(jobId) {
        return new Promise((resolve, reject) => {
            const ws = new WebSocket(`ws://localhost:8080/api/v1/ws/${jobId}`);
            
            ws.on('open', () => {
                console.log('📡 WebSocket connected');
            });

            ws.on('message', (data) => {
                try {
                    const update = JSON.parse(data);
                    console.log(`Progress: ${update.progress}% - ${update.message || update.status}`);
                    
                    if (update.status === 'completed') {
                        ws.close();
                        resolve(update);
                    } else if (update.status === 'failed') {
                        ws.close();
                        reject(new Error(update.error || 'Conversion failed'));
                    }
                } catch (error) {
                    console.error('Error parsing WebSocket message:', error);
                }
            });

            ws.on('error', (error) => {
                console.error('WebSocket error:', error);
                reject(error);
            });

            ws.on('close', () => {
                console.log('📡 WebSocket disconnected');
            });

            // Timeout after 5 minutes
            setTimeout(() => {
                ws.close();
                reject(new Error('Timeout waiting for conversion'));
            }, 300000);
        });
    }

    /**
     * Wait for job completion (polling method)
     */
    async waitForCompletion(jobId, timeout = 300, pollInterval = 2) {
        const startTime = Date.now();
        
        while ((Date.now() - startTime) / 1000 < timeout) {
            const status = await this.getJobStatus(jobId);
            
            if (!status) {
                console.error('Failed to get job status');
                return null;
            }
            
            console.log(`Status: ${status.status} - Progress: ${status.progress}%`);
            
            if (status.status === 'completed') {
                return status;
            } else if (status.status === 'failed') {
                console.error(`Job failed: ${status.error || 'Unknown error'}`);
                return status;
            }
            
            await new Promise(resolve => setTimeout(resolve, pollInterval * 1000));
        }
        
        console.error('Timeout waiting for job completion');
        return null;
    }

    /**
     * Download converted file
     */
    async downloadFile(jobId, fileType = 'jmx', outputPath = null) {
        if (!['har', 'jmx'].includes(fileType)) {
            console.error(`Invalid file type: ${fileType}`);
            return null;
        }

        if (!outputPath) {
            outputPath = `output_${jobId}.${fileType}`;
        }

        try {
            const response = await axios.get(
                `${this.apiBase}/download/${jobId}/${fileType}`,
                {
                    responseType: 'stream'
                }
            );

            const writer = fs.createWriteStream(outputPath);
            response.data.pipe(writer);

            return new Promise((resolve, reject) => {
                writer.on('finish', () => {
                    console.log(`✅ File saved to: ${outputPath}`);
                    resolve(outputPath);
                });
                writer.on('error', reject);
            });
        } catch (error) {
            console.error('Error downloading file:', error.message);
            return null;
        }
    }

    /**
     * List recent conversions
     */
    async listConversions(limit = 20, offset = 0) {
        try {
            const response = await axios.get(
                `${this.apiBase}/conversions`,
                {
                    params: { limit, offset }
                }
            );
            return response.data;
        } catch (error) {
            console.error('Error listing conversions:', error.message);
            return null;
        }
    }
}

// Example usage
async function main() {
    const client = new TCPDumpToJMXClient('http://localhost:8080');
    
    // Check health
    console.log('🏥 Checking API health...');
    const health = await client.checkHealth();
    if (health) {
        console.log(`✅ API is ${health.status}`);
    } else {
        console.log('❌ API is not responding');
        process.exit(1);
    }
    
    // Convert a file
    const pcapFile = 'sample.pcap';
    
    if (!fs.existsSync(pcapFile)) {
        console.log(`⚠️  Sample file '${pcapFile}' not found`);
        console.log('Creating a dummy file for demonstration...');
        fs.writeFileSync(pcapFile, 'dummy pcap data');
    }
    
    console.log(`\n📤 Converting ${pcapFile}...`);
    const jobId = await client.convertFile(pcapFile, {
        correlation: 'true',
        parameterization: 'true',
        threads: '10',
        rampup: '5'
    });
    
    if (!jobId) {
        console.log('❌ Failed to start conversion');
        process.exit(1);
    }
    
    console.log(`✅ Job started with ID: ${jobId}`);
    
    // Choose monitoring method
    const useWebSocket = true;
    
    if (useWebSocket) {
        // Monitor via WebSocket
        console.log('\n🔌 Monitoring progress via WebSocket...');
        try {
            await client.monitorProgress(jobId);
            console.log('✅ Conversion completed successfully!');
        } catch (error) {
            console.error('❌ Conversion failed:', error.message);
            process.exit(1);
        }
    } else {
        // Monitor via polling
        console.log('\n⏳ Waiting for conversion to complete...');
        const finalStatus = await client.waitForCompletion(jobId, 60);
        
        if (!finalStatus || finalStatus.status !== 'completed') {
            console.log('❌ Conversion failed or timed out');
            process.exit(1);
        }
        console.log('✅ Conversion completed successfully!');
    }
    
    // Download files
    console.log('\n📥 Downloading HAR file...');
    const harFile = await client.downloadFile(jobId, 'har');
    
    console.log('\n📥 Downloading JMX file...');
    const jmxFile = await client.downloadFile(jobId, 'jmx');
    
    if (harFile && jmxFile) {
        console.log('\n🎉 Success! Files downloaded:');
        console.log(`  - HAR: ${harFile}`);
        console.log(`  - JMX: ${jmxFile}`);
    }
    
    // List recent conversions
    console.log('\n📋 Recent conversions:');
    const conversions = await client.listConversions(5);
    if (conversions && conversions.jobs) {
        conversions.jobs.forEach(job => {
            console.log(`  - ${job.id}: ${job.status} (${job.file_name})`);
        });
    }
}

// Run if executed directly
if (require.main === module) {
    main().catch(console.error);
}

module.exports = TCPDumpToJMXClient;