# V2.0 迭代11 接口规范

> 本文档定义迭代11所有新增、改造和删除的 API 接口。
> 基础接口规范请参考：[迭代10接口规范](./iteration10_api_spec.md)

---

## 1. 接口变更总览

| 接口 | 方法 | 变更类型 | 说明 |
|------|------|---------|------|
| `/api/classes` | POST | **重构** | 创建班级时同步创建分身，新增分身信息参数和 is_public |
| `/api/classes/:id` | PUT | **增强** | 支持更新 is_public 字段 |
| `/api/personas` | POST | **重构** | 教师角色禁止独立创建分身 |
| `/api/personas` | GET | **增强** | 返回 bound_class_id、bound_class_name、is_public |
| `/api/personas/:id/switch` | PUT | **删除** | 新模式下不需要切换分身 |
| `/api/personas/:id/activate` | PUT | **删除** | 分身随班级管理 |
| `/api/personas/:id/deactivate` | PUT | **删除** | 分身随班级管理 |
| `/api/test-student` | GET | **新增** | 获取自测学生信息 |
| `/api/test-student/reset` | POST | **新增** | 重置自测学生数据 |
| `/api/auth/complete-profile` | POST | **改造** | 教师注册时不再创建分身，改为创建自测学生 |

---

## 2. 班级管理接口

### 2.1 创建班级（重构）

**POST** `/api/classes`

**鉴权**：需要（Bearer Token，教师角色）

**变更说明**：创建班级时**同步创建班级专属分身**。请求体新增分身信息字段和 `is_public` 字段。

**请求体**：
```json
{
  "name": "高一(3)班",
  "description": "2026级高一3班",
  "persona_nickname": "王老师",
  "persona_school": "北京大学",
  "persona_description": "物理学教授，擅长启发式教学",
  "is_public": true
}
```

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| name | string | ✅ | - | 班级名称 |
| description | string | ❌ | "" | 班级描述 |
| persona_nickname | string | ✅ | - | 分身昵称（教师在该班级的显示名） |
| persona_school | string | ✅ | - | 学校名称 |
| persona_description | string | ✅ | - | 分身描述（教学风格、擅长领域等） |
| is_public | bool | ❌ | true | 是否公开（公开班级展示在发现页） |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 10,
    "name": "高一(3)班",
    "description": "2026级高一3班",
    "is_public": true,
    "persona_id": 15,
    "persona_nickname": "王老师",
    "persona_school": "北京大学",
    "persona_description": "物理学教授，擅长启发式教学",
    "share_code": "abc123",
    "share_url": "https://example.com/share/abc123",
    "created_at": "2026-04-04T17:30:00Z"
  }
}
```

**错误响应**：

| 场景 | HTTP | code | message |
|------|------|------|---------|
| 缺少班级名称 | 400 | 40004 | 班级名称不能为空 |
| 缺少分身昵称 | 400 | 40004 | 分身昵称不能为空 |
| 缺少学校名称 | 400 | 40004 | 学校名称不能为空 |
| 缺少分身描述 | 400 | 40004 | 分身描述不能为空 |
| 同名班级已存在 | 409 | 40030 | 该班级名称已存在 |

**业务逻辑**：
1. 创建班级记录（含 is_public 字段）
2. 创建班级专属分身（`bound_class_id = 班级ID`）
3. 自动生成分享码和分享链接
4. 如果教师有自测学生，自动将自测学生加入该班级（自动审批）
5. 以上操作在同一事务中完成

---

### 2.2 更新班级（增强）

**PUT** `/api/classes/:id`

**鉴权**：需要（Bearer Token，教师角色，班级所有者）

**变更说明**：新增 `is_public` 字段支持。

**请求体**：
```json
{
  "name": "高一(3)班（已更名）",
  "description": "更新后的描述",
  "is_public": false
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | ❌ | 班级名称 |
| description | string | ❌ | 班级描述 |
| is_public | bool | ❌ | 是否公开 |

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 10,
    "name": "高一(3)班（已更名）",
    "description": "更新后的描述",
    "is_public": false,
    "updated_at": "2026-04-04T18:00:00Z"
  }
}
```

---

## 3. 分身管理接口

### 3.1 创建分身（重构）

**POST** `/api/personas`

**鉴权**：需要（Bearer Token）

**变更说明**：教师角色**不再允许**通过此接口独立创建分身。

**请求体**（不变）：
```json
{
  "role": "student",
  "nickname": "小明",
  "school": "",
  "description": ""
}
```

**业务逻辑变更**：
1. 如果 `role=teacher`，返回错误：
   ```json
   {
     "code": 40040,
     "message": "教师分身随班级创建，请通过创建班级来创建分身"
   }
   ```
2. 如果 `role=student`，保持原有逻辑不变

---

### 3.2 获取分身列表（增强）

**GET** `/api/personas`

**鉴权**：需要（Bearer Token）

**变更说明**：返回结果中增加 `bound_class_id`、`bound_class_name`、`is_public` 字段。

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "personas": [
      {
        "id": 15,
        "role": "teacher",
        "nickname": "王老师",
        "school": "北京大学",
        "description": "物理学教授",
        "avatar": "",
        "is_active": true,
        "is_public": true,
        "bound_class_id": 10,
        "bound_class_name": "高一(3)班",
        "student_count": 25,
        "created_at": "2026-04-04T17:30:00Z"
      },
      {
        "id": 16,
        "role": "teacher",
        "nickname": "王老师",
        "school": "北京大学",
        "description": "数学教学",
        "avatar": "",
        "is_active": true,
        "is_public": false,
        "bound_class_id": 11,
        "bound_class_name": "高二(1)班",
        "student_count": 30,
        "created_at": "2026-04-04T17:35:00Z"
      }
    ]
  }
}
```

