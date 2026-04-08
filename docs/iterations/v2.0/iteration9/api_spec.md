# V2.0 迭代9 API 接口规范

## 版本信息
- **文档版本**: v1.0
- **创建日期**: 2026-04-03
- **关联需求**: `iteration9_requirements.md`

---

## API 列表

| 编号 | 方法 | 路径 | 说明 | 优先级 |
|------|------|------|------|--------|
| API-102 | POST | /api/chat/stream | SSE 流式对话（增强 thinking_step 事件） | P0 |
| API-103 | POST | /api/conversations/sessions | 创建新会话 | P0 |
| API-104 | GET | /api/conversations/sessions | 获取会话列表（增强 teacher_persona_id 过滤） | P0 |
| API-105 | POST | /api/conversations/sessions/:session_id/title | 异步生成会话标题 | P0 |
| API-106 | GET | /api/classes/:id | 获取班级详情（学生查看） | P1 |
| API-107 | GET | /api/students/:id/profile | 获取学生详情（教师查看） | P1 |
| API-108 | PUT | /api/students/:id/evaluation | 更新学生评语 | P1 |
| API-109 | POST | /api/courses | 发布课程信息 | P1 |
| API-110 | GET | /api/courses | 获取课程列表 | P1 |
| API-111 | PUT | /api/courses/:id | 更新课程信息 | P1 |
| API-112 | DELETE | /api/courses/:id | 删除课程信息 | P1 |
| API-113 | POST | /api/courses/:id/push | 推送课程通知 | P1 |
| API-114 | POST | /api/wx/subscribe | 订阅微信消息 | P1 |
| API-115 | GET | /api/wx/subscribe/status | 获取订阅状态 | P1 |

---

## API-102: SSE 流式对话（增强）

### 接口说明
在现有 SSE 流式对话基础上，新增 `thinking_step` 类型事件，用于实时展示系统思考过程。

### 请求
复用现有 `POST /api/chat/stream` 接口，无需修改请求参数。

### 响应（新增事件类型）

#### thinking_step 事件格式

```json
{
  "type": "thinking_step",
  "step": "rag_search",
  "status": "start",
  "message": "🔍 正在检索知识库...",
  "timestamp": 1712123456789
}
```

#### 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| type | string | 固定值 `"thinking_step"` |
| step | string | 步骤标识：`rag_search` / `memory_recall` / `tool_call` / `llm_thinking` |
| status | string | 状态：`start` / `done` |
| message | string | 展示给用户的提示文案 |
| detail | string | 完成时的详细信息（可选，仅 status=done 时） |
| duration_ms | number | 步骤耗时毫秒（可选，仅 status=done 时） |
| timestamp | number | 时间戳（毫秒） |

### 向后兼容
旧版客户端忽略未知事件类型即可，不影响现有功能。

---

## API-103: 创建新会话

### 接口说明
创建新的对话会话，同时异步为上一个活跃会话生成标题。

### 请求

```http
POST /api/conversations/sessions
Authorization: Bearer <token>
Content-Type: application/json

{
  "teacher_persona_id": 123,
  "student_persona_id": 456
}
```

### 响应

```json
{
  "code": 0,
  "data": {
    "session_id": "sess_abc123",
    "created_at": "2026-04-03T12:00:00Z"
  }
}
```

### 业务逻辑
1. 生成新的 `session_id`（UUID v4 格式）
2. 查询上一个活跃会话（按 `updated_at` 排序）
3. 异步触发上一个会话的标题生成任务
4. 返回新会话 ID

---

## API-104: 获取会话列表（增强）

### 接口说明
获取指定教师分身下的会话列表，增强支持按 `teacher_persona_id` 过滤。

### 请求

```http
GET /api/conversations/sessions?teacher_persona_id=123&page=1&page_size=20
Authorization: Bearer <token>
```

