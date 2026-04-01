# V2.0 迭代3 API 接口规范

## 1. 通用约定

> 继承 V2.0 迭代2 的通用约定，以下仅列出新增和变更部分。

### 1.1 基础信息

| 项目 | 说明 |
|------|------|
| Base URL | `http://localhost:8080` |
| 协议 | HTTP（开发环境） |
| 数据格式 | JSON（`Content-Type: application/json`） |
| 认证方式 | Bearer Token（`Authorization: Bearer <token>`） |

### 1.2 新增错误码

| 错误码 | HTTP Status | 说明 |
|--------|-------------|------|
| 40025 | 403 | 教师分身已停用，无法发起对话 |
| 40026 | 403 | 班级已停用，无法发起对话 |
| 40027 | 403 | 学生访问权限已关闭，无法发起对话 |
| 40028 | 400 | 预览 ID 无效或已过期 |

---

## 2. 新增接口

### 2.1 教师分身仪表盘

**GET** `/api/personas/:id/dashboard`

**鉴权**：需要（Bearer Token，本人分身）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 教师分身 ID |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "persona": {
      "id": 3,
      "nickname": "王老师",
      "school": "北京大学",
      "description": "物理学教授",
      "is_active": true
    },
    "pending_count": 3,
    "classes": [
      {
        "id": 1,
        "name": "高一(3)班",
        "description": "2026级高一3班",
        "member_count": 32,
        "is_active": true
      },
      {
        "id": 2,
        "name": "高二(1)班",
        "description": "",
        "member_count": 28,
        "is_active": true
      }
    ],
    "latest_share": {
      "id": 1,
      "share_code": "ABC12345",
      "class_id": 1,
      "class_name": "高一(3)班",
      "used_count": 12,
      "max_uses": 50,
      "is_active": true,
      "expires_at": "2026-04-08T09:00:00Z"
    },
    "stats": {
      "total_students": 60,
      "total_documents": 15,
      "total_classes": 2
    }
  }
}
```

**字段说明**：

| 字段 | 说明 |
|------|------|
| persona | 当前教师分身基本信息 |
| pending_count | 待审批的学生申请数量 |
| classes | 该分身下的班级列表（含成员数和启停状态） |
| latest_share | 最近一个有效的分享码（如果没有则为 null） |
| stats | 统计信息（总学生数、总文档数、总班级数） |

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 分身不存在 | 40013 | 分身不存在 |
| 分身不属于当前用户 | 40014 | 分身不属于当前用户 |

---

### 2.2 启用/停用班级

**PUT** `/api/classes/:id/toggle`

**鉴权**：需要（Bearer Token，角色：teacher）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 班级 ID |

**请求体**：
```json
{
  "is_active": false
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| is_active | bool | ✅ | true=启用，false=停用 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "高一(3)班",
    "is_active": false,
    "affected_students": 32,
    "updated_at": "2026-04-15T10:00:00Z"
  }
}
```

**字段说明**：

| 字段 | 说明 |
|------|------|
| affected_students | 受影响的学生数量（用于前端展示确认信息） |

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 班级不存在 | 40016 | 班级不存在 |
| 班级不属于当前分身 | 40017 | 班级不属于当前教师分身 |

---

### 2.3 启用/停用学生访问权限

**PUT** `/api/relations/:id/toggle`

**鉴权**：需要（Bearer Token，角色：teacher）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 师生关系 ID |

**请求体**：
```json
{
  "is_active": false
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| is_active | bool | ✅ | true=启用，false=停用 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "student_persona_id": 5,
    "student_nickname": "张三",
    "is_active": false,
    "updated_at": "2026-04-15T10:00:00Z"
  }
}
```

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 关系不存在 | 40007 | 师生关系不存在 |
| 关系不属于当前教师分身 | 40014 | 无权操作该师生关系 |

---

### 2.4 文档预览（文本）

**POST** `/api/documents/preview`

**鉴权**：需要（Bearer Token，角色：teacher）

**请求体**：
```json
{
  "title": "牛顿运动定律",
  "content": "牛顿第一定律，也称为惯性定律...",
  "tags": "物理,力学"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | ✅ | 文档标题 |
| content | string | ✅ | 文档内容 |
| tags | string | ❌ | 标签（逗号分隔） |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "preview_id": "tmp_abc123def456",
    "title": "牛顿运动定律",
    "tags": "物理,力学",
    "total_chars": 3500,
    "chunks": [
      {
        "index": 0,
        "content": "牛顿第一定律，也称为惯性定律，是经典力学的基础...",
        "char_count": 850
      },
      {
        "index": 1,
        "content": "牛顿第二定律描述了力与加速度之间的关系...",
        "char_count": 920
      },
      {
        "index": 2,
        "content": "牛顿第三定律指出，对于每一个作用力...",
        "char_count": 880
      },
      {
        "index": 3,
        "content": "牛顿运动定律的应用范围非常广泛...",
        "char_count": 850
      }
    ],
    "chunk_count": 4
  }
}
```

**字段说明**：

| 字段 | 说明 |
|------|------|
| preview_id | 预览 ID，用于后续确认入库（有效期 30 分钟） |
| total_chars | 文档总字符数 |
| chunks | 切片列表 |
| chunks[].index | 切片序号（从 0 开始） |
| chunks[].content | 切片内容（完整内容） |
| chunks[].char_count | 切片字符数 |
| chunk_count | 切片总数 |

---

### 2.5 文档预览（文件上传）

**POST** `/api/documents/preview-upload`

**鉴权**：需要（Bearer Token，角色：teacher）

**Content-Type**: `multipart/form-data`

**表单字段**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | ✅ | 文件（PDF/DOCX/TXT/MD） |
| title | string | ❌ | 文档标题（不传则使用文件名） |
| tags | string | ❌ | 标签（逗号分隔） |

**成功响应** `200`：同 2.4 格式。

---

### 2.6 文档预览（URL 导入）

**POST** `/api/documents/preview-url`

**鉴权**：需要（Bearer Token，角色：teacher）

**请求体**：
```json
{
  "url": "https://example.com/article",
  "title": "可选标题",
  "tags": "物理,力学"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| url | string | ✅ | 网页 URL |
| title | string | ❌ | 文档标题（不传则从网页提取） |
| tags | string | ❌ | 标签（逗号分隔） |

**成功响应** `200`：同 2.4 格式。

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| URL 不可达 | 40012 | URL 不可达或解析失败 |

---

### 2.7 确认文档入库

**POST** `/api/documents/confirm`

**鉴权**：需要（Bearer Token，角色：teacher）

**请求体**：
```json
{
  "preview_id": "tmp_abc123def456",
  "title": "牛顿运动定律",
  "tags": "物理,力学",
  "scope": "class",
  "scope_ids": [1, 2]
}
```

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| preview_id | string | ✅ | - | 预览 ID |
| title | string | ❌ | - | 文档标题（可在确认时修改） |
| tags | string | ❌ | - | 标签（可在确认时修改） |
| scope | string | ❌ | global | 作用域：global / class / student |
| scope_ids | int[] | scope≠global 时 ✅ | [] | 班级 ID 列表或学生分身 ID 列表 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "documents": [
      {
        "id": 10,
        "title": "牛顿运动定律",
        "scope": "class",
        "scope_id": 1,
        "scope_name": "高一(3)班",
        "chunks_count": 4
      },
      {
        "id": 11,
        "title": "牛顿运动定律",
        "scope": "class",
        "scope_id": 2,
        "scope_name": "高二(1)班",
        "chunks_count": 4
      }
    ],
    "total_documents": 2
  }
}
```

> **说明**：当 scope_ids 包含多个 ID 时，为每个 ID 创建一条 document 记录（相同内容，不同 scope_id），返回所有创建的文档。

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 预览 ID 无效或已过期 | 40028 | 预览 ID 无效或已过期，请重新预览 |
| 无效的 scope 值 | 40023 | 无效的知识库作用域 |
| 班级不属于当前分身 | 40017 | 班级不属于当前教师分身 |

---

## 3. 改造接口

### 3.1 获取教师列表（改造）

**GET** `/api/teachers`

**改造说明**：学生视角下，仅返回与当前学生分身有 approved + is_active 关系的教师分身。

**当前行为**：返回所有教师分身列表。

**改造后行为**：
- **教师角色调用**：返回所有教师分身列表（不变）
- **学生角色调用**：仅返回与当前学生分身有 `teacher_student_relations.status='approved' AND teacher_student_relations.is_active=1` 关系的教师分身

**成功响应** `200`（学生视角）：
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
        "description": "物理学教授",
        "document_count": 5,
        "created_at": "2026-04-01T09:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 3.2 添加知识文档（改造）

**POST** `/api/documents`

**改造说明**：新增 `scope_ids` 字段，支持多班级/多学生选择。向后兼容 `scope_id`。

**请求体**：
```json
{
  "title": "牛顿运动定律",
  "content": "牛顿第一定律...",
  "tags": "物理,力学",
  "scope": "class",
  "scope_ids": [1, 2]
}
```

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| title | string | ✅ | - | 文档标题 |
| content | string | ✅ | - | 文档内容 |
| tags | string | ❌ | - | 标签（逗号分隔） |
| scope | string | ❌ | global | 作用域：global / class / student |
| scope_ids | int[] | scope≠global 时 ✅ | [] | 班级 ID 列表或学生分身 ID 列表 |
| scope_id | int | ❌ | 0 | **向后兼容**：单个 scope_id |

**向后兼容规则**：
- 如果请求中同时包含 `scope_ids` 和 `scope_id`，优先使用 `scope_ids`
- 如果只有 `scope_id`（旧格式），自动转换为 `scope_ids: [scope_id]`
- 如果 `scope_ids` 为空且 `scope_id` 为 0，且 scope 不是 global，返回错误

**成功响应** `200`（多 scope_id 时返回多条文档）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "documents": [
      {
        "id": 10,
        "title": "牛顿运动定律",
        "scope": "class",
        "scope_id": 1,
        "chunks_count": 4
      },
      {
        "id": 11,
        "title": "牛顿运动定律",
        "scope": "class",
        "scope_id": 2,
        "chunks_count": 4
      }
    ]
  }
}
```

