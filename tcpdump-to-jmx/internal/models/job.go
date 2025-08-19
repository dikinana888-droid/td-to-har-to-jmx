package models

import (
	"time"
)

// JobStatus represents the status of a conversion job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// Job represents a conversion job
type Job struct {
	ID        string            `json:"id"`
	Status    JobStatus         `json:"status"`
	Progress  int               `json:"progress"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	FileName  string            `json:"file_name"`
	FileSize  int64             `json:"file_size"`
	Options   ConversionOptions `json:"options"`
	HarS3URL  string            `json:"har_s3_url,omitempty"`
	JmxS3URL  string            `json:"jmx_s3_url,omitempty"`
	Error     string            `json:"error,omitempty"`
}

// ConversionOptions represents options for conversion
type ConversionOptions struct {
	FilterPort             string `json:"filter_port"`
	FilterHost             string `json:"filter_host"`
	EnableCorrelation      bool   `json:"enable_correlation"`
	EnableParameterization bool   `json:"enable_parameterization"`
	ThreadCount            string `json:"thread_count"`
	RampUpTime             string `json:"ramp_up_time"`
	LoopCount              string `json:"loop_count"`
}

// ProgressUpdate represents a progress update for WebSocket
type ProgressUpdate struct {
	Type     string    `json:"type"`
	JobID    string    `json:"job_id"`
	Status   JobStatus `json:"status"`
	Progress int       `json:"progress"`
	Message  string    `json:"message,omitempty"`
	Error    string    `json:"error,omitempty"`
}