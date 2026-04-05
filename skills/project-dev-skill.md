# Digital-Twin 迭代开发 Skill（项目特定）

> **通用核心规则**请加载：`../../skills/shared/iteration-dev/core.md`（~15KB，每次必须加载）
> **通用详细参考**按需查阅：`../../skills/shared/iteration-dev/reference.md`（~75KB，按需读取对应章节）
> **项目技术规范**请参考：`skills/project-tech-spec.md`
> **历史迭代归档**：`docs/archive/iteration-modules-archive.md`（V1.0 + V2.0 iter4~7）

---

## 1. 项目概述

| 项目 | 说明 |
|------|------|
| **项目名称** | digital-twin（数字分身教学助手） |
| **Go 模块路径** | `digital-twin` |
| **核心架构** | Harness 插件管道架构 |
| **核心接口** | `src/harness/core/plugin.go`（已冻结） |
| **配置文件** | `configs/harness.yaml` |
| **环境变量** | `.env` |

---

## 2. 冒烟测试用例（端到端验证）

> **完整用例文档**：`docs/testing/smoke_test_plan.md`（26 条用例，覆盖 13 个功能模块、24 个页面、双角色）
> **维护规则**：每个迭代在 Phase 0（环境准备）阶段，主 Agent 必须：①根据最新需求文档重新审视并更新冒烟测试计划文档（R17）；②Review 已有的集成测试代码（`tests/integration/`）和 E2E 测试代码（`src/frontend/e2e/`），对照最新需求进行删减、补充或调整，确保测试代码与冒烟测试计划一致（R17）
> **执行时机**：集成测试 + 代码&架构 Review 全部通过后
> **执行方式**：ci_test_agent 在集成测试通过后继续作为**编排者**，为每个冒烟用例拉起独立 sub agent
> **小程序端**：通过 miniprogram-automator SDK 控制微信开发者工具模拟器自动化执行（详见通用 Skill §10.5.5）
> **H5端**：通过 Playwright 控制浏览器模拟移动端设备执行（详见通用 Skill §10.5.6）

---

## 3. V2.0 迭代8 模块划分（当前迭代）

### 3.1 后端模块划分

| 模块编号 | 模块名称 | 优先级 | 依赖 | 说明 |
|----------|----------|--------|------|------|
| V2-IT8-BE-M1 | 数据库变更 + User表精简 | P0-第1层 | 无 | 移除users表冗余字段；扩展classes表；新增knowledge_items表；修改class_members表；新增chat_pins表 |
| V2-IT8-BE-M2 | 智能知识库上传API | P0-第2层 | M1 | POST /api/knowledge/smart-upload（URL/文字/文件统一入口）|
| V2-IT8-BE-M3 | 知识库管理API（搜索/CRUD） | P0-第2层 | M1 | GET /api/knowledge（支持搜索筛选）、PUT/DELETE /api/knowledge/:id |
| V2-IT8-BE-M4 | 班级管理增强API | P0-第2层 | M1 | 扩展班级创建字段；生成分享链接/二维码；班级分享信息接口 |
| V2-IT8-BE-M5 | 学生加入审批API | P0-第2层 | M1 | 申请加入、待审批列表、审批处理（含学生信息登记）|
| V2-IT8-BE-M6 | 聊天列表重构API | P0-第2层 | M1 | 学生端老师列表、教师端按班级组织的学生列表 |
| V2-IT8-BE-M7 | 置顶功能API | P0-第2层 | M1 | POST/DELETE /api/chat/pin、GET /api/chat/pins |
| V2-IT8-BE-M8 | 会话管理API | P1-第2层 | M1 | 新会话开启、快捷指令获取 |
| V2-IT8-BE-M9 | 发现页API | P1-第2层 | M1 | 推荐班级/老师、搜索班级 |

