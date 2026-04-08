package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"digital-twin/src/backend/database"
	"digital-twin/src/harness/core"
	"digital-twin/src/plugins/auth"

	"github.com/gin-gonic/gin"
)

// H5Handler H5相关处理器
type H5Handler struct {
	authPlugin *auth.AuthPlugin
}

// NewH5Handler 创建H5处理器
func NewH5Handler(authPlugin *auth.AuthPlugin) *H5Handler {
	return &H5Handler{
		authPlugin: authPlugin,
	}
}

// HandleWxH5LoginURL H5微信登录URL
func (h *H5Handler) HandleWxH5LoginURL(c *gin.Context) {
	redirectURI := c.Query("redirect_uri")
	state := c.Query("state")

	if redirectURI == "" {
		Error(c, http.StatusBadRequest, 40004, "缺少 redirect_uri 参数")
		return
	}

	output, err := h.authPlugin.Execute(context.Background(), &core.PluginInput{
		Data: map[string]interface{}{
			"action":       "wx-h5-login-url",
			"redirect_uri": redirectURI,
			"state":        state,
		},
	})

	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "生成授权URL失败")
		return
	}

	if !output.Success {
		Error(c, http.StatusBadRequest, 40004, output.Error)
		return
	}

	Success(c, gin.H{
		"login_url": output.Data["auth_url"],
	})
}

// HandleWxH5Callback H5微信授权回调
func (h *H5Handler) HandleWxH5Callback(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40004, "缺少 code 参数")
		return
	}

	output, err := h.authPlugin.Execute(context.Background(), &core.PluginInput{
		Data: map[string]interface{}{
			"action": "wx-h5-callback",
			"code":   req.Code,
		},
	})

	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "微信授权失败")
		return
	}

	if !output.Success {
		// 检查是否是用户禁用错误
		if errorCode, ok := output.Data["error_code"]; ok {
			if toInt(errorCode, 0) == 40003 {
				Error(c, http.StatusForbidden, 40003, output.Error)
				return
			}
		}
		Error(c, http.StatusBadRequest, 40004, output.Error)
		return
	}

	Success(c, output.Data)
}

// HandleH5Upload H5文件上传
func (h *H5Handler) HandleH5Upload(c *gin.Context) {
	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		Error(c, http.StatusBadRequest, 40004, "请选择要上传的文件")
		return
	}
	defer file.Close()

	// 获取用户ID
	userID, _ := c.Get("user_id")
	userIDInt64, _ := userID.(int64)

	// 文件大小限制 (10MB)
	const maxSize = 10 * 1024 * 1024
	if header.Size > maxSize {
		Error(c, http.StatusBadRequest, 40004, "文件大小不能超过10MB")
		return
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	// 允许的文件类型
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".xls":  true,
		".xlsx": true,
		".txt":  true,
	}
	if !allowedExts[ext] {
		Error(c, http.StatusBadRequest, 40004, "不支持的文件类型")
		return
	}

	// 生成文件名
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("h5_%d_%s%s", userIDInt64, timestamp, ext)

	// 创建上传目录
	uploadDir := filepath.Join("uploads", "h5")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建上传目录失败")
		return
	}

	// 保存文件
	filePath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "保存文件失败")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "写入文件失败")
		return
	}

	// 返回文件URL
	fileURL := fmt.Sprintf("/uploads/h5/%s", filename)

	Success(c, gin.H{
		"filename":  filename,
		"original":  header.Filename,
		"url":       fileURL,
		"size":      header.Size,
		"upload_at": time.Now().Format(time.RFC3339),
	})
}

// ======================== 平台配置API ========================

// PlatformConfig 平台配置
type PlatformConfig struct {
	AppName       string            `json:"app_name"`
	Version       string            `json:"version"`
	Features      map[string]bool   `json:"features"`
	Contact       map[string]string `json:"contact"`
	UploadMaxSize int64             `json:"upload_max_size"`
	AllowedExts   []string          `json:"allowed_extensions"`
}

// HandleGetPlatformConfig 获取平台配置
func HandleGetPlatformConfig(c *gin.Context) {
	config := PlatformConfig{
		AppName: "Digital Twin",
		Version: "2.0.0",
		Features: map[string]bool{
			"wechat_login":    true,
			"wechat_h5_login": true,
			"file_upload":     true,
			"feedback":        true,
		},
		Contact: map[string]string{
			"email":  "support@example.com",
			"wechat": "digital_twin_support",
		},
		UploadMaxSize: 10 * 1024 * 1024, // 10MB
		AllowedExts:   []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".txt"},
	}

	Success(c, config)
}

// ======================== 用户禁用检查中间件 ========================

// UserStatusChecker 用户状态检查中间件
// 用于在JWT验证后检查用户是否被禁用
func UserStatusChecker(userRepo *database.UserRepository) gin.HandlerFunc {
	// 内存缓存禁用用户（简化实现）
	disabledUsersCache := make(map[int64]bool)

	// 定期清理缓存（每5分钟）
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			disabledUsersCache = make(map[int64]bool)
		}
	}()

	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		var userIDInt64 int64
		switch v := userID.(type) {
		case int64:
			userIDInt64 = v
		case float64:
			userIDInt64 = int64(v)
		}

		// 检查缓存
		if disabledUsersCache[userIDInt64] {
			Error(c, http.StatusForbidden, 40003, "用户已被禁用")
			c.Abort()
			return
		}

		// 从数据库查询用户状态
		user, err := userRepo.GetByID(userIDInt64)
		if err != nil || user == nil {
			c.Next()
			return
		}

		if user.Status == "disabled" {
			// 加入缓存
			disabledUsersCache[userIDInt64] = true
			Error(c, http.StatusForbidden, 40003, "用户已被禁用")
			c.Abort()
			return
		}

		c.Next()
	}
}
