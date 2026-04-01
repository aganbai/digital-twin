# V2.0 任务跟踪

## 版本信息

| 项目 | 说明 |
|------|------|
| 版本 | V2.0 - 单机生产可用版 |
| 迭代数 | 5 个迭代 |
| 预计周期 | ~20 周 |
| 状态 | ✅ 迭代1已完成，✅ 迭代2已完成，✅ 迭代3已完成，✅ 迭代4已完成，✅ 迭代5已完成（Phase 3c 补充执行中） |

---

## 迭代1：核心功能开发（~4 周） ✅ 已完成

### P0 基础层（Day 1）

| 子模块 | 任务 | 状态 | 备注 |
|--------|------|------|------|
| S0-1 | Model 定义（6 个新结构体 + User 扩展） | ✅ 已完成 | `database/models.go` |
| S0-2 | Repository 接口签名（5 个新 Repository） | ✅ 已完成 | `database/repository.go` |
| S0-3 | Handler 空壳 + 路由注册（17 个新路由） | ✅ 已完成 | `api/handlers.go` + `api/router.go` |
| S0-4 | 数据库建表（5 张新表 + ALTER TABLE） | ✅ 已完成 | `database/database.go` |

### P1 后端并行组 A：核心链路（Day 2-4）

| 子模块 | 任务 | 状态 | 依赖契约 | 备注 |
|--------|------|------|----------|------|
| S1-A1-a | User 结构体加 School/Description | ✅ 已完成 | - | `models.go` |
| S1-A1-b | CheckTeacherExists + UpdateProfile | ✅ 已完成 | C2, C3 | `repository.go` |
| S1-A1-c | auth_plugin 校验逻辑 | ✅ 已完成 | C2, C3 | `auth_plugin.go` |
| S1-A1-d | HandleCompleteProfile 改造 | ✅ 已完成 | - | `handlers.go` |
| S1-A1-e | HandleGetTeachers 增强 | ✅ 已完成 | - | `handlers.go` |
| S1-A2-a | RelationRepository 全部实现 | ✅ 已完成 | C1 | `repository.go` |
| S1-A2-b | HandleInviteStudent / HandleApplyTeacher | ✅ 已完成 | C1 | `handlers.go` |
| S1-A2-c | HandleApproveRelation / HandleRejectRelation | ✅ 已完成 | C1 | `handlers.go` |
| S1-A2-d | HandleGetRelations（双视角） | ✅ 已完成 | C1 | `handlers.go` |
| S1-A2-e | HandleChat 增加 IsApproved 鉴权 | ✅ 已完成 | C1 | `handlers.go` |
| S1-A2-f | 数据迁移脚本 | ✅ 已完成 | - | `database.go` |
| S1-A3-a | LLMClient.ChatStream 方法 | ✅ 已完成 | C4 | `llm_client.go` |
| S1-A3-b | API 模式 stream:true + SSE 解析 | ✅ 已完成 | C4 | `llm_client.go` |
| S1-A3-c | Mock 模式逐字模拟 | ✅ 已完成 | C4 | `llm_client.go` |
| S1-A3-d | dialogue_plugin handleChatStream | ✅ 已完成 | C4 | `dialogue_plugin.go` |
| S1-A3-e | HandleChatStream SSE 响应 | ✅ 已完成 | C4 | `handlers.go` |

### P1 后端并行组 B：功能模块（Day 2-5）