**开发顺序**：
```
第1层: V2-IT8-BE-M1 数据库变更
      ↓
第2层（尽量并行）:
  ├── V2-IT8-BE-M2 智能知识库上传
  ├── V2-IT8-BE-M3 知识库管理
  ├── V2-IT8-BE-M4 班级管理增强
  ├── V2-IT8-BE-M5 学生加入审批
  ├── V2-IT8-BE-M6 聊天列表重构
  ├── V2-IT8-BE-M7 置顶功能
  ├── V2-IT8-BE-M8 会话管理
  └── V2-IT8-BE-M9 发现页
```

### 3.2 前端模块划分

| 模块编号 | 模块名称 | 优先级 | 涉及页面 | 对应后端接口 |
|----------|----------|--------|----------|-------------|
| V2-IT8-FE-M1 | 知识库上传页面改版 | P0 | `knowledge/add` | smart-upload API |
| V2-IT8-FE-M2 | 知识库管理页面 | P0 | `knowledge/index`（改版）、`knowledge/preview` | knowledge CRUD API |
| V2-IT8-FE-M3 | 班级创建页面扩展 | P0 | `class-create/index`（表单扩展） | classes API |
| V2-IT8-FE-M4 | 班级创建后引导 | P0 | 新增引导弹窗/页面 | 无（纯前端）|
| V2-IT8-FE-M5 | 分享链接落地页 | P0 | `share-join/index`（改版） | join-request API |
| V2-IT8-FE-M6 | 学生审批页面 | P0 | 新增 `student-approve/index` | pending-requests/approve-request API |
| V2-IT8-FE-M7 | 学生端聊天列表 | P0 | 新增 `chat-list/index` | chat/teacher-list API |
| V2-IT8-FE-M8 | 聊天页改版（学生端） | P0 | `chat/index`（仿微信风格） | new-session/quick-actions API |
| V2-IT8-FE-M9 | 教师端聊天列表改版 | P0 | 新增 `teacher-chat-list/index` | chat/student-list/pins API |
| V2-IT8-FE-M10 | 发现页实现 | P1 | `discover/index`（完善） | discover API |
| V2-IT8-FE-M11 | 注册流程优化（移除字段） | P0 | `login/index`、`role-select/index` | 现有注册接口 |
| V2-IT8-FE-M12 | 教师首次引导流程 | P0 | 新增引导页面/弹窗 | 无（纯前端）|

**开发顺序**：
```
无强依赖，基于接口文档并行开发
建议优先完成：FE-M11（注册优化）、FE-M3（班级创建）、FE-M1/M2（知识库）
```

### 3.3 集成测试用例规划

| 用例编号 | 测试场景 | 涉及模块 | 数据依赖 |
|----------|----------|----------|----------|
| IT-501 | 知识库URL上传 + 解析 | BE-M2 | 无 |
| IT-502 | 知识库文字上传 | BE-M2 | 无 |
| IT-503 | 知识库文件批量上传 | BE-M2 | 无 |
| IT-504 | 知识库列表搜索筛选 | BE-M3 | IT-501~503 |
| IT-505 | 知识库编辑删除 | BE-M3 | IT-501 |
| IT-506 | 班级创建（扩展字段）+ 分享链接生成 | BE-M4 | 无 |
| IT-507 | 学生通过分享链接申请加入 | BE-M5 | IT-506 |
| IT-508 | 教师审批通过 + 填写学生信息 | BE-M5 | IT-507 |
| IT-509 | 教师审批拒绝 | BE-M5 | IT-507 |
| IT-510 | 学生端获取多老师聊天列表 | BE-M6 | IT-508 |
| IT-511 | 教师端获取按班级组织的学生列表 | BE-M6 | IT-508 |
| IT-512 | 置顶班级/学生 | BE-M7 | IT-511 |
| IT-513 | 取消置顶 | BE-M7 | IT-512 |
| IT-514 | 开启新会话 + 快捷指令 | BE-M8 | IT-510 |
| IT-515 | 发现页推荐 + 搜索 | BE-M9 | IT-506 |
| IT-516 | 全链路：注册→创建班级→分享→申请→审批→聊天→新会话 | 全部 | 无 |

