package api

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 文件上传相关常量
const (
	maxUploadSize = 10 << 20 // 10MB
	uploadBaseDir = "uploads"
)

// 允许的文件类型
var allowedMimeTypes = map[string]string{
	".pdf":  "application/pdf",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".txt":  "text/plain",
	".md":   "text/markdown",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
}

// HandleUpload 通用文件上传
// POST /api/upload
func (h *Handler) HandleUpload(c *gin.Context) {
	// 限制请求体大小
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		if err.Error() == "http: request body too large" {
			Error(c, http.StatusBadRequest, 40036, "文件大小超出限制（最大 10MB）")
			return
		}
		Error(c, http.StatusBadRequest, 40004, "请求参数无效: "+err.Error())
		return
	}
	defer file.Close()

	// 检查文件大小
	if header.Size > maxUploadSize {
		Error(c, http.StatusBadRequest, 40036, "文件大小超出限制（最大 10MB）")
		return
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(header.Filename))
	mimeType, ok := allowedMimeTypes[ext]
	if !ok {
		Error(c, http.StatusBadRequest, 40035, fmt.Sprintf("不支持的文件类型: %s，支持: PDF, DOCX, TXT, MD, JPG, JPEG, PNG", ext))
		return
	}

	// 读取文件内容用于计算哈希
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "读取文件失败")
		return
	}

	// 生成文件名：原始名_hash.ext
	hash := fmt.Sprintf("%x", md5.Sum(fileBytes))[:8]
	baseName := strings.TrimSuffix(header.Filename, ext)
	// 清理文件名中的特殊字符
	baseName = sanitizeFilename(baseName)
	if baseName == "" {
		baseName = "file"
	}
	newFilename := fmt.Sprintf("%s_%s%s", baseName, hash, ext)

	// 构建存储路径：uploads/{year}/{month}/{filename}
	now := time.Now()
	relDir := filepath.Join(uploadBaseDir, fmt.Sprintf("%d", now.Year()), fmt.Sprintf("%02d", now.Month()))
	absDir := relDir

	// 创建目录
	if err := os.MkdirAll(absDir, 0755); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建上传目录失败")
		return
	}

	// 保存文件
	absPath := filepath.Join(absDir, newFilename)
	if err := os.WriteFile(absPath, fileBytes, 0644); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "保存文件失败")
		return
	}

	// 返回文件信息
	fileURL := "/" + filepath.Join(relDir, newFilename)

	Success(c, gin.H{
		"url":           fileURL,
		"filename":      newFilename,
		"original_name": header.Filename,
		"size":          header.Size,
		"mime_type":     mimeType,
	})
}

// sanitizeFilename 清理文件名中的特殊字符
func sanitizeFilename(name string) string {
	// 只保留字母、数字、中文、下划线、短横线
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '_' || r == '-' || (r >= 0x4e00 && r <= 0x9fff) {
			result.WriteRune(r)
		}
	}
	return result.String()
}
