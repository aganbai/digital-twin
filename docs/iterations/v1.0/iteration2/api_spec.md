# 第二迭代 API 接口规范

## 1. 通用约定

> 本文档仅描述第二迭代**新增和变更**的接口。第一迭代已有接口（认证、对话、知识库、记忆、系统）的规范不变，请参考 `iteration1_api_spec.md`。

### 1.1 基础信息

| 项目 | 说明 |
|------|------|
| Base URL | `http://localhost:8080`（开发环境） |
| 协议 | HTTP（开发环境），HTTPS（生产环境） |
| 数据格式 | JSON（`Content-Type: application/json`） |
| 字符编码 | UTF-8 |
| 认证方式 | Bearer Token（`Authorization: Bearer <token>`） |

### 1.2 统一响应格式（不变）

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

### 1.3 错误码表（不变，沿用第一迭代）

| 错误码 | HTTP Status | 说明 |
|--------|-------------|------|
| 0 | 200 | 成功 |
| 40001 | 401 | 未认证 / 令牌无效 |
| 40002 | 401 | 令牌已过期 |
| 40003 | 403 | 权限不足 |
| 40004 | 400 | 请求参数校验失败 |
| 40005 | 404 | 资源不存在 |
| 40006 | 409 | 用户名已存在 |
| 50001 | 500 | 服务器内部错误 |
| 50002 | 502 | 大模型调用失败 |
| 50003 | 502 | 向量数据库错误 |
| 50004 | 504 | 管道执行超时 |

---

## 2. 🆕 新增接口

### 2.1 微信登录

**POST** `/api/auth/wx-login`

**鉴权**：无

**说明**：小程序前端调用 `wx.login()` 获取临时 code，将 code 发送给后端。后端用 code 调用微信 `jscode2session` 接口获取 openid，然后根据 openid 查找或创建用户，返回 JWT Token。

**请求体**：
```json
{
  "code": "0a3Xyz000abc12Ghi3000jkl4m3Xyz0A"
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | ✅ | 微信 wx.login() 返回的临时 code |

**成功响应** `200`（已有用户）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 1,
    "token": "eyJ...",
    "role": "teacher",
    "nickname": "王老师",
    "is_new_user": false,
    "expires_at": "2026-04-02T13:00:00Z"
  }
}
```

**成功响应** `200`（新用户）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 5,
    "token": "eyJ...",
    "role": "",
    "nickname": "",
    "is_new_user": true,
    "expires_at": "2026-04-02T13:00:00Z"
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | int | 用户 ID |
| token | string | JWT Token |
| role | string | 用户角色，新用户为空字符串 |
| nickname | string | 昵称，新用户为空字符串 |
| is_new_user | bool | 是否为新用户（前端据此决定是否跳转角色选择页） |
| expires_at | string | Token 过期时间 |

**错误响应**：

| 场景 | code | message |
|------|------|---------|
| code 为空 | 40004 | 缺少 code 参数 |
| 微信 API 调用失败 | 50001 | 微信登录失败 |
| code 无效/过期 | 40004 | 无效的登录凭证 |

**实现说明**：
- 后端定义 `WxClient` 接口，生产环境调用 `https://api.weixin.qq.com/sns/jscode2session`
- 测试/开发环境通过 `WX_MODE=mock` 使用 `MockWxClient`（code → `mock_openid_{code}`）
- 根据 openid 查询 `users` 表：
  - 找到 → 已有用户，`is_new_user = false`
  - 未找到 → 创建新用户（role 为空，nickname 为空），`is_new_user = true`
- 无论新旧用户，都生成 JWT Token 返回

---

### 2.2 新用户补全信息

**POST** `/api/auth/complete-profile`

**鉴权**：需要（Bearer Token）

**说明**：微信登录后的新用户补全角色和昵称信息。只有 `role` 为空的用户才能调用此接口。

**请求体**：
```json
{
  "role": "teacher",
  "nickname": "王老师"
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| role | string | ✅ | 角色，可选值：`teacher` / `student` |
| nickname | string | ✅ | 昵称，1-20 字符 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 5,
    "role": "teacher",
    "nickname": "王老师"
  }
}
```

