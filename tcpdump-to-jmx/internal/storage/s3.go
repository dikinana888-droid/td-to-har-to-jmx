package storage

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
	"github.com/tcpdump-to-jmx/internal/config"
)

type S3Storage struct {
	client     *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	bucket     string
}

func NewS3Storage(cfg config.AWSConfig) (*S3Storage, error) {
	awsConfig := &aws.Config{
		Region: aws.String(cfg.Region),
	}

	// Set credentials if provided
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		awsConfig.Credentials = credentials.NewStaticCredentials(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)
	}

	// Set custom endpoint for S3-compatible services
	if cfg.Endpoint != "" {
		awsConfig.Endpoint = aws.String(cfg.Endpoint)
		awsConfig.S3ForcePathStyle = aws.Bool(true)
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	s3Client := s3.New(sess)
	
	// Create bucket if it doesn't exist
	_, err = s3Client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(cfg.BucketName),
	})
	
	if err != nil {
		logrus.Infof("Bucket %s doesn't exist, creating...", cfg.BucketName)
		_, err = s3Client.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(cfg.BucketName),
		})
		if err != nil {
			logrus.Warnf("Failed to create bucket: %v", err)
		}
	}

	return &S3Storage{
		client:     s3Client,
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
		bucket:     cfg.BucketName,
	}, nil
}

// Upload uploads a file to S3
func (s *S3Storage) Upload(key string, data []byte, contentType string) (string, error) {
	input := &s3manager.UploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		Metadata: map[string]*string{
			"uploaded_at": aws.String(time.Now().Format(time.RFC3339)),
		},
	}

	result, err := s.uploader.Upload(input)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return result.Location, nil
}

// UploadStream uploads a stream to S3
func (s *S3Storage) UploadStream(key string, reader io.Reader, contentType string) (string, error) {
	input := &s3manager.UploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
		Metadata: map[string]*string{
			"uploaded_at": aws.String(time.Now().Format(time.RFC3339)),
		},
	}

	result, err := s.uploader.Upload(input)
	if err != nil {
		return "", fmt.Errorf("failed to upload stream: %w", err)
	}

	return result.Location, nil
}

// Download downloads a file from S3
func (s *S3Storage) Download(key string) ([]byte, error) {
	buf := &aws.WriteAtBuffer{}
	
	_, err := s.downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return buf.Bytes(), nil
}

// DownloadStream downloads a file from S3 as a stream
func (s *S3Storage) DownloadStream(key string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return result.Body, nil
}

// Delete deletes a file from S3
func (s *S3Storage) Delete(key string) error {
	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Exists checks if a file exists in S3
func (s *S3Storage) Exists(key string) (bool, error) {
	_, err := s.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	
	if err != nil {
		// Check if the error is because the object doesn't exist
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotFound" {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetPresignedURL generates a presigned URL for downloading
func (s *S3Storage) GetPresignedURL(key string, expiration time.Duration) (string, error) {
	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	urlStr, err := req.Presign(expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return urlStr, nil
}

// ListFiles lists files with a given prefix
func (s *S3Storage) ListFiles(prefix string) ([]*s3.Object, error) {
	result, err := s.client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return result.Contents, nil
}