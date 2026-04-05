# V2.0 需求规格说明书（全量版）

> **最后更新**: 2026-04-04 | **文档版本**: v2.0.0 | **状态**: 迭代10已完成 + TabBar调整

---

## 1. 版本概述

| 项目 | 说明 |
|------|------|
| **版本名称** | V2.0 - 单机生产可用版 |
| **版本目标** | 功能完整 + 单机部署 + 生产可用 + 多端支持 |
| **迭代数** | 10 个迭代（已全部完成） |
| **前置依赖** | V1.0 全部完成（3 个迭代，39 个集成测试通过） |

### 1.1 V1.0 已完成内容回顾

| V1.0 迭代 | 内容 | 状态 |
|------------|------|------|
| 迭代1 | Harness 框架 + 4 个核心插件（认证/知识库/记忆/对话） + HTTP API | ✅ |
| 迭代2 | Taro 小程序前端 10 页面 + 后端微信登录适配 | ✅ |
| 迭代3 | 管道编排落地 + 记忆自动提取 + 令牌刷新 + 配置校验 + 超时控制 | ✅ |

### 1.2 V2.0 目标

> **单机部署、生产可用、功能完整、多端支持** —— 一台服务器跑起来，能真正给师生用；同时支持小程序和H5多端访问。

### 1.3 明确排除

| 事项 | 原因 |
|------|------|
| 多机部署 / K8s / 自动扩缩容 | 保持单机部署 |
| 多租户支持 | 企业级特性 |
| 插件市场 / API 开放平台 | 企业级特性 |
| 智能推荐 | 优先级低，可后续补充 |

---

## 2. 迭代规划总览

| 迭代 | 主题 | 核心内容 | 状态 |
|------|------|----------|------|
| 迭代1 | 核心功能开发 | 师生授权 + 注册增强 + 评语 + 问答风格 + 作业系统 + 文件上传 + URL导入 + SSE流式输出 | ✅ |
| 迭代2 | 多角色多分身架构 | 多分身体系 + 班级管理 + 分身分享 + 知识库精细化 + 分身化改造 | ✅ |
| 迭代3 | UI重构与体验优化 | 教师/学生仪表盘 + 班级详情 + 启停管理 + 知识库scope多选 + 上传预览 + 分享码优化 | ✅ |
| 迭代4 | 分身广场与教师介入 | 分身广场/定向邀请 + 教师真人介入对话 + LLM智能摘要 + 分身概览 | ✅ |
| 迭代5 | LlamaIndex语义检索 + UX重构 | Python LlamaIndex服务 + 教师Dashboard + 学生管理合并 + 对话附件 + 广场独立化 | ✅ |
| 迭代6 | 记忆增强 + 对话风格 + 部署 | 记忆三层分层 + 6种对话风格模板 + 二维码分享 + 聊天记录导入 + 自定义TabBar + Docker部署 | ✅ |
| 迭代7 | 教材配置 + 安全加固 | 8档学段教材配置 + Adaptive RAG + 批量上传 + API限流 + 隐私防护 + 语音/Emoji + 班级推送 + 用户画像 | ✅ |
| 迭代8 | 易用性优化 | 知识库统一输入框 + 班级管理增强 + 审批流程 + 聊天列表仿微信改版 + 发现页 | ✅ |
| 迭代9 | 对话体验增强 + 教学管理 | 思考过程展示 + 语音恢复 + +号多功能按钮 + 会话列表优化 + 课程发布 + 画像隐私保护 | ✅ |
| 迭代10 | 管理员H5后台 + 多端支持 | 微信H5 OAuth登录 + 管理员监控面板 + 教师/学生H5页面 + 操作日志流水 | ✅ |

**迭代后变更**：
- 教师端 TabBar 调整：移除"工作台"，改为"聊天列表"（按班级划分） ✅

---

## 3. 功能模块总览

### 3.1 认证与用户管理

| 功能 | 说明 | 来源迭代 |
|------|------|----------|
| 微信小程序登录 | wx.login → code换token → JWT | V1.0 |
| 微信H5网页授权 | OAuth 2.0 snsapi_userinfo → JWT | 迭代10 |
| 角色选择 | 新用户选择教师/学生角色 | V1.0 |
| 教师注册增强 | 微信昵称自动获取，首次登录引导创建班级 | 迭代1→迭代8简化 |
| JWT令牌管理 | 签发/刷新/过期/黑名单 | V1.0+迭代10 |
| 用户禁用/启用 | 管理员可禁用用户，token立即失效，所有请求返回403 | 迭代10 |
| 管理员角色 | 通过数据库脚本设置，不通过前端注册 | 迭代10 |

### 3.2 多角色多分身体系

| 功能 | 说明 | 来源迭代 |
|------|------|----------|
| 分身创建 | 教师可创建多个AI分身（不同学科/班级） | 迭代2 |
| 分身切换 | 教师/学生可切换当前活跃分身 | 迭代2 |
| 分身启停 | 教师可启用/停用分身 | 迭代3 |
| 分身概览 | 分身详情页（基本信息+统计+设置） | 迭代4 |
| 分身广场 | 公开分身展示，学生可浏览和申请 | 迭代4→迭代5独立化 |
| 分身分享 | 分享链接/二维码/定向邀请 | 迭代2+迭代4+迭代6 |

### 3.3 师生关系与班级管理

| 功能 | 说明 | 来源迭代 |
|------|------|----------|
| 班级创建 | 教师昵称+学科+年龄范畴+班级名称+简介 | 迭代2→迭代8增强 |
| 班级分享链接 | 创建后自动生成专属分享链接和二维码 | 迭代2+迭代6 |
| 学生审批流程 | 学生申请加入→教师填写学生信息→审批通过/拒绝 | 迭代8 |
| 审批信息 | 评价/特点、年龄、性别、家庭情况（用于Prompt注入） | 迭代8 |
| 学生管理（合并页） | 全部学生/按班级/待审批/班级设置 四Tab合一 | 迭代5合并 |
| 批量添加学生 | LLM智能解析文本+分享码加入+信息丰富化引导 | 迭代7 |
| 班级启停 | 教师可启用/停用班级 | 迭代3 |