**错误响应**：

| 场景 | code | message |
|------|------|---------|
| role 不合法 | 40004 | 角色只能是 teacher 或 student |
| nickname 为空 | 40004 | 昵称不能为空 |
| 用户已有角色（重复调用） | 40004 | 用户信息已完善，无需重复设置 |

**实现说明**：
- 从 JWT Token 中解析 `user_id`
- 查询用户，检查 `role` 是否为空（只有新用户才能补全）
- 更新 `users` 表的 `role` 和 `nickname` 字段

---

### 2.3 获取教师列表

**GET** `/api/teachers`

**鉴权**：需要（Bearer Token）

**说明**：学生端用于选择教师数字分身。返回所有角色为 `teacher` 的用户列表，以及每位教师的知识库文档数量。

**Query 参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | ❌ | 1 | 页码 |
| page_size | int | ❌ | 20 | 每页数量，最大 100 |
| keyword | string | ❌ | - | 按昵称/用户名模糊搜索（P1，本迭代可不实现） |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "username": "teacher_wang",
        "nickname": "王老师",
        "role": "teacher",
        "document_count": 5,
        "created_at": "2026-04-01T09:00:00Z"
      },
      {
        "id": 2,
        "username": "teacher_li",
        "nickname": "李老师",
        "role": "teacher",
        "document_count": 12,
        "created_at": "2026-04-02T10:00:00Z"
      }
    ],
    "total": 2,
    "page": 1,
    "page_size": 20
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int | 教师用户 ID |
| username | string | 用户名 |
| nickname | string | 昵称（可能为空） |
| role | string | 固定为 "teacher" |
| document_count | int | 该教师的知识库文档数量 |
| created_at | string | 注册时间 |

**实现说明**：
- 查询 `users` 表中 `role = 'teacher'` 的记录
- `document_count` 通过 LEFT JOIN `documents` 表按 `teacher_id` 统计
- 不返回密码等敏感字段

---

### 2.4 获取当前用户信息

**GET** `/api/user/profile`

**鉴权**：需要（Bearer Token）

**说明**：前端个人中心页面使用，获取当前登录用户的详细信息。

**请求体**：无

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 3,
    "username": "student_li",
    "nickname": "小李",
    "role": "student",
    "email": "",
    "created_at": "2026-04-01T10:00:00Z",
    "stats": {
      "conversation_count": 15,
      "memory_count": 8
    }
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int | 用户 ID |
| username | string | 用户名 |
| nickname | string | 昵称 |
| role | string | 角色（teacher/student/admin） |
| email | string | 邮箱（可能为空） |
| created_at | string | 注册时间 |
| stats | object | 用户统计信息 |
| stats.conversation_count | int | 对话总数（学生：自己的对话数；教师：被提问的对话数） |
| stats.memory_count | int | 记忆总数（仅学生有意义） |

**实现说明**：
- 从 JWT Token 中解析 `user_id`
- 查询 `users` 表获取用户信息
- `conversation_count` 通过 COUNT `conversations` 表获取
- `memory_count` 通过 COUNT `memories` 表获取
- 教师角色时，`stats` 中可额外返回 `document_count`（知识库文档数）

