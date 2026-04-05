# V2.0 迭代10 API 接口规范

## 版本信息
- **文档版本**: v1.0
- **创建日期**: 2026-04-04
- **关联需求**: `iteration10_requirements.md`

---

## API 列表

| 编号 | 方法 | 路径 | 说明 | 优先级 |
|------|------|------|------|--------|
| API-201 | GET | /api/auth/wx-h5-login-url | 获取微信H5授权跳转URL | P0 |
| API-202 | POST | /api/auth/wx-h5-callback | 微信H5授权回调 | P0 |
| API-203 | GET | /api/admin/dashboard/overview | 系统总览 | P0 |
| API-204 | GET | /api/admin/dashboard/user-stats | 用户统计 | P0 |
| API-205 | GET | /api/admin/dashboard/chat-stats | 对话统计 | P0 |
| API-206 | GET | /api/admin/dashboard/knowledge-stats | 知识库统计 | P0 |
| API-207 | GET | /api/admin/dashboard/active-users | 活跃用户排行 | P0 |
| API-208 | GET | /api/admin/users | 用户管理列表 | P0 |
| API-209 | PUT | /api/admin/users/:id/role | 修改用户角色 | P0 |
| API-210 | PUT | /api/admin/users/:id/status | 启用/禁用用户 | P0 |
| API-211 | GET | /api/admin/feedbacks | 反馈管理列表 | P0 |
| API-212 | GET | /api/admin/logs | 查询操作日志 | P0 |
| API-213 | GET | /api/admin/logs/stats | 日志统计 | P0 |
| API-214 | GET | /api/admin/logs/export | 导出日志(CSV) | P0 |
| API-215 | GET | /api/platform/config | 获取平台配置 | P0 |
| API-216 | POST | /api/upload/h5 | H5端文件上传 | P1 |

---

## API-201: 获取微信H5授权跳转URL

### 接口说明
H5前端调用此接口获取微信网页授权的跳转URL，前端拿到URL后跳转到微信授权页面。

### 请求

```http
GET /api/auth/wx-h5-login-url?redirect_uri=https://your-domain.com/h5/auth/callback
```

### 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| redirect_uri | string | 是 | 授权后的回调地址 |

### 响应

```json
{
  "code": 0,
  "data": {
    "auth_url": "https://open.weixin.qq.com/connect/oauth2/authorize?appid=wx_app_id&redirect_uri=https%3A%2F%2Fyour-domain.com%2Fh5%2Fauth%2Fcallback&response_type=code&scope=snsapi_userinfo&state=random_state_string#wechat_redirect",
    "state": "random_state_string"
  }
}
```

### 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| auth_url | string | 微信授权跳转URL，前端直接 window.location.href 跳转 |
| state | string | 防CSRF的随机字符串，回调时需校验 |

### Mock模式

开发环境下（`WX_H5_MOCK_ENABLED=true`），返回一个模拟的授权URL，跳转后直接回调到 redirect_uri 并带上 mock code。

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40004 | 缺少 redirect_uri 参数 |
| 50001 | 微信配置不可用 |

---

## API-202: 微信H5授权回调

### 接口说明
微信授权后，前端将回调中的 code 和 state 发送到后端，后端用 code 换取用户信息并返回 JWT token。

### 请求

```http
POST /api/auth/wx-h5-callback
Content-Type: application/json

{
  "code": "微信回调的authorization_code",
  "state": "防CSRF的state字符串"
}
```

### 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | 微信授权回调的 code |
| state | string | 是 | 防CSRF的 state，需与请求时一致 |

### 响应

```json
{
  "code": 0,
  "data": {
    "user_id": 1,
    "token": "jwt_token_string",
    "role": "teacher",
    "nickname": "曹老师",
    "is_new_user": false,
    "expires_at": "2026-04-05T00:00:00Z",
    "personas": [
      {
        "id": 1,
        "role": "teacher",
        "nickname": "曹老师",
        "is_active": true
      }
    ],
    "current_persona": {
      "id": 1,
      "role": "teacher",
      "nickname": "曹老师",
      "is_active": true
    }
  }
}
```

