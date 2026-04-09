# V2.0 IT13 UI规范文档

## 迭代信息

- **版本**: V2.0
- **迭代**: IT13
- **主题**: 全平台UI规范（基于代码全量扫描）
- **生成时间**: 2026-04-09
- **覆盖范围**: Taro小程序主应用(34页面) + H5教师端(13页面) + H5学生端(7页面) + H5管理后台(7页面)

---

## 目录

- [一、项目架构总览](#一项目架构总览)
- [二、Taro小程序主应用](#二taro小程序主应用)
- [三、H5教师端](#三h5教师端)
- [四、H5学生端](#四h5学生端)
- [五、H5管理后台](#五h5管理后台)
- [六、公共组件](#六公共组件)
- [七、API接口汇总](#七api接口汇总)

---

## 一、项目架构总览

### 1.1 技术栈

| 应用 | 技术栈 | UI框架 | 端口 |
|---|---|---|---|
| 小程序主应用 (frontend) | Taro 3 + React + TSX | Taro UI | - |
| H5教师端 (h5-teacher) | Vue 3 + Vue Router + Pinia | Element Plus | 3001 |
| H5学生端 (h5-student) | Vue 3 + Vue Router + Pinia | Vant 4 | 3002 |
| H5管理后台 (h5-admin) | Vue 3 + Vue Router + Pinia | Element Plus + ECharts | 5173 |

### 1.2 角色体系

| 角色 | 说明 | 可访问端 |
|---|---|---|
| 教师 (teacher) | 创建班级、管理学生、知识库、课程、查看对话 | 小程序 + H5教师端 |
| 学生 (student) | AI对话、加入班级、查看课程、提交反馈 | 小程序 + H5学生端 |
| 管理员 (admin) | 系统管理、用户管理、反馈处理、日志查看 | H5管理后台 |

### 1.3 登录认证

| 端 | 登录方式 | API |
|---|---|---|
| 小程序 | 微信小程序登录 | `POST /api/auth/wx-login` |
| H5端 | 微信H5 OAuth | `GET /api/auth/wx-h5-login-url` → `POST /api/auth/wx-h5-callback` |

### 1.4 路由守卫

所有端均实现路由守卫：未登录用户访问需认证页面时，自动跳转到 `/login`。

---

## 二、Taro小程序主应用

### 2.1 TabBar配置

**学生端 TabBar（3项）**：

| 序号 | Tab名称 | 图标 | 路径 |
|---|---|---|---|
| 1 | 聊天 | chat | /pages/chat-list/index |
| 2 | 发现 | discover | /pages/discover/index |
| 3 | 我的 | profile | /pages/profile/index |

> 注意：对话页（/pages/chat/index）由聊天列表点击会话进入，**不作为独立 Tab 展示**。

**教师端 TabBar（4项）**：

| 序号 | Tab名称 | 图标 | 路径 |
|---|---|---|---|
| 1 | 聊天 | chat | /pages/chat-list/index |
| 2 | 学生 | student | /pages/teacher-students/index |
| 3 | 知识库 | knowledge | /pages/knowledge/index |
| 4 | 我的 | profile | /pages/profile/index |

> 注意：教师端不展示"对话"和"发现" Tab。对话页由聊天列表点击学生进入。

---

### 2.2 登录页 (login)

**路径**: `/pages/login/index`

| 元素 | 类型 | 内容 |
|---|---|---|
| Logo | 图片 | 应用Logo |
| 标题 | 文本 | 数字分身 |
| 登录按钮 | 按钮 | `微信登录` |
| API | 接口 | `POST /api/auth/wx-login` |

**跳转逻辑**:
- `is_new_user` = true → `/pages/role-select/index`
- personas数量 > 1 → `/pages/persona-select/index`
- 否则 → 首页

---

### 2.3 角色选择页 (role-select)

**路径**: `/pages/role-select/index`

| 元素 | 类型 | 内容 |
|---|---|---|
| 页面标题 | 文本 | `选择你的身份` |
| 教师卡片 | 卡片 | 👨‍🏫 图标 + `我是教师` |
| 学生卡片 | 卡片 | 👨‍🎓 图标 + `我是学生` |
| API | 接口 | `POST /api/personas` (创建分身) |

**跳转逻辑**: 选择角色后创建分身 → 首页

---

### 2.4 分身选择页 (persona-select)

**路径**: `/pages/persona-select/index`

| 元素 | 类型 | 内容 |
|---|---|---|
| 教师分身分组 | 列表 | 教师角色分身列表 |
| 学生分身分组 | 列表 | 学生角色分身列表 |
| 创建新身份按钮 | 按钮 | `创建新身份` |
| API | 接口 | `GET /api/personas` (获取分身列表) |

---

### 2.5 分身概览页 (persona-overview)

**路径**: `/pages/persona-overview/index`

| 元素 | 类型 | 内容 |
|---|---|---|
| 页面标题 | 导航栏 | `分身概览` |
| 分身卡片列表 | 列表 | 昵称、学校、统计数据 |
| 公开/私有开关 | Switch | 切换分身可见性 |
| 统计数据 | 文本 | 学生数、文档数、班级数 |
| API | 接口 | `GET /api/personas/{id}/dashboard` |

---

### 2.6 首页 (home)

**路径**: `/pages/home/index`

根据当前分身角色渲染不同组件：
- **教师角色** → `TeacherDashboard` 组件
- **学生角色** → `StudentHome` 组件

---

### 2.7 聊天列表页 (chat-list) ⭐核心页面

**路径**: `/pages/chat-list/index` | **代码行数**: 681行

#### 学生端视图

| 元素 | 类型 | 内容 |
|---|---|---|
| 老师分组 | 分组列表 | 按老师分身ID分组 |
| 老师头像/昵称 | 列表头 | 老师信息 |
| 会话列表 | 子列表 | 最新消息、时间、未读数 |
| 历史会话展开 | 折叠 | 展开历史会话 |
| 新建对话按钮 | 按钮 | 创建新会话 |
| 置顶 | 操作 | 长按置顶老师 |
| 快捷指令 | 入口 | 快捷发送预设指令 |

#### 教师端视图

| 元素 | 类型 | 内容 |
|---|---|---|
| 班级分组 | 分组列表 | 按班级分组 |
| 学生子列表 | 子列表 | 头像、昵称、最新消息、未读badge |
| 展开/收起 | 箭头 | 折叠班级内学生 |
| 置顶 | 图标 | ⭐ 置顶图标 |

**API**: `GET /api/chat-list/student` | `GET /api/chat-list/teacher`

**时间格式化**: 刚刚 → X分钟前 → 时:分 → 昨天 → 周X → 月/日

---

### 2.8 对话页 (chat) ⭐核心页面

**路径**: `/pages/chat/index` | **代码行数**: 810行

> **入口**：从聊天列表点击会话进入，不在 TabBar 单独展示。

#### 学生端视图

| 元素 | 类型 | 内容 |
|---|---|---|
| 导航栏 | NavBar | 老师昵称（居中）；**无右上角"详情"按钮**（避免与小程序胶囊冲突） |
| 消息列表 | 滚动列表 | 聊天气泡(文本/图片/语音)，AI气泡左白色，自己气泡右绿色 |
| SSE流式显示 | 实时 | AI回复逐字显示 (enableChunked) |
| 思考步骤 | 折叠面板 | ThinkingPanel 组件，显示 AI 推理步骤 |
| 教师接管标识 | 状态标签 | 🤖 AI自动回复 / 👨‍🏫 教师已接管 |
| 输入框 | Input | `placeholder="输入消息..."` |
| 语音按钮 | 圆形细描边 + 内置麦克风SVG线条图标 | VoiceInput 组件，切换语音/文字模式 |
| Emoji面板 | 面板 | EmojiPanel 组件，点击😊触发 |
| + 号面板 | 面板 | PlusPanel 组件，包含：拍照/相册/文件/**新会话** |
| **无发送按钮** | — | 输入框区域不展示发送按钮 |
| **暂停按钮** | 圆形双竖线图标 | AI 回复流式输出时，在消息流底部显示暂停圆形按钮，点击打断回复 |
| 引用回复 | 引用条 | 引用消息内容，显示在输入框上方 |

#### 教师端视图（从聊天列表→学生进入）

| 元素 | 类型 | 内容 |
|---|---|---|
| 导航栏 | NavBar | 学生昵称（居中）；**左返回，右侧无任何按钮** |
| 消息列表顶部 | 提示条 | 橙色左边框提示："👁️ 教师视角：正在查看学生与AI的对话" |
| 消息列表 | 滚动列表 | 显示学生与AI的完整对话记录 |
| 接管标识 | 居中分割线 | `——— 👨‍🏫 教师已接管 ———`，以分割线形式插入消息流中间 |
| 输入栏 | 与学生端一致 | 🔊语音(白色圆形) + 输入框 + 😊Emoji + ➕ + 发送 |
| + 号面板 | 面板 | 含：拍照/相册/文件；**不含"新会话"入口** |

**API**:
- `POST /api/chat/send` (发送消息)
- `GET /api/chat/messages` (历史消息)
- SSE: 微信小程序 `wx.request` + `enableChunked`
- `POST /api/chat/teacher-reply` (教师直接回复)
- `POST /api/chat/takeover` (教师接管)

---

### 2.9 对话历史页 (history)

**路径**: `/pages/history/index`

| 元素 | 类型 | 内容 |
|---|---|---|
| 会话列表 | 列表 | 老师昵称、最后消息、消息数 |
| 时间显示 | 文本 | 格式化时间 |
| API | 接口 | `GET /api/sessions` |

---

### 2.10 发现页 (discover) ⭐

**路径**: `/pages/discover/index` | **代码行数**: 420行

| 元素 | 类型 | 内容 |
|---|---|---|
| 搜索框 | SearchBar | `placeholder="搜索老师、班级..."` |
| 学科浏览 | 标签组 | 学科分类标签 |
| 热门班级 | 卡片列表 | 班级名称、老师、学生数 |
| 推荐老师 | 卡片列表 | TeacherCard 组件 |
| 广场 | 列表 | 公开分身/班级兼容 |
| API | 接口 | `GET /api/discover/recommend`, `GET /api/discover/search` |

---

### 2.11 个人中心页 (profile)

**路径**: `/pages/profile/index`

| 元素 | 类型 | 内容 |
|---|---|---|
| 用户信息卡 | 卡片 | 头像、昵称、角色标签 |
| 统计数据 | 数据行 | 对话数/知识数/班级数 等 |

#### 功能菜单（按角色过滤）

**教师菜单**:
| 菜单项 | 跳转 |
|---|---|
| 分身概览 | `/pages/persona-overview/index` |
| 分享管理 | `/pages/share-manage/index` |
| 反馈管理 | `/pages/feedback-manage/index` |

**切换角色入口**：用户名右侧展示「🔄 切换角色」胶囊按钮，点击进入 `/pages/persona-select/index`

**学生菜单**:
| 菜单项 | 跳转 | 说明 |
|---|---|---|
| 我的教师 | `/pages/my-teachers/index` | 已授权/审批中教师列表 |
| 我的记忆 | `/pages/memories/index` | AI 对话记忆管理 |
| 我的评语 | `/pages/my-comments/index` | 查看教师写给自己的鼓励性评语 |
| 意见反馈 | `/pages/feedback/index` | 提交使用反馈 |

**切换身份入口**：用户名右侧展示「🔄 切换身份」绿色胶囊按钮，点击进入 `/pages/persona-select/index`

> **备注与评语区分说明**：
> - **学生备注** `📋 🔒 仅教师可见`：教师在学生详情页填写的私密记录（性格/进度/注意事项），学生无法查看
> - **学生评语** `✍️ 👁️ 学生可见`：教师写的鼓励性点评/阶段总结，学生可在「我的评语」页查看

---

### 2.12 学生资料填写页 (student-profile)

**路径**: `/pages/student-profile/index`

| 字段 | 类型 | 说明 |
|---|---|---|
| 年龄 | 数字输入 | 学生年龄 |
| 性别 | 选择器 | 男/女 |
| 家庭情况 | 文本域 | 可选填写 |
| 跳过按钮 | 按钮 | `跳过` (可跳过) |
| 提交按钮 | 按钮 | `完成` |
| API | 接口 | `PUT /api/users/student-profile` |

---

### 2.13 知识库列表页 (knowledge)

**路径**: `/pages/knowledge/index`

| 元素 | 类型 | 内容 |
|---|---|---|
| 搜索框 | SearchBar | 搜索知识条目 |
| 类型筛选Tab | Tab组 | `全部` / `链接` / `文字` / `文件` |
| 知识列表 | 列表 | 标题、类型标签、时间 |
| 统一输入框 | Input | 智能识别URL/文本 |
| 文件上传 | 按钮 | 选择文件上传 |
| 左滑操作 | 滑动 | 删除/重命名 |
| 添加按钮 | FAB | 跳转添加页 |
| API | 接口 | `GET /api/knowledge`, `DELETE /api/knowledge/{id}` |

---

### 2.14 知识添加页 (knowledge/add) ⭐

**路径**: `/pages/knowledge/add` | **代码行数**: 811行

#### Tab切换

| Tab | 内容 |
|---|---|
| 文本 | 文本内容输入 + 标题 |
| 文件 | 文件选择上传 |
| URL | URL链接输入 |
| 聊天记录 | 选择聊天记录导入 |
| 批量上传 | 多文件批量上传 |

#### Scope选择

| 选项 | 说明 |
|---|---|
| 全部 | 所有班级/学生可用 |
| 指定班级 | 选择特定班级 |
| 指定学生 | 选择特定学生 |

**API**: `POST /api/knowledge` + 批量上传状态轮询

---

### 2.15 学生管理页 (teacher-students) ⭐

**路径**: `/pages/teacher-students/index` | **代码行数**: 632行

#### Tab切换

| Tab | 内容 |
|---|---|
| 全部学生 | 所有学生列表 |
| 按班级 | 按班级分组查看 |
| 待审批 | 待审批加入请求 |
| 班级设置 | 班级管理操作 |

| 元素 | 类型 | 内容 |
|---|---|---|
| 邀请学生弹窗 | 弹窗 | 分享码/链接邀请 |
| 启停开关 | Switch | 启用/禁用学生 |
| API | 接口 | `GET /api/personas/search-students`, `PUT /api/relations/{id}/toggle` |

---

### 2.16 创建班级页 (class-create)

**路径**: `/pages/class-create/index`

**V11版本**: 创建班级时同步创建分身

#### 表单字段

| 字段 | 必填 | placeholder | maxLength |
|---|---|---|---|
| 分身昵称 | ✅ | `请输入分身昵称（如：王老师）` | 30 |
| 学校名称 | ✅ | `请输入学校名称` | 50 |
| 分身描述 | ✅ | `请输入分身描述（教学风格、擅长领域等）` | 200 |
| 班级名称 | ✅ | `请输入班级名称` | 50 |
| 班级描述 | ❌ | `请输入班级描述（可选）` | 200 |
| 公开开关 | ❌ | Switch: 公开/私密 | - |

**公开提示文案**:
- 公开: `当前班级公开，所有学生可见` (绿色)
- 私密: `当前班级私密，仅受邀学生可加入` (橙色)

**创建成功展示**:
- 班级分身信息卡片(昵称、ID、学校)
- 分享链接(可复制)
- 分享码(可复制)
- 提示: `💡 将分享链接或分享码发给学生，即可邀请他们加入班级`

**API**: `POST /api/classes` (V11版本含分身创建)

---

### 2.17 班级详情页 (class-detail)

**路径**: `/pages/class-detail/index`

| 元素 | 类型 | 内容 |
|---|---|---|
| 班级信息 | 描述列表 | 名称、学生数、创建时间、描述 |
| 学生列表 | 表格 | 昵称、状态(正常/禁用)、最后活跃 |
| 添加学生 | 按钮 | 打开添加弹窗 |
| 快捷操作 | 按钮组 | 对话记录/写评语/设置风格 |
| 启停班级/学生 | Switch | 启用/禁用 |

---

### 2.18 编辑班级页 (class-edit)

**路径**: `/pages/class-edit/index`

| 字段 | 说明 |
|---|---|
| 班级名称 | 可编辑 |
| 班级描述 | 可编辑 |
| 公开设置 | Switch切换 |

**API**: `PUT /api/classes/{id}`

---

### 2.19 加入分享页 (share-join)

**路径**: `/pages/share-join/index`

**流程**: 输入分享码 → 查看老师信息 → 选择分身 → 确认加入

#### 5种状态

| 状态 | 说明 | 展示 |
|---|---|---|
| can_join | 可以加入 | 显示确认按钮 |
| already_joined | 已加入 | 提示已加入 |
| not_target | 非目标用户 | 提示不符合条件 |
| need_login | 需要登录 | 跳转登录 |
| need_persona | 需要创建分身 | 跳转创建分身 |

---

### 2.20 分享管理页 (share-manage)

**路径**: `/pages/share-manage/index`

| 元素 | 类型 | 内容 |
|---|---|---|
| 有效分享码列表 | 列表 | 分享码、关联班级、使用次数 |
| 失效分享码列表 | 列表 | 已过期/停用的分享码 |
| 二维码生成 | Canvas | 生成分享二维码 |
| 复制按钮 | 按钮 | 复制分享码/链接 |
| 停用按钮 | 按钮 | 停用分享码 |
| 新建弹窗 | 弹窗 | 选择关联班级生成新分享码 |

---

### 2.21 学生对话记录页 (student-chat-history)

**路径**: `/pages/student-chat-history/index`

**教师视角**: 查看学生与AI的对话

| 元素 | 类型 | 内容 |
|---|---|---|
| 对话消息列表 | 列表 | 学生消息 + AI回复 |
| 教师真人回复 | 标识 | 区分AI回复和教师回复 |
| 接管按钮 | 按钮 | 教师接管对话 |
| 退出接管 | 按钮 | 退出接管模式 |
| 引用回复 | 操作 | 引用某条消息回复 |

---

### 2.22 学生详情页 (student-detail)

**路径**: `/pages/student-detail/index`

#### 问答风格设置

| 设置项 | 类型 | 选项/内容 |
|---|---|---|
| 教学风格 | 6选1 | 严谨/温和/幽默/鼓励/引导/简洁 |
| 引导程度 | 滑块 | 低→高 |
| 风格描述 | 文本域 | 自定义风格描述 |

#### 学生备注 🔒 仅教师可见

> 用于教师私密记录学生性格特点、学习进度、注意事项等，**不对学生展示**。

| 元素 | 说明 |
|---|---|
| 区块标题 | `📋 学生备注` + 橙色「🔒 仅教师可见」标签 |
| 副提示 | `记录学生性格、学习进度、注意事项等（学生不可见）` |
| 备注卡片 | 橙色左边框 + 备注内容 + 进度日期 |
| 添加按钮 | `+ 添加` |

#### 学生评语 👁️ 学生可见

> 教师写给学生的鼓励性点评或阶段总结，**学生可以在「我的评语」中查看**。

| 元素 | 说明 |
|---|---|
| 区块标题 | `✍️ 学生评语` + 绿色「👁️ 学生可见」标签 |
| 副提示 | `鼓励性点评、阶段总结等（学生可在「我的评语」中查看）` |
| 评语卡片 | 绿色左边框 + 评语正文 + 日期 |
| 写评语按钮 | `+ 写评语` |

---

### 2.23 记忆页-学生端 (memories)

**路径**: `/pages/memories/index`

| 元素 | 类型 | 内容 |
|---|---|---|
| 教师筛选 | Picker | 选择老师过滤记忆 |
| 记忆层级Tab | Tab组 | `全部` / `核心` / `情景` / `归档` |
| 记忆列表 | 卡片列表 | 记忆内容、类型标签、重要性标签 |

---

### 2.24 记忆管理页-教师端 (memory-manage)

**路径**: `/pages/memory-manage/index`

| 元素 | 类型 | 内容 |
|---|---|---|
| 层级筛选 | Tab/Select | 筛选记忆层级 |
| 记忆列表 | 列表 | 记忆内容、层级、时间 |
| 摘要合并 | 按钮 | 合并多条记忆为摘要 |
| 编辑记忆 | 操作 | 编辑记忆内容 |
| 删除记忆 | 操作 | 删除记忆 |
| 分页 | 分页器 | 分页加载 |

---

### 2.25 我的教师页 (my-teachers)

**路径**: `/pages/my-teachers/index`

| 元素 | 类型 | 内容 |
|---|---|---|
| 已授权教师列表 | 列表 | 教师头像、昵称、学校 |
| 审批中教师列表 | 列表 | 等待审批的教师 |

---

### 2.26 我的评语页 (my-comments)

**路径**: `/pages/my-comments/index` | **角色**: 学生端

> 展示教师写给该学生的「学生评语」（`👁️ 学生可见` 部分），**不含**教师私密的「学生备注」。

| 元素 | 类型 | 内容 |
|---|---|---|
| 页面标题 | NavBar | 我的评语（带返回按钮） |
| 评语卡片列表 | 卡片 | 教师昵称（绿色） + 日期 + 评语正文（绿色左边框卡片） |
| 空状态 | Empty | `暂无评语` |

---

### 2.27 教材配置页 (curriculum-config) ⭐

**路径**: `/pages/curriculum-config/index` | **代码行数**: 457行

| 元素 | 类型 | 内容 |
|---|---|---|
| 引导步骤指示器 | 步骤条 | 步骤进度 |
| 学段选择 | Picker | K12/大学/成人 |
| 年级选择 | 选择器 | 根据学段动态显示 |
| 教材版本 | 多选标签 + 手动输入 | 人教版/苏教版/北师大版 等 |
| 学科选择 | 选择器 | 语文/数学/英语 等 |
| 教学进度 | Input | 当前教学进度 |
| 已有配置列表 | 列表 | 编辑/删除已有配置 |

---

### 2.28 意见反馈页 (feedback)

**路径**: `/pages/feedback/index`

| 字段 | 类型 | 验证 |
|---|---|---|
| 反馈类型 | 选择器 | 必选 |
| 反馈内容 | 文本域 | ≥5个字 |
| 设备上下文 | 自动采集 | 系统/版本/网络 |

**API**: `POST /api/feedbacks`

---

### 2.29 反馈管理页 (feedback-manage)

**路径**: `/pages/feedback-manage/index`

| 元素 | 类型 | 内容 |
|---|---|---|
| 状态筛选Tab | Tab组 | 全部/待处理/已查看/已解决 |
| 反馈卡片 | 卡片列表 | 类型、状态标签、内容、用户、时间 |
| 状态更新 | 点击操作 | 待处理→已查看→已解决 |

---

### 2.30 审批管理页 (approval-manage)

**路径**: `/pages/approval-manage/index`

| 元素 | 类型 | 内容 |
|---|---|---|
| 待审批列表 | 列表 | 学生昵称、申请时间、来源班级 |
| 点击跳转 | 导航 | 跳转审批详情页 |

---

### 2.31 审批详情页 (approval-detail)

**路径**: `/pages/approval-detail/index`

| 元素 | 类型 | 内容 |
|---|---|---|
| 学生信息 | 描述列表 | 年龄/性别/家庭情况(可编辑) |
| 教师评价 | 文本域 | 审批评价 |
| 通过按钮 | 按钮 | `通过` (绿色) |
| 拒绝按钮 | 按钮 | `拒绝` (红色) |

---

### 2.32 课程列表页 (course-list)

**路径**: `/pages/course-list/index`

| 元素 | 类型 | 内容 |
|---|---|---|
| 班级筛选Tab | Tab组 | 按班级筛选 |
| 课程卡片 | 卡片列表 | 标题、时间、内容摘要 |
| 编辑按钮 | 按钮 | 编辑课程 |
| 删除按钮 | 按钮 | 删除课程 |
| 推送按钮 | 按钮 | 推送课程给学生 |
| 新建浮动按钮 | FAB | 跳转课程发布页 |

---

### 2.33 课程发布页 (course-publish)

**路径**: `/pages/course-publish/index`

| 字段 | 类型 | 说明 |
|---|---|---|
| 课程标题 | Input | 必填 |
| 课程内容 | 文本域 | 必填 |
| 选择班级 | Picker | 选择关联班级 |
| 推送给学生 | Switch | 是否立即推送 |

---

### 2.34 自测学生页 (test-student)

**路径**: `/pages/test-student/index`

| 元素 | 类型 | 内容 |
|---|---|---|
| 测试学生信息卡 | 卡片 | 测试学生基本信息 |
| 已加入班级 | 列表 | 测试学生加入的班级 |
| 模拟登录按钮 | 按钮 | 模拟学生身份登录 |
| 重置数据按钮 | 按钮 | 重置测试数据 |
| API | 接口 | `GET /api/test-student`, `POST /api/test-student/reset` |

---

### 2.35 消息推送页 (teacher-message)

**路径**: `/pages/teacher-message/index`

| 字段 | 类型 | 说明 |
|---|---|---|
| 推送目标 | 选择器 | 班级/指定学生 |
| 班级选择 | Picker | 选择班级 |
| 学生ID输入 | Input | 指定学生ID |
| 消息内容 | 文本域 | 推送消息内容 |
| 推送历史列表 | 列表 | 历史推送记录 |

---

### 2.36 批量添加学生页 (student-batch) ⭐

**路径**: `/pages/student-batch/index` | **代码行数**: 398行

**三步流程**:

| 步骤 | 内容 |
|---|---|
| 1. 输入文本 | 粘贴/输入花名册文本 |
| 2. AI识别确认 | AI智能解析，可展开编辑每个学生的详细信息 |
| 3. 创建结果 | 显示批量创建结果(成功/失败数) |

---

## 三、H5教师端

### 3.1 路由结构

| 路径 | 名称 | meta.title | 认证 |
|---|---|---|---|
| `/login` | Login | - | 否 |
| `/role-select` | RoleSelect | - | 否 |
| `/` | Layout | - | 是 |
| `/home` | Home | `工作台` | 是 |
| `/chat-list` | ChatList | `消息列表` | 是 |
| `/chat/:id` | Chat | `对话` | 是 |
| `/classes` | Classes | `班级管理` | 是 |
| `/class/:id` | ClassDetail | `班级详情` | 是 |
| `/knowledge` | Knowledge | `知识库` | 是 |
| `/courses` | Courses | `课程管理` | 是 |
| `/personas` | Personas | `分身管理` | 是 |
| `/profile` | Profile | `个人中心` | 是 |
| `/:pathMatch(.*)*` | NotFound | - | 否 |

默认重定向: `/` → `/chat-list`

---

### 3.2 布局框架 (Layout)

| 元素 | 内容 |
|---|---|
| 侧边栏Logo | `教师工作台` |
| 侧边栏菜单 | 聊天列表(ChatDotSquare) / 学生管理(School) / 知识库(Collection) / 我的(User) / 课程管理(Document) / 分身管理(Avatar) |
| 顶部栏 | 折叠按钮 + 页面标题 + 用户下拉(头像+昵称→退出登录) |
| 退出确认 | `确定要退出登录吗？` |

---

### 3.3 登录页 (Login)

| 元素 | 内容 |
|---|---|
| 标题 | `数字分身教师端` |
| 副标题 | `微信授权登录` |
| 状态 | 加载中: `正在登录中...` / 错误: 动态错误信息 / `重新登录` 按钮 |
| 主按钮 | `微信授权登录` |
| 角色校验 | 非教师 → `您不是教师角色，无法访问此页面` |

---

### 3.4 角色选择页 (RoleSelect)

| 元素 | 内容 |
|---|---|
| 标题 | `选择您的角色` / `请选择您在系统中的身份` |
| 角色卡片 | `我是教师` / `我是学生` |
| 昵称输入 | `placeholder="请输入您的昵称"` |
| 确认按钮 | `确认` (需选角色+输入昵称) |
| 跳转 | 教师→`/home`，学生→外部跳转`/h5-student/` |

---

### 3.5 工作台首页 (Home)

**统计卡片（4张）**:
| 卡片 | 字段 |
|---|---|
| 我的学生 | `studentCount` |
| 班级数量 | `classCount` |
| 今日消息 | `todayMessages` |
| 知识条目 | `knowledgeCount` |

**最近对话表格**: 学生 / 最新消息 / 时间

**快捷操作**: `管理班级`(primary) / `添加知识`(success) / `发布课程`(warning)

---

### 3.6 聊天列表 (ChatList)

| 元素 | 内容 |
|---|---|
| 标题 | `学生消息` / `共 X 个班级` |
| 搜索框 | `placeholder="搜索学生姓名"` |
| 班级层级 | 班级名称 + 科目标签 + 学生数量 + 置顶⭐ + 展开箭头 |
| 学生层级 | 头像(昵称首字母) + 昵称 + 置顶 + 时间 + 最后消息 + 未读badge |
| 空状态 | `暂无班级聊天记录` / `创建班级并添加学生后即可查看` |
| 班级空 | `该班级暂无学生` |

**API**: `GET /api/chat-list/teacher`

---

### 3.7 对话详情 (Chat)

| 元素 | 内容 |
|---|---|
| 头部 | 学生昵称 |
| 消息列表 | 内容 + 时间，自己消息靠右蓝色背景 |
| 输入框 | `placeholder="输入消息..."` + Enter发送 |
| 发送按钮 | `发送` |

---

### 3.8 班级管理 (Classes) ⭐

**班级列表表格**:
| 列 | 说明 |
|---|---|
| 班级名称 | prop: name |
| 分身昵称 | prop: persona_nickname |
| 公开状态 | 标签: 公开(success)/私密(info) |
| 学生数 | prop: student_count |
| 创建时间 | prop: created_at |
| 操作 | 详情/编辑/删除 |

**创建班级弹窗** (600px, 含分身同步创建):

分身信息区:
| 字段 | 必填 | placeholder | maxLength |
|---|---|---|---|
| 分身昵称 | ✅ | `请输入分身昵称（如：王老师）` | 30 |
| 学校名称 | ✅ | `请输入学校名称` | 50 |
| 分身描述 | ✅ | `请输入分身描述（教学风格、擅长领域等）` | 200 |

班级信息区:
| 字段 | 必填 | placeholder | maxLength |
|---|---|---|---|
| 班级名称 | ✅ | `请输入班级名称` | 50 |
| 班级描述 | ❌ | `请输入班级描述（可选）` | 200 |
| 公开班级 | ❌ | Switch: 公开/私密 | - |

提示: `创建班级时会同步创建该班级专属的分身` (el-alert)

**创建成功弹窗**: 班级分身信息 + 分享链接/码 + 复制按钮

**编辑班级弹窗** (500px): 班级名称/描述/公开设置

**API**: `GET /api/classes` / `POST /api/classes` / `PUT /api/classes/{id}` / `DELETE /api/classes/{id}`

---

### 3.9 班级详情 (ClassDetail)

| 元素 | 内容 |
|---|---|
| 班级信息 | 名称、学生数、创建时间、描述 |
| 学生表格 | 昵称、状态(正常/禁用)、最后活跃、操作(移除) |
| 添加学生弹窗 | `请输入学生ID或手机号` |
| 移除确认 | `确定要将 "XXX" 移出班级吗？` |

---

### 3.10 知识库 (Knowledge)

**Tab切换**: `文档知识` / `问答知识`

**文档知识表格**: 标题 / 类型 / 上传时间 / 操作(预览/删除)

**问答知识表格**: 问题 / 答案 / 创建时间 / 操作(编辑/删除)

**上传弹窗**: 拖拽上传区域，提示: `支持 PDF、Word、TXT 等格式文件`

**问答编辑弹窗**: 问题(textarea) + 答案(textarea)

---

### 3.11 课程管理 (Courses)

**课程表格**: 课程标题 / 所属班级 / 状态(已发布success/草稿info) / 创建时间 / 操作(编辑/发布/删除)

**发布课程弹窗**:
| 字段 | 类型 |
|---|---|
| 课程标题 | Input |
| 所属班级 | Select (班级列表) |
| 课程内容 | Textarea (6行) |
| 附件 | Upload (点击上传) |

---

### 3.12 分身管理 (Personas)

| 元素 | 内容 |
|---|---|
| 标题 | `班级分身管理` |
| 创建按钮 | 禁用状态，tooltip: `分身随班级创建，请前往班级管理创建班级` |
| 空状态 | `暂无班级分身` / `请先创建班级，分身将随班级自动创建` |

**分身表格**:
| 列 | 说明 |
|---|---|
| 分身昵称 | nickname |
| 学校 | school |
| 绑定班级 | 标签(success) 或 `-` |
| 公开状态 | 已公开(success)/未公开(info) |
| 学生数 | student_count |
| 文档数 | document_count |
| 创建时间 | 格式化日期 |
| 操作 | 进入管理/班级详情/设为公开(私有) |

**API**: `GET /api/personas` / `PUT /api/personas/{id}/visibility` / `GET /api/personas/{id}/dashboard`

---

### 3.13 个人中心 (Profile)

| 元素 | 内容 |
|---|---|
| 用户信息 | 头像、昵称、角色(教师)、状态(正常) |
| 功能菜单 | 修改资料 / 意见反馈 / 关于我们 / 退出登录 |

---

## 四、H5学生端

### 4.1 路由结构

| 路径 | 名称 | meta.title | 认证 |
|---|---|---|---|
| `/login` | Login | - | 否 |
| `/` | Layout | - | 是 |
| `/chat` | Chat | `对话` | 是 |
| `/history` | History | `历史记录` | 是 |
| `/discover` | Discover | `发现` | 是 |
| `/profile` | Profile | `我的` | 是 |
| `/:pathMatch(.*)*` | NotFound | - | 否 |

默认重定向: `/` → `/chat`

---

### 4.2 布局 (Layout) - 底部Tabbar

| Tab | 图标 | 路径 |
|---|---|---|
| 对话 | chat-o | /chat |
| 历史 | clock-o | /history |
| 发现 | apps-o | /discover |
| 我的 | user-o | /profile |

---

### 4.3 登录页 (Login)

| 元素 | 内容 |
|---|---|
| 标题 | `数字分身` |
| 副标题 | `学生端` |
| 状态 | 加载中: `登录中...` (van-loading) |
| 主按钮 | `微信授权登录` |
| 角色校验 | 非学生 → `您不是学生，无法访问此页面` |
| 跳转 | 登录成功 → `/chat` |

---

### 4.4 AI对话页 (Chat)

| 元素 | 内容 |
|---|---|
| 导航栏 | `AI助手` |
| 消息列表 | 头像 + 气泡(AI白色/自己绿色#07c160) + 时间 |
| 初始消息 | `你好，我是你的AI助手，有什么可以帮助你的吗？` |
| 输入框 | `placeholder="输入消息..."` + Enter发送 |
| 发送按钮 | `发送` (van-button, primary, small) |
| 空输入提示 | Toast: `请输入消息` |

**API**: `GET /api/student/conversations/{id}/messages` / `POST /api/student/conversations/{id}/messages`

---

### 4.5 历史记录页 (History)

| 元素 | 内容 |
|---|---|
| 导航栏 | `历史记录` |
| 搜索框 | `placeholder="搜索对话"` |
| 对话列表 | 标题 + 时间，左侧chat-o图标 |
| 空状态 | `暂无历史记录` |
| 跳转 | 点击 → `/chat?id=xxx` |

---

### 4.6 发现页 (Discover)

| 元素 | 内容 |
|---|---|
| 导航栏 | `发现` |
| 功能入口 | `我的班级`(cluster) / `课程列表`(orders-o) / `学习资料`(description) |
| 推荐内容 | 卡片列表: 标题+描述+缩略图 + `查看详情` 按钮 |

---

### 4.7 个人中心 (Profile)

| 元素 | 内容 |
|---|---|
| 导航栏 | `我的` |
| 用户信息 | 头像 + 昵称 + ID |
| 功能菜单 | 个人资料(edit) / 意见反馈(chat-o) / 关于我们(info-o) / 退出登录(revoke) |
| 退出确认 | `确定要退出登录吗？` |

---

## 五、H5管理后台

### 5.1 路由结构

| 路径 | 名称 | meta.title | 认证 |
|---|---|---|---|
| `/login` | Login | - | 否 |
| `/` | Layout | - | 是 |
| `/dashboard` | Dashboard | `仪表盘` | 是 |
| `/users` | Users | `用户管理` | 是 |
| `/feedbacks` | Feedbacks | `反馈管理` | 是 |
| `/logs` | Logs | `操作日志` | 是 |
| `/:pathMatch(.*)*` | NotFound | - | 否 |

默认重定向: `/` → `/dashboard`

---

### 5.2 布局框架 (Layout)

| 元素 | 内容 |
|---|---|
| 侧边栏Logo | Logo图片 + `管理后台` |
| 导航菜单 | 仪表盘(DataAnalysis) / 用户管理(User) / 反馈管理(ChatDotSquare) / 操作日志(Document) |
| 顶部栏 | 折叠按钮 + 页面标题 + 用户下拉(退出登录) |
| 侧边栏样式 | 深色背景 #304156，激活色 #409eff |

---

### 5.3 登录页 (Login)

| 元素 | 内容 |
|---|---|
| 标题 | `数字分身管理后台` |
| 副标题 | `管理员登录` |
| 状态 | `正在登录中...` / 错误信息 / `重新登录` |
| 主按钮 | `微信授权登录` |
| 角色校验 | 非管理员 → `您不是管理员，无法访问此页面` |

---

### 5.4 仪表盘 (Dashboard) ⭐

#### 统计卡片（6张）

| 卡片 | 数据 | 附加 |
|---|---|---|
| 总用户数 | total_users | 今日新增: today_new_users |
| 教师数 | teacher_count | - |
| 学生数 | student_count | - |
| 总对话数 | total_conversations | - |
| 总消息数 | total_messages | 今日: today_messages |
| 今日活跃 | today_active_users | - |

#### 图表区域

| 图表 | 类型 | 说明 |
|---|---|---|
| 用户趋势 | 折线图 | 支持 7天/30天/90天 切换 |
| 角色分布 | 饼图 | 教师/学生/管理员 |
| 对话时段分布 | 柱状图 | 按小时统计 |
| 活跃用户排行 | 表格 | 昵称/角色/消息数/最后活跃 |

**API**: `GET /api/admin/dashboard/overview` / `GET /api/admin/dashboard/user-stats` / `GET /api/admin/dashboard/chat-stats` / `GET /api/admin/dashboard/active-users`

---

### 5.5 用户管理 (Users) ⭐

#### 搜索表单

| 字段 | 类型 | placeholder | 选项 |
|---|---|---|---|
| 昵称 | Input | `请输入昵称` | - |
| 角色 | Select | `请选择角色` | 教师/学生/管理员 |
| 状态 | Select | `请选择状态` | 正常(active)/禁用(disabled) |

#### 用户表格

| 列 | 标签类型 |
|---|---|
| ID | - |
| 昵称 | - |
| 角色 | 管理员(danger)/教师(primary)/学生(success) |
| 状态 | 正常(success)/禁用(danger) |
| 注册时间 | - |
| 操作 | 修改角色/禁用(启用) |

#### 修改角色弹窗 (400px)

| 字段 | 说明 |
|---|---|
| 用户昵称 | 只读 |
| 新角色 | Select: 教师/学生/管理员 |

#### 禁用/启用确认

`确定要{禁用/启用}用户 "{昵称}" 吗？`

**API**: `GET /api/admin/users` / `PUT /api/admin/users/{id}/role` / `PUT /api/admin/users/{id}/status`

---

### 5.6 反馈管理 (Feedbacks) ⭐

#### 搜索表单

| 字段 | 选项 |
|---|---|
| 状态 | 待处理(pending)/处理中(processing)/已解决(resolved) |

#### 反馈表格

| 列 | 标签类型 |
|---|---|
| ID | - |
| 用户昵称 | - |
| 用户角色 | 教师(primary)/学生(success) |
| 反馈内容 | 溢出隐藏 |
| 状态 | 待处理(warning)/处理中(primary)/已解决(success) |
| 提交时间 | - |
| 操作 | 查看 |

#### 反馈详情弹窗 (600px)

| 元素 | 内容 |
|---|---|
| 描述列表 | 用户昵称/角色/反馈内容/图片(可预览)/提交时间/状态/回复 |
| 更新状态 | Select: 待处理/处理中/已解决 |
| 回复内容 | Textarea: `请输入回复内容` |
| 按钮 | 取消/保存 |

**API**: `GET /api/admin/feedbacks` / `PUT /api/admin/feedbacks/{id}`

---

### 5.7 操作日志 (Logs) ⭐

#### 搜索表单

| 字段 | 类型 | placeholder | 选项 |
|---|---|---|---|
| 用户ID | Input | `请输入用户ID` | - |
| 操作类型 | Select | `请选择操作类型` | 用户登录/发送消息/创建班级/上传知识/创建分身 |
| 平台 | Select | `请选择平台` | 小程序(miniapp)/H5(h5)/API(api) |
| 时间范围 | DateRange | `开始日期` 至 `结束日期` | - |

#### 日志表格

| 列 | 说明 |
|---|---|
| ID | - |
| 用户昵称 | - |
| 角色 | 教师(primary)/学生(success) |
| 操作类型 | action |
| 资源类型 | resource |
| 详情 | 溢出隐藏 |
| 平台 | 小程序/H5/API 标签 |
| 状态码 | status_code |
| 耗时(ms) | duration_ms |
| 时间 | 格式: YYYY-MM-DD HH:mm:ss |

**操作按钮**: 搜索/重置/导出CSV

**API**: `GET /api/admin/logs` / `GET /api/admin/logs/export`

---

## 六、公共组件

### 6.1 小程序端组件 (12个)

| 组件 | 说明 | 使用场景 |
|---|---|---|
| AvatarPopup | 头像弹窗，查看教师详情 | 对话页点击头像 |
| ChatBubble | 聊天气泡组件 | 对话页消息展示 |
| CustomTabBar | 自定义底部TabBar | 全局底部导航 |
| EmojiPanel | Emoji表情面板 | 对话页表情选择 |
| Empty | 空状态组件 | 列表为空时展示 |
| PlusPanel | 附件面板(拍照/相册/文件) | 对话页附件发送 |
| StudentHome | 学生首页组件 | 首页学生视图 |
| TagInput | 标签输入组件 | 教材配置等场景 |
| TeacherCard | 教师信息卡片 | 发现页推荐老师 |
| TeacherDashboard | 教师工作台组件 | 首页教师视图 |
| ThinkingPanel | AI思考步骤面板 | 对话页AI推理展示 |
| VoiceInput | 语音输入组件 | 对话页语音输入 |

---

## 七、API接口汇总

### 7.1 认证模块

| 方法 | 路径 | 说明 | 端 |
|---|---|---|---|
| POST | `/api/auth/wx-login` | 微信小程序登录 | 小程序 |
| GET | `/api/auth/wx-h5-login-url` | 获取H5微信授权URL | H5全端 |
| POST | `/api/auth/wx-h5-callback` | H5微信登录回调 | H5全端 |
| POST | `/api/auth/complete-profile` | 补全用户信息 | 全端 |

### 7.2 分身模块

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/personas` | 获取分身列表 |
| POST | `/api/personas` | 创建分身 |
| GET | `/api/personas/{id}` | 获取分身详情 |
| PUT | `/api/personas/{id}` | 更新分身信息 |
| PUT | `/api/personas/{id}/visibility` | 设置公开/私有 |
| GET | `/api/personas/{id}/dashboard` | 分身仪表盘数据 |
| GET | `/api/personas/search-students` | 搜索学生 |

### 7.3 对话模块

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/chat/send` | 发送消息(SSE) |
| GET | `/api/chat/messages` | 获取历史消息 |
| POST | `/api/chat/teacher-reply` | 教师直接回复 |
| POST | `/api/chat/takeover` | 教师接管对话 |
| GET | `/api/chat-list/student` | 学生端聊天列表 |
| GET | `/api/chat-list/teacher` | 教师端聊天列表 |
| POST | `/api/chat-pins` | 置顶聊天 |
| DELETE | `/api/chat-pins/{type}/{id}` | 取消置顶 |

### 7.4 会话模块

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/sessions` | 获取会话列表 |
| GET | `/api/student/conversations` | 学生对话列表(H5) |
| POST | `/api/student/conversations` | 创建新对话(H5) |
| GET | `/api/student/conversations/{id}/messages` | 获取消息(H5) |
| POST | `/api/student/conversations/{id}/messages` | 发送消息(H5) |

### 7.5 班级模块

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/classes` | 班级列表 |
| POST | `/api/classes` | 创建班级(V11含分身) |
| PUT | `/api/classes/{id}` | 更新班级 |
| DELETE | `/api/classes/{id}` | 删除班级 |
| GET | `/api/teacher/classes` | 教师班级列表(H5) |
| GET | `/api/teacher/classes/{id}` | 班级详情(H5) |
| POST | `/api/teacher/classes/{id}/students` | 添加学生到班级(H5) |
| DELETE | `/api/teacher/classes/{id}/students/{sid}` | 移除学生(H5) |

### 7.6 知识库模块

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/knowledge` | 知识库列表 |
| POST | `/api/knowledge` | 添加知识 |
| DELETE | `/api/knowledge/{id}` | 删除知识 |
| GET | `/api/teacher/knowledge` | 教师知识列表(H5) |
| POST | `/api/teacher/knowledge/upload` | 上传文档(H5) |
| POST | `/api/teacher/knowledge/qa` | 创建问答(H5) |
| PUT | `/api/teacher/knowledge/qa/{id}` | 更新问答(H5) |

### 7.7 课程模块

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/teacher/courses` | 课程列表 |
| POST | `/api/teacher/courses` | 创建课程 |
| PUT | `/api/teacher/courses/{id}` | 更新课程 |
| POST | `/api/teacher/courses/{id}/publish` | 发布课程 |
| DELETE | `/api/teacher/courses/{id}` | 删除课程 |
| GET | `/api/student/courses` | 学生课程列表 |
| GET | `/api/student/courses/{id}` | 课程详情 |
| GET | `/api/student/materials` | 学习资料 |
| GET | `/api/student/my-class` | 我的班级 |

### 7.8 师生关系模块

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/relations` | 关系列表 |
| POST | `/api/relations/invite` | 邀请学生 |
| POST | `/api/relations/apply` | 申请加入 |
| PUT | `/api/relations/{id}/approve` | 审批通过 |
| PUT | `/api/relations/{id}/reject` | 审批拒绝 |
| PUT | `/api/relations/{id}/toggle` | 启停关系 |

### 7.9 用户模块

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/users/me` | 当前用户信息 |
| PUT | `/api/users/student-profile` | 更新学生资料 |

### 7.10 管理后台模块

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/admin/dashboard/overview` | 系统总览 |
| GET | `/api/admin/dashboard/user-stats` | 用户统计 |
| GET | `/api/admin/dashboard/chat-stats` | 对话统计 |
| GET | `/api/admin/dashboard/active-users` | 活跃用户 |
| GET | `/api/admin/users` | 用户列表 |
| PUT | `/api/admin/users/{id}/role` | 修改角色 |
| PUT | `/api/admin/users/{id}/status` | 启禁用户 |
| GET | `/api/admin/feedbacks` | 反馈列表 |
| PUT | `/api/admin/feedbacks/{id}` | 更新反馈 |
| GET | `/api/admin/logs` | 日志列表 |
| GET | `/api/admin/logs/export` | 导出日志CSV |

### 7.11 其他模块

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/discover/recommend` | 发现页推荐 |
| GET | `/api/discover/search` | 发现页搜索 |
| POST | `/api/feedbacks` | 提交反馈 |
| GET | `/api/test-student` | 测试学生信息 |
| POST | `/api/test-student/reset` | 重置测试数据 |
| GET | `/api/share/{code}` | 获取分享信息 |
| POST | `/api/share/join` | 加入分享 |
| GET | `/api/platform/config` | 平台配置 |

---

## 附录：状态标签映射

### 角色标签

| 角色 | 文本 | Element类型 | 颜色 |
|---|---|---|---|
| teacher | 教师 | primary | 蓝色 |
| student | 学生 | success | 绿色 |
| admin | 管理员 | danger | 红色 |

### 反馈状态

| 状态 | 文本 | 类型 |
|---|---|---|
| pending | 待处理 | warning |
| processing | 处理中 | primary |
| resolved | 已解决 | success |

### 班级公开状态

| 状态 | 文本 | 类型 |
|---|---|---|
| true | 公开 | success |
| false | 私密 | info |

### 用户状态

| 状态 | 文本 | 类型 |
|---|---|---|
| active | 正常 | success |
| disabled | 禁用 | danger |

### 课程状态

| 状态 | 文本 | 类型 |
|---|---|---|
| published | 已发布 | success |
| draft | 草稿 | info |

### 记忆层级

| 层级 | 说明 |
|---|---|
| 核心 | 长期重要记忆 |
| 情景 | 具体场景记忆 |
| 归档 | 已归档的历史记忆 |

---

*文档基于项目源代码全量扫描自动生成，覆盖 4 个前端应用、61+ 页面/组件、80+ API接口。*
