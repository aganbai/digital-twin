# V2.0 迭代8 API 接口规范

> **重要**：本文档以实际后端实现的路由为准（v1.1修正版），消除了原始规范与实现之间的路径歧义。

## 1. 通用约定

### 1.1 基础信息

| 项目 | 说明 |
|------|------|
| Base URL | `http://localhost:8080` |
| 协议 | HTTP（开发环境）/ HTTPS（生产环境） |
| 数据格式 | JSON（`Content-Type: application/json`） |
| 认证方式 | Bearer Token（`Authorization: Bearer <token>`） |

### 1.2 新增错误码

| 错误码 | HTTP Status | 说明 |
|--------|-------------|------|
| 40060 | 400 | URL解析失败，无法获取内容 |
| 40061 | 400 | 不支持的内容类型 |
| 40062 | 400 | 班级分享链接已失效 |
| 40063 | 400 | 学生申请已存在，请勿重复申请 |
| 40064 | 404 | 待审批的申请不存在 |
| 40065 | 400 | 置顶数量超限（最多置顶10个） |

---

## 2. 知识库相关接口

### 2.1 智能上传（统一接口）

**POST** `/api/knowledge/upload`

**鉴权**：需要（Bearer Token，仅教师角色）

**请求体**（JSON）：

```json
{
  "type": "text",
  "content": "这里是教学笔记内容...",
  "title": "可选的自定义标题",
  "persona_id": 1,
  "tags": ["素描", "基础"],
  "scope": "class",
  "scope_id": 1
}
```

**type 取值**：
- `url`：需提供 `url` 字段
- `text`：需提供 `content` 字段
- `file`：需提供 `file_urls` 字段（文件已上传到OSS的URL列表）

**字段说明**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | ✅ | url / text / file |
| url | string | ❌ | URL地址（type=url时必填） |
| content | string | ❌ | 文字内容（type=text时必填） |
| file_urls | []string | ❌ | 文件URL列表（type=file时必填） |
| title | string | ❌ | 自定义标题 |
| persona_id | int64 | ❌ | 分身ID（不传则使用默认分身） |
| tags | []string | ❌ | 标签列表 |
| scope | string | ❌ | 作用域：global / class / student |
| scope_id | int64 | ❌ | 作用域ID |

**成功响应** `200`：

```json
{
  "items": [
    {
      "id": 42,
      "title": "素描基础教程",
      "type": "file",
      "status": "success",
      "message": ""
    }
  ]
}
```

---

### 2.2 知识库列表（支持搜索）

**GET** `/api/knowledge`

**鉴权**：需要（Bearer Token，仅教师角色）

**查询参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | ❌ | 关键词搜索（标题） |
| item_type | string | ❌ | 类型筛选：url/text/file |
| scope | string | ❌ | 作用域筛选 |
| page | int | ❌ | 页码，默认1 |
| page_size | int | ❌ | 每页数量，默认20 |

**成功响应** `200`：

```json
{
  "items": [...],
  "total": 15,
  "page": 1,
  "page_size": 20,
  "total_pages": 1
}
```

---

### 2.3 更新知识库条目

**PUT** `/api/knowledge/:id`

**鉴权**：需要（Bearer Token，仅教师角色）

**请求体**：

```json
{
  "title": "新的标题",
  "content": "更新内容",
  "tags": ["新标签"],
  "scope": "class",
  "scope_id": 1
}
```

---

### 2.4 删除知识库条目

**DELETE** `/api/knowledge/:id`

**鉴权**：需要（Bearer Token，仅教师角色）

**成功响应** `200`：

```json
{
  "success": true
}
```

---

### 2.5 获取知识库详情

**GET** `/api/knowledge/:id`

**鉴权**：需要（Bearer Token，仅教师角色）

**成功响应** `200`：返回完整的 KnowledgeItem 对象

---

## 3. 班级相关接口

### 3.1 创建班级（扩展版）

**POST** `/api/classes/v8`

**鉴权**：需要（Bearer Token，仅教师角色）

**请求体**：

