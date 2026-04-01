# 第二迭代任务跟踪

## 迭代信息

| 项目 | 说明 |
|------|------|
| 迭代名称 | Sprint 2 - 小程序前端开发 & 后端适配改造 |
| 迭代周期 | 2026-04-15 至 2026-04-28 |
| 状态 | ✅ 已完成 |

---

## 前端模块跟踪

| 模块编号 | 模块名称 | 负责 Agent | 状态 | 开发 | 单测 | 备注 |
|----------|----------|------------|------|------|------|------|
| FE-M1 | 项目脚手架 | dev_frontent_agent | ✅ 完成 | ✅ | - | Taro 3.6.31 + React 18 + Zustand 4 + TypeScript |
| FE-M2 | 网络请求层 | dev_frontent_agent | ✅ 完成 | ✅ | - | request.ts + 7个 API 模块（auth/teacher/chat/knowledge/memory/user） |
| FE-M3 | 状态管理层 | dev_frontent_agent | ✅ 完成 | ✅ | - | userStore + chatStore + teacherStore（zustand） |
| FE-M4 | 微信登录 | dev_frontent_agent | ✅ 完成 | ✅ | - | 登录页 + 角色选择页 + 路由守卫（app.ts） |
| FE-M5 | 学生首页 | dev_frontent_agent | ✅ 完成 | ✅ | - | 首页 + TeacherCard 组件 + Empty 组件 |
| FE-M6 | 对话模块 | dev_frontent_agent | ✅ 完成 | ✅ | - | 对话页 + ChatBubble 组件 + AI思考动画 |
| FE-M7 | 知识库模块 | dev_frontent_agent | ✅ 完成 | ✅ | - | 知识库管理 + 添加文档 + TagInput 组件 |
| FE-M8 | 个人中心 | dev_frontent_agent | ✅ 完成 | ✅ | - | 个人中心 + CustomTabBar 占位 |
| FE-M9 | 对话历史 | dev_frontent_agent | ✅ 完成 | ✅ | - | 对话历史页（会话列表） |
| FE-M10 | 记忆查看 | dev_frontent_agent | ✅ 完成 | ✅ | - | 记忆查看页（教师筛选 + 类型筛选） |

## 后端模块跟踪

| 模块编号 | 模块名称 | 负责 Agent | 状态 | 开发 | 单测 | 备注 |
|----------|----------|------------|------|------|------|------|
| BE-M1 | 微信登录接口 | dev_backend_agent | ✅ 完成 | ✅ | ✅ 14个 | wx_client.go + auth_plugin 新增 wx-login/complete-profile + openid 字段 |
| BE-M2 | 补全信息接口 | dev_backend_agent | ✅ 完成 | ✅ | ✅ | 包含在 BE-M1 中（handleCompleteProfile） |
| BE-M3 | 教师列表接口 | dev_backend_agent | ✅ 完成 | ✅ | ✅ 4个 | GET /api/teachers + GetTeachers + TeacherWithDocCount |
| BE-M4 | 用户信息接口 | dev_backend_agent | ✅ 完成 | ✅ | ✅ 3个 | GET /api/user/profile + GetUserStats |
| BE-M5 | 会话列表接口 | dev_backend_agent | ✅ 完成 | ✅ | ✅ 5个 | GET /api/conversations/sessions + GetSessionsByStudent |
| BE-M6 | CORS 配置适配 | 主 Agent | ✅ 完成 | ✅ | - | harness.yaml 更新 CORS 配置 |
| BE-M7 | 对话历史增强 | dev_backend_agent | ✅ 完成 | ✅ | ✅ 4个 | teacher_id 改为可选 + session_id 筛选 |

## 集成测试跟踪

| 用例编号 | 测试场景 | 状态 | 备注 |
|----------|----------|------|------|
| IT-18 | 微信登录（mock）→ 返回 token + is_new_user=true | ✅ 通过 | 0.04s |
| IT-19 | 新用户补全信息 → 设置角色和昵称 | ✅ 通过 | 0.00s |
| IT-20 | 同一 openid 再次登录 → is_new_user=false | ✅ 通过 | 0.00s |
| IT-21 | 获取教师列表 → 返回教师数组 + document_count | ✅ 通过 | 0.05s |
| IT-22 | 获取用户信息 → 返回 profile + stats | ✅ 通过 | 0.00s |
| IT-23 | 获取会话列表 → 返回会话摘要 | ✅ 通过 | 0.00s |
| IT-24 | 对话历史不传 teacher_id → 返回所有对话 | ✅ 通过 | 0.00s |
| IT-25 | 微信登录→补全信息→教师列表→对话 全链路 | ✅ 通过 | 0.05s |
| IT-26 | 教师微信登录→补全→添加文档→学生对话引用知识 | ✅ 通过 | 0.09s |
| IT-27 | 多轮对话→会话列表→对话历史→记忆 | ✅ 通过 | 0.09s |

## 单元测试统计

