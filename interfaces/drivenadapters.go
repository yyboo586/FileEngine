package interfaces

import (
	"context"
	"io"
	"time"
)

type StorageAdapter interface {
	// 上传文件到存储
	Upload(ctx context.Context, bucketID, objectName string, reader io.Reader, size int64, contentType string) error
	// 从存储下载文件
	Download(ctx context.Context, bucketID, objectName string) (io.ReadCloser, error)
	// 从存储删除文件
	Delete(ctx context.Context, bucketID, objectName string) error
	// 检查文件是否存在
	FileExists(ctx context.Context, bucketID, objectName string) (bool, error)
	// 获取文件信息
	GetFileInfo(ctx context.Context, bucketID, objectName string) (*StorageFileInfo, error)

	// 新增：生成预签名下载URL
	GeneratePresignedDownloadURL(ctx context.Context, bucketID, objectName string, expiration time.Duration) (string, error)
	// 新增：生成预签名上传URL
	GeneratePresignedUploadURL(ctx context.Context, bucketID, objectName string, expiration time.Duration) (string, error)
}

type StorageFileInfo struct {
	Size         int64
	ContentType  string
	LastModified string
	ETag         string
}
