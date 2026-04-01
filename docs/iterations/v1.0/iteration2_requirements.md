# 第二迭代需求规格说明书

## 1. 迭代概述

| 项目 | 说明 |
|------|------|
| **迭代名称** | Sprint 2 - 小程序前端开发 & 后端适配改造 |
| **迭代目标** | 完成 Taro 小程序前端全部页面，后端配合前端做接口适配和功能增强 |
| **迭代周期** | 2周（2026-04-15 至 2026-04-28） |
| **交付标准** | 小程序可在微信开发者工具中运行，前后端联调通过所有集成测试 |
| **前端依赖** | 第一迭代后端 API 全部就绪 |

## 2. 迭代目标

### 2.1 核心目标
> **完成小程序前端 MVP，实现教师和学生的核心使用流程**

具体来说：
1. ✅ 小程序框架搭建：Taro + Vant Weapp + 状态管理
2. ✅ 用户认证流程：微信登录 + 新用户角色选择，JWT Token 管理
3. ✅ 学生对话页面：与数字分身的苏格拉底式对话交互
4. ✅ 教师知识库管理：文档列表、添加文档、删除文档
5. ✅ 个人中心：用户信息展示、角色切换入口
6. ✅ 后端 CORS 适配：支持小程序请求
7. ✅ 后端接口增强：用户信息查询、教师列表等前端所需接口

### 2.2 不在本迭代范围
- ❌ 文件上传（PDF/DOCX 解析）—— 知识库仍为文本录入
- ❌ 流式输出（SSE/WebSocket）—— 对话仍为请求-响应模式
- ❌ 记忆衰减机制
- ❌ 数据分析看板
- ❌ 留言系统
- ❌ 真机发布（仅在开发者工具中验证）

---

## 3. 前端需求

### 3.1 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| Taro | 3.x | 跨端小程序框架 |
| React | 18.x | UI 框架（Taro 默认） |
| Vant Weapp | 1.x | 小程序 UI 组件库 |
| Taro UI | 3.x | 备选 UI 组件库（如 Vant 不兼容则使用） |
| TypeScript | 5.x | 类型安全 |
| Zustand / Taro 内置 | - | 轻量状态管理 |

### 3.2 页面清单

| 页面编号 | 页面名称 | 路径 | 角色 | 优先级 |
|----------|----------|------|------|--------|
| FE-P1 | 登录页 | `/pages/login/index` | 所有 | P0 |
| FE-P2 | 角色选择页（新用户） | `/pages/role-select/index` | 新用户 | P0 |
| FE-P3 | 首页（学生） | `/pages/home/index` | student | P0 |
| FE-P4 | 教师选择页 | `/pages/teachers/index` | student | P0 |
| FE-P5 | 对话页 | `/pages/chat/index` | student | P0 |
| FE-P6 | 对话历史页 | `/pages/history/index` | student | P1 |
| FE-P7 | 知识库管理页 | `/pages/knowledge/index` | teacher | P0 |
| FE-P8 | 添加文档页 | `/pages/knowledge/add` | teacher | P0 |
| FE-P9 | 个人中心页 | `/pages/profile/index` | 所有 | P0 |
| FE-P10 | 记忆查看页 | `/pages/memories/index` | student | P1 |

---

### 3.3 页面详细设计

#### FE-P1: 登录页

**功能描述**：用户通过微信一键登录。

**UI 元素**：
- Logo + 应用名称（"AI 数字分身"）
- 应用简介文案（"基于苏格拉底式教学的 AI 数字分身"）
- 微信登录按钮（绿色，微信图标 + "微信登录"）
- 底部用户协议/隐私政策链接（占位，本迭代不实现跳转）
- 加载状态（登录中...）

**交互逻辑**：
1. 点击"微信登录"按钮：
   a. 调用 `wx.login()` 获取临时 `code`
   b. 调用 `POST /api/auth/wx-login` 将 code 发送给后端
   c. 后端用 code 换取微信 openid → 查找/创建用户 → 返回 JWT Token
