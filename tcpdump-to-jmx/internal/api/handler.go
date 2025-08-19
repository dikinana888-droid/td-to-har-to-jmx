package api

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/tcpdump-to-jmx/internal/config"
	"github.com/tcpdump-to-jmx/internal/models"
	"github.com/tcpdump-to-jmx/internal/storage"
	"github.com/tcpdump-to-jmx/internal/worker"
)

type Handler struct {
	storage     *storage.S3Storage
	config      *config.Config
	jobManager  *worker.JobManager
	wsUpgrader  websocket.Upgrader
}

func NewHandler(storage *storage.S3Storage, config *config.Config) *Handler {
	return &Handler{
		storage:    storage,
		config:     config,
		jobManager: worker.NewJobManager(storage, config),
		wsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
	}
}

// HealthCheck returns the health status of the service
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"version": "1.0.0",
		"time":    time.Now().UTC(),
	})
}

// ConvertTCPDump handles the TCP dump file upload and conversion
func (h *Handler) ConvertTCPDump(c *gin.Context) {
	// Parse multipart form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to get file from request",
		})
		return
	}
	defer file.Close()

	// Check file size
	if header.Size > h.config.Server.MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("File size exceeds maximum allowed size of %d MB", h.config.Server.MaxFileSize/(1024*1024)),
		})
		return
	}

	// Check file extension
	ext := filepath.Ext(header.Filename)
	if ext != ".pcap" && ext != ".pcapng" && ext != ".cap" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file format. Only .pcap, .pcapng, and .cap files are supported",
		})
		return
	}

	// Parse conversion options
	options := models.ConversionOptions{
		FilterPort:          c.DefaultQuery("port", ""),
		FilterHost:          c.DefaultQuery("host", ""),
		EnableCorrelation:   c.DefaultQuery("correlation", "true") == "true",
		EnableParameterization: c.DefaultQuery("parameterization", "true") == "true",
		ThreadCount:         c.DefaultQuery("threads", "10"),
		RampUpTime:          c.DefaultQuery("rampup", "10"),
		LoopCount:           c.DefaultQuery("loops", "1"),
	}

	// Generate job ID
	jobID := uuid.New().String()

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read file content",
		})
		return
	}

	// Save original file to S3
	originalKey := fmt.Sprintf("%s/original/%s", jobID, header.Filename)
	_, err = h.storage.Upload(originalKey, fileContent, "application/octet-stream")
	if err != nil {
		logrus.Errorf("Failed to upload original file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save original file",
		})
		return
	}

	// Create job
	job := &models.Job{
		ID:        jobID,
		Status:    models.JobStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		FileName:  header.Filename,
		FileSize:  header.Size,
		Options:   options,
	}

	// Start conversion job
	h.jobManager.SubmitJob(job, fileContent)

	// Return job information
	c.JSON(http.StatusAccepted, gin.H{
		"job_id":  jobID,
		"status":  job.Status,
		"message": "Conversion job started successfully",
		"links": gin.H{
			"status":    fmt.Sprintf("/api/v1/status/%s", jobID),
			"websocket": fmt.Sprintf("/api/v1/ws/%s", jobID),
		},
	})
}

// GetJobStatus returns the status of a conversion job
func (h *Handler) GetJobStatus(c *gin.Context) {
	jobID := c.Param("jobId")
	
	job, exists := h.jobManager.GetJob(jobID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
		})
		return
	}

	response := gin.H{
		"job_id":     job.ID,
		"status":     job.Status,
		"progress":   job.Progress,
		"created_at": job.CreatedAt,
		"updated_at": job.UpdatedAt,
		"file_name":  job.FileName,
		"file_size":  job.FileSize,
	}

	if job.Status == models.JobStatusCompleted {
		response["har_file"] = fmt.Sprintf("/api/v1/download/%s/har", jobID)
		response["jmx_file"] = fmt.Sprintf("/api/v1/download/%s/jmx", jobID)
		response["har_s3_url"] = job.HarS3URL
		response["jmx_s3_url"] = job.JmxS3URL
	}

	if job.Status == models.JobStatusFailed {
		response["error"] = job.Error
	}

	c.JSON(http.StatusOK, response)
}

// DownloadFile handles file download requests
func (h *Handler) DownloadFile(c *gin.Context) {
	jobID := c.Param("jobId")
	fileType := c.Param("fileType")

	job, exists := h.jobManager.GetJob(jobID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
		})
		return
	}

	if job.Status != models.JobStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job is not completed yet",
		})
		return
	}

	var key string
	var filename string
	var contentType string

	switch fileType {
	case "har":
		key = fmt.Sprintf("%s/output/%s.har", jobID, jobID)
		filename = fmt.Sprintf("%s.har", jobID)
		contentType = "application/json"
	case "jmx":
		key = fmt.Sprintf("%s/output/%s.jmx", jobID, jobID)
		filename = fmt.Sprintf("%s.jmx", jobID)
		contentType = "application/xml"
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file type. Use 'har' or 'jmx'",
		})
		return
	}

	// Generate presigned URL for direct download
	presignedURL, err := h.storage.GetPresignedURL(key, 1*time.Hour)
	if err != nil {
		// If presigned URL fails, try direct download
		stream, err := h.storage.DownloadStream(key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to download file",
			})
			return
		}
		defer stream.Close()

		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Header("Content-Type", contentType)
		io.Copy(c.Writer, stream)
		return
	}

	// Redirect to presigned URL
	c.Redirect(http.StatusTemporaryRedirect, presignedURL)
}

// WebSocketHandler handles WebSocket connections for real-time progress updates
func (h *Handler) WebSocketHandler(c *gin.Context) {
	jobID := c.Param("jobId")
	
	conn, err := h.wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Errorf("Failed to upgrade WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Subscribe to job updates
	updates := h.jobManager.Subscribe(jobID)
	defer h.jobManager.Unsubscribe(jobID, updates)

	// Send initial job status
	job, exists := h.jobManager.GetJob(jobID)
	if exists {
		conn.WriteJSON(gin.H{
			"type":     "status",
			"job_id":   job.ID,
			"status":   job.Status,
			"progress": job.Progress,
		})
	}

	// Listen for updates
	for update := range updates {
		if err := conn.WriteJSON(update); err != nil {
			logrus.Errorf("Failed to write WebSocket message: %v", err)
			break
		}
	}
}

// ListConversions returns a list of recent conversions
func (h *Handler) ListConversions(c *gin.Context) {
	limit := c.DefaultQuery("limit", "20")
	offset := c.DefaultQuery("offset", "0")
	
	jobs := h.jobManager.ListJobs(limit, offset)
	
	c.JSON(http.StatusOK, gin.H{
		"jobs":   jobs,
		"total":  len(jobs),
		"limit":  limit,
		"offset": offset,
	})
}