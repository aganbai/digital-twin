# V2.0 迭代7 API 接口规范

## 1. 通用约定

> 继承 V2.0 迭代6 的通用约定，以下仅列出新增和变更部分。

### 1.1 基础信息

| 项目 | 说明 |
|------|------|
| Go 后端 Base URL | `http://localhost:8080` |
| Python LlamaIndex 服务 Base URL | `http://localhost:8100`（内部调用，不对外暴露） |
| 协议 | HTTP（开发环境）/ HTTPS（生产环境，通过 Nginx） |
| 数据格式 | JSON（`Content-Type: application/json`） |
| 认证方式 | Bearer Token（`Authorization: Bearer <token>`） |

### 1.2 新增错误码

| 错误码 | HTTP Status | 说明 |
|--------|-------------|------|
| 40041 | 400 | 无效的学段类型（grade_level 不在8档枚举中） |
| 40042 | 400 | 无效的反馈类型（feedback_type 不在枚举中） |
| 40043 | 400 | 批量上传文件数超限（最多20个） |
| 40044 | 400 | 批量上传总大小超限（最大100MB） |
| 40045 | 400 | 不支持的文件格式（仅支持 PDF/DOCX/TXT/MD） |
| 40046 | 404 | 批量任务不存在 |
| 40047 | 400 | LLM解析学生文本失败 |
| 40048 | 400 | 批量创建学生数据为空 |
| 40049 | 400 | 推送消息内容为空 |
| 40050 | 429 | 教师每日推送次数超限（最多20条/天） |
| 40051 | 429 | API请求频率超限 |

---

## 2. 新增接口

### 2.1 教材配置 - 创建/更新

**POST** `/api/curriculum-configs`

**鉴权**：需要（Bearer Token，仅教师角色）

**请求体**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| persona_id | int64 | ✅ | 教师分身 ID |
| grade_level | string | ✅ | 学段标识（8档枚举，见下方） |
| grade | string | ❌ | 具体年级（如"三年级"、"高一"） |
| textbook_versions | []string | ❌ | 教材版本数组（如 `["人教版", "北师大版"]`） |
| subjects | []string | ❌ | 学科数组（如 `["数学", "物理"]`；成人学段为课程类别） |
| current_progress | object | ❌ | 当前教学进度（JSON对象） |
| region | string | ❌ | 地区（用于推荐默认教材版本） |

**grade_level 枚举值**：

| 值 | 名称 | 年级范围 |
|----|------|---------|
| `preschool` | 学前班 | 幼儿园大班~学前 |
| `primary_lower` | 小学低年级 | 1-3年级 |
| `primary_upper` | 小学高年级 | 4-6年级 |
| `junior` | 初中 | 7-9年级 |
| `senior` | 高中 | 10-12年级 |
| `university` | 大学及以上 | 大学/研究生/博士 |
| `adult_life` | 成人生活技能 | 烹饪/健身/手工等 |
| `adult_professional` | 成人职业培训 | 职业技能/考证等 |

**请求示例**：
```json
{
  "persona_id": 1,
  "grade_level": "primary_upper",
  "grade": "五年级",
  "textbook_versions": ["人教版"],
  "subjects": ["数学", "语文"],
  "current_progress": {"数学": "第三章 小数乘法"},
  "region": "北京"
}
```

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "teacher_id": 10,
    "persona_id": 1,
    "grade_level": "primary_upper",
    "grade": "五年级",
    "textbook_versions": ["人教版"],
    "subjects": ["数学", "语文"],
    "current_progress": {"数学": "第三章 小数乘法"},
    "region": "北京",
    "is_active": true,
    "created_at": "2026-04-01T10:00:00Z",
    "updated_at": "2026-04-01T10:00:00Z"
  }
}
```

**失败响应**：

| 场景 | 错误码 | 说明 |
|------|--------|------|
| 无效学段 | 40041 | grade_level 不在枚举中 |
| 非教师角色 | 40003 | 权限不足 |

---

### 2.2 教材配置 - 查询

**GET** `/api/curriculum-configs`

**鉴权**：需要（Bearer Token，仅教师角色）

**查询参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| persona_id | int64 | ✅ | 教师分身 ID |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "teacher_id": 10,
      "persona_id": 1,
      "grade_level": "primary_upper",
      "grade": "五年级",
      "textbook_versions": ["人教版"],
      "subjects": ["数学", "语文"],
      "current_progress": {"数学": "第三章 小数乘法"},
      "region": "北京",
      "is_active": true,
      "created_at": "2026-04-01T10:00:00Z",
      "updated_at": "2026-04-01T10:00:00Z"
    }
  ]
}
```

---

### 2.3 教材配置 - 更新

**PUT** `/api/curriculum-configs/:id`

**鉴权**：需要（Bearer Token，仅教师角色）

