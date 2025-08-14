package interfaces

import (
	"context"
)

type DBFile interface {
	// 创建文件记录
	CreateFile(ctx context.Context, file *FileInfo) error
	// 根据ID获取文件
	GetFileByID(ctx context.Context, fileID string) (*FileInfo, error)
	// 根据名称获取文件
	GetFileByName(ctx context.Context, name string) (*FileInfo, error)
	// 删除文件记录
	DeleteFile(ctx context.Context, fileID string) error
	// 获取文件列表
	GetFileList(ctx context.Context, bucketID string, page, pageSize int) ([]*FileInfo, int64, error)
}
