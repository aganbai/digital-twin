# V2.0 迭代2 API 接口规范

## 1. 通用约定

> 继承 V2.0 迭代1 的通用约定，以下仅列出新增和变更部分。

### 1.1 基础信息

| 项目 | 说明 |
|------|------|
| Base URL | `http://localhost:8080` |
| 协议 | HTTP（开发环境） |
| 数据格式 | JSON（`Content-Type: application/json`） |
| 认证方式 | Bearer Token（`Authorization: Bearer <token>`） |

### 1.2 JWT Token 结构变更

```json
// 迭代1 的 JWT Claims
{
  "user_id": 1,
  "role": "teacher",
  "exp": 1234567890
}

// 迭代2 的 JWT Claims（新增 persona_id）
{
  "user_id": 1,
  "persona_id": 3,
  "role": "teacher",
  "exp": 1234567890
}
```

**向后兼容**：`persona_id` 为 0 时，后端自动查找用户的 `default_persona_id`。

### 1.3 新增错误码

| 错误码 | HTTP Status | 说明 |
|--------|-------------|------|
| 40013 | 404 | 分身不存在 |
| 40014 | 403 | 分身不属于当前用户 |
| 40015 | 409 | 该学校已有同名教师分身 |
| 40016 | 404 | 班级不存在 |
| 40017 | 403 | 班级不属于当前教师分身 |
| 40018 | 409 | 同名班级已存在 |
| 40019 | 409 | 学生已在该班级中 |
| 40020 | 400 | 分享码无效或已过期 |
| 40021 | 400 | 分享码使用次数已达上限 |
| 40022 | 400 | 需要先创建学生分身 |
| 40023 | 400 | 无效的知识库作用域 |
| 40024 | 400 | 班级有成员，无法删除 |

---

## 2. 改造接口

### 2.1 微信登录（改造）

**POST** `/api/auth/wx-login`

**改造说明**：登录成功后返回用户的分身列表和当前分身信息。

**请求体**（不变）：
```json
{
  "code": "wx_login_code"
}
```

**成功响应** `200`（新用户）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "user_id": 1,
    "is_new_user": true,
    "role": "",
    "personas": []
  }
}
```

**成功响应** `200`（老用户，有分身）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "user_id": 1,
    "is_new_user": false,
    "role": "teacher",
    "current_persona": {
      "id": 3,
      "role": "teacher",
      "nickname": "王老师",
      "school": "北京大学",
      "description": "物理学教授"
    },
    "personas": [
      {
        "id": 3,
        "role": "teacher",
        "nickname": "王老师",
        "school": "北京大学",
        "description": "物理学教授"
      },
      {
        "id": 5,
        "role": "student",
        "nickname": "音乐培训班学生",
        "school": "",
        "description": ""
      }
    ]
  }
}
```

