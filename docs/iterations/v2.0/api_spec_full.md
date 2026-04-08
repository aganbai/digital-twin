# V2.0 接口完整清单

> **最后更新**: 2026-04-05 | **文档版本**: v1.1.0
> **来源**: `src/backend/api/router.go` + `src/knowledge-service/app/main.py`
> **总计**: **136 个接口**（Go 后端 132 个 + Python 服务 4 个）
> **关联文档**: [V2.0 需求规格说明书](./requirements.md)

---

## 迭代11接口变更速览

| 接口 | 方法 | 变更类型 | 说明 |
|------|------|---------|------|
| `/api/personas/:id/switch` | PUT | **删除** | 无主分身，不需要切换 |
| `/api/personas/:id/activate` | PUT | **删除** | 分身随班级管理 |
| `/api/personas/:id/deactivate` | PUT | **删除** | 分身随班级管理 |
| `/api/personas` | POST | **重构** | 教师角色禁止独立创建，返回引导信息 |
| `/api/personas` | GET | **增强** | 返回 bound_class_id、bound_class_name、is_public |
| `/api/classes` | POST | **重构** | 同步创建班级专属分身，新增分身信息参数和 is_public |
| `/api/auth/complete-profile` | POST | **改造** | 教师注册不再创建分身，改为创建自测学生 |
| `/api/test-student` | GET | **新增** | 获取自测学生信息 |
| `/api/test-student/reset` | POST | **新增** | 重置自测学生数据 |

---

## 1. 认证接口（无需鉴权）

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 1 | POST | `/api/auth/register` | 用户注册 | 所有 | V1.0 |
| 2 | POST | `/api/auth/login` | 用户登录 | 所有 | V1.0 |
| 3 | POST | `/api/auth/wx-login` | 微信小程序登录 | 所有 | V1.0 |
| 4 | POST | `/api/auth/refresh` | 刷新JWT令牌 | 所有 | V1.0 |
| 5 | GET | `/api/auth/wx-h5-login-url` | 获取H5微信授权URL | 所有 | 迭代10 |
| 6 | POST | `/api/auth/wx-h5-callback` | H5微信授权回调 | 所有 | 迭代10 |
| 7 | POST | `/api/auth/complete-profile` | 补全用户信息（需鉴权）；迭代11改造：教师注册自动创建自测学生 | 所有 | V1.0→**迭代11改造** |

## 2. 用户接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 8 | GET | `/api/user/profile` | 获取当前用户信息 | 所有 | V1.0 |
| 9 | PUT | `/api/user/student-profile` | 更新学生基础信息 | student | 迭代8 |
| 10 | GET | `/api/teachers` | 获取教师列表 | 所有 | V1.0 |

## 3. 分身管理接口（迭代11重构）

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 11 | POST | `/api/personas` | 创建分身（**教师角色返回引导错误 40040**；学生角色正常） | 所有 | 迭代2→**迭代11重构** |
| 12 | GET | `/api/personas` | 获取分身列表（**新增 bound_class_id/bound_class_name/is_public**） | 所有 | 迭代2→**迭代11增强** |
| 13 | PUT | `/api/personas/:id` | 编辑分身信息 | 所有 | 迭代2 |
| ~~14~~ | ~~PUT~~ | ~~`/api/personas/:id/activate`~~ | ~~启用分身~~ **（迭代11删除）** | — | — |
| ~~15~~ | ~~PUT~~ | ~~`/api/personas/:id/deactivate`~~ | ~~停用分身~~ **（迭代11删除）** | — | — |
| ~~16~~ | ~~PUT~~ | ~~`/api/personas/:id/switch`~~ | ~~切换当前分身~~ **（迭代11删除）** | — | — |
| 14 | GET | `/api/personas/:id/dashboard` | 分身Dashboard数据 | 所有 | 迭代3 |
| 15 | GET | `/api/personas/marketplace` | 分身广场（仅展示 is_public=true 的班级） | 所有 | 迭代4→迭代11适配 |
| 16 | PUT | `/api/personas/:id/visibility` | 设置分身公开/私密 | 所有 | 迭代4 |

