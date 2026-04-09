# V2.0 迭代13 API规范文档

## 概述

本文档定义 V2.0 IT13 教材配置流程重构涉及的所有 API 变更，包括接口扩展、请求/响应格式、错误码定义和影响范围说明。

---

## 1. 接口变更总览

### 1.1 变更汇总表

| 接口 | 方法 | 变更类型 | 优先级 | 影响模块 | 向后兼容 |
|------|------|----------|--------|----------|----------|
| POST /api/classes | 扩展 | P0 | B1→F2 | 是 |
| PUT /api/classes/:id | 扩展 | P0 | B2→F3 | 是 |
| GET /api/classes/:id | 增强 | P1 | B3→F3 | 是 |

### 1.2 现有接口保留

| 接口 | 说明 | 备注 |
|------|------|------|
| POST /api/curriculum-configs | 创建教材配置 | 保留用于独立页面降级 |
| GET /api/curriculum-configs | 查询教材配置列表 | 保留用于独立页面降级 |
| PUT /api/curriculum-configs/:id | 更新教材配置 | 保留用于独立页面降级 |
| DELETE /api/curriculum-configs/:id | 删除教材配置 | 保留用于独立页面降级 |

---

## 2. 接口详细规范

### 2.1 POST /api/classes（B1扩展）

#### 基本信息

- **接口地址**: `/api/classes`
- **请求方法**: POST
- **认证要求**: 需要登录（教师角色）
- **变更说明**: 请求体新增可选的 `curriculum_config` 字段

#### 请求参数

**Header 参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| Authorization | string | 是 | Bearer Token |

**Body 参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| name | string | 是 | 班级名称，1-50字符 |
| description | string | 否 | 班级描述，最多200字符 |
| persona_nickname | string | 是 | 分身昵称，1-30字符 |
| persona_school | string | 是 | 学校名称，1-50字符 |
| persona_description | string | 是 | 分身描述，最多200字符 |
| is_public | boolean | 否 | 是否公开，默认true |
| **curriculum_config** | object | 否 | **新增：教材配置信息** |
| curriculum_config.grade_level | string | 否 | 学段枚举值 |
| curriculum_config.grade | string | 否 | 年级名称 |
| curriculum_config.subjects | string[] | 否 | 学科列表 |
| curriculum_config.textbook_versions | string[] | 否 | 教材版本列表 |
| curriculum_config.custom_textbooks | string[] | 否 | 自定义教材列表 |
| curriculum_config.current_progress | string | 否 | 当前教学进度 |

** curricula_config 字段详细说明**:

| 字段 | 类型 | 约束 | 选项值 |
|------|------|------|--------|
| grade_level | string | 可选 | preschool, primary_lower, primary_upper, junior, senior, university, adult_life, adult_professional |
| grade | string | 可选 | 根据 grade_level 动态 |
| subjects | string[] | 可选，K12必填 | K12_SUBJECTS 或 ADULT_*_CATEGORIES |
| textbook_versions | string[] | 可选 | TEXTBOOK_VERSIONS |
| custom_textbooks | string[] | 可选，university必填 | 用户自定义输入 |
| current_progress | string | 可选 | 自由文本，如"第三单元" |

**请求示例（完整）**:
```json
{
  "name": "三年级数学班",
  "description": "小学数学培优班级",
  "persona_nickname": "王老师",
  "persona_school": "实验小学",
  "persona_description": "10年数学教学经验，专注小学奥数",
  "is_public": true,
  "curriculum_config": {
    "grade_level": "primary_lower",
    "grade": "三年级",
    "subjects": ["数学"],
    "textbook_versions": ["人教版", "北师大版"],
    "custom_textbooks": ["《小学奥数启蒙》"],
    "current_progress": "第三单元 乘法初步"
  }
}
```

**请求示例（无教材配置）**:
```json
{
  "name": "临时班级",
  "description": "",
  "persona_nickname": "李老师",
  "persona_school": "实验中学",
  "persona_description": "临时班级，暂不需要教材配置",
  "is_public": false
}
```

#### 响应参数

**成功响应（200 OK）**:
| 字段 | 类型 | 说明 |
|------|------|------|
| code | integer | 0 表示成功 |
| message | string | 成功信息 |
| data | object | 创建的班级信息 |
| data.id | integer | 班级ID |
| data.name | string | 班级名称 |
| data.description | string | 班级描述 |
| data.is_public | boolean | 是否公开 |
| data.persona_id | integer | 关联的分身ID |
| data.persona_nickname | string | 分身昵称 |
| data.persona_school | string | 学校名称 |
| data.persona_description | string | 分身描述 |
| data.teacher_id | integer | 教师用户ID |
| data.created_at | string | 创建时间(ISO8601) |
| data.token | string | 新生成的JWT（包含新分身） |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 123,
    "name": "三年级数学班",
    "description": "小学数学培优班级",
    "is_public": true,
    "persona_id": 456,
    "persona_nickname": "王老师",
    "persona_school": "实验小学",
    "persona_description": "10年数学教学经验，专注小学奥数",
    "teacher_id": 789,
    "created_at": "2026-04-09T10:30:00Z",
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