**成功响应** `200`（老用户，无分身 — 迁移前的旧用户）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "user_id": 1,
    "is_new_user": false,
    "role": "",
    "personas": []
  }
}
```

> **前端行为**：
> - `is_new_user=true` 或 `personas` 为空 → 跳转创建分身页
> - `personas` 只有 1 个 → 直接进入对应首页
> - `personas` 有多个 → 跳转分身选择页

---

### 2.2 补全用户信息（改造）

**POST** `/api/auth/complete-profile`

**改造说明**：内部转换为创建分身，向后兼容原有请求格式。

**请求体**（不变，向后兼容）：
```json
{
  "role": "teacher",
  "nickname": "王老师",
  "school": "北京大学",
  "description": "物理学教授"
}
```

**成功响应** `200`（新增 persona 信息）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 1,
    "role": "teacher",
    "nickname": "王老师",
    "school": "北京大学",
    "description": "物理学教授",
    "persona_id": 3,
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

> **行为变更**：
> 1. 不再检查 `user.Role != ""`（允许已有角色的用户创建新分身）
> 2. 内部调用 PersonaRepository.Create 创建分身
> 3. 返回新的 JWT token（包含 persona_id）

---

### 2.3 发送对话消息（改造）

**POST** `/api/chat`

**改造说明**：`teacher_id` 改为 `teacher_persona_id`，向后兼容 `teacher_id`。

**请求体**：
```json
{
  "message": "什么是牛顿第一定律?",
  "teacher_persona_id": 3,
  "session_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| message | string | ✅ | 用户消息 |
| teacher_persona_id | int | ✅ | 教师分身 ID（向后兼容：也接受 teacher_id） |
| session_id | string | ❌ | 会话 ID |

**向后兼容**：如果请求中包含 `teacher_id` 而非 `teacher_persona_id`，后端自动查找该 teacher 的默认分身。

---

### 2.4 SSE 流式对话（改造）

**POST** `/api/chat/stream`

**改造说明**：同 2.3，`teacher_id` 改为 `teacher_persona_id`。

---

### 2.5 获取教师列表（改造）

**GET** `/api/teachers`

**改造说明**：返回教师分身列表（而非用户列表）。

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 3,
        "persona_id": 3,
        "user_id": 1,
        "nickname": "王老师",
        "role": "teacher",
        "school": "北京大学",
        "description": "物理学教授，专注力学和热力学教学",
        "document_count": 5,
        "created_at": "2026-04-01T09:00:00Z"
      }
    ],
    "total": 3,
    "page": 1,
    "page_size": 20
  }
}
```

> **注意**：`id` 和 `persona_id` 相同，都是分身 ID。保留 `id` 是为了向后兼容。

---

### 2.6 添加知识文档（改造）

**POST** `/api/documents`

**改造说明**：新增 `scope` 和 `scope_id` 字段。

**请求体**：
```json
{
  "title": "牛顿运动定律",
  "content": "牛顿第一定律...",
  "tags": "物理,力学",
  "scope": "class",
  "scope_id": 1
}
```

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| title | string | ✅ | - | 文档标题 |
| content | string | ✅ | - | 文档内容 |
| tags | string | ❌ | - | 标签（逗号分隔） |
| scope | string | ❌ | global | 作用域：global / class / student |
| scope_id | int | scope≠global 时 ✅ | 0 | 班级 ID 或学生分身 ID |

**校验规则**：
- `scope=class` 时，`scope_id` 必须是当前教师分身的班级
- `scope=student` 时，`scope_id` 必须是与当前教师分身有 approved 关系的学生分身

**新增错误响应**：
| 场景 | code | message |
|------|------|---------|
| 无效的 scope 值 | 40023 | 无效的知识库作用域，仅支持 global/class/student |
| scope=class 但班级不属于当前分身 | 40017 | 班级不属于当前教师分身 |
| scope=student 但无师生关系 | 40007 | 未获得该学生的授权关系 |

---

### 2.7 获取文档列表（改造）

**GET** `/api/documents`

**改造说明**：新增 scope 筛选参数。

**Query 参数**（新增）：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| scope | string | ❌ | 按作用域筛选：global / class / student |
| scope_id | int | ❌ | 按作用域 ID 筛选 |

**成功响应** `200`（新增 scope 字段）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 10,
        "teacher_id": 1,
        "persona_id": 3,
        "title": "牛顿运动定律",
        "doc_type": "text",
        "tags": "物理,力学",
        "status": "active",
        "scope": "class",
        "scope_id": 1,
        "scope_name": "高一(3)班",
        "chunks_count": 8,
        "created_at": "2026-04-01T09:00:00Z"
      }
    ],
    "total": 5,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 2.8 文件上传（改造）

**POST** `/api/documents/upload`

**改造说明**：新增 scope 和 scope_id 表单字段。

**表单字段**（新增）：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| scope | string | ❌ | 作用域，默认 global |
| scope_id | int | scope≠global 时 ✅ | 班级 ID 或学生分身 ID |

---

### 2.9 URL 导入（改造）

**POST** `/api/documents/import-url`

**改造说明**：新增 scope 和 scope_id 字段。

**请求体**（新增字段）：
```json
{
  "url": "https://example.com/article",
  "title": "可选标题",
  "tags": "物理,力学",
  "scope": "student",
  "scope_id": 5
}
```

---

### 2.10 师生关系接口（改造）

所有师生关系接口中的 `teacher_id` / `student_id` 改为 `teacher_persona_id` / `student_persona_id`，向后兼容旧字段。

#### 教师邀请学生（改造）

**POST** `/api/relations/invite`

**请求体**：
```json
{
  "student_persona_id": 5
}
```

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "teacher_persona_id": 3,
    "student_persona_id": 5,
    "status": "approved",
    "initiated_by": "teacher",
    "created_at": "2026-04-01T10:00:00Z"
  }
}
```

#### 学生申请使用分身（改造）

**POST** `/api/relations/apply`

**请求体**：
```json
{
  "teacher_persona_id": 3
}
```

