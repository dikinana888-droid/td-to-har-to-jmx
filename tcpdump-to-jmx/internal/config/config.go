package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server ServerConfig
	AWS    AWSConfig
	Worker WorkerConfig
}

type ServerConfig struct {
	Port         int
	Mode         string
	ReadTimeout  int
	WriteTimeout int
	MaxFileSize  int64
}

type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Endpoint        string // For S3-compatible services like MinIO
}

type WorkerConfig struct {
	MaxWorkers       int
	JobTimeout       int // in seconds
	CleanupInterval  int // in seconds
	RetentionPeriod  int // in hours
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			Mode:         getEnv("GIN_MODE", "debug"),
			ReadTimeout:  getEnvAsInt("SERVER_READ_TIMEOUT", 60),
			WriteTimeout: getEnvAsInt("SERVER_WRITE_TIMEOUT", 60),
			MaxFileSize:  getEnvAsInt64("MAX_FILE_SIZE", 500*1024*1024), // 500MB default
		},
		AWS: AWSConfig{
			Region:          getEnv("AWS_REGION", "us-east-1"),
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			BucketName:      getEnv("S3_BUCKET_NAME", "tcpdump-conversions"),
			Endpoint:        getEnv("S3_ENDPOINT", ""),
		},
		Worker: WorkerConfig{
			MaxWorkers:      getEnvAsInt("MAX_WORKERS", 10),
			JobTimeout:      getEnvAsInt("JOB_TIMEOUT", 3600),      // 1 hour
			CleanupInterval: getEnvAsInt("CLEANUP_INTERVAL", 3600), // 1 hour
			RetentionPeriod: getEnvAsInt("RETENTION_PERIOD", 168),  // 7 days
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultValue
}