### 3.4 对话系统

| 功能 | 说明 | 来源迭代 |
|------|------|----------|
| SSE流式输出 | 大模型流式调用→SSE推送→前端逐字渲染 | 迭代1 |
| 师生授权鉴权 | 未授权学生对话返回403 | 迭代1 |
| 个性化问答风格 | 教师可针对每个学生设置对话风格 | 迭代1 |
| 教师真人介入 | 教师可接管AI对话，发送真人消息 | 迭代4 |
| 对话附件支持 | 学生可在对话中发送文件，AI自动识别并点评 | 迭代5 |
| 6种对话风格模板 | 苏格拉底式/直接讲解/鼓励引导/严格训练/游戏化/自适应 | 迭代6 |
| 思考过程展示 | RAG检索/记忆检索/工具调用/LLM思考 实时展示 | 迭代9 |
| 语音输入 | 长按录音→语音转文字→填入输入框 | 迭代7→迭代9恢复 |
| Emoji表情 | 表情面板，支持常用表情分类 | 迭代7 |
| +号多功能按钮 | 文件/相册/拍摄 功能面板 | 迭代9 |
| 新会话指令触发 | 用户发送"新话题"等关键词创建新session | 迭代9 |
| 会话列表优化 | 按老师→会话二级结构，历史会话LLM总结标题 | 迭代9 |
| 仿微信聊天UI | 底部输入栏：语音→输入框→Emoji→+号 | 迭代8 |

### 3.5 知识库系统

| 功能 | 说明 | 来源迭代 |
|------|------|----------|
| 文本录入 | 直接输入文本内容 | V1.0 |
| 文件上传 | PDF/DOCX/TXT/MD，单文件≤20MB | 迭代1 |
| URL导入 | 网页抓取→HTML解析→正文提取 | 迭代1 |
| 聊天记录导入 | JSON格式对话记录导入 | 迭代6 |
| 批量上传 | 最多20个文件，总大小≤100MB，异步处理 | 迭代7 |
| 统一输入框 | 智能识别URL/文字/文件，拖拽上传 | 迭代8 |
| LlamaIndex语义检索 | Python服务，DashScope Embedding向量化+语义检索 | 迭代5 |
| Scope精细化 | 全局/班级/学生 三级scope控制 | 迭代2+迭代3多选 |
| 知识库预览 | 点击文档可预览内容 | 迭代3 |
| 搜索与筛选 | 按标题关键词搜索，按类型筛选 | 迭代8 |
| 课程类型 | item_type='course'，课程信息存入知识库 | 迭代9 |

### 3.6 记忆系统

| 功能 | 说明 | 来源迭代 |
|------|------|----------|
| 记忆自动提取 | 对话后自动提取关键记忆 | V1.0 |
| 三层分层存储 | core(核心)/episodic(情景)/archived(归档) | 迭代6 |
| 记忆摘要合并 | 定时合并相似记忆，LLM生成摘要 | 迭代6 |
| 用户画像持久化 | 记忆合并时提炼学生/教师画像，存入profile_snapshot | 迭代7 |
| 画像隐私保护 | 画像仅内部使用，API不返回profile_snapshot | 迭代9 |
| 记忆管理API | 查看/编辑/删除记忆 | 迭代6 |

### 3.7 教学管理

| 功能 | 说明 | 来源迭代 |
|------|------|----------|
| 学生备注 | 教师私密备注（原评语），学生不可见，用于AI个性化 | 迭代1→迭代5改名 |
| 教材配置体系 | 8档学段 + 教材版本 + 学科 + 进度配置 | 迭代7 |
| 学段Prompt模板 | 每个学段独立模板，约束语言风格和回答深度 | 迭代7 |
| 课程信息发布 | 教师发布课程→存入知识库→可推送给学生 | 迭代9 |
| 班级消息推送 | 教师向班级/指定学生推送消息，显示在聊天页 | 迭代7 |
| 微信订阅消息 | 课程更新/消息推送通过微信订阅消息触达 | 迭代9 |
| 用户反馈系统 | 功能建议/Bug报告/内容问题/其他 | 迭代7 |
| ~~作业系统~~ | ~~学生提交作业+AI/教师点评~~ **（迭代7已移除）** | ~~迭代1~~ |

### 3.8 前端UI/UX（小程序端）

#### 教师端 TabBar（最新）

| Tab | 图标 | 页面 | 功能 |
|-----|------|------|------|
| 聊天列表 | 💬 | chat-list | 按班级分组的学生聊天列表 + 置顶 + 未读消息 |
| 学生管理 | 👥 | teacher-students | 全部学生/按班级/待审批/班级设置 |
| 知识库 | 📚 | knowledge | 知识库管理（统一输入框 + 搜索筛选） |
| 我的 | 👤 | profile | 分身概览 + 课程管理 + 设置 + 退出 |

> **变更说明**：原"工作台"Tab已移除，替换为"聊天列表"Tab（按班级划分学生）。

#### 学生端 TabBar

| Tab | 图标 | 页面 | 功能 |
|-----|------|------|------|
| 对话 | 💬 | home | 我的老师列表 + 开始对话 |
| 发现 | 🌐 | discover | 教师广场 + 分享码加入 |
| 我的 | 👤 | profile | 个人信息 + 设置 + 退出 |

#### 聊天列表页（教师端）

- 按班级分组展示学生
- 每个班级可展开/收起
- 支持班级置顶和学生置顶（📌标识）
- 显示未读消息数量角标
- 默认每班显示前5个学生，可展开查看更多
- 点击学生进入聊天详情

