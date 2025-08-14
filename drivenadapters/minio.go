package drivenadapters

import (
	"FileEngine/interfaces"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

type MinioAdapter struct {
	client   *minio.Client
	bucketID string
}

func NewMinioAdapter() interfaces.StorageAdapter {
	return &MinioAdapter{
		client:   minioClient,
		bucketID: config.Minio.BucketID,
	}
}

func (m *MinioAdapter) Upload(ctx context.Context, bucketID, objectName string, reader io.Reader, size int64, contentType string) error {
	// 确保bucket存在
	exists, err := m.client.BucketExists(ctx, bucketID)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = m.client.MakeBucket(ctx, bucketID, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	// 上传文件
	_, err = m.client.PutObject(ctx, bucketID, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})

	return err
}

func (m *MinioAdapter) Download(ctx context.Context, bucketID, objectName string) (io.ReadCloser, error) {
	obj, err := m.client.GetObject(ctx, bucketID, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return obj, nil
}

func (m *MinioAdapter) Delete(ctx context.Context, bucketID, objectName string) error {
	return m.client.RemoveObject(ctx, bucketID, objectName, minio.RemoveObjectOptions{})
}

func (m *MinioAdapter) FileExists(ctx context.Context, bucketID, objectName string) (bool, error) {
	_, err := m.client.StatObject(ctx, bucketID, objectName, minio.StatObjectOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (m *MinioAdapter) GetFileInfo(ctx context.Context, bucketID, objectName string) (*interfaces.StorageFileInfo, error) {
	info, err := m.client.StatObject(ctx, bucketID, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}

	return &interfaces.StorageFileInfo{
		Size:         info.Size,
		ContentType:  info.ContentType,
		LastModified: info.LastModified.Format("2006-01-02 15:04:05"),
		ETag:         info.ETag,
	}, nil
}

// 生成预签名下载URL
func (m *MinioAdapter) GeneratePresignedDownloadURL(ctx context.Context, bucketID, objectName string, expiration time.Duration) (string, error) {
	// 检查bucket是否存在
	exists, err := m.client.BucketExists(ctx, bucketID)
	if err != nil {
		return "", fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		return "", fmt.Errorf("bucket %s does not exist", bucketID)
	}

	// 检查对象是否存在
	exists, err = m.FileExists(ctx, bucketID, objectName)
	if err != nil {
		return "", fmt.Errorf("failed to check object existence: %w", err)
	}

	if !exists {
		return "", fmt.Errorf("object %s does not exist in bucket %s", objectName, bucketID)
	}

	// 生成预签名URL
	presignedURL, err := m.client.PresignedGetObject(ctx, bucketID, objectName, expiration, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

// 生成预签名上传URL
func (m *MinioAdapter) GeneratePresignedUploadURL(ctx context.Context, bucketID, objectName string, expiration time.Duration) (string, error) {
	// 检查bucket是否存在，不存在则创建
	exists, err := m.client.BucketExists(ctx, bucketID)
	if err != nil {
		return "", fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = m.client.MakeBucket(ctx, bucketID, minio.MakeBucketOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	// 生成预签名上传URL
	presignedURL, err := m.client.PresignedPutObject(ctx, bucketID, objectName, expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return presignedURL.String(), nil
}
