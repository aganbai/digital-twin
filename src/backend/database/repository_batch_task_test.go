package database

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupBatchTaskTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS batch_tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id TEXT NOT NULL UNIQUE,
			persona_id INTEGER NOT NULL DEFAULT 0,
			knowledge_base_id INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'pending',
			total_files INTEGER NOT NULL DEFAULT 0,
			success_files INTEGER NOT NULL DEFAULT 0,
			failed_files INTEGER NOT NULL DEFAULT 0,
			result_json TEXT DEFAULT '{}',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	return db
}

func TestBatchTaskRepository_Create(t *testing.T) {
	db := setupBatchTaskTestDB(t)
	defer db.Close()

	repo := NewBatchTaskRepository(db)
	task := &BatchTask{
		TaskID:          "task_abc123",
		PersonaID:       1,
		KnowledgeBaseID: 2,
		Status:          "pending",
		TotalFiles:      5,
	}

	id, err := repo.Create(task)
	if err != nil {
		t.Fatalf("创建批量任务失败: %v", err)
	}
	if id <= 0 {
		t.Fatal("创建的任务ID应大于0")
	}

	t.Logf("创建批量任务成功, id=%d", id)
}

func TestBatchTaskRepository_GetByTaskID(t *testing.T) {
	db := setupBatchTaskTestDB(t)
	defer db.Close()

	repo := NewBatchTaskRepository(db)

	// 创建任务
	task := &BatchTask{
		TaskID:     "task_get_test",
		PersonaID:  1,
		Status:     "pending",
		TotalFiles: 3,
	}
	_, err := repo.Create(task)
	if err != nil {
		t.Fatalf("创建任务失败: %v", err)
	}

	// 查询任务
	found, err := repo.GetByTaskID("task_get_test")
	if err != nil {
		t.Fatalf("查询任务失败: %v", err)
	}
	if found == nil {
		t.Fatal("应该找到任务")
	}
	if found.TaskID != "task_get_test" {
		t.Errorf("任务ID不匹配: %s", found.TaskID)
	}
	if found.TotalFiles != 3 {
		t.Errorf("文件数不匹配: %d", found.TotalFiles)
	}

	// 查询不存在的任务
	notFound, err := repo.GetByTaskID("nonexistent")
	if err != nil {
		t.Fatalf("查询不存在任务不应报错: %v", err)
	}
	if notFound != nil {
		t.Fatal("不存在的任务应返回nil")
	}
}

func TestBatchTaskRepository_UpdateStatus(t *testing.T) {
	db := setupBatchTaskTestDB(t)
	defer db.Close()

	repo := NewBatchTaskRepository(db)

	// 创建任务
	task := &BatchTask{
		TaskID:     "task_update_test",
		PersonaID:  1,
		Status:     "pending",
		TotalFiles: 5,
	}
	_, err := repo.Create(task)
	if err != nil {
		t.Fatalf("创建任务失败: %v", err)
	}

	// 更新状态
	err = repo.UpdateStatus("task_update_test", "success", 4, 1, `{"results":[]}`)
	if err != nil {
		t.Fatalf("更新状态失败: %v", err)
	}

	// 验证更新
	found, err := repo.GetByTaskID("task_update_test")
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if found.Status != "success" {
		t.Errorf("状态不匹配: %s", found.Status)
	}
	if found.SuccessFiles != 4 {
		t.Errorf("成功文件数不匹配: %d", found.SuccessFiles)
	}
	if found.FailedFiles != 1 {
		t.Errorf("失败文件数不匹配: %d", found.FailedFiles)
	}
}