### 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | number | 用户ID |
| token | string | JWT token |
| role | string | 用户角色（teacher/student/admin），新用户为空 |
| nickname | string | 用户昵称 |
| is_new_user | boolean | 是否新用户（需补全角色信息） |
| expires_at | string | token 过期时间 |
| personas | array | 分身列表（新用户为空数组） |
| current_persona | object | 当前分身（新用户为 null） |

### 业务逻辑

1. 校验 state 防CSRF
2. 用 code 调用微信接口换取 access_token + openid
3. 用 access_token 调用微信接口获取用户信息（昵称、头像）
4. 通过 openid 查找用户：
   - 已存在：更新昵称/头像，生成 JWT token
   - 不存在：创建新用户（role 为空），标记 is_new_user=true
5. 检查用户状态：如果 status=disabled，返回 40003 错误
6. 返回 token 和用户信息

### Mock模式

开发环境下，code 以 `mock_` 开头时，跳过微信接口调用，直接使用 mock 数据：
- `mock_teacher_001` → 教师用户
- `mock_student_001` → 学生用户
- `mock_admin_001` → 管理员用户
- `mock_new_001` → 新用户

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40001 | state 校验失败 |
| 40003 | 账号已被禁用，请联系管理员 |
| 40004 | 缺少 code 或 state 参数 |
| 50001 | 微信接口调用失败 |

---

## API-203: 系统总览

### 接口说明
获取系统核心业务指标和趋势数据。

### 请求

```http
GET /api/admin/dashboard/overview?period=7d
Authorization: Bearer <admin_token>
```

### 参数

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| period | string | 否 | 7d | 趋势数据时间范围：7d / 30d / 90d |

### 响应

```json
{
  "code": 0,
  "data": {
    "total_users": 1200,
    "total_teachers": 50,
    "total_students": 1150,
    "total_conversations": 35000,
    "total_messages": 280000,
    "total_knowledge_items": 500,
    "total_classes": 80,
    "active_users_today": 320,
    "new_users_today": 15,
    "messages_today": 8500,
    "period_trend": {
      "dates": ["2026-03-28", "2026-03-29", "2026-03-30"],
      "new_users": [12, 15, 8],
      "active_users": [280, 310, 295],
      "messages": [7200, 8500, 6800]
    }
  }
}
```

### 权限
- 仅 admin 角色可访问

---

## API-204: 用户统计

### 接口说明
获取用户维度的统计数据。

### 请求

```http
GET /api/admin/dashboard/user-stats?period=30d
Authorization: Bearer <admin_token>
```

### 响应

```json
{
  "code": 0,
  "data": {
    "role_distribution": [
      { "role": "teacher", "count": 50 },
      { "role": "student", "count": 1150 },
      { "role": "admin", "count": 2 }
    ],
    "registration_trend": {
      "dates": ["2026-03-01", "2026-03-02"],
      "counts": [5, 8]
    },
    "activity_distribution": [
      { "level": "high", "count": 200, "description": "近7天活跃≥5天" },
      { "level": "medium", "count": 350, "description": "近7天活跃2-4天" },
      { "level": "low", "count": 300, "description": "近7天活跃1天" },
      { "level": "silent", "count": 350, "description": "近7天无活跃" }
    ]
  }
}
```

### 权限
- 仅 admin 角色可访问

---

## API-205: 对话统计

### 接口说明
获取对话维度的统计数据。

### 请求

```http
GET /api/admin/dashboard/chat-stats?period=7d
Authorization: Bearer <admin_token>
```

### 响应

```json
{
  "code": 0,
  "data": {
    "daily_messages": {
      "dates": ["2026-03-28", "2026-03-29"],
      "counts": [7200, 8500]
    },
    "avg_turns_per_session": 6.5,
    "hourly_distribution": {
      "hours": [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23],
      "counts": [120, 50, 30, 20, 15, 25, 80, 350, 600, 800, 750, 680, 500, 550, 620, 700, 780, 850, 900, 820, 650, 450, 300, 180]
    }
  }
}
```

