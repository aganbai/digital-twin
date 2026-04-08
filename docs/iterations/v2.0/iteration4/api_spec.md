# V2.0 迭代4 API 接口规范

## 1. 通用约定

> 继承 V2.0 迭代3 的通用约定，以下仅列出新增和变更部分。

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
| 40029 | 403 | 该分享码仅对特定学生可用 |
| 40030 | 200 | 当前会话已被教师接管，请等待教师回复（降级响应） |
| 40031 | 400 | 接管记录不存在或已结束 |
| 40032 | 403 | 无权操作该会话（非该分身的教师） |
| 40033 | 200 | LLM 摘要生成失败（降级响应，字段为空） |

---

## 2. 新增接口

### 2.1 获取分身广场列表

**GET** `/api/personas/marketplace`

**鉴权**：需要（Bearer Token）

**查询参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| keyword | string | ❌ | - | 搜索关键词（匹配昵称或学校） |
| page | int | ❌ | 1 | 页码 |
| page_size | int | ❌ | 20 | 每页数量（最大 50） |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 5,
        "nickname": "李老师",
        "school": "清华大学",
        "description": "高中数学",
        "student_count": 45,
        "document_count": 12,
        "application_status": ""
      },
      {
        "id": 8,
        "nickname": "张老师",
        "school": "复旦大学",
        "description": "英语教学",
        "student_count": 30,
        "document_count": 8,
        "application_status": "pending"
      }
    ],
    "total": 15,
    "page": 1,
    "page_size": 20
  }
}
```

**字段说明**：

| 字段 | 说明 |
|------|------|
| id | 教师分身 ID |
| nickname | 教师昵称 |
| school | 学校 |
| description | 分身描述 |
| student_count | 学生数量 |
| document_count | 知识库文档数量 |
| application_status | 当前学生对该教师的申请状态：空字符串（未申请）/ `pending`（申请中）/ `approved`（已授权，不应出现在广场） |

> **说明**：
> - 仅返回 `is_public=1 AND is_active=1` 的教师分身
> - 排除当前学生已有 `approved` 关系的教师分身（这些在"我的老师"中展示）
> - `application_status` 用于前端展示按钮状态（"申请使用" / "申请中..."）

---

### 2.2 设置分身公开/私有

**PUT** `/api/personas/:id/visibility`

**鉴权**：需要（Bearer Token，本人分身）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 分身 ID |

**请求体**：
```json
{
  "is_public": true
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| is_public | bool | ✅ | true=公开到广场，false=私有 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 3,
    "nickname": "王老师",
    "is_public": true,
    "updated_at": "2026-03-31T10:00:00Z"
  }
}
```

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 分身不存在 | 40013 | 分身不存在 |
| 分身不属于当前用户 | 40014 | 分身不属于当前用户 |
| 仅教师分身可公开 | 40032 | 仅教师分身可设置公开状态 |

---

### 2.3 搜索已注册学生

**GET** `/api/students/search`

**鉴权**：需要（Bearer Token，角色：teacher）

**查询参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| keyword | string | ✅ | - | 搜索关键词（匹配学生昵称，至少 2 个字符） |
| page | int | ❌ | 1 | 页码 |
| page_size | int | ❌ | 20 | 每页数量（最大 50） |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "persona_id": 5,
        "user_id": 10,
        "nickname": "小明",
        "created_at": "2026-03-20T09:00:00Z"
      },
      {
        "persona_id": 8,
        "user_id": 15,
        "nickname": "小红",
        "created_at": "2026-03-22T14:00:00Z"
      }
    ],
    "total": 2,
    "page": 1,
    "page_size": 20
  }
}
```

> **说明**：仅返回 `role='student' AND is_active=1` 的分身。

---

### 2.4 教师真人回复

**POST** `/api/chat/teacher-reply`

**鉴权**：需要（Bearer Token，角色：teacher）

**请求体**：
```json
{
  "student_persona_id": 5,
  "session_id": "uuid-session-123",
  "content": "同学你好！惯性其实就是物体保持原来运动状态的性质...",
  "reply_to_id": 42
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| student_persona_id | int | ✅ | 学生分身 ID |
| session_id | string | ✅ | 会话 ID |
| content | string | ✅ | 回复内容 |
| reply_to_id | int | ❌ | 引用的消息 ID（0 或不传表示不引用） |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "conversation_id": 100,
    "sender_type": "teacher",
    "reply_to_id": 42,
    "reply_to_content": "我还是不太理解惯性的概念",
    "takeover_status": "active",
    "created_at": "2026-03-31T14:30:00Z"
  }
}
```

**字段说明**：

| 字段 | 说明 |
|------|------|
| conversation_id | 新创建的对话记录 ID |
| sender_type | 发送者类型（固定为 `teacher`） |
| reply_to_id | 引用的消息 ID |
| reply_to_content | 引用的消息内容摘要（前 100 字符） |
| takeover_status | 接管状态（首次回复自动变为 `active`） |