#### 获取师生关系列表（改造）

**GET** `/api/relations`

**成功响应** `200`（教师视角）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "student_persona_id": 5,
        "student_nickname": "小李",
        "class_name": "高一(3)班",
        "status": "approved",
        "initiated_by": "teacher",
        "created_at": "2026-04-01T10:00:00Z"
      }
    ],
    "total": 2,
    "page": 1,
    "page_size": 20
  }
}
```

**成功响应** `200`（学生视角）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "teacher_persona_id": 3,
        "teacher_nickname": "王老师",
        "teacher_school": "北京大学",
        "teacher_description": "物理学教授",
        "status": "approved",
        "initiated_by": "teacher",
        "created_at": "2026-04-01T10:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 2.11 评语接口（改造）

**POST** `/api/comments`

**请求体**：
```json
{
  "student_persona_id": 5,
  "content": "该生学习态度认真...",
  "progress_summary": "牛顿定律掌握80%..."
}
```

---

### 2.12 问答风格接口（改造）

**PUT** `/api/students/:id/dialogue-style`

**改造说明**：路径中的 `:id` 改为学生分身 ID（persona_id）。

---

### 2.13 作业接口（改造）

**POST** `/api/assignments`

**请求体**：
```json
{
  "teacher_persona_id": 3,
  "title": "牛顿定律作业",
  "content": "牛顿第一定律是指..."
}
```

---

## 3. 分身管理接口

### 3.1 创建分身

**POST** `/api/personas`

**鉴权**：需要（Bearer Token）

**请求体**（教师分身）：
```json
{
  "role": "teacher",
  "nickname": "王老师",
  "school": "北京大学",
  "description": "物理学教授，专注力学和热力学教学"
}
```

**请求体**（学生分身）：
```json
{
  "role": "student",
  "nickname": "音乐培训班学生"
}
```

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
|------|------|------|----------|------|
| role | string | ✅ | 枚举：teacher/student | 分身角色 |
| nickname | string | ✅ | 1-64 字符 | 分身昵称 |
| school | string | 教师 ✅ | 1-128 字符 | 学校名称 |
| description | string | 教师 ✅ | 1-500 字符 | 分身描述 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 3,
    "user_id": 1,
    "role": "teacher",
    "nickname": "王老师",
    "school": "北京大学",
    "description": "物理学教授，专注力学和热力学教学",
    "is_active": true,
    "created_at": "2026-04-01T09:00:00Z",
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

> **注意**：创建分身后返回新的 JWT token（包含 persona_id），前端需要更新本地 token。

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 教师缺少 school | 40004 | 教师分身必须填写学校名称 |
| 教师缺少 description | 40004 | 教师分身必须填写分身描述 |
| 同名+同校教师分身已存在 | 40015 | 该学校已有同名教师分身，请修改名称 |

---

### 3.2 获取分身列表

**GET** `/api/personas`

**鉴权**：需要（Bearer Token）

**Query 参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| role | string | ❌ | 按角色筛选：teacher/student |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 3,
        "role": "teacher",
        "nickname": "王老师",
        "school": "北京大学",
        "description": "物理学教授",
        "is_active": true,
        "student_count": 15,
        "document_count": 8,
        "class_count": 2,
        "created_at": "2026-04-01T09:00:00Z"
      },
      {
        "id": 5,
        "role": "student",
        "nickname": "音乐培训班学生",
        "school": "",
        "description": "",
        "is_active": true,
        "teacher_count": 1,
        "created_at": "2026-04-01T10:00:00Z"
      }
    ],
    "current_persona_id": 3
  }
}
```

---

### 3.3 编辑分身

**PUT** `/api/personas/:id`

**鉴权**：需要（Bearer Token，本人分身）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 分身 ID |

**请求体**：
```json
{
  "nickname": "王教授",
  "school": "北京大学",
  "description": "物理学教授，专注量子力学教学"
}
```

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 3,
    "nickname": "王教授",
    "school": "北京大学",
    "description": "物理学教授，专注量子力学教学",
    "updated_at": "2026-04-01T11:00:00Z"
  }
}
```

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 分身不存在 | 40013 | 分身不存在 |
| 分身不属于当前用户 | 40014 | 分身不属于当前用户 |
| 同名+同校冲突 | 40015 | 该学校已有同名教师分身 |

---

### 3.4 启用分身

**PUT** `/api/personas/:id/activate`

**鉴权**：需要（Bearer Token，本人分身）

**请求体**：无

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 3,
    "is_active": true,
    "updated_at": "2026-04-01T11:00:00Z"
  }
}
```