### 权限
- 仅 admin 角色可访问

---

## API-206: 知识库统计

### 接口说明
获取知识库维度的统计数据。

### 请求

```http
GET /api/admin/dashboard/knowledge-stats?period=30d
Authorization: Bearer <admin_token>
```

### 响应

```json
{
  "code": 0,
  "data": {
    "type_distribution": [
      { "type": "text", "count": 200 },
      { "type": "url", "count": 100 },
      { "type": "file", "count": 150 },
      { "type": "course", "count": 50 }
    ],
    "growth_trend": {
      "dates": ["2026-03-01", "2026-03-02"],
      "counts": [3, 5]
    },
    "total_items": 500
  }
}
```

### 权限
- 仅 admin 角色可访问

---

## API-207: 活跃用户排行

### 接口说明
获取最近N天内最活跃的用户排行。

### 请求

```http
GET /api/admin/dashboard/active-users?days=7&role=teacher&limit=20
Authorization: Bearer <admin_token>
```

### 参数

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| days | number | 否 | 7 | 统计天数 |
| role | string | 否 | 空（全部） | 角色筛选：teacher / student |
| limit | number | 否 | 20 | 返回条数，最大50 |

### 响应

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "user_id": 1,
        "nickname": "曹老师",
        "role": "teacher",
        "message_count": 520,
        "session_count": 45,
        "last_active_at": "2026-04-03T18:30:00Z"
      }
    ]
  }
}
```

### 权限
- 仅 admin 角色可访问

---

## API-208: 用户管理列表

### 接口说明
分页获取用户列表，支持搜索和筛选。

### 请求

```http
GET /api/admin/users?page=1&page_size=20&keyword=曹&role=teacher&status=active
Authorization: Bearer <admin_token>
```

### 参数

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | number | 否 | 1 | 页码 |
| page_size | number | 否 | 20 | 每页条数，最大100 |
| keyword | string | 否 | 空 | 搜索关键词（匹配昵称/用户名） |
| role | string | 否 | 空（全部） | 角色筛选 |
| status | string | 否 | 空（全部） | 状态筛选：active / disabled |

### 响应

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 1,
        "username": "user_wx_abc123",
        "nickname": "曹老师",
        "role": "teacher",
        "status": "active",
        "school": "北京市第一中学",
        "persona_count": 2,
        "created_at": "2026-03-01T10:00:00Z",
        "last_active_at": "2026-04-03T18:30:00Z"
      }
    ],
    "total": 1200,
    "page": 1,
    "page_size": 20
  }
}
```

### 权限
- 仅 admin 角色可访问
- 不返回 profile_snapshot 字段（隐私保护）

---

## API-209: 修改用户角色

### 接口说明
管理员修改指定用户的角色。

### 请求

```http
PUT /api/admin/users/:id/role
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "role": "teacher"
}
```

### 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | number | 是 | 用户ID（URL参数） |
| role | string | 是 | 新角色：teacher / student / admin |

### 响应

```json
{
  "code": 0,
  "message": "角色修改成功"
}
```

### 业务逻辑
1. 校验目标角色合法性
2. 不允许修改自己的角色
3. 更新用户角色
4. 记录操作日志（action: admin.update_role）

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40004 | 无效的角色值 |
| 40005 | 用户不存在 |
| 40020 | 不允许修改自己的角色 |

---

## API-210: 启用/禁用用户

### 接口说明
管理员启用或禁用指定用户。禁用后该用户的token立即失效。

### 请求

```http
PUT /api/admin/users/:id/status
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "status": "disabled"
}
```

### 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | number | 是 | 用户ID（URL参数） |
| status | string | 是 | 新状态：active / disabled |

### 响应

```json
{
  "code": 0,
  "message": "用户已禁用"
}
```

### 业务逻辑
1. 校验目标状态合法性
2. 不允许禁用自己
3. 更新用户状态
4. 如果是禁用操作：将用户加入黑名单缓存（内存），JWT中间件检查时立即拒绝
5. 记录操作日志（action: admin.toggle_user）

