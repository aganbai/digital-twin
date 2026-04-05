package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// LogDatabase 操作日志数据库管理器（独立数据库文件）
type LogDatabase struct {
	DB *sql.DB
}

// NewLogDatabase 创建日志数据库连接
func NewLogDatabase(dbPath string) (*LogDatabase, error) {
	// 自动创建目录
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志数据库目录失败: %w", err)
	}

	// 打开数据库连接
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开日志数据库失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("日志数据库连接测试失败: %w", err)
	}

	// 设置 WAL 模式
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("设置日志数据库 WAL 模式失败: %w", err)
	}

	logDB := &LogDatabase{DB: db}

	// 自动建表
	if err := logDB.autoMigrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("日志数据库自动建表失败: %w", err)
	}

	return logDB, nil
}

// Close 关闭数据库连接
func (d *LogDatabase) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

// autoMigrate 自动创建表结构
func (d *LogDatabase) autoMigrate() error {
	tables := []string{
		createOperationLogsTable,
		createOperationLogsUserIndex,
		createOperationLogsActionIndex,
		createOperationLogsCreatedIndex,
	}

	for _, ddl := range tables {
		if _, err := d.DB.Exec(ddl); err != nil {
			return fmt.Errorf("执行建表语句失败: %w", err)
		}
	}

	return nil
}

// CleanupOldLogs 清理超过指定天数的日志
func (d *LogDatabase) CleanupOldLogs(retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays).Format("2006-01-02")
	result, err := d.DB.Exec(
		"DELETE FROM operation_logs WHERE created_at < ?",
		cutoff,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// ======================== 操作日志查询方法 ========================

// OperationLog 操作日志模型
type OperationLog struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	UserRole   string    `json:"user_role"`
	PersonaID  int64     `json:"persona_id"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id"`
	Detail     string    `json:"detail"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
	Platform   string    `json:"platform"`
	StatusCode int       `json:"status_code"`
	DurationMs int       `json:"duration_ms"`
	CreatedAt  time.Time `json:"created_at"`
}

// LogQueryParams 日志查询参数
type LogQueryParams struct {
	UserID    int64
	UserRole  string
	Action    string
	Resource  string
	StartDate string
	EndDate   string
	Page      int
	PageSize  int
}

// QueryLogs 查询操作日志
func (d *LogDatabase) QueryLogs(params LogQueryParams) ([]OperationLog, int, error) {
	// 构建查询条件
	whereClauses := []string{"1=1"}
	args := []interface{}{}

	if params.UserID > 0 {
		whereClauses = append(whereClauses, "user_id = ?")
		args = append(args, params.UserID)
	}
	if params.UserRole != "" {
		whereClauses = append(whereClauses, "user_role = ?")
		args = append(args, params.UserRole)
	}
	if params.Action != "" {
		whereClauses = append(whereClauses, "action = ?")
		args = append(args, params.Action)
	}
	if params.Resource != "" {
		whereClauses = append(whereClauses, "resource = ?")
		args = append(args, params.Resource)
	}
	if params.StartDate != "" {
		whereClauses = append(whereClauses, "created_at >= ?")
		args = append(args, params.StartDate)
	}
	if params.EndDate != "" {
		whereClauses = append(whereClauses, "created_at <= ?")
		args = append(args, params.EndDate+" 23:59:59")
	}

	whereClause := ""
	for i, clause := range whereClauses {
		if i == 0 {
			whereClause = clause
		} else {
			whereClause += " AND " + clause
		}
	}

	// 查询总数
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM operation_logs WHERE %s", whereClause)
	err := d.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询日志总数失败: %w", err)
	}

	// 分页参数
	page := params.Page
	pageSize := params.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 查询列表
	query := fmt.Sprintf(`
		SELECT id, user_id, user_role, persona_id, action, resource, resource_id,
		       detail, ip, user_agent, platform, status_code, duration_ms, created_at
		FROM operation_logs
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, pageSize, offset)
	rows, err := d.DB.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询日志列表失败: %w", err)
	}
	defer rows.Close()

	var logs []OperationLog
	for rows.Next() {
		var log OperationLog
		if err := rows.Scan(
			&log.ID, &log.UserID, &log.UserRole, &log.PersonaID,
			&log.Action, &log.Resource, &log.ResourceID,
			&log.Detail, &log.IP, &log.UserAgent, &log.Platform,
			&log.StatusCode, &log.DurationMs, &log.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("扫描日志记录失败: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}

// LogStats 日志统计结果
type LogStats struct {
	TotalCount   int            `json:"total_count"`
	ActionCounts map[string]int `json:"action_counts"`
	TopUsers     []UserLogCount `json:"top_users"`
}

// UserLogCount 用户日志统计
type UserLogCount struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Count    int    `json:"count"`
}

// GetLogStats 获取日志统计
func (d *LogDatabase) GetLogStats(startDate, endDate string) (*LogStats, error) {
	stats := &LogStats{
		ActionCounts: make(map[string]int),
		TopUsers:     []UserLogCount{},
	}

	// 构建时间条件
	whereClause := "1=1"
	args := []interface{}{}
	if startDate != "" {
		whereClause += " AND created_at >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		whereClause += " AND created_at <= ?"
		args = append(args, endDate+" 23:59:59")
	}

	// 总数
	var total int
	err := d.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM operation_logs WHERE %s", whereClause), args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("查询日志总数失败: %w", err)
	}
	stats.TotalCount = total

	// 按 action 统计
	rows, err := d.DB.Query(fmt.Sprintf(`
		SELECT action, COUNT(*) as count FROM operation_logs
		WHERE %s
		GROUP BY action
		ORDER BY count DESC
	`, whereClause), args...)
	if err != nil {
		return nil, fmt.Errorf("查询日志action统计失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var action string
		var count int
		if err := rows.Scan(&action, &count); err != nil {
			return nil, fmt.Errorf("扫描action统计失败: %w", err)
		}
		stats.ActionCounts[action] = count
	}

	return stats, nil
}
