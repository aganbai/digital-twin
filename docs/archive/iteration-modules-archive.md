# Digital-Twin 历史迭代归档

> 本文件归档已完成迭代的模块划分和测试用例，仅供历史查阅。
> 当前迭代请查看：`skills/project-dev-skill.md`

---

## 1. V1.0 模块划分

### 1.1 后端模块划分

| 模块编号 | 模块名称 | 目录 | 依赖 | 优先级 |
|----------|----------|------|------|--------|
| BE-M1 | 配置管理 | `src/harness/config/` | 无 | P0-第1层 |
| BE-M2 | 数据库层 | `src/backend/database/` | 无 | P0-第1层 |
| BE-M3 | 认证插件 | `src/plugins/auth/` | M1, M2 | P0-第2层 |
| BE-M4 | 知识库插件 | `src/plugins/knowledge/` | M1, M2 | P0-第2层 |
| BE-M5 | 记忆插件 | `src/plugins/memory/` | M1, M2 | P0-第2层 |
| BE-M6 | 对话插件 | `src/plugins/dialogue/` | M1, M2 | P0-第2层 |
| BE-M7 | 管理器补全 | `src/harness/manager/` | M1~M6 | P0-第3层 |
| BE-M8 | HTTP API层 | `src/backend/api/` + `src/cmd/server/` | M7 | P0-第4层 |

### 1.2 集成测试用例

| 用例编号 | 测试场景 | 接口 | 验证点 | 数据依赖 |
|----------|----------|------|--------|----------|
| IT-01 | 用户注册（教师） | POST /api/auth/register | 返回 user_id + token，密码 bcrypt 加密 | 无 |
| IT-02 | 用户注册（学生） | POST /api/auth/register | 返回 user_id + token | 无 |
| IT-03 | 重复注册 | POST /api/auth/register | 返回 40006 错误 | IT-01 |
| IT-04 | 用户登录 | POST /api/auth/login | 返回有效 JWT | IT-01 |
| IT-05 | 登录密码错误 | POST /api/auth/login | 返回 40001 错误 | IT-01 |
| IT-06 | 无效Token访问 | POST /api/chat | 返回 401 | 无 |
| IT-07 | 添加知识文档 | POST /api/documents | 教师角色成功，返回 document_id | IT-01 |
| IT-08 | 学生添加文档（权限） | POST /api/documents | 返回 40003 权限不足 | IT-02 |
| IT-09 | 获取文档列表 | GET /api/documents | 返回文档数组 | IT-07 |
| IT-10 | 删除文档 | DELETE /api/documents/:id | 返回 deleted: true | IT-07 |
| IT-11 | 学生对话（全链路） | POST /api/chat | 走完4个插件，返回苏格拉底式回复 | IT-02 |
| IT-12 | 对话历史查询 | GET /api/conversations | 返回对话记录 | IT-11 |
| IT-13 | 记忆列表查询 | GET /api/memories | 返回记忆数组 | IT-11 |
| IT-14 | 健康检查 | GET /api/system/health | 返回 status: running | 无 |
| IT-15 | 插件列表（admin） | GET /api/system/plugins | 返回4个插件 | IT-01 |
| IT-16 | 管道列表（admin） | GET /api/system/pipelines | 返回2个管道 | IT-01 |
| IT-17 | 令牌刷新 | POST /api/auth/refresh | 返回新 token | IT-04 |

---

## 2. V2.0 迭代4 模块划分

### 2.1 后端模块划分

| 模块编号 | 模块名称 | 优先级 | 依赖 |
|----------|----------|--------|------|
| V2-IT4-BE-M1 | 数据库迁移 | P0-第1层 | 无 |
| V2-IT4-BE-M2 | 老师分身广场 API | P0-第2层 | M1 |
| V2-IT4-BE-M3 | 定向邀请 API | P0-第2层 | M1 |
| V2-IT4-BE-M4 | 教师真人介入对话 API | P0-第2层 | M1 |
| V2-IT4-BE-M5 | 分身概览改造 | P0-第2层 | M1 |
| V2-IT4-BE-M6 | 知识库 LLM 智能摘要 | P0-第2层 | M1 |