**教师角色的响应示例**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "username": "teacher_wang",
    "nickname": "王老师",
    "role": "teacher",
    "email": "wang@example.com",
    "created_at": "2026-04-01T09:00:00Z",
    "stats": {
      "document_count": 5,
      "conversation_count": 42
    }
  }
}
```

---

### 2.5 获取会话列表

**GET** `/api/conversations/sessions`

**鉴权**：需要（Bearer Token）

**说明**：前端对话历史页使用，返回当前学生与各教师的会话摘要列表。每个会话显示最后一条消息和时间。

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
        "session_id": "550e8400-e29b-41d4-a716-446655440000",
        "teacher_id": 1,
        "teacher_nickname": "王老师",
        "last_message": "你觉得一个物体在没有外力作用时会怎样运动呢？",
        "last_message_role": "assistant",
        "message_count": 12,
        "updated_at": "2026-04-05T14:30:00Z"
      },
      {
        "session_id": "660e8400-e29b-41d4-a716-446655440001",
        "teacher_id": 2,
        "teacher_nickname": "李老师",
        "last_message": "什么是光合作用？",
        "last_message_role": "user",
        "message_count": 4,
        "updated_at": "2026-04-04T10:15:00Z"
      }
    ],
    "total": 2,
    "page": 1,
    "page_size": 20
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| session_id | string | 会话 ID |
| teacher_id | int | 教师 ID |
| teacher_nickname | string | 教师昵称 |
| last_message | string | 最后一条消息内容（截断至 100 字符） |
| last_message_role | string | 最后一条消息的角色（user/assistant） |
| message_count | int | 该会话的消息总数 |
| updated_at | string | 最后消息时间 |

**实现说明**：
- 从 JWT Token 中解析 `student_id`
- 查询 `conversations` 表，按 `session_id` 分组
- 每组取最后一条消息作为摘要
- LEFT JOIN `users` 表获取教师昵称
- 按 `updated_at` 降序排列（最近的会话在前）

**SQL 参考**：
```sql
SELECT 
    c.session_id,
    c.teacher_id,
    u.nickname as teacher_nickname,
    c.content as last_message,
    c.role as last_message_role,
    COUNT(*) OVER (PARTITION BY c.session_id) as message_count,
    c.created_at as updated_at
FROM conversations c
JOIN users u ON c.teacher_id = u.id
WHERE c.student_id = ? 
    AND c.id IN (
        SELECT MAX(id) FROM conversations 
        WHERE student_id = ? 
        GROUP BY session_id
    )
ORDER BY c.created_at DESC
LIMIT ? OFFSET ?
```

---

## 3. 🔧 变更接口

> **说明**：第一迭代的 `POST /api/auth/login` 和 `POST /api/auth/register` 接口保留不删除（用于集成测试和后台管理），但前端不再使用。这两个接口的响应中已包含 `nickname` 字段（第一迭代 auth_plugin.go 的 handleLogin 和 handleRegister 已返回 nickname），无需额外修改。

### 3.1 获取对话历史（增强）

**GET** `/api/conversations`

**变更说明**：`teacher_id` 参数从必填改为可选。不传时返回当前学生与所有教师的对话记录。

**变更前**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| teacher_id | int | ✅ | 教师 ID |

**变更后**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| teacher_id | int | ❌ | - | 教师 ID，不传则返回所有教师的对话 |
| session_id | string | ❌ | - | 会话 ID，不传则返回所有会话 |
| page | int | ❌ | 1 | 页码 |
| page_size | int | ❌ | 20 | 每页数量 |

**向后兼容**：✅ 参数从必填改为可选，已有调用方传 teacher_id 的行为不变。

---

## 4. CORS 配置变更

### 4.1 新增允许的来源

```yaml
# harness.yaml 变更
security:
  cors:
    allowed_origins:
      - "http://localhost:3000"
      - "http://localhost:10086"      # 🆕 Taro 开发服务器
      - "http://localhost:*"          # 🆕 通配本地开发端口
      - "https://servicewechat.com"   # 🆕 微信小程序
      - "https://your-app.com"
    allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]  # 🆕 增加 OPTIONS
    allowed_headers: ["Content-Type", "Authorization", "X-Requested-With"]  # 🆕 增加 X-Requested-With
    allow_credentials: true  # 🆕 允许携带凭证
```

### 4.2 小程序请求适配说明

微信小程序的网络请求有以下特殊性，后端需注意：

1. **无 CORS 预检**：小程序不走浏览器 CORS 机制，但开发者工具中会模拟浏览器行为，需要正确处理 OPTIONS 请求
2. **请求头限制**：小程序只能设置有限的请求头，`Authorization` 是允许的
3. **域名白名单**：正式发布时需要在小程序后台配置服务器域名，开发阶段可在开发者工具中关闭域名校验
4. **HTTPS 要求**：正式环境必须 HTTPS，开发环境可用 HTTP

---

## 5. 前端 API 调用规范

### 5.1 请求封装示例

```typescript
// src/api/request.ts
import Taro from '@tarojs/taro'