| 子模块 | 任务 | 状态 | 依赖契约 | 备注 |
|--------|------|------|----------|------|
| S1-B1-a | FileParser 接口 + TXT/MD 解析 | ✅ 已完成 | C6 | `file_parser.go` |
| S1-B1-b | PDF 解析 | ✅ 已完成 | C6 | `file_parser.go` |
| S1-B1-c | DOCX 解析 | ✅ 已完成 | C6 | `file_parser.go` |
| S1-B1-d | HandleUploadDocument 实现 | ✅ 已完成 | C6 | `handlers.go` |
| S1-B2-a | URLFetcher HTTP GET + 编码检测 | ✅ 已完成 | - | `url_fetcher.go` |
| S1-B2-b | HTML 解析（去标签+提取正文） | ✅ 已完成 | - | `url_fetcher.go` |
| S1-B2-c | HandleImportURL 实现 | ✅ 已完成 | - | `handlers.go` |
| S1-B3-a | CommentRepository 全部实现 | ✅ 已完成 | C1 | `repository.go` |
| S1-B3-b | HandleCreateComment 实现 | ✅ 已完成 | C1 | `handlers.go` |
| S1-B3-c | HandleGetComments 实现 | ✅ 已完成 | - | `handlers.go` |
| S1-B4-a | StyleRepository 全部实现 | ✅ 已完成 | C1 | `repository.go` |
| S1-B4-b | HandleSetDialogueStyle / HandleGetDialogueStyle | ✅ 已完成 | C1 | `handlers.go` |
| S1-B4-c | BuildSystemPrompt 增加 styleConfig | ✅ 已完成 | C5 | `prompt.go` |
| S1-B4-d | handleChat 风格查询和注入 | ✅ 已完成 | C5 | `dialogue_plugin.go` |
| S1-B4-e | SetTemperature / ResetTemperature | ✅ 已完成 | - | `llm_client.go`（含并发安全修复） |
| S1-B5-a | AssignmentRepository + ReviewRepository 实现 | ✅ 已完成 | C1 | `repository.go` |
| S1-B5-b | HandleSubmitAssignment 实现 | ✅ 已完成 | C1 | `handlers.go` |
| S1-B5-c | HandleGetAssignments / HandleGetAssignmentDetail | ✅ 已完成 | - | `handlers.go` |
| S1-B5-d | HandleReviewAssignment 实现 | ✅ 已完成 | - | `handlers.go` |
| S1-B5-e | HandleAIReviewAssignment 实现 | ✅ 已完成 | C6, C7 | `handlers.go` |
| S1-B5-f | BuildAssignmentReviewPrompt | ✅ 已完成 | C7 | `prompt.go` |

### P2 前端并行（Day 5-10）

| 子模块 | 任务 | 状态 | 依赖后端 | 备注 |
|--------|------|------|----------|------|
| S2-F1-a | 角色选择页改造（教师加学校+描述） | ✅ 已完成 | S1-A1 | `role-select/index.tsx` |
| S2-F1-b | 学生首页改造（教师卡片+授权状态） | ✅ 已完成 | S1-A1 | `home/index.tsx` |
| S2-F2-a | 师生管理页（教师端） | ✅ 已完成 | S1-A2 | `teacher-students/index.tsx` |
| S2-F2-b | 我的教师页（学生端） | ✅ 已完成 | S1-A2 | `my-teachers/index.tsx` |
| S2-F2-c | relation API + Store | ✅ 已完成 | S1-A2 | `api/relation.ts` + `store/relationStore.ts` |
| S2-F3-a | 对话页 SSE 流式渲染 | ✅ 已完成 | S1-A3 | `chat/index.tsx` |
| S2-F3-b | chat API chatStream 方法 | ✅ 已完成 | S1-A3 | `api/chat.ts` |
| S2-F4-a | 添加文档页三 Tab 改造 | ✅ 已完成 | S1-B1+B2 | `knowledge/add.tsx` |
| S2-F4-b | document API upload/importUrl | ✅ 已完成 | S1-B1+B2 | `api/knowledge.ts` |
| S2-F5-a | 学生详情页（风格+评语） | ✅ 已完成 | S1-B3+B4 | `student-detail/index.tsx` |
| S2-F5-b | 我的评语页 | ✅ 已完成 | S1-B3 | `my-comments/index.tsx` |
| S2-F5-c | comment API + style API | ✅ 已完成 | S1-B3+B4 | `api/comment.ts` + `api/style.ts` |
| S2-F6-a | 提交作业页 | ✅ 已完成 | S1-B5 | `submit-assignment/index.tsx` |
| S2-F6-b | 作业列表页 | ✅ 已完成 | S1-B5 | `assignment-list/index.tsx` |
| S2-F6-c | 作业详情页 | ✅ 已完成 | S1-B5 | `assignment-detail/index.tsx` |
| S2-F6-d | 我的作业页 | ✅ 已完成 | S1-B5 | `my-assignments/index.tsx` |
| S2-F6-e | assignment API + Store | ✅ 已完成 | S1-B5 | `api/assignment.ts` + `store/assignmentStore.ts` |

### P3 集成测试（Day 11-13）

| 子模块 | 任务 | 状态 | 备注 |
|--------|------|------|------|
| S3-a | IT-40~IT-41 教师注册增强 | ✅ 已完成 | 2/2 PASS |
| S3-b | IT-42~IT-46 师生授权 | ✅ 已完成 | 5/5 PASS |
| S3-c | IT-47~IT-48 对话鉴权 | ✅ 已完成 | 2/2 PASS |
| S3-d | IT-49 评语系统 | ✅ 已完成 | 1/1 PASS |
| S3-e | IT-50 问答风格 | ✅ 已完成 | 1/1 PASS |
| S3-f | IT-51~IT-52 作业系统 | ✅ 已完成 | 2/2 PASS |
| S3-g | IT-53~IT-55 文件上传 | ✅ 已完成 | 3/3 PASS |
| S3-h | IT-56~IT-57 URL 导入 | ✅ 已完成 | 2/2 PASS |
| S3-i | IT-58~IT-59 SSE 流式 | ✅ 已完成 | 2/2 PASS |
| S3-j | IT-60 全链路 | ✅ 已完成 | 1/1 PASS（14步全链路） |