> **说明**：迭代11删除了 activate/deactivate/switch 三个接口，请求这些路径将返回 `404 Not Found`。

## 4. 班级管理接口（迭代11重构）

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 17 | POST | `/api/classes` | 创建班级（**同步创建班级专属分身，新增 persona_nickname/persona_school/persona_description/is_public**） | teacher | 迭代2→**迭代11重构** |
| 18 | GET | `/api/classes` | 获取班级列表 | teacher | 迭代2 |
| 19 | PUT | `/api/classes/:id` | 更新班级（**新增 is_public 字段支持**） | teacher | 迭代2→**迭代11增强** |
| 20 | DELETE | `/api/classes/:id` | 删除班级（关联分身标记为停用） | teacher | 迭代2 |
| 21 | GET | `/api/classes/:id/members` | 获取班级成员 | teacher | 迭代2 |
| 22 | POST | `/api/classes/:id/members` | 添加班级成员 | teacher | 迭代2 |
| 23 | DELETE | `/api/classes/:id/members/:member_id` | 移除班级成员 | teacher | 迭代2 |
| 24 | PUT | `/api/classes/:id/toggle` | 启停班级 | teacher | 迭代3 |
| 25 | POST | `/api/classes/v8` | 创建班级（V8增强版） | teacher | 迭代8 |
| 26 | GET | `/api/classes/:id/share-info` | 获取班级分享信息 | teacher | 迭代8 |
| 27 | GET | `/api/classes/:id/members/v8` | 获取班级成员（V8增强版） | teacher | 迭代8 |
| 28 | GET | `/api/classes/:id` | 学生查看班级详情 | 所有 | 迭代9 |

## 5. 班级加入申请接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 29 | POST | `/api/classes/join` | 学生申请加入班级 | student | 迭代8 |
| 30 | GET | `/api/join-requests/pending` | 获取待审批申请列表 | teacher | 迭代8 |
| 31 | PUT | `/api/join-requests/:id/approve` | 审批通过 | teacher | 迭代8 |
| 32 | PUT | `/api/join-requests/:id/reject` | 审批拒绝 | teacher | 迭代8 |

## 6. 师生关系接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 33 | POST | `/api/relations/invite` | 教师邀请学生 | teacher | 迭代1 |
| 34 | POST | `/api/relations/apply` | 学生申请使用分身 | student | 迭代1 |
| 35 | PUT | `/api/relations/:id/approve` | 教师审批同意 | teacher | 迭代1 |
| 36 | PUT | `/api/relations/:id/reject` | 教师审批拒绝 | teacher | 迭代1 |
| 37 | GET | `/api/relations` | 获取师生关系列表 | 所有 | 迭代1 |
| 38 | PUT | `/api/relations/:id/toggle` | 启停师生关系 | teacher | 迭代3 |

## 7. 对话接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 39 | POST | `/api/chat` | 发送消息（非流式） | 所有 | V1.0 |
| 40 | POST | `/api/chat/stream` | SSE流式对话 | 所有 | 迭代1 |
| 41 | GET | `/api/conversations` | 获取对话历史 | 所有 | V1.0 |
| 42 | GET | `/api/conversations/sessions` | 获取会话列表 | 所有 | V1.0 |
| 43 | POST | `/api/chat/new-session` | 创建新会话 | 所有 | 迭代8 |
| 44 | GET | `/api/chat/quick-actions` | 获取快捷指令 | 所有 | 迭代8 |
| 45 | POST | `/api/conversations/sessions/:session_id/title` | 生成会话标题 | 所有 | 迭代9 |

> **向量召回优化（迭代11）**：`/api/chat` 和 `/api/chat/stream` 的知识库检索环节已优化为：召回100条 → 置信度过滤(score≥0.3) → 最多20条 → scope过滤 → 返回≤5条。接口格式不变。

