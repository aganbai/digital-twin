# V2.0 需求规格说明书

## 1. 版本概述

| 项目 | 说明 |
|------|------|
| **版本名称** | V2.0 - 单机生产可用版 |
| **版本目标** | 功能完整 + 单机部署 + 生产可用 |
| **迭代数** | 3 个迭代 |
| **预计周期** | ~11 周 |
| **前置依赖** | V1.0 全部完成（3 个迭代，39 个集成测试通过） |

### 1.1 V1.0 已完成内容回顾

| V1.0 迭代 | 内容 | 状态 |
|------------|------|------|
| 迭代1 | Harness 框架 + 4 个核心插件（认证/知识库/记忆/对话） + HTTP API | ✅ |
| 迭代2 | Taro 小程序前端 10 页面 + 后端微信登录适配 | ✅ |
| 迭代3 | 管道编排落地 + 记忆自动提取 + 令牌刷新 + 配置校验 + 超时控制 | ✅ |

### 1.2 V2.0 目标

> **单机部署、生产可用、功能完整** —— 一台服务器跑起来，能真正给师生用。

### 1.3 明确排除

| 事项 | 原因 |
|------|------|
| 多机部署 / K8s / 自动扩缩容 | 保持单机部署 |
| 多租户支持 | 企业级特性 |
| 插件市场 / API 开放平台 | 企业级特性 |
| 留言系统 | 优先级低，可后续补充 |
| 智能推荐 | 优先级低，可后续补充 |

---

## 2. 迭代规划总览

```
迭代1: 核心功能开发（用户需求优先）     ~4 周
迭代2: 多角色多分身架构                ~4 周
迭代3: 生产就绪与上线                  ~3 周
                              总计 ~11 周
```

| 迭代 | 主题 | 核心内容 |
|------|------|----------|
| 迭代1 | 核心功能开发 | 师生授权 + 注册增强 + 评语 + 问答风格 + 作业系统 + 文件上传 + URL 导入 + SSE 流式输出 |
| 迭代2 | 多角色多分身架构 | 多分身体系 + 班级管理 + 分身分享 + 知识库精细化 + 现有模块分身化改造 |
| 迭代3 | UI 重构与体验优化 | 教师/学生仪表盘首页 + 班级详情页 + 启停管理 + 知识库 scope 多选 + 上传预览 + 分享码入口优化 |

---

## 3. 迭代1：核心功能开发（~4 周）

> **目标**：完成所有用户可感知的新功能，让系统从 MVP 变成功能完整的产品

### 3.1 师生关系与教学管理（🆕 用户需求）

#### 3.1.1 数据库变更

**新增表**：

```sql
-- 师生授权关系表
CREATE TABLE IF NOT EXISTS teacher_student_relations (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    teacher_id      INTEGER NOT NULL,
    student_id      INTEGER NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',  -- pending/approved/rejected
    initiated_by    TEXT NOT NULL,                     -- teacher(邀请) / student(申请)
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (teacher_id) REFERENCES users(id),
    FOREIGN KEY (student_id) REFERENCES users(id),
    UNIQUE(teacher_id, student_id)
);

-- 教师评语表
CREATE TABLE IF NOT EXISTS teacher_comments (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    teacher_id      INTEGER NOT NULL,
    student_id      INTEGER NOT NULL,
    content         TEXT NOT NULL,           -- 评语内容
    progress_summary TEXT,                   -- 学习进度摘要
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (teacher_id) REFERENCES users(id),
    FOREIGN KEY (student_id) REFERENCES users(id)
);

-- 学生问答风格配置表
CREATE TABLE IF NOT EXISTS student_dialogue_styles (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    teacher_id      INTEGER NOT NULL,
    student_id      INTEGER NOT NULL,
    style_config    TEXT NOT NULL,           -- JSON 格式配置
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (teacher_id) REFERENCES users(id),
    FOREIGN KEY (student_id) REFERENCES users(id),
    UNIQUE(teacher_id, student_id)
);

-- 学生作业/成果表
CREATE TABLE IF NOT EXISTS assignments (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id      INTEGER NOT NULL,
    teacher_id      INTEGER NOT NULL,
    title           TEXT NOT NULL,
    content         TEXT,                    -- 文本内容
    file_path       TEXT,                    -- 上传文件路径
    file_type       TEXT,                    -- 文件类型
    status          TEXT DEFAULT 'submitted', -- submitted/reviewed
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES users(id),
    FOREIGN KEY (teacher_id) REFERENCES users(id)
);

-- 作业点评表（AI 和教师共用）
CREATE TABLE IF NOT EXISTS assignment_reviews (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    assignment_id   INTEGER NOT NULL,
    reviewer_type   TEXT NOT NULL,            -- ai / teacher
    reviewer_id     INTEGER,                  -- teacher 时为教师 ID，ai 时为 NULL
    content         TEXT NOT NULL,            -- 点评内容
    score           REAL,                     -- 评分（可选，0-100）
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (assignment_id) REFERENCES assignments(id)
);
```