**成功响应** `200`（单 scope_id，向后兼容格式）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "document_id": 10,
    "title": "牛顿运动定律",
    "chunks_count": 4
  }
}
```

> **说明**：当 scope_ids 只有 1 个元素时，返回旧格式（单条文档），保持向后兼容。

---

### 3.3 文件上传（改造）

**POST** `/api/documents/upload`

**改造说明**：新增 `scope_ids` 表单字段。

**表单字段**（新增）：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| scope_ids | string | scope≠global 时 ✅ | 逗号分隔的 ID 列表，如 "1,2,3" |

> **说明**：multipart/form-data 不支持数组类型，使用逗号分隔的字符串表示。

---

### 3.4 URL 导入（改造）

**POST** `/api/documents/import-url`

**改造说明**：新增 `scope_ids` 字段。

**请求体**（新增字段）：
```json
{
  "url": "https://example.com/article",
  "title": "可选标题",
  "tags": "物理,力学",
  "scope": "class",
  "scope_ids": [1, 2]
}
```

---

### 3.5 发送对话消息（改造）

**POST** `/api/chat`

**改造说明**：对话鉴权增加启停状态检查。

**新增错误响应**：
| 场景 | code | message |
|------|------|---------|
| 教师分身已停用 | 40025 | 教师分身已停用，无法发起对话 |
| 班级已停用 | 40026 | 您所在的班级已停用，无法发起对话 |
| 学生访问权限已关闭 | 40027 | 您的访问权限已关闭，请联系教师 |

**鉴权逻辑**（改造后）：
```
学生发起对话:
  1. 检查 teacher_student_relations.status == 'approved'
     → 否 → 40007 "未获得该教师授权"
  2. 检查 teacher_student_relations.is_active == 1
     → 否 → 40027 "您的访问权限已关闭，请联系教师"
  3. 检查教师分身 personas.is_active == 1
     → 否 → 40025 "教师分身已停用，无法发起对话"
  4. 查询学生在该教师分身下的班级
     → 如果有班级，检查 classes.is_active == 1
     → 否 → 40026 "您所在的班级已停用，无法发起对话"
     → 如果不在任何班级，跳过此检查
  5. 所有检查通过 → 允许对话
