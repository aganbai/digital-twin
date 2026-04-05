-- ============================================================
-- V2.0 迭代11 数据库迁移脚本
-- 日期: 2026-04-04
-- 说明: 支持班级公开、教师分身绑定班级、自测学生功能
-- ============================================================

-- 注意：SQLite 不支持在 ALTER TABLE 中添加带有复杂约束的列
-- 以下 SQL 语句在 Go 代码的 autoMigrate 中自动执行
-- 本文件作为迁移记录和手动执行参考

-- ========================
-- 1. Users 表变更
-- ========================

-- 新增字段：是否为自测学生账号
ALTER TABLE users ADD COLUMN is_test_student INTEGER NOT NULL DEFAULT 0;

-- 新增字段：自测学生所属教师ID
ALTER TABLE users ADD COLUMN test_teacher_id INTEGER DEFAULT 0;

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_is_test_student ON users(is_test_student);
CREATE INDEX IF NOT EXISTS idx_users_test_teacher_id ON users(test_teacher_id);

-- ========================
-- 2. Classes 表变更
-- ========================

-- 新增字段：班级是否公开（默认公开）
ALTER TABLE classes ADD COLUMN is_public INTEGER NOT NULL DEFAULT 1;

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_classes_is_public ON classes(is_public);

-- ========================
-- 3. Personas 表变更
-- ========================

-- 新增字段：绑定的班级ID（教师分身必填）
ALTER TABLE personas ADD COLUMN bound_class_id INTEGER DEFAULT NULL;

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_personas_bound_class_id ON personas(bound_class_id);

-- ========================
-- 数据迁移说明
-- ========================

-- 1. 现有班级默认为公开（is_public = 1）
-- 2. 现有教师分身的 bound_class_id 保持为 NULL
-- 3. 现有用户的 is_test_student 默认为 FALSE（0）

-- ========================
-- 验证查询
-- ========================

-- 验证 users 表新字段
SELECT is_test_student, test_teacher_id FROM users LIMIT 1;

-- 验证 classes 表新字段
SELECT is_public FROM classes LIMIT 1;

-- 验证 personas 表新字段
SELECT bound_class_id FROM personas LIMIT 1;

-- 验证索引是否创建成功
SELECT name FROM sqlite_master WHERE type='index' AND name LIKE 'idx_%' ORDER BY name;
