package interfaces

import (
	"context"
	"io"
	"mime/multipart"
	"time"
)

type LogicsFile interface {
	// 上传文件
	Upload(ctx context.Context, file *multipart.FileHeader) (*FileInfo, error)
	// 下载文件
	Download(ctx context.Context, fileID string) (*FileDownload, error)
	// 生成预签名上传URL
	GenerateUploadURL(ctx context.Context, filename string, contentType string, size int64) (*UploadURL, error)
	// 生成预签名下载URL
	GenerateDownloadURL(ctx context.Context, fileID string) (*DownloadURL, error)

	// 删除文件
	Delete(ctx context.Context, fileID string) error
	// GetMeta
	GetMeta(ctx context.Context, fileID string) (*FileInfo, error)
	// 获取文件列表
	GetList(ctx context.Context, page, pageSize int) ([]*FileInfo, int64, error)
}

// 下载URL信息
type DownloadURL struct {
	URL            string    `json:"url"`
	ExpiresAt      time.Time `json:"expires_at"`
	ExpiresIn      int64     `json:"expires_in"` // 过期时间（秒）
	FileInfo       *FileInfo `json:"file_info"`
	DirectDownload bool      `json:"direct_download"` // 是否支持直接下载
}

// 上传URL信息
type UploadURL struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
	ExpiresIn int64     `json:"expires_in"` // 过期时间（秒）
}

type FileInfo struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	ContentType string     `json:"content_type"`
	BucketID    string     `json:"bucket_id"`
	Size        int64      `json:"size"`
	Icon        string     `json:"icon"`
	CreateTime  *time.Time `json:"create_time"`
	UpdateTime  *time.Time `json:"update_time"`
}

// 组合模式的简单实现
type FileDownload struct {
	File   *FileInfo
	Reader io.ReadCloser
}

func (f *FileDownload) Close() {
	if f.Reader != nil {
		f.Reader.Close()
	}
}

func GetContentType(file *multipart.FileHeader) string {
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return contentType
}