**推荐分批方案**：
- 第 1 批（知识库）：IT-501 ~ IT-505
- 第 2 批（班级+审批）：IT-506 ~ IT-509
- 第 3 批（聊天列表+置顶）：IT-510 ~ IT-513
- 第 4 批（会话+发现+全链路）：IT-514 ~ IT-516

---

## 4. 执行指令模板

> **使用说明**：以下模板中的 `{X}` 和 `{N}` 为占位符，使用时请替换为实际的版本号和迭代号。例如 V1.0 第一迭代：`{X}` → `1`，`{N}` → `1`。

### 4.1 启动后端开发 Agent

```
@dev_backend_agent

## 任务：第N迭代后端开发

### 输入文档
- 需求文档: docs/iterations/v{X}.0/iteration{N}_requirements.md
- 接口规范: docs/iterations/v{X}.0/iteration{N}_api_spec.md
- 核心接口: src/harness/core/plugin.go
- 已有代码: src/harness/manager/ (harness_manager.go, pipeline.go, event_bus.go)
- 配置文件: configs/harness.yaml
- 环境变量: .env
- 项目技术规范: skills/project-tech-spec.md
- 通用核心Skill: ../../skills/shared/iteration-dev/core.md

### 环境保护规则
⚠️ 请严格遵守项目技术规范第2节"环境保护规则"，禁止修改 .env、core/plugin.go、configs/ 等受保护文件。

### 模块划分
（根据当前迭代的模块划分填写，按依赖层级顺序开发）

### 开发规则
1. **每个模块必须拉起独立的 sub agent，严禁合并多个模块到一个 sub agent 中**（通用Skill R4）
2. 每个模块需要2个独立的 sub agent：1个负责开发代码，1个负责编写并执行单元测试 + 代码Review（通用Skill R1）
3. 同一层级的开发 sub agent 尽量并行拉起，开发完成后再并行拉起对应的单测&Review sub agent
4. **sub agent 的 prompt 只包含结论和指令，不包含编排者的推理过程**（通用Skill R5）
5. 单测不通过或Review发现问题时，反馈给开发 sub agent 修复，最多3轮
6. 所有模块完成后，执行 `go build ./...` 确认整体编译通过
7. 执行 `go test ./...` 确认所有单元测试通过
8. 所有插件的 Execute 方法必须 merge 上游 Data（见项目技术规范 3.1）

### 技术约束
- Go 模块路径: digital-twin
- 使用 modernc.org/sqlite（纯Go，无需CGO）
- JWT: github.com/golang-jwt/jwt/v5
- HTTP: github.com/gin-gonic/gin
- YAML: gopkg.in/yaml.v3
- 大模型: OpenAI兼容格式（通义千问 qwen-turbo）
```

### 4.2 启动集成测试 Agent