#### 聊天列表页（学生端）

- 按老师→会话二级结构展示
- 最新会话显示最后一条消息摘要
- 历史会话显示LLM总结的标题
- 支持老师置顶
- 底部仿微信输入栏 + 新会话按钮

### 3.9 H5多端支持（迭代10）

#### 管理员H5后台

| 页面 | 功能 |
|------|------|
| 仪表盘 | 系统总览（用户数/对话数/消息数等）+ 趋势图（7/30/90天）+ 用户统计 + 对话统计 + 知识库统计 + 活跃排行 |
| 用户管理 | 用户列表 + 搜索 + 修改角色 + 启用/禁用 |
| 反馈管理 | 反馈列表 + 状态筛选 + 状态更新 |
| 日志查询 | 操作日志列表 + 多条件筛选 + 统计 + CSV导出 |

- 技术栈：Vue 3 + Element Plus
- 布局：左侧导航栏 + 顶部标题栏

#### 教师H5页面

功能对齐小程序教师端，UI独立设计。核心页面：
- 聊天列表（按班级组织）、对话页、班级管理、知识库、课程管理、分身管理、分享管理、记忆管理、审批管理、学生详情、教材配置、反馈、个人中心

侧边栏导航顺序：聊天列表 → 学生管理 → 知识库 → 我的 → 课程管理 → 分身管理

- 技术栈：Vue 3 + Element Plus
- 默认进入聊天列表页

#### 学生H5页面

功能对齐小程序学生端，UI独立设计。核心页面：
- 聊天列表、对话页、发现页、分享加入、历史记录、我的教师、我的评语、反馈、个人中心

底部TabBar：对话 / 发现 / 我的

- 技术栈：Vue 3 + Vant 4（移动端组件库）

### 3.10 管理员后台功能

#### 操作日志流水（迭代10）

| 功能 | 说明 |
|------|------|
| 全局中间件采集 | 自动记录所有API请求（方法/路径/状态码/耗时/用户/IP/平台） |
| 语义增强 | 关键操作补充业务语义（action/resource/resource_id/detail） |
| 独立存储 | 独立SQLite数据库文件，与业务数据库分离 |
| 保留策略 | 90天自动清理（每日凌晨定时任务） |
| 日志查询 | 多条件筛选（用户/操作类型/资源/时间/平台） |
| 日志统计 | 操作频次/平台分布/时段热力图/活跃用户 |
| 日志导出 | CSV格式导出（应用当前筛选条件） |

操作类型枚举：`user.login` / `user.register` / `chat.send_message` / `chat.create_session` / `class.create` / `knowledge.upload` / `persona.create` / `share.create` / `course.push` / `admin.update_role` / `admin.toggle_user` 等 30+ 种。

### 3.11 运维与部署

| 功能 | 说明 | 来源迭代 |
|------|------|----------|
| Docker Compose | Go后端 + Python LlamaIndex + Nginx 三容器编排 | 迭代6 |
| Nginx + HTTPS | Let's Encrypt证书、反向代理、静态文件托管 | 迭代6 |
| SQLite生产加固 | WAL模式、数据目录挂载、定时备份 | 迭代6 |
| API限流 | 全局请求限流 + 对话接口单独限流 | 迭代7 |
| Prompt Injection防御 | System Prompt安全指令 + 记忆注入脱敏 | 迭代7 |
| 日志持久化 | 结构化日志输出到文件 + logrotate | 迭代6 |
| 环境配置分离 | `.env.production` + `harness.production.yaml` | 迭代6 |
| Adaptive RAG | LLM自主决策是否需要外部搜索 | 迭代7 |

---

## 4. 技术架构（最新版）

```
┌──────────────────────────────────────────────────────────────┐
│                      服务器（单机部署）                         │
│                                                                │
│  ┌──────────┐  ┌───────────────┐  ┌────────────────────┐     │
│  │  Nginx    │  │   Go 后端      │  │  Python LlamaIndex │     │
│  │ (HTTPS)   │→ │   (:8080)     │→ │  服务 (:8100)      │     │
│  │ (:443)    │  │               │  │                    │     │
│  └──────────┘  └───────────────┘  └────────────────────┘     │
│       ↑              ↑                     ↑                  │
│  静态文件        SQLite.db           SimpleVectorStore        │
│  (前端dist)     (业务数据)           (本地JSON文件持久化)       │
│  (H5 dist)          ↑                     ↑                  │
│              operation_logs.db      DashScope Embedding       │
│              (操作日志-独立)        (text-embedding-v3)        │
│                     ↑                                         │
│              uploads/ (文件存储)                               │
│              (知识库文件 + 对话附件)                            │
│                                                                │
│  docker-compose 统一编排                                       │
└──────────────────────────────────────────────────────────────┘
         ↑                    ↑
    微信小程序            H5页面（微信内浏览器）
    (Taro)               ├── 管理员后台 (Vue3+ElementPlus)
                         ├── 教师端 (Vue3+ElementPlus)
                         └── 学生端 (Vue3+Vant4)
```

### 4.1 技术栈

| 组件 | 技术选型 |
|------|---------|
| 后端 | Go + Gin |
| 小程序前端 | Taro 3 + React + TypeScript |
| 管理员H5 | Vue 3 + Element Plus |
| 教师H5 | Vue 3 + Element Plus |
| 学生H5 | Vue 3 + Vant 4 |
| 语义检索 | Python + FastAPI + LlamaIndex |
| Embedding | 通义千问 text-embedding-v3 (DashScope) |
| LLM | 通义千问 (DashScope) |
| 向量存储 | LlamaIndex SimpleVectorStore (本地JSON) |
| 业务数据库 | SQLite (WAL模式) |
| 日志数据库 | SQLite (独立文件) |
| 容器化 | Docker Compose |
| 反向代理 | Nginx + HTTPS |