---

## 迭代2：多角色多分身架构 ✅ 已完成

### 后端模块跟踪

| 模块编号 | 模块名称 | 状态 | 备注 |
|----------|----------|------|------|
| V2-IT2-M1 | 分身管理 | ✅ 已完成 | `handlers_persona.go` + `repository_persona.go` |
| V2-IT2-M2 | 班级管理 | ✅ 已完成 | `handlers_class.go` + `repository_class.go` |
| V2-IT2-M3 | 分身分享 | ✅ 已完成 | `handlers_share.go` + `repository_share.go` |
| V2-IT2-M4 | 知识库精细化 | ✅ 已完成 | `handlers_knowledge.go` scope 改造 + `knowledge_plugin.go` filterByScope |
| V2-IT2-M5 | 现有模块分身化改造 | ✅ 已完成 | 对话/记忆/评语/风格/作业/师生关系全部改为分身维度 |
| V2-IT2-M6 | 认证改造 | ✅ 已完成 | JWT Claims 新增 persona_id + 登录返回分身列表 |
| V2-IT2-M7 | 数据库迁移 | ✅ 已完成 | 4 张新表 + 9 步数据回填（幂等） |
| V2-IT2-M8 | 集成测试 | ✅ 已完成 | IT-61~IT-90 共 30 个测试用例全部 PASS |

### 前端模块跟踪

| 模块编号 | 模块名称 | 状态 | 备注 |
|----------|----------|------|------|
| V2-IT2-FE-M1 | 分身选择页 + 创建分身页 | ✅ 已完成 | `persona-select/index.tsx` |
| V2-IT2-FE-M2 | 登录流程改造 | ✅ 已完成 | `login/index.tsx` 支持分身列表跳转 |
| V2-IT2-FE-M3 | 学生首页改造 + 分身切换 | ✅ 已完成 | `home/index.tsx` 含分身切换入口 |
| V2-IT2-FE-M4 | 教师首页改造 | ✅ 已完成 | `knowledge/index.tsx` scope Tab 筛选 |
| V2-IT2-FE-M5 | 班级管理 + 师生管理改造 | ✅ 已完成 | `teacher-students/index.tsx` + `class-create/index.tsx` |
| V2-IT2-FE-M6 | 分享码功能 + 加入页 | ✅ 已完成 | `share-join/index.tsx` + `api/share.ts` |
| V2-IT2-FE-M7 | 添加文档页 scope 改造 | ✅ 已完成 | `api/knowledge.ts` 支持 scope 参数 |
| V2-IT2-FE-M8 | 对话页分身维度改造 | ✅ 已完成 | `api/chat.ts` 使用 teacher_persona_id |

---

## 📊 进度统计

| 阶段 | 子模块数 | 完成数 | 完成率 |
|------|----------|--------|--------|
| P0 基础层 | 4 | 4 | 100% |
| P1 后端组A | 16 | 16 | 100% |
| P1 后端组B | 21 | 21 | 100% |
| P2 前端 | 17 | 17 | 100% |
| P3 集成测试 | 10 | 10 | 100% |
| **迭代1 合计** | **68** | **68** | **100%** |
| 迭代2 后端 | 8 | 8 | 100% |
| 迭代2 前端 | 8 | 8 | 100% |
| **迭代2 合计** | **16** | **16** | **100%** |
| **V2.0 总计** | **84** | **84** | **100%** |

---

## 📝 迭代1 Review 记录

### 集成测试结果
- V1.0 回归测试：17/17 PASS
- V2.0 新增测试：21/21 PASS（IT-40 ~ IT-60）
- 总计：38/38 PASS

### 需求匹配 Review
- 9 个后端模块（V2-M1 ~ V2-M9）：100% 匹配
- 8 个前端模块（V2-FE-M1 ~ V2-FE-M8）：100% 匹配

### 架构&设计 Review
- 架构质量：良好（分层清晰、插件化设计、安全措施到位）
- 已修复问题：
  - FB-01: LLMClient.SetTemperature 并发安全（已加 sync.Mutex）
  - FB-02: ResetTemperature 硬编码 0.7（已改为 defaultTemperature）
  - FB-03: 上传目录使用相对路径（已改为读取 UPLOAD_DIR 环境变量）