```

---

### 3.6 SSE 流式对话（改造）

**POST** `/api/chat/stream`

**改造说明**：同 3.5，对话鉴权增加启停状态检查。

---

### 3.7 获取班级列表（改造）

**GET** `/api/classes`

**改造说明**：返回结果新增 `is_active` 字段。

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
        "is_active": true,
        "created_at": "2026-04-01T09:00:00Z"
      }
    ],
    "total": 2
  }
}
```

---

### 3.8 获取师生关系列表（改造）

**GET** `/api/relations`

**改造说明**：返回结果新增 `is_active` 字段。

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
        "is_active": true,
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

---

## 4. 接口总览

### 4.1 新增接口（5 个）

| 编号 | 方法 | 路径 | 说明 | 角色 |
|------|------|------|------|------|
| API-38 | GET | `/api/personas/:id/dashboard` | 教师分身仪表盘 | 登录用户（本人分身） |
| API-39 | PUT | `/api/classes/:id/toggle` | 启用/停用班级 | teacher |
| API-40 | PUT | `/api/relations/:id/toggle` | 启用/停用学生访问权限 | teacher |
| API-41 | POST | `/api/documents/preview` | 文档预览（文本） | teacher |
| API-42 | POST | `/api/documents/preview-upload` | 文档预览（文件上传） | teacher |
| API-43 | POST | `/api/documents/preview-url` | 文档预览（URL 导入） | teacher |
| API-44 | POST | `/api/documents/confirm` | 确认文档入库 | teacher |