const BASE_URL = 'http://localhost:8080'

interface RequestOptions {
  url: string
  method: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: any
  header?: Record<string, string>
}

interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

export async function request<T = any>(options: RequestOptions): Promise<ApiResponse<T>> {
  const token = Taro.getStorageSync('token')
  
  const header: Record<string, string> = {
    'Content-Type': 'application/json',
    ...options.header,
  }
  
  if (token) {
    header['Authorization'] = `Bearer ${token}`
  }
  
  try {
    const response = await Taro.request({
      url: `${BASE_URL}${options.url}`,
      method: options.method,
      data: options.data,
      header,
    })
    
    const result = response.data as ApiResponse<T>
    
    // 统一错误处理
    if (result.code === 40001 || result.code === 40002) {
      // Token 无效或过期
      Taro.removeStorageSync('token')
      Taro.removeStorageSync('userInfo')
      Taro.redirectTo({ url: '/pages/login/index' })
      throw new Error('登录已过期，请重新登录')
    }
    
    if (result.code !== 0) {
      Taro.showToast({ title: result.message, icon: 'none' })
      throw new Error(result.message)
    }
    
    return result
  } catch (error) {
    if (error instanceof Error && error.message.includes('request:fail')) {
      Taro.showToast({ title: '网络异常，请检查网络连接', icon: 'none' })
    }
    throw error
  }
}
```

### 5.2 API 定义示例

```typescript
// src/api/auth.ts
import Taro from '@tarojs/taro'
import { request } from './request'

// 微信登录：先调用 wx.login 获取 code，再发送给后端
export async function wxLogin() {
  const { code } = await Taro.login()
  return request({
    url: '/api/auth/wx-login',
    method: 'POST',
    data: { code },
  })
}

// 新用户补全信息
export function completeProfile(role: string, nickname: string) {
  return request({
    url: '/api/auth/complete-profile',
    method: 'POST',
    data: { role, nickname },
  })
}

// src/api/teacher.ts
import { request } from './request'

export function getTeachers(page = 1, pageSize = 20) {
  return request({
    url: `/api/teachers?page=${page}&page_size=${pageSize}`,
    method: 'GET',
  })
}

// src/api/chat.ts
import { request } from './request'

export function sendMessage(message: string, teacherId: number, sessionId?: string) {
  return request({
    url: '/api/chat',
    method: 'POST',
    data: { message, teacher_id: teacherId, session_id: sessionId },
  })
}

export function getConversations(params: {
  teacher_id?: number
  session_id?: string
  page?: number
  page_size?: number
}) {
  const query = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined) query.append(key, String(value))
  })
  return request({
    url: `/api/conversations?${query.toString()}`,
    method: 'GET',
  })
}

export function getSessions(page = 1, pageSize = 20) {
  return request({
    url: `/api/conversations/sessions?page=${page}&page_size=${pageSize}`,
    method: 'GET',
  })
}

// src/api/user.ts
import { request } from './request'

