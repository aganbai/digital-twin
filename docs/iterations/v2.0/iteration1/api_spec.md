# V2.0 迭代1 API 接口规范

## 1. 通用约定

> 继承 V1.0 的通用约定，以下仅列出新增和变更部分。

### 1.1 基础信息

| 项目 | 说明 |
|------|------|
| Base URL | `http://localhost:8080` |
| 协议 | HTTP（开发环境），HTTPS（V2.0 迭代2 生产环境） |
| 数据格式 | JSON（`Content-Type: application/json`），文件上传使用 `multipart/form-data` |
| 认证方式 | Bearer Token（`Authorization: Bearer <token>`） |

### 1.2 新增错误码

| 错误码 | HTTP Status | 说明 |
|--------|-------------|------|
| 40007 | 403 | 未获得该教师授权 |
| 40008 | 409 | 该学校已有同名教师 |
| 40009 | 409 | 师生关系已存在 |
| 40010 | 400 | 文件格式不支持 |
| 40011 | 400 | 文件大小超限 |
| 40012 | 400 | URL 不可达或解析失败 |

---

## 2. 改造接口

### 2.1 补全用户信息（改造）

**POST** `/api/auth/complete-profile`

**鉴权**：需要（Bearer Token）

**改造说明**：教师角色新增 `school` 和 `description` 必填字段，校验 nickname + school 唯一。

**请求体**（教师）：
```json
{
  "role": "teacher",
  "nickname": "王老师",
  "school": "北京大学",
  "description": "物理学教授，专注力学和热力学教学"
}
```

**请求体**（学生，不变）：
```json
{
  "role": "student",
  "nickname": "小李"
}
```

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
|------|------|------|----------|------|
| role | string | ✅ | 枚举：teacher/student | 用户角色 |
| nickname | string | ✅ | 1-64 字符 | 昵称 |
| school | string | 教师 ✅ | 1-128 字符 | 学校名称 |
| description | string | 教师 ✅ | 1-500 字符 | 分身简短描述 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 1,
    "role": "teacher",
    "nickname": "王老师",
    "school": "北京大学",
    "description": "物理学教授，专注力学和热力学教学"
  }
}
```

**新增错误响应**：
| 场景 | code | message |
|------|------|---------|
| 教师缺少 school | 40004 | 教师角色必须填写学校名称 |
| 教师缺少 description | 40004 | 教师角色必须填写分身描述 |
| 同名+同校教师已存在 | 40008 | 该学校已有同名教师，请修改名称 |

---

### 2.2 发送对话消息（改造）

**POST** `/api/chat`

**改造说明**：增加师生授权鉴权 + 个性化问答风格注入。

**新增错误响应**：
| 场景 | code | HTTP Status | message |
|------|------|-------------|---------|
| 未获得教师授权 | 40007 | 403 | 未获得该教师授权，请先申请 |

**行为变更**：
1. 学生角色发起对话前，检查 `teacher_student_relations` 表中是否存在 `status=approved` 的记录
2. 如果存在个性化问答风格配置，自动注入到 system prompt

---

## 3. 师生关系接口

### 3.1 教师邀请学生

**POST** `/api/relations/invite`

**鉴权**：需要（Bearer Token，角色：teacher）

**请求体**：
```json
{
  "student_id": 5
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| student_id | int | ✅ | 被邀请的学生 ID |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "teacher_id": 1,
    "student_id": 5,
    "status": "approved",
    "initiated_by": "teacher",
    "created_at": "2026-05-06T10:00:00Z"
  }
}
```

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 学生不存在 | 40005 | 学生不存在 |
| 目标用户不是学生 | 40004 | 目标用户不是学生角色 |
| 关系已存在 | 40009 | 师生关系已存在 |

---

### 3.2 学生申请使用分身

**POST** `/api/relations/apply`

**鉴权**：需要（Bearer Token，角色：student）

**请求体**：
```json
{
  "teacher_id": 1
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| teacher_id | int | ✅ | 目标教师 ID |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2,
    "teacher_id": 1,
    "student_id": 5,
    "status": "pending",
    "initiated_by": "student",
    "created_at": "2026-05-06T10:05:00Z"
  }
}
```

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 教师不存在 | 40005 | 教师不存在 |
| 目标用户不是教师 | 40004 | 目标用户不是教师角色 |
| 关系已存在 | 40009 | 师生关系已存在 |

---

### 3.3 教师审批同意

**PUT** `/api/relations/:id/approve`

**鉴权**：需要（Bearer Token，角色：teacher）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 关系记录 ID |

**请求体**：无

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2,
    "status": "approved",
    "updated_at": "2026-05-06T10:10:00Z"
  }
}
```

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 关系不存在 | 40005 | 关系记录不存在 |
| 非当前教师的关系 | 40003 | 无权操作此关系 |