**业务逻辑**：
1. 验证教师分身与学生分身有 approved 关系
2. 验证 session_id 属于该学生
3. 如果当前无 active 接管记录，自动创建（status='active'）
4. 保存对话记录（role='assistant', sender_type='teacher', reply_to_id）
5. 返回成功

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 无师生关系 | 40007 | 未获得该学生的授权关系 |
| 会话不存在 | 40031 | 会话不存在 |
| 无权操作 | 40032 | 无权操作该会话 |

---

### 2.5 查询接管状态

**GET** `/api/chat/takeover-status`

**鉴权**：需要（Bearer Token）

**查询参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| session_id | string | ✅ | 会话 ID |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "is_taken_over": true,
    "teacher_persona_id": 3,
    "teacher_nickname": "王老师",
    "started_at": "2026-03-31T14:25:00Z"
  }
}
```

**字段说明**：

| 字段 | 说明 |
|------|------|
| is_taken_over | 是否处于接管状态 |
| teacher_persona_id | 接管的教师分身 ID（未接管时为 0） |
| teacher_nickname | 接管的教师昵称（未接管时为空） |
| started_at | 接管开始时间（未接管时为空） |

> **说明**：学生和教师都可以调用此接口查询接管状态。

---

### 2.6 教师退出接管

**POST** `/api/chat/end-takeover`

**鉴权**：需要（Bearer Token，角色：teacher）

**请求体**：
```json
{
  "session_id": "uuid-session-123"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| session_id | string | ✅ | 会话 ID |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "session_id": "uuid-session-123",
    "status": "ended",
    "ended_at": "2026-03-31T15:00:00Z"
  }
}
```

**业务逻辑**：
1. 查找该 session_id 的 active 接管记录
2. 验证当前教师是接管者
3. 更新 status='ended', ended_at=NOW()
4. 返回成功

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 无活跃接管记录 | 40031 | 接管记录不存在或已结束 |
| 非接管教师 | 40032 | 无权操作该会话 |

---

### 2.7 教师查看学生对话记录

**GET** `/api/conversations/student/:student_persona_id`

**鉴权**：需要（Bearer Token，角色：teacher）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| student_persona_id | int | 学生分身 ID |

**查询参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| session_id | string | ❌ | - | 指定会话 ID（不传则返回最近会话） |
| page | int | ❌ | 1 | 页码 |
| page_size | int | ❌ | 50 | 每页数量（最大 100） |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "student_persona_id": 5,
    "student_nickname": "小明",
    "session_id": "uuid-session-123",
    "takeover_status": "active",
    "messages": [
      {
        "id": 40,
        "role": "user",
        "content": "牛顿第一定律是什么？",
        "sender_type": "student",
        "reply_to_id": 0,
        "created_at": "2026-03-31T14:00:00Z"
      },
      {
        "id": 41,
        "role": "assistant",
        "content": "这是一个很好的问题！让我们一起来思考...",
        "sender_type": "ai",
        "reply_to_id": 0,
        "created_at": "2026-03-31T14:00:05Z"
      },
      {
        "id": 42,
        "role": "user",
        "content": "我还是不太理解惯性的概念",
        "sender_type": "student",
        "reply_to_id": 0,
        "created_at": "2026-03-31T14:05:00Z"
      },
      {
        "id": 100,
        "role": "assistant",
        "content": "同学你好！惯性其实就是物体保持原来运动状态的性质...",
        "sender_type": "teacher",
        "reply_to_id": 42,
        "reply_to_content": "我还是不太理解惯性的概念",
        "created_at": "2026-03-31T14:30:00Z"
      }
    ],
    "total": 4,
    "page": 1,
    "page_size": 50
  }
}
```

**字段说明**：

| 字段 | 说明 |
|------|------|
| takeover_status | 当前会话的接管状态：`active`（接管中）/ `ended`（已结束）/ `none`（未接管） |
| messages[].sender_type | 消息发送者类型：`student` / `ai` / `teacher` |
| messages[].reply_to_id | 引用的消息 ID（0 表示非引用回复） |
| messages[].reply_to_content | 引用的消息内容摘要（仅 reply_to_id > 0 时返回） |

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 无师生关系 | 40007 | 未获得该学生的授权关系 |
| 学生分身不存在 | 40013 | 分身不存在 |

---

## 3. 改造接口

### 3.1 创建分享码（改造）

**POST** `/api/shares`

**改造说明**：新增 `target_student_persona_id` 字段，支持定向邀请。

**请求体**（新增字段）：
```json
{
  "class_id": 1,
  "max_uses": 50,
  "expires_hours": 168,
  "target_student_persona_id": 5
}
```

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| target_student_persona_id | int | ❌ | 0 | 目标学生分身 ID（0=不限定，>0=仅该学生可用） |