```
@ci_test_agent

## 任务：第N迭代集成测试 + 迭代级Review

### 输入文档
- 需求文档: docs/iterations/v{X}.0/iteration{N}_requirements.md
- 接口规范: docs/iterations/v{X}.0/iteration{N}_api_spec.md
- 项目技术规范: skills/project-tech-spec.md
- 通用核心Skill: ../../skills/shared/iteration-dev/core.md

### 环境保护规则
⚠️ 请严格遵守项目技术规范第2节"环境保护规则"。测试数据库使用独立文件（如 data/test_integration.db），不要修改主数据库。

### ⚠️ 重要约束：不允许直接修改代码
集成测试 Agent **严禁直接修改任何代码文件**。发现问题时，只能通过反馈机制让对应的开发 Agent 进行修改。

### 测试准备
1. 读取需求文档和接口规范
2. **阅读本文档第 6 节「已知陷阱清单」，了解历史踩坑点**
3. 在 tests/integration/ 目录下编写**端到端（E2E）集成测试**用例
4. **分批编写执行**：每批 ≤ 5 个用例，以 API 规范文档为准编写测试（如发现代码与规范不一致，反馈给主 Agent 转交开发 Agent 修改），执行通过后再编写下一批
5. 测试用例需**尽量覆盖全面**，包括但不限于：
   - 认证流程（注册、登录、Token刷新、Token过期）
   - 知识库CRUD（创建、查询、删除、权限控制）
   - 对话全链路（正常对话、上下文传递、插件链路完整性）
   - 记忆查询（记忆创建、列表查询、记忆关联）
   - 系统接口（健康检查、插件列表、管道列表）
   - 边界条件（空参数、超长输入、特殊字符）
   - 并发场景（多用户同时操作）
6. ⚠️ 推荐使用 Go 的 httptest 包启动测试服务（见项目技术规范 3.6）

### 测试执行
1. 等待后端开发完成（go build 和 go test 通过）
2. 使用 httptest 在测试代码中直接启动 HTTP server
3. 按顺序执行测试用例（先注册登录，再业务操作）
4. 记录每个用例的通过/失败状态

### 失败处理
1. 测试不通过时，定位问题归属（后端哪个模块）
2. 生成失败报告：用例编号 + 期望结果 + 实际结果 + 错误信息
3. 反馈给 @dev_backend_agent 进行修复（**不要自己修改代码**）
4. 修复后重新执行失败的用例
5. 最多3轮修复循环

### Review + 集成测试 + 冒烟验证
（参考通用Skill §10.4~§10.5，冒烟用例列表见本文档第2节）

### Base URL
使用 httptest.NewServer 动态分配
```

### 4.3 启动前端开发 Agent

```
@dev_frontent_agent

## 任务：第N迭代前端开发

### 输入文档
- 需求文档: docs/iterations/v{X}.0/iteration{N}_requirements.md
- 接口规范: docs/iterations/v{X}.0/iteration{N}_api_spec.md
- 项目技术规范: skills/project-tech-spec.md
- 通用核心Skill: ../../skills/shared/iteration-dev/core.md

### 环境保护规则
⚠️ 请严格遵守项目技术规范第2节"环境保护规则"。

### 模块划分
（根据当前迭代的前端模块划分填写，按依赖层级顺序开发）

### 前端测试技术栈
| 测试层 | 小程序端 (Taro + React) | H5端 (Vue3 + Vite) |
|--------|------------------------|--------------------|
| **单元测试** | Jest | Vitest |
| **组件测试** | Jest + @testing-library/react + MSW | Vitest + @vue/test-utils + MSW |
| **E2E冒烟测试** | Jest + miniprogram-automator | Playwright |
| **API Mock** | MSW (Mock Service Worker) | MSW |

### 开发规则
1. **每个模块必须拉起独立的 sub agent，严禁合并多个模块到一个 sub agent 中**（通用Skill R4）
2. 每个模块需要2个独立的 sub agent：1个负责开发代码，1个负责编写并执行单元测试 + 组件测试 + 代码Review（通用Skill R1 + R20）
3. 同层级的开发 sub agent 尽量并行拉起，开发完成后再并行拉起对应的单测&Review sub agent
4. **sub agent 的 prompt 只包含结论和指令，不包含编排者的推理过程**（通用Skill R5）
5. 单测/组件测试不通过或Review发现问题时，反馈给开发 sub agent 修复，最多3轮
6. 所有模块完成后，执行整体构建检查
7. **前端单测和组件测试必须同步编写**（通用Skill R20）：
   - 小程序端：`npx jest` 运行单测，`npx jest e2e/` 运行 E2E
   - H5端：`npx vitest run` 运行单测，`npx playwright test` 运行 E2E
8. **组件测试必须使用 MSW 进行 API Mock**，验证组件在不同 API 响应下的行为（正常/异常/空数据）
9. **E2E 测试用例必须按平台分别编写**：小程序端用例放 `src/frontend/e2e/`，H5端用例放 `src/h5-teacher/e2e/`（或 `src/h5-student/e2e/`）

### 注意事项
- 纯后端迭代时，前端 Agent 无需参与开发，但主 Agent 必须在 Phase 2 模块完整性检查时确认
```

---