---

### 3.4 教师审批拒绝

**PUT** `/api/relations/:id/reject`

**鉴权**：需要（Bearer Token，角色：teacher）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 关系记录 ID |

**请求体**：无

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2,
    "status": "rejected",
    "updated_at": "2026-05-06T10:10:00Z"
  }
}
```

**错误响应**：同 3.3

---

### 3.5 获取师生关系列表

**GET** `/api/relations`

**鉴权**：需要（Bearer Token）

**Query 参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| status | string | ❌ | - | 状态筛选：pending/approved/rejected |
| page | int | ❌ | 1 | 页码 |
| page_size | int | ❌ | 20 | 每页数量 |

**成功响应** `200`（教师视角）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "student_id": 5,
        "student_nickname": "小李",
        "status": "approved",
        "initiated_by": "teacher",
        "created_at": "2026-05-06T10:00:00Z"
      },
      {
        "id": 2,
        "student_id": 6,
        "student_nickname": "小王",
        "status": "pending",
        "initiated_by": "student",
        "created_at": "2026-05-06T10:05:00Z"
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
        "teacher_id": 1,
        "teacher_nickname": "王老师",
        "teacher_school": "北京大学",
        "teacher_description": "物理学教授",
        "status": "approved",
        "initiated_by": "teacher",
        "created_at": "2026-05-06T10:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 4. 教师评语接口

### 4.1 教师写评语

**POST** `/api/comments`

**鉴权**：需要（Bearer Token，角色：teacher）

**请求体**：
```json
{
  "student_id": 5,
  "content": "该生学习态度认真，对力学概念理解较好，但在热力学方面还需加强",
  "progress_summary": "牛顿定律掌握80%，热力学掌握40%"
}
```

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
|------|------|------|----------|------|
| student_id | int | ✅ | - | 学生 ID |
| content | string | ✅ | 1-2000 字符 | 评语内容 |
| progress_summary | string | ❌ | 最长 500 字符 | 学习进度摘要 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "teacher_id": 1,
    "student_id": 5,
    "content": "该生学习态度认真...",
    "progress_summary": "牛顿定律掌握80%...",
    "created_at": "2026-05-06T10:00:00Z"
  }
}
```

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 师生关系未授权 | 40007 | 未获得该学生的授权关系 |
| 学生不存在 | 40005 | 学生不存在 |

---

### 4.2 获取评语列表

**GET** `/api/comments`

**鉴权**：需要（Bearer Token）

**Query 参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| student_id | int | ❌ | - | 学生 ID（教师筛选特定学生） |
| teacher_id | int | ❌ | - | 教师 ID（学生筛选特定教师） |
| page | int | ❌ | 1 | 页码 |
| page_size | int | ❌ | 20 | 每页数量 |

**行为说明**：
- 教师：返回自己写的评语（可按 student_id 筛选）
- 学生：返回收到的评语（可按 teacher_id 筛选）

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "teacher_id": 1,
        "teacher_nickname": "王老师",
        "student_id": 5,
        "student_nickname": "小李",
        "content": "该生学习态度认真...",
        "progress_summary": "牛顿定律掌握80%...",
        "created_at": "2026-05-06T10:00:00Z"
      }
    ],
    "total": 3,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 5. 个性化问答风格接口

