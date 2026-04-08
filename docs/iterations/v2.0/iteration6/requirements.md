# V2.0 迭代6 需求规格说明书

## 1. 迭代概述

| 项目 | 说明 |
|------|------|
| **迭代名称** | V2.0 Sprint 6 - 记忆增强 + 对话风格 + 分享优化 + UI重构 + 生产部署 |
| **迭代目标** | 记忆系统分层改造、对话风格灵活化、分享码二维码+知识库聊天记录导入、教师/学生 TabBar 重设计、Docker Compose 生产部署 |
| **迭代周期** | ~3 周 |
| **交付标准** | 所有新功能通过集成测试 + Docker Compose 一键部署成功 |
| **前置依赖** | V2.0 迭代5 全部完成（161/162 集成测试通过） |

## 2. 迭代目标

### 2.1 核心目标

> **记忆系统分层 + 对话风格灵活化 + 分享码二维码 + 聊天记录导入知识库 + TabBar 重设计 + Docker Compose 生产部署**

具体来说：
1. **记忆系统分层改造（R1）**：memories 表新增 `memory_layer` 字段（`core` / `episodic` / `archived`），核心记忆支持更新覆盖，情景记忆支持摘要合并
2. **对话风格灵活化（R2）**：`StyleConfig` 新增 `teaching_style` 字段，提供多种教学风格模板（苏格拉底式、讲解式、鼓励式、严格式、陪伴式、自定义），去除苏格拉底硬编码
3. **分享码二维码生成（R3）**：前端将分享码转为二维码图片展示，教师可保存分享给学生
4. **扫码落地页优化（R4）**：`GET /api/shares/:code/info` 增加 `join_status` 字段，非目标学生友好引导申请
5. **聊天记录 JSON 导入知识库（R5）**：新增 `POST /api/documents/import-chat`（仅教师），支持上传其他智能体平台导出的 JSON 聊天记录入库
6. **教师/学生 TabBar 重设计（R6）**：自定义 TabBar 组件，教师端（工作台/学生/知识库/我的）、学生端（对话/历史/发现/我的）
7. **Docker Compose 生产部署（R7）**：Go 后端 + Python LlamaIndex 服务 + Nginx 反向代理，一键部署脚本

### 2.2 不在本迭代范围
- ❌ CI/CD 自动化流水线（后续迭代）
- ❌ Prometheus + Grafana 监控（后续迭代）
- ❌ API 限流 / 安全加固（后续迭代）
- ❌ 旧数据迁移脚本（记忆分层对旧数据向后兼容）
- ❌ 微信小程序审核发布（需人工操作）

### 2.3 与迭代5的关系
本迭代在迭代5的基础上，核心升级是**记忆系统架构**（三层分层 + 更新覆盖 + 摘要合并）、**对话风格灵活化**（去除苏格拉底硬编码）、**分享体验优化**（二维码 + 非目标学生引导）、**知识库扩展**（聊天记录导入）、**前端 UI 重构**（自定义 TabBar）和**生产部署**（Docker Compose）。同时清理迭代2~5积累的技术债务。

---

## 3. 问题分析与解决方案

### 3.1 用户反馈的核心问题

| # | 问题 | 当前状态 | 根因 |
|---|------|---------|------|
| 1 | 记忆系统只有单层存储，无法区分核心画像和临时事件 | memories 表无分层字段 | 缺少记忆分层机制 |
| 2 | 对话风格硬编码为苏格拉底式，教师无法选择其他风格 | prompt.go 硬编码苏格拉底 | 缺少风格模板机制 |
| 3 | 分享码只有文本，学生需要手动输入 | 前端只展示文本分享码 | 缺少二维码生成 |
| 4 | 非目标学生扫码后直接报错，体验差 | 返回 40029 错误 | 缺少友好引导 |
| 5 | 知识库只支持手动输入/文件/URL，无法导入聊天记录 | 3种数据源 | 缺少聊天记录导入 |
| 6 | 教师/学生共用 TabBar，功能入口不够精准 | 3 Tab 共用 | 缺少角色自适应 TabBar |
| 7 | 服务无法在生产环境部署 | 只有开发启动脚本 | 缺少 Docker 部署方案 |
| 8 | ListMemories 在应用层做分页和筛选，性能差 | 技术债务 P1 | SQL 层未优化 |

### 3.2 解决方案总览

