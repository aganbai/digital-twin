# 迭代9冒烟测试异常报告

**测试时间**: 2026-04-03 22:05  
**测试结果**: 4/7 通过 (57.1%)  
**失败用例**: SM-05, SM-06, SM-07

---

## 异常1: SM-05 - AvatarPopup组件未显示

### 测试路径
```
学生登录 → 聊天列表页面 → 点击老师进入聊天页面 → 点击老师头像
```

### 失败详情
- **错误信息**: `AvatarPopup 组件未显示`
- **测试代码位置**: `src/frontend/e2e/iteration9-smoke.test.js:860-925`

### 定位线索

#### 1. 测试脚本尝试的选择器
```javascript
// 尝试的导航栏头像选择器（均未找到）
'.chat-page__teacher-avatar'
'.chat-header__avatar'  
'.chat-page__navbar-back + .chat-page__navbar-title'

// 尝试的弹窗选择器
'.avatar-popup'  // 未找到
```

#### 2. 需要检查的前端代码
- **聊天页面**: `src/frontend/src/pages/chat/index.tsx`
  - 检查是否存在老师头像元素
  - 检查头像的点击事件是否触发
  - 检查是否调用 `setShowAvatarPopup(true)`
  
- **AvatarPopup组件**: `src/frontend/src/components/AvatarPopup/index.tsx`
  - 检查组件是否正确渲染
  - 检查组件的条件渲染逻辑 `showAvatarPopup && <AvatarPopup />`

#### 3. 可能原因
1. **选择器不匹配**: 前端头像元素类名与测试脚本不一致
2. **组件未挂载**: AvatarPopup组件未引入或条件渲染失败
3. **事件未绑定**: 头像元素的点击事件未正确绑定

### 修复建议
```bash
# 检查聊天页面的头像元素类名
grep -n "avatar" src/frontend/src/pages/chat/index.tsx

# 检查AvatarPopup组件是否存在
ls -la src/frontend/src/components/AvatarPopup/

# 检查AvatarPopup的引入和使用
grep -n "AvatarPopup" src/frontend/src/pages/chat/index.tsx
```

---

## 异常2: SM-06 - 教师没有班级

### 测试路径
```
教师登录 → 聊天列表页面
```

### 失败详情
- **错误信息**: `教师没有班级`
- **测试代码位置**: `src/frontend/e2e/iteration9-smoke.test.js:935-1020`

### 定位线索

#### 1. 数据准备状态
测试脚本中的数据准备代码位于 `prepareTestData()` 函数：
```javascript
// 已创建的班级（来自SM-01/SM-04测试）
classId: 16  // 班级ID
teacher_persona_id: 51  // 教师分身ID
student_persona_id: 50  // 学生分身ID
```

#### 2. 教师登录状态
```javascript
// 教师测试账号
teacherCode: 'v9iter_tea'
// 对应分身ID: 51
```

#### 3. 需要检查的数据
```bash
# 检查教师的班级列表
sqlite3 data/digital-twin.db "SELECT * FROM classes WHERE persona_id = 51"

# 检查教师的用户和分身关系
sqlite3 data/digital-twin.db "SELECT p.id, p.user_id, p.role, u.default_persona_id FROM personas p JOIN users u ON p.user_id = u.id WHERE p.id = 51"

# 检查教师聊天列表API响应
curl -H "Authorization: Bearer <teacher_token>" http://localhost:8080/api/chat-list/teacher
```

#### 4. 可能原因
1. **教师token问题**: 教师登录后token中的persona_id不正确
2. **班级归属问题**: 班级创建时关联的教师分身ID与登录教师不匹配
3. **API查询逻辑**: 教师聊天列表API的查询条件有误

### 修复建议
```bash
# 检查教师登录后的token内容
# 在测试脚本中添加：
console.log('Teacher token payload:', JSON.parse(atob(teacherToken.split('.')[1])))

# 检查后端教师聊天列表API逻辑
grep -n "GetTeacherChatList" src/backend/api/handlers_chat_list_v8.go
```

---

## 异常3: SM-07 - 课程发布后未跳转

### 测试路径
```
教师登录 → 首页 → 点击发布课程 → 填写表单 → 点击发布
```

### 失败详情
- **错误信息**: `课程发布后未跳转，当前页面: pages/home/index`
- **测试代码位置**: `src/frontend/e2e/iteration9-smoke.test.js:1030-1140`

### 定位线索

#### 1. 测试期望行为
```javascript
// 课程发布成功后应该跳转到课程详情页
// 预期页面: pages/course-detail/index
// 实际页面: pages/home/index
```

#### 2. 需要检查的前端代码
- **课程发布页面**: `src/frontend/src/pages/course-publish/index.tsx`
  - 检查发布按钮的事件处理
  - 检查API调用成功后的路由跳转逻辑
  
- **课程详情页面**: `src/frontend/src/pages/course-detail/index.tsx`
  - 检查页面是否存在且配置正确

#### 3. 可能原因
1. **API返回格式**: 发布成功但返回的课程ID不正确
2. **路由跳转失败**: `Taro.navigateTo` 调用失败或路径错误
3. **页面未配置**: course-detail页面未在app.config.js中注册

### 修复建议
```bash
# 检查课程发布页面的跳转逻辑
grep -n "navigateTo.*course-detail\|发布成功" src/frontend/src/pages/course-publish/index.tsx

# 检查课程详情页配置
grep -n "course-detail" src/frontend/src/app.config.ts

# 检查后端课程发布API返回
grep -n "CreateCourse\|发布课程" src/backend/api/handlers_course.go
```

---

## 开发Agent任务分配

### 任务1: 修复SM-05 AvatarPopup显示问题
**负责人**: dev_frontend_agent  
**优先级**: P0  
**验证方式**: 重新运行SM-05测试用例

### 任务2: 修复SM-06 教师班级列表问题  
**负责人**: dev_backend_agent  
**优先级**: P0  
**验证方式**: 
```bash
# 验证教师能看到班级
curl -H "Authorization: Bearer <teacher_token>" http://localhost:8080/api/chat-list/teacher
```

### 任务3: 修复SM-07 课程发布跳转问题
**负责人**: dev_frontend_agent  
**优先级**: P1  
**验证方式**: 重新运行SM-07测试用例

---

## 附录: 测试结果摘要

```
✅ SM-01: 消息发送成功，AI已回复（快速响应模式，未展示思考过程）
✅ SM-02: 语音按钮显示正常，录音界面正常显示
✅ SM-03: PlusPanel显示正常，功能选项完整(文件, 相册, 拍摄)
✅ SM-04: 会话列表二级结构正常，展开/折叠交互正常，历史会话8条
❌ SM-05: AvatarPopup 组件未显示
❌ SM-06: 教师没有班级
❌ SM-07: 课程发布后未跳转，当前页面: pages/home/index
```

**下一步**: 请相关开发agent根据上述定位线索排查问题，修复后重新运行冒烟测试验证。