2. 登录成功后判断 `is_new_user`：
   - `is_new_user = true` → 跳转角色选择页 `/pages/role-select/index`
   - `is_new_user = false` → 存储 token → 根据角色跳转：
     - student → 首页 `/pages/home/index`
     - teacher → 知识库管理 `/pages/knowledge/index`
3. 失败 → 显示错误提示（Toast："登录失败，请重试"）

**接口依赖**：
- `POST /api/auth/wx-login`（🆕 新增接口，替代原 login/register）

---

#### FE-P2: 角色选择页（新用户）

**功能描述**：微信登录后的新用户选择角色并填写昵称。

**UI 元素**：
- 标题（"欢迎加入 AI 数字分身"）
- 副标题（"请选择你的身份"）
- 角色选择卡片（两张大卡片，互斥选择）：
  - 教师卡片：教师图标 + "我是教师" + "创建知识库，打造你的数字分身"
  - 学生卡片：学生图标 + "我是学生" + "与教师的数字分身对话学习"
- 昵称输入框（placeholder: "请输入你的昵称"）
- 确认按钮（"开始使用"）

**交互逻辑**：
1. 必须选择一个角色
2. 昵称必填（1-20 字符）
3. 点击"开始使用" → 调用 `POST /api/auth/complete-profile`
4. 成功 → 根据角色跳转对应首页
5. 失败 → 显示错误提示

**接口依赖**：
- `POST /api/auth/complete-profile`（🆕 新增接口）

---

#### FE-P3: 首页（学生视角）

**功能描述**：学生的主页面，展示可对话的教师列表和快捷入口。

**UI 元素**：
- 顶部问候语（"你好，{nickname}"）
- 搜索框（搜索教师，P1 优先级，本迭代可不实现搜索功能，仅展示 UI）
- 教师卡片列表（每个卡片：头像占位 + 教师昵称 + 简介 + "开始对话"按钮）
- 底部 TabBar：首页 | 历史 | 我的

**交互逻辑**：
1. 进入页面 → 调用 `GET /api/teachers` 获取教师列表
2. 点击教师卡片 → 跳转对话页 `/pages/chat/index?teacher_id={id}`
3. TabBar 切换页面

**接口依赖**：
- `GET /api/teachers`（🆕 新增接口）

---

#### FE-P4: 教师选择页

**功能描述**：学生选择要对话的教师（当教师数量较多时的完整列表页）。

> 本迭代可与 FE-P3 合并，首页即为教师选择页。如教师数量少（<10），无需独立页面。

**UI 元素**：
- 教师列表（头像 + 昵称 + 知识库文档数 + 最近对话时间）
- 下拉刷新

**接口依赖**：
- `GET /api/teachers`（🆕 新增接口）

---

#### FE-P5: 对话页（核心页面）

**功能描述**：学生与教师数字分身的对话交互页面。

**UI 元素**：
- 顶部导航栏：教师昵称 + 返回按钮
- 消息列表区域（聊天气泡样式）：
  - 用户消息：右侧蓝色气泡
  - AI 回复：左侧白色气泡 + 教师头像占位
  - 时间戳（每隔 5 分钟显示一次）
- 底部输入区域：
  - 文本输入框（placeholder: "请输入你的问题..."）
  - 发送按钮（图标）
- 加载状态：AI 思考中动画（三个跳动的点）
- 空状态：首次对话提示（"向 {teacher_name} 的数字分身提问吧！"）

**交互逻辑**：
1. 进入页面 → 调用 `GET /api/conversations?teacher_id={id}&page_size=50` 加载历史消息
2. 输入消息 → 点击发送 / 键盘回车：
   a. 立即在列表中显示用户消息（乐观更新）
   b. 显示 AI 思考中动画
   c. 调用 `POST /api/chat`
   d. 收到回复 → 隐藏动画 → 显示 AI 回复气泡
   e. 自动滚动到底部
3. 发送失败 → 消息气泡显示"发送失败，点击重试"
4. 上拉加载更多历史消息（分页）

**接口依赖**：
- `POST /api/chat`
- `GET /api/conversations`

---

#### FE-P6: 对话历史页

**功能描述**：查看与不同教师的对话会话列表。

**UI 元素**：
- 会话列表（每项：教师头像 + 教师昵称 + 最后一条消息摘要 + 时间）
- 点击进入对话详情
- 空状态提示

