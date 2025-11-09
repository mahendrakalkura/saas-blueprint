package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/mahendrakalkura/saas-blueprint/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Service struct {
	client *minio.Client
	bucket string
}

type UploadResult struct {
	Key       string
	URL       string
	Size      int64
	MimeType  string
	UploadedAt time.Time
}

func NewService(cfg *config.StorageConfig) (*Service, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// Create bucket if it doesn't exist
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &Service{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// Upload uploads a file to storage
func (s *Service) Upload(ctx context.Context, path string, reader io.Reader, size int64, mimeType string) (*UploadResult, error) {
	// Generate unique key
	key := generateUniqueKey(path)

	_, err := s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: mimeType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &UploadResult{
		Key:        key,
		Size:       size,
		MimeType:   mimeType,
		UploadedAt: time.Now(),
	}, nil
}

// UploadAvatar uploads a user avatar
func (s *Service) UploadAvatar(ctx context.Context, userID string, reader io.Reader, size int64, mimeType string) (*UploadResult, error) {
	path := fmt.Sprintf("avatars/%s", userID)
	return s.Upload(ctx, path, reader, size, mimeType)
}

// GetPresignedURL generates a presigned URL for downloading a file
func (s *Service) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucket, key, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// Delete deletes a file from storage
func (s *Service) Delete(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetObject retrieves a file from storage
func (s *Service) GetObject(ctx context.Context, key string) (*minio.Object, error) {
	object, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return object, nil
}

// ListObjects lists objects with a given prefix
func (s *Service) ListObjects(ctx context.Context, prefix string) <-chan minio.ObjectInfo {
	return s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})
}

// generateUniqueKey generates a unique key for storage
func generateUniqueKey(path string) string {
	ext := filepath.Ext(path)
	id := uuid.New().String()
	dir := filepath.Dir(path)

	if dir == "." || dir == "" {
		return fmt.Sprintf("%s%s", id, ext)
	}

	return fmt.Sprintf("%s/%s%s", dir, id, ext)
}