---

### 3.5 停用分身

**PUT** `/api/personas/:id/deactivate`

**鉴权**：需要（Bearer Token，本人分身）

**请求体**：无

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 3,
    "is_active": false,
    "updated_at": "2026-04-01T11:00:00Z"
  }
}
```

---

### 3.6 切换分身

**PUT** `/api/personas/:id/switch`

**鉴权**：需要（Bearer Token，本人分身）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 目标分身 ID |

**请求体**：无

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "persona_id": 5,
    "role": "student",
    "nickname": "音乐培训班学生",
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

> **注意**：切换分身后返回新的 JWT token，前端需要更新本地 token 和用户信息。

---

## 4. 班级管理接口

### 4.1 创建班级

**POST** `/api/classes`

**鉴权**：需要（Bearer Token，角色：teacher）

**请求体**：
```json
{
  "name": "高一(3)班",
  "description": "2026级高一3班"
}
```

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
|------|------|------|----------|------|
| name | string | ✅ | 1-64 字符 | 班级名称 |
| description | string | ❌ | 最长 200 字符 | 班级描述 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "persona_id": 3,
    "name": "高一(3)班",
    "description": "2026级高一3班",
    "created_at": "2026-04-01T09:00:00Z"
  }
}
```

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 同名班级已存在 | 40018 | 同名班级已存在 |

---

### 4.2 获取班级列表

**GET** `/api/classes`

**鉴权**：需要（Bearer Token，角色：teacher）

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "name": "高一(3)班",
        "description": "2026级高一3班",
        "member_count": 15,
        "created_at": "2026-04-01T09:00:00Z"
      },
      {
        "id": 2,
        "name": "高二(1)班",
        "description": "",
        "member_count": 8,
        "created_at": "2026-04-01T09:30:00Z"
      }
    ],
    "total": 2
  }
}
```

---

### 4.3 编辑班级

**PUT** `/api/classes/:id`

**鉴权**：需要（Bearer Token，角色：teacher）

**请求体**：
```json
{
  "name": "高一(3)班（理科）",
  "description": "2026级高一3班理科方向"
}
```

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "高一(3)班（理科）",
    "description": "2026级高一3班理科方向",
    "updated_at": "2026-04-01T11:00:00Z"
  }
}
```

---

### 4.4 删除班级

**DELETE** `/api/classes/:id`

**鉴权**：需要（Bearer Token，角色：teacher）

**请求体**：无

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "deleted": true
  }
}
```

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 班级不存在 | 40016 | 班级不存在 |
| 班级不属于当前分身 | 40017 | 班级不属于当前教师分身 |
| 班级有成员 | 40024 | 班级有成员，无法删除，请先移除所有成员 |

---

### 4.5 添加班级成员

**POST** `/api/classes/:id/members`

**鉴权**：需要（Bearer Token，角色：teacher）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 班级 ID |

**请求体**：
```json
{
  "student_persona_id": 5
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| student_persona_id | int | ✅ | 学生分身 ID |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "class_id": 1,
    "student_persona_id": 5,
    "student_nickname": "小李",
    "joined_at": "2026-04-01T10:00:00Z"
  }
}
```

**副作用**：如果师生关系不存在，自动创建 `teacher_student_relation`（status=approved）。

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 班级不存在 | 40016 | 班级不存在 |
| 班级不属于当前分身 | 40017 | 班级不属于当前教师分身 |
| 学生分身不存在 | 40013 | 分身不存在 |
| 学生已在班级中 | 40019 | 学生已在该班级中 |

---

### 4.6 移除班级成员

**DELETE** `/api/classes/:id/members/:member_id`

**鉴权**：需要（Bearer Token，角色：teacher）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 班级 ID |
| member_id | int | 成员记录 ID |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "deleted": true
  }
}
```

> **注意**：移除班级成员不会删除师生关系。

---

### 4.7 获取班级成员列表

**GET** `/api/classes/:id/members`

**鉴权**：需要（Bearer Token，角色：teacher）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 班级 ID |

**Query 参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | ❌ | 1 | 页码 |
| page_size | int | ❌ | 20 | 每页数量 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "student_persona_id": 5,
        "student_nickname": "小李",
        "joined_at": "2026-04-01T10:00:00Z"
      },
      {
        "id": 2,
        "student_persona_id": 6,
        "student_nickname": "小王",
        "joined_at": "2026-04-01T10:05:00Z"
      }
    ],
    "total": 2,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 5. 分身分享接口