### 5.1 设置学生问答风格

**PUT** `/api/students/:id/dialogue-style`

**鉴权**：需要（Bearer Token，角色：teacher）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 学生 ID |

**请求体**：
```json
{
  "temperature": 0.7,
  "guidance_level": "medium",
  "style_prompt": "对该学生请多用鼓励性语言，注重基础概念的巩固",
  "max_turns_per_topic": 5
}
```

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
|------|------|------|----------|------|
| temperature | float | ❌ | 0.1-1.0 | 回复随机性，默认 0.7 |
| guidance_level | string | ❌ | 枚举：low/medium/high | 引导程度，默认 medium |
| style_prompt | string | ❌ | 最长 500 字符 | 自定义风格描述 |
| max_turns_per_topic | int | ❌ | 1-20 | 每个话题最大追问轮次，默认 5 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "teacher_id": 1,
    "student_id": 5,
    "style_config": {
      "temperature": 0.7,
      "guidance_level": "medium",
      "style_prompt": "对该学生请多用鼓励性语言...",
      "max_turns_per_topic": 5
    },
    "updated_at": "2026-05-06T10:00:00Z"
  }
}
```

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 师生关系未授权 | 40007 | 未获得该学生的授权关系 |
| temperature 超出范围 | 40004 | temperature 必须在 0.1-1.0 之间 |

---

### 5.2 获取学生问答风格

**GET** `/api/students/:id/dialogue-style`

**鉴权**：需要（Bearer Token，角色：teacher/student）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 学生 ID |

**Query 参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| teacher_id | int | 学生角色 ✅ | 教师 ID（学生查看时需指定） |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "teacher_id": 1,
    "student_id": 5,
    "style_config": {
      "temperature": 0.7,
      "guidance_level": "medium",
      "style_prompt": "对该学生请多用鼓励性语言...",
      "max_turns_per_topic": 5
    },
    "created_at": "2026-05-06T10:00:00Z",
    "updated_at": "2026-05-06T10:00:00Z"
  }
}
```

**未设置时响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

## 6. 作业接口

### 6.1 学生提交作业

**POST** `/api/assignments`

**鉴权**：需要（Bearer Token，角色：student）

**请求体**（JSON 或 multipart/form-data）：

JSON 模式：
```json
{
  "teacher_id": 1,
  "title": "牛顿定律作业",
  "content": "牛顿第一定律是指..."
}
```

multipart/form-data 模式（含文件）：
```
teacher_id: 1
title: 牛顿定律作业
content: 牛顿第一定律是指...
file: (binary)
```

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
|------|------|------|----------|------|
| teacher_id | int | ✅ | - | 目标教师 ID |
| title | string | ✅ | 1-200 字符 | 作业标题 |
| content | string | ❌ | 最长 10000 字符 | 文本内容（content 和 file 至少一个） |
| file | file | ❌ | PDF/DOCX/TXT/MD，≤50MB | 附件文件 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "student_id": 5,
    "teacher_id": 1,
    "title": "牛顿定律作业",
    "status": "submitted",
    "created_at": "2026-05-06T10:00:00Z"
  }
}
```

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 师生关系未授权 | 40007 | 未获得该教师授权 |
| 标题为空 | 40004 | 标题不能为空 |
| 内容和文件都为空 | 40004 | 作业内容和附件至少提供一个 |

---

### 6.2 获取作业列表

**GET** `/api/assignments`

**鉴权**：需要（Bearer Token）

**Query 参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| teacher_id | int | ❌ | - | 教师 ID 筛选 |
| student_id | int | ❌ | - | 学生 ID 筛选（教师用） |
| status | string | ❌ | - | 状态筛选：submitted/reviewed |
| page | int | ❌ | 1 | 页码 |
| page_size | int | ❌ | 20 | 每页数量 |

**行为说明**：
- 教师：返回提交给自己的作业（可按 student_id / status 筛选）
- 学生：返回自己提交的作业（可按 teacher_id / status 筛选）

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "student_id": 5,
        "student_nickname": "小李",
        "teacher_id": 1,
        "teacher_nickname": "王老师",
        "title": "牛顿定律作业",
        "status": "submitted",
        "has_file": true,
        "review_count": 1,
        "created_at": "2026-05-06T10:00:00Z"
      }
    ],
    "total": 5,
    "page": 1,
    "page_size": 20
  }
}
```