```json
{
  "name": "初一美术一班",
  "teacher_display_name": "曹老师",
  "subject": "美术",
  "age_group": "junior",
  "description": "专注培养学生的艺术鉴赏能力"
}
```

**字段说明**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | ✅ | 班级名称 |
| teacher_display_name | string | ❌ | 教师显示昵称 |
| subject | string | ❌ | 学科 |
| age_group | string | ❌ | 年龄范畴：preschool/primary_lower/primary_upper/junior/senior/adult |
| description | string | ❌ | 班级简介 |

**成功响应** `200`：

```json
{
  "id": 1,
  "name": "初一美术一班",
  "invite_code": "abc12345",
  "share_link": "/pages/share-join/index?code=abc12345",
  "qr_code_url": "/api/qrcode?text=abc12345"
}
```

---

### 3.2 获取班级分享信息

**GET** `/api/classes/:id/share-info`

**鉴权**：需要（Bearer Token，仅教师角色）

**成功响应** `200`：

```json
{
  "id": 1,
  "name": "初一美术一班",
  "description": "...",
  "teacher_display_name": "曹老师",
  "subject": "美术",
  "age_group": "junior",
  "share_link": "...",
  "invite_code": "abc12345",
  "qr_code_url": "...",
  "member_count": 18
}
```

---

### 3.3 学生申请加入班级

**POST** `/api/classes/join`

**鉴权**：需要（Bearer Token，仅学生角色）

**请求体**：

```json
{
  "invite_code": "abc12345",
  "request_message": "我想加入这个班级",
  "age": 12,
  "gender": "male",
  "family_info": "爷爷奶奶带大"
}
```

**成功响应** `200`：

```json
{
  "request_id": 1,
  "status": "pending",
  "message": "申请已提交，等待教师审批"
}
```

---

### 3.4 获取待审批学生列表

**GET** `/api/join-requests/pending`

**鉴权**：需要（Bearer Token，仅教师角色）

**成功响应** `200`：

```json
{
  "requests": [...],
  "total": 3
}
```

---

### 3.5 审批通过

**PUT** `/api/join-requests/:id/approve`

**鉴权**：需要（Bearer Token，仅教师角色）

**请求体**：

```json
{
  "teacher_evaluation": "该生素描基础较好，性格活泼开朗"
}
```

**成功响应** `200`：

```json
{
  "success": true,
  "message": "已批准加入申请"
}
```

---

### 3.6 审批拒绝

**PUT** `/api/join-requests/:id/reject`

**鉴权**：需要（Bearer Token，仅教师角色）

**请求体**：

```json
{
  "teacher_evaluation": "班级人数已满"
}
```

**成功响应** `200`：

```json
{
  "success": true,
  "message": "已拒绝加入申请"
}
```

---

### 3.7 获取班级成员列表（增强版）

**GET** `/api/classes/:id/members/v8`

**鉴权**：需要（Bearer Token，仅教师角色）

**成功响应** `200`：

```json
{
  "members": [
    {
      "id": 1,
      "student_persona_id": 201,
      "student_nickname": "小明",
      "student_avatar": "...",
      "age": 12,
      "gender": "male",
      "family_info": "...",
      "teacher_evaluation": "...",
      "joined_at": "2026-04-02T10:30:00Z",
      "approval_time": "2026-04-02T11:00:00Z"
    }
  ],
  "total": 18
}
```

---

## 4. 聊天相关接口

### 4.1 获取学生端老师聊天列表

**GET** `/api/chat-list/student`

**鉴权**：需要（Bearer Token，仅学生角色）

**成功响应** `200`：

```json
{
  "teachers": [
    {
      "teacher_persona_id": 1,
      "teacher_nickname": "曹老师",
      "teacher_avatar": "...",
      "teacher_school": "...",
      "subject": "美术",
      "last_message": "明天的素描课记得带铅笔",
      "last_message_time": "2026-04-02T10:30:00Z",
      "is_pinned": false
    }
  ],
  "total": 2
}
```

---

### 4.2 获取教师端聊天列表（按班级组织）