```
┌──────────────────────────────────────────────────────────────┐
│  迭代6 核心改造                                                │
│                                                                │
│  🧠 记忆系统增强                                                │
│  ├── memories 表新增 memory_layer 字段（core/episodic/archived）│
│  ├── 核心记忆更新覆盖机制（同类 UPDATE 而非 INSERT）             │
│  ├── 情景记忆摘要合并（LLM 压缩为核心记忆）                      │
│  ├── ListMemories SQL 层优化（WHERE + LIMIT/OFFSET + 索引）     │
│  └── 记忆管理 API + 前端（教师可查看/编辑/删除学生记忆）          │
│                                                                │
│  🎨 对话风格灵活化                                              │
│  ├── StyleConfig 新增 teaching_style 字段                       │
│  ├── prompt.go 改为根据 teaching_style 动态选择模板              │
│  └── 前端新增"教学风格"选择器                                    │
│                                                                │
│  🔗 分享与知识沉淀                                              │
│  ├── 前端二维码生成组件                                          │
│  ├── 扫码落地页 + 非目标学生友好引导                              │
│  ├── 聊天记录 JSON 导入知识库 API                                │
│  └── 导入聊天记录前端交互                                        │
│                                                                │
│  📱 UI 重构                                                     │
│  ├── 自定义 TabBar 组件（角色自适应）                             │
│  ├── 教师端 4 Tab（工作台/学生/知识库/我的）                      │
│  └── 学生端 4 Tab（对话/历史/发现/我的）                          │
│                                                                │
│  🚀 生产部署                                                    │
│  ├── Dockerfile.backend（Go 多阶段构建）                         │
│  ├── Dockerfile.knowledge（Python 3.11 + LlamaIndex）           │
│  ├── docker-compose.yml（三服务编排）                             │
│  ├── Nginx 反向代理（HTTPS + SSE 支持）                          │
│  └── 一键部署脚本                                                │
└──────────────────────────────────────────────────────────────┘
```

---

## 4. 架构设计

### 4.1 整体架构（改造后）

```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Compose 生产部署                     │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐     │
│  │              Nginx 反向代理 :80/:443                   │     │
│  │              HTTPS + SSE 支持 + 文件上传 20M           │     │
│  └──────────────────────┬──────────────────────────────┘     │
│                         │ /api/*                              │
│  ┌──────────────────────▼──────────────────────────────┐     │
│  │              Go 后端 :8080                            │     │
│  │                                                       │     │
│  │  API Layer → KnowledgePlugin → VectorClient           │     │
│  │              DialoguePlugin → PromptBuilder            │     │
│  │              MemoryPlugin → 三层记忆（core/episodic）  │     │
│  │              MemorySummarizer → LLM 摘要合并           │     │
│  └──────────────────────┬──────────────────────────────┘     │
│                         │ HTTP                                │
│  ┌──────────────────────▼──────────────────────────────┐     │
│  │         Python LlamaIndex 服务 :8100                  │     │
│  │         FastAPI + DashScope Embedding                  │     │
│  │         SimpleVectorStore (本地 JSON 持久化)           │     │
│  └─────────────────────────────────────────────────────┘     │
│                                                               │
│  微信小程序 ──HTTPS──→ Nginx :443                             │
└─────────────────────────────────────────────────────────────┘
```

### 4.2 记忆系统三层架构

```
┌─────────────────────────────────────────────────────────────┐
│                    记忆系统三层架构                             │
│                                                               │
│  Layer 1: 工作记忆（Working Memory）                          │
│  ├── 存储位置：对话上下文（不落库）                              │
│  ├── 生命周期：单次会话                                        │
│  └── 内容：当前对话的上下文窗口                                 │
│                                                               │
│  Layer 2: 核心记忆（Core Memory）                             │
│  ├── 存储位置：memories 表，memory_layer = 'core'              │
│  ├── 生命周期：长期保留，可更新覆盖                              │
│  ├── 内容：学生画像结论（擅长什么、薄弱什么、学习风格等）         │
│  └── 更新机制：同类记忆 UPDATE 而非 INSERT                     │
│                                                               │
│  Layer 3: 情景记忆（Episodic Memory）                         │
│  ├── 存储位置：memories 表，memory_layer = 'episodic'          │
│  ├── 生命周期：定期摘要压缩为核心记忆                            │
│  ├── 内容：具体学习事件（某次对话中讨论了什么）                   │
│  └── 压缩机制：LLM 将多条 episodic 合并为 1 条 core            │
│                                                               │
│  归档层: 已归档记忆（Archived Memory）                         │
│  ├── 存储位置：memories 表，memory_layer = 'archived'          │
│  ├── 生命周期：永久保留，不参与检索                              │
│  └── 来源：被摘要合并后的原始 episodic 记忆                     │
└─────────────────────────────────────────────────────────────┘
```

### 4.3 对话风格模板架构

```
┌─────────────────────────────────────────────────────────────┐
│                    对话风格模板机制                             │
│                                                               │
│  StyleConfig.teaching_style                                   │
│  ├── "socratic"     → 苏格拉底式提问（现有默认）               │
│  ├── "explanatory"  → 讲解式教学                              │
│  ├── "encouraging"  → 鼓励式教学                              │
│  ├── "strict"       → 严格式教学                              │
│  ├── "companion"    → 陪伴式学习                              │
│  └── "custom"       → 完全由 style_prompt 决定                │
│                                                               │
│  prompt.go 改造：                                             │
│  ├── systemPromptTemplate → 根据 teaching_style 动态选择      │
│  ├── buildStyleText → guidance_level 去掉苏格拉底绑定          │
│  └── 向后兼容：teaching_style 为空时默认 "socratic"            │
└─────────────────────────────────────────────────────────────┘
```

---