---

### 6.3 获取作业详情

**GET** `/api/assignments/:id`

**鉴权**：需要（Bearer Token）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 作业 ID |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "student_id": 5,
    "student_nickname": "小李",
    "teacher_id": 1,
    "teacher_nickname": "王老师",
    "title": "牛顿定律作业",
    "content": "牛顿第一定律是指...",
    "file_path": "/uploads/assignments/5/uuid_homework.pdf",
    "file_type": "pdf",
    "status": "reviewed",
    "created_at": "2026-05-06T10:00:00Z",
    "reviews": [
      {
        "id": 1,
        "reviewer_type": "ai",
        "reviewer_id": null,
        "content": "优点：概念理解准确...\n改进：缺少实例说明...",
        "score": 78,
        "created_at": "2026-05-06T10:05:00Z"
      },
      {
        "id": 2,
        "reviewer_type": "teacher",
        "reviewer_id": 1,
        "content": "整体不错，注意公式推导过程",
        "score": 85,
        "created_at": "2026-05-06T11:00:00Z"
      }
    ]
  }
}
```

---

### 6.4 教师点评作业

**POST** `/api/assignments/:id/review`

**鉴权**：需要（Bearer Token，角色：teacher）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 作业 ID |

**请求体**：
```json
{
  "content": "整体不错，注意公式推导过程需要更严谨",
  "score": 85
}
```

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
|------|------|------|----------|------|
| content | string | ✅ | 1-2000 字符 | 点评内容 |
| score | float | ❌ | 0-100 | 评分 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2,
    "assignment_id": 1,
    "reviewer_type": "teacher",
    "reviewer_id": 1,
    "content": "整体不错...",
    "score": 85,
    "created_at": "2026-05-06T11:00:00Z"
  }
}
```

**副作用**：作业状态自动变为 `reviewed`

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 作业不存在 | 40005 | 作业不存在 |
| 非该教师的作业 | 40003 | 无权点评此作业 |

---

### 6.5 AI 自动点评

**POST** `/api/assignments/:id/ai-review`

**鉴权**：需要（Bearer Token，角色：student/teacher）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 作业 ID |

**请求体**：无

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "assignment_id": 1,
    "reviewer_type": "ai",
    "content": "## 作业点评\n\n### 优点\n1. 概念理解准确...\n\n### 改进建议\n1. 缺少实例说明...\n\n### 评分: 78/100",
    "score": 78,
    "created_at": "2026-05-06T10:05:00Z",
    "token_usage": {
      "prompt_tokens": 1200,
      "completion_tokens": 350,
      "total_tokens": 1550
    }
  }
}
```

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 作业不存在 | 40005 | 作业不存在 |
| 大模型调用失败 | 50002 | AI 点评生成失败 |

---

## 7. 知识库增强接口

### 7.1 文件上传

**POST** `/api/documents/upload`

**鉴权**：需要（Bearer Token，角色：teacher）

**Content-Type**：`multipart/form-data`

**表单字段**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | ✅ | 文件（PDF/DOCX/TXT/MD，≤50MB） |
| title | string | ❌ | 文档标题（不传则使用文件名） |
| tags | string | ❌ | 标签（逗号分隔） |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "document_id": 10,
    "title": "牛顿运动定律",
    "doc_type": "pdf",
    "chunks_count": 8,
    "file_size": 1048576,
    "status": "active"
  }
}
```

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| 文件格式不支持 | 40010 | 不支持的文件格式，仅支持 PDF/DOCX/TXT/MD |
| 文件大小超限 | 40011 | 文件大小超过限制（最大 50MB） |
| 文件解析失败 | 50001 | 文件内容解析失败 |