**交互逻辑**：
1. 进入页面 → 调用 `GET /api/conversations/sessions`（🆕 新增接口）
2. 点击某个会话 → 跳转对话页 `/pages/chat/index?teacher_id={id}&session_id={sid}`

**接口依赖**：
- `GET /api/conversations/sessions`（🆕 新增接口）

---

#### FE-P7: 知识库管理页（教师视角）

**功能描述**：教师查看和管理自己的知识库文档。

**UI 元素**：
- 顶部标题："我的知识库"
- 文档统计（"共 {n} 篇文档"）
- 文档列表（每项：标题 + 标签 + 创建时间 + 删除按钮）
- 右下角浮动"添加"按钮（FAB）
- 空状态提示（"还没有文档，点击右下角添加"）
- 下拉刷新
- 底部 TabBar：知识库 | 我的

**交互逻辑**：
1. 进入页面 → 调用 `GET /api/documents` 获取文档列表
2. 点击"添加" → 跳转添加文档页
3. 左滑/长按文档 → 显示删除确认 → 调用 `DELETE /api/documents/:id`
4. 删除成功 → 刷新列表

**接口依赖**：
- `GET /api/documents`
- `DELETE /api/documents/:id`

---

#### FE-P8: 添加文档页

**功能描述**：教师录入知识文档（本迭代为纯文本录入）。

**UI 元素**：
- 标题输入框（placeholder: "请输入文档标题"）
- 内容输入框（多行文本域，placeholder: "请输入文档内容..."，最小高度 300px）
- 标签输入（可添加多个标签，Tag 样式）
- 提交按钮
- 字数统计（右下角显示当前字数）

**交互逻辑**：
1. 输入校验：标题必填（1-200 字符），内容必填（1-100000 字符）
2. 点击提交 → 调用 `POST /api/documents`
3. 成功 → Toast 提示"添加成功" → 返回知识库列表页
4. 失败 → 显示错误提示

**接口依赖**：
- `POST /api/documents`

---

#### FE-P9: 个人中心页

**功能描述**：展示用户信息，提供设置入口。

**UI 元素**：
- 用户头像占位 + 昵称 + 角色标签（教师/学生）
- 功能列表：
  - 我的记忆（学生可见）→ 跳转记忆查看页
  - 对话历史（学生可见）→ 跳转对话历史页
  - 我的知识库（教师可见）→ 跳转知识库管理页
  - 关于系统 → 显示版本信息
  - 退出登录 → 清除 token → 跳转登录页

**交互逻辑**：
1. 进入页面 → 调用 `GET /api/user/profile`（🆕 新增接口）
2. 根据角色动态显示功能列表
3. 退出登录 → 清除本地存储 → 跳转登录页

**接口依赖**：
- `GET /api/user/profile`（🆕 新增接口）

---

#### FE-P10: 记忆查看页

**功能描述**：学生查看系统为自己记录的学习记忆。

**UI 元素**：
- 教师筛选（下拉选择教师）
- 记忆类型筛选（Tab：全部 | 对话记忆 | 学习进度 | 个性特征）
- 记忆列表（每项：记忆内容 + 重要性标签 + 时间）
- 空状态提示

**交互逻辑**：
1. 进入页面 → 调用 `GET /api/memories?teacher_id={id}`
2. 切换教师/类型 → 重新请求
3. 下拉刷新

**接口依赖**：
- `GET /api/memories`
- `GET /api/teachers`（获取教师列表用于筛选）

---

### 3.4 前端公共模块

#### FE-C1: 网络请求封装

**功能**：统一的 HTTP 请求工具，封装 Taro.request。

**要求**：
- 自动附加 `Authorization: Bearer <token>` 请求头
- 统一错误处理：
  - 401 → 清除 token → 跳转登录页
  - 网络错误 → Toast 提示"网络异常"
  - 业务错误 → Toast 提示 message 字段
- 请求/响应拦截器
- 支持 loading 状态管理
- Base URL 可配置（开发环境 `http://localhost:8080`）

#### FE-C2: 状态管理（Store）

**功能**：全局状态管理。