### 5.1 生成分享码

**POST** `/api/shares`

**鉴权**：需要（Bearer Token，角色：teacher）

**请求体**：
```json
{
  "class_id": 1,
  "expires_hours": 168,
  "max_uses": 50
}
```

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| class_id | int | ❌ | - | 指定加入的班级（不传则只建立师生关系） |
| expires_hours | int | ❌ | 168（7天） | 过期时间（小时），0=永不过期 |
| max_uses | int | ❌ | 0 | 最大使用次数，0=不限 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "teacher_persona_id": 3,
    "share_code": "ABC12345",
    "class_id": 1,
    "class_name": "高一(3)班",
    "expires_at": "2026-04-08T09:00:00Z",
    "max_uses": 50,
    "used_count": 0,
    "is_active": true,
    "created_at": "2026-04-01T09:00:00Z"
  }
}
```

---

### 5.2 获取分享码列表

**GET** `/api/shares`

**鉴权**：需要（Bearer Token，角色：teacher）

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "share_code": "ABC12345",
        "class_id": 1,
        "class_name": "高一(3)班",
        "expires_at": "2026-04-08T09:00:00Z",
        "max_uses": 50,
        "used_count": 12,
        "is_active": true,
        "created_at": "2026-04-01T09:00:00Z"
      }
    ],
    "total": 3
  }
}
```

---

### 5.3 停用分享码

**PUT** `/api/shares/:id/deactivate`

**鉴权**：需要（Bearer Token，角色：teacher）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 分享码记录 ID |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "is_active": false
  }
}
```

---

### 5.4 获取分享码信息（预览）

**GET** `/api/shares/:code/info`

**鉴权**：需要（Bearer Token，但不限角色）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| code | string | 分享码 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "teacher_persona_id": 3,
    "teacher_nickname": "王老师",
    "teacher_school": "北京大学",
    "teacher_description": "物理学教授，专注力学和热力学教学",
    "class_name": "高一(3)班",
    "is_valid": true
  }
}
```

**分享码无效时**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "is_valid": false,
    "reason": "分享码已过期"
  }
}
```

---

### 5.5 通过分享码加入

**POST** `/api/shares/:code/join`

**鉴权**：需要（Bearer Token）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| code | string | 分享码 |

**请求体**：
```json
{
  "student_persona_id": 5
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| student_persona_id | int | ❌ | 使用哪个学生分身加入（不传则自动选择） |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "relation_id": 10,
    "teacher_persona_id": 3,
    "teacher_nickname": "王老师",
    "teacher_school": "北京大学",
    "student_persona_id": 5,
    "class_id": 1,
    "class_name": "高一(3)班"
  }
}
```

**需要选择学生分身时** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "need_select_persona": true,
    "student_personas": [
      {
        "id": 5,
        "nickname": "音乐培训班学生"
      },
      {
        "id": 8,
        "nickname": "数学提高班学生"
      }
    ]
  }
}
```

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 分享码无效 | 40020 | 分享码无效或已过期 |
| 使用次数已达上限 | 40021 | 分享码使用次数已达上限 |
| 没有学生分身 | 40022 | 需要先创建学生分身才能加入 |
| 已经关联 | 40009 | 师生关系已存在 |

---

## 6. 获取用户信息（增强）

### 6.1 获取用户信息

**GET** `/api/user/profile`

**改造说明**：返回当前分身信息和分身列表。

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 1,
    "username": "wx_user_abc",
    "current_persona": {
      "id": 3,
      "role": "teacher",
      "nickname": "王老师",
      "school": "北京大学",
      "description": "物理学教授"
    },
    "personas": [
      {
        "id": 3,
        "role": "teacher",
        "nickname": "王老师",
        "school": "北京大学",
        "description": "物理学教授"
      },
      {
        "id": 5,
        "role": "student",
        "nickname": "音乐培训班学生",
        "school": "",
        "description": ""
      }
    ],
    "created_at": "2026-04-01T09:00:00Z"
  }
}
```

---

## 7. 接口总览

### 7.1 新增接口（18 个）

| 编号 | 方法 | 路径 | 说明 | 角色 |
|------|------|------|------|------|
| API-20 | POST | `/api/personas` | 创建分身 | 登录用户 |
| API-21 | GET | `/api/personas` | 获取分身列表 | 登录用户 |
| API-22 | PUT | `/api/personas/:id` | 编辑分身 | 登录用户 |
| API-23 | PUT | `/api/personas/:id/activate` | 启用分身 | 登录用户 |
| API-24 | PUT | `/api/personas/:id/deactivate` | 停用分身 | 登录用户 |
| API-25 | PUT | `/api/personas/:id/switch` | 切换分身 | 登录用户 |
| API-26 | POST | `/api/classes` | 创建班级 | teacher |
| API-27 | GET | `/api/classes` | 获取班级列表 | teacher |
| API-28 | PUT | `/api/classes/:id` | 编辑班级 | teacher |
| API-29 | DELETE | `/api/classes/:id` | 删除班级 | teacher |
| API-30 | POST | `/api/classes/:id/members` | 添加班级成员 | teacher |
| API-31 | DELETE | `/api/classes/:id/members/:member_id` | 移除班级成员 | teacher |
| API-32 | GET | `/api/classes/:id/members` | 获取班级成员列表 | teacher |
| API-33 | POST | `/api/shares` | 生成分享码 | teacher |
| API-34 | GET | `/api/shares` | 获取分享码列表 | teacher |
| API-35 | PUT | `/api/shares/:id/deactivate` | 停用分享码 | teacher |
| API-36 | GET | `/api/shares/:code/info` | 获取分享码信息 | 登录用户 |
| API-37 | POST | `/api/shares/:code/join` | 通过分享码加入 | 登录用户 |

### 7.2 改造接口（13 个）

| 接口 | 改造内容 |
|------|----------|
| `POST /api/auth/wx-login` | 返回分身列表 + 当前分身 |
| `POST /api/auth/complete-profile` | 内部转为创建分身 |
| `POST /api/chat` | teacher_id → teacher_persona_id |
| `POST /api/chat/stream` | teacher_id → teacher_persona_id |
| `GET /api/teachers` | 返回教师分身列表 |
| `POST /api/documents` | 新增 scope / scope_id |
| `GET /api/documents` | 新增 scope 筛选 |
| `POST /api/documents/upload` | 新增 scope / scope_id |
| `POST /api/documents/import-url` | 新增 scope / scope_id |
| `POST /api/relations/invite` | student_id → student_persona_id |
| `POST /api/relations/apply` | teacher_id → teacher_persona_id |
| `GET /api/relations` | 使用 persona_id |
| `GET /api/user/profile` | 返回分身信息 |

---

## 8. 路由注册参考

```go
// router.go 新增路由

// 分身管理
personas := authorized.Group("/personas")
{
    personas.POST("", handler.HandleCreatePersona)
    personas.GET("", handler.HandleGetPersonas)
    personas.PUT("/:id", handler.HandleEditPersona)
    personas.PUT("/:id/activate", handler.HandleActivatePersona)
    personas.PUT("/:id/deactivate", handler.HandleDeactivatePersona)
    personas.PUT("/:id/switch", handler.HandleSwitchPersona)
}

// 班级管理
classes := authorized.Group("/classes")
{
    classes.POST("", auth.RoleRequired("teacher"), handler.HandleCreateClass)
    classes.GET("", auth.RoleRequired("teacher"), handler.HandleGetClasses)
    classes.PUT("/:id", auth.RoleRequired("teacher"), handler.HandleEditClass)
    classes.DELETE("/:id", auth.RoleRequired("teacher"), handler.HandleDeleteClass)
    classes.POST("/:id/members", auth.RoleRequired("teacher"), handler.HandleAddClassMember)
    classes.DELETE("/:id/members/:member_id", auth.RoleRequired("teacher"), handler.HandleRemoveClassMember)
    classes.GET("/:id/members", auth.RoleRequired("teacher"), handler.HandleGetClassMembers)
}

// 分享码
shares := authorized.Group("/shares")
{
    shares.POST("", auth.RoleRequired("teacher"), handler.HandleCreateShare)
    shares.GET("", auth.RoleRequired("teacher"), handler.HandleGetShares)
    shares.PUT("/:id/deactivate", auth.RoleRequired("teacher"), handler.HandleDeactivateShare)
    shares.GET("/:code/info", handler.HandleGetShareInfo)
    shares.POST("/:code/join", handler.HandleJoinByShare)
}
```

---

**文档版本**: v1.0.0
**创建日期**: 2026-03-29
**最后更新**: 2026-03-29