## 5. 已验证的项目特定最佳实践

| 实践 | 说明 |
|------|------|
| httptest 集成测试 | 避免端口占用、进程管理等复杂问题，测试自包含可重复 |
| flattenConfig 适配层 | 当配置结构（嵌套）和消费方接口（扁平 map）不匹配时，用 flatten 适配 |
| 插件 Data merge | 每个插件的 Execute 方法必须 merge 上游 Data，否则上游数据丢失 |

---

## 6. 已知陷阱清单（随迭代积累）

> ci_test_agent 编写集成测试前必须阅读本清单，避免重复踩坑。开发 Agent 也应参考。

| # | 陷阱描述 | 来源迭代 | 影响 | 防护措施 |
|---|----------|----------|------|----------|
| T1 | API 请求参数名与实际 handler 不一致（如测试用 `username` 但 handler 期望 `phone`） | V2.0-iter4 | 集成测试全部失败，需逐个修复 | ci_test_agent 以 API 规范为准编写测试，发现不一致则反馈开发 Agent 修改代码 |
| T2 | 新增数据库字段后，只改了模型定义和 DDL，遗漏了 INSERT SQL 和 handler 响应 | V2.0-iter4 | 字段写入为空值，响应缺少字段 | 开发 Agent 必须先输出「字段影响清单」（R15） |
| T3 | 测试用例间存在隐式数据依赖（如 IT-03 依赖 IT-01 创建的用户），但未显式声明 | V2.0-iter4 | 用例执行顺序变化时随机失败 | 用例规划表增加「数据依赖」列，显式声明 |
| T4 | 集成测试一次性编写 17 个用例，上下文过长导致后半部分用例质量下降 | V2.0-iter4 | 后半部分用例参数错误率高 | 分批编写执行，每批 ≤ 5 个用例 |
| T5 | 测试数据库未隔离，多个测试用例操作同一张表导致数据冲突 | V2.0-iter4 | 并发测试随机失败 | 每个测试函数使用独立的数据库文件或事务回滚 |
| T6 | 前端模块在迭代执行中被遗漏，主 Agent 只开发了后端就进入了 Phase 3 | V2.0-iter4 | 前端功能缺失，需要回退补开发 | Phase 2 结束前必须对照需求文档确认后端+前端模块全部已开发（R16） |
| T7 | 主 Agent 直接编写前端代码，未调用 dev_frontent_agent | V2.0-iter4 | 代码质量不可控，缺少单测和 Review | 主 Agent 禁止直接编写代码（R16） |
| T8 | Python 服务未启动时 Go 后端未做降级处理，导致知识库功能完全不可用 | V2.0-iter5 | 知识库添加/检索全部失败 | VectorClient 必须实现降级逻辑（Python 不可用时返回空结果） |

---

**文档版本**: v4.0.0
**创建日期**: 2026-03-28
**更新日期**: 2026-04-04
**最新迭代**: V2.0 迭代8
**适用项目**: digital-twin
**关联文档**:
- 通用核心Skill: `../../skills/shared/iteration-dev/core.md`
- 通用详细参考: `../../skills/shared/iteration-dev/reference.md`
- 项目技术规范: `skills/project-tech-spec.md`
- 历史迭代归档: `docs/archive/iteration-modules-archive.md`
- 冒烟测试计划: `docs/testing/smoke_test_plan.md`
**变更记录**:
- v4.0.0: 目录重组 + 内容精简 — 历史迭代数据（V1.0 + V2.0 iter4~7）归档到 `iteration-dev-skill-archive.md`；章节重编号（§2 冒烟测试、§3 当前迭代、§4 模板、§5 实践、§6 陷阱）；所有内部引用路径更新为新目录结构；变更记录精简
- v3.6.0: 前端测试全覆盖（R20）
- v3.5.0: 新增 V2.0 迭代8 模块划分
- v3.0.0~v3.4.0: 迭代4~7 模块划分（已归档）
- v2.0.0~v2.9.0: 拆分重构、路径修复、模板优化（已归档）
