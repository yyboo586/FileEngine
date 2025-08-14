package logics

import (
	"FileEngine/common"
	"FileEngine/interfaces"
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type LogicsFile struct {
	uploadTimeout   time.Duration
	downloadTimeout time.Duration
	defaultBucketID string
	dbFile          interfaces.DBFile
	storage         interfaces.StorageAdapter
}

var (
	logicsFileOnce sync.Once
	logicsFile     *LogicsFile
)

func NewLogicsFile() interfaces.LogicsFile {
	logicsFileOnce.Do(func() {
		logicsFile = &LogicsFile{
			uploadTimeout:   config.Server.UploadTimeout,
			downloadTimeout: config.Server.DownloadTimeout,
			defaultBucketID: config.Minio.BucketID,
			dbFile:          dbFile,
			storage:         storageAdapter,
		}
	})
	return logicsFile
}

func (l *LogicsFile) Upload(ctx context.Context, file *multipart.FileHeader) (fileInfo *interfaces.FileInfo, err error) {
	log.Printf("[DEBUG] file header: %+v", file.Header)
	// 文件校验
	if err = l.validateFile(file); err != nil {
		return
	}

	// 2. 生成唯一文件名
	originalName := filepath.Base(file.Filename)

	// 检查文件是否已存在
	existingFile, err := l.dbFile.GetFileByName(ctx, originalName)
	if err == nil && existingFile != nil {
		err = common.NewHTTPError(http.StatusBadRequest, "File with name already exists", []map[string]interface{}{
			{
				"error":   "File with name already exists",
				"message": fmt.Sprintf("file with name %s already exists", originalName),
			},
		})
		return
	}

	// 上传到MinIO
	// 打开文件
	src, err := file.Open()
	if err != nil {
		err = common.NewHTTPError(http.StatusInternalServerError, "Failed to open uploaded file", []map[string]interface{}{
			{
				"error":   "Failed to open uploaded file",
				"message": err.Error(),
			},
		})
		return
	}
	defer src.Close()

	// 获取文件大小
	fileSize := file.Size
	// 获取Content-Type
	contentType := interfaces.GetContentType(file)
	// 上传到存储
	err = l.storage.Upload(ctx, l.defaultBucketID, originalName, src, fileSize, contentType)
	if err != nil {
		err = common.NewHTTPError(http.StatusInternalServerError, "Failed to upload file to storage", []map[string]interface{}{
			{
				"error":   "Failed to upload file to storage",
				"message": err.Error(),
			},
		})
		return
	}

	// 创建数据库记录
	fileInfo = &interfaces.FileInfo{
		ID:          uuid.New().String(),
		Name:        originalName,
		BucketID:    l.defaultBucketID,
		Icon:        "",
		Size:        fileSize,
		ContentType: contentType,
	}

	err = l.dbFile.CreateFile(ctx, fileInfo)
	if err != nil {
		// 如果数据库插入失败，需要从存储中删除已上传的文件
		l.storage.Delete(ctx, l.defaultBucketID, originalName)
		err = common.NewHTTPError(http.StatusInternalServerError, "Failed to create file record", []map[string]interface{}{
			{
				"error":   "Failed to create file record",
				"message": err.Error(),
			},
		})
		return
	}

	return fileInfo, nil
}

func (l *LogicsFile) Download(ctx context.Context, fileID string) (fileDownloadInfo *interfaces.FileDownload, err error) {
	// 从数据库获取文件信息
	fileInfo, err := l.dbFile.GetFileByID(ctx, fileID)
	if err != nil {
		err = common.NewHTTPError(http.StatusNotFound, "File not found", []map[string]interface{}{
			{
				"error":   "File not found",
				"message": err.Error(),
			},
		})
		return
	}

	// 检查存储中文件是否存在
	exists, err := l.storage.FileExists(ctx, fileInfo.BucketID, fileInfo.Name)
	if err != nil {
		err = common.NewHTTPError(http.StatusInternalServerError, "Failed to check file existence", []map[string]interface{}{
			{
				"error":   "Failed to check file existence",
				"message": err.Error(),
			},
		})
		return
	}
	if !exists {
		err = common.NewHTTPError(http.StatusNotFound, "File not found in storage", nil)
	}

	fileReaderCloser, err := l.storage.Download(ctx, fileInfo.BucketID, fileInfo.Name)
	if err != nil {
		err = common.NewHTTPError(http.StatusInternalServerError, "Failed to download file", []map[string]interface{}{
			{
				"error":   "Failed to download file",
				"message": err.Error(),
			},
		})
		return
	}

	fileDownloadInfo = &interfaces.FileDownload{
		File:   fileInfo,
		Reader: fileReaderCloser,
	}
	return
}

// 生成预签名上传URL
func (l *LogicsFile) GenerateUploadURL(ctx context.Context, filename string, contentType string, size int64) (*interfaces.UploadURL, error) {
	// 文件校验
	if err := l.validateFileInfo(filename, contentType, size); err != nil {
		return nil, err
	}

	// 生成预签名上传URL
	presignedURL, err := l.storage.GeneratePresignedUploadURL(ctx, l.defaultBucketID, filename, l.uploadTimeout)
	if err != nil {
		return nil, common.NewHTTPError(http.StatusInternalServerError, "Failed to generate upload URL", []map[string]interface{}{
			{"error": "Failed to generate upload URL", "message": err.Error()},
		})
	}
	presignedURL, err = url.QueryUnescape(presignedURL) // presignedURL包含%2F，需要解码
	if err != nil {
		return nil, common.NewHTTPError(http.StatusInternalServerError, "Failed to unescape upload URL", []map[string]interface{}{
			{"error": "Failed to unescape upload URL", "message": err.Error()},
		})
	}

	// 创建数据库记录
	fileInfo := &interfaces.FileInfo{
		ID:          uuid.New().String(),
		Name:        filename,
		BucketID:    l.defaultBucketID,
		Icon:        "",
		Size:        size,
		ContentType: contentType,
	}
	err = l.dbFile.CreateFile(ctx, fileInfo)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, common.NewHTTPError(http.StatusBadRequest, "File with name already exists", []map[string]interface{}{
				{"error": "File with name already exists", "message": fmt.Sprintf("file with name %s already exists", filename)},
			})
		}
		return nil, common.NewHTTPError(http.StatusInternalServerError, "Failed to create file record", []map[string]interface{}{
			{"error": "Failed to create file record", "message": err.Error()},
		})
	}

	// 计算过期时间
	expiresAt := time.Now().Add(l.uploadTimeout)
	expiresIn := int64(l.uploadTimeout.Seconds())

	return &interfaces.UploadURL{
		ID:        fileInfo.ID,
		URL:       presignedURL,
		ExpiresAt: expiresAt,
		ExpiresIn: expiresIn,
	}, nil
}

// 生成预签名下载URL
func (l *LogicsFile) GenerateDownloadURL(ctx context.Context, fileID string) (*interfaces.DownloadURL, error) {
	// 权限检查（可以添加用户权限验证）
	if err := l.checkDownloadPermission(ctx, fileID); err != nil {
		return nil, err
	}

	// 获取文件信息
	fileInfo, err := l.dbFile.GetFileByID(ctx, fileID)
	if err != nil {
		return nil, common.NewHTTPError(http.StatusNotFound, "File not found", []map[string]interface{}{
			{"error": "File not found", "message": err.Error()},
		})
	}

	// 检查存储中文件是否存在
	exists, err := l.storage.FileExists(ctx, fileInfo.BucketID, fileInfo.Name)
	if err != nil {
		return nil, common.NewHTTPError(http.StatusInternalServerError, "Failed to check file existence", []map[string]interface{}{
			{"error": "Failed to check file existence", "message": err.Error()},
		})
	}

	if !exists {
		return nil, common.NewHTTPError(http.StatusNotFound, "File not found in storage", nil)
	}

	// 生成预签名URL
	presignedURL, err := l.storage.GeneratePresignedDownloadURL(ctx, fileInfo.BucketID, fileInfo.Name, l.downloadTimeout)
	if err != nil {
		return nil, common.NewHTTPError(http.StatusInternalServerError, "Failed to generate download URL", []map[string]interface{}{
			{"error": "Failed to generate download URL", "message": err.Error()},
		})
	}
	// 计算过期时间
	expiresAt := time.Now().Add(l.downloadTimeout)
	expiresIn := int64(l.downloadTimeout.Seconds())

	return &interfaces.DownloadURL{
		URL:            presignedURL,
		ExpiresAt:      expiresAt,
		ExpiresIn:      expiresIn,
		FileInfo:       fileInfo,
		DirectDownload: true,
	}, nil
}

func (l *LogicsFile) Delete(ctx context.Context, fileID string) error {
	// 从数据库获取文件信息
	fileInfo, err := l.dbFile.GetFileByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// 从存储中删除文件
	err = l.storage.Delete(ctx, fileInfo.BucketID, fileInfo.Name)
	if err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}

	// 从数据库删除记录
	err = l.dbFile.DeleteFile(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to delete file record: %w", err)
	}

	return nil
}

func (l *LogicsFile) GetMeta(ctx context.Context, fileID string) (*interfaces.FileInfo, error) {
	return l.dbFile.GetFileByID(ctx, fileID)
}

func (l *LogicsFile) GetList(ctx context.Context, page, pageSize int) ([]*interfaces.FileInfo, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return l.dbFile.GetFileList(ctx, l.defaultBucketID, page, pageSize)
}

// 文件校验
func (l *LogicsFile) validateFile(file *multipart.FileHeader) (err error) {
	// 检查文件名是否合法
	if err = ValidFileName(file.Filename); err != nil {
		return
	}

	// 检查文件大小
	if err = ValidFileSize(file.Size); err != nil {
		return
	}

	// 检查文件扩展名
	if err = ValidFileExtension(file); err != nil {
		return
	}

	return nil
}

// 权限检查（示例实现）
func (l *LogicsFile) checkDownloadPermission(ctx context.Context, fileID string) error {
	// 这里可以添加：
	// - 用户身份验证
	// - 文件访问权限检查
	// - 下载频率限制
	// - 文件状态检查（是否被锁定、删除等）

	// 暂时返回nil，表示允许下载
	return nil
}

// 验证文件信息
func (l *LogicsFile) validateFileInfo(filename string, contentType string, size int64) (err error) {
	err = ValidFileName(filename)
	if err != nil {
		return err
	}

	err = ValidFileSize(size)
	if err != nil {
		return err
	}

	return nil
}