---

### 3.3 删除的接口

以下接口在本迭代中**移除**，请求将返回 `404 Not Found`：

| 接口 | 方法 | 原功能 | 移除原因 |
|------|------|--------|---------|
| `/api/personas/:id/switch` | PUT | 切换当前分身 | 没有主分身概念，不需要切换 |
| `/api/personas/:id/activate` | PUT | 启用分身 | 分身随班级管理 |
| `/api/personas/:id/deactivate` | PUT | 停用分身 | 分身随班级管理 |

---

## 4. 自测学生接口

### 4.1 获取自测学生信息

**GET** `/api/test-student`

**鉴权**：需要（Bearer Token，教师角色）

**说明**：获取当前教师的自测学生账号信息。

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 20,
    "username": "teacher_1_test",
    "persona_id": 21,
    "nickname": "测试学生",
    "password_hint": "abc12345",
    "is_active": true,
    "joined_classes": [
      {
        "class_id": 10,
        "class_name": "高一(3)班",
        "persona_id": 15
      }
    ],
    "created_at": "2026-04-04T17:30:00Z"
  }
}
```

**错误响应**：

| 场景 | HTTP | code | message |
|------|------|------|---------|
| 非教师角色 | 403 | 40039 | 仅教师角色可使用此功能 |
| 自测学生不存在 | 404 | 40004 | 自测学生账号不存在 |

---

### 4.2 重置自测学生

**POST** `/api/test-student/reset`

**鉴权**：需要（Bearer Token，教师角色）

**说明**：清空自测学生与该教师所有班级分身的对话记录和记忆，保留师生关系和班级成员关系。

**请求体**：无

**成功响应** `200`：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "cleared_conversations": 15,
    "cleared_memories": 8,
    "message": "自测学生数据已重置"
  }
}
```

**清空范围**：
1. 自测学生与该教师所有班级分身的对话记录（`conversations` 表）
2. 自测学生与该教师所有班级分身的记忆（`memories` 表）
3. **不清空**：师生关系、班级成员关系、知识库数据

---

## 5. 认证接口改造

### 5.1 完善个人资料（改造）

**POST** `/api/auth/complete-profile`

**鉴权**：需要（Bearer Token）