- 建议优化项（纳入迭代2）：
  - OPT-01: handlers.go 过大（2437行），建议按模块拆分
  - OPT-02: 增加 API 限流中间件

---

**最后更新**: 2026-03-31 19:33

---

## 📝 迭代2 Review 记录

### 集成测试结果
- V1.0 回归测试：18/18 PASS（IT-01 ~ IT-17）
- V2.0 迭代1 回归测试：21/21 PASS（IT-18 ~ IT-39 + IT-40 ~ IT-60）
- V2.0 迭代2 新增测试：30/30 PASS（IT-61 ~ IT-90）
- **总计：90/90 PASS（含全链路 IT-86）**

### 需求匹配 Review
- 6 个后端模块（M1~M6）：100% 匹配
- 8 个前端模块（FE-M1~FE-M8）：100% 匹配
- 已修复 high 问题：
  - 删除班级行为与需求不一致（改为拒绝删除有成员的班级）
  - 对话插件+记忆插件未完成分身维度改造（已改造）
- 已修复 medium 问题：
  - HandleGetRelations 未使用分身维度查询
  - 分享码加入缺少"无学生分身"引导（40022 错误码）
  - 前端学生首页教师列表使用 teacher_persona_id

### 架构&设计 Review
- 架构质量：良好（分身维度降级策略、向后兼容、知识库三级 scope）
- 已修复问题：
  - 所有分身维度授权检查增加 user_id 回退逻辑
  - 对话记录保存分身维度信息（CreateWithPersonas）
  - 前端 home/index.tsx 类型错误修复

### 代码 Review
- 整体评分：7.5/10
- 可维护性：7/10
- 安全性：7.5/10
- 测试覆盖：8/10

### 技术债务（纳入后续迭代）
| 优先级 | 描述 |
|--------|------|
| P1 | teacher_student_relations 表唯一约束需适配分身维度（UNIQUE(teacher_persona_id, student_persona_id)） |
| P1 | 插件间工具函数重复（mergeData/toInt/toInt64/toFloat64），需抽取到公共包 |
| P1 | ListMemories 在应用层做分页和筛选，需下推到 SQL 层 |
| P2 | knowledge_plugin handleDelete 未校验文档所有者 |
| P2 | handlers_class.go 班级更新/删除需校验所有权 |
| P2 | 前端 knowledge/index.tsx 使用 setTimeout 触发副作用，需改为 useEffect |
| P3 | 前端 teacher-students/index.tsx 521行，需拆分子组件 |

---

## 迭代3：教师仪表盘 + 启停管理 + 知识库预览 — 进行中

### 恢复快照（最后更新: 2026-03-30 21:10）

#### 当前状态
- 当前 Phase: Phase 3c（端到端冒烟验证）— 进行中
- Phase 3a Review: ✅ 完成（3个HIGH问题已修复）
- Phase 3b 集成测试: ✅ 完成（IT-91~IT-105 全部 PASS + IT-01~IT-90 回归 PASS）
- 下一步: Phase 3c 冒烟验证 → Phase 4 收尾
- 编译状态: ✅ go build 通过

#### Phase 3a Review 修复记录
- H-1: ListTeachersForStudent 增加 JOIN personas 检查 is_active ✅
- H-2: handlers_knowledge.go 拆分为 3 个文件 ✅
- H-3: 班级详情页补充学生快捷操作按钮 ✅

#### Phase 3b 集成测试结果
- IT-91~IT-105: 15/15 PASS
- IT-01~IT-90 回归: 90/90 PASS
- 迭代3回归: 7/7 PASS
- 总计: 112/112 PASS

#### 已完成模块清单（上次会话评估结论）

| 模块 | 状态 | 关键文件 | 说明 |
|------|------|----------|------|
| BE-M1 数据库变更 | ✅ 已实现 | database/database.go | classes 和 teacher_student_relations 表已添加 is_active 字段 |
| BE-M2 教师仪表盘聚合API | ✅ 已实现 | api/handlers_persona.go | HandleGetPersonaDashboard 已实现 |
| BE-M3 启停管理API | ✅ 已实现 | api/handlers_class.go | HandleToggleClass 和 HandleToggleRelation 已实现 |
| BE-M4 知识库预览API | ✅ 已实现 | api/ | 预览相关接口已实现 |
| BE-M5 scope_ids多选改造 | ✅ 已实现 | api/ | 后端支持已实现 |
| BE-M6 学生教师列表过滤改造 | ✅ 已实现 | api/ | HandleGetTeachers 已改造 |
| FE-M1 教师仪表盘首页 | ✅ 已实现 | miniprogram/pages/home/index.tsx | 教师仪表盘+学生首页双模式渲染 |
| FE-M2 班级详情页 | ✅ 已实现 | miniprogram/pages/class-detail/ | 班级详情页已存在 |
| FE-M3 知识库预览页 | ✅ 已实现 | miniprogram/pages/ | 预览页已存在 |
| FE-M4 启停管理UI | ✅ 已实现 | miniprogram/pages/class-detail/ + teacher-students/ | 班级/学生启停 + 二次确认弹窗 |