### 禁用后的行为
- 被禁用用户的所有请求返回 403
- 错误信息：`{"code": 40003, "message": "您的账号已被禁用，请联系管理员"}`
- 被禁用用户无法重新登录（登录接口也检查状态）

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40004 | 无效的状态值 |
| 40005 | 用户不存在 |
| 40020 | 不允许禁用自己 |

---

## API-211: 反馈管理列表

### 接口说明
管理员查看所有用户反馈。

### 请求

```http
GET /api/admin/feedbacks?page=1&page_size=20&status=pending
Authorization: Bearer <admin_token>
```

### 参数

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | number | 否 | 1 | 页码 |
| page_size | number | 否 | 20 | 每页条数 |
| status | string | 否 | 空（全部） | 状态筛选：pending / processing / resolved |

### 响应

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 1,
        "user_id": 123,
        "user_nickname": "小明",
        "user_role": "student",
        "feedback_type": "bug",
        "content": "对话页面偶尔卡顿",
        "status": "pending",
        "context_info": "{}",
        "created_at": "2026-04-03T10:00:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 20
  }
}
```

### 权限
- 仅 admin 角色可访问

---

## API-212: 查询操作日志

### 接口说明
分页查询操作日志，支持多条件筛选。

### 请求

```http
GET /api/admin/logs?page=1&page_size=20&user_id=123&action=chat.send_message&start_time=2026-04-01&end_time=2026-04-03&platform=h5
Authorization: Bearer <admin_token>
```

### 参数

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | number | 否 | 1 | 页码 |
| page_size | number | 否 | 20 | 每页条数，最大100 |
| user_id | number | 否 | 空 | 用户ID筛选 |
| action | string | 否 | 空 | 操作类型筛选 |
| resource | string | 否 | 空 | 资源类型筛选 |
| platform | string | 否 | 空 | 平台筛选：miniapp / h5 |
| start_time | string | 否 | 空 | 开始时间（YYYY-MM-DD） |
| end_time | string | 否 | 空 | 结束时间（YYYY-MM-DD） |

### 响应

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 10001,
        "user_id": 123,
        "user_nickname": "小明",
        "user_role": "student",
        "persona_id": 456,
        "action": "chat.send_message",
        "resource": "conversation",
        "resource_id": "sess_abc123",
        "detail": "{\"message_length\": 50, \"teacher_persona_id\": 789}",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0...",
        "platform": "h5",
        "status_code": 200,
        "duration_ms": 120,
        "created_at": "2026-04-03T10:30:00Z"
      }
    ],
    "total": 5000,
    "page": 1,
    "page_size": 20
  }
}
```

### 权限
- 仅 admin 角色可访问

### 说明
- 日志从独立的日志数据库中查询
- user_nickname 通过关联业务数据库获取

---

## API-213: 日志统计

### 接口说明
获取操作日志的统计数据。

### 请求

```http
GET /api/admin/logs/stats?period=7d&group_by=action
Authorization: Bearer <admin_token>
```

### 参数

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| period | string | 否 | 7d | 统计时间范围：7d / 30d / 90d |
| group_by | string | 否 | action | 分组维度：action / platform / user |

### 响应

```json
{
  "code": 0,
  "data": {
    "total_operations": 50000,
    "unique_users": 800,
    "by_action": [
      { "action": "chat.send_message", "count": 28000 },
      { "action": "user.login", "count": 5000 },
      { "action": "knowledge.upload", "count": 200 }
    ],
    "by_platform": [
      { "platform": "miniapp", "count": 35000 },
      { "platform": "h5", "count": 15000 }
    ],
    "hourly_heatmap": {
      "hours": [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23],
      "counts": [120, 50, 30, 20, 15, 25, 80, 350, 600, 800, 750, 680, 500, 550, 620, 700, 780, 850, 900, 820, 650, 450, 300, 180]
    }
  }
}
```

### 权限
- 仅 admin 角色可访问

---