#### 错误码定义

| 错误码 | HTTP状态码 | 错误信息 | 说明 |
|--------|------------|----------|------|
| 40004 | 400 | 请求参数无效 | 参数校验失败 |
| 40030 | 409 | 该班级名称已存在 | 同一教师下班级名重复 |
| 40041 | 400 | 无效的学段类型 | grade_level 不在枚举值中 |
| 40301 | 403 | 仅教师角色可创建班级 | 非教师角色访问 |
| 50001 | 500 | 数据库服务不可用 | 数据库连接异常 |

#### 业务处理逻辑

```
1. 参数校验
   - 校验必填字段（name, persona_nickname, persona_school, persona_description）
   - 校验班级名称唯一性（在当前用户范围内）
   - 如有 curriculum_config，校验 grade_level 枚举值

2. 开启数据库事务

3. 创建班级专属分身（INSERT personas）

4. 创建班级记录（INSERT classes）

5. 更新分身 bound_class_id

6. 如提供了 curriculum_config:
   a. 序列化 JSON 字段（textbook_versions, subjects, custom_textbooks）
   b. INSERT teacher_curriculum_configs（关联 persona_id）
   c. 如异常：记录错误日志，继续提交事务（不影响班级创建）

7. 提交事务

8. 返回结果（含新生成的token）
```

---

### 2.2 PUT /api/classes/:id（B2扩展）

#### 基本信息

- **接口地址**: `/api/classes/:id`
- **请求方法**: PUT
- **认证要求**: 需要登录（教师角色）
- **变更说明**: 请求体新增可选的 `curriculum_config` 字段

#### 请求参数

**Path 参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | 班级ID |

**Header 参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| Authorization | string | 是 | Bearer Token |

**Body 参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| name | string | 否 | 班级名称 |
| description | string | 否 | 班级描述 |
| is_public | boolean | 否 | 是否公开 |
| **curriculum_config** | object | 否 | **新增：教材配置信息** |
| curriculum_config.grade_level | string | 否 | 学段枚举值 |
| curriculum_config.grade | string | 否 | 年级名称 |
| curriculum_config.subjects | string[] | 否 | 学科列表 |
| curriculum_config.textbook_versions | string[] | 否 | 教材版本列表 |
| curriculum_config.custom_textbooks | string[] | 否 | 自定义教材列表 |
| curriculum_config.current_progress | string | 否 | 当前教学进度 |

> 注意：curriculum_config 整体为可选，但内部字段在传入时按必填处理

**请求示例（更新班级和配置）**:
```json
{
  "name": "三年级数学班（已更名）",
  "description": "更新后的描述",
  "is_public": true,
  "curriculum_config": {
    "grade_level": "primary_lower",
    "grade": "三年级",
    "subjects": ["数学", "奥数"],
    "textbook_versions": ["人教版"],
    "custom_textbooks": ["《小学奥数进阶》"],
    "current_progress": "第四单元 除法"
  }
}
```

**请求示例（仅更新班级信息，保留原配置）**:
```json
{
  "name": "三年级数学班（已更名）",
  "description": "仅更新描述，不更改配置"
}
```

#### 响应参数

**成功响应（200 OK）**:
| 字段 | 类型 | 说明 |
|------|------|------|
| code | integer | 0 表示成功 |
| message | string | 成功信息 |
| data | object | 更新后的班级信息 |
| data.id | integer | 班级ID |
| data.name | string | 班级名称 |
| data.description | string | 班级描述 |
| data.persona_id | integer | 关联的分身ID |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 123,
    "name": "三年级数学班（已更名）",
    "description": "更新后的描述",
    "persona_id": 456
  }
}
```

#### 错误码定义

| 错误码 | HTTP状态码 | 错误信息 | 说明 |
|--------|------------|----------|------|
| 40004 | 400 | 请求参数无效 | 参数校验失败 |
| 40016 | 409 | 班级名称已存在 | 与其他班级名重复 |
| 40017 | 404 | 班级不存在 | 班级ID无效 |
| **40018** | **403** | **无权操作此班级** | 班级不属于当前教师分身 |
| 40041 | 400 | 无效的学段类型 | grade_level 不在枚举值中 |
| 50001 | 500 | 数据库服务不可用 | 数据库连接异常 |

#### 业务处理逻辑

```
1. 参数校验
   - 校验班级ID有效性
   - 如有 curriculum_config，校验 grade_level 枚举值