**users 表变更**：

```sql
ALTER TABLE users ADD COLUMN school TEXT DEFAULT '';          -- 学校名称（教师必填）
ALTER TABLE users ADD COLUMN description TEXT DEFAULT '';     -- 分身简短描述（教师必填）
-- 教师 nickname + school 联合唯一索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_teacher_school ON users(nickname, school) WHERE role = 'teacher';
```

#### 3.1.2 新增接口

**师生授权**：

| 方法 | 路径 | 说明 | 角色 |
|------|------|------|------|
| POST | `/api/relations/invite` | 教师邀请学生 | teacher |
| POST | `/api/relations/apply` | 学生申请使用分身 | student |
| PUT | `/api/relations/:id/approve` | 教师审批同意 | teacher |
| PUT | `/api/relations/:id/reject` | 教师审批拒绝 | teacher |
| GET | `/api/relations` | 获取师生关系列表 | 所有 |

**教师评语**：

| 方法 | 路径 | 说明 | 角色 |
|------|------|------|------|
| POST | `/api/comments` | 教师写评语 | teacher |
| GET | `/api/comments` | 获取评语列表 | 所有 |

**个性化问答风格**：

| 方法 | 路径 | 说明 | 角色 |
|------|------|------|------|
| PUT | `/api/students/:id/dialogue-style` | 设置学生问答风格 | teacher |
| GET | `/api/students/:id/dialogue-style` | 获取学生问答风格 | teacher/student |

**作业系统**：

| 方法 | 路径 | 说明 | 角色 |
|------|------|------|------|
| POST | `/api/assignments` | 学生提交作业 | student |
| GET | `/api/assignments` | 获取作业列表 | 所有 |
| GET | `/api/assignments/:id` | 获取作业详情 | 所有 |
| POST | `/api/assignments/:id/review` | 教师点评作业 | teacher |
| POST | `/api/assignments/:id/ai-review` | AI 自动点评 | student/teacher |

#### 3.1.3 改造接口

| 接口 | 改造内容 |
|------|----------|
| `POST /api/auth/complete-profile` | 教师注册必填 school + description，校验 nickname+school 唯一 |
| `POST /api/chat` | 增加师生授权鉴权（未授权返回 403）+ 注入个性化问答风格 |

#### 3.1.4 核心业务逻辑

**师生授权机制**：
```
教师邀请学生:
  POST /api/relations/invite {student_id}
  → 创建关系 status=approved, initiated_by=teacher（邀请即同意）

学生申请使用分身:
  POST /api/relations/apply {teacher_id}
  → 创建关系 status=pending, initiated_by=student

教师审批:
  PUT /api/relations/:id/approve → status=approved
  PUT /api/relations/:id/reject  → status=rejected

对话鉴权（改造 POST /api/chat）:
  学生发起对话 → 检查 teacher_student_relations
  → status != 'approved' → 403 "未获得该教师授权，请先申请"
```

**教师注册唯一性校验**：
```
POST /api/auth/complete-profile:
  if role == "teacher":
    必填: nickname, school, description
    校验: SELECT COUNT(*) FROM users WHERE nickname=? AND school=? AND role='teacher'
    → 已存在 → 409 "该学校已有同名教师，请修改名称"
```

**个性化问答风格**（`style_config` JSON 结构）：
```json
{
  "temperature": 0.7,           // 回复随机性 0.1-1.0
  "guidance_level": "medium",   // 引导程度: low/medium/high
  "style_prompt": "对该学生请多用鼓励性语言，注重基础概念的巩固",
  "max_turns_per_topic": 5      // 每个话题最大追问轮次
}
```

对话时，`socratic-dialogue` 插件先查询 `student_dialogue_styles`，将 `style_config` 注入系统提示词。