**GET** `/api/chat-list/teacher`

**鉴权**：需要（Bearer Token，仅教师角色）

**成功响应** `200`：

```json
{
  "classes": [
    {
      "class_id": 1,
      "class_name": "初一美术一班",
      "subject": "美术",
      "is_pinned": true,
      "students": [
        {
          "student_persona_id": 201,
          "student_nickname": "小明",
          "student_avatar": "...",
          "last_message": "老师好！",
          "last_message_time": "2026-04-02T10:00:00Z",
          "unread_count": 0,
          "is_pinned": false
        }
      ]
    }
  ],
  "total": 2
}
```

---

### 4.3 置顶聊天

**POST** `/api/chat-pins`

**鉴权**：需要（Bearer Token）

**请求体**：

```json
{
  "target_type": "class",
  "target_id": 1
}
```

**target_type 取值**：`teacher` / `student` / `class`

**成功响应** `200`：

```json
{
  "pin_id": 1,
  "success": true
}
```

---

### 4.4 取消置顶

**DELETE** `/api/chat-pins/:type/:id`

**鉴权**：需要（Bearer Token）

**URL参数**：
- `type`：target_type（teacher/student/class）
- `id`：target_id

**成功响应** `200`：

```json
{
  "success": true
}
```

---

### 4.5 获取置顶列表

**GET** `/api/chat-pins`

**鉴权**：需要（Bearer Token）

**成功响应** `200`：

```json
{
  "pins": [
    {
      "id": 1,
      "target_type": "class",
      "target_id": 1,
      "target_name": "初一美术一班",
      "avatar": "",
      "pinned_at": "2026-04-02 10:30:00"
    }
  ],
  "total": 1
}
```

---

### 4.6 开启新会话 ⚠️ 待实现

**POST** `/api/chat/new-session`

**鉴权**：需要（Bearer Token）

**请求体**：

```json
{
  "teacher_persona_id": 1,
  "initial_message": "回顾上次内容"
}
```

**成功响应** `200`：

```json
{
  "session_id": "sess_abc123",
  "message": "好的，让我帮你回顾一下我们上次学习的内容..."
}
```

---

### 4.7 获取快捷指令 ⚠️ 待实现

**GET** `/api/chat/quick-actions`

**鉴权**：需要（Bearer Token）

**查询参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| teacher_persona_id | int64 | ✅ | 教师分身ID |

**成功响应** `200`：

```json
[
  {
    "id": "review",
    "label": "📚 回顾上次内容",
    "action": "回顾上次学习的内容"
  },
  {
    "id": "summarize",
    "label": "📝 总结已学知识",
    "action": "帮我总结一下已学的知识点"
  },
  {
    "id": "practice",
    "label": "✏️ 开始练习",
    "action": "我想开始练习"
  },
  {
    "id": "question",
    "label": "❓ 提个问题",
    "action": "我有一个问题"
  }
]
```

---

## 5. 发现页接口

### 5.1 获取推荐班级/老师

**GET** `/api/discover`

**鉴权**：需要（Bearer Token）

**查询参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | ❌ | class / teacher / all（默认） |
| page | int | ❌ | 页码，默认1 |
| page_size | int | ❌ | 每页数量，默认10 |

**成功响应** `200`：

```json
{
  "items": [
    {
      "id": 1,
      "type": "class",
      "title": "曹老师的素描工作室",
      "description": "...",
      "avatar": "...",
      "tags": ["美术", "junior"],
      "member_count": 128,
      "teacher_name": "曹老师",
      "subject": "美术",
      "age_group": "junior"
    }
  ],
  "total": 10,
  "page": 1,
  "page_size": 10,
  "total_pages": 1
}
```

---

### 5.2 获取发现页详情

**GET** `/api/discover/detail`

**鉴权**：需要（Bearer Token）

**查询参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | ✅ | class / teacher |
| id | int64 | ✅ | 班级ID或教师分身ID |

**成功响应** `200`：