**Store 模块**：
- `userStore`：用户信息、token、角色
- `chatStore`：当前对话消息列表、发送状态
- `teacherStore`：教师列表缓存

#### FE-C3: 路由守卫

**功能**：页面访问权限控制。

**规则**：
- 未登录 → 只能访问登录页和注册页
- 已登录 → 自动跳过登录页
- 角色不匹配 → 重定向到对应首页

#### FE-C4: TabBar 配置

**学生 TabBar**：
| Tab | 图标 | 页面 |
|-----|------|------|
| 首页 | home | `/pages/home/index` |
| 历史 | chat | `/pages/history/index` |
| 我的 | user | `/pages/profile/index` |

**教师 TabBar**：
| Tab | 图标 | 页面 |
|-----|------|------|
| 知识库 | book | `/pages/knowledge/index` |
| 我的 | user | `/pages/profile/index` |

> **注意**：小程序 TabBar 是全局配置，不同角色的 TabBar 需要通过自定义 TabBar 组件实现。

---

### 3.5 前端模块划分

| 模块编号 | 模块名称 | 包含页面/组件 | 优先级 |
|----------|----------|---------------|--------|
| FE-M1 | 项目脚手架 | Taro 初始化 + 目录结构 + 依赖安装 + 全局配置 | P0-第1层 |
| FE-M2 | 网络请求层 | request 封装 + 拦截器 + API 定义 | P0-第1层 |
| FE-M3 | 状态管理层 | userStore + chatStore + teacherStore | P0-第1层 |
| FE-M4 | 微信登录 | FE-P1 登录页 + FE-P2 角色选择页 + 路由守卫 | P0-第2层 |
| FE-M5 | 学生首页 | FE-P3 首页 + 教师卡片组件 | P0-第2层 |
| FE-M6 | 对话模块 | FE-P5 对话页 + 消息气泡组件 + 输入组件 | P0-第3层 |
| FE-M7 | 知识库模块 | FE-P7 知识库管理 + FE-P8 添加文档 | P0-第3层 |
| FE-M8 | 个人中心 | FE-P9 个人中心 + 自定义 TabBar | P0-第3层 |
| FE-M9 | 对话历史 | FE-P6 对话历史页 | P1-第4层 |
| FE-M10 | 记忆查看 | FE-P10 记忆查看页 | P1-第4层 |

**开发顺序**（按依赖层级）：
```
第1层（并行）: FE-M1 脚手架 + FE-M2 网络请求 + FE-M3 状态管理
      ↓
第2层（并行）: FE-M4 微信登录 + FE-M5 学生首页
      ↓
第3层（并行）: FE-M6 对话模块 + FE-M7 知识库模块 + FE-M8 个人中心
      ↓
第4层（并行）: FE-M9 对话历史 + FE-M10 记忆查看
```

---

## 4. 后端改造需求

### 4.1 新增接口

第一迭代的接口面向 curl/Postman 验证，第二迭代需要新增以下接口以支撑前端：

| 编号 | 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|------|
| BE-API-1 | POST | `/api/auth/wx-login` | 微信登录（code 换 token） | 无 |
| BE-API-2 | POST | `/api/auth/complete-profile` | 新用户补全信息（角色+昵称） | 需要 |
| BE-API-3 | GET | `/api/teachers` | 获取教师列表（学生选择教师用） | 需要 |
| BE-API-4 | GET | `/api/user/profile` | 获取当前用户信息 | 需要 |
| BE-API-5 | GET | `/api/conversations/sessions` | 获取会话列表（按教师分组） | 需要 |

### 4.2 接口改造

| 编号 | 接口 | 改造内容 |
|------|------|----------|
| BE-MOD-1 | CORS 配置 | 允许小程序域名，增加 `http://localhost:*` 通配 |
| BE-MOD-2 | `GET /api/conversations` | 支持不传 `teacher_id` 时返回所有教师的对话 |
| BE-MOD-3 | `users` 表 | 新增 `openid` 字段（微信 openid，唯一索引） |
| BE-MOD-4 | 认证插件 | 新增 `wx-login` 和 `complete-profile` action |

