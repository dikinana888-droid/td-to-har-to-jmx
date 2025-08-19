package worker

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tcpdump-to-jmx/internal/config"
	"github.com/tcpdump-to-jmx/internal/converter"
	"github.com/tcpdump-to-jmx/internal/models"
	"github.com/tcpdump-to-jmx/internal/storage"
)

// JobManager manages conversion jobs
type JobManager struct {
	storage      *storage.S3Storage
	config       *config.Config
	jobs         map[string]*models.Job
	jobsMutex    sync.RWMutex
	workers      chan struct{}
	subscribers  map[string][]chan models.ProgressUpdate
	subMutex     sync.RWMutex
}

// NewJobManager creates a new job manager
func NewJobManager(storage *storage.S3Storage, config *config.Config) *JobManager {
	jm := &JobManager{
		storage:     storage,
		config:      config,
		jobs:        make(map[string]*models.Job),
		workers:     make(chan struct{}, config.Worker.MaxWorkers),
		subscribers: make(map[string][]chan models.ProgressUpdate),
	}
	
	// Initialize worker pool
	for i := 0; i < config.Worker.MaxWorkers; i++ {
		jm.workers <- struct{}{}
	}
	
	// Start cleanup routine
	go jm.cleanupRoutine()
	
	return jm
}

// SubmitJob submits a new conversion job
func (jm *JobManager) SubmitJob(job *models.Job, pcapData []byte) {
	jm.jobsMutex.Lock()
	jm.jobs[job.ID] = job
	jm.jobsMutex.Unlock()
	
	// Start processing in background
	go jm.processJob(job, pcapData)
}

// GetJob returns a job by ID
func (jm *JobManager) GetJob(jobID string) (*models.Job, bool) {
	jm.jobsMutex.RLock()
	defer jm.jobsMutex.RUnlock()
	
	job, exists := jm.jobs[jobID]
	return job, exists
}

// ListJobs returns a list of jobs
func (jm *JobManager) ListJobs(limit, offset string) []*models.Job {
	jm.jobsMutex.RLock()
	defer jm.jobsMutex.RUnlock()
	
	jobs := make([]*models.Job, 0, len(jm.jobs))
	for _, job := range jm.jobs {
		jobs = append(jobs, job)
	}
	
	// Sort by creation time (newest first)
	for i := 0; i < len(jobs)-1; i++ {
		for j := i + 1; j < len(jobs); j++ {
			if jobs[i].CreatedAt.Before(jobs[j].CreatedAt) {
				jobs[i], jobs[j] = jobs[j], jobs[i]
			}
		}
	}
	
	return jobs
}

// Subscribe subscribes to job updates
func (jm *JobManager) Subscribe(jobID string) chan models.ProgressUpdate {
	jm.subMutex.Lock()
	defer jm.subMutex.Unlock()
	
	ch := make(chan models.ProgressUpdate, 10)
	jm.subscribers[jobID] = append(jm.subscribers[jobID], ch)
	
	return ch
}

// Unsubscribe unsubscribes from job updates
func (jm *JobManager) Unsubscribe(jobID string, ch chan models.ProgressUpdate) {
	jm.subMutex.Lock()
	defer jm.subMutex.Unlock()
	
	if subs, exists := jm.subscribers[jobID]; exists {
		for i, sub := range subs {
			if sub == ch {
				jm.subscribers[jobID] = append(subs[:i], subs[i+1:]...)
				close(ch)
				break
			}
		}
	}
}