**成功响应** `200`（新增字段）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 10,
    "share_code": "XYZ789",
    "class_id": 1,
    "class_name": "高一(3)班",
    "target_student_persona_id": 5,
    "target_student_nickname": "小明",
    "max_uses": 50,
    "used_count": 0,
    "is_active": true,
    "expires_at": "2026-04-07T10:00:00Z",
    "created_at": "2026-03-31T10:00:00Z"
  }
}
```

> **向后兼容**：`target_student_persona_id` 不传或为 0 时，行为与旧版一致。

---

### 3.2 使用分享码加入（改造）

**POST** `/api/shares/:code/join`

**改造说明**：增加目标学生校验。

**新增错误响应**：
| 场景 | code | message |
|------|------|---------|
| 非目标学生 | 40029 | 该分享码仅对特定学生可用 |

**校验逻辑**：
```
学生使用分享码:
  1. 查询分享码信息
  2. 如果 target_student_persona_id > 0:
     → 检查当前学生分身 ID == target_student_persona_id
     → 不匹配 → 返回 40029
  3. 如果 target_student_persona_id == 0:
     → 不做限制（向后兼容）
  4. 其他校验不变（过期、使用次数等）
```

---

### 3.3 获取分享码信息（改造）

**GET** `/api/shares/:code/info`

**改造说明**：返回结果新增目标学生信息。

**成功响应** `200`（新增字段）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "teacher_persona_id": 3,
    "teacher_nickname": "王老师",
    "teacher_school": "北京大学",
    "teacher_description": "物理学教授",
    "class_name": "高一(3)班",
    "target_student_persona_id": 5,
    "target_student_nickname": "小明",
    "is_valid": true,
    "reason": ""
  }
}
```

| 字段 | 说明 |
|------|------|
| target_student_persona_id | 目标学生分身 ID（0 表示不限定） |
| target_student_nickname | 目标学生昵称（不限定时为空） |

---

### 3.4 发送对话消息（改造）

**POST** `/api/chat`

**改造说明**：增加接管状态检查。

**新增行为**：
```
学生发送消息:
  1. 原有鉴权检查（授权 + 启停状态）
  2. 🆕 检查 teacher_takeovers 表:
     → 存在 session_id + status='active' 的记录
     → 保存学生消息（sender_type='student'）
     → 不调用 AI 生成回复
     → 返回特殊响应（code=40030）
  3. 无接管 → 正常 AI 回复流程
     → 保存学生消息（sender_type='student'）
     → AI 回复保存（sender_type='ai'）
```

**接管状态下的响应** `200`：
```json
{
  "code": 40030,
  "message": "老师正在亲自回复中，请等待老师回复",
  "data": {
    "conversation_id": 101,
    "sender_type": "student",
    "takeover_info": {
      "teacher_nickname": "王老师",
      "started_at": "2026-03-31T14:25:00Z"
    }
  }
}
```

> **说明**：HTTP Status 仍为 200，但 code=40030 表示消息已保存但 AI 未回复。前端据此展示提示。

---

### 3.5 SSE 流式对话（改造）

**POST** `/api/chat/stream`

**改造说明**：同 3.4，增加接管状态检查。接管状态下不发起 SSE 流，直接返回 JSON 响应（同 3.4 格式）。

---

### 3.6 获取对话记录（改造）

**GET** `/api/conversations`

**改造说明**：返回结果新增 `sender_type` 和 `reply_to_id` 字段。

**成功响应** `200`（新增字段）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 40,
        "session_id": "uuid-session-123",
        "role": "user",
        "content": "牛顿第一定律是什么？",
        "sender_type": "student",
        "reply_to_id": 0,
        "token_count": 15,
        "created_at": "2026-03-31T14:00:00Z"
      },
      {
        "id": 41,
        "session_id": "uuid-session-123",
        "role": "assistant",
        "content": "这是一个很好的问题！...",
        "sender_type": "ai",
        "reply_to_id": 0,
        "token_count": 120,
        "created_at": "2026-03-31T14:00:05Z"
      },
      {
        "id": 100,
        "session_id": "uuid-session-123",
        "role": "assistant",
        "content": "同学你好！惯性其实就是...",
        "sender_type": "teacher",
        "reply_to_id": 42,
        "reply_to_content": "我还是不太理解惯性的概念",
        "token_count": 0,
        "created_at": "2026-03-31T14:30:00Z"
      }
    ],
    "total": 3,
    "page": 1,
    "page_size": 50
  }
}
```

> **向后兼容**：旧数据的 `sender_type` 为空字符串（数据库迁移会回填为 `student`/`ai`），`reply_to_id` 为 0。

---

### 3.7 获取分身列表（改造）

**GET** `/api/personas`

**改造说明**：返回结果新增 `is_public` 字段。

**成功响应** `200`（新增字段）：
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
        "is_public": true,
        "student_count": 60,
        "document_count": 15,
        "class_count": 2,
        "created_at": "2026-03-20T09:00:00Z"
      }
    ],
    "total": 3
  }
}
```