> **注意**：第一迭代的 `POST /api/auth/login` 和 `POST /api/auth/register` 接口**保留不删除**（用于集成测试和后台管理），但前端不再使用。

### 4.3 后端模块划分

| 模块编号 | 模块名称 | 改动范围 | 优先级 |
|----------|----------|----------|--------|
| BE-M1 | 微信登录接口 | `auth_plugin.go` + `wx_client.go` + `handlers.go` + `router.go` + `repository.go`（users 表加 openid） | P0 |
| BE-M2 | 补全信息接口 | `auth_plugin.go` + `handlers.go` + `router.go` | P0 |
| BE-M3 | 教师列表接口 | `handlers.go` + `router.go` + `repository.go` | P0 |
| BE-M4 | 用户信息接口 | `handlers.go` + `router.go` | P0 |
| BE-M5 | 会话列表接口 | `handlers.go` + `router.go` + `repository.go` | P1 |
| BE-M6 | CORS 配置适配 | `middleware.go` + `harness.yaml` | P0 |
| BE-M7 | 对话历史增强 | `handlers.go`（teacher_id 改为可选） | P0 |

**开发顺序**：
```
第1层（并行）: BE-M1 微信登录 + BE-M6 CORS
      ↓
第2层（并行）: BE-M2 补全信息 + BE-M3 教师列表 + BE-M4 用户信息 + BE-M7 对话历史增强
      ↓
第3层: BE-M5 会话列表
```

### 4.4 微信登录后端实现要点

#### 4.4.1 WxClient 接口设计

```go
// WxClient 微信 API 客户端接口
type WxClient interface {
    // Code2Session 用 code 换取 openid 和 session_key
    Code2Session(code string) (*WxSessionResult, error)
}

type WxSessionResult struct {
    OpenID     string `json:"openid"`
    SessionKey string `json:"session_key"`
    UnionID    string `json:"unionid"`
    ErrCode    int    `json:"errcode"`
    ErrMsg     string `json:"errmsg"`
}
```

#### 4.4.2 Mock 模式

集成测试和开发环境使用 `MockWxClient`，不依赖真实微信服务器：

```go
// MockWxClient 测试用 mock 客户端
type MockWxClient struct{}

func (m *MockWxClient) Code2Session(code string) (*WxSessionResult, error) {
    // 用 code 作为 openid 的一部分，方便测试区分不同用户
    return &WxSessionResult{
        OpenID: "mock_openid_" + code,
    }, nil
}
```

通过环境变量 `WX_MODE` 控制：
- `WX_MODE=mock` → 使用 MockWxClient
- `WX_MODE=real`（默认） → 使用真实微信 API

#### 4.4.3 数据库变更

`users` 表新增字段：

```sql
ALTER TABLE users ADD COLUMN openid TEXT DEFAULT '' ;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_openid ON users(openid) WHERE openid != '';
```

#### 4.4.4 环境变量新增

| 变量名 | 必填 | 默认值 | 说明 |
|--------|------|--------|------|
| `WX_APPID` | 生产必填 | - | 微信小程序 AppID |
| `WX_SECRET` | 生产必填 | - | 微信小程序 AppSecret |
| `WX_MODE` | ❌ | `real` | 微信 API 模式：`real` / `mock` |

---

## 5. 前端目录结构