---

## 5. 数据库设计（完整表清单）

### 5.1 业务数据库表

| 表名 | 说明 | 来源迭代 |
|------|------|----------|
| `users` | 用户表（含role/openid/wx_unionid/status/profile_snapshot） | V1.0+多次扩展 |
| `personas` | 分身表（教师创建的AI分身） | 迭代2 |
| `classes` | 班级表（含teacher_display_name/subject/age_group/share_link/invite_code/qr_code_url） | 迭代2+迭代8增强 |
| `class_members` | 班级成员表（含approval_status/teacher_evaluation/age/gender/family_info） | 迭代2+迭代8增强 |
| `class_join_requests` | 班级加入申请表 | 迭代8 |
| `conversations` | 对话记录表 | V1.0 |
| `knowledge_items` | 知识库条目表（含item_type: text/url/file/course） | 迭代2+迭代9扩展 |
| `memories` | 记忆表（三层分层: core/episodic/archived） | V1.0+迭代6增强 |
| `teacher_student_relations` | 师生授权关系表 | 迭代1 |
| `teacher_comments` | 教师评语/学生备注表 | 迭代1 |
| `student_dialogue_styles` | 学生问答风格配置表 | 迭代1 |
| `persona_shares` | 分身分享码表 | 迭代2 |
| `teacher_curriculum_configs` | 教材配置表（学段/年级/教材版本/学科/进度） | 迭代7 |
| `feedbacks` | 用户反馈表 | 迭代7 |
| `batch_tasks` | 批量任务表（批量上传进度跟踪） | 迭代7 |
| `teacher_messages` | 教师推送消息表 | 迭代7 |
| `chat_pins` | 聊天置顶记录表 | 迭代8 |
| `course_notifications` | 课程推送通知记录表 | 迭代9 |
| `wx_subscriptions` | 微信订阅状态记录表 | 迭代9 |
| `session_titles` | 会话标题表（LLM总结） | 迭代9 |

### 5.2 日志数据库表（独立文件）

| 表名 | 说明 | 来源迭代 |
|------|------|----------|
| `operation_logs` | 操作日志表（user_id/action/resource/ip/platform/duration_ms等） | 迭代10 |

---

## 6. 前端页面总览

### 6.1 小程序页面

| 页面 | 路径 | 角色 | 来源 |
|------|------|------|------|
| 登录页 | pages/login | 所有 | V1.0 |
| 角色选择页 | pages/role-select | 新用户 | V1.0→迭代8简化 |
| 分身选择页 | pages/persona-select | 所有 | 迭代2 |
| 首页 | pages/home | 所有 | V1.0→多次重构 |
| 对话页 | pages/chat | 所有 | V1.0→迭代9增强 |
| 聊天列表页 | pages/chat-list | 所有 | 迭代8 |
| 对话历史页 | pages/history | student | V1.0 |
| 知识库管理页 | pages/knowledge | teacher | V1.0→迭代8增强 |
| 添加知识页 | pages/knowledge/add | teacher | V1.0→迭代8统一输入框 |
| 知识库预览页 | pages/knowledge/preview | teacher | 迭代3 |
| 个人中心页 | pages/profile | 所有 | V1.0 |
| 记忆查看页 | pages/memories | student | V1.0 |
| 记忆管理页 | pages/memory-manage | teacher | 迭代6 |
| 学生管理页 | pages/teacher-students | teacher | 迭代1→迭代5合并 |
| 我的教师页 | pages/my-teachers | student | 迭代1 |
| 学生详情页 | pages/student-detail | teacher | 迭代1 |
| 学生档案页 | pages/student-profile | teacher | 迭代8 |
| 分享加入页 | pages/share-join | student | 迭代2 |
| 分享管理页 | pages/share-manage | teacher | 迭代2→迭代5移入分身管理 |
| 班级创建页 | pages/class-create | teacher | 迭代2 |
| 班级详情页 | pages/class-detail | teacher | 迭代3 |
| 分身概览页 | pages/persona-overview | teacher | 迭代4 |
| 学生聊天历史页 | pages/student-chat-history | teacher | 迭代4 |
| 发现页 | pages/discover | student | 迭代5独立化→迭代8增强 |
| 教师消息页 | pages/teacher-message | teacher | 迭代7 |
| 教材配置页 | pages/curriculum-config | teacher | 迭代7 |
| 反馈页 | pages/feedback | 所有 | 迭代7 |
| 反馈管理页 | pages/feedback-manage | teacher | 迭代7 |
| 批量添加学生页 | pages/student-batch | teacher | 迭代7 |
| 审批管理页 | pages/approval-manage | teacher | 迭代8 |
| 审批详情页 | pages/approval-detail | teacher | 迭代8 |
| 我的评语页 | pages/my-comments | student | 迭代1（学生端入口已移除） |
| 课程发布页 | pages/course-publish | teacher | 迭代9 |
| 课程列表页 | pages/course-list | teacher | 迭代9 |

### 6.2 H5页面

#### 管理员后台（src/h5-admin/）
- 登录页、仪表盘、用户管理、反馈管理、日志查询、日志统计

#### 教师端（src/h5-teacher/）
- 登录页、角色选择、聊天列表、对话页、班级管理、班级创建、班级详情、知识库、知识库添加、知识库预览、课程管理、课程发布、分身管理、分享管理、记忆管理、审批管理、学生详情、教材配置、反馈、个人中心

#### 学生端（src/h5-student/）
- 登录页、角色选择、首页、聊天列表、对话页、发现页、分享加入、历史记录、我的教师、我的评语、分身管理、反馈、个人中心

---

## 7. 接口完整清单（基于 router.go 实际代码）

> 以下接口清单从 `src/backend/api/router.go` 和 `src/knowledge-service/app/main.py` 中逐行提取，共计 **102 个接口**。