## 5. 数据库设计

### 5.1 表结构变更

#### 5.1.1 memories 表新增字段

```sql
ALTER TABLE memories ADD COLUMN memory_layer TEXT NOT NULL DEFAULT 'episodic';
-- 可选值: 'core' / 'episodic' / 'archived'

-- 新增索引（优化 ListMemories 查询性能）
CREATE INDEX IF NOT EXISTS idx_memories_layer ON memories(teacher_persona_id, student_persona_id, memory_layer);
CREATE INDEX IF NOT EXISTS idx_memories_type_layer ON memories(teacher_persona_id, student_persona_id, memory_type, memory_layer);
```

#### 5.1.2 documents 表新增 doc_type 值

`doc_type` 字段新增 `chat` 值，标识来源为聊天记录导入。无需 ALTER TABLE，仅代码层面新增枚举值。

#### 5.1.3 documents 表新增可选字段

```sql
ALTER TABLE documents ADD COLUMN source_session_id TEXT DEFAULT '';
-- 记录聊天记录导入的来源会话ID（可选，便于溯源）
```

### 5.2 Memory 模型变更

```go
// Memory 学生记忆模型（迭代6 改造）
type Memory struct {
    ID               int64      `json:"id"`
    StudentID        int64      `json:"student_id"`
    TeacherID        int64      `json:"teacher_id"`
    TeacherPersonaID int64      `json:"teacher_persona_id"`
    StudentPersonaID int64      `json:"student_persona_id"`
    MemoryType       string     `json:"memory_type"`
    MemoryLayer      string     `json:"memory_layer"`  // 新增: core / episodic / archived
    Content          string     `json:"content"`
    Importance       float64    `json:"importance"`
    LastAccessed     *time.Time `json:"last_accessed,omitempty"`
    CreatedAt        time.Time  `json:"created_at"`
    UpdatedAt        time.Time  `json:"updated_at"`
}
```

### 5.3 StyleConfig 模型变更

```go
// StyleConfig 风格配置（迭代6 改造）
type StyleConfig struct {
    Temperature      float64 `json:"temperature"`
    GuidanceLevel    string  `json:"guidance_level"`    // low / medium / high
    TeachingStyle    string  `json:"teaching_style"`    // 新增: socratic / explanatory / encouraging / strict / companion / custom
    StylePrompt      string  `json:"style_prompt"`
    MaxTurnsPerTopic int     `json:"max_turns_per_topic"`
}
```

---

## 6. 功能需求详细描述

### 6.1 R1 - 记忆系统分层改造

#### 6.1.1 功能描述
将现有的单层记忆系统改造为三层架构（core / episodic / archived），核心记忆支持更新覆盖，情景记忆支持 LLM 摘要合并。

#### 6.1.2 用户故事
- 作为教师，我希望 AI 能区分学生的长期画像（核心记忆）和具体学习事件（情景记忆），提供更精准的个性化教学
- 作为教师，我希望能查看和管理学生的记忆，了解 AI 对学生的认知

#### 6.1.3 子模块

| 子模块 | 内容 | 优先级 |
|--------|------|--------|
| R1-A | memories 表新增 `memory_layer` 字段 + 索引 | P0 |
| R1-B | 核心记忆更新机制：存储时检查同 `memory_type` 的 core 记忆，只保留最近一条（有则 UPDATE，多条时更新最近的） | P0 |
| R1-C | ListMemories SQL 优化：WHERE + LIMIT/OFFSET 替代应用层分页 | P0 |
| R1-D | 记忆价值评估：LLM 判断提取的信息属于 core 还是 episodic | P1 |
| R1-E | 记忆摘要合并：定时任务（每天凌晨自动）+ 手动触发，episodic 超过 50 条时 LLM 压缩为 core，原始标记 archived | P1 |
| R1-F | 记忆管理 API：教师可查看/编辑/删除学生记忆，触发摘要合并 | P1 |

#### 6.1.4 技术要求

1. **记忆存储改造**：`memory_plugin.go` 中 `handleStore` 方法改造
   - 新增 LLM 调用判断记忆层级（core / episodic）
   - core 记忆：先查询是否已有同 `memory_type` 的 core 记忆，只保留最近一条（有则 UPDATE content + updated_at，多条时更新 `ORDER BY updated_at DESC LIMIT 1` 的那条），无则 INSERT
   - episodic 记忆：直接 INSERT
   - episodic 上限：每个学生分身的 episodic 记忆上限 50 条

2. **ListMemories SQL 优化**：
   - 现有问题：应用层先查全部再分页筛选
   - 改造为：`SELECT ... WHERE memory_layer IN ('core','episodic') AND teacher_persona_id=? AND student_persona_id=? ORDER BY importance DESC, updated_at DESC LIMIT ? OFFSET ?`

3. **记忆摘要合并**：
   - 新增 `POST /api/memories/summarize` 接口（教师手动触发）
   - 新增定时任务（每天凌晨 2:00 自动执行），扫描所有学生的 episodic 记忆，超过 50 条的自动触发摘要合并
   - 查询指定学生的所有 episodic 记忆
   - 调用 LLM 将多条 episodic 压缩为 1~3 条 core 记忆
   - 原始 episodic 记忆标记为 `archived`

