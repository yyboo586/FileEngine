package driveradapters

import (
	"FileEngine/common"
	"FileEngine/interfaces"
	"FileEngine/logics"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	fileHandlerOnce sync.Once
	fileHandler     *FileHandler
)

type FileHandler struct {
	logicsFile interfaces.LogicsFile
}

func NewFileHandler() interfaces.RESTHandler {
	fileHandlerOnce.Do(func() {
		fileHandler = &FileHandler{
			logicsFile: logics.NewLogicsFile(),
		}
	})
	return fileHandler
}

func (handler *FileHandler) RegisterPublic(engine *gin.Engine) {
	engine.Use(handler.authMiddleware())
	engine.POST("/api/v1/file-engine/files", handler.uploadFile)
	engine.GET("/api/v1/file-engine/files/:fileID", handler.downloadFile)

	engine.POST("/api/v2/file-engine/files", handler.getUploadURL)
	engine.GET("/api/v2/file-engine/files/:fileID", handler.getDownloadURL)

	engine.GET("/api/v1/file-engine/files/:fileID/meta", handler.getFileMeta)
	engine.DELETE("/api/v1/file-engine/files/:fileID", handler.deleteFile)
}

func (handler *FileHandler) RegisterPrivate(engine *gin.Engine) {
}

// 文件上传
func (handler *FileHandler) uploadFile(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		err = common.NewHTTPError(http.StatusBadRequest, "No file uploaded, please select a file to upload", nil)
		common.ReplyError(c, err)
		return
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 调用业务逻辑上传文件
	fileInfo, err := handler.logicsFile.Upload(ctx, file)
	if err != nil {
		common.ReplyError(c, err)
		return
	}

	data := map[string]interface{}{
		"id":           fileInfo.ID,
		"name":         fileInfo.Name,
		"content_type": fileInfo.ContentType,
		"size":         fileInfo.Size,
		"icon":         fileInfo.Icon,
		"create_time":  fileInfo.CreateTime.Format("2006-01-02 15:04:05"),
		"update_time":  fileInfo.UpdateTime.Format("2006-01-02 15:04:05"),
	}
	common.ReplyOK(c, http.StatusOK, data)
}

// 获取文件信息
func (handler *FileHandler) getFileMeta(c *gin.Context) {
	fileID := c.Param("fileID")
	if fileID == "" {
		err := common.NewHTTPError(http.StatusBadRequest, "File ID is required", nil)
		common.ReplyError(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 从数据库获取文件信息
	fileInfo, err := handler.logicsFile.GetMeta(ctx, fileID)
	if err != nil {
		common.ReplyError(c, err)
		return
	}

	data := map[string]interface{}{
		"id":           fileInfo.ID,
		"name":         fileInfo.Name,
		"content_type": fileInfo.ContentType,
		"size":         fileInfo.Size,
		"icon":         fileInfo.Icon,
		"create_time":  fileInfo.CreateTime.Format("2006-01-02 15:04:05"),
		"update_time":  fileInfo.UpdateTime.Format("2006-01-02 15:04:05"),
	}
	common.ReplyOK(c, http.StatusOK, data)
}

// 文件下载
func (handler *FileHandler) downloadFile(c *gin.Context) {
	fileID := c.Param("fileID")
	if fileID == "" {
		err := common.NewHTTPError(http.StatusBadRequest, "File ID is required", nil)
		common.ReplyError(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 获取文件信息
	fileDownloadInfo, err := handler.logicsFile.Download(ctx, fileID)
	if err != nil {
		common.ReplyError(c, err)
		return
	}
	defer fileDownloadInfo.Close()

	extraHeaders := map[string]string{
		"Content-Disposition": fmt.Sprintf("attachment; filename=%s", fileDownloadInfo.File.Name),
	}
	c.DataFromReader(http.StatusOK, fileDownloadInfo.File.Size, fileDownloadInfo.File.ContentType, fileDownloadInfo.Reader, extraHeaders)
}

// 删除文件
func (handler *FileHandler) deleteFile(c *gin.Context) {
	fileID := c.Param("fileID")
	if fileID == "" {
		err := common.NewHTTPError(http.StatusBadRequest, "File ID is required", nil)
		common.ReplyError(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 删除文件
	err := handler.logicsFile.Delete(ctx, fileID)
	if err != nil {
		common.ReplyError(c, err)
		return
	}

	common.ReplyOK(c, http.StatusOK, nil)
}

// 获取预签名上传URL
func (handler *FileHandler) getUploadURL(c *gin.Context) {
	var request struct {
		Filename    string `json:"filename" binding:"required"`
		ContentType string `json:"content_type"`
		Size        int64  `json:"size" binding:"required"`
		Expires     int    `json:"expires"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		err := common.NewHTTPError(http.StatusBadRequest, "Invalid request parameters", []map[string]interface{}{
			{"error": "Invalid request parameters", "message": err.Error()},
		})
		common.ReplyError(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 生成上传URL
	uploadURL, err := handler.logicsFile.GenerateUploadURL(ctx, request.Filename, request.ContentType, request.Size)
	if err != nil {
		common.ReplyError(c, err)
		return
	}

	data := map[string]interface{}{
		"id":         uploadURL.ID,
		"url":        uploadURL.URL,
		"expires_at": uploadURL.ExpiresAt.Format("2006-01-02 15:04:05"),
		"expires_in": uploadURL.ExpiresIn,
	}
	common.ReplyOK(c, http.StatusOK, data)
}

// 获取预签名下载URL
func (handler *FileHandler) getDownloadURL(c *gin.Context) {
	fileID := c.Param("fileID")
	if fileID == "" {
		err := common.NewHTTPError(http.StatusBadRequest, "File ID is required", nil)
		common.ReplyError(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 生成下载URL
	downloadURL, err := handler.logicsFile.GenerateDownloadURL(ctx, fileID)
	if err != nil {
		common.ReplyError(c, err)
		return
	}

	data := map[string]interface{}{
		"url":             downloadURL.URL,
		"expires_at":      downloadURL.ExpiresAt.Format("2006-01-02 15:04:05"),
		"expires_in":      downloadURL.ExpiresIn,
		"direct_download": downloadURL.DirectDownload,
	}
	common.ReplyOK(c, http.StatusOK, data)
}

// 认证中间件
func (handler *FileHandler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以添加JWT token验证等认证逻辑
		// 暂时跳过认证
		c.Next()
	}
}