**AI 点评逻辑**：
```
学生提交作业 → 可触发 AI 点评:
  1. 读取作业内容（文本 or 解析文件）
  2. 检索该教师知识库中的相关知识
  3. 构建 prompt: "你是{教师名}的数字分身，请根据以下知识库内容，对学生的作业进行点评..."
  4. 调用大模型生成点评
  5. 存入 assignment_reviews (reviewer_type='ai')

教师也可手动点评:
  POST /api/assignments/:id/review → 存入 assignment_reviews (reviewer_type='teacher')
```

---

### 3.2 知识库增强

| 事项 | 说明 |
|------|------|
| 文件上传（PDF/DOCX/TXT/MD） | 文件接收 → 内容解析 → 分块 → 向量化存储 |
| URL 页面导入 | 网页抓取 → HTML 解析 → 正文提取 → 存储 |

**新增接口**：

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/documents/upload` | 文件上传（multipart/form-data） |
| POST | `/api/documents/import-url` | URL 导入 |

**文件上传规格**：
- 支持格式：PDF / DOCX / TXT / MD
- 单文件最大：50MB
- 存储路径：`uploads/documents/{teacher_id}/{filename}`
- 解析后自动分块（chunk_size=1000, overlap=200）→ 向量化存入 Chroma

**URL 导入规格**：
- 后端抓取网页内容（HTTP GET + User-Agent）
- HTML 解析：去除标签、提取正文、处理编码
- 异常处理：URL 不可达、内容为空、超时（10s）、内容过长（截断 100000 字符）
- 解析后内容自动填充标题和正文，教师可编辑后提交

---

### 3.3 对话增强

| 事项 | 说明 |
|------|------|
| 流式输出（SSE） | 大模型流式调用 → SSE 推送 → 前端逐字渲染 |

**改造接口**：

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/chat/stream` | SSE 流式对话（新增，与 `/api/chat` 并存） |

**SSE 响应格式**：
```
data: {"type": "start", "session_id": "uuid"}

data: {"type": "delta", "content": "这是"}

data: {"type": "delta", "content": "一个"}

data: {"type": "delta", "content": "很好的问题"}

data: {"type": "done", "conversation_id": 42, "token_usage": {"prompt_tokens": 850, "completion_tokens": 120, "total_tokens": 970}}
```

---

### 3.4 前端页面变更

#### 改造页面

| 页面 | 改造内容 |
|------|----------|
| 角色选择页 FE-P2 | 教师注册增加"学校"和"分身描述"输入框 |
| 学生首页 FE-P3 | 教师卡片增加授权状态（"申请使用" / "已授权" / "审批中"） |
| 对话页 FE-P5 | SSE 流式输出，逐字渲染 AI 回复 |
| 添加文档页 FE-P8 | 三 Tab 切换：文本录入 / 文件上传 / URL 导入 |

#### 新增页面

| 页面 | 角色 | 说明 |
|------|------|------|
| 🆕 师生管理页 | teacher | 学生列表、审批申请、邀请学生 |
| 🆕 我的教师页 | student | 已授权 / 待审批教师列表 |
| 🆕 学生详情页 | teacher | 评语 + 问答风格设置 + 学生基础信息 |
| 🆕 我的评语页 | student | 查看教师评语 |
| 🆕 提交作业页 | student | 标题 + 内容 + 文件上传 |
| 🆕 作业列表页 | teacher | 查看学生提交的作业 |
| 🆕 作业详情页 | 所有 | 作业内容 + AI 点评 + 教师点评 |
| 🆕 我的作业页 | student | 查看自己的作业和点评 |

---

### 3.5 后端模块划分

| 模块编号 | 模块名称 | 优先级 | 涉及文件 |
|----------|----------|--------|----------|
| V2-M1 | 教师注册增强 | P0 | auth_plugin.go, database.go, repository.go, models.go |
| V2-M2 | 师生授权机制 | P0 | handlers.go, router.go, repository.go, models.go |
| V2-M3 | 对话鉴权改造 | P0 | handlers.go, dialogue_plugin.go |
| V2-M4 | 教师评语系统 | P1 | handlers.go, router.go, repository.go, models.go |
| V2-M5 | 个性化问答风格 | P1 | handlers.go, router.go, repository.go, dialogue_plugin.go, prompt.go |
| V2-M6 | 作业系统 | P1 | handlers.go, router.go, repository.go, models.go, llm_client.go |
| V2-M7 | 文件上传 | P1 | handlers.go, router.go, knowledge_plugin.go, 新增 file_parser.go |
| V2-M8 | URL 导入 | P1 | handlers.go, router.go, 新增 url_fetcher.go |
| V2-M9 | SSE 流式输出 | P1 | handlers.go, router.go, dialogue_plugin.go, llm_client.go |
| V2-M10 | 集成测试 | P0 | tests/integration/ |