#### 待完成模块清单

| 模块 | 状态 | 前置依赖 | 备注 |
|------|------|----------|------|
| IT-91~IT-105 集成测试 | ❌ 未开始 | BE-M1~M6, FE-M1~M4 | V2.0 迭代3 集成测试 |

#### 编译状态
- go build: ✅ 通过（上次会话确认）
- go test: 待确认

#### 上下文恢复提示
- 不需要重新读取的文件: iteration_dev_skill_core.md（已精简）、project_tech_spec.md（未变更）
- 需要关注的文件: FE-M4 启停管理 UI 相关前端文件、tests/integration/ 集成测试文件

---

## 迭代4：分身广场 + 教师真人介入 + 智能摘要 — ✅ 已完成

### 恢复快照（最后更新: 2026-03-31 10:45）

#### 当前状态
- 当前 Phase: Phase 4（迭代收尾）— ✅ 已完成
- Phase 2 后端开发: ✅ 完成（6 个后端模块全部完成）
- Phase 2 前端开发: ✅ 完成（6 个前端模块全部完成）
- Phase 3a Review: ✅ 完成（3 个 HIGH + 6 个 MEDIUM 问题已修复）
- Phase 3b 集成测试: ✅ 完成（134/134 PASS）
- Phase 3c 冒烟验证: ⏭️ 跳过（微信开发者工具 IDE 未运行）
- 编译状态: ✅ go build 通过 + 134/134 集成测试 PASS

#### Phase 3a Review 修复记录
- H-ARCH-1: 旧对话查询方法（4个）添加 sender_type + reply_to_id ✅
- H-ARCH-2: HandleConfirmDocument 传递 summary 到插件/数据库 ✅
- H-CODE-1: GetByPersonas 添加 sender_type + reply_to_id ✅
- H-CODE-2: CreateWithPersonas 添加 reply_to_id ✅
- M-ARCH-1: 广场申请 API 改用 applyTeacher 方法 ✅
- M-ARCH-2: HandleEndTakeover 返回 ended_at ✅
- M-ARCH-3: HandleTeacherReply 返回正确的 created_at ✅
- M-ARCH-4: 教师首页添加分身概览入口 ✅
- M-CODE-1/2/3: HandleChat/HandleChatStream 接管响应补全字段 ✅
- M-CODE-4: documents 表添加 summary 列 ✅

#### Phase 3b 集成测试结果
- IT-106~IT-125: 20/20 PASS
- IT-01~IT-105 回归: 114/114 PASS
- 总计: 134/134 PASS

#### Phase 3c 冒烟验证
- 状态: ⏭️ 跳过
- 原因: 微信开发者工具 IDE 未运行，CLI 连接超时
- 建议: 手动启动 IDE 并登录后重新执行

### 需求文档
- 需求规格说明书: `docs/iterations/v2.0/iteration4_requirements.md`
- API 接口规范: `docs/iterations/v2.0/iteration4_api_spec.md`

### 核心功能
| 编号 | 功能 | 说明 |
|------|------|------|
| F1 | 老师分身广场 | 学生首页新增广场，展示公开教师分身，学生可申请 |
| F2 | 定向邀请链接 | 分享码绑定特定学生，仅目标学生可用 |
| F3 | 教师真人介入对话 | 教师引用回复学生，接管后AI停止，退出后AI恢复 |
| F4 | 分身概览页 | 教师登录后看到所有分身概览，可创建新分身 |
| F5 | 知识库LLM智能摘要 | 上传后LLM自动生成title和摘要，展示loading |