```
src/frontend/
├── project.config.json          # 小程序项目配置
├── package.json                 # 依赖管理
├── tsconfig.json                # TypeScript 配置
├── babel.config.js              # Babel 配置
├── config/
│   ├── index.ts                 # Taro 编译配置
│   ├── dev.ts                   # 开发环境配置
│   └── prod.ts                  # 生产环境配置
├── src/
│   ├── app.ts                   # 应用入口
│   ├── app.config.ts            # 全局配置（页面路由、TabBar）
│   ├── app.scss                 # 全局样式
│   ├── api/
│   │   ├── request.ts           # 🆕 网络请求封装
│   │   ├── auth.ts              # 🆕 认证相关 API
│   │   ├── chat.ts              # 🆕 对话相关 API
│   │   ├── teacher.ts           # 🆕 教师相关 API
│   │   ├── knowledge.ts         # 🆕 知识库相关 API
│   │   ├── memory.ts            # 🆕 记忆相关 API
│   │   └── user.ts              # 🆕 用户相关 API
│   ├── store/
│   │   ├── index.ts             # 🆕 Store 入口
│   │   ├── userStore.ts         # 🆕 用户状态
│   │   ├── chatStore.ts         # 🆕 对话状态
│   │   └── teacherStore.ts      # 🆕 教师状态
│   ├── components/
│   │   ├── ChatBubble/          # 🆕 消息气泡组件
│   │   │   ├── index.tsx
│   │   │   └── index.scss
│   │   ├── TeacherCard/         # 🆕 教师卡片组件
│   │   │   ├── index.tsx
│   │   │   └── index.scss
│   │   ├── CustomTabBar/        # 🆕 自定义 TabBar
│   │   │   ├── index.tsx
│   │   │   └── index.scss
│   │   ├── TagInput/            # 🆕 标签输入组件
│   │   │   ├── index.tsx
│   │   │   └── index.scss
│   │   └── Empty/               # 🆕 空状态组件
│   │       ├── index.tsx
│   │       └── index.scss
│   ├── pages/
│   │   ├── login/               # 🆕 登录页
│   │   │   ├── index.tsx
│   │   │   ├── index.config.ts
│   │   │   └── index.scss
│   │   ├── role-select/         # 🆕 角色选择页（新用户）
│   │   │   ├── index.tsx
│   │   │   ├── index.config.ts
│   │   │   └── index.scss
│   │   ├── home/                # 🆕 首页（学生）
│   │   │   ├── index.tsx
│   │   │   ├── index.config.ts
│   │   │   └── index.scss
│   │   ├── chat/                # 🆕 对话页
│   │   │   ├── index.tsx
│   │   │   ├── index.config.ts
│   │   │   └── index.scss
│   │   ├── history/             # 🆕 对话历史页
│   │   │   ├── index.tsx
│   │   │   ├── index.config.ts
│   │   │   └── index.scss
│   │   ├── knowledge/           # 🆕 知识库管理
│   │   │   ├── index.tsx        # 文档列表
│   │   │   ├── add.tsx          # 添加文档
│   │   │   ├── index.config.ts
│   │   │   ├── add.config.ts
│   │   │   └── index.scss
│   │   ├── memories/            # 🆕 记忆查看
│   │   │   ├── index.tsx
│   │   │   ├── index.config.ts
│   │   │   └── index.scss
│   │   └── profile/             # 🆕 个人中心
│   │       ├── index.tsx
│   │       ├── index.config.ts
│   │       └── index.scss
│   └── utils/
│       ├── storage.ts           # 🆕 本地存储工具
│       ├── format.ts            # 🆕 格式化工具（时间、文本截断等）
│       └── constants.ts         # 🆕 常量定义
```

**统计**：🆕 新建约 45 个文件

---

## 6. UI 设计规范

### 6.1 色彩方案

| 用途 | 色值 | 说明 |
|------|------|------|
| 主色 | `#1890FF` | 按钮、链接、高亮 |
| 成功 | `#52C41A` | 成功提示 |
| 警告 | `#FAAD14` | 警告提示 |
| 错误 | `#FF4D4F` | 错误提示 |
| 背景 | `#F5F5F5` | 页面背景 |
| 卡片背景 | `#FFFFFF` | 卡片、输入框背景 |
| 主文字 | `#333333` | 标题、正文 |
| 次文字 | `#999999` | 辅助信息、时间戳 |
| 用户气泡 | `#1890FF` | 用户消息背景 |
| AI 气泡 | `#FFFFFF` | AI 回复背景 |

### 6.2 字体规范

| 用途 | 大小 | 粗细 |
|------|------|------|
| 页面标题 | 18px | bold |
| 卡片标题 | 16px | medium |
| 正文 | 14px | normal |
| 辅助文字 | 12px | normal |
| 消息文字 | 15px | normal |

### 6.3 间距规范

| 用途 | 值 |
|------|-----|
| 页面边距 | 16px |
| 卡片间距 | 12px |
| 卡片内边距 | 16px |
| 列表项间距 | 8px |
| 组件间距 | 12px |

---

## 7. 数据流设计