2. 权限校验
   - 查询班级信息
   - 验证班级属于当前教师分身（class.persona_id == ctx.persona_id）

3. 开启数据库事务

4. 更新班级基本信息（如有变更）

5. 如提供了 curriculum_config:
   a. 通过班级 persona_id 查询现有教材配置
   b. 如存在配置:
      - UPDATE teacher_curriculum_configs
   c. 如不存在配置:
      - INSERT teacher_curriculum_configs

6. 提交事务

7. 返回更新结果
```

---

### 2.3 GET /api/classes/:id（B3增强）

#### 基本信息

- **接口地址**: `/api/classes/:id`
- **请求方法**: GET
- **认证要求**: 需要登录（教师角色）
- **变更说明**: 响应体新增 `curriculum_config` 字段

#### 请求参数

**Path 参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | integer | 是 | 班级ID |

**Header 参数**:
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| Authorization | string | 是 | Bearer Token |

#### 响应参数

**成功响应（200 OK）**:
| 字段 | 类型 | 说明 |
|------|------|------|
| code | integer | 0 表示成功 |
| message | string | 成功信息 |
| data | object | 班级详情 |
| data.id | integer | 班级ID |
| data.name | string | 班级名称 |
| data.description | string | 班级描述 |
| data.is_public | boolean | 是否公开 |
| data.is_active | boolean | 是否启用 |
| data.persona_id | integer | 关联的分身ID |
| data.teacher_id | integer | 教师用户ID |
| data.created_at | string | 创建时间 |
| data.updated_at | string | 更新时间 |
| **data.curriculum_config** | object\|null | **新增：教材配置信息** |
| curriculum_config.id | integer | 配置ID |
| curriculum_config.grade_level | string | 学段 |
| curriculum_config.grade | string | 年级 |
| curriculum_config.subjects | string[] | 学科列表 |
| curriculum_config.textbook_versions | string[] | 教材版本 |
| curriculum_config.custom_textbooks | string[] | 自定义教材 |
| curriculum_config.current_progress | string | 教学进度 |

**响应示例（有教材配置）**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 123,
    "name": "三年级数学班",
    "description": "小学数学培优班级",
    "is_public": true,
    "is_active": true,
    "persona_id": 456,
    "teacher_id": 789,
    "created_at": "2026-04-09T10:30:00Z",
    "updated_at": "2026-04-09T10:30:00Z",
    "curriculum_config": {
      "id": 1001,
      "grade_level": "primary_lower",
      "grade": "三年级",
      "subjects": ["数学"],
      "textbook_versions": ["人教版", "北师大版"],
      "custom_textbooks": ["《小学奥数启蒙》"],
      "current_progress": "第三单元 乘法初步"
    }
  }
}
```

**响应示例（无教材配置）**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 123,
    "name": "临时班级",
    "description": "",
    "is_public": false,
    "is_active": true,
    "persona_id": 456,
    "teacher_id": 789,
    "created_at": "2026-04-09T10:30:00Z",
    "updated_at": "2026-04-09T10:30:00Z",
    "curriculum_config": null
  }
}
```

#### 错误码定义

| 错误码 | HTTP状态码 | 错误信息 | 说明 |
|--------|------------|----------|------|
| 40004 | 400 | 无效的班级ID | 参数格式错误 |
| 40017 | 404 | 班级不存在 | 班级ID不存在 |
| 40018 | 403 | 无权操作此班级 | 班级不属于当前教师分身 |
| 50001 | 500 | 数据库服务不可用 | 数据库连接异常 |

#### 业务处理逻辑

```
1. 参数校验
   - 校验班级ID有效性

2. 权限校验
   - 查询班级信息
   - 验证班级属于当前教师分身

3. 查询班级详情
   - SELECT * FROM classes WHERE id = ?

4. 查询关联教材配置
   - SELECT * FROM teacher_curriculum_configs
     WHERE persona_id = ? AND is_active = 1
     ORDER BY updated_at DESC LIMIT 1

5. 解析 JSON 字段（textbook_versions, subjects, custom_textbooks）