## 8. 教师介入对话接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 46 | POST | `/api/chat/teacher-reply` | 教师真人回复 | teacher | 迭代4 |
| 47 | GET | `/api/chat/takeover-status` | 获取接管状态 | 所有 | 迭代4 |
| 48 | POST | `/api/chat/end-takeover` | 结束接管 | teacher | 迭代4 |
| 49 | GET | `/api/conversations/student/:student_persona_id` | 查看学生对话记录 | teacher | 迭代4 |

## 9. 聊天列表接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 50 | GET | `/api/chat-list/teacher` | 教师端聊天列表（按班级组织） | teacher | 迭代8 |
| 51 | GET | `/api/chat-list/student` | 学生端聊天列表（按老师分组） | student | 迭代8 |
| 52 | POST | `/api/chat-pins` | 置顶聊天 | 所有 | 迭代8 |
| 53 | DELETE | `/api/chat-pins/:type/:id` | 取消置顶 | 所有 | 迭代8 |
| 54 | GET | `/api/chat-pins` | 获取置顶列表 | 所有 | 迭代8 |

## 10. 知识库接口（旧版 documents）

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 55 | POST | `/api/documents` | 添加文档 | teacher/admin | V1.0 |
| 56 | GET | `/api/documents` | 获取文档列表 | teacher/admin | V1.0 |
| 57 | DELETE | `/api/documents/:id` | 删除文档 | teacher/admin | V1.0 |
| 58 | POST | `/api/documents/upload` | 文件上传 | teacher/admin | 迭代1 |
| 59 | POST | `/api/documents/import-url` | URL导入 | teacher/admin | 迭代1 |
| 60 | POST | `/api/documents/preview` | 预览文档 | teacher/admin | 迭代3 |
| 61 | POST | `/api/documents/preview-upload` | 预览上传文件 | teacher/admin | 迭代3 |
| 62 | POST | `/api/documents/preview-url` | 预览URL内容 | teacher/admin | 迭代3 |
| 63 | POST | `/api/documents/confirm` | 确认添加文档 | teacher/admin | 迭代3 |
| 64 | POST | `/api/documents/import-chat` | 聊天记录导入 | teacher/admin | 迭代6 |
| 65 | POST | `/api/documents/batch-upload` | 批量文件上传 | teacher/admin | 迭代7 |
| 66 | GET | `/api/batch-tasks/:task_id` | 查询批量任务状态 | teacher/admin | 迭代7 |

## 11. 知识库接口（V8增强版 knowledge）

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 67 | POST | `/api/knowledge/upload` | 智能知识库上传（统一输入框） | teacher/admin | 迭代8 |
| 68 | GET | `/api/knowledge` | 搜索知识库列表 | teacher/admin | 迭代8 |
| 69 | GET | `/api/knowledge/:id` | 获取知识详情 | teacher/admin | 迭代8 |
| 70 | PUT | `/api/knowledge/:id` | 更新知识 | teacher/admin | 迭代8 |
| 71 | DELETE | `/api/knowledge/:id` | 删除知识 | teacher/admin | 迭代8 |

## 12. 记忆接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 72 | GET | `/api/memories` | 获取记忆列表 | 所有 | V1.0 |
| 73 | PUT | `/api/memories/:id` | 更新记忆 | teacher | 迭代6 |
| 74 | DELETE | `/api/memories/:id` | 删除记忆 | teacher | 迭代6 |
| 75 | POST | `/api/memories/summarize` | 手动触发记忆合并 | teacher | 迭代6 |

## 13. 评语与问答风格接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 76 | POST | `/api/comments` | 教师写评语/备注 | teacher | 迭代1 |
| 77 | GET | `/api/comments` | 获取评语列表 | 所有 | 迭代1 |
| 78 | PUT | `/api/students/:id/dialogue-style` | 设置学生问答风格 | teacher | 迭代1 |
| 79 | GET | `/api/students/:id/dialogue-style` | 获取学生问答风格 | 所有 | 迭代1 |

## 14. 对话风格接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 80 | PUT | `/api/styles` | 设置对话风格配置 | teacher | 迭代6 |
| 81 | GET | `/api/styles` | 获取对话风格配置 | 所有 | 迭代6 |