### 2.2 前端模块划分

| 模块编号 | 模块名称 | 优先级 | 涉及页面 | 对应后端接口 |
|----------|----------|--------|----------|-------------|
| V2-IT4-FE-M1 | 老师分身广场 UI | P0 | home/index.tsx | BE-M2 接口 |
| V2-IT4-FE-M2 | 定向邀请 UI | P0 | share 相关页面 | BE-M3 接口 |
| V2-IT4-FE-M3 | 对话页 sender_type 标识 + 引用回复 | P0 | chat/index.tsx | BE-M4 接口 |
| V2-IT4-FE-M4 | 教师对话记录页 + 真人回复 | P0 | student-chat-history/index.tsx | BE-M4 接口 |
| V2-IT4-FE-M5 | 分身概览页 | P0 | persona-overview/index.tsx | BE-M5 接口 |
| V2-IT4-FE-M6 | 知识库 LLM 摘要 UI | P0 | knowledge/preview.tsx + knowledge/add.tsx | BE-M6 接口 |

### 2.3 集成测试用例

| 用例编号 | 测试场景 | 涉及模块 | 数据依赖 |
|----------|----------|----------|----------|
| IT-106 ~ IT-125 | 广场、定向邀请、真人介入、概览、摘要、全链路 | 全部 | 见原文档 |

---

## 3. V2.0 迭代5 模块划分

### 3.1 后端模块划分

| 模块编号 | 模块名称 | 优先级 | 依赖 |
|----------|----------|--------|------|
| V2-IT5-BE-M1 | Python LlamaIndex 服务搭建 | P0-第1层 | 无 |
| V2-IT5-BE-M2 | Go 向量客户端适配层 | P0-第2层 | M1 |
| V2-IT5-BE-M3 | 对话附件支持 | P1-第2层 | 无 |
| V2-IT5-BE-M4 | 文件上传接口 | P1-第2层 | 无 |
| V2-IT5-BE-M5 | 评语接口权限调整 | P1-第2层 | 无 |

### 3.2 前端模块划分

| 模块编号 | 模块名称 | 优先级 | 涉及页面 |
|----------|----------|--------|----------|
| V2-IT5-FE-M1 | 教师首页 Dashboard 重构 | P0 | home/index（教师模式） |
| V2-IT5-FE-M2 | 学生首页重构 + 路由优化 | P0 | home/index（学生模式） |
| V2-IT5-FE-M3 | 学生管理页合并 | P0 | teacher-students/index |
| V2-IT5-FE-M4 | 对话附件发送 | P1 | chat/index |
| V2-IT5-FE-M5 | 独立发现页 | P1 | discover/index |
| V2-IT5-FE-M6 | 评语→学生备注 + 分享码入口调整 | P1 | student-detail/index |

### 3.3 集成测试用例

| 用例编号 | 测试场景 | 涉及模块 | 数据依赖 |
|----------|----------|----------|----------|
| IT-201 ~ IT-213 | Python服务、Go集成、附件、评语、全链路 | 全部 | 见原文档 |

---

## 4. V2.0 迭代6 模块划分

### 4.1 后端模块划分

| 模块编号 | 模块名称 | 优先级 | 依赖 |
|----------|----------|--------|------|
| V2-IT6-BE-M1 | 数据库变更（记忆分层） | P0-第1层 | 无 |
| V2-IT6-BE-M2 | 记忆存储改造 | P0-第2层 | M1 |
| V2-IT6-BE-M3 | ListMemories SQL 优化 | P0-第2层 | M1 |
| V2-IT6-BE-M4 | 记忆管理 API | P1-第2层 | M1 |
| V2-IT6-BE-M5 | 对话风格模板改造 | P0-第2层 | 无 |
| V2-IT6-BE-M6 | 分享码信息增强 | P0-第2层 | 无 |
| V2-IT6-BE-M7 | 聊天记录导入 API | P1-第2层 | 无 |
| V2-IT6-BE-M8 | 部署配置文件 | P0-第1层 | 无 |