4. **记忆管理 API**：
   - `GET /api/memories` 增加 `layer` 查询参数
   - `PUT /api/memories/:id` 教师编辑记忆内容
   - `DELETE /api/memories/:id` 教师删除记忆

#### 6.1.5 向后兼容
- 旧记忆数据 `memory_layer` 默认为 `episodic`（通过 DEFAULT 约束）
- 不需要数据迁移脚本

---

### 6.2 R2 - 对话风格灵活化

#### 6.2.1 功能描述
去除系统提示词中苏格拉底式教学法的硬编码，引入"教学风格模板"机制，允许教师为每个学生选择不同的教学风格。

#### 6.2.2 用户故事
- 作为教师，我希望能为不同学生选择不同的教学风格，而不是所有学生都用苏格拉底式提问

#### 6.2.3 风格模板定义

| 风格 ID | 名称 | 系统提示词基底 |
|---------|------|---------------|
| `socratic` | 苏格拉底式提问 | 不直接给答案，通过提问引导思考（现有默认） |
| `explanatory` | 讲解式教学 | 详细讲解知识点，配合举例说明 |
| `encouraging` | 鼓励式教学 | 多用肯定语言，循序渐进引导 |
| `strict` | 严格式教学 | 严格要求，注重准确性和规范性 |
| `companion` | 陪伴式学习 | 像朋友一样陪伴学习，轻松氛围 |
| `custom` | 自定义 | 完全由教师的 `style_prompt` 决定 |

#### 6.2.4 技术要求

1. **后端 prompt.go 改造**：
   - `systemPromptTemplate` 常量改为 `teachingStyleTemplates` map
   - `BuildSystemPrompt` 根据 `styleConfig.TeachingStyle` 选择对应模板
   - `buildStyleText` 中 `guidance_level` 描述去掉苏格拉底绑定
   - 向后兼容：`TeachingStyle` 为空时默认 `socratic`

2. **后端 StyleConfig 改造**：
   - `database/models.go` 中 `StyleConfig` 新增 `TeachingStyle` 字段
   - `dialogue/prompt.go` 中本地 `StyleConfig` 同步新增

3. **前端改造**：
   - `student-detail/index.tsx` 新增"教学风格"Picker 选择器
   - `api/style.ts` 中 `StyleConfig` 接口新增 `teaching_style` 字段
   - `login/index.tsx` slogan 从"基于苏格拉底式教学的 AI 数字分身"改为通用描述
   - `student-detail/index.tsx` 引导程度标签去掉"苏格拉底"字样

---

### 6.3 R3 - 分享码二维码生成

#### 6.3.1 功能描述
教师在分享码管理页面，可以将分享码生成为二维码图片，方便学生扫码加入。

#### 6.3.2 用户故事
- 作为教师，我希望能生成分享码的二维码图片，发给学生扫码即可加入

#### 6.3.3 技术要求

1. **前端引入二维码库**：使用 `weapp-qrcode` 或 Taro 兼容的二维码生成库
2. **二维码内容格式**：`https://{domain}/join/{share_code}` 或小程序码路径 `pages/share-join/index?code={share_code}`
3. **展示位置**：分享码管理页（`share-manage/index.tsx`）每个分享码卡片新增"显示二维码"按钮
4. **支持保存**：用户可长按保存二维码图片到相册
5. **后端无改动**：纯前端功能

---

### 6.4 R4 - 扫码落地页优化

#### 6.4.1 功能描述
优化分享码加入流程，对非目标学生提供友好引导（提示申请），而非直接报错。

#### 6.4.2 用户故事
- 作为非目标学生，我扫码后希望看到友好的提示和申请入口，而不是冷冰冰的错误信息

#### 6.4.3 交互流程

```
学生扫码 → GET /api/shares/:code/info
     │
     ▼
展示教师分身信息（头像、昵称、学校、简介）
     │
     ├── join_status = "can_join"      → 显示"加入"按钮 → POST join → 成功
     ├── join_status = "already_joined" → 显示"已加入"提示
     ├── join_status = "not_target"    → 显示"该邀请码是老师专门发给特定同学的，你可以向老师申请"
     │                                   → 显示"向老师申请"按钮 → POST /api/relations/apply
     ├── join_status = "need_login"    → 提示"请先登录" → 跳转登录页
     └── join_status = "need_persona"  → 提示"请先创建学生分身" → 跳转创建分身页
```

#### 6.4.4 技术要求

1. **后端 `GET /api/shares/:code/info` 增强**：
   - 新增 `join_status` 字段（需要登录态，无登录态返回 `need_login`）
   - 可选值：`can_join` / `already_joined` / `not_target` / `need_login` / `need_persona`

2. **后端 `POST /api/shares/:code/join` 改造**：
   - 非目标学生不再返回 `40029` 错误
   - 改为返回引导信息：`{ "join_status": "not_target", "can_apply": true, "message": "..." }`