## 15. 学生管理接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 82 | GET | `/api/students/search` | 搜索学生（自测账号不出现） | teacher | 迭代1→迭代11 |
| 83 | GET | `/api/students/:id/profile` | 查看学生详情 | teacher | 迭代9 |
| 84 | PUT | `/api/students/:id/evaluation` | 更新学生评语 | teacher | 迭代9 |
| 85 | POST | `/api/students/parse-text` | LLM解析学生文本 | teacher | 迭代7 |
| 86 | POST | `/api/students/batch-create` | 批量创建学生 | teacher | 迭代7 |

## 16. 教材配置接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 87 | POST | `/api/curriculum-configs` | 创建教材配置 | teacher | 迭代7 |
| 88 | GET | `/api/curriculum-configs` | 获取教材配置列表 | teacher | 迭代7 |
| 89 | PUT | `/api/curriculum-configs/:id` | 更新教材配置 | teacher | 迭代7 |
| 90 | DELETE | `/api/curriculum-configs/:id` | 删除教材配置 | teacher | 迭代7 |
| 91 | GET | `/api/curriculum-versions` | 获取教材版本列表 | 所有 | 迭代7 |

## 17. 课程接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 92 | POST | `/api/courses` | 发布课程 | teacher | 迭代9 |
| 93 | GET | `/api/courses` | 课程列表 | teacher | 迭代9 |
| 94 | PUT | `/api/courses/:id` | 更新课程 | teacher | 迭代9 |
| 95 | DELETE | `/api/courses/:id` | 删除课程 | teacher | 迭代9 |
| 96 | POST | `/api/courses/:id/push` | 推送课程给学生 | teacher | 迭代9 |

## 18. 教师消息推送接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 97 | POST | `/api/teacher-messages` | 教师推送消息 | teacher | 迭代7 |
| 98 | GET | `/api/teacher-messages/history` | 推送历史 | teacher | 迭代7 |

## 19. 分享码接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 99 | POST | `/api/shares` | 创建分享码 | teacher | 迭代2 |
| 100 | GET | `/api/shares` | 获取分享码列表 | teacher | 迭代2 |
| 101 | POST | `/api/shares/:code/join` | 通过分享码加入 | 所有 | 迭代2 |
| 102 | PUT | `/api/shares/:id/deactivate` | 废止分享码 | teacher | 迭代2 |
| 103 | GET | `/api/shares/:code/info` | 获取分享码信息（可选鉴权） | 所有 | 迭代2 |

## 20. 反馈接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 104 | POST | `/api/feedbacks` | 提交反馈 | 所有 | 迭代7 |
| 105 | GET | `/api/feedbacks` | 反馈列表 | teacher/admin | 迭代7 |
| 106 | PUT | `/api/feedbacks/:id/status` | 更新反馈状态 | teacher/admin | 迭代7 |

## 21. 发现页接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 107 | GET | `/api/discover` | 发现页数据（迭代11：仅展示 is_public=true 的班级） | 所有 | 迭代8→迭代11适配 |
| 108 | GET | `/api/discover/detail` | 发现详情 | 所有 | 迭代8 |
| 109 | GET | `/api/discover/search` | 发现页搜索 | 所有 | 迭代8 |

## 22. 通用文件上传接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 110 | POST | `/api/upload` | 通用文件上传（小程序） | 所有 | 迭代5 |
| 111 | POST | `/api/upload/h5` | H5文件上传 | 所有 | 迭代10 |

## 23. 平台配置接口（无需鉴权）

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 112 | GET | `/api/platform/config` | 平台配置（小程序/H5差异化配置） | 所有 | 迭代10 |

## 24. 自测学生接口（迭代11新增）

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 113 | GET | `/api/test-student` | 获取教师自测学生信息（user_id/username/persona_id/joined_classes等） | teacher | 迭代11 |
| 114 | POST | `/api/test-student/reset` | 重置自测学生（清空与该教师所有班级分身的对话记录和记忆） | teacher | 迭代11 |