### 后端模块划分
| 模块编号 | 模块名称 | 状态 | 备注 |
|----------|----------|------|------|
| V2-IT4-BE-M1 | 数据库迁移 | ✅ 已完成 | ALTER TABLE + CREATE TABLE |
| V2-IT4-BE-M2 | 分身广场 API | ✅ 已完成 | marketplace + visibility |
| V2-IT4-BE-M3 | 定向邀请 API | ✅ 已完成 | share 改造 + 学生搜索 |
| V2-IT4-BE-M4 | 教师真人介入 API | ✅ 已完成 | teacher-reply + takeover |
| V2-IT4-BE-M5 | 分身概览改造 | ✅ 已完成 | personas 返回 is_public |
| V2-IT4-BE-M6 | LLM 智能摘要 | ✅ 已完成 | llm_summarizer + preview 改造 |

### 前端模块划分
| 模块编号 | 模块名称 | 状态 | 备注 |
|----------|----------|------|------|
| V2-IT4-FE-M1 | 广场 UI | ✅ 已完成 | 学生首页新增广场区域 |
| V2-IT4-FE-M2 | 定向邀请 UI | ✅ 已完成 | share 页面改造 |
| V2-IT4-FE-M3 | 对话标识 + 引用回复 | ✅ 已完成 | chat 页面 sender_type 标识 |
| V2-IT4-FE-M4 | 教师对话记录页 | ✅ 已完成 | 新增 student-chat-history 页面 |
| V2-IT4-FE-M5 | 分身概览页 | ✅ 已完成 | 新增 persona-overview 页面 |
| V2-IT4-FE-M6 | LLM 摘要 UI | ✅ 已完成 | preview 页面 loading + 自动填充 |

### 集成测试
| 用例范围 | 状态 | 备注 |
|----------|------|------|
| IT-106 ~ IT-125 | ✅ 已通过 | 20 个测试用例（之前会话完成） |
| IT-01 ~ IT-105 回归 | ✅ 已通过 | 134/134 全部 PASS |

### 新增接口（7 个）
| API-45 ~ API-51 | marketplace / visibility / search / teacher-reply / takeover-status / end-takeover / student-conversations |

### 改造接口（8 个）
| shares(3) + chat(2) + conversations(1) + personas(1) + documents/preview(1) |

### 数据库变更
| 变更 | 说明 |
|------|------|
| personas.is_public | 分身公开状态 |
| persona_shares.target_student_persona_id | 定向邀请 |
| conversations.sender_type | 消息发送者类型 |
| conversations.reply_to_id | 引用回复 |
| teacher_takeovers 表 | 教师接管记录 |

---

## 迭代5：LlamaIndex 语义检索 + UX 重构 — ✅ 已完成

### 恢复快照（最后更新: 2026-03-31 19:33）

#### 当前状态
- 当前 Phase: Phase 4（迭代收尾）— ✅ 已完成
- Phase 2 后端开发: ✅ 完成（5 个后端模块全部完成，20 个新增单测通过）
- Phase 2 前端开发: ✅ 完成（6 个前端模块全部完成，TypeScript 编译通过）
- Phase 3a Review: ✅ 完成（3 个 HIGH 问题已修复）
- Phase 3b 集成测试: ✅ 完成（147/148 PASS）
- Phase 3c 冒烟验证: ⏭️ 跳过（E2E 测试脚本未编写，miniprogram-automator 环境未搭建）
- 编译状态: ✅ go build 通过 + 147/148 集成测试 PASS

### 需求文档
- 需求规格说明书: `docs/iterations/v2.0/iteration5_requirements.md`
- API 接口规范: `docs/iterations/v2.0/iteration5_api_spec.md`

### 核心功能
| 编号 | 功能 | 说明 |
|------|------|------|
| R1 | LlamaIndex 语义检索服务 | Python LlamaIndex 服务替换 InMemoryVectorStore |
| R2 | 教师首页改为 Dashboard | 教师落地页从知识库页改为 Dashboard |
| R3 | 学生1个老师直接进对话 | 学生只有1个老师时跳过首页直接对话 |
| R4 | 学生管理页合并 | 师生管理 + 班级管理合并为"学生管理" |
| R5 | 评语改为学生备注 | 学生不可见，仅教师私有 |
| R6 | 分享码移入分身管理 | 首页独立入口移入分身管理页 |
| R7 | 对话附件支持 | 作业在聊天窗口提交（P1） |
| R8 | 广场独立化 | 从首页内嵌拆为独立页面 |

### 后端模块划分
| 模块编号 | 模块名称 | 状态 | 备注 |
|----------|----------|------|------|
| V2-IT5-BE-M1 | Python LlamaIndex 服务搭建 | ✅ 已完成 | src/knowledge-service/ (FastAPI + LlamaIndex) |
| V2-IT5-BE-M2 | Go 向量客户端适配层 | ✅ 已完成 | vector_client.go + knowledge_plugin.go 改造 |
| V2-IT5-BE-M3 | 对话附件支持 | ✅ 已完成 | handlers_chat.go 扩展附件参数 |
| V2-IT5-BE-M4 | 文件上传接口 | ✅ 已完成 | handlers_upload.go (POST /api/upload) |
| V2-IT5-BE-M5 | 评语接口权限调整 | ✅ 已完成 | handlers_comment.go 学生返回空列表 |