### 7.1 微信登录流程

```
用户点击"微信登录"按钮
    ↓
wx.login() → 获取临时 code
    ↓
调用 POST /api/auth/wx-login {code}
    ↓
后端: code → 微信 jscode2session → openid
    ↓
后端: 根据 openid 查找用户
    ├── 已有用户 → 生成 JWT → 返回 {token, is_new_user: false, role, nickname}
    └── 新用户 → 创建用户(role=空) → 生成 JWT → 返回 {token, is_new_user: true}
    ↓
前端: 存储 token 到 Taro.setStorageSync
    ↓
判断 is_new_user:
  true → 跳转 /pages/role-select/index（角色选择页）
  false → 根据 role 跳转：
    student → /pages/home/index
    teacher → /pages/knowledge/index
```

### 7.1.1 新用户补全信息流程

```
用户在角色选择页选择角色 + 填写昵称
    ↓
调用 POST /api/auth/complete-profile {role, nickname}
    ↓
后端: 更新用户的 role 和 nickname
    ↓
前端: 更新 userStore → 根据 role 跳转对应首页
```

### 7.2 对话流程

```
用户输入消息 → 点击发送
    ↓
chatStore.addMessage({role: "user", content: message})  // 乐观更新
    ↓
chatStore.setLoading(true)  // 显示 AI 思考中
    ↓
调用 POST /api/chat {message, teacher_id, session_id}
    ↓
成功 → chatStore.addMessage({role: "assistant", content: reply})
    ↓
chatStore.setLoading(false)
    ↓
scrollToBottom()  // 滚动到底部
```

### 7.3 Token 刷新流程

```
请求返回 401（token 过期）
    ↓
拦截器检查是否有 refresh token（本迭代暂不实现 refresh）
    ↓
清除本地 token → 跳转登录页
    ↓
Toast 提示"登录已过期，请重新登录"
```

---

## 8. 验收标准

### 8.1 前端功能验收

| 编号 | 验收项 | 验证方式 |
|------|--------|----------|
| FE-AC-01 | 微信登录按钮正常显示，点击触发 wx.login | 微信开发者工具 |
| FE-AC-02 | 新用户登录后跳转角色选择页，选择后跳转首页 | 微信开发者工具 |
| FE-AC-03 | 学生首页展示教师列表 | 微信开发者工具 |
| FE-AC-04 | 对话页能发送消息并收到 AI 回复 | 微信开发者工具 |
| FE-AC-05 | 对话页显示历史消息 | 微信开发者工具 |
| FE-AC-06 | 教师知识库列表正常展示 | 微信开发者工具 |
| FE-AC-07 | 添加文档成功并返回列表 | 微信开发者工具 |
| FE-AC-08 | 删除文档成功并刷新列表 | 微信开发者工具 |
| FE-AC-09 | 个人中心展示用户信息 | 微信开发者工具 |
| FE-AC-10 | 退出登录清除 token 并跳转 | 微信开发者工具 |
| FE-AC-11 | 未登录访问受保护页面自动跳转登录 | 微信开发者工具 |
| FE-AC-12 | 网络错误有友好提示 | 微信开发者工具 |

### 8.2 后端改造验收

| 编号 | 验收项 | 验证方式 |
|------|--------|----------|
| BE-AC-01 | `POST /api/auth/wx-login` 微信登录返回 token | curl（mock 模式） |
| BE-AC-02 | `POST /api/auth/complete-profile` 补全角色和昵称 | curl |
| BE-AC-03 | `GET /api/teachers` 返回教师列表 | curl |
| BE-AC-04 | `GET /api/user/profile` 返回当前用户信息 | curl |
| BE-AC-05 | `GET /api/conversations/sessions` 返回会话列表 | curl |
| BE-AC-06 | CORS 允许小程序域名 | 前端联调 |
| BE-AC-07 | `users` 表 openid 字段正常工作 | 集成测试 |

### 8.3 集成测试验收

> **测试策略**：分两层测试。第1层为后端 API 集成测试（Go httptest，`ci_test_agent` 负责），第2层为前端组件单测（各前端模块的单测阶段覆盖）。

#### 第1层：后端 API 集成测试（沿用 httptest 模式）