export function getUserProfile() {
  return request({
    url: '/api/user/profile',
    method: 'GET',
  })
}
```

---

## 6. 接口完整清单（第二迭代全量）

### 6.1 第一迭代已有接口

| 方法 | 路径 | 说明 | 鉴权 | 状态 |
|------|------|------|------|------|
| POST | `/api/auth/register` | 用户注册（保留，前端不使用） | 无 | 不变 |
| POST | `/api/auth/login` | 用户登录（保留，前端不使用） | 无 | 不变 |
| POST | `/api/auth/refresh` | 刷新令牌 | 需要 | 不变 |
| POST | `/api/chat` | 发送对话消息 | 需要 | 不变 |
| GET | `/api/conversations` | 获取对话历史 | 需要 | 🔧 teacher_id 改为可选 |
| POST | `/api/documents` | 添加知识文档 | 需要（teacher） | 不变 |
| GET | `/api/documents` | 获取文档列表 | 需要（teacher） | 不变 |
| DELETE | `/api/documents/:id` | 删除文档 | 需要（teacher） | 不变 |
| GET | `/api/memories` | 获取学生记忆 | 需要 | 不变 |
| GET | `/api/system/health` | 健康检查 | 无 | 不变 |
| GET | `/api/system/plugins` | 插件列表 | 需要（admin） | 不变 |
| GET | `/api/system/pipelines` | 管道列表 | 需要（admin） | 不变 |

### 6.2 第二迭代新增接口

| 方法 | 路径 | 说明 | 鉴权 | 状态 |
|------|------|------|------|------|
| POST | `/api/auth/wx-login` | 微信登录 | 无 | 🆕 新增 |
| POST | `/api/auth/complete-profile` | 新用户补全信息 | 需要 | 🆕 新增 |
| GET | `/api/teachers` | 获取教师列表 | 需要 | 🆕 新增 |
| GET | `/api/user/profile` | 获取当前用户信息 | 需要 | 🆕 新增 |
| GET | `/api/conversations/sessions` | 获取会话列表 | 需要 | 🆕 新增 |

---

## 7. 测试用例（curl 命令）

### 微信登录（mock 模式）
```bash
# 需要设置环境变量 WX_MODE=mock
curl -X POST http://localhost:8080/api/auth/wx-login \
  -H "Content-Type: application/json" \
  -d '{"code": "test_teacher_001"}'
# 期望响应: is_new_user=true, role=""
```

### 新用户补全信息
```bash
curl -X POST http://localhost:8080/api/auth/complete-profile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <new_user_token>" \
  -d '{"role": "teacher", "nickname": "王老师"}'
# 期望响应: role="teacher", nickname="王老师"
```

### 同一 openid 再次登录
```bash
curl -X POST http://localhost:8080/api/auth/wx-login \
  -H "Content-Type: application/json" \
  -d '{"code": "test_teacher_001"}'
# 期望响应: is_new_user=false, role="teacher", nickname="王老师"
```

### 获取教师列表
```bash
curl -X GET "http://localhost:8080/api/teachers?page=1&page_size=20" \
  -H "Authorization: Bearer <student_token>"
```

### 获取当前用户信息
```bash
curl -X GET http://localhost:8080/api/user/profile \
  -H "Authorization: Bearer <token>"
```

### 获取会话列表
```bash
curl -X GET "http://localhost:8080/api/conversations/sessions?page=1&page_size=20" \
  -H "Authorization: Bearer <student_token>"
```

### 获取对话历史（不传 teacher_id）
```bash
curl -X GET "http://localhost:8080/api/conversations?page=1&page_size=50" \
  -H "Authorization: Bearer <student_token>"
```

---

## 8. 环境变量（新增/变更）

| 变量名 | 必填 | 默认值 | 说明 | 状态 |
|--------|------|--------|------|------|
| `TARO_APP_API` | ❌ | `http://localhost:8080` | 前端 API 地址 | 🆕 前端使用 |
| `WX_APPID` | 生产必填 | - | 微信小程序 AppID | 🆕 后端使用 |
| `WX_SECRET` | 生产必填 | - | 微信小程序 AppSecret | 🆕 后端使用 |
| `WX_MODE` | ❌ | `real` | 微信 API 模式：`real` / `mock` | 🆕 后端使用 |

> 其他后端环境变量不变，沿用第一迭代的 `.env` 配置。
> 开发和测试环境设置 `WX_MODE=mock`，无需真实的 AppID 和 Secret。

---

**文档版本**: v1.1.0
**创建日期**: 2026-03-28
**最后更新**: 2026-03-28
**变更记录**:
- v1.0.0: 初始版本（用户名密码登录）
- v1.1.0: 登录方式改为微信登录，新增 wx-login 和 complete-profile 接口，移除 login/register 响应增强变更
