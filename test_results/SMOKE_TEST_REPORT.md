# 迭代11冒烟测试报告

**测试时间**: 2026-04-05 00:25:00
**测试环境**: macOS, Node.js, 后端服务 localhost:8080
**测试文件**: smoke-v11-quick-check.js, smoke-v11-core-test.js

## 测试摘要

| 用例ID | 测试项 | 状态 | 说明 |
|--------|--------|------|------|
| SM-AD04 | 废弃接口返回404 | ✅ 通过 | /api/personas/:id/{switch,activate,deactivate} 均返回404 |
| SM-AD02 | 教师禁止独立创建分身 | ❌ 失败 | 返回409/40015而非400/40040，需检查JWT role字段 |
| SM-AD01 | 班级创建同步创建分身 | ❌ 失败 | 返回401，JWT token验证失败 |
| SM-AE01 | 自测学生功能 | ⚠️ 未实现 | /api/test-student 接口不存在 |

## 详细分析

### 1. SM-AD04: 废弃接口返回404 ✅

**测试结果**:
```
/api/personas/1/switch: 404 ✓
/api/personas/1/activate: 404 ✓
/api/personas/1/deactivate: 404 ✓
```

**结论**: 后端路由修复生效，废弃接口已正确移除。

### 2. SM-AD02: 教师禁止独立创建分身 ❌

**期望**: 返回 400 状态码，错误码 40040
**实际**: 返回 409 状态码，错误码 40015

**问题分析**:
- 代码中第24行有教师角色检查：`if roleStr == "teacher" { Error(..., 40040) }`
- 但实际执行到了第62行的唯一性检查，说明role字段可能未正确传递
- JWT中间件设置了role字段，但可能类型不匹配

**需要检查**:
1. JWT token中的role字段值
2. `handlers_persona.go`中`c.Get("role")`的返回值类型

### 3. SM-AD01: 班级创建同步创建分身 ❌

**期望**: 班级创建成功，返回persona_id、bound_class_id
**实际**: 返回401未授权错误

**问题分析**:
- JWT token验证失败
- 可能是user_id字段类型断言失败
- 代码中使用`userIDInt64, ok := userID.(int64)`，但JWT中间件可能设置的是float64

**需要检查**:
1. JWT中间件中user_id的类型
2. 类型断言逻辑

### 4. SM-AE01: 自测学生功能 ⚠️

**状态**: 功能未实现
- `/api/test-student` 接口返回404
- 需要实现自测学生管理接口

## 数据库验证

**迭代11字段已存在**:
```sql
sqlite> PRAGMA table_info(personas);
...
10|is_public|INTEGER|0|0|0
11|bound_class_id|INTEGER|0|NULL|0
```

**索引已创建**:
- idx_classes_is_public
- idx_personas_bound_class_id

## 后续行动

### 紧急修复 (P0)
1. 检查JWT中间件的user_id类型设置，确保为int64
2. 验证role字段的传递逻辑

### 中等优先级 (P1)
1. 实现自测学生接口（/api/test-student）
2. 添加更详细的错误日志

### 低优先级 (P2)
1. 优化测试脚本，添加自动清理机制
2. 完善测试文档

## 测试数据

**测试数据库**: `/Users/aganbai/Desktop/WorkSpace/digital-twin/data/digital-twin.db`
**测试结果**: `/Users/aganbai/Desktop/WorkSpace/digital-twin/test_results/smoke-v11-quick-check.json`

---

**生成时间**: 2026-04-05 00:25:00
**执行人**: 自动化测试脚本