| 测试包 | 用例数 | 状态 |
|--------|--------|------|
| src/backend/database | 30+ | ✅ 全部通过 |
| src/plugins/auth | 20+ | ✅ 全部通过 |
| src/plugins/dialogue | 已有 | ✅ 全部通过 |
| src/plugins/knowledge | 已有 | ✅ 全部通过 |
| src/plugins/memory | 已有 | ✅ 全部通过 |
| src/harness/config | 已有 | ✅ 全部通过 |
| src/harness/manager | 已有 | ✅ 全部通过 |
| tests/integration | 27 | ✅ 全部通过（IT-01~IT-27） |

---

## 开发顺序（实际执行）

```
Phase 0: 环境准备 ✅
  ├── go build ./... 验证第一迭代代码正常
  └── 更新 harness.yaml CORS 配置

Phase 1: 后端第1层 ✅
  └── BE-M1 微信登录（开发 + 14个单测）+ BE-M6 CORS

Phase 1: 后端第2层 ✅
  └── BE-M3 教师列表 + BE-M4 用户信息 + BE-M7 对话历史增强（开发 + 11个单测）

Phase 1: 后端第3层 ✅
  └── BE-M5 会话列表（开发 + 5个单测）

Phase 2: 前端第1层 ✅
  └── FE-M1 脚手架 + FE-M2 网络请求 + FE-M3 状态管理

Phase 2: 前端第2层 ✅
  └── FE-M4 微信登录 + FE-M5 学生首页

Phase 2: 前端第3层 ✅
  └── FE-M6 对话模块 + FE-M7 知识库模块 + FE-M8 个人中心

Phase 2: 前端第4层 ✅
  └── FE-M9 对话历史 + FE-M10 记忆查看

Phase 3: 集成测试 ✅
  └── IT-18 ~ IT-27 全部通过

Phase 4: 收尾 ✅
  └── 更新 task_tracking.md
```

## 新增/修改文件清单

### 后端（Go）
| 文件 | 操作 | 说明 |
|------|------|------|
| src/plugins/auth/wx_client.go | 🆕 新建 | WxClient 接口 + RealWxClient + MockWxClient |
| src/plugins/auth/auth_plugin.go | 修改 | 新增 wx-login + complete-profile action |
| src/plugins/auth/auth_plugin_test.go | 修改 | 新增 9 个微信登录/补全信息测试 |
| src/backend/database/models.go | 修改 | User 新增 OpenID + TeacherWithDocCount + SessionSummary |
| src/backend/database/database.go | 修改 | users 表新增 openid 字段 + 唯一索引 |
| src/backend/database/repository.go | 修改 | 新增 GetByOpenID/CreateWithOpenID/UpdateRoleAndNickname/GetTeachers/GetUserStats/GetSessionsByStudent 等方法 |
| src/backend/database/database_test.go | 修改 | 新增 16 个 repository 层测试 |
| src/backend/api/handlers.go | 修改 | 新增 HandleWxLogin/HandleCompleteProfile/HandleGetTeachers/HandleGetUserProfile/HandleGetSessions + 改造 HandleGetConversations |
| src/backend/api/router.go | 修改 | 新增 5 个路由 |
| configs/harness.yaml | 修改 | CORS 配置更新 |
| tests/integration/integration_test.go | 修改 | TestMain 添加 WX_MODE=mock |
| tests/integration/integration_v2_test.go | 🆕 新建 | V2 集成测试 10 个用例 |

### 前端（TypeScript/React）
| 文件 | 操作 | 说明 |
|------|------|------|
| src/frontend/package.json | 🆕 新建 | 项目依赖 |
| src/frontend/tsconfig.json | 🆕 新建 | TypeScript 配置 |
| src/frontend/babel.config.js | 🆕 新建 | Babel 配置 |
| src/frontend/project.config.json | 🆕 新建 | 小程序项目配置 |
| src/frontend/config/*.ts | 🆕 新建 | Taro 编译配置（3个文件） |
| src/frontend/src/app.ts | 🆕 新建 | 应用入口 + 路由守卫 |
| src/frontend/src/app.config.ts | 🆕 新建 | 全局配置（页面路由 + TabBar） |
| src/frontend/src/app.scss | 🆕 新建 | 全局样式 |
| src/frontend/src/api/*.ts | 🆕 新建 | 网络请求层（7个文件） |
| src/frontend/src/store/*.ts | 🆕 新建 | 状态管理（4个文件） |
| src/frontend/src/utils/*.ts | 🆕 新建 | 工具函数（3个文件） |
| src/frontend/src/components/**/index.tsx+scss | 🆕 新建 | 5个组件（ChatBubble/TeacherCard/Empty/TagInput/CustomTabBar） |
| src/frontend/src/pages/**/index.tsx+scss+config.ts | 🆕 新建 | 9个页面（login/role-select/home/chat/history/knowledge/knowledge-add/memories/profile） |

---

**最后更新**: 2026-03-28