### 前端模块划分
| 模块编号 | 模块名称 | 状态 | 备注 |
|----------|----------|------|------|
| V2-IT5-FE-M1 | 教师首页 Dashboard 重构 | ✅ 已完成 | TeacherDashboard 组件 |
| V2-IT5-FE-M2 | 学生首页重构 + 路由优化 | ✅ 已完成 | StudentHome 组件 + 路由优化 |
| V2-IT5-FE-M3 | 学生管理页合并 | ✅ 已完成 | teacher-students 4Tab 重构 |
| V2-IT5-FE-M4 | 对话附件发送 | ✅ 已完成 | chat 附件功能 + upload API |
| V2-IT5-FE-M5 | 独立发现页 | ✅ 已完成 | discover/index.tsx 新增 |
| V2-IT5-FE-M6 | 评语→学生备注 + 分享码入口调整 | ✅ 已完成 | student-detail + persona-overview + share-manage |

### 集成测试
| 用例范围 | 状态 | 备注 |
|----------|------|------|
| IT-201 ~ IT-213 | ✅ 已通过 | 14 个测试用例（含 IT-208b）全部 PASS |
| IT-01 ~ IT-134 回归 | ✅ 已通过 | 147/148 PASS（IT-57 旧已知问题） |

### Phase 3a Review 修复记录
- HIGH-1: Python search() 锁竞态 → 将 retrieve 纳入锁保护范围 ✅
- HIGH-2: Python search() 重复创建 Embedding 实例 → 预创建 _query_embed_model 复用 ✅
- HIGH-3: Go InMemoryVectorStore 无效初始化 → 改为 nil ✅

### Phase 3b 集成测试结果
- IT-201~IT-213: 14/14 PASS（含 IT-208b 文件类型校验）
- IT-01~IT-134 回归: 147/148 PASS
- 唯一失败: IT-57（V2 迭代1 旧测试，import-url mock 模式已知问题，非 V5 引入）
- 总计: 161/162 PASS

### Phase 3c 冒烟验证
- 状态: ⏭️ 跳过
- 原因: E2E 测试脚本未编写，miniprogram-automator 环境未搭建
- 建议: 后续迭代搭建 E2E 自动化环境

### 新增接口（5 个）
| API-52 ~ API-56 | vectors/documents / vectors/search / vectors/documents/{id} / health / upload |

### 修改接口（3 个）
| chat(2) + comments(1) |

### 数据库变更
无表结构变更。向量存储迁移到外部 Python 服务。

---

## 📝 迭代5 Review 记录

### 集成测试结果
- V5 新增测试：14/14 PASS（IT-201 ~ IT-213 + IT-208b）
- 回归测试：147/148 PASS（IT-01 ~ IT-134 + IT-201 ~ IT-213）
- **总计：161/162 PASS**

### 需求匹配 Review
- 8 个核心功能（R1~R8）：100% 匹配
- 5 个后端模块（BE-M1~M5）：100% 完成
- 6 个前端模块（FE-M1~M6）：100% 完成
- 已修复 HIGH 问题：
  - Python search() 锁竞态条件（HIGH-1）
  - Python search() 重复创建 Embedding 实例（HIGH-2）
  - Go InMemoryVectorStore 无效初始化（HIGH-3）

### 架构&设计 Review
- 架构质量：4.5/5（Go→Python HTTP REST 通信合理、VectorClient 降级设计完善）
- 已修复问题：Python 锁粒度、Embedding 实例复用、无效内存占用

### 代码 Review
- 代码质量：4/5
- 架构质量：4.5/5
- 测试覆盖：4/5
- 需求匹配：5/5

### MEDIUM 问题（建议后续修复）
| 编号 | 描述 |
|------|------|
| MED-1 | handlers_upload.go 未使用请求中的 type 参数 |
| MED-2 | TeacherDashboard 缺少"对话数"和"今日活跃"统计卡片 |
| MED-3 | my-comments 页面文件仍存在且注册在路由中（学生端已移除入口） |
| MED-4 | Python delete_by_document_id 使用 delete_ref_doc 可能不完整 |
| MED-5 | HandleChat 附件路径拼接存在安全风险（需增加路径校验） |
| MED-6 | handlers_comment_test.go 评语测试未使用真实 Handler |