**变更说明**：教师角色注册时**不再自动创建教师分身**（因为没有主分身概念），改为自动创建自测学生。

**请求体**（不变）：
```json
{
  "role": "teacher",
  "nickname": "王老师",
  "school": "北京大学",
  "description": "物理学教授"
}
```

**业务逻辑变更**：

**教师角色**：
1. 更新用户角色为 teacher
2. ~~创建教师主分身~~ → **不再创建教师分身**
3. **新增**：创建自测学生用户（`teacher_{user_id}_test`，自动生成密码）
4. **新增**：创建自测学生分身

**学生角色**：保持不变

**成功响应** `200`（教师角色）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 1,
    "role": "teacher",
    "nickname": "王老师",
    "test_student": {
      "user_id": 20,
      "username": "teacher_1_test",
      "persona_id": 21,
      "password_hint": "abc12345"
    },
    "message": "教师资料已完善。已为您创建测试学生角色，请通过创建班级来开始使用。"
  }
}
```

---

## 6. 知识库接口（无变更，仅优化说明）

### 6.1 向量召回策略优化

**影响接口**：`POST /api/chat`、`POST /api/chat/stream`（管道中的知识库检索环节）

**优化内容**（后端内部逻辑，不影响 API 接口格式）：

| 参数 | 旧值 | 新值 | 说明 |
|------|------|------|------|
| 向量召回数量 | 20 | 100 | 增大召回池 |
| 置信度阈值 | 无 | 0.3 | 过滤低质量结果 |
| 阈值后最大数量 | 无 | 20 | 控制过滤后数量 |
| scope 过滤后最大数量 | 5 | 5 | 不变 |

**检索流程**：
```
向量召回(100条) → 置信度过滤(score≥0.3, ≤20条) → scope过滤 → 返回(≤5条)
```

---

## 7. 记忆接口（无变更）

以下接口在本迭代中**无需任何调整**，与班级分身模式完全兼容：

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/memories` | GET | 记忆列表（支持 layer 筛选 + SQL 层分页） |
| `/api/memories/:id` | PUT | 编辑记忆 |
| `/api/memories/:id` | DELETE | 删除记忆 |
| `/api/memories/summarize` | POST | 记忆摘要合并 |

**兼容性说明**：
- 记忆按 `teacher_persona_id + student_persona_id` 隔离
- 每个班级分身有独立的 `persona_id`，记忆自然隔离
- 不同班级分身的记忆互不影响

---

## 8. 其他现有接口（无变更）

以下接口组在本迭代中**无需调整**：

| 接口组 | 说明 |
|--------|------|
| 对话接口 (`/api/chat`, `/api/chat/stream`) | 通过 `teacher_persona_id` 路由到正确分身 |
| 知识库 CRUD (`/api/knowledge/*`) | 按 `persona_id` 隔离，scope 机制兼容 |
| 文档管理 (`/api/documents/*`) | 按 `persona_id` 隔离 |
| 师生关系 (`/api/relations/*`) | 按分身维度管理 |
| 班级成员 (`/api/classes/:id/members/*`) | 不变 |
| 置顶功能 (`/api/chat-pins/*`) | 不变 |
| 发现页 (`/api/discover/*`) | 需适配 `is_public` 字段（仅展示公开班级） |
| 课程管理 (`/api/courses/*`) | 不变 |
| 管理员接口 (`/api/admin/*`) | 不变 |

---

## 9. 错误码新增

| 错误码 | HTTP 状态码 | 说明 |
|--------|-----------|------|
| 40040 | 400 | 教师分身随班级创建，请通过创建班级来创建分身 |
| 40041 | 404 | 自测学生账号不存在 |
| 40042 | 400 | 自测学生数据重置失败 |

---

**文档版本**：v1.0
**创建时间**：2026-04-04
**适用迭代**：V2.0 迭代11
**关联文档**：
- [迭代11需求文档](./iteration11_requirements.md)
- [迭代10接口规范](./iteration10_api_spec.md)