**开发顺序**：
```
第1层（并行）: V2-M1 教师注册增强 + V2-M7 文件上传 + V2-M8 URL 导入
      ↓
第2层（并行）: V2-M2 师生授权 + V2-M3 对话鉴权 + V2-M9 SSE 流式输出
      ↓
第3层（并行）: V2-M4 评语系统 + V2-M5 问答风格 + V2-M6 作业系统
      ↓
第4层: V2-M10 集成测试
```

### 3.6 前端模块划分

| 模块编号 | 模块名称 | 优先级 |
|----------|----------|--------|
| V2-FE-M1 | 角色选择页改造（学校+描述） | P0 |
| V2-FE-M2 | 学生首页改造（授权状态） | P0 |
| V2-FE-M3 | 师生管理页 + 我的教师页 | P0 |
| V2-FE-M4 | 对话页 SSE 改造 | P1 |
| V2-FE-M5 | 添加文档页改造（文件上传 + URL） | P1 |
| V2-FE-M6 | 学生详情页（评语 + 风格设置） | P1 |
| V2-FE-M7 | 我的评语页 | P1 |
| V2-FE-M8 | 作业相关页面（提交/列表/详情/我的作业） | P1 |

---

### 3.7 迭代1 交付标准

| 编号 | 验收项 |
|------|--------|
| AC-01 | 教师注册必填学校和描述，同名+同校不允许注册（409） |
| AC-02 | 教师可邀请学生（邀请即同意），学生可申请（需审批） |
| AC-03 | 未授权学生对话返回 403 |
| AC-04 | 教师可对学生写评语，学生可查看 |
| AC-05 | 教师可针对每个学生设置问答风格，对话时生效 |
| AC-06 | 学生可提交作业（文本+文件），AI 自动点评，教师可手动点评 |
| AC-07 | 教师可上传 PDF/DOCX/TXT/MD 文件，自动解析入库 |
| AC-08 | 教师可输入 URL，自动抓取网页内容入库 |
| AC-09 | 对话 SSE 流式输出，前端逐字渲染 |
| AC-10 | 所有新功能通过集成测试 |

---

## 4. 迭代2：生产就绪与上线（~3 周）

> **目标**：让系统安全、稳定地跑在服务器上，补全运营能力，完成发布

### 4.1 生产基础设施

| 编号 | 事项 | 说明 |
|------|------|------|
| V2-PROD-01 | Docker 容器化 | Dockerfile + docker-compose（Go 后端 + Chroma DB + Nginx 三容器） |
| V2-PROD-02 | Chroma DB 持久化 | Docker 运行 Chroma，替换内存向量库，数据挂载到宿主机 |
| V2-PROD-03 | Nginx + HTTPS | Let's Encrypt 证书、反向代理、前端静态文件托管 |
| V2-PROD-04 | SQLite 生产加固 | WAL 模式、数据目录挂载到宿主机、定时备份脚本 |
| V2-PROD-05 | 安全加固 | JWT Secret 环境变量强制校验、API 限流落地、CORS 收紧为实际域名 |
| V2-PROD-06 | 日志持久化 | 结构化日志输出到文件 + logrotate 配置 |
| V2-PROD-07 | 环境配置分离 | `.env.production` 模板 + `harness.production.yaml` |

### 4.2 体验与运营

| 编号 | 事项 | 说明 |
|------|------|------|
| V2-EXP-01 | 记忆衰减机制 | 艾宾浩斯遗忘曲线，定时清理低强度记忆（30 天未触发清理） |
| V2-EXP-02 | 数据分析看板（教师端） | 学生对话统计、知识库热度、学习进度概览 |
| V2-EXP-03 | 对话导出 | 对话记录导出为 TXT |

### 4.3 上线收尾

| 编号 | 事项 | 说明 |
|------|------|------|
| V2-REL-01 | 小程序审核发布 | 域名备案、服务器域名配置、微信提审 |
| V2-REL-02 | 数据备份策略 | SQLite + Chroma 定时备份（cron）+ 保留策略 |
| V2-REL-03 | 监控告警 | 健康检查失败告警、磁盘/内存阈值告警 |
| V2-REL-04 | 全链路回归测试 | 生产环境全流程验证 |

### 4.4 后端模块划分