6. 组装响应数据
```

---

## 3. 枚举值定义

### 3.1 GradeLevel 学段

| 枚举值 | 标签 | 年级选项 |
|--------|------|----------|
| preschool | 学前班 | 幼儿园大班、学前 |
| primary_lower | 小学低年级 | 一年级、二年级、三年级 |
| primary_upper | 小学高年级 | 四年级、五年级、六年级 |
| junior | 初中 | 七年级、八年级、九年级 |
| senior | 高中 | 高一、高二、高三 |
| university | 大学及以上 | 大一、大二、大三、大四、研究生、博士 |
| adult_life | 成人生活技能 | （无固定年级） |
| adult_professional | 成人职业培训 | （无固定年级） |

### 3.2 K12 Subjects 学科

```
['语文', '数学', '英语', '物理', '化学', '生物', '历史', '地理', '政治', '音乐', '美术', '体育', '信息技术']
```

### 3.3 Adult Categories 成人课程类别

**成人生活技能**:
```
['中餐', '西餐', '烘焙', '力量训练', '有氧运动', '瑜伽', '手工', '园艺', '摄影', '绘画']
```

**成人职业培训**:
```
['编程', '设计', '会计', '法律', '医学', '教育', '管理', '营销', '外语', '考证培训']
```

### 3.4 Textbook Versions 教材版本

```
['人教版', '北师大版', '苏教版', '沪教版', '部编版', '外研版', '浙教版', '冀教版']
```

---

## 4. 前后端接口约定

### 4.1 前端数据结构约定

**CurriculumConfigValue (前端使用)**:
```typescript
interface CurriculumConfigValue {
  grade_level?: string;        // 学段值
  grade?: string;              // 年级名称
  subjects?: string[];         // 学科数组
  textbook_versions?: string[]; // 教材版本数组
  custom_textbooks?: string[];  // 自定义教材数组
  current_progress?: string;   // 教学进度文本
}
```

### 4.2 数据转换规则

**前端 → 后端**:
```javascript
// 表单收集的数据
const formData = {
  grade_level: 'primary_lower',
  grade: '一年级',
  subjects: ['语文', '数学'],
  textbook_versions: ['人教版'],
  custom_textbooks: ['教辅A'],
  current_progress: '第二单元'
};

// 组装到创建/更新请求
const requestBody = {
  name: '班级名称',
  // ... 其他字段
  curriculum_config: formData  // 直接传递对象
};
```

**后端 → 前端**:
```javascript
// 后端返回的数据（已解析JSON）
const response = {
  curriculum_config: {
    id: 1001,
    grade_level: 'primary_lower',
    grade: '一年级',
    subjects: ['语文', '数学'],      // 后端已解析为数组
    textbook_versions: ['人教版'],    // 后端已解析为数组
    custom_textbooks: ['教辅A'],      // 后端已解析为数组
    current_progress: '第二单元'
  }
};

// 直接用于表单初始值填充
setFormValue(response.curriculum_config || {});
```

### 4.3 空值处理约定

| 场景 | 前端行为 | 后端行为 |
|------|----------|----------|
| 用户不填任何配置 | 不传递 curriculum_config 字段 或 传递 null | 不创建/不更新配置 |
| 用户清空已有配置 | 传递 curriculum_config: {}（空对象） | 仅更新为激活状态的配置数据 |
| 配置字段留空 | 传递空字符串或空数组 | 按空值存储 |
| 查询无配置 | - | 返回 curriculum_config: null |

---

## 5. 接口变更影响范围

### 5.1 后端影响

| 文件 | 变更内容 | 影响 |
|------|----------|------|
| api/handlers_class.go | HandleCreateClass 扩展 | 新增教材配置创建逻辑 |
| api/handlers_class.go | HandleUpdateClass 扩展 | 新增教材配置更新逻辑 |
| api/handlers_class.go | 新增/修改查询逻辑 | 返回教材配置详情 |
| database/repository_curriculum.go | 可能需要新增方法 | 按persona_id查询活跃配置 |

### 5.2 前端影响

| 文件 | 变更内容 | 影响 |
|------|----------|------|
| pages/class-create/index.tsx | 集成表单组件 | 新增教材配置区域 |
| pages/class-edit/index.tsx | 集成表单组件+加载配置 | 新增教材配置编辑区域 |
| api/class.ts | 更新类型定义 | 添加 curriculum_config 字段 |
| pages/class-detail/index.tsx | 移除配置按钮 | 不再跳转到 curriculum-config |
| app.config.ts | 废弃入口 | 可能从菜单移除 curriculum-config |
| custom-tab-bar/index.tsx | 废弃导航入口 | 移除教材配置菜单项 |

### 5.3 向后兼容性

| 变更 | 兼容策略 |
|------|----------|
| POST /api/classes 扩展 | curriculum_config 为可选字段，不传则现有逻辑不变 |
| PUT /api/classes/:id 扩展 | curriculum_config 为可选字段，不传则配置不变 |
| GET /api/classes/:id 增强 | 新增字段不影响现有解析 |
| 废弃 curriculum-config 入口 | 保留页面代码，仅移除跳转，URL 直接访问可用 |

---

## 6. API 版本说明

本迭代 API 变更采用**向后兼容扩展**策略，不提升 API 版本号。

- 原有接口 `/api/classes` 保持功能不变
- 新增字段均为可选（optional）
- 前端可根据需要选择是否传递新字段

如需强制要求新版本，可考虑后续添加 API 版本头:
```
API-Version: 2.1
```