## API-214: 导出日志(CSV)

### 接口说明
导出操作日志为CSV格式文件。应用当前的筛选条件。

### 请求

```http
GET /api/admin/logs/export?user_id=123&action=chat.send_message&start_time=2026-04-01&end_time=2026-04-03
Authorization: Bearer <admin_token>
```

### 参数
与 API-212 相同的筛选参数（不含分页参数）。

### 响应

- Content-Type: `text/csv; charset=utf-8`
- Content-Disposition: `attachment; filename=operation_logs_20260403.csv`

CSV 字段：
```
时间,用户ID,用户昵称,角色,操作类型,资源类型,资源ID,详情,IP,平台,状态码,耗时(ms)
2026-04-03 10:30:00,123,小明,student,chat.send_message,conversation,sess_abc123,"{""message_length"":50}",192.168.1.100,h5,200,120
```

### 权限
- 仅 admin 角色可访问

### 限制
- 单次导出最多 10000 条记录
- 超过限制时返回错误提示

---

## API-215: 获取平台配置

### 接口说明
H5前端获取平台差异化配置。

### 请求

```http
GET /api/platform/config?platform=h5
```

### 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| platform | string | 是 | 平台标识：h5 / miniapp |

### 响应

```json
{
  "code": 0,
  "data": {
    "platform": "h5",
    "features": {
      "voice_input": false,
      "wx_subscribe": false,
      "file_upload": true,
      "share_to_wechat": false,
      "camera": false
    },
    "upload": {
      "max_file_size_mb": 20,
      "allowed_types": ["image/*", "application/pdf", ".doc", ".docx", ".txt", ".md"]
    },
    "auth": {
      "login_type": "wx_h5_oauth",
      "need_complete_profile": true
    }
  }
}
```

### 说明
- 此接口无需鉴权
- 前端根据返回的 features 控制功能的显示/隐藏

---

## API-216: H5端文件上传

### 接口说明
H5端文件上传，使用标准的 multipart/form-data 格式（不走微信临时文件路径）。

### 请求

```http
POST /api/upload/h5
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: (binary)
```

### 响应

```json
{
  "code": 0,
  "data": {
    "url": "/uploads/2026/04/abc123.pdf",
    "filename": "课件.pdf",
    "size": 1048576,
    "mime_type": "application/pdf"
  }
}
```