### 7.1 认证接口（无需鉴权）

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 1 | POST | `/api/auth/register` | 用户注册 | 所有 | V1.0 |
| 2 | POST | `/api/auth/login` | 用户登录 | 所有 | V1.0 |
| 3 | POST | `/api/auth/wx-login` | 微信小程序登录 | 所有 | V1.0 |
| 4 | POST | `/api/auth/refresh` | 刷新JWT令牌 | 所有 | V1.0 |
| 5 | GET | `/api/auth/wx-h5-login-url` | 获取H5微信授权URL | 所有 | 迭代10 |
| 6 | POST | `/api/auth/wx-h5-callback` | H5微信授权回调 | 所有 | 迭代10 |
| 7 | POST | `/api/auth/complete-profile` | 补全用户信息（需鉴权） | 所有 | V1.0 |

### 7.2 用户接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 8 | GET | `/api/user/profile` | 获取当前用户信息 | 所有 | V1.0 |
| 9 | PUT | `/api/user/student-profile` | 更新学生基础信息 | student | 迭代8 |
| 10 | GET | `/api/teachers` | 获取教师列表 | 所有 | V1.0 |

### 7.3 分身管理接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 11 | POST | `/api/personas` | 创建分身 | 所有 | 迭代2 |
| 12 | GET | `/api/personas` | 获取分身列表 | 所有 | 迭代2 |
| 13 | PUT | `/api/personas/:id` | 编辑分身 | 所有 | 迭代2 |
| 14 | PUT | `/api/personas/:id/activate` | 激活分身 | 所有 | 迭代2 |
| 15 | PUT | `/api/personas/:id/deactivate` | 停用分身 | 所有 | 迭代2 |
| 16 | PUT | `/api/personas/:id/switch` | 切换当前分身 | 所有 | 迭代2 |
| 17 | GET | `/api/personas/:id/dashboard` | 分身Dashboard数据 | 所有 | 迭代3 |
| 18 | GET | `/api/personas/marketplace` | 分身广场 | 所有 | 迭代4 |
| 19 | PUT | `/api/personas/:id/visibility` | 设置分身公开/私密 | 所有 | 迭代4 |

### 7.4 班级管理接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 20 | POST | `/api/classes` | 创建班级 | teacher | 迭代2 |
| 21 | GET | `/api/classes` | 获取班级列表 | teacher | 迭代2 |
| 22 | PUT | `/api/classes/:id` | 更新班级 | teacher | 迭代2 |
| 23 | DELETE | `/api/classes/:id` | 删除班级 | teacher | 迭代2 |
| 24 | GET | `/api/classes/:id/members` | 获取班级成员 | teacher | 迭代2 |
| 25 | POST | `/api/classes/:id/members` | 添加班级成员 | teacher | 迭代2 |
| 26 | DELETE | `/api/classes/:id/members/:member_id` | 移除班级成员 | teacher | 迭代2 |
| 27 | PUT | `/api/classes/:id/toggle` | 启停班级 | teacher | 迭代3 |
| 28 | POST | `/api/classes/v8` | 创建班级（V8增强版） | teacher | 迭代8 |
| 29 | GET | `/api/classes/:id/share-info` | 获取班级分享信息 | teacher | 迭代8 |
| 30 | GET | `/api/classes/:id/members/v8` | 获取班级成员（V8增强版） | teacher | 迭代8 |
| 31 | GET | `/api/classes/:id` | 学生查看班级详情 | 所有 | 迭代9 |

### 7.5 班级加入申请接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 32 | POST | `/api/classes/join` | 学生申请加入班级 | student | 迭代8 |
| 33 | GET | `/api/join-requests/pending` | 获取待审批申请列表 | teacher | 迭代8 |
| 34 | PUT | `/api/join-requests/:id/approve` | 审批通过 | teacher | 迭代8 |
| 35 | PUT | `/api/join-requests/:id/reject` | 审批拒绝 | teacher | 迭代8 |

### 7.6 师生关系接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 36 | POST | `/api/relations/invite` | 教师邀请学生 | teacher | 迭代1 |
| 37 | POST | `/api/relations/apply` | 学生申请使用分身 | student | 迭代1 |
| 38 | PUT | `/api/relations/:id/approve` | 教师审批同意 | teacher | 迭代1 |
| 39 | PUT | `/api/relations/:id/reject` | 教师审批拒绝 | teacher | 迭代1 |
| 40 | GET | `/api/relations` | 获取师生关系列表 | 所有 | 迭代1 |
| 41 | PUT | `/api/relations/:id/toggle` | 启停师生关系 | teacher | 迭代3 |

### 7.7 对话接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 42 | POST | `/api/chat` | 发送消息（非流式） | 所有 | V1.0 |
| 43 | POST | `/api/chat/stream` | SSE流式对话 | 所有 | 迭代1 |
| 44 | GET | `/api/conversations` | 获取对话历史 | 所有 | V1.0 |
| 45 | GET | `/api/conversations/sessions` | 获取会话列表 | 所有 | V1.0 |
| 46 | POST | `/api/chat/new-session` | 创建新会话 | 所有 | 迭代8 |
| 47 | GET | `/api/chat/quick-actions` | 获取快捷指令 | 所有 | 迭代8 |
| 48 | POST | `/api/conversations/sessions/:session_id/title` | 生成会话标题 | 所有 | 迭代9 |

### 7.8 教师介入对话接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 49 | POST | `/api/chat/teacher-reply` | 教师真人回复 | teacher | 迭代4 |
| 50 | GET | `/api/chat/takeover-status` | 获取接管状态 | 所有 | 迭代4 |
| 51 | POST | `/api/chat/end-takeover` | 结束接管 | teacher | 迭代4 |
| 52 | GET | `/api/conversations/student/:student_persona_id` | 查看学生对话记录 | teacher | 迭代4 |

