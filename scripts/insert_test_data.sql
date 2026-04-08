-- ==========================================
-- 冒烟测试数据初始化脚本
-- 创建时间：2026-04-03
-- 用途：为端到端冒烟测试创建必要的测试数据
-- ==========================================

-- 清理可能存在的旧测试数据（避免冲突）
DELETE FROM class_members WHERE class_id IN (SELECT id FROM classes WHERE name = '冒烟测试班-自动化');
DELETE FROM classes WHERE name = '冒烟测试班-自动化';
DELETE FROM teacher_student_relations WHERE teacher_id IN (SELECT id FROM users WHERE username IN ('13800138001', '13900139001'));
DELETE FROM personas WHERE user_id IN (SELECT id FROM users WHERE username IN ('13800138001', '13900139001'));
DELETE FROM users WHERE username IN ('13800138001', '13900139001');

-- ==========================================
-- 第一步：创建测试账号
-- ==========================================

-- 创建教师账号（手机号：13800138001）
INSERT INTO users (username, password, role, nickname, school, description, created_at, updated_at)
VALUES (
    '13800138001',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.iW8jY9D6aV3xV9xKCu',  -- 密码: test123456
    'teacher',
    '测试教师',
    '测试学校',
    '自动化测试教师账号',
    datetime('now', 'localtime'),
    datetime('now', 'localtime')
);

-- 创建学生账号（手机号：13900139001）
INSERT INTO users (username, password, role, nickname, school, description, created_at, updated_at)
VALUES (
    '13900139001',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.iW8jY9D6aV3xV9xKCu',  -- 密码: test123456
    'student',
    '测试学生',
    '',
    '自动化测试学生账号',
    datetime('now', 'localtime'),
    datetime('now', 'localtime')
);

-- 获取新创建的用户ID
-- 教师ID
SELECT '教师用户ID: ' || id FROM users WHERE username = '13800138001';
-- 学生ID
SELECT '学生用户ID: ' || id FROM users WHERE username = '13900139001';

-- ==========================================
-- 第二步：创建分身（Personas）
-- ==========================================

-- 为教师创建教师分身
INSERT INTO personas (user_id, role, nickname, school, description, is_active, is_public, created_at, updated_at)
SELECT 
    id,
    'teacher',
    '测试教师分身',
    '测试学校',
    '自动化测试教师分身',
    1,
    1,
    datetime('now', 'localtime'),
    datetime('now', 'localtime')
FROM users WHERE username = '13800138001';

-- 为学生创建学生分身
INSERT INTO personas (user_id, role, nickname, school, description, is_active, is_public, created_at, updated_at)
SELECT 
    id,
    'student',
    '测试学生分身',
    '',
    '自动化测试学生分身',
    1,
    1,
    datetime('now', 'localtime'),
    datetime('now', 'localtime')
FROM users WHERE username = '13900139001';

-- 更新用户的默认分身ID
UPDATE users SET default_persona_id = (
    SELECT id FROM personas WHERE user_id = (SELECT id FROM users WHERE username = '13800138001') AND role = 'teacher' LIMIT 1
) WHERE username = '13800138001';

UPDATE users SET default_persona_id = (
    SELECT id FROM personas WHERE user_id = (SELECT id FROM users WHERE username = '13900139001') AND role = 'student' LIMIT 1
) WHERE username = '13900139001';

-- ==========================================
-- 第三步：创建班级
-- ==========================================

-- 创建测试班级（关联教师分身）
INSERT INTO classes (persona_id, name, description, is_active, created_at, updated_at)
SELECT 
    p.id,
    '冒烟测试班-自动化',
    '自动化测试专用班级',
    1,
    datetime('now', 'localtime'),
    datetime('now', 'localtime')
FROM personas p
INNER JOIN users u ON p.user_id = u.id
WHERE u.username = '13800138001' AND p.role = 'teacher'
LIMIT 1;

-- ==========================================
-- 第四步：建立师生关系
-- ==========================================

-- 创建师生授权关系
INSERT INTO teacher_student_relations (
    teacher_id, 
    student_id, 
    teacher_persona_id, 
    student_persona_id, 
    status, 
    initiated_by, 
    is_active,
    created_at, 
    updated_at
)
SELECT 
    (SELECT id FROM users WHERE username = '13800138001'),
    (SELECT id FROM users WHERE username = '13900139001'),
    (SELECT id FROM personas WHERE user_id = (SELECT id FROM users WHERE username = '13800138001') AND role = 'teacher' LIMIT 1),
    (SELECT id FROM personas WHERE user_id = (SELECT id FROM users WHERE username = '13900139001') AND role = 'student' LIMIT 1),
    'approved',
    'teacher',
    1,
    datetime('now', 'localtime'),
    datetime('now', 'localtime');

-- ==========================================
-- 第五步：将学生加入班级
-- ==========================================

-- 将学生分身加入班级
INSERT INTO class_members (class_id, student_persona_id, joined_at)
SELECT 
    c.id,
    (SELECT id FROM personas WHERE user_id = (SELECT id FROM users WHERE username = '13900139001') AND role = 'student' LIMIT 1),
    datetime('now', 'localtime')
FROM classes c
WHERE c.name = '冒烟测试班-自动化';

-- ==========================================
-- 验证数据创建结果
-- ==========================================

SELECT '========== 数据创建完成 ==========';
SELECT '教师账号信息：';
SELECT u.id as user_id, u.username, u.role, u.nickname, p.id as persona_id, p.nickname as persona_name
FROM users u
LEFT JOIN personas p ON p.user_id = u.id
WHERE u.username = '13800138001';

SELECT '学生账号信息：';
SELECT u.id as user_id, u.username, u.role, u.nickname, p.id as persona_id, p.nickname as persona_name
FROM users u
LEFT JOIN personas p ON p.user_id = u.id
WHERE u.username = '13900139001';

SELECT '班级信息：';
SELECT c.id, c.name, c.persona_id, p.nickname as teacher_name
FROM classes c
LEFT JOIN personas p ON c.persona_id = p.id
WHERE c.name = '冒烟测试班-自动化';

SELECT '班级成员信息：';
SELECT cm.class_id, c.name as class_name, cm.student_persona_id, p.nickname as student_name
FROM class_members cm
LEFT JOIN classes c ON cm.class_id = c.id
LEFT JOIN personas p ON cm.student_persona_id = p.id
WHERE c.name = '冒烟测试班-自动化';

SELECT '师生关系信息：';
SELECT r.id, r.teacher_id, u1.nickname as teacher_name, r.student_id, u2.nickname as student_name, 
       r.teacher_persona_id, r.student_persona_id, r.status
FROM teacher_student_relations r
LEFT JOIN users u1 ON r.teacher_id = u1.id
LEFT JOIN users u2 ON r.student_id = u2.id
WHERE r.teacher_id = (SELECT id FROM users WHERE username = '13800138001')
  AND r.student_id = (SELECT id FROM users WHERE username = '13900139001');