3. **前端 `share-join/index.tsx` 改造**：
   - 根据 `join_status` 展示不同 UI
   - `not_target` 状态显示"向老师申请"按钮，调用 `POST /api/relations` 发起申请

---

### 6.5 R5 - 聊天记录 JSON 导入知识库

#### 6.5.1 功能描述
教师可以上传其他智能体平台导出的 JSON 格式聊天记录，系统自动解析为 Q&A 文本并入库。仅教师可用，学生不支持此功能。

#### 6.5.2 用户故事
- 作为教师，我希望能将学生在其他 AI 平台的聊天记录导入知识库，丰富教学素材

#### 6.5.3 支持的 JSON 格式

```json
// 格式1: OpenAI 风格
{
  "messages": [
    {"role": "user", "content": "什么是二叉树？"},
    {"role": "assistant", "content": "二叉树是一种树形数据结构..."}
  ]
}

// 格式2: 带时间戳
{
  "conversations": [
    {"sender": "学生", "text": "老师，这道题怎么做？", "time": "2026-03-20 10:00"},
    {"sender": "AI", "text": "我们来分析一下...", "time": "2026-03-20 10:01"}
  ]
}

// 格式3: 通用数组
[
  {"role": "user", "content": "..."},
  {"role": "assistant", "content": "..."}
]
```

#### 6.5.4 技术要求

1. **新增 API**：`POST /api/documents/import-chat`（仅教师，multipart/form-data）
2. **新增解析器**：`ChatJSONParser`，自动识别常见 JSON 聊天格式
3. **解析策略**：提取对话对，拼接为结构化 Q&A 文本
4. **入库流程**：复用现有 knowledge 插件的 add action
5. **doc_type**：记为 `chat`
6. **前端**：知识库管理页新增"导入聊天记录"入口（仅教师端显示）

#### 6.5.5 前端交互

```
知识库管理页（教师端）：
┌─────────────────────────────────────────┐
│  ➕ 添加知识                             │
│  ├── 📝 手动输入                         │
│  ├── 📄 上传文件（PDF/DOCX/TXT/MD）      │
│  ├── 🔗 导入网页                         │
│  └── 💬 导入聊天记录（JSON）  ← 新增      │
└─────────────────────────────────────────┘

点击"导入聊天记录"后：
┌─────────────────────────────────────────┐
│  导入聊天记录                            │
│                                          │
│  📎 选择 JSON 文件    [选择文件]          │
│                                          │
│  ── 解析预览 ──                          │
│  Q: 什么是二叉树？                       │
│  A: 二叉树是一种树形数据结构...           │
│  （共 12 轮对话）                         │
│                                          │
│  标题：[关于二叉树的问答        ]         │
│  标签：[二叉树,数据结构          ]        │
│  范围：[全局 ▼]  [选择班级/学生]          │
│                                          │
│         [取消]  [确认导入]                │
└─────────────────────────────────────────┘
```

---

### 6.6 R6 - 教师/学生 TabBar 重设计

#### 6.6.1 功能描述
使用自定义 TabBar 组件替代微信小程序原生 TabBar，实现教师和学生展示不同的 Tab 布局。

#### 6.6.2 用户故事
- 作为教师，我希望底部 Tab 直接展示我最常用的功能入口（工作台、学生管理、知识库、我的）
- 作为学生，我希望底部 Tab 直接展示对话、历史、发现、我的

#### 6.6.3 TabBar 布局设计

**教师端 4 Tab：**

| Tab | 图标 | 页面 | 功能 |
|-----|------|------|------|
| 工作台 | 🏠 | home | 统计数据 + 班级列表 + 待审批提醒 + 分享码 |
| 学生 | 👥 | teacher-students | 学生管理 + 审批（带角标显示待审批数） |
| 知识库 | 📚 | knowledge | 知识库管理（教师高频操作，提升为一级入口） |
| 我的 | 👤 | profile | 分身概览 + 作业管理 + 设置 + 退出 |

**学生端 4 Tab：**

| Tab | 图标 | 页面 | 功能 |
|-----|------|------|------|
| 对话 | 💬 | home | 我的老师列表 + 开始对话 |
| 历史 | 📋 | history | 对话历史记录 |
| 发现 | 🌐 | discover | 教师广场 + 分享码加入 |
| 我的 | 👤 | profile | 我的作业 + 我的记忆 + 设置 + 退出 |

#### 6.6.4 视觉规范

| 属性 | 值 |
|------|-----|
| 未选中颜色 | `#999999` |
| 选中颜色（品牌色） | `#4F46E5` |
| 背景色 | `#FFFFFF` |
| 顶部分隔线 | `1px solid #F0F0F0` |
| 角标颜色 | `#FF4D4F`（红色小圆点/数字） |

#### 6.6.5 技术要求

1. **自定义 TabBar 组件**：`components/CustomTabBar/index.tsx`
   - 根据用户角色（teacher/student）展示不同 Tab 列表
   - 支持角标（Badge）显示待审批数量
   - 使用 `Taro.switchTab` 或 `Taro.redirectTo` 切换页面