### 7.9 聊天列表接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 53 | GET | `/api/chat-list/teacher` | 教师端聊天列表（按班级组织） | teacher | 迭代8 |
| 54 | GET | `/api/chat-list/student` | 学生端聊天列表（按老师分组） | student | 迭代8 |
| 55 | POST | `/api/chat-pins` | 置顶聊天 | 所有 | 迭代8 |
| 56 | DELETE | `/api/chat-pins/:type/:id` | 取消置顶 | 所有 | 迭代8 |
| 57 | GET | `/api/chat-pins` | 获取置顶列表 | 所有 | 迭代8 |

### 7.10 知识库接口（旧版 documents）

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 58 | POST | `/api/documents` | 添加文档 | teacher/admin | V1.0 |
| 59 | GET | `/api/documents` | 获取文档列表 | teacher/admin | V1.0 |
| 60 | DELETE | `/api/documents/:id` | 删除文档 | teacher/admin | V1.0 |
| 61 | POST | `/api/documents/upload` | 文件上传 | teacher/admin | 迭代1 |
| 62 | POST | `/api/documents/import-url` | URL导入 | teacher/admin | 迭代1 |
| 63 | POST | `/api/documents/preview` | 预览文档 | teacher/admin | 迭代3 |
| 64 | POST | `/api/documents/preview-upload` | 预览上传文件 | teacher/admin | 迭代3 |
| 65 | POST | `/api/documents/preview-url` | 预览URL内容 | teacher/admin | 迭代3 |
| 66 | POST | `/api/documents/confirm` | 确认添加文档 | teacher/admin | 迭代3 |
| 67 | POST | `/api/documents/import-chat` | 聊天记录导入 | teacher/admin | 迭代6 |
| 68 | POST | `/api/documents/batch-upload` | 批量文件上传 | teacher/admin | 迭代7 |
| 69 | GET | `/api/batch-tasks/:task_id` | 查询批量任务状态 | teacher/admin | 迭代7 |

### 7.11 知识库接口（V8增强版 knowledge）

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 70 | POST | `/api/knowledge/upload` | 智能知识库上传（统一输入框） | teacher/admin | 迭代8 |
| 71 | GET | `/api/knowledge` | 搜索知识库列表 | teacher/admin | 迭代8 |
| 72 | GET | `/api/knowledge/:id` | 获取知识详情 | teacher/admin | 迭代8 |
| 73 | PUT | `/api/knowledge/:id` | 更新知识 | teacher/admin | 迭代8 |
| 74 | DELETE | `/api/knowledge/:id` | 删除知识 | teacher/admin | 迭代8 |

### 7.12 记忆接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 75 | GET | `/api/memories` | 获取记忆列表 | 所有 | V1.0 |
| 76 | PUT | `/api/memories/:id` | 更新记忆 | teacher | 迭代6 |
| 77 | DELETE | `/api/memories/:id` | 删除记忆 | teacher | 迭代6 |
| 78 | POST | `/api/memories/summarize` | 手动触发记忆合并 | teacher | 迭代6 |

### 7.13 评语与问答风格接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 79 | POST | `/api/comments` | 教师写评语/备注 | teacher | 迭代1 |
| 80 | GET | `/api/comments` | 获取评语列表 | 所有 | 迭代1 |
| 81 | PUT | `/api/students/:id/dialogue-style` | 设置学生问答风格 | teacher | 迭代1 |
| 82 | GET | `/api/students/:id/dialogue-style` | 获取学生问答风格 | 所有 | 迭代1 |

### 7.14 对话风格接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 83 | PUT | `/api/styles` | 设置对话风格配置 | teacher | 迭代6 |
| 84 | GET | `/api/styles` | 获取对话风格配置 | 所有 | 迭代6 |

### 7.15 学生管理接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 85 | GET | `/api/students/search` | 搜索学生 | teacher | 迭代1 |
| 86 | GET | `/api/students/:id/profile` | 查看学生详情 | teacher | 迭代9 |
| 87 | PUT | `/api/students/:id/evaluation` | 更新学生评语 | teacher | 迭代9 |
| 88 | POST | `/api/students/parse-text` | LLM解析学生文本 | teacher | 迭代7 |
| 89 | POST | `/api/students/batch-create` | 批量创建学生 | teacher | 迭代7 |

### 7.16 教材配置接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 90 | POST | `/api/curriculum-configs` | 创建教材配置 | teacher | 迭代7 |
| 91 | GET | `/api/curriculum-configs` | 获取教材配置列表 | teacher | 迭代7 |
| 92 | PUT | `/api/curriculum-configs/:id` | 更新教材配置 | teacher | 迭代7 |
| 93 | DELETE | `/api/curriculum-configs/:id` | 删除教材配置 | teacher | 迭代7 |
| 94 | GET | `/api/curriculum-versions` | 获取教材版本列表 | 所有 | 迭代7 |

### 7.17 课程接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 95 | POST | `/api/courses` | 发布课程 | teacher | 迭代9 |
| 96 | GET | `/api/courses` | 课程列表 | teacher | 迭代9 |
| 97 | PUT | `/api/courses/:id` | 更新课程 | teacher | 迭代9 |
| 98 | DELETE | `/api/courses/:id` | 删除课程 | teacher | 迭代9 |
| 99 | POST | `/api/courses/:id/push` | 推送课程给学生 | teacher | 迭代9 |

### 7.18 教师消息推送接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 100 | POST | `/api/teacher-messages` | 教师推送消息 | teacher | 迭代7 |
| 101 | GET | `/api/teacher-messages/history` | 推送历史 | teacher | 迭代7 |