### 限制
- 最大文件大小：20MB
- 允许的文件类型：image/*, application/pdf, .doc, .docx, .txt, .md

### 错误码

| 错误码 | 说明 |
|--------|------|
| 40004 | 缺少文件 |
| 40021 | 文件大小超过限制 |
| 40022 | 不支持的文件类型 |

---

## 全局错误码补充

本迭代新增的错误码：

| 错误码 | 说明 | 使用场景 |
|--------|------|---------|
| 40003 | 账号已被禁用，请联系管理员 | 被禁用用户的所有请求 |
| 40020 | 不允许操作自己 | 管理员修改自己的角色/禁用自己 |
| 40021 | 文件大小超过限制 | H5文件上传 |
| 40022 | 不支持的文件类型 | H5文件上传 |

---

## 用户禁用机制（全局中间件增强）

### JWT中间件增强

现有的 JWT 认证中间件需要增加用户状态检查：

```
请求 → JWT验证 → 用户状态检查 → 业务处理
                      ↓
              status=disabled → 返回 403 + "您的账号已被禁用，请联系管理员"
```

### 实现方式

1. 维护一个内存中的**禁用用户ID集合**（Set）
2. 管理员禁用用户时，将用户ID加入集合
3. 管理员启用用户时，将用户ID从集合中移除
4. JWT中间件在验证token后，检查用户ID是否在禁用集合中
5. 服务启动时，从数据库加载所有 status=disabled 的用户ID到集合中

---

## 操作日志中间件

### 中间件位置

```
请求 → Recovery → RequestLog → CORS → RateLimit → 操作日志中间件 → JWT → 业务处理
                                                        ↓
                                                  记录请求信息
                                                        ↓
                                                  业务处理完成后
                                                        ↓
                                                  补充响应信息
                                                        ↓
                                                  异步写入日志DB
```

### 平台识别

通过请求头 `X-Platform` 或 User-Agent 识别平台：
- `X-Platform: miniapp` → 小程序
- `X-Platform: h5` → H5页面
- 未设置时，通过 User-Agent 推断

### Action 映射

中间件根据 HTTP 方法 + 路径自动映射 action：

| 方法 + 路径模式 | Action |
|----------------|--------|
| POST /api/auth/login | user.login |
| POST /api/auth/wx-login | user.login |
| POST /api/auth/wx-h5-callback | user.login |
| POST /api/auth/register | user.register |
| POST /api/auth/complete-profile | user.profile_update |
| POST /api/chat | chat.send_message |
| POST /api/chat/stream | chat.send_message |
| POST /api/chat/new-session | chat.create_session |
| POST /api/chat/teacher-reply | chat.teacher_reply |
| POST /api/classes | class.create |
| PUT /api/classes/:id | class.update |
| DELETE /api/classes/:id | class.delete |
| POST /api/classes/:id/members | class.add_member |
| DELETE /api/classes/:id/members/:member_id | class.remove_member |
| POST /api/knowledge/upload | knowledge.upload |
| PUT /api/knowledge/:id | knowledge.update |
| DELETE /api/knowledge/:id | knowledge.delete |
| POST /api/personas | persona.create |
| PUT /api/personas/:id/switch | persona.switch |
| PUT /api/personas/:id | persona.update |
| POST /api/shares | share.create |
| POST /api/shares/:code/join | share.join |
| POST /api/relations/invite | relation.invite |
| PUT /api/relations/:id/approve | relation.approve |
| PUT /api/relations/:id/reject | relation.reject |
| POST /api/courses | course.create |
| PUT /api/courses/:id | course.update |
| DELETE /api/courses/:id | course.delete |
| POST /api/courses/:id/push | course.push |
| PUT /api/admin/users/:id/role | admin.update_role |
| PUT /api/admin/users/:id/status | admin.toggle_user |
| POST /api/documents | document.upload |
| DELETE /api/documents/:id | document.delete |
| PUT /api/memories/:id | memory.update |
| DELETE /api/memories/:id | memory.delete |
| POST /api/feedbacks | feedback.create |
| 其他 GET 请求 | api.read |
| 其他 POST/PUT/DELETE 请求 | api.write |

---

## 数据库变更汇总

### 新增表（独立日志数据库）

```sql
-- 文件：LOG_DB_PATH 指定的独立数据库
CREATE TABLE IF NOT EXISTS operation_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL DEFAULT 0,
    user_role TEXT NOT NULL DEFAULT '',
    persona_id INTEGER NOT NULL DEFAULT 0,
    action TEXT NOT NULL DEFAULT '',
    resource TEXT NOT NULL DEFAULT '',
    resource_id TEXT NOT NULL DEFAULT '',
    detail TEXT NOT NULL DEFAULT '',
    ip TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    platform TEXT NOT NULL DEFAULT '',
    status_code INTEGER NOT NULL DEFAULT 0,
    duration_ms INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_operation_logs_user_created ON operation_logs(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_operation_logs_action_created ON operation_logs(action, created_at);
CREATE INDEX IF NOT EXISTS idx_operation_logs_created ON operation_logs(created_at);
```

### 修改表（业务数据库）

```sql
-- users 表新增字段
ALTER TABLE users ADD COLUMN status TEXT NOT NULL DEFAULT 'active';
ALTER TABLE users ADD COLUMN wx_unionid TEXT DEFAULT '';
```

### 管理员初始化脚本

```sql
-- scripts/set_admin.sql
-- 将指定用户设为管理员（替换 <USER_ID> 为目标用户ID）
UPDATE users SET role = 'admin' WHERE id = <USER_ID>;

-- 通过微信openid设置管理员
-- UPDATE users SET role = 'admin' WHERE openid = '<OPENID>';
```

---

## 路由注册规划

### 新增路由组

```go
// H5认证接口（无需鉴权）
authGroup.GET("/wx-h5-login-url", handler.HandleWxH5LoginURL)
authGroup.POST("/wx-h5-callback", handler.HandleWxH5Callback)

// 平台配置（无需鉴权）
api.GET("/platform/config", handler.HandleGetPlatformConfig)

// 管理员路由组（需鉴权 + admin角色）
admin := api.Group("/admin")
admin.Use(auth.JWTAuthMiddleware(jwtManager))
admin.Use(auth.RoleRequired("admin"))
{
    // 仪表盘
    admin.GET("/dashboard/overview", handler.HandleAdminDashboardOverview)
    admin.GET("/dashboard/user-stats", handler.HandleAdminUserStats)
    admin.GET("/dashboard/chat-stats", handler.HandleAdminChatStats)
    admin.GET("/dashboard/knowledge-stats", handler.HandleAdminKnowledgeStats)
    admin.GET("/dashboard/active-users", handler.HandleAdminActiveUsers)

    // 用户管理
    admin.GET("/users", handler.HandleAdminGetUsers)
    admin.PUT("/users/:id/role", handler.HandleAdminUpdateUserRole)
    admin.PUT("/users/:id/status", handler.HandleAdminUpdateUserStatus)

    // 反馈管理
    admin.GET("/feedbacks", handler.HandleAdminGetFeedbacks)

    // 操作日志
    admin.GET("/logs", handler.HandleAdminGetLogs)
    admin.GET("/logs/stats", handler.HandleAdminGetLogStats)
    admin.GET("/logs/export", handler.HandleAdminExportLogs)
}

// H5文件上传（需鉴权）
authorized.POST("/upload/h5", handler.HandleH5Upload)
```

---

## 前端项目结构规划

### 管理员H5（Vue 3 + Element Plus）

```
src/frontend-admin/
├── public/
├── src/
│   ├── api/           # API 接口封装
│   ├── assets/        # 静态资源
│   ├── components/    # 公共组件
│   ├── layouts/       # 布局组件
│   ├── router/        # 路由配置
│   ├── stores/        # Pinia 状态管理
│   ├── views/         # 页面视图
│   │   ├── login/     # 登录页
│   │   ├── dashboard/ # 仪表盘
│   │   ├── users/     # 用户管理
│   │   ├── logs/      # 日志管理
│   │   └── feedbacks/ # 反馈管理
│   ├── App.vue
│   └── main.ts
├── package.json
├── vite.config.ts
└── tsconfig.json
```

### 教师H5（Vue 3 + Element Plus）

```
src/frontend-teacher/
├── src/
│   ├── api/
│   ├── components/
│   ├── layouts/
│   ├── router/
│   ├── stores/
│   ├── views/
│   │   ├── login/
│   │   ├── home/
│   │   ├── chat/
│   │   ├── classes/
│   │   ├── knowledge/
│   │   ├── courses/
│   │   ├── personas/
│   │   ├── shares/
│   │   ├── memories/
│   │   ├── approvals/
│   │   ├── curriculum/
│   │   ├── feedback/
│   │   └── profile/
│   ├── App.vue
│   └── main.ts
├── package.json
└── vite.config.ts
```

### 学生H5（Vue 3 + Vant 4）

```
src/frontend-student/
├── src/
│   ├── api/
│   ├── components/
│   ├── layouts/
│   ├── router/
│   ├── stores/
│   ├── views/
│   │   ├── login/
│   │   ├── home/
│   │   ├── chat/
│   │   ├── discover/
│   │   ├── history/
│   │   ├── teachers/
│   │   ├── comments/
│   │   ├── personas/
│   │   ├── share-join/
│   │   ├── feedback/
│   │   └── profile/
│   ├── App.vue
│   └── main.ts
├── package.json
└── vite.config.ts
```

---

**文档版本**: v1.0
**创建日期**: 2026-04-04
**关联需求**: 迭代10 管理员H5后台 + 教师/学生H5页面 + 操作日志流水