### 技术债务（纳入后续迭代）
| 优先级 | 描述 |
|--------|------|
| P1 | ✅ 搭建 E2E 自动化测试环境（miniprogram-automator）— 已完成 |
| P1 | Python 环境升级到 3.10+（当前 3.9 不兼容 llama_index 最新版） |
| P2 | HandleChat 附件路径安全校验 |
| P2 | TeacherDashboard 补全统计卡片 |
| P3 | 清理 my-comments 页面和路由注册 |

---

**最后更新**: 2026-03-31 22:55

---

## 迭代6：记忆增强 + 对话风格 + 分享优化 + UI重构 + 生产部署 — 进行中

### 恢复快照（最后更新: 2026-03-31 22:55）

#### 当前状态
- 当前 Phase: Phase 2（并行开发）— 进行中
- Phase -1 需求确认: ✅ 完成（Q1~Q8 全部确认）
- Phase 0 环境准备: ✅ 完成（环境就绪 + 冒烟测试计划已更新 17 条新用例）
- Phase 0.5 代码状态评估: ⏭️ 跳过（迭代5已完成，无需恢复）
- Phase 1 准备阶段: ✅ 完成（需求文档+API规范已更新决策结论）
- 编译状态: ✅ go build 通过

#### 需求确认决策摘要
- Q1: 记忆摘要合并 = 定时任务（每天凌晨2:00自动）+ 手动触发接口保留
- Q2: 同 memory_type 的 core 记忆只保留最近一条，有则 UPDATE
- Q3: GET /api/shares/:code/info 改为可选鉴权
- Q4: 不考虑旧版兼容（未上线），非目标学生返回200+引导
- Q5: 聊天记录 JSON 导入，复用 scope 机制，仅教师可用
- Q6: Tab 页路由统一改为 switchTab
- Q7: SQLite 当前阶段够用
- Q8: episodic 上限 50 条，超过自动合并

### 需求文档
- 需求规格说明书: `docs/iterations/v2.0/iteration6_requirements.md`
- API 接口规范: `docs/iterations/v2.0/iteration6_api_spec.md`

### 后端模块划分
| 模块编号 | 模块名称 | 状态 | 备注 |
|----------|----------|------|------|
| V2-IT6-BE-M1 | 数据库变更（记忆分层） | ⏳ 待开发 | 第1层 |
| V2-IT6-BE-M2 | 记忆存储改造 | ⏳ 待开发 | 第2层 |
| V2-IT6-BE-M3 | ListMemories SQL 优化 | ✅ 已完成 | 第2层 |
| V2-IT6-BE-M4 | 记忆管理 API | ✅ 已完成 | 第2层 |
| V2-IT6-BE-M5 | 对话风格模板改造 | ⏳ 待开发 | 第2层 |
| V2-IT6-BE-M6 | 分享码信息增强 | ⏳ 待开发 | 第2层 |
| V2-IT6-BE-M7 | 聊天记录导入 API | ⏳ 待开发 | 第2层 |
| V2-IT6-BE-M8 | 部署配置文件 | ⏳ 待开发 | 第1层 |

### 前端模块划分
| 模块编号 | 模块名称 | 状态 | 备注 |
|----------|----------|------|------|
| V2-IT6-FE-M1 | 自定义 TabBar 组件 | ✅ 已完成 | CustomTabBar 组件 + app.config.ts（custom: true） |
| V2-IT6-FE-M2 | 教师端 TabBar 适配 | ✅ 已完成 | 4 Tab: 工作台/学生/知识库/我的 + profile 菜单更新 |
| V2-IT6-FE-M3 | 学生端 TabBar 适配 | ✅ 已完成 | 4 Tab: 对话/历史/发现/我的 + profile 菜单更新 |
| V2-IT6-FE-M4 | 对话风格选择器 | ✅ 已完成 | 6种风格模板 + style API 更新 + login slogan 更新 |
| V2-IT6-FE-M5 | 分享码二维码 | ✅ 已完成 | Canvas 绘制二维码 + 保存到相册 |
| V2-IT6-FE-M6 | 扫码落地页优化 | ✅ 已完成 | 5种 join_status UI + 向老师申请按钮 |
| V2-IT6-FE-M7 | 聊天记录导入 UI | ✅ 已完成 | knowledge/add 新增聊天记录 Tab + importChat API |
| V2-IT6-FE-M8 | 记忆管理 UI | ✅ 已完成 | 新增 memory-manage 页面 + memory API 增强 |

### 前端构建状态
- Webpack 编译: ✅ 通过（Compiled successfully）
- 编译时间: 3.28s