### 4.2 前端模块划分

| 模块编号 | 模块名称 | 优先级 | 涉及页面 |
|----------|----------|--------|----------|
| V2-IT6-FE-M1 | 自定义 TabBar 组件 | P0 | CustomTabBar 组件 |
| V2-IT6-FE-M2 | 教师端 TabBar 适配 | P0 | home + teacher-students + knowledge + profile |
| V2-IT6-FE-M3 | 学生端 TabBar 适配 | P0 | home + history + discover + profile |
| V2-IT6-FE-M4 | 对话风格选择器 | P0 | student-detail + login |
| V2-IT6-FE-M5 | 分享码二维码 | P1 | share-manage |
| V2-IT6-FE-M6 | 扫码落地页优化 | P1 | share-join |
| V2-IT6-FE-M7 | 聊天记录导入 UI | P1 | knowledge/add |
| V2-IT6-FE-M8 | 记忆管理 UI | P1 | memory-manage |

### 4.3 集成测试用例

| 用例编号 | 测试场景 | 涉及模块 | 数据依赖 |
|----------|----------|----------|----------|
| IT-301 ~ IT-317 | 记忆系统、对话风格、分享优化、知识库、全链路 | 全部 | 见原文档 |

---

## 5. V2.0 迭代7 模块划分

### 5.1 后端模块划分

| 模块编号 | 模块名称 | 优先级 | 依赖 | 已有代码状态 |
|----------|----------|--------|------|-------------|
| V2-IT7-BE-M1 | 数据库变更 + 作业清理 | P0-第1层 | 无 | ⚠️ 部分完成 |
| V2-IT7-BE-M2 | 学段模板配置化 | P0-第1层 | 无 | ⚠️ 需改造 |
| V2-IT7-BE-M3 | 教材配置API补全 | P0-第2层 | M1 | ✅ 大部分完成 |
| V2-IT7-BE-M4 | Prompt鲁棒性增强 | P0-第2层 | M2 | ✅ 已实现 |
| V2-IT7-BE-M5 | 批量上传Go端+Python脚本 | P1-第2层 | M1 | 🟡 仅DB表 |
| V2-IT7-BE-M6 | API限流中间件 | P1-第2层 | 无 | 🟡 仅配置结构 |
| V2-IT7-BE-M7 | Adaptive RAG | P1-第2层 | 无 | ❌ 未实现 |

### 5.2 前端模块划分

| 模块编号 | 模块名称 | 优先级 | 涉及页面 | 已有代码状态 |
|----------|----------|--------|----------|-------------|
| V2-IT7-FE-M1 | 作业功能清理 | P0 | chat/index.tsx + app.config.ts | ⚠️ 残留未清理 |
| V2-IT7-FE-M2 | 教材配置页完善 | P0 | curriculum-config/ | ✅ 已有页面 |
| V2-IT7-FE-M3 | 反馈页完善 | P0 | feedback/ | ✅ 已有页面 |
| V2-IT7-FE-M4 | 批量添加学生页完善 | P0 | student-batch/ | ✅ 已有页面 |
| V2-IT7-FE-M5 | 消息推送页完善 | P1 | teacher-message/ | ✅ 已有页面 |
| V2-IT7-FE-M6 | 批量上传UI | P1 | knowledge/add | ❌ 未实现 |
| V2-IT7-FE-M7 | 语音输入 | P1 | chat/index.tsx | ❌ 未实现 |
| V2-IT7-FE-M8 | Emoji表情面板 | P2 | chat/index.tsx | ❌ 未实现 |

### 5.3 集成测试用例

| 用例编号 | 测试场景 | 涉及模块 | 数据依赖 |
|----------|----------|----------|----------|
| IT-401 ~ IT-417 | 教材配置、反馈、学生、推送、上传、安全、全链路 | 全部 | 见原文档 |

---

**归档日期**: 2026-04-04
**归档范围**: V1.0 + V2.0 迭代4~7