2. **app.config.ts 改造**：
   - `tabBar.custom: true`（启用自定义 TabBar）
   - `tabBar.list` 保留所有可能的 Tab 页面（微信要求）

3. **页面路由调整**：
   - 知识库页面（`knowledge/index`）提升为 Tab 页
   - 发现页面（`discover/index`）提升为 Tab 页

---

### 6.7 R7 - Docker Compose 生产部署

#### 6.7.1 功能描述
提供 Docker Compose 一键部署方案，包含 Go 后端、Python LlamaIndex 服务和 Nginx 反向代理。

#### 6.7.2 用户故事
- 作为开发者，我希望在一台服务器上一条命令就能启动所有服务

#### 6.7.3 服务编排

| 服务 | 镜像 | 端口 | 依赖 |
|------|------|------|------|
| knowledge | Dockerfile.knowledge (Python 3.11 + LlamaIndex) | 8100 | 无 |
| backend | Dockerfile.backend (Go 1.25 + SQLite) | 8080 | knowledge (healthy) |
| nginx | nginx:alpine | 80, 443 | backend |

#### 6.7.4 文件清单

```
deployments/
├── docker/
│   ├── Dockerfile.backend       # Go 后端多阶段构建
│   └── Dockerfile.knowledge     # Python LlamaIndex 服务
├── nginx/
│   ├── nginx.conf               # Nginx 配置（HTTPS + SSE）
│   └── ssl/                     # SSL 证书目录
├── scripts/
│   ├── deploy.sh                # 一键部署脚本
│   └── backup.sh                # 数据备份脚本
├── docker-compose.yml           # 服务编排
└── .env.production              # 生产环境变量模板
```

#### 6.7.5 技术要求

1. **Dockerfile.backend**：
   - 多阶段构建（golang:1.25-alpine → alpine:3.20）
   - CGO_ENABLED=1（SQLite 依赖）
   - 健康检查：`curl -f http://localhost:8080/health`

2. **Dockerfile.knowledge**：
   - 基础镜像 python:3.11-slim
   - 预下载 NLTK 数据（LlamaIndex SentenceSplitter 依赖）
   - 内存限制 2G（LlamaIndex + numpy 较吃内存）
   - 健康检查：`curl -f http://localhost:8100/api/v1/health`

3. **docker-compose.yml**：
   - 服务启动顺序：knowledge → backend → nginx
   - Volume 持久化：SQLite 数据库、上传文件、向量索引
   - 环境变量通过 `.env` 文件注入

4. **Nginx 配置**：
   - HTTP → HTTPS 重定向
   - SSE 流式响应支持（`proxy_buffering off`）
   - 文件上传限制 20M（`client_max_body_size 20M`）

5. **一键部署脚本**：
   - 检查 Docker 环境
   - 检查 `.env` 必要变量（JWT_SECRET、OPENAI_API_KEY）
   - 构建镜像 + 启动服务 + 等待就绪

#### 6.7.6 服务器配置建议

| 配置项 | 最低要求 | 推荐配置 |
|--------|----------|----------|
| CPU | 2 核 | 4 核 |
| 内存 | 4G | 8G |
| 磁盘 | 20G | 50G |
| 网络 | 需要访问外网（DashScope API） | — |

---

## 7. 模块划分

### 7.1 后端模块划分

| 模块编号 | 模块名称 | 优先级 | 依赖 | 说明 |
|----------|----------|--------|------|------|
| V2-IT6-BE-M1 | 数据库变更（记忆分层） | P0-第1层 | 无 | memories 表新增 memory_layer + 索引 |
| V2-IT6-BE-M2 | 记忆存储改造 | P0-第2层 | M1 | 核心记忆更新覆盖 + LLM 层级判断 |
| V2-IT6-BE-M3 | ListMemories SQL 优化 | P0-第2层 | M1 | SQL 层 WHERE + LIMIT/OFFSET |
| V2-IT6-BE-M4 | 记忆管理 API | P1-第2层 | M1 | 编辑/删除/摘要合并接口 |
| V2-IT6-BE-M5 | 对话风格模板改造 | P0-第2层 | 无 | prompt.go 动态模板 + StyleConfig 扩展 |
| V2-IT6-BE-M6 | 分享码信息增强 | P0-第2层 | 无 | join_status 字段 + 非目标学生引导 |
| V2-IT6-BE-M7 | 聊天记录导入 API | P1-第2层 | 无 | ChatJSONParser + import-chat 接口 |
| V2-IT6-BE-M8 | 部署配置文件 | P0-第1层 | 无 | Dockerfile + docker-compose + Nginx |

**开发顺序：**
```
第1层（并行）:
  ├── V2-IT6-BE-M1 数据库变更
  └── V2-IT6-BE-M8 部署配置文件

第2层（并行）:
  ├── V2-IT6-BE-M2 记忆存储改造
  ├── V2-IT6-BE-M3 ListMemories SQL 优化
  ├── V2-IT6-BE-M4 记忆管理 API
  ├── V2-IT6-BE-M5 对话风格模板改造
  ├── V2-IT6-BE-M6 分享码信息增强
  └── V2-IT6-BE-M7 聊天记录导入 API
```