### 4.2 改造接口（7 个）

| 接口 | 改造内容 |
|------|----------|
| `GET /api/teachers` | 学生视角仅返回已授权+启用的教师分身 |
| `POST /api/documents` | 新增 scope_ids 多选支持 |
| `POST /api/documents/upload` | 新增 scope_ids 多选支持 |
| `POST /api/documents/import-url` | 新增 scope_ids 多选支持 |
| `POST /api/chat` | 对话鉴权增加启停状态检查 |
| `POST /api/chat/stream` | 对话鉴权增加启停状态检查 |
| `GET /api/classes` | 返回 is_active 字段 |
| `GET /api/relations` | 返回 is_active 字段 |

---

## 5. 路由注册参考

```go
// router.go 新增路由（V2.0 迭代3）

// 教师仪表盘
personas.GET("/:id/dashboard", handler.HandleGetPersonaDashboard)

// 班级启停
classes.PUT("/:id/toggle", auth.RoleRequired("teacher"), handler.HandleToggleClass)

// 师生关系启停
relations.PUT("/:id/toggle", auth.RoleRequired("teacher"), handler.HandleToggleRelation)

// 文档预览和确认
authorized.POST("/documents/preview", auth.RoleRequired("teacher", "admin"), handler.HandlePreviewDocument)
authorized.POST("/documents/preview-upload", auth.RoleRequired("teacher", "admin"), handler.HandlePreviewUpload)
authorized.POST("/documents/preview-url", auth.RoleRequired("teacher", "admin"), handler.HandlePreviewURL)
authorized.POST("/documents/confirm", auth.RoleRequired("teacher", "admin"), handler.HandleConfirmDocument)
```

---

**文档版本**: v1.0.0
**创建日期**: 2026-03-30
**最后更新**: 2026-03-30
