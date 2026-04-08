# V2.0 IT12 冒烟测试报告

## 执行概要

| 指标 | 数值 |
|------|------|
| 总用例数 | 30 |
| 通过数 | 9 |
| 失败数 | 21 |
| 通过率 | 30.0% |
| 执行时间 | 2026-04-08 20:28:57 ~ 20:28:58 |

### P0优先级统计

| 指标 | 数值 |
|------|------|
| P0总数 | 21 |
| P0通过 | 4 |
| P0失败 | 17 |
| P0通过率 | 19.0% |

## 环境信息

| 项目 | 状态 |
|------|------|
| 后端服务 | http://localhost:8080 - ✅ 运行中 |
| 知识库服务 | http://localhost:8100 - ✅ 运行中 |
| 测试框架 | Python Requests |
| 测试环境 | 本地沙箱 |

## 用例执行详情

### SMOKE-030 - 系统健康检查与反馈提交 ❌

- **优先级**: P0
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:57.407220
- **结束时间**: 2026-04-08T20:28:57.415359
- **失败步骤**: Step 3: POST /api/feedbacks 提交一条反馈
- **错误信息**: HTTP 401

**步骤执行结果**:

- Step 1: ✅ GET /api/system/health 验证后端健康...
- Step 2: ✅ GET http://localhost:8100/api/v1/health 验证Python服务...
- Step 3: ❌ POST /api/feedbacks 提交一条反馈...
  - 错误: HTTP 401

---

### SMOKE-001 - 微信登录与教师注册（含自测学生创建） ❌

- **优先级**: P0
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:57.415501
- **结束时间**: 2026-04-08T20:28:57.508514
- **失败步骤**: Step 1: POST /api/auth/wx-login { code: 'smoke_teacher_001' }
- **错误信息**: HTTP 401: Unknown

**步骤执行结果**:

- Step 1: ❌ POST /api/auth/wx-login { code: 'smoke_teacher_001...
  - 错误: HTTP 401: Unknown

---

### SMOKE-015 - Python LlamaIndex语义检索服务 ❌

- **优先级**: P0
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:57.508562
- **结束时间**: 2026-04-08T20:28:57.512434
- **失败步骤**: Step 2: POST /api/v1/vectors/documents 存储文档向量
- **错误信息**: HTTP 401

**步骤执行结果**:

- Step 1: ✅ GET http://localhost:8100/api/v1/health 健康检查...
- Step 2: ❌ POST /api/v1/vectors/documents 存储文档向量...
  - 错误: HTTP 401

---

### SMOKE-023 - 管理员仪表盘全套统计 ❌

- **优先级**: P0
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:57.512490
- **结束时间**: 2026-04-08T20:28:58.103678
- **失败步骤**: Step 4: GET /api/admin/dashboard/knowledge-stats 知识库统计
- **错误信息**: HTTP 403

**步骤执行结果**:

- Step 1: ✅ GET /api/admin/dashboard/overview 系统总览...
- Step 2: ✅ GET /api/admin/dashboard/user-stats 用户统计...
- Step 3: ✅ GET /api/admin/dashboard/chat-stats 对话统计...
- Step 4: ❌ GET /api/admin/dashboard/knowledge-stats 知识库统计...
  - 错误: HTTP 403

---

### SMOKE-026 - H5微信OAuth登录 + 平台配置 + 文件上传 ❌

- **优先级**: P0
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:58.103942
- **结束时间**: 2026-04-08T20:28:58.175335
- **失败步骤**: Step 1: GET /api/auth/wx-h5-login-url 获取H5授权URL
- **错误信息**: HTTP 400

**步骤执行结果**:

- Step 1: ❌ GET /api/auth/wx-h5-login-url 获取H5授权URL...
  - 错误: HTTP 400

---

### SMOKE-027 - 鉴权与角色隔离 ❌

- **优先级**: P0
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:58.175378
- **结束时间**: 2026-04-08T20:28:58.224733
- **失败步骤**: Step 2: POST /api/classes（学生token）验证返回 403
- **错误信息**: HTTP 403

**步骤执行结果**:

- Step 1: ✅ GET /api/personas（无Authorization）验证返回 401...
- Step 2: ❌ POST /api/classes（学生token）验证返回 403...
  - 错误: HTTP 403

---

### SMOKE-002 - 学生注册与登录 ✅

- **优先级**: P0
- **状态**: passed
- **开始时间**: 2026-04-08T20:28:58.224791
- **结束时间**: 2026-04-08T20:28:58.316910

**步骤执行结果**:

- Step 1: ✅ POST /api/auth/wx-login { code: 'smoke_student_001...
- Step 2: ✅ POST /api/auth/complete-profile { role: 'student',...
- Step 3: ✅ 验证返回 persona_id + 新 token...
- Step 4: ✅ POST /api/auth/refresh 验证Token刷新...

---

### SMOKE-003 - 创建班级（同步创建分身）+ 分身列表验证 ❌

- **优先级**: P0
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:58.316935
- **结束时间**: 2026-04-08T20:28:58.363325
- **失败步骤**: Step 1: POST /api/classes { name, persona_nickname, persona_school, persona_description, is_public: true }
- **错误信息**: HTTP 403

**步骤执行结果**:

- Step 1: ❌ POST /api/classes { name, persona_nickname, person...
  - 错误: HTTP 403

---

### SMOKE-004 - 分享码创建 + 学生扫码加入 + 审批通过 ❌

- **优先级**: P0
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:58.363373
- **结束时间**: 2026-04-08T20:28:58.409674
- **失败步骤**: Step 1: POST /api/shares { persona_id } 创建分享码
- **错误信息**: HTTP 403

**步骤执行结果**:

- Step 1: ❌ POST /api/shares { persona_id } 创建分享码...
  - 错误: HTTP 403

---

### SMOKE-005 - 教师禁止独立创建分身 + 废弃接口404验证 ❌

- **优先级**: P0
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:58.409717
- **结束时间**: 2026-04-08T20:28:58.457776
- **失败步骤**: Step 2: PUT /api/personas/:id/switch 验证返回 404
- **错误信息**: 期望404, 实际-1

**步骤执行结果**:

- Step 1: ✅ POST /api/personas { role: 'teacher' } 验证返回 400, c...
- Step 2: ❌ PUT /api/personas/:id/switch 验证返回 404...
  - 错误: 期望404, 实际-1

---

### SMOKE-006 - 普通消息发送与AI回复 ❌

- **优先级**: P0
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:58.457840
- **结束时间**: 2026-04-08T20:28:58.503802
- **失败步骤**: Step 1: POST /api/chat { teacher_persona_id, message: '你好' }
- **错误信息**: HTTP 403

**步骤执行结果**:

- Step 1: ❌ POST /api/chat { teacher_persona_id, message: '你好'...
  - 错误: HTTP 403

---

### SMOKE-007 - SSE流式对话 + 思考过程展示 ✅

- **优先级**: P0
- **状态**: passed
- **开始时间**: 2026-04-08T20:28:58.503848
- **结束时间**: 2026-04-08T20:28:58.505075

**步骤执行结果**:

- Step 1: ✅ POST /api/chat/stream { session_id, message, teach...
- Step 2: ✅ 监听 SSE 事件流...
- Step 3: ✅ 验证收到 thinking_step 事件（RAG/记忆/LLM）...
- Step 4: ✅ 验证收到流式回复内容...
- Step 5: ✅ 验证收到 [DONE] 事件...

---

### SMOKE-008 - 流式中断功能（迭代12新增） ❌

- **优先级**: P0
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:58.505093
- **结束时间**: 2026-04-08T20:28:58.506137
- **失败步骤**: Step 3: 验证返回 200, {code: 0, data: {aborted: true}}
- **错误信息**: HTTP -1

**步骤执行结果**:

- Step 1: ✅ 发送消息触发流式回复...
- Step 2: ✅ 流式进行中 GET /api/chat/stream/:session_id/abort...
- Step 3: ❌ 验证返回 200, {code: 0, data: {aborted: true}}...
  - 错误: HTTP -1

---

### SMOKE-009 - 中断接口异常场景 ❌

- **优先级**: P0
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:58.506176
- **结束时间**: 2026-04-08T20:28:58.506947
- **失败步骤**: Step 1: GET /api/chat/stream/non-existent/abort 验证返回 404
- **错误信息**: 期望404, 实际-1

**步骤执行结果**:

- Step 1: ❌ GET /api/chat/stream/non-existent/abort 验证返回 404...
  - 错误: 期望404, 实际-1

---

### SMOKE-010 - 教师真人介入对话 ❌

- **优先级**: P0
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:58.506979
- **结束时间**: 2026-04-08T20:28:58.553451
- **失败步骤**: Step 1: POST /api/chat/teacher-reply { student_persona_id, message }
- **错误信息**: HTTP 403

**步骤执行结果**:

- Step 1: ❌ POST /api/chat/teacher-reply { student_persona_id,...
  - 错误: HTTP 403

---

### SMOKE-011 - 新会话创建 + 会话列表 + 标题生成 ❌

- **优先级**: P0
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:58.553500
- **结束时间**: 2026-04-08T20:28:58.554371
- **失败步骤**: Step 1: POST /api/chat/new-session 创建新会话
- **错误信息**: HTTP 403

**步骤执行结果**:

- Step 1: ❌ POST /api/chat/new-session 创建新会话...
  - 错误: HTTP 403

---

### SMOKE-012 - 会话列表侧边栏（迭代12新增） ❌

- **优先级**: P0
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:58.554415
- **结束时间**: 2026-04-08T20:28:58.558773
- **失败步骤**: Step 6: 点击+新会话按钮
- **错误信息**: HTTP 400

**步骤执行结果**:

- Step 1: ✅ 打开左上角会话入口，侧边栏滑入...
- Step 2: ✅ GET /api/sessions 加载会话列表...
- Step 3: ✅ 验证当前会话高亮标识...
- Step 4: ✅ 点击历史会话切换...
- Step 5: ✅ 验证消息列表更新...
- Step 6: ❌ 点击+新会话按钮...
  - 错误: HTTP 400

---

### SMOKE-013 - 指令系统全面测试（迭代12新增） ❌

- **优先级**: P1
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:58.558882
- **结束时间**: 2026-04-08T20:28:58.559737
- **失败步骤**: Step 1: 输入 #新会话 验证清空消息列表、重置session_id
- **错误信息**: HTTP 400

**步骤执行结果**:

- Step 1: ❌ 输入 #新会话 验证清空消息列表、重置session_id...
  - 错误: HTTP 400

---

### SMOKE-014 - 知识库CRUD + 统一输入框 ❌

- **优先级**: P0
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:58.559773
- **结束时间**: 2026-04-08T20:28:58.605824
- **失败步骤**: Step 1: POST /api/documents { title, content } 添加文本文档
- **错误信息**: HTTP 403

**步骤执行结果**:

- Step 1: ❌ POST /api/documents { title, content } 添加文本文档...
  - 错误: HTTP 403

---

### SMOKE-016 - 向量召回策略 + Scope控制（迭代11优化） ❌

- **优先级**: P1
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:58.605863
- **结束时间**: 2026-04-08T20:28:58.606668
- **失败步骤**: Step 1: 教师上传≥10条知识库文档（含scope=global）
- **错误信息**: HTTP 403

**步骤执行结果**:

- Step 1: ❌ 教师上传≥10条知识库文档（含scope=global）...
  - 错误: HTTP 403

---

### SMOKE-017 - 记忆三层存储 + 摘要合并 ✅

- **优先级**: P0
- **状态**: passed
- **开始时间**: 2026-04-08T20:28:58.606708
- **结束时间**: 2026-04-08T20:28:58.654324

**步骤执行结果**:

- Step 1: ✅ GET /api/memories 查看记忆列表（含 memory_layer）...
- Step 2: ✅ PUT /api/memories/:id 编辑记忆...
- Step 3: ✅ POST /api/memories/summarize 触发记忆摘要合并...
- Step 4: ✅ 验证 core/episodic/archived 三层分类正确...

---

### SMOKE-018 - 画像隐私保护 ✅

- **优先级**: P1
- **状态**: passed
- **开始时间**: 2026-04-08T20:28:58.654345
- **结束时间**: 2026-04-08T20:28:58.655515

**步骤执行结果**:

- Step 1: ✅ GET /api/teachers 查看教师列表...
- Step 2: ✅ 验证返回数据不包含 profile_snapshot 字段...

---

### SMOKE-019 - 学生端与教师端聊天列表 ❌

- **优先级**: P0
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:58.655532
- **结束时间**: 2026-04-08T20:28:58.656358
- **失败步骤**: Step 1: GET /api/chat-list/student 学生端列表
- **错误信息**: HTTP 403

**步骤执行结果**:

- Step 1: ❌ GET /api/chat-list/student 学生端列表...
  - 错误: HTTP 403

---

### SMOKE-020 - 课程发布与推送 ❌

- **优先级**: P1
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:58.656399
- **结束时间**: 2026-04-08T20:28:58.702484
- **失败步骤**: Step 1: POST /api/courses { title, description } 发布课程
- **错误信息**: HTTP 403

**步骤执行结果**:

- Step 1: ❌ POST /api/courses { title, description } 发布课程...
  - 错误: HTTP 403

---

### SMOKE-021 - 对话风格模板设置 ❌

- **优先级**: P1
- **状态**: failed
- **开始时间**: 2026-04-08T20:28:58.702530
- **结束时间**: 2026-04-08T20:28:58.748499
- **失败步骤**: Step 1: PUT /api/styles { style_template: 'socratic' } 设置苏格拉底式
- **错误信息**: HTTP 400

**步骤执行结果**:

- Step 1: ❌ PUT /api/styles { style_template: 'socratic' } 设置苏...
  - 错误: HTTP 400

---

### SMOKE-022 - 教材配置与教师消息推送 ✅

- **优先级**: P1
- **状态**: passed
- **开始时间**: 2026-04-08T20:28:58.748539
- **结束时间**: 2026-04-08T20:28:58.794557

**步骤执行结果**:

- Step 1: ✅ POST /api/curriculum-configs 创建教材配置（学段+教材版本）...
- Step 2: ✅ GET /api/curriculum-configs 验证配置...
- Step 3: ✅ POST /api/teacher-messages 推送消息...

---

### SMOKE-024 - 管理员用户管理 + 禁用用户 ✅

- **优先级**: P0
- **状态**: passed
- **开始时间**: 2026-04-08T20:28:58.794573
- **结束时间**: 2026-04-08T20:28:58.797521

**步骤执行结果**:

- Step 1: ✅ GET /api/admin/users 获取用户列表...
- Step 2: ✅ PUT /api/admin/users/:id/role 修改角色...
- Step 3: ✅ PUT /api/admin/users/:id/status { status: 'disable...
- Step 4: ✅ 被禁用用户访问API 验证返回 403...
- Step 5: ✅ 被禁用用户登录 验证返回 40003...

---

### SMOKE-025 - 操作日志查询与CSV导出 ✅

- **优先级**: P1
- **状态**: passed
- **开始时间**: 2026-04-08T20:28:58.797538
- **结束时间**: 2026-04-08T20:28:58.799521

**步骤执行结果**:

- Step 1: ✅ GET /api/admin/logs 查询日志（支持多条件筛选）...
- Step 2: ✅ GET /api/admin/logs/stats 日志统计...
- Step 3: ✅ GET /api/admin/logs/export 导出CSV...

---

### SMOKE-028 - API限流保护 ✅

- **优先级**: P1
- **状态**: passed
- **开始时间**: 2026-04-08T20:28:58.799537
- **结束时间**: 2026-04-08T20:28:58.821901

**步骤执行结果**:

- Step 1: ✅ 快速连续发送20+请求...
- Step 2: ✅ 验证超限后返回 429 Too Many Requests...
- Step 3: ✅ 等待限流窗口恢复后验证请求正常...

---

### SMOKE-029 - 发现页 + 广场（is_public过滤） ✅

- **优先级**: P1
- **状态**: passed
- **开始时间**: 2026-04-08T20:28:58.821918
- **结束时间**: 2026-04-08T20:28:58.825451

**步骤执行结果**:

- Step 1: ✅ GET /api/discover 验证仅返回 is_public=true 的班级...
- Step 2: ✅ GET /api/discover/search?q=关键词 搜索验证...
- Step 3: ✅ GET /api/personas/marketplace 验证公开班级分身...
- Step 4: ✅ 教师 PUT /api/classes/:id { is_public: false }...
- Step 5: ✅ 再次查询发现页验证该班级不再出现...

---

## 失败用例分析

### SMOKE-030 - 系统健康检查与反馈提交

- **问题归属**: integration
- **错误详情**: HTTP 401
- **失败步骤**: Step 3
- **截图**: screenshots/SMOKE-030/error.json

### SMOKE-001 - 微信登录与教师注册（含自测学生创建）

- **问题归属**: integration
- **错误详情**: HTTP 401: Unknown
- **失败步骤**: Step 1
- **截图**: screenshots/SMOKE-001/error.json

### SMOKE-015 - Python LlamaIndex语义检索服务

- **问题归属**: integration
- **错误详情**: HTTP 401
- **失败步骤**: Step 2
- **截图**: screenshots/SMOKE-015/error.json

### SMOKE-023 - 管理员仪表盘全套统计

- **问题归属**: integration
- **错误详情**: HTTP 403
- **失败步骤**: Step 4
- **截图**: screenshots/SMOKE-023/error.json

### SMOKE-026 - H5微信OAuth登录 + 平台配置 + 文件上传

- **问题归属**: integration
- **错误详情**: HTTP 400
- **失败步骤**: Step 1
- **截图**: screenshots/SMOKE-026/error.json

### SMOKE-027 - 鉴权与角色隔离

- **问题归属**: integration
- **错误详情**: HTTP 403
- **失败步骤**: Step 2
- **截图**: screenshots/SMOKE-027/error.json

### SMOKE-003 - 创建班级（同步创建分身）+ 分身列表验证

- **问题归属**: integration
- **错误详情**: HTTP 403
- **失败步骤**: Step 1
- **截图**: screenshots/SMOKE-003/error.json

### SMOKE-004 - 分享码创建 + 学生扫码加入 + 审批通过

- **问题归属**: integration
- **错误详情**: HTTP 403
- **失败步骤**: Step 1
- **截图**: screenshots/SMOKE-004/error.json

### SMOKE-005 - 教师禁止独立创建分身 + 废弃接口404验证

- **问题归属**: integration
- **错误详情**: 期望404, 实际-1
- **失败步骤**: Step 2
- **截图**: screenshots/SMOKE-005/error.json

### SMOKE-006 - 普通消息发送与AI回复

- **问题归属**: integration
- **错误详情**: HTTP 403
- **失败步骤**: Step 1
- **截图**: screenshots/SMOKE-006/error.json

### SMOKE-008 - 流式中断功能（迭代12新增）

- **问题归属**: integration
- **错误详情**: HTTP -1
- **失败步骤**: Step 3
- **截图**: screenshots/SMOKE-008/error.json

### SMOKE-009 - 中断接口异常场景

- **问题归属**: integration
- **错误详情**: 期望404, 实际-1
- **失败步骤**: Step 1
- **截图**: screenshots/SMOKE-009/error.json

### SMOKE-010 - 教师真人介入对话

- **问题归属**: integration
- **错误详情**: HTTP 403
- **失败步骤**: Step 1
- **截图**: screenshots/SMOKE-010/error.json

### SMOKE-011 - 新会话创建 + 会话列表 + 标题生成

- **问题归属**: integration
- **错误详情**: HTTP 403
- **失败步骤**: Step 1
- **截图**: screenshots/SMOKE-011/error.json

### SMOKE-012 - 会话列表侧边栏（迭代12新增）

- **问题归属**: integration
- **错误详情**: HTTP 400
- **失败步骤**: Step 6
- **截图**: screenshots/SMOKE-012/error.json

### SMOKE-013 - 指令系统全面测试（迭代12新增）

- **问题归属**: integration
- **错误详情**: HTTP 400
- **失败步骤**: Step 1
- **截图**: screenshots/SMOKE-013/error.json

### SMOKE-014 - 知识库CRUD + 统一输入框

- **问题归属**: integration
- **错误详情**: HTTP 403
- **失败步骤**: Step 1
- **截图**: screenshots/SMOKE-014/error.json

### SMOKE-016 - 向量召回策略 + Scope控制（迭代11优化）

- **问题归属**: integration
- **错误详情**: HTTP 403
- **失败步骤**: Step 1
- **截图**: screenshots/SMOKE-016/error.json

### SMOKE-019 - 学生端与教师端聊天列表

- **问题归属**: integration
- **错误详情**: HTTP 403
- **失败步骤**: Step 1
- **截图**: screenshots/SMOKE-019/error.json

### SMOKE-020 - 课程发布与推送

- **问题归属**: integration
- **错误详情**: HTTP 403
- **失败步骤**: Step 1
- **截图**: screenshots/SMOKE-020/error.json

### SMOKE-021 - 对话风格模板设置

- **问题归属**: integration
- **错误详情**: HTTP 400
- **失败步骤**: Step 1
- **截图**: screenshots/SMOKE-021/error.json

## 结论与建议

根据测试结果判断:

- ❌ 存在P0用例失败，冒烟测试未通过
- 需要修复以下问题后才能继续:
  - SMOKE-030: HTTP 401...
  - SMOKE-001: HTTP 401: Unknown...
  - SMOKE-015: HTTP 401...
  - SMOKE-023: HTTP 403...
  - SMOKE-026: HTTP 400...