### 7.2 前端模块划分

| 模块编号 | 模块名称 | 优先级 | 涉及页面 | 对应后端接口 |
|----------|----------|--------|----------|-------------|
| V2-IT6-FE-M1 | 自定义 TabBar 组件 | P0 | CustomTabBar 组件 + app.config.ts | 无 |
| V2-IT6-FE-M2 | 教师端 TabBar 适配 | P0 | home + teacher-students + knowledge + profile | 已有接口 |
| V2-IT6-FE-M3 | 学生端 TabBar 适配 | P0 | home + history + discover + profile | 已有接口 |
| V2-IT6-FE-M4 | 对话风格选择器 | P0 | student-detail + login | style API |
| V2-IT6-FE-M5 | 分享码二维码 | P1 | share-manage | 无（纯前端） |
| V2-IT6-FE-M6 | 扫码落地页优化 | P1 | share-join | shares info API |
| V2-IT6-FE-M7 | 聊天记录导入 UI | P1 | knowledge/add | import-chat API |
| V2-IT6-FE-M8 | 记忆管理 UI | P1 | 新增 memory-manage 页面 | memories API |

**开发顺序：**
```
8 个前端模块之间无强依赖，基于接口文档直接并行开发
建议优先完成 FE-M1（TabBar 组件），其他模块依赖 TabBar 框架
```

---

## 8. 集成测试用例规划

> **分批执行规则**：每批 ≤ 5 个用例，ci_test_agent 以 API 规范文档为准编写测试。

| 用例编号 | 测试场景 | 涉及模块 | 数据依赖 |
|----------|----------|----------|----------|
| IT-301 | 记忆存储 - episodic 层级 | BE-M1, M2 | 无 |
| IT-302 | 记忆存储 - core 层级 + 更新覆盖 | BE-M1, M2 | IT-301 |
| IT-303 | ListMemories 按 layer 筛选 | BE-M1, M3 | IT-301, IT-302 |
| IT-304 | 记忆编辑 + 删除 | BE-M4 | IT-301 |
| IT-305 | 记忆摘要合并 | BE-M4 | IT-301 |
| IT-306 | 对话风格 - socratic 模板 | BE-M5 | 无 |
| IT-307 | 对话风格 - explanatory 模板 | BE-M5 | 无 |
| IT-308 | 对话风格 - custom 模板 | BE-M5 | 无 |
| IT-309 | 对话风格 - 向后兼容（teaching_style 为空） | BE-M5 | 无 |
| IT-310 | 分享码信息 - join_status = can_join | BE-M6 | 无 |
| IT-311 | 分享码信息 - join_status = not_target | BE-M6 | 无 |
| IT-312 | 分享码信息 - join_status = already_joined | BE-M6 | IT-310 |
| IT-313 | 分享码加入 - 非目标学生友好引导 | BE-M6 | 无 |
| IT-314 | 聊天记录导入 - OpenAI 格式 JSON | BE-M7 | 无 |
| IT-315 | 聊天记录导入 - 带时间戳格式 JSON | BE-M7 | 无 |
| IT-316 | 聊天记录导入 - 非教师角色拒绝 | BE-M7 | 无 |
| IT-317 | 全链路：教师设置风格 → 学生对话 → 记忆分层存储 → 教师查看记忆 | 全部 | 无 |

**推荐分批方案：**
- 第 1 批（记忆系统）：IT-301 ~ IT-305
- 第 2 批（对话风格）：IT-306 ~ IT-309
- 第 3 批（分享优化）：IT-310 ~ IT-313
- 第 4 批（知识库 + 全链路）：IT-314 ~ IT-317

---

## 9. 前端页面结构（最终版）

### 9.1 教师端

```
🏠 工作台 (Tab 1)
├── 📊 数据统计卡片（学生数/文档数/班级数/对话数）
├── ⏳ 待审批提醒
├── 📋 我的班级列表
└── 🔗 最新分享码

👥 学生 (Tab 2)
├── [全部学生] Tab
├── [按班级] Tab
├── [待审批] Tab（带角标）
├── [班级设置] Tab
└── 点击学生 → 学生详情（风格设置/备注/对话记录）
    └── 🎨 教学风格选择器（新增）

📚 知识库 (Tab 3)
├── 文档列表（scope 筛选）
├── ➕ 添加知识
│   ├── 📝 手动输入
│   ├── 📄 上传文件
│   ├── 🔗 导入网页
│   └── 💬 导入聊天记录（新增）
└── 文档预览

👤 我的 (Tab 4)
├── 个人信息 + 统计
├── 🧑‍🏫 分身概览
├── 📝 作业管理
├── 🔗 分享管理（含二维码）
├── 🧠 学生记忆管理（新增）
├── ℹ️ 关于系统
└── 🚪 退出登录
```

### 9.2 学生端