### 响应

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "session_id": "sess_abc123",
        "title": null,
        "last_message": "老师，色彩搭配有什么技巧？",
        "last_message_role": "user",
        "message_count": 12,
        "is_active": true,
        "updated_at": "2026-03-28T15:30:00Z"
      },
      {
        "session_id": "sess_def456",
        "title": "素描技法问答",
        "last_message": "谢谢老师的指导！",
        "last_message_role": "user",
        "message_count": 8,
        "is_active": false,
        "updated_at": "2026-03-25T10:00:00Z"
      }
    ],
    "total": 3
  }
}
```

### 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| session_id | string | 会话唯一标识 |
| title | string? | 会话标题（最新会话为 null） |
| last_message | string | 最后一条消息内容 |
| last_message_role | string | 最后一条消息角色 |
| message_count | number | 消息总数 |
| is_active | boolean | 是否为当前活跃会话 |
| updated_at | string | 最后更新时间 |

---

## API-105: 异步生成会话标题

### 接口说明
异步为指定会话生成标题，由后端定时任务或手动触发。

### 请求

```http
POST /api/conversations/sessions/:session_id/title
Authorization: Bearer <token>
```

### 响应

```json
{
  "code": 0,
  "message": "标题生成任务已提交"
}
```

### 业务逻辑
1. 检查会话是否存在且属于当前用户
2. 检查是否已有标题（幂等）
3. 异步调用 LLM 生成标题
4. 标题存入 `session_titles` 表

### LLM Prompt

```
请为以下对话生成一个简短的标题（10-20字），概括对话的主要话题：

{最近10条消息内容}

要求：
- 标题简洁明了
- 体现对话的核心话题
- 不超过20个字
```

---

## API-106: 获取班级详情

### 接口说明
学生点击老师头像后查看班级信息。

### 请求

```http
GET /api/classes/:id
Authorization: Bearer <token>
```

### 响应

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "name": "初一美术一班",
    "subject": "美术",
    "description": "面向初一学生的美术基础班",
    "teacher_name": "曹老师",
    "member_count": 25,
    "created_at": "2026-03-01T00:00:00Z"
  }
}
```

### 权限
- 学生必须是班级成员才能查看
- `profile_snapshot` 字段不返回

---

## API-107: 获取学生详情

### 接口说明
教师点击学生头像后查看学生信息。

### 请求

```http
GET /api/students/:id/profile
Authorization: Bearer <token>
```

### 响应

```json
{
  "code": 0,
  "data": {
    "id": 123,
    "nickname": "张三",
    "age": 13,
    "gender": "男",
    "family_info": "父母均为教师",
    "teacher_evaluation": "学习认真，进度良好",
    "class_name": "初一美术一班"
  }
}
```

### 权限
- 教师必须与该学生有师生关系才能查看
- `profile_snapshot` 字段不返回

---

## API-108: 更新学生评语

### 接口说明
教师更新对学生的评语，覆盖更新。

### 请求

```http
PUT /api/students/:id/evaluation
Authorization: Bearer <token>
Content-Type: application/json

{
  "evaluation": "该学生进步明显，色彩感知能力强"
}
```

### 响应

```json
{
  "code": 0,
  "message": "评语更新成功"
}
```

### 权限
- 教师必须与该学生有师生关系才能修改

---

## API-109: 发布课程信息

### 接口说明
教师发布课程信息，存入知识库。

### 请求

```http
POST /api/courses
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "色彩基础理论",
  "content": "本次课程介绍色彩的三要素：色相、明度、纯度...",
  "class_id": 1,
  "push_to_students": true
}
```

### 响应

```json
{
  "code": 0,
  "data": {
    "id": 123,
    "title": "色彩基础理论",
    "created_at": "2026-04-03T12:00:00Z"
  }
}
```

### 业务逻辑
1. 创建 `knowledge_items` 记录，`item_type = 'course'`
2. 如果 `push_to_students = true`，异步发送推送通知

---

## API-110: 获取课程列表

### 接口说明
教师查看已发布的课程列表。

### 请求

```http
GET /api/courses?class_id=1&page=1&page_size=20
Authorization: Bearer <token>
```

### 响应

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 123,
        "title": "色彩基础理论",
        "content": "本次课程介绍...",
        "class_id": 1,
        "class_name": "初一美术一班",
        "created_at": "2026-04-03T12:00:00Z",
        "pushed": true
      }
    ],
    "total": 5
  }
}
```

---

## API-111: 更新课程信息

### 接口说明
教师修改已发布的课程信息。

### 请求

```http
PUT /api/courses/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "色彩基础理论（修订版）",
  "content": "更新后的内容..."
}
```

### 响应

```json
{
  "code": 0,
  "message": "课程更新成功"
}
```

---

## API-112: 删除课程信息

### 接口说明
教师删除已发布的课程信息。

### 请求

```http
DELETE /api/courses/:id
Authorization: Bearer <token>
```

### 响应

```json
{
  "code": 0,
  "message": "课程删除成功"
}
```

---

## API-113: 推送课程通知

### 接口说明
向班级学生推送课程更新通知。

### 请求

```http
POST /api/courses/:id/push
Authorization: Bearer <token>
Content-Type: application/json