在 `tests/integration/integration_test.go` 中新增以下用例（编号从 IT-18 开始，延续第一迭代）：

| 编号 | 验收项 | 验证方式 |
|------|--------|----------|
| IT-18 | 微信登录（mock 模式）→ 返回 token + is_new_user=true | httptest |
| IT-19 | 新用户补全信息 → 设置角色和昵称 | httptest |
| IT-20 | 同一 openid 再次登录 → is_new_user=false | httptest |
| IT-21 | 获取教师列表 → 返回教师数组 + document_count | httptest |
| IT-22 | 获取用户信息 → 返回 profile + stats | httptest |
| IT-23 | 获取会话列表 → 返回会话摘要 | httptest |
| IT-24 | 对话历史不传 teacher_id → 返回所有对话 | httptest |
| IT-25 | 微信登录→补全信息→教师列表→对话 全链路 | httptest |
| IT-26 | 教师微信登录→补全信息→添加文档→学生对话引用知识 | httptest |
| IT-27 | 多轮对话→查看会话列表→查看对话历史→查看记忆 | httptest |

**微信登录测试策略**：
- 后端实现 `WxClient` 接口，生产环境调用真实微信 API
- 测试环境注入 `MockWxClient`，通过 `WX_MODE=mock` 环境变量控制
- MockWxClient 用 code 生成固定的 openid（如 `mock_openid_{code}`），无需真实微信服务器

#### 第2层：前端组件测试（各模块单测阶段覆盖）

前端不做独立集成测试，在各模块的单测阶段由 `dev_frontent_agent` 完成：
- 组件渲染正确性
- 状态管理逻辑
- API 调用参数正确性（mock request）
- 本迭代不做前端 E2E 测试（在开发者工具中手动验证）

---

## 9. 外部依赖清单

### 9.1 前端依赖

| 依赖 | 版本 | 用途 |
|------|------|------|
| `@tarojs/cli` | 3.x | Taro CLI |
| `@tarojs/taro` | 3.x | Taro 运行时 |
| `@tarojs/components` | 3.x | Taro 组件 |
| `@tarojs/runtime` | 3.x | Taro 运行时 |
| `react` | 18.x | UI 框架 |
| `react-dom` | 18.x | React DOM |
| `typescript` | 5.x | TypeScript |
| `sass` | latest | CSS 预处理器 |
| `zustand` | 4.x | 状态管理（轻量） |

> **关于 UI 组件库**：Taro 3.x 对 Vant Weapp 的兼容性需要验证。如果不兼容，使用 Taro UI 或纯手写组件。前端 Agent 在 FE-M1 脚手架搭建时需要验证并确定最终方案。

### 9.2 后端新增依赖

| 依赖 | 用途 |
|------|------|
| `net/http`（标准库） | 调用微信 jscode2session API |

> 无需新增第三方 Go 依赖，微信 API 调用使用标准库 `net/http` + `encoding/json` 即可。

---

## 10. 风险与应对

| 风险 | 影响 | 应对方案 |
|------|------|----------|
| Taro + Vant Weapp 兼容性问题 | 前端 UI 组件不可用 | 备选 Taro UI 或手写组件 |
| 小程序 TabBar 不支持动态切换 | 不同角色 TabBar 不同 | 使用自定义 TabBar 组件 |
| 微信开发者工具环境差异 | 本地开发正常但工具中异常 | 持续在开发者工具中验证 |
| 对话响应时间长（大模型调用） | 用户体验差 | 添加 loading 动画 + 超时提示 |
| CORS 跨域问题 | 前端请求被拦截 | 后端 CORS 配置适配 |
| 微信 jscode2session 调用失败 | 无法获取 openid | 开发/测试环境使用 MockWxClient，不依赖微信服务器 |
| 微信 AppID/Secret 泄露 | 安全风险 | 通过环境变量注入，不写入代码和配置文件 |

---

**文档版本**: v1.1.0
**创建日期**: 2026-03-28
**最后更新**: 2026-03-28
**变更记录**:
- v1.0.0: 初始版本（用户名密码登录）
- v1.1.0: 登录方式改为微信登录，新增集成测试详细方案