---

### 7.2 URL 导入

**POST** `/api/documents/import-url`

**鉴权**：需要（Bearer Token，角色：teacher）

**请求体**：
```json
{
  "url": "https://example.com/physics/newton-laws",
  "title": "牛顿运动定律（可选，不传则自动提取）",
  "tags": "物理,力学"
}
```

| 字段 | 类型 | 必填 | 校验规则 | 说明 |
|------|------|------|----------|------|
| url | string | ✅ | 合法 URL 格式 | 目标网页 URL |
| title | string | ❌ | 最长 200 字符 | 文档标题（不传则从 `<title>` 提取） |
| tags | string | ❌ | - | 标签（逗号分隔） |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "document_id": 11,
    "title": "牛顿运动定律 - 维基百科",
    "doc_type": "url",
    "chunks_count": 12,
    "content_length": 15000,
    "source_url": "https://example.com/physics/newton-laws",
    "status": "active"
  }
}
```

**错误响应**：
| 场景 | code | message |
|------|------|---------|
| URL 格式无效 | 40004 | 无效的 URL 格式 |
| URL 不可达 | 40012 | 无法访问目标 URL |
| 内容为空 | 40012 | 网页内容为空或无法解析 |
| 抓取超时 | 40012 | 网页抓取超时（10秒） |

---

## 8. 流式对话接口

### 8.1 SSE 流式对话

**POST** `/api/chat/stream`

**鉴权**：需要（Bearer Token）

**请求体**（同 `/api/chat`）：
```json
{
  "message": "什么是牛顿第一定律?",
  "teacher_id": 1,
  "session_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| message | string | ✅ | 用户消息，最长 2000 字符 |
| teacher_id | int | ✅ | 目标教师 ID |
| session_id | string | ❌ | 会话 ID，不传则自动生成 |

**响应 Headers**：
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
X-Accel-Buffering: no
```

**SSE 事件流**：

**1. 开始事件**：
```
data: {"type":"start","session_id":"550e8400-e29b-41d4-a716-446655440000"}

```

**2. 内容增量事件**（多次）：
```
data: {"type":"delta","content":"这是"}

data: {"type":"delta","content":"一个"}

data: {"type":"delta","content":"很好的问题！"}

data: {"type":"delta","content":"在我们讨论"}

data: {"type":"delta","content":"牛顿第一定律之前，"}

```

**3. 完成事件**：
```
data: {"type":"done","conversation_id":42,"token_usage":{"prompt_tokens":850,"completion_tokens":120,"total_tokens":970}}

```

**4. 错误事件**（异常时）：
```
data: {"type":"error","code":50002,"message":"大模型调用失败"}

```

**SSE 事件类型说明**：

| type | 说明 | 字段 |
|------|------|------|
| start | 流开始 | session_id |
| delta | 内容增量 | content（文本片段） |
| done | 流结束 | conversation_id, token_usage |
| error | 错误 | code, message |

**错误响应**（非 SSE，在流开始前返回）：
| 场景 | code | HTTP Status | message |
|------|------|-------------|---------|
| 未获得教师授权 | 40007 | 403 | 未获得该教师授权，请先申请 |
| 教师不存在 | 40005 | 404 | 教师不存在 |
| 参数校验失败 | 40004 | 400 | 请求参数无效 |

---

## 9. 教师列表接口（增强）

### 9.1 获取教师列表（增强）

**GET** `/api/teachers`

**改造说明**：返回数据新增 `school` 和 `description` 字段。

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

---

## 10. 接口总览

### 10.1 新增接口（17 个）

| 编号 | 方法 | 路径 | 说明 | 角色 |
|------|------|------|------|------|
| API-01 | POST | `/api/relations/invite` | 教师邀请学生 | teacher |
| API-02 | POST | `/api/relations/apply` | 学生申请使用分身 | student |
| API-03 | PUT | `/api/relations/:id/approve` | 教师审批同意 | teacher |
| API-04 | PUT | `/api/relations/:id/reject` | 教师审批拒绝 | teacher |
| API-05 | GET | `/api/relations` | 获取师生关系列表 | 所有 |
| API-06 | POST | `/api/comments` | 教师写评语 | teacher |
| API-07 | GET | `/api/comments` | 获取评语列表 | 所有 |
| API-08 | PUT | `/api/students/:id/dialogue-style` | 设置学生问答风格 | teacher |
| API-09 | GET | `/api/students/:id/dialogue-style` | 获取学生问答风格 | teacher/student |
| API-10 | POST | `/api/assignments` | 学生提交作业 | student |
| API-11 | GET | `/api/assignments` | 获取作业列表 | 所有 |
| API-12 | GET | `/api/assignments/:id` | 获取作业详情 | 所有 |
| API-13 | POST | `/api/assignments/:id/review` | 教师点评作业 | teacher |
| API-14 | POST | `/api/assignments/:id/ai-review` | AI 自动点评 | student/teacher |
| API-15 | POST | `/api/documents/upload` | 文件上传 | teacher |
| API-16 | POST | `/api/documents/import-url` | URL 导入 | teacher |
| API-17 | POST | `/api/chat/stream` | SSE 流式对话 | student |

### 10.2 改造接口（2 个）

| 接口 | 改造内容 |
|------|----------|
| `POST /api/auth/complete-profile` | 教师必填 school + description，唯一校验 |
| `POST /api/chat` | 师生授权鉴权 + 个性化问答风格注入 |

### 10.3 增强接口（1 个）

| 接口 | 增强内容 |
|------|----------|
| `GET /api/teachers` | 返回新增 school + description 字段 |

---

## 11. 路由注册参考

```go
// router.go 新增路由

// 师生关系
relations := authorized.Group("/relations")
{
    relations.POST("/invite", auth.RoleRequired("teacher"), handler.HandleInviteStudent)
    relations.POST("/apply", auth.RoleRequired("student"), handler.HandleApplyTeacher)
    relations.PUT("/:id/approve", auth.RoleRequired("teacher"), handler.HandleApproveRelation)
    relations.PUT("/:id/reject", auth.RoleRequired("teacher"), handler.HandleRejectRelation)
    relations.GET("", handler.HandleGetRelations)
}

// 评语
authorized.POST("/comments", auth.RoleRequired("teacher"), handler.HandleCreateComment)
authorized.GET("/comments", handler.HandleGetComments)

// 问答风格
authorized.PUT("/students/:id/dialogue-style", auth.RoleRequired("teacher"), handler.HandleSetDialogueStyle)
authorized.GET("/students/:id/dialogue-style", handler.HandleGetDialogueStyle)

// 作业
assignments := authorized.Group("/assignments")
{
    assignments.POST("", auth.RoleRequired("student"), handler.HandleSubmitAssignment)
    assignments.GET("", handler.HandleGetAssignments)
    assignments.GET("/:id", handler.HandleGetAssignmentDetail)
    assignments.POST("/:id/review", auth.RoleRequired("teacher"), handler.HandleReviewAssignment)
    assignments.POST("/:id/ai-review", handler.HandleAIReviewAssignment)
}

// 知识库增强
authorized.POST("/documents/upload", auth.RoleRequired("teacher", "admin"), handler.HandleUploadDocument)
authorized.POST("/documents/import-url", auth.RoleRequired("teacher", "admin"), handler.HandleImportURL)

// SSE 流式对话
authorized.POST("/chat/stream", handler.HandleChatStream)
```

---

**文档版本**: v1.0.0
**创建日期**: 2026-03-28
**最后更新**: 2026-03-28