{
  "push_type": "in_app"
}
```

### 响应

```json
{
  "code": 0,
  "data": {
    "pushed_count": 25,
    "failed_count": 0
  }
}
```

### 推送类型
- `in_app`: 应用内推送
- `wechat`: 微信订阅消息推送

---

## API-114: 订阅微信消息

### 接口说明
用户主动订阅微信消息推送。

### 请求

```http
POST /api/wx/subscribe
Authorization: Bearer <token>
Content-Type: application/json

{
  "template_id": "course_update",
  "subscribe": true
}
```

### 响应

```json
{
  "code": 0,
  "message": "订阅成功"
}
```

---

## API-115: 获取订阅状态

### 接口说明
查询用户的微信消息订阅状态。

### 请求

```http
GET /api/wx/subscribe/status?template_id=course_update
Authorization: Bearer <token>
```

### 响应

```json
{
  "code": 0,
  "data": {
    "template_id": "course_update",
    "is_subscribed": true,
    "last_subscribe_time": "2026-04-03T10:00:00Z"
  }
}
```

---

## 数据库变更

### 新增表

#### course_notifications

```sql
CREATE TABLE IF NOT EXISTS course_notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    course_item_id INTEGER NOT NULL,
    class_id INTEGER NOT NULL,
    teacher_id INTEGER NOT NULL,
    persona_id INTEGER NOT NULL,
    push_type TEXT NOT NULL,  -- 'in_app' / 'wechat'
    status TEXT NOT NULL,      -- 'pending' / 'sent' / 'failed'
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### wx_subscriptions

```sql
CREATE TABLE IF NOT EXISTS wx_subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    template_id TEXT NOT NULL,
    is_subscribed INTEGER DEFAULT 0,
    last_subscribe_time DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, template_id)
);
```

#### session_titles

```sql
CREATE TABLE IF NOT EXISTS session_titles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL UNIQUE,
    student_persona_id INTEGER NOT NULL,
    teacher_persona_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

---

## 画像隐私保护（R7）

### 需要过滤 profile_snapshot 的接口

以下接口在返回用户信息时，必须过滤 `profile_snapshot` 字段：

| 接口 | 说明 |
|------|------|
| GET /api/user/profile | 获取当前用户信息 |
| GET /api/teachers | 获取教师列表 |
| GET /api/students | 获取学生列表 |
| GET /api/classes/:id/members | 获取班级成员 |
| GET /api/personas | 获取分身列表 |
| 所有返回 User 对象的接口 | 统一过滤 |

### 实现方式

在 Repository 层统一处理：

```go
// 返回用户信息时，清空 profile_snapshot
func (r *UserRepository) GetByID(id int64) (*User, error) {
    user, err := r.getByID(id)
    if err != nil {
        return nil, err
    }
    user.ProfileSnapshot = "" // 隐私保护
    return user, nil
}
```

---

## 前端接口对接清单

| 功能 | 接口 | 方法 |
|------|------|------|
| 思考过程展示 | /api/chat/stream | SSE 事件处理 |
| 语音输入 | 微信 API | wx.startRecord |
| +号面板 | 微信 API | wx.chooseMedia |
| 新会话指令 | /api/conversations/sessions | POST |
| 会话列表 | /api/conversations/sessions | GET |
| 班级信息 | /api/classes/:id | GET |
| 学生信息 | /api/students/:id/profile | GET |
| 评语修改 | /api/students/:id/evaluation | PUT |
| 课程发布 | /api/courses | POST |
| 课程列表 | /api/courses | GET |
| 课程编辑 | /api/courses/:id | PUT |
| 课程删除 | /api/courses/:id | DELETE |
| 课程推送 | /api/courses/:id/push | POST |
| 微信订阅 | /api/wx/subscribe | POST |
| 订阅状态 | /api/wx/subscribe/status | GET |

---

**文档版本**: v1.0
**创建日期**: 2026-04-03
**关联需求**: 迭代9 对话体验增强 + 教学管理增强
