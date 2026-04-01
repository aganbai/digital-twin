package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// Database 数据库管理器
type Database struct {
	DB *sql.DB
}

// NewDatabase 创建数据库连接
// 自动创建数据库文件所在的目录，并执行自动建表
func NewDatabase(dbPath string) (*Database, error) {
	// 自动创建目录
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据库目录失败: %w", err)
	}

	// 打开数据库连接
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 启用外键约束
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("启用外键约束失败: %w", err)
	}

	// 设置 WAL 模式提升并发性能
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("设置 WAL 模式失败: %w", err)
	}

	database := &Database{DB: db}

	// 自动建表
	if err := database.autoMigrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("自动建表失败: %w", err)
	}

	return database, nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

// autoMigrate 自动创建表结构
func (d *Database) autoMigrate() error {
	tables := []string{
		createUsersTable,
		createUsersOpenIDIndex,
		createDocumentsTable,
		createConversationsTable,
		createMemoriesTable,
		// V2.0 迭代1 新增表
		createTeacherStudentRelationsTable,
		createTeacherCommentsTable,
		createStudentDialogueStylesTable,
		createAssignmentsTable,
		createAssignmentReviewsTable,
		// V2.0 迭代2 新增表
		createPersonasTable,
		createClassesTable,
		createClassMembersTable,
		createPersonaSharesTable,
		// V2.0 迭代4 新增表
		createTeacherTakeoversTable,
	}

	for _, ddl := range tables {
		if _, err := d.DB.Exec(ddl); err != nil {
			return fmt.Errorf("执行建表语句失败: %w", err)
		}
	}

	// ALTER TABLE 语句（忽略 "duplicate column" 错误）
	alterStatements := []string{
		alterUsersAddSchool,
		alterUsersAddDescription,
		// V2.0 迭代2 ALTER TABLE
		alterUsersAddDefaultPersonaID,
		alterRelationsAddTeacherPersonaID,
		alterRelationsAddStudentPersonaID,
		alterDocumentsAddScope,
		alterDocumentsAddScopeID,
		alterDocumentsAddPersonaID,
		alterConversationsAddTeacherPersonaID,
		alterConversationsAddStudentPersonaID,
		alterMemoriesAddTeacherPersonaID,
		alterMemoriesAddStudentPersonaID,
		alterCommentsAddTeacherPersonaID,
		alterCommentsAddStudentPersonaID,
		alterStylesAddTeacherPersonaID,
		alterStylesAddStudentPersonaID,
		alterAssignmentsAddTeacherPersonaID,
		alterAssignmentsAddStudentPersonaID,
		// V2.0 迭代3 ALTER TABLE
		alterClassesAddIsActive,
		alterRelationsAddIsActive,
		// V2.0 迭代4 ALTER TABLE
		alterPersonasAddIsPublic,
		alterSharesAddTargetStudentPersonaID,
		alterConversationsAddSenderType,
		alterConversationsAddReplyToID,
		alterDocumentsAddSummary,
		// V2.0 迭代6 ALTER TABLE
		alterMemoriesAddMemoryLayer,
		alterDocumentsAddSourceSessionID,
	}
	for _, stmt := range alterStatements {
		if _, err := d.DB.Exec(stmt); err != nil {
			// SQLite ALTER TABLE ADD COLUMN 如果列已存在会报 "duplicate column name"
			if !strings.Contains(err.Error(), "duplicate column") {
				return fmt.Errorf("执行 ALTER TABLE 失败: %w", err)
			}
		}
	}

	// 创建依赖 ALTER TABLE 新增列的索引（必须在 ALTER TABLE 之后执行）
	// 先处理旧数据中可能存在的重复 (nickname, school) 组合，避免唯一约束冲突
	_, _ = d.DB.Exec(`
		UPDATE users SET school = 'unknown_' || CAST(id AS TEXT)
		WHERE role = 'teacher' AND (school IS NULL OR school = '')
		AND id NOT IN (
			SELECT MIN(id) FROM users
			WHERE role = 'teacher' AND (school IS NULL OR school = '')
			GROUP BY nickname
		)
	`)
	if _, err := d.DB.Exec(createTeacherSchoolIndex); err != nil {
		// 索引已存在时忽略错误
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("创建教师学校索引失败: %w", err)
		}
	}

	// 数据迁移：为已有对话关系自动创建 approved 记录
	_, _ = d.DB.Exec(`
		INSERT OR IGNORE INTO teacher_student_relations (teacher_id, student_id, status, initiated_by)
		SELECT DISTINCT teacher_id, student_id, 'approved', 'teacher'
		FROM conversations
		WHERE NOT EXISTS (
			SELECT 1 FROM teacher_student_relations
			WHERE teacher_student_relations.teacher_id = conversations.teacher_id
			AND teacher_student_relations.student_id = conversations.student_id
		)
	`)

	// V2.0 迭代2 索引
	if _, err := d.DB.Exec(createPersonaTeacherSchoolIndex); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("创建分身教师学校索引失败: %w", err)
		}
	}

	// V2.0 迭代2 数据迁移：为现有用户创建默认分身
	_, _ = d.DB.Exec(`
		INSERT OR IGNORE INTO personas (user_id, role, nickname, school, description)
		SELECT id, role, COALESCE(nickname, username), COALESCE(school, ''), COALESCE(description, '')
		FROM users
		WHERE role != '' AND role IS NOT NULL
		AND NOT EXISTS (SELECT 1 FROM personas WHERE personas.user_id = users.id)
	`)

	// 回填 users.default_persona_id
	_, _ = d.DB.Exec(`
		UPDATE users SET default_persona_id = (
			SELECT p.id FROM personas p WHERE p.user_id = users.id LIMIT 1
		) WHERE default_persona_id = 0 AND EXISTS (SELECT 1 FROM personas WHERE personas.user_id = users.id)
	`)

	// 回填 teacher_student_relations 的 persona_id
	_, _ = d.DB.Exec(`
		UPDATE teacher_student_relations SET
			teacher_persona_id = COALESCE((SELECT p.id FROM personas p WHERE p.user_id = teacher_student_relations.teacher_id AND p.role = 'teacher' LIMIT 1), 0),
			student_persona_id = COALESCE((SELECT p.id FROM personas p WHERE p.user_id = teacher_student_relations.student_id AND p.role = 'student' LIMIT 1), 0)
		WHERE teacher_persona_id = 0 OR student_persona_id = 0
	`)

	// 回填 documents.persona_id
	_, _ = d.DB.Exec(`
		UPDATE documents SET persona_id = COALESCE((
			SELECT p.id FROM personas p WHERE p.user_id = documents.teacher_id AND p.role = 'teacher' LIMIT 1
		), 0) WHERE persona_id = 0
	`)

	// 回填 conversations 的 persona_id
	_, _ = d.DB.Exec(`
		UPDATE conversations SET
			teacher_persona_id = COALESCE((SELECT p.id FROM personas p WHERE p.user_id = conversations.teacher_id AND p.role = 'teacher' LIMIT 1), 0),
			student_persona_id = COALESCE((SELECT p.id FROM personas p WHERE p.user_id = conversations.student_id AND p.role = 'student' LIMIT 1), 0)
		WHERE teacher_persona_id = 0 OR student_persona_id = 0
	`)

	// 回填 memories 的 persona_id
	_, _ = d.DB.Exec(`
		UPDATE memories SET
			teacher_persona_id = COALESCE((SELECT p.id FROM personas p WHERE p.user_id = memories.teacher_id AND p.role = 'teacher' LIMIT 1), 0),
			student_persona_id = COALESCE((SELECT p.id FROM personas p WHERE p.user_id = memories.student_id AND p.role = 'student' LIMIT 1), 0)
		WHERE teacher_persona_id = 0 OR student_persona_id = 0
	`)

	// 回填 teacher_comments 的 persona_id
	_, _ = d.DB.Exec(`
		UPDATE teacher_comments SET
			teacher_persona_id = COALESCE((SELECT p.id FROM personas p WHERE p.user_id = teacher_comments.teacher_id AND p.role = 'teacher' LIMIT 1), 0),
			student_persona_id = COALESCE((SELECT p.id FROM personas p WHERE p.user_id = teacher_comments.student_id AND p.role = 'student' LIMIT 1), 0)
		WHERE teacher_persona_id = 0 OR student_persona_id = 0
	`)

	// 回填 student_dialogue_styles 的 persona_id
	_, _ = d.DB.Exec(`
		UPDATE student_dialogue_styles SET
			teacher_persona_id = COALESCE((SELECT p.id FROM personas p WHERE p.user_id = student_dialogue_styles.teacher_id AND p.role = 'teacher' LIMIT 1), 0),
			student_persona_id = COALESCE((SELECT p.id FROM personas p WHERE p.user_id = student_dialogue_styles.student_id AND p.role = 'student' LIMIT 1), 0)
		WHERE teacher_persona_id = 0 OR student_persona_id = 0
	`)

	// 回填 assignments 的 persona_id
	_, _ = d.DB.Exec(`
		UPDATE assignments SET
			teacher_persona_id = COALESCE((SELECT p.id FROM personas p WHERE p.user_id = assignments.teacher_id AND p.role = 'teacher' LIMIT 1), 0),
			student_persona_id = COALESCE((SELECT p.id FROM personas p WHERE p.user_id = assignments.student_id AND p.role = 'student' LIMIT 1), 0)
		WHERE teacher_persona_id = 0 OR student_persona_id = 0
	`)

	// V2.0 迭代4 索引
	if _, err := d.DB.Exec(createTakeoverSessionIndex); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("创建接管索引失败: %w", err)
		}
	}

	// V2.0 迭代4 数据回填：旧对话数据 sender_type
	_, _ = d.DB.Exec(`UPDATE conversations SET sender_type = 'student' WHERE role = 'user' AND sender_type = ''`)
	_, _ = d.DB.Exec(`UPDATE conversations SET sender_type = 'ai' WHERE role = 'assistant' AND sender_type = ''`)

	// V2.0 迭代6 索引
	if _, err := d.DB.Exec(createMemoriesLayerIndex); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("创建记忆分层索引失败: %w", err)
		}
	}
	if _, err := d.DB.Exec(createMemoriesTypeLayerIndex); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("创建记忆类型分层索引失败: %w", err)
		}
	}

	return nil
}