| 模块编号 | 模块名称 | 优先级 |
|----------|----------|--------|
| V2-PROD-M1 | Docker 容器化 + Chroma 持久化 | P0 |
| V2-PROD-M2 | Nginx + HTTPS | P0 |
| V2-PROD-M3 | SQLite 加固 + 安全加固 | P0 |
| V2-PROD-M4 | 日志 + 环境配置 | P0 |
| V2-EXP-M1 | 记忆衰减机制 | P1 |
| V2-EXP-M2 | 数据分析看板 | P1 |
| V2-EXP-M3 | 对话导出 | P2 |
| V2-REL-M1 | 小程序发布 + 备份 + 监控 | P0 |
| V2-REL-M2 | 全链路回归测试 | P0 |

**开发顺序**：
```
第1层（并行）: V2-PROD-M1 Docker + V2-PROD-M2 Nginx + V2-PROD-M3 安全
      ↓
第2层（并行）: V2-PROD-M4 日志配置 + V2-EXP-M1 记忆衰减 + V2-EXP-M2 看板
      ↓
第3层（并行）: V2-EXP-M3 导出 + V2-REL-M1 发布
      ↓
第4层: V2-REL-M2 全链路回归
```

### 4.5 前端模块划分

| 模块编号 | 模块名称 | 优先级 |
|----------|----------|--------|
| V2-FE-PROD-M1 | 数据看板页（教师端） | P1 |
| V2-FE-PROD-M2 | 对话导出功能 | P2 |
| V2-FE-PROD-M3 | CustomTabBar 完善（角色动态切换） | P1 |

### 4.6 迭代2 交付标准

| 编号 | 验收项 |
|------|--------|
| AC-11 | `docker-compose up -d` 一键启动全部服务 |
| AC-12 | HTTPS 可访问、API 限流生效 |
| AC-13 | 向量数据和 SQLite 数据持久化到宿主机 |
| AC-14 | 记忆自动衰减，30 天未触发的记忆清理 |
| AC-15 | 教师可查看数据分析看板 |
| AC-16 | 小程序通过微信审核，可正式使用 |
| AC-17 | 备份策略生效，可验证恢复 |
| AC-18 | 监控告警可触发通知 |

---

## 5. 技术架构（单机部署）

```
┌──────────────────────────────────────────────────┐
│                   服务器（单机）                    │
│                                                  │
│  ┌──────────┐  ┌───────────┐  ┌──────────────┐  │
│  │  Nginx    │  │  Go 后端   │  │  Chroma DB   │  │
│  │ (HTTPS)   │→ │  (:8080)  │→ │  (:8000)     │  │
│  │ (:443)    │  │           │  │              │  │
│  └──────────┘  └───────────┘  └──────────────┘  │
│       ↑             ↑               ↑            │
│  静态文件       SQLite.db       向量数据          │
│  (前端dist)    (挂载到宿主机)   (挂载到宿主机)     │
│                     ↑                            │
│              uploads/ (文件存储)                   │
│              (作业文件 + 知识库文件)                │
│                                                  │
│  docker-compose 统一编排                          │
└──────────────────────────────────────────────────┘
         ↑
    微信小程序 ←→ 微信服务器
```

---

## 6. 前端页面总览（V2.0 完成后）

| 编号 | 页面 | 角色 | 来源 |
|------|------|------|------|
| P1 | 登录页 | 所有 | V1.0 已有 |
| P2 | 角色选择页（增强：学校+描述） | 新用户 | V1.0 → 迭代1 改造 |
| P3 | 学生首页（增强：授权状态） | student | V1.0 → 迭代1 改造 |
| P4 | 对话页（增强：SSE 流式） | student | V1.0 → 迭代1 改造 |
| P5 | 对话历史页 | student | V1.0 已有 |
| P6 | 知识库管理页 | teacher | V1.0 已有 |
| P7 | 添加文档页（增强：文件上传+URL） | teacher | V1.0 → 迭代1 改造 |
| P8 | 个人中心页 | 所有 | V1.0 已有 |
| P9 | 记忆查看页 | student | V1.0 已有 |
| P10 | 🆕 师生管理页 | teacher | 迭代1 新增 |
| P11 | 🆕 我的教师页 | student | 迭代1 新增 |
| P12 | 🆕 学生详情页（评语+风格设置） | teacher | 迭代1 新增 |
| P13 | 🆕 我的评语页 | student | 迭代1 新增 |
| P14 | 🆕 提交作业页 | student | 迭代1 新增 |
| P15 | 🆕 作业列表页 | teacher | 迭代1 新增 |
| P16 | 🆕 作业详情页（含点评） | 所有 | 迭代1 新增 |
| P17 | 🆕 我的作业页 | student | 迭代1 新增 |
| P18 | 🆕 数据看板页 | teacher | 迭代2 新增 |