**请求体**：与 POST 相同（除 persona_id 外，所有字段可选更新）

**成功响应** `200`：返回更新后的完整配置对象

**失败响应**：

| 场景 | 错误码 | 说明 |
|------|--------|------|
| 配置不存在 | 40004 | 资源不存在 |
| 无效学段 | 40041 | grade_level 不在枚举中 |

---

### 2.4 教材配置 - 删除

**DELETE** `/api/curriculum-configs/:id`

**鉴权**：需要（Bearer Token，仅教师角色）

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 2.5 教材版本列表

**GET** `/api/curriculum-versions`

**鉴权**：需要（Bearer Token）

**查询参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| grade_level | string | ❌ | 按学段筛选可用教材版本 |
| region | string | ❌ | 按地区筛选推荐版本 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "versions": ["人教版", "北师大版", "苏教版", "沪教版", "部编版", "外研版"],
    "recommended": "人教版"
  }
}
```

---

### 2.6 提交反馈

**POST** `/api/feedbacks`

**鉴权**：需要（Bearer Token）

**请求体**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| feedback_type | string | ✅ | 反馈类型：`suggestion` / `bug` / `content_issue` / `other` |
| content | string | ✅ | 反馈内容（最大2000字） |
| context_info | object | ❌ | 上下文信息（页面、设备等，前端自动填充） |

**请求示例**：
```json
{
  "feedback_type": "suggestion",
  "content": "希望能支持语音输入功能",
  "context_info": {
    "page": "chat",
    "device": "iPhone 15",
    "os": "iOS 18"
  }
}
```

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "user_id": 10,
    "feedback_type": "suggestion",
    "content": "希望能支持语音输入功能",
    "status": "pending",
    "context_info": {"page": "chat", "device": "iPhone 15", "os": "iOS 18"},
    "created_at": "2026-04-01T10:00:00Z"
  }
}
```

---

### 2.7 反馈列表

**GET** `/api/feedbacks`

**鉴权**：需要（Bearer Token，仅教师/管理员角色）

**查询参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| status | string | ❌ | - | 按状态筛选：`pending` / `reviewed` / `resolved` |
| page | int | ❌ | 1 | 页码 |
| page_size | int | ❌ | 20 | 每页数量（最大50） |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 5
  }
}
```

---

### 2.8 更新反馈状态

**PUT** `/api/feedbacks/:id/status`

**鉴权**：需要（Bearer Token，仅教师/管理员角色）

**请求体**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status | string | ✅ | 新状态：`reviewed` / `resolved` |

---

### 2.9 批量文件上传

**POST** `/api/documents/batch-upload`

**鉴权**：需要（Bearer Token，仅教师角色）

**请求格式**：`multipart/form-data`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| files | []file | ✅ | 文件列表（最多20个，总大小≤100MB） |
| persona_id | int64 | ✅ | 教师分身 ID |
| knowledge_base_id | int64 | ❌ | 知识库 ID（不传则使用默认知识库） |

**文件限制**：

| 限制项 | 值 |
|--------|-----|
| 最大文件数 | 20 |
| 总大小上限 | 100MB |
| 支持格式 | PDF / DOCX / TXT / MD |

**成功响应** `202`：
```json
{
  "code": 0,
  "message": "任务已提交，正在后台处理",
  "data": {
    "task_id": "task_abc123",
    "status": "pending",
    "total_files": 5
  }
}
```

---

### 2.10 查询批量任务状态

**GET** `/api/batch-tasks/:task_id`

**鉴权**：需要（Bearer Token）

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "task_abc123",
    "status": "partial_success",
    "total_files": 5,
    "success_files": 4,
    "failed_files": 1,
    "result_json": {
      "results": [
        {"filename": "ch1.pdf", "status": "success", "document_id": 42},
        {"filename": "ch2.docx", "status": "success", "document_id": 43},
        {"filename": "bad.txt", "status": "failed", "error": "文件内容为空"}
      ]
    },
    "created_at": "2026-04-01T10:00:00Z",
    "updated_at": "2026-04-01T10:05:00Z"
  }
}
```

**status 枚举**：`pending` / `processing` / `success` / `partial_success` / `failed`

---

### 2.11 LLM解析学生文本

**POST** `/api/students/parse-text`

**鉴权**：需要（Bearer Token，仅教师角色）

**请求体**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| text | string | ✅ | 教师粘贴的文本（花名册等） |