// processJob processes a conversion job
func (jm *JobManager) processJob(job *models.Job, pcapData []byte) {
	// Acquire worker
	<-jm.workers
	defer func() {
		jm.workers <- struct{}{}
	}()
	
	// Update job status
	jm.updateJobStatus(job, models.JobStatusProcessing, 0, "Starting conversion...")
	
	// Step 1: Convert PCAP to HAR (0-50% progress)
	logrus.Infof("Converting PCAP to HAR for job %s", job.ID)
	jm.updateJobStatus(job, models.JobStatusProcessing, 10, "Parsing PCAP file...")
	
	pcapConverter := converter.NewPcapToHarConverter()
	if job.Options.FilterPort != "" {
		// Parse port filter
		// pcapConverter.SetPortFilter(port)
	}
	if job.Options.FilterHost != "" {
		pcapConverter.SetHostFilter(job.Options.FilterHost)
	}
	
	har, err := pcapConverter.Convert(pcapData)
	if err != nil {
		jm.updateJobStatus(job, models.JobStatusFailed, 0, fmt.Sprintf("Failed to convert PCAP: %v", err))
		return
	}
	
	jm.updateJobStatus(job, models.JobStatusProcessing, 30, "PCAP converted to HAR successfully")
	
	// Save HAR to S3
	harData, err := json.MarshalIndent(har, "", "  ")
	if err != nil {
		jm.updateJobStatus(job, models.JobStatusFailed, 0, fmt.Sprintf("Failed to marshal HAR: %v", err))
		return
	}
	
	harKey := fmt.Sprintf("%s/output/%s.har", job.ID, job.ID)
	harURL, err := jm.storage.Upload(harKey, harData, "application/json")
	if err != nil {
		jm.updateJobStatus(job, models.JobStatusFailed, 0, fmt.Sprintf("Failed to upload HAR: %v", err))
		return
	}
	job.HarS3URL = harURL
	
	jm.updateJobStatus(job, models.JobStatusProcessing, 50, "HAR file saved to S3")
	
	// Step 2: Convert HAR to JMX (50-100% progress)
	logrus.Infof("Converting HAR to JMX for job %s", job.ID)
	jm.updateJobStatus(job, models.JobStatusProcessing, 60, "Converting HAR to JMX...")
	
	jmxConverter := converter.NewHarToJmxConverter(job.Options)
	jmxData, err := jmxConverter.Convert(har)
	if err != nil {
		jm.updateJobStatus(job, models.JobStatusFailed, 0, fmt.Sprintf("Failed to convert HAR to JMX: %v", err))
		return
	}
	
	jm.updateJobStatus(job, models.JobStatusProcessing, 80, "JMX generated with correlation and parameterization")
	
	// Save JMX to S3
	jmxKey := fmt.Sprintf("%s/output/%s.jmx", job.ID, job.ID)
	jmxURL, err := jm.storage.Upload(jmxKey, jmxData, "application/xml")
	if err != nil {
		jm.updateJobStatus(job, models.JobStatusFailed, 0, fmt.Sprintf("Failed to upload JMX: %v", err))
		return
	}
	job.JmxS3URL = jmxURL
	
	jm.updateJobStatus(job, models.JobStatusProcessing, 95, "JMX file saved to S3")
	
	// Mark job as completed
	jm.updateJobStatus(job, models.JobStatusCompleted, 100, "Conversion completed successfully")
	
	logrus.Infof("Job %s completed successfully", job.ID)
}

// updateJobStatus updates job status and notifies subscribers
func (jm *JobManager) updateJobStatus(job *models.Job, status models.JobStatus, progress int, message string) {
	jm.jobsMutex.Lock()
	job.Status = status
	job.Progress = progress
	job.UpdatedAt = time.Now()
	if status == models.JobStatusFailed {
		job.Error = message
	}
	jm.jobsMutex.Unlock()
	
	// Notify subscribers
	jm.notifySubscribers(job.ID, models.ProgressUpdate{
		Type:     "progress",
		JobID:    job.ID,
		Status:   status,
		Progress: progress,
		Message:  message,
	})
}

// notifySubscribers notifies all subscribers of a job update
func (jm *JobManager) notifySubscribers(jobID string, update models.ProgressUpdate) {
	jm.subMutex.RLock()
	defer jm.subMutex.RUnlock()
	
	if subs, exists := jm.subscribers[jobID]; exists {
		for _, ch := range subs {
			select {
			case ch <- update:
			default:
				// Channel is full, skip
			}
		}
	}
}

// cleanupRoutine periodically cleans up old jobs
func (jm *JobManager) cleanupRoutine() {
	ticker := time.NewTicker(time.Duration(jm.config.Worker.CleanupInterval) * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		jm.cleanupOldJobs()
	}
}

// cleanupOldJobs removes jobs older than retention period
func (jm *JobManager) cleanupOldJobs() {
	jm.jobsMutex.Lock()
	defer jm.jobsMutex.Unlock()
	
	retentionDuration := time.Duration(jm.config.Worker.RetentionPeriod) * time.Hour
	cutoffTime := time.Now().Add(-retentionDuration)
	
	for jobID, job := range jm.jobs {
		if job.CreatedAt.Before(cutoffTime) {
			// Delete files from S3
			go func(id string) {
				prefix := fmt.Sprintf("%s/", id)
				files, err := jm.storage.ListFiles(prefix)
				if err != nil {
					logrus.Errorf("Failed to list files for cleanup: %v", err)
					return
				}
				
				for _, file := range files {
					if file.Key != nil {
						if err := jm.storage.Delete(*file.Key); err != nil {
							logrus.Errorf("Failed to delete file %s: %v", *file.Key, err)
						}
					}
				}
			}(jobID)
			
			// Remove job from memory
			delete(jm.jobs, jobID)
			logrus.Infof("Cleaned up job %s", jobID)
		}
	}
}