---

### 3.8 文档预览接口（改造）

以下三个预览接口统一新增 LLM 摘要字段：

- **POST** `/api/documents/preview`
- **POST** `/api/documents/preview-upload`
- **POST** `/api/documents/preview-url`

**成功响应新增字段**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "preview_id": "tmp_abc123",
    "title": "牛顿运动定律",
    "llm_title": "牛顿三大运动定律详解",
    "llm_summary": "本文详细介绍了牛顿三大运动定律的内容、推导过程和实际应用。第一定律阐述了惯性原理，第二定律建立了力与加速度的定量关系，第三定律揭示了作用力与反作用力的对称性。",
    "tags": "物理,力学",
    "total_chars": 3500,
    "chunks": [
      {
        "index": 0,
        "content": "牛顿第一定律...",
        "char_count": 850
      }
    ],
    "chunk_count": 4
  }
}
```

| 字段 | 说明 |
|------|------|
| llm_title | LLM 自动生成的标题（生成失败时为空字符串） |
| llm_summary | LLM 自动生成的摘要（生成失败时为空字符串） |

> **说明**：
> - LLM 摘要调用超时设置为 10 秒，超时后降级为空字段
> - 如果用户在请求中已提供 title，llm_title 仅作为参考
> - 前端收到响应后，如果 title 为空则自动使用 llm_title 填充

---

### 3.9 确认文档入库（改造）

**POST** `/api/documents/confirm`

**改造说明**：新增 `summary` 字段。

**请求体**（新增字段）：
```json
{
  "preview_id": "tmp_abc123",
  "title": "牛顿三大运动定律详解",
  "summary": "本文详细介绍了牛顿三大运动定律...",
  "tags": "物理,力学",
  "scope": "class",
  "scope_ids": [1, 2]
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| summary | string | ❌ | 文档摘要（可由 LLM 生成后教师修改） |

---

## 4. 接口总览

### 4.1 新增接口（7 个）

| 编号 | 方法 | 路径 | 说明 | 角色 |
|------|------|------|------|------|
| API-45 | GET | `/api/personas/marketplace` | 分身广场列表 | 登录用户 |
| API-46 | PUT | `/api/personas/:id/visibility` | 设置分身公开/私有 | 登录用户（本人分身） |
| API-47 | GET | `/api/students/search` | 搜索已注册学生 | teacher |
| API-48 | POST | `/api/chat/teacher-reply` | 教师真人回复 | teacher |
| API-49 | GET | `/api/chat/takeover-status` | 查询接管状态 | 登录用户 |
| API-50 | POST | `/api/chat/end-takeover` | 教师退出接管 | teacher |
| API-51 | GET | `/api/conversations/student/:student_persona_id` | 教师查看学生对话记录 | teacher |

### 4.2 改造接口（8 个）

| 接口 | 改造内容 |
|------|----------|
| `POST /api/shares` | 新增 target_student_persona_id 定向邀请 |
| `POST /api/shares/:code/join` | 增加目标学生校验 |
| `GET /api/shares/:code/info` | 返回目标学生信息 |
| `POST /api/chat` | 增加接管状态检查 |
| `POST /api/chat/stream` | 增加接管状态检查 |
| `GET /api/conversations` | 返回 sender_type + reply_to_id |
| `GET /api/personas` | 返回 is_public 字段 |
| `POST /api/documents/preview*` + `confirm` | 返回 llm_title + llm_summary + summary |

---

## 5. 路由注册参考

```go
// router.go 新增路由（V2.0 迭代4）

// 分身广场
personas.GET("/marketplace", handler.HandleGetMarketplace)

// 分身公开设置
personas.PUT("/:id/visibility", handler.HandleSetVisibility)

// 搜索学生
authorized.GET("/students/search", auth.RoleRequired("teacher"), handler.HandleSearchStudents)

// 教师真人回复
authorized.POST("/chat/teacher-reply", auth.RoleRequired("teacher"), handler.HandleTeacherReply)

// 接管状态
authorized.GET("/chat/takeover-status", handler.HandleGetTakeoverStatus)

// 退出接管
authorized.POST("/chat/end-takeover", auth.RoleRequired("teacher"), handler.HandleEndTakeover)

// 教师查看学生对话记录
authorized.GET("/conversations/student/:student_persona_id", auth.RoleRequired("teacher"), handler.HandleGetStudentConversations)
```

---

**文档版本**: v1.0.0
**创建日期**: 2026-03-31
**最后更新**: 2026-03-31