---

## 7. 接口总览（V2.0 新增/改造）

### 7.1 迭代1 新增接口

| 编号 | 方法 | 路径 | 说明 | 鉴权 |
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

### 7.2 迭代1 改造接口

| 接口 | 改造内容 |
|------|----------|
| `POST /api/auth/complete-profile` | 教师必填 school + description，nickname+school 唯一校验 |
| `POST /api/chat` | 师生授权鉴权 + 个性化问答风格注入 |

### 7.3 迭代2 新增接口

| 编号 | 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|------|
| API-18 | GET | `/api/analytics/dashboard` | 教师数据看板 | teacher |
| API-19 | GET | `/api/conversations/export` | 对话导出（TXT） | 所有 |

---

## 8. 新增错误码

| 错误码 | 说明 | HTTP Status |
|--------|------|-------------|
| 40007 | 未获得该教师授权 | 403 |
| 40008 | 该学校已有同名教师 | 409 |
| 40009 | 师生关系已存在 | 409 |
| 40010 | 文件格式不支持 | 400 |
| 40011 | 文件大小超限 | 400 |
| 40012 | URL 不可达或解析失败 | 400 |

---

## 9. 新增环境变量

| 变量名 | 必填 | 默认值 | 说明 | 迭代 |
|--------|------|--------|------|------|
| `WX_APPID` | 生产必填 | - | 微信小程序 AppID | 迭代2 |
| `WX_SECRET` | 生产必填 | - | 微信小程序 AppSecret | 迭代2 |
| `UPLOAD_DIR` | ❌ | `./uploads` | 文件上传存储目录 | 迭代1 |
| `MAX_UPLOAD_SIZE` | ❌ | `52428800` | 最大上传文件大小（字节，默认 50MB） | 迭代1 |

---

## 10. 需求追溯

| 需求来源 | 事项 | V2.0 迭代 |
|----------|------|-----------|
| 🆕 用户需求 | 师生授权机制 | 迭代1 |
| 🆕 用户需求 | 教师注册增强（名称+学校唯一） | 迭代1 |
| 🆕 用户需求 | 教师评语系统 | 迭代1 |
| 🆕 用户需求 | 个性化问答风格 | 迭代1 |
| 🆕 用户需求 | 学生作业/成果上传 + AI/教师点评 | 迭代1 |
| 全局需求 3.2 | 文档上传 PDF/DOCX | 迭代1 |
| V3.0 BL-001 | URL 导入 | 迭代1 |
| 全局需求 3.3 | 流式输出 | 迭代1 |
| 生产必备 | Docker + HTTPS + 安全 | 迭代2 |
| 全局需求 3.5 | 记忆衰减 | 迭代2 |
| 全局需求 3.5 | 数据分析看板 | 迭代2 |
| 全局需求 3.5 | 导出功能 | 迭代2 |
| 生产必备 | 小程序发布 | 迭代2 |

---

## 11. 时间线

```
Week 1~4:   迭代1 - 核心功能开发（师生授权 + 作业 + 文件上传 + URL + SSE）
Week 5~7:   迭代2 - 生产就绪与上线（Docker + 安全 + 看板 + 发布）

总计: ~7 周
```

---

## 12. 风险与应对

| 风险 | 影响 | 应对方案 |
|------|------|----------|
| PDF/DOCX 解析库兼容性 | 文件上传功能不可用 | 优先支持 TXT/MD，PDF/DOCX 使用成熟 Go 库 |
| 网页抓取被反爬 | URL 导入失败 | 设置合理 User-Agent + 超时 + 错误提示 |
| SSE 小程序兼容性 | 流式输出不可用 | 小程序使用 wx.request + 轮询降级方案 |
| 微信小程序审核不通过 | 无法正式发布 | 提前了解审核规则，准备备案材料 |
| SQLite 并发写入瓶颈 | 高并发下性能下降 | WAL 模式 + 写入队列 + 连接池优化 |
| 师生授权改造影响现有对话 | 已有用户无法对话 | 数据迁移脚本：为已有对话关系自动创建 approved 记录 |

---

**文档版本**: v1.0.0
**创建日期**: 2026-03-28
**最后更新**: 2026-03-28