**请求示例**：
```json
{
  "text": "张三 男 13岁 数学好\n李四 女 12岁 英语好\n王五 男 13岁"
}
```

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "students": [
      {
        "name": "张三",
        "gender": "male",
        "age": 13,
        "strengths": "数学",
        "weaknesses": "",
        "notes": ""
      },
      {
        "name": "李四",
        "gender": "female",
        "age": 12,
        "strengths": "英语",
        "weaknesses": "",
        "notes": ""
      },
      {
        "name": "王五",
        "gender": "male",
        "age": 13,
        "strengths": "",
        "weaknesses": "",
        "notes": ""
      }
    ],
    "parse_method": "llm"
  }
}
```

> `parse_method` 为 `llm` 表示LLM解析成功，为 `rule` 表示LLM失败后规则回退解析。

---

### 2.12 批量创建学生

**POST** `/api/students/batch-create`

**鉴权**：需要（Bearer Token，仅教师角色）

**请求体**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| persona_id | int64 | ✅ | 教师分身 ID |
| students | []object | ✅ | 学生信息数组 |

**students 数组元素**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | ✅ | 姓名 |
| gender | string | ✅ | 性别：`male` / `female` |
| age | int | ❌ | 年龄 |
| student_id | string | ❌ | 学号 |
| strengths | string | ❌ | 擅长学科 |
| weaknesses | string | ❌ | 薄弱学科 |
| learning_style | string | ❌ | 学习风格偏好 |
| personality_tags | []string | ❌ | 性格标签（多选） |
| interests | string | ❌ | 兴趣爱好 |
| specialties | string | ❌ | 特长 |
| parent_notes | string | ❌ | 家长备注 |
| extra_info | object | ❌ | 自定义扩展字段 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 3,
    "success": 3,
    "failed": 0,
    "results": [
      {"name": "张三", "status": "success", "user_id": 101, "persona_id": 201},
      {"name": "李四", "status": "success", "user_id": 102, "persona_id": 202},
      {"name": "王五", "status": "success", "user_id": 103, "persona_id": 203}
    ]
  }
}
```

---

### 2.13 教师推送消息

**POST** `/api/teacher-messages`

**鉴权**：需要（Bearer Token，仅教师角色）

**请求体**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| target_type | string | ✅ | 目标类型：`class` / `student` |
| target_id | int64 | ✅ | 目标 ID（class_id 或 student_persona_id） |
| content | string | ✅ | 消息内容（最大1000字） |
| persona_id | int64 | ✅ | 教师分身 ID |

**请求示例**：
```json
{
  "target_type": "class",
  "target_id": 1,
  "content": "同学们，明天数学课请带好三角尺",
  "persona_id": 1
}
```

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "teacher_id": 10,
    "target_type": "class",
    "target_id": 1,
    "content": "同学们，明天数学课请带好三角尺",
    "status": "sent",
    "created_at": "2026-04-01T10:00:00Z"
  }
}
```

**失败响应**：

| 场景 | 错误码 | 说明 |
|------|--------|------|
| 内容为空 | 40049 | 推送消息内容为空 |
| 频率超限 | 40050 | 每日推送次数超限（20条/天） |

**行为说明**：
1. 消息写入 `teacher_messages` 表
2. 同时写入 `conversations` 表（`role='system'`, `sender_type='teacher_push'`），推送到学生聊天页
3. 如 target_type 为 class，遍历班级所有学生，每人写入一条 conversation 记录

---

### 2.14 教师推送历史

**GET** `/api/teacher-messages/history`

**鉴权**：需要（Bearer Token，仅教师角色）

**查询参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| persona_id | int64 | ✅ | - | 教师分身 ID |
| page | int | ❌ | 1 | 页码 |
| page_size | int | ❌ | 20 | 每页数量 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 10
  }
}
```

---

### 2.15 语音转文字

**POST** `/api/speech-to-text`

**鉴权**：需要（Bearer Token）

**请求格式**：`multipart/form-data`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| audio | file | ✅ | 音频文件（微信录音格式 silk/mp3/wav） |

**文件限制**：

| 限制项 | 值 |
|--------|-----|
| 最大时长 | 60秒 |
| 最大文件大小 | 2MB |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "text": "什么是勾股定理",
    "confidence": 0.95
  }
}
```

> **注**：V1版本前端优先使用微信同声传译插件在端侧完成语音转文字，此接口作为备选方案。

---

## 3. 修改接口

### 3.1 对话接口 - Adaptive RAG 增强

**POST** `/api/chat`

**行为变更**：
1. 对话插件新增 Function Calling 能力，LLM 可自主决定是否调用 `web_search` 工具
2. 搜索结果摘要注入上下文后再生成最终回复
3. 响应新增 `tools_used` 字段

**响应变更**：

| 新增字段 | 类型 | 说明 |
|----------|------|------|
| tools_used | []string | 本次对话使用的工具列表（如 `["web_search"]`，未使用则为空数组） |

**响应示例**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "reply": "根据最新研究...",
    "conversation_id": 42,
    "tools_used": ["web_search"]
  }
}
```

---

## 4. 接口变更汇总

### 4.1 新增接口（15 个）