```json
{
  "id": 1,
  "type": "class",
  "title": "初一美术一班",
  "description": "...",
  "avatar": "...",
  "teacher_name": "曹老师",
  "subject": "美术",
  "age_group": "junior",
  "member_count": 18,
  "tags": ["美术", "junior"],
  "share_link": "...",
  "invite_code": "abc12345",
  "is_joined": false,
  "application_status": "pending"
}
```

---

### 5.3 搜索班级/老师 ⚠️ 待实现

**GET** `/api/discover/search`

**鉴权**：需要（Bearer Token）

**查询参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | ✅ | 搜索关键词 |
| type | string | ❌ | class / teacher / all |
| page | int | ❌ | 页码 |
| page_size | int | ❌ | 每页数量 |

**成功响应** `200`：结构同 `/api/discover`

---

## 6. 接口变更汇总

### 6.1 新增接口（20个）

| 编号 | 接口 | 方法 | 说明 | 状态 |
|------|------|------|------|------|
| API-80 | `/api/knowledge/upload` | POST | 智能上传（URL/文字/文件） | ✅ 已实现 |
| API-81 | `/api/knowledge` | GET | 知识库列表（支持搜索） | ✅ 已实现 |
| API-82 | `/api/knowledge/:id` | PUT | 更新知识库条目 | ✅ 已实现 |
| API-83 | `/api/knowledge/:id` | DELETE | 删除知识库条目 | ✅ 已实现 |
| API-84 | `/api/knowledge/:id` | GET | 获取知识库详情 | ✅ 已实现 |
| API-85 | `/api/classes/v8` | POST | 创建班级（扩展版） | ✅ 已实现 |
| API-86 | `/api/classes/:id/share-info` | GET | 获取班级分享信息 | ✅ 已实现 |
| API-87 | `/api/classes/:id/members/v8` | GET | 获取班级成员（增强版） | ✅ 已实现 |
| API-88 | `/api/classes/join` | POST | 学生申请加入 | ✅ 已实现 |
| API-89 | `/api/join-requests/pending` | GET | 获取待审批列表 | ✅ 已实现 |
| API-90 | `/api/join-requests/:id/approve` | PUT | 审批通过 | ✅ 已实现 |
| API-91 | `/api/join-requests/:id/reject` | PUT | 审批拒绝 | ✅ 已实现 |
| API-92 | `/api/chat-list/student` | GET | 学生端老师聊天列表 | ✅ 已实现 |
| API-93 | `/api/chat-list/teacher` | GET | 教师端聊天列表 | ✅ 已实现 |
| API-94 | `/api/chat-pins` | POST | 置顶聊天 | ✅ 已实现 |
| API-95 | `/api/chat-pins/:type/:id` | DELETE | 取消置顶 | ✅ 已实现 |
| API-96 | `/api/chat-pins` | GET | 获取置顶列表 | ✅ 已实现 |
| API-97 | `/api/discover` | GET | 发现页推荐 | ✅ 已实现 |
| API-98 | `/api/discover/detail` | GET | 发现页详情 | ✅ 已实现 |
| API-99 | `/api/chat/new-session` | POST | 开启新会话 | ⚠️ 待实现 |
| API-100 | `/api/chat/quick-actions` | GET | 获取快捷指令 | ⚠️ 待实现 |
| API-101 | `/api/discover/search` | GET | 搜索班级/老师 | ⚠️ 待实现 |

### 6.2 数据库变更

| 变更 | 说明 |
|------|------|
| **新增表** `knowledge_items` | 整合知识库条目 |
| **新增表** `class_join_requests` | 班级加入申请 |
| **新增表** `chat_pins` | 置顶记录 |
| **修改表** `classes` | 新增：teacher_display_name, subject, age_group, share_link, invite_code, qr_code_url |
| **修改表** `class_members` | 新增：approval_status, teacher_evaluation, age, gender, family_info, request_time, approval_time |

---

**文档版本**: v1.1（修正版，以实际实现路由为准）
**创建日期**: 2026-04-02
**状态**: 已创建
**关联需求**: docs/iterations/v2.0/iteration8_requirements.md