// 建表 SQL 语句
const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    username    TEXT NOT NULL UNIQUE,
    password    TEXT NOT NULL,
    role        TEXT NOT NULL DEFAULT 'student',
    nickname    TEXT,
    email       TEXT,
    openid      TEXT DEFAULT '',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);`

// openid 唯一索引（排除空字符串）
const createUsersOpenIDIndex = `
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_openid ON users(openid) WHERE openid != '';`

const createDocumentsTable = `
CREATE TABLE IF NOT EXISTS documents (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    teacher_id  INTEGER NOT NULL,
    title       TEXT NOT NULL,
    content     TEXT NOT NULL,
    doc_type    TEXT DEFAULT 'text',
    tags        TEXT,
    status      TEXT DEFAULT 'active',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (teacher_id) REFERENCES users(id)
);`

const createConversationsTable = `
CREATE TABLE IF NOT EXISTS conversations (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id      INTEGER NOT NULL,
    teacher_id      INTEGER NOT NULL,
    session_id      TEXT NOT NULL,
    role            TEXT NOT NULL,
    content         TEXT NOT NULL,
    token_count     INTEGER DEFAULT 0,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES users(id),
    FOREIGN KEY (teacher_id) REFERENCES users(id)
);`

const createMemoriesTable = `
CREATE TABLE IF NOT EXISTS memories (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id      INTEGER NOT NULL,
    teacher_id      INTEGER NOT NULL,
    memory_type     TEXT NOT NULL,
    content         TEXT NOT NULL,
    importance      REAL DEFAULT 0.5,
    last_accessed   DATETIME,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES users(id),
    FOREIGN KEY (teacher_id) REFERENCES users(id)
);`

// ======================== V2.0 迭代1 新增表 ========================

const alterUsersAddSchool = `ALTER TABLE users ADD COLUMN school TEXT DEFAULT '';`
const alterUsersAddDescription = `ALTER TABLE users ADD COLUMN description TEXT DEFAULT '';`
const createTeacherSchoolIndex = `CREATE UNIQUE INDEX IF NOT EXISTS idx_teacher_school ON users(nickname, school) WHERE role = 'teacher';`

const createTeacherStudentRelationsTable = `
CREATE TABLE IF NOT EXISTS teacher_student_relations (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    teacher_id      INTEGER NOT NULL,
    student_id      INTEGER NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    initiated_by    TEXT NOT NULL,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (teacher_id) REFERENCES users(id),
    FOREIGN KEY (student_id) REFERENCES users(id),
    UNIQUE(teacher_id, student_id)
);`

const createTeacherCommentsTable = `
CREATE TABLE IF NOT EXISTS teacher_comments (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    teacher_id      INTEGER NOT NULL,
    student_id      INTEGER NOT NULL,
    content         TEXT NOT NULL,
    progress_summary TEXT,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (teacher_id) REFERENCES users(id),
    FOREIGN KEY (student_id) REFERENCES users(id)
);`

const createStudentDialogueStylesTable = `
CREATE TABLE IF NOT EXISTS student_dialogue_styles (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    teacher_id      INTEGER NOT NULL,
    student_id      INTEGER NOT NULL,
    style_config    TEXT NOT NULL,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (teacher_id) REFERENCES users(id),
    FOREIGN KEY (student_id) REFERENCES users(id),
    UNIQUE(teacher_id, student_id)
);`

const createAssignmentsTable = `
CREATE TABLE IF NOT EXISTS assignments (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id      INTEGER NOT NULL,
    teacher_id      INTEGER NOT NULL,
    title           TEXT NOT NULL,
    content         TEXT,
    file_path       TEXT,
    file_type       TEXT,
    status          TEXT DEFAULT 'submitted',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES users(id),
    FOREIGN KEY (teacher_id) REFERENCES users(id)
);`

const createAssignmentReviewsTable = `
CREATE TABLE IF NOT EXISTS assignment_reviews (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    assignment_id   INTEGER NOT NULL,
    reviewer_type   TEXT NOT NULL,
    reviewer_id     INTEGER,
    content         TEXT NOT NULL,
    score           REAL,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (assignment_id) REFERENCES assignments(id)
);`

// ======================== V2.0 迭代2 新增表 ========================

const createPersonasTable = `
CREATE TABLE IF NOT EXISTS personas (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL,
    role            TEXT NOT NULL,
    nickname        TEXT NOT NULL,
    school          TEXT DEFAULT '',
    description     TEXT DEFAULT '',
    avatar          TEXT DEFAULT '',
    is_active       INTEGER DEFAULT 1,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);`

const createPersonaTeacherSchoolIndex = `CREATE UNIQUE INDEX IF NOT EXISTS idx_persona_teacher_school ON personas(nickname, school) WHERE role = 'teacher';`

const createClassesTable = `
CREATE TABLE IF NOT EXISTS classes (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    persona_id      INTEGER NOT NULL,
    name            TEXT NOT NULL,
    description     TEXT DEFAULT '',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (persona_id) REFERENCES personas(id),
    UNIQUE(persona_id, name)
);`

const createClassMembersTable = `
CREATE TABLE IF NOT EXISTS class_members (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id            INTEGER NOT NULL,
    student_persona_id  INTEGER NOT NULL,
    joined_at           DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (class_id) REFERENCES classes(id),
    FOREIGN KEY (student_persona_id) REFERENCES personas(id),
    UNIQUE(class_id, student_persona_id)
);`

const createPersonaSharesTable = `
CREATE TABLE IF NOT EXISTS persona_shares (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    teacher_persona_id  INTEGER NOT NULL,
    share_code          TEXT NOT NULL UNIQUE,
    class_id            INTEGER,
    expires_at          DATETIME,
    max_uses            INTEGER DEFAULT 0,
    used_count          INTEGER DEFAULT 0,
    is_active           INTEGER DEFAULT 1,
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (teacher_persona_id) REFERENCES personas(id),
    FOREIGN KEY (class_id) REFERENCES classes(id)
);`

// V2.0 迭代2 ALTER TABLE 语句
const alterUsersAddDefaultPersonaID = `ALTER TABLE users ADD COLUMN default_persona_id INTEGER DEFAULT 0;`

const alterRelationsAddTeacherPersonaID = `ALTER TABLE teacher_student_relations ADD COLUMN teacher_persona_id INTEGER DEFAULT 0;`
const alterRelationsAddStudentPersonaID = `ALTER TABLE teacher_student_relations ADD COLUMN student_persona_id INTEGER DEFAULT 0;`

const alterDocumentsAddScope = `ALTER TABLE documents ADD COLUMN scope TEXT DEFAULT 'global';`
const alterDocumentsAddScopeID = `ALTER TABLE documents ADD COLUMN scope_id INTEGER DEFAULT 0;`
const alterDocumentsAddPersonaID = `ALTER TABLE documents ADD COLUMN persona_id INTEGER DEFAULT 0;`

const alterConversationsAddTeacherPersonaID = `ALTER TABLE conversations ADD COLUMN teacher_persona_id INTEGER DEFAULT 0;`
const alterConversationsAddStudentPersonaID = `ALTER TABLE conversations ADD COLUMN student_persona_id INTEGER DEFAULT 0;`

const alterMemoriesAddTeacherPersonaID = `ALTER TABLE memories ADD COLUMN teacher_persona_id INTEGER DEFAULT 0;`
const alterMemoriesAddStudentPersonaID = `ALTER TABLE memories ADD COLUMN student_persona_id INTEGER DEFAULT 0;`

const alterCommentsAddTeacherPersonaID = `ALTER TABLE teacher_comments ADD COLUMN teacher_persona_id INTEGER DEFAULT 0;`
const alterCommentsAddStudentPersonaID = `ALTER TABLE teacher_comments ADD COLUMN student_persona_id INTEGER DEFAULT 0;`

const alterStylesAddTeacherPersonaID = `ALTER TABLE student_dialogue_styles ADD COLUMN teacher_persona_id INTEGER DEFAULT 0;`
const alterStylesAddStudentPersonaID = `ALTER TABLE student_dialogue_styles ADD COLUMN student_persona_id INTEGER DEFAULT 0;`

const alterAssignmentsAddTeacherPersonaID = `ALTER TABLE assignments ADD COLUMN teacher_persona_id INTEGER DEFAULT 0;`
const alterAssignmentsAddStudentPersonaID = `ALTER TABLE assignments ADD COLUMN student_persona_id INTEGER DEFAULT 0;`

// ======================== V2.0 迭代3 ALTER TABLE 语句 ========================

const alterClassesAddIsActive = `ALTER TABLE classes ADD COLUMN is_active INTEGER DEFAULT 1;`
const alterRelationsAddIsActive = `ALTER TABLE teacher_student_relations ADD COLUMN is_active INTEGER DEFAULT 1;`

// ======================== V2.0 迭代4 DDL ========================

const alterPersonasAddIsPublic = `ALTER TABLE personas ADD COLUMN is_public INTEGER DEFAULT 0;`
const alterSharesAddTargetStudentPersonaID = `ALTER TABLE persona_shares ADD COLUMN target_student_persona_id INTEGER DEFAULT 0;`
const alterConversationsAddSenderType = `ALTER TABLE conversations ADD COLUMN sender_type TEXT DEFAULT '';`
const alterConversationsAddReplyToID = `ALTER TABLE conversations ADD COLUMN reply_to_id INTEGER DEFAULT 0;`
const alterDocumentsAddSummary = `ALTER TABLE documents ADD COLUMN summary TEXT DEFAULT '';`

const createTeacherTakeoversTable = `
CREATE TABLE IF NOT EXISTS teacher_takeovers (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    teacher_persona_id  INTEGER NOT NULL,
    student_persona_id  INTEGER NOT NULL,
    session_id          TEXT NOT NULL,
    status              TEXT NOT NULL DEFAULT 'active',
    started_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    ended_at            DATETIME,
    FOREIGN KEY (teacher_persona_id) REFERENCES personas(id),
    FOREIGN KEY (student_persona_id) REFERENCES personas(id)
);`

const createTakeoverSessionIndex = `CREATE INDEX IF NOT EXISTS idx_takeover_session ON teacher_takeovers(session_id, status);`

// ======================== V2.0 迭代6 DDL ========================

const alterMemoriesAddMemoryLayer = `ALTER TABLE memories ADD COLUMN memory_layer TEXT NOT NULL DEFAULT 'episodic';`
const createMemoriesLayerIndex = `CREATE INDEX IF NOT EXISTS idx_memories_layer ON memories(teacher_persona_id, student_persona_id, memory_layer);`
const createMemoriesTypeLayerIndex = `CREATE INDEX IF NOT EXISTS idx_memories_type_layer ON memories(teacher_persona_id, student_persona_id, memory_type, memory_layer);`
const alterDocumentsAddSourceSessionID = `ALTER TABLE documents ADD COLUMN source_session_id TEXT DEFAULT '';`