### 7.19 分享码接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 102 | POST | `/api/shares` | 创建分享码 | teacher | 迭代2 |
| 103 | GET | `/api/shares` | 获取分享码列表 | teacher | 迭代2 |
| 104 | POST | `/api/shares/:code/join` | 通过分享码加入 | 所有 | 迭代2 |
| 105 | PUT | `/api/shares/:id/deactivate` | 废止分享码 | teacher | 迭代2 |
| 106 | GET | `/api/shares/:code/info` | 获取分享码信息（可选鉴权） | 所有 | 迭代2 |

### 7.20 反馈接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 107 | POST | `/api/feedbacks` | 提交反馈 | 所有 | 迭代7 |
| 108 | GET | `/api/feedbacks` | 反馈列表 | teacher/admin | 迭代7 |
| 109 | PUT | `/api/feedbacks/:id/status` | 更新反馈状态 | teacher/admin | 迭代7 |

### 7.21 发现页接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 110 | GET | `/api/discover` | 发现页数据 | 所有 | 迭代8 |
| 111 | GET | `/api/discover/detail` | 发现详情 | 所有 | 迭代8 |
| 112 | GET | `/api/discover/search` | 发现页搜索 | 所有 | 迭代8 |

### 7.22 通用文件上传接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 113 | POST | `/api/upload` | 通用文件上传 | 所有 | 迭代5 |
| 114 | POST | `/api/upload/h5` | H5文件上传 | 所有 | 迭代10 |

### 7.23 平台配置接口（无需鉴权）

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 115 | GET | `/api/platform/config` | 平台配置（小程序/H5差异） | 所有 | 迭代10 |

### 7.24 管理员接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 116 | GET | `/api/admin/dashboard/overview` | 管理员仪表盘总览 | admin | 迭代10 |
| 117 | GET | `/api/admin/dashboard/user-stats` | 用户统计 | admin | 迭代10 |
| 118 | GET | `/api/admin/dashboard/chat-stats` | 对话统计 | admin | 迭代10 |
| 119 | GET | `/api/admin/dashboard/knowledge-stats` | 知识库统计 | admin | 迭代10 |
| 120 | GET | `/api/admin/dashboard/active-users` | 活跃用户排行 | admin | 迭代10 |
| 121 | GET | `/api/admin/users` | 用户列表 | admin | 迭代10 |
| 122 | PUT | `/api/admin/users/:id/role` | 修改用户角色 | admin | 迭代10 |
| 123 | PUT | `/api/admin/users/:id/status` | 启禁用户 | admin | 迭代10 |
| 124 | GET | `/api/admin/feedbacks` | 反馈管理列表 | admin | 迭代10 |
| 125 | GET | `/api/admin/logs` | 操作日志查询 | admin | 迭代10 |
| 126 | GET | `/api/admin/logs/stats` | 日志统计 | admin | 迭代10 |
| 127 | GET | `/api/admin/logs/export` | 日志导出CSV | admin | 迭代10 |

### 7.25 系统接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 128 | GET | `/api/system/health` | 健康检查 | 所有 | V1.0 |
| 129 | GET | `/api/system/plugins` | 插件列表 | admin | V1.0 |
| 130 | GET | `/api/system/pipelines` | 管道列表 | admin | V1.0 |

### 7.26 Python LlamaIndex 服务接口（内部，端口8100）

| # | 方法 | 路径 | 说明 | 来源 |
|---|------|------|------|------|
| 131 | POST | `/api/v1/vectors/documents` | 存储文档向量（接收已分好的chunks） | 迭代5 |
| 132 | POST | `/api/v1/vectors/search` | 语义检索（返回top-k相似文档块） | 迭代5 |
| 133 | DELETE | `/api/v1/vectors/documents/{doc_id}` | 删除指定文档的所有向量 | 迭代5 |
| 134 | GET | `/api/v1/health` | 健康检查 | 迭代5 |

### 7.27 接口统计

| 分类 | 数量 |
|------|------|
| Go 后端接口（`/api/`） | 130 个 |
| Python 服务接口（`/api/v1/`） | 4 个 |
| **总计** | **134 个** |

| 按来源迭代统计 | 数量 |
|----------------|------|
| V1.0 | 12 |
| 迭代1 | 12 |
| 迭代2 | 14 |
| 迭代3 | 7 |
| 迭代4 | 6 |
| 迭代5 | 5 |
| 迭代6 | 7 |
| 迭代7 | 14 |
| 迭代8 | 22 |
| 迭代9 | 11 |
| 迭代10 | 20 |

| 按角色统计 | 数量 |
|------------|------|
| 所有角色可访问 | 48 |
| teacher 专属 | 52 |
| student 专属 | 3 |
| teacher/admin | 17 |
| admin 专属 | 14 |

---

## 8. 环境变量清单

| 变量名 | 必填 | 说明 | 来源迭代 |
|--------|------|------|----------|
| `JWT_SECRET` | ✅ | JWT签名密钥 | V1.0 |
| `DASHSCOPE_API_KEY` | ✅ | 通义千问API密钥 | V1.0 |
| `WX_APPID` | 生产必填 | 微信小程序AppID | 迭代2 |
| `WX_SECRET` | 生产必填 | 微信小程序AppSecret | 迭代2 |
| `WX_H5_APP_ID` | H5必填 | 微信公众号AppID | 迭代10 |
| `WX_H5_APP_SECRET` | H5必填 | 微信公众号AppSecret | 迭代10 |
| `WX_H5_REDIRECT_URI` | H5必填 | H5授权回调URL | 迭代10 |
| `WX_H5_MOCK_ENABLED` | ❌ | H5 Mock模式（开发环境） | 迭代10 |
| `UPLOAD_DIR` | ❌ | 文件上传目录（默认./uploads） | 迭代1 |
| `MAX_UPLOAD_SIZE` | ❌ | 最大上传大小（默认50MB） | 迭代1 |
| `LLAMAINDEX_URL` | ❌ | LlamaIndex服务地址（默认http://localhost:8100） | 迭代5 |
| `LOG_DB_PATH` | ❌ | 日志数据库路径 | 迭代10 |
| `LOG_RETENTION_DAYS` | ❌ | 日志保留天数（默认90） | 迭代10 |
| `LOG_CLEANUP_ENABLED` | ❌ | 是否启用日志清理 | 迭代10 |
| `LOG_CLEANUP_HOUR` | ❌ | 日志清理执行时间（默认凌晨3点） | 迭代10 |