```
💬 对话 (Tab 1)
├── 我的老师列表 + 开始对话
├── 加入班级入口
└── 1个老师时直接进对话

📋 历史 (Tab 2)
└── 对话历史记录

🌐 发现 (Tab 3)
├── 教师广场搜索
├── 分享码加入（扫码落地页优化）
└── 推荐教师

👤 我的 (Tab 4)
├── 个人信息 + 统计
├── 📝 我的作业
├── 🧠 我的记忆
├── ℹ️ 关于系统
└── 🚪 退出登录
```

---

## 10. 接口变更总览

### 10.1 新增接口

| 编号 | 接口 | 方法 | 说明 | 所属模块 |
|------|------|------|------|----------|
| API-57 | `/api/memories/summarize` | POST | 记忆摘要合并（教师触发） | BE-M4 |
| API-58 | `/api/memories/:id` | PUT | 编辑记忆内容 | BE-M4 |
| API-59 | `/api/memories/:id` | DELETE | 删除记忆 | BE-M4 |
| API-60 | `/api/documents/import-chat` | POST | 聊天记录 JSON 导入知识库 | BE-M7 |

### 10.2 修改接口

| 编号 | 接口 | 变更内容 | 所属模块 |
|------|------|----------|----------|
| API-M4 | `GET /api/memories` | 新增 `layer` 查询参数 | BE-M3 |
| API-M5 | `GET /api/shares/:code/info` | 新增 `join_status` 字段 | BE-M6 |
| API-M6 | `POST /api/shares/:code/join` | 非目标学生返回引导而非报错 | BE-M6 |
| API-M7 | `PUT /api/styles` | StyleConfig 新增 `teaching_style` 字段 | BE-M5 |
| API-M8 | `GET /api/styles` | StyleConfig 新增 `teaching_style` 字段 | BE-M5 |

### 10.3 内部改造（无接口变更）

| 改造点 | 说明 |
|--------|------|
| `memory_plugin.go` handleStore | 新增 LLM 层级判断 + core 记忆更新覆盖 |
| `memory_plugin.go` handleList | SQL 层 WHERE + LIMIT/OFFSET |
| `prompt.go` systemPromptTemplate | 改为 teachingStyleTemplates map |
| `prompt.go` buildStyleText | guidance_level 去掉苏格拉底绑定 |

---

## 11. 关键场景操作路径对比

| 场景 | 迭代5（当前） | 迭代6（目标） | 变化 |
|------|-------------|-------------|------|
| 教师底部 Tab | 首页/历史/我的（3 Tab 共用） | **工作台/学生/知识库/我的**（4 Tab 专属） | 功能入口更精准 |
| 学生底部 Tab | 首页/历史/我的（3 Tab 共用） | **对话/历史/发现/我的**（4 Tab 专属） | 发现升级为一级入口 |
| 教师设置教学风格 | 只能调引导程度（绑定苏格拉底） | **选择教学风格模板**（6种可选） | 灵活度大幅提升 |
| 教师分享码 | 只有文本码 | **二维码图片**（可保存分享） | 分享更便捷 |
| 非目标学生扫码 | 报错 40029 | **友好引导 + 申请按钮** | 体验大幅改善 |
| 教师导入知识 | 手动/文件/URL | 新增**聊天记录 JSON 导入** | 数据源更丰富 |
| 记忆存储 | 单层存储，无区分 | **三层分层**（core/episodic/archived） | 记忆更精准 |
| 记忆管理 | 无管理入口 | 教师可**查看/编辑/删除/合并** | 可控性提升 |
| 生产部署 | 只有开发脚本 | **Docker Compose 一键部署** | 可上线 |

---

## 12. 技术债务清理

本迭代顺带清理以下技术债务：

| 来源 | 描述 | 关联模块 | 状态 |
|------|------|----------|------|
| 迭代2-P1 | ListMemories 应用层分页 → SQL 层 | R1-C 直接解决 | 待修复 |
| 迭代5-P1 | Python 环境升级到 3.10+ | R7 Docker 镜像用 3.11 | 待修复 |
| 迭代5-MED-2 | TeacherDashboard 缺少统计卡片 | R6 TabBar 重构时顺带补全 | 待修复 |
| 迭代5-MED-5 | HandleChat 附件路径安全校验 | 可顺带修复 | 待修复 |
| 迭代5-P3 | 清理 my-comments 页面和路由 | R6 TabBar 重构时清理 | 待修复 |

---

## 13. 工作量估算

| 主题 | 后端工作量 | 前端工作量 | 总计 |
|------|-----------|-----------|------|
| R1. 记忆系统增强 | 中~大 | 中 | **大** |
| R2. 对话风格灵活化 | 小 | 小 | **小** |
| R3. 分享码二维码 | 无 | 小 | **小** |
| R4. 扫码落地页优化 | 小 | 小 | **小** |
| R5. 聊天记录导入 | 小~中 | 中 | **中** |
| R6. TabBar 重设计 | 无 | 中 | **中** |
| R7. Docker 部署 | 小（配置文件） | 无 | **小** |
| **总计** | **中** | **中~大** | **大** |