## 25. 管理员接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 115 | GET | `/api/admin/dashboard/overview` | 管理员仪表盘总览 | admin | 迭代10 |
| 116 | GET | `/api/admin/dashboard/user-stats` | 用户统计（趋势图可选7/30/90天） | admin | 迭代10 |
| 117 | GET | `/api/admin/dashboard/chat-stats` | 对话统计 | admin | 迭代10 |
| 118 | GET | `/api/admin/dashboard/knowledge-stats` | 知识库统计 | admin | 迭代10 |
| 119 | GET | `/api/admin/dashboard/active-users` | 活跃用户排行 | admin | 迭代10 |
| 120 | GET | `/api/admin/users` | 用户列表 | admin | 迭代10 |
| 121 | PUT | `/api/admin/users/:id/role` | 修改用户角色 | admin | 迭代10 |
| 122 | PUT | `/api/admin/users/:id/status` | 启禁用户（立即失效token） | admin | 迭代10 |
| 123 | GET | `/api/admin/feedbacks` | 反馈管理列表 | admin | 迭代10 |
| 124 | GET | `/api/admin/logs` | 操作日志查询（多条件筛选） | admin | 迭代10 |
| 125 | GET | `/api/admin/logs/stats` | 日志统计（操作频次/平台/时段热力图） | admin | 迭代10 |
| 126 | GET | `/api/admin/logs/export` | 日志导出CSV | admin | 迭代10 |

## 26. 系统接口

| # | 方法 | 路径 | 说明 | 角色 | 来源 |
|---|------|------|------|------|------|
| 127 | GET | `/api/system/health` | 健康检查（含uptime/database状态） | 所有 | V1.0 |
| 128 | GET | `/api/system/plugins` | 插件列表 | admin | V1.0 |
| 129 | GET | `/api/system/pipelines` | 管道列表 | admin | V1.0 |

## 27. Python LlamaIndex 服务接口（内部，端口 8100）

| # | 方法 | 路径 | 说明 | 来源 |
|---|------|------|------|------|
| 130 | POST | `/api/v1/vectors/documents` | 存储文档向量（接收已分好的 chunks） | 迭代5 |
| 131 | POST | `/api/v1/vectors/search` | 语义检索（返回 top-k 相似文档块） | 迭代5 |
| 132 | DELETE | `/api/v1/vectors/documents/{doc_id}` | 删除指定文档的所有向量 | 迭代5 |
| 133 | GET | `/api/v1/health` | 健康检查 | 迭代5 |

---

## 接口统计

### 按来源统计

| 来源 | Go 接口数 | 说明 |
|------|-----------|------|
| V1.0 | 12 | |
| 迭代1 | 12 | |
| 迭代2 | 14 | |
| 迭代3 | 7 | |
| 迭代4 | 6 | |
| 迭代5 | 5 | 含 Python 服务 4 个 |
| 迭代6 | 7 | |
| 迭代7 | 14 | |
| 迭代8 | 22 | |
| 迭代9 | 11 | |
| 迭代10 | 20 | |
| 迭代11 | +2新增 / -3删除 / 4改造 | 净增 -1，实际 Go 后端 129→129 个（含改造） |
| **Go 后端合计** | **132** | |
| Python 服务 | 4 | |
| **总计** | **136** | 活跃接口（不含已删除的3个） |

### 按角色统计

| 角色 | 数量 |
|------|------|
| 所有角色可访问 | 48 |
| teacher 专属 | 53 |
| student 专属 | 3 |
| teacher/admin | 17 |
| admin 专属 | 14 |

### 错误码（迭代11新增）

| 错误码 | HTTP | 说明 |
|--------|------|------|
| 40040 | 400 | 教师分身随班级创建，请通过创建班级来创建分身 |
| 40041 | 404 | 自测学生账号不存在 |
| 40042 | 400 | 自测学生数据重置失败 |

---

**文档版本**: v1.1.0
**创建日期**: 2026-04-05
**维护说明**: 本文档为 V2.0 全量接口清单。各迭代接口变更详情见 `iteration{N}_api_spec.md`，V2.0 总体需求见 `requirements.md`