---

## 9. Prompt体系

### 9.1 Prompt组装优先级（从高到低）

1. **安全规则** ← 最高优先级，不可覆盖（隐私防护+Injection防御）
2. **学段基础模板** ← 硬约束，决定语言风格和回答深度（8档学段）
3. **教学风格模板** ← 软约束，决定教学方法（6种风格）
4. **个性化教学要求** ← 教师自定义补充
5. **教学背景** ← 教材版本、学科、进度
6. **学生个人信息** ← 用于个性化服务（审批时填写的信息）
7. **用户画像** ← 记忆合并时提炼的画像
8. **行为约束** ← 知识库为空时的行为规则
9. **相关知识 + 学生记忆** ← 动态内容（RAG检索+记忆检索）

### 9.2 学段划分（8档）

| 学段标识 | 名称 | 年级范围 |
|---------|------|---------|
| `preschool` | 学前班 | 幼儿园大班~学前 |
| `primary_lower` | 小学低年级 | 1-3年级 |
| `primary_upper` | 小学高年级 | 4-6年级 |
| `junior` | 初中 | 7-9年级 |
| `senior` | 高中 | 10-12年级 |
| `university` | 大学及以上 | 大学/研究生/博士 |
| `adult_life` | 成人生活技能 | 烹饪/健身/手工等 |
| `adult_professional` | 成人职业培训 | 职业技能/考证等 |

### 9.3 对话风格模板（6种）

| 风格 | 说明 |
|------|------|
| 苏格拉底式 | 通过提问引导学生自主思考 |
| 直接讲解 | 直接给出清晰的解释和答案 |
| 鼓励引导 | 多用鼓励性语言，注重正向反馈 |
| 严格训练 | 高标准要求，注重准确性 |
| 游戏化 | 趣味互动，适合低龄学生 |
| 自适应 | 根据学生表现自动调整风格 |

---

## 10. 已移除的功能

| 功能 | 原迭代 | 移除迭代 | 原因 |
|------|--------|----------|------|
| 作业系统（提交/列表/点评） | 迭代1 | 迭代7 | 功能优先级低，改为对话中附件提交 |
| 记忆衰减机制 | 原规划迭代2 | — | 改为三层分层存储+摘要合并 |
| 数据分析看板（教师端独立页） | 原规划迭代2 | — | 改为Dashboard概览+管理员后台统计 |
| 对话导出TXT | 原规划迭代2 | — | 优先级低，未实现 |

---

## 11. 技术债务与已知问题

| 问题 | 说明 | 优先级 |
|------|------|--------|
| InMemoryVectorStore残留 | 已被LlamaIndex替代，代码保留作降级方案 | 低 |
| 教师端工作台页面 | TabBar已移除工作台入口，但home页面代码仍存在 | 低 |
| 微信订阅消息 | 需要微信后台配置消息模板，当前为Mock | 中 |
| 小程序审核发布 | 域名备案、服务器域名配置、微信提审 | 高 |
| H5端语音输入 | H5不支持微信语音SDK，暂不可用 | 低 |

---

## 12. 需求追溯

| 需求来源 | 事项 | 实现迭代 |
|----------|------|----------|
| 用户需求 | 师生授权机制 | 迭代1 |
| 用户需求 | 教师注册增强 | 迭代1→迭代8简化 |
| 用户需求 | 教师评语/学生备注 | 迭代1→迭代5改名 |
| 用户需求 | 个性化问答风格 | 迭代1 |
| 全局需求 | 文档上传PDF/DOCX | 迭代1 |
| 全局需求 | URL导入 | 迭代1 |
| 全局需求 | SSE流式输出 | 迭代1 |
| 架构需求 | 多角色多分身体系 | 迭代2 |
| 架构需求 | 班级管理 | 迭代2→迭代8增强 |
| UX需求 | 教师Dashboard | 迭代3→迭代5重构 |
| UX需求 | 启停管理 | 迭代3 |
| 用户需求 | 分身广场 | 迭代4→迭代5独立化 |
| 用户需求 | 教师真人介入 | 迭代4 |
| 架构需求 | LlamaIndex语义检索 | 迭代5 |
| UX需求 | 学生管理合并 | 迭代5 |
| 全局需求 | 记忆三层分层 | 迭代6 |
| 全局需求 | 对话风格模板 | 迭代6 |
| 生产必备 | Docker部署 | 迭代6 |
| 用户需求 | 教材配置体系 | 迭代7 |
| 安全需求 | API限流+隐私防护 | 迭代7 |
| 用户需求 | 批量添加学生 | 迭代7 |
| UX需求 | 知识库统一输入框 | 迭代8 |
| UX需求 | 仿微信聊天改版 | 迭代8 |
| 用户需求 | 思考过程展示 | 迭代9 |
| 用户需求 | 会话列表优化 | 迭代9 |
| 用户需求 | 课程发布 | 迭代9 |
| 运营需求 | 管理员H5后台 | 迭代10 |
| 多端需求 | 教师/学生H5页面 | 迭代10 |
| 运营需求 | 操作日志流水 | 迭代10 |
| 用户需求 | 教师端TabBar调整（聊天列表替换工作台） | 迭代后变更 |

---

**文档版本**: v2.0.0
**创建日期**: 2026-03-28
**最后更新**: 2026-04-04
**维护说明**: 本文档为 V2.0 全量需求汇总，各迭代详细需求见 `iteration{N}_requirements.md`，API详细规范见 `iteration{N}_api_spec.md`