| 编号 | 服务 | 接口 | 方法 | 说明 | 优先级 |
|------|------|------|------|------|--------|
| API-61 | Go | `/api/curriculum-configs` | POST | 创建教材配置 | P0 |
| API-62 | Go | `/api/curriculum-configs` | GET | 查询教材配置 | P0 |
| API-63 | Go | `/api/curriculum-configs/:id` | PUT | 更新教材配置 | P0 |
| API-64 | Go | `/api/curriculum-configs/:id` | DELETE | 删除教材配置 | P0 |
| API-65 | Go | `/api/curriculum-versions` | GET | 教材版本列表 | P0 |
| API-66 | Go | `/api/feedbacks` | POST | 提交反馈 | P0 |
| API-67 | Go | `/api/feedbacks` | GET | 反馈列表 | P0 |
| API-68 | Go | `/api/feedbacks/:id/status` | PUT | 更新反馈状态 | P0 |
| API-69 | Go | `/api/documents/batch-upload` | POST | 批量文件上传 | P1 |
| API-70 | Go | `/api/batch-tasks/:task_id` | GET | 查询批量任务状态 | P1 |
| API-71 | Go | `/api/students/parse-text` | POST | LLM解析学生文本 | P0 |
| API-72 | Go | `/api/students/batch-create` | POST | 批量创建学生 | P0 |
| API-73 | Go | `/api/teacher-messages` | POST | 教师推送消息 | P1 |
| API-74 | Go | `/api/teacher-messages/history` | GET | 推送历史 | P1 |
| API-75 | Go | `/api/speech-to-text` | POST | 语音转文字 | P1 |

### 4.2 修改接口（1 个）

| 编号 | 接口 | 变更 |
|------|------|------|
| API-M9 | `POST /api/chat` | 新增 Adaptive RAG + `tools_used` 响应字段 |

### 4.3 移除接口

| 接口 | 说明 |
|------|------|
| ~~`POST /api/assignments`~~ | 作业提交（已移除） |
| ~~`GET /api/assignments`~~ | 作业列表（已移除） |

### 4.4 数据库变更

| 变更 | 说明 |
|------|------|
| **新增表** `teacher_curriculum_configs` | 教材配置（teacher_id, persona_id, grade_level, grade, textbook_versions, subjects, current_progress, region, is_active） |
| **新增表** `feedbacks` | 用户反馈（user_id, feedback_type, content, status, context_info） |
| **新增表** `batch_tasks` | 批量任务（task_id, persona_id, knowledge_base_id, status, total_files, success_files, failed_files, result_json） |
| **新增表** `teacher_messages` | 教师消息推送（teacher_id, target_type, target_id, content, status） |
| **修改表** `users` | 新增 `profile_snapshot` TEXT 字段（JSON格式用户画像） |
| **移除表** `assignments` | 作业表（如存在则移除） |

### 4.5 内部改造（无接口变更）

| 改造点 | 说明 | 优先级 |
|--------|------|--------|
| `prompt.go` 安全规则 | 最高优先级安全指令，防止Prompt Injection泄露教师评价 | P0 |
| `prompt.go` 学段模板 | 根据 grade_level 从配置文件加载对应学段Prompt模板 | P0 |
| `prompt.go` 教材配置注入 | 教材版本、学科、进度注入System Prompt【教学背景】段落 | P0 |
| `prompt.go` 用户画像注入 | profile_snapshot 注入System Prompt【用户画像】段落 | P0 |
| `prompt.go` 知识库为空行为 | 知识库为空时诚实回答"不知道"，不编造信息 | P0 |
| `memory_summarizer.go` 画像提炼 | 记忆合并时LLM提炼用户画像，覆盖写入 profile_snapshot | P0 |
| `buildMemoryText` 脱敏 | personality_traits 类记忆加 `[内部参考-禁止透露]` 标记 | P0 |
| API限流中间件 | 全局限流 + 对话接口单独限流 | P1 |
| Adaptive RAG | Function Calling框架 + web_search工具 | P1 |

### 4.6 学段Prompt模板配置

> 学段模板以配置文件形式保存在 `configs/grade_level_templates.yaml`，运行时加载。

**配置文件结构**：
```yaml
grade_level_templates:
  preschool:
    name: "学前班"
    description: "幼儿园大班~学前"
    prompt_template: |
      你正在和一个学前班的小朋友对话。请使用：
      - 极简、活泼的语言，多用比喻和故事
      - 游戏化的互动方式
      - 大量鼓励和正面反馈
      - 安全感优先，不要使用可能引起恐惧的表述
      - 每次只讲一个知识点，用具体的例子
  primary_lower:
    name: "小学低年级"
    prompt_template: |
      ...
  # ... 其他6个学段
```

---

**文档版本**: v1.0
**创建日期**: 2026-04-02
**状态**: 已创建
**关联需求**: docs/iterations/v2.0/iteration7_requirements.md