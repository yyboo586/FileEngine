package logics

import (
	"FileEngine/common"
	"FileEngine/interfaces"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	config         *common.Config
	dbFile         interfaces.DBFile
	storageAdapter interfaces.StorageAdapter
)

func SetConfig(i *common.Config) {
	config = i
}

func SetDBFile(i interfaces.DBFile) {
	dbFile = i
}

func SetStorageAdapter(i interfaces.StorageAdapter) {
	storageAdapter = i
}

// ValidFileName 检查文件名是否有效
func ValidFileName(filename string) (err error) {
	if filename == "" || len(filename) > 255 {
		return common.NewHTTPError(http.StatusBadRequest, "Invalid filename", []map[string]interface{}{
			{
				"error":   "Invalid filename",
				"message": fmt.Sprintf("file name %s is invalid", filename),
			},
		})
	}

	// 检查是否包含非法字符
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*", "\\", "/"}
	for _, char := range invalidChars {
		if strings.Contains(filename, char) {
			return common.NewHTTPError(http.StatusBadRequest, "Invalid filename", []map[string]interface{}{
				{
					"error":   "Invalid filename",
					"message": fmt.Sprintf("file name %s is invalid", filename),
				},
			})
		}
	}

	return nil
}

// 检查文件大小 (限制为5GB)
func ValidFileSize(size int64) (err error) {
	const maxFileSize = 5 * 1024 * 1024 * 1024
	if size > maxFileSize {
		err = common.NewHTTPError(http.StatusBadRequest, "File size exceeds maximum allowed size", []map[string]interface{}{
			{
				"error":   "File size exceeds maximum allowed size",
				"message": fmt.Sprintf("file size %d exceeds maximum allowed size %d", size, maxFileSize),
			},
		})
		return err
	}

	return nil
}

// 检查文件扩展名
func ValidFileExtension(file *multipart.FileHeader) (err error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{
		".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", // 图片
		".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv", // 视频
		".exe", ".msi", ".dmg", ".pkg", // 应用
		".zip", ".rar", ".7z", ".tar", ".gz", // 压缩包
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", // 文档
		".txt", ".md", ".json", ".xml", ".csv", // 文本
	}

	isAllowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		err = common.NewHTTPError(http.StatusBadRequest, "File extension is not allowed", []map[string]interface{}{
			{
				"error":   "File extension is not allowed",
				"message": fmt.Sprintf("file extension %s is not allowed", ext),
			},
		})
		return err
	}

	return nil
}

// 生成唯一对象名
func generateUniqueObjectName(filename string) string {
	// 移除扩展名
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

	// 生成时间戳
	timestamp := time.Now().Format("20060102150405")

	// 生成随机字符串
	randomStr := uuid.New().String()[:8]

	// 组合对象名: 原文件名_时间戳_随机字符串.扩展名
	return fmt.Sprintf("%s_%s_%s%s", nameWithoutExt, timestamp, randomStr, filepath.Ext(filename))
}
