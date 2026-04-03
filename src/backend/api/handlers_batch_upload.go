package api

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"digital-twin/src/backend/database"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 批量上传相关常量
const (
	maxBatchFiles     = 20
	maxBatchTotalMB   = 100
	maxBatchTotalSize = maxBatchTotalMB * 1024 * 1024 // 100MB
)

// allowedBatchFormats 允许的文件格式
var allowedBatchFormats = map[string]bool{
	".pdf":  true,
	".docx": true,
	".txt":  true,
	".md":   true,
}

// HandleBatchUpload 批量文件上传
// POST /api/documents/batch-upload
func (h *Handler) HandleBatchUpload(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userIDInt64, ok := userID.(int64)
	if !ok {
		Error(c, http.StatusUnauthorized, 40001, "用户信息无效")
		return
	}

	// 解析 multipart form
	if err := c.Request.ParseMultipartForm(maxBatchTotalSize); err != nil {
		Error(c, http.StatusBadRequest, 40044, "请求体过大或解析失败: "+err.Error())
		return
	}

	form := c.Request.MultipartForm
	if form == nil || form.File == nil {
		Error(c, http.StatusBadRequest, 40004, "未上传任何文件")
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		Error(c, http.StatusBadRequest, 40004, "未上传任何文件")
		return
	}

	// 校验文件数量
	if len(files) > maxBatchFiles {
		Error(c, http.StatusBadRequest, 40043, fmt.Sprintf("文件数量超限，最多允许 %d 个文件", maxBatchFiles))
		return
	}

	// 校验总大小和文件格式
	var totalSize int64
	for _, file := range files {
		totalSize += file.Size
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !allowedBatchFormats[ext] {
			Error(c, http.StatusBadRequest, 40045, fmt.Sprintf("不支持的文件格式: %s（仅支持 PDF/DOCX/TXT/MD）", ext))
			return
		}
	}
	if totalSize > maxBatchTotalSize {
		Error(c, http.StatusBadRequest, 40044, fmt.Sprintf("文件总大小超限，最大允许 %dMB", maxBatchTotalMB))
		return
	}

	// 获取 persona_id
	personaIDStr := c.PostForm("persona_id")
	if personaIDStr == "" {
		Error(c, http.StatusBadRequest, 40004, "缺少 persona_id 参数")
		return
	}
	var personaID int64
	fmt.Sscanf(personaIDStr, "%d", &personaID)
	if personaID <= 0 {
		Error(c, http.StatusBadRequest, 40004, "无效的 persona_id")
		return
	}

	// 可选 knowledge_base_id
	var knowledgeBaseID int64
	kbIDStr := c.PostForm("knowledge_base_id")
	if kbIDStr != "" {
		fmt.Sscanf(kbIDStr, "%d", &knowledgeBaseID)
	}

	// 生成任务ID
	taskID := "task_" + uuid.New().String()[:8]

	// 创建落盘目录
	inputDir := filepath.Join("data", "pending-imports", taskID)
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建临时目录失败: "+err.Error())
		return
	}

	// 落盘文件
	for _, file := range files {
		src, err := file.Open()
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "读取上传文件失败: "+err.Error())
			return
		}

		dstPath := filepath.Join(inputDir, file.Filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			src.Close()
			Error(c, http.StatusInternalServerError, 50001, "保存文件失败: "+err.Error())
			return
		}

		_, err = io.Copy(dst, src)
		src.Close()
		dst.Close()
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "写入文件失败: "+err.Error())
			return
		}
	}

	// 创建 batch_tasks 记录
	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	repo := database.NewBatchTaskRepository(db)
	task := &database.BatchTask{
		TaskID:          taskID,
		PersonaID:       personaID,
		KnowledgeBaseID: knowledgeBaseID,
		Status:          "pending",
		TotalFiles:      len(files),
	}

	if _, err := repo.Create(task); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "创建任务记录失败: "+err.Error())
		return
	}

	// 异步调用 Python 脚本处理
	go processBatchUpload(db, taskID, inputDir, personaID, knowledgeBaseID, userIDInt64)

	// 返回 202 Accepted
	c.JSON(http.StatusAccepted, gin.H{
		"code":    0,
		"message": "任务已提交，正在后台处理",
		"data": gin.H{
			"task_id":     taskID,
			"status":      "pending",
			"total_files": len(files),
		},
	})
}

// processBatchUpload 异步处理批量上传（调用Python脚本）
func processBatchUpload(db *sql.DB, taskID, inputDir string, personaID, knowledgeBaseID, userID int64) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[BatchUpload] panic recovered: %v", r)
		}
	}()

	repo := database.NewBatchTaskRepository(db)

	// 更新状态为 processing
	_ = repo.UpdateStatus(taskID, "processing", 0, 0, "{}")

	// 查找 Python 可执行文件
	pythonCmd := findPython()
	if pythonCmd == "" {
		log.Printf("[BatchUpload] Python 不可用，标记任务失败: %s", taskID)
		_ = repo.UpdateStatus(taskID, "failed", 0, 0, `{"error":"Python not available"}`)
		return
	}

	// 构建 Python 脚本命令
	scriptPath := filepath.Join("scripts", "import_documents.py")
	dbPath := filepath.Join("data", "digital-twin.db")

	args := []string{
		scriptPath,
		"--task-id", taskID,
		"--input-dir", inputDir,
		"--db-path", dbPath,
		"--persona-id", fmt.Sprintf("%d", personaID),
	}
	if knowledgeBaseID > 0 {
		args = append(args, "--knowledge-base-id", fmt.Sprintf("%d", knowledgeBaseID))
	}

	cmd := exec.Command(pythonCmd, args...)
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("[BatchUpload] Python 脚本执行失败: %v, output: %s", err, string(output))
		_ = repo.UpdateStatus(taskID, "failed", 0, 0, fmt.Sprintf(`{"error":"%s","output":"%s"}`, err.Error(), string(output)))
		return
	}

	log.Printf("[BatchUpload] 任务 %s 处理完成", taskID)
	// Python 脚本应自行更新 batch_tasks 表状态
}

// findPython 查找可用的 Python 命令
func findPython() string {
	for _, cmd := range []string{"python3", "python"} {
		if _, err := exec.LookPath(cmd); err == nil {
			return cmd
		}
	}
	return ""
}

// HandleGetBatchTask 查询批量任务状态
// GET /api/batch-tasks/:task_id
func (h *Handler) HandleGetBatchTask(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		Error(c, http.StatusBadRequest, 40004, "缺少 task_id 参数")
		return
	}

	db := h.manager.GetDB()
	if db == nil {
		Error(c, http.StatusInternalServerError, 50001, "数据库服务不可用")
		return
	}

	repo := database.NewBatchTaskRepository(db)
	task, err := repo.GetByTaskID(taskID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "查询任务失败: "+err.Error())
		return
	}
	if task == nil {
		Error(c, http.StatusNotFound, 40046, "批量任务不存在")
		return
	}

	Success(c, gin.H{
		"task_id":       task.TaskID,
		"status":        task.Status,
		"total_files":   task.TotalFiles,
		"success_files": task.SuccessFiles,
		"failed_files":  task.FailedFiles,
		"result_json":   task.ResultJSON,
		"created_at":    task.CreatedAt.Format(time.RFC3339),
		"updated_at":    task.UpdatedAt.Format(time.RFC3339),
	})
}
