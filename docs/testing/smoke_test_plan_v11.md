# 迭代11 冒烟测试计划（新增模块）

> 本文档是 `smoke_test_plan.md` 的补充，定义迭代11新增的冒烟测试用例。
> **核心变更**：分身机制重构（无主分身，班级绑定分身）、自测学生功能、向量召回策略优化。

---

## 模块 AD：班级绑定分身 - 迭代11新增（5 条）

| 编号 | 场景 | 前置条件 | 操作步骤 | 预期结果 | 清理操作 |
|------|------|----------|----------|----------|----------|
| **SM-AD01** | 教师创建班级同步创建分身 | 教师已登录，无班级 | 1. 进入班级创建页<br>2. 填写班级名称"测试班级"<br>3. 填写分身昵称"张老师"<br>4. 填写学校"测试大学"<br>5. 填写分身描述"测试教师"<br>6. 点击创建 | 1. 创建成功<br>2. 返回 persona_id<br>3. 分身 bound_class_id = 班级ID<br>4. is_public 默认为 true | 删除班级和对应分身 |
| **SM-AD02** | 教师禁止独立创建分身 | 教师已登录 | 1. 调用 POST /api/personas<br>2. role=teacher<br>3. 填写分身信息 | 1. 返回错误码 40040<br>2. 提示"教师分身随班级创建，请通过创建班级来创建分身" | 无需清理 |
| **SM-AD03** | 分身列表展示班级信息 | 教师已创建班级 | 1. 进入个人中心<br>2. 查看分身列表 | 1. 每个分身显示 bound_class_id<br>2. 每个分身显示 bound_class_name<br>3. 每个分身显示 is_public 状态 | 无需清理 |
| **SM-AD04** | 已删除的接口返回404 | 教师已登录，已有分身 | 1. 调用 PUT /api/personas/:id/switch<br>2. 调用 PUT /api/personas/:id/activate<br>3. 调用 PUT /api/personas/:id/deactivate | 1. 三个接口均返回 404<br>2. 提示"接口已移除" | 无需清理 |
| **SM-AD05** | 班级 is_public 设置 | 教师已创建班级 | 1. 进入班级设置页<br>2. 查看 is_public 开关<br>3. 切换为非公开<br>4. 保存 | 1. is_public 默认显示为公开<br>2. 切换后状态更新<br>3. 引导语正确展示 | 恢复 is_public 为 true |

---

## 模块 AE：自测学生 - 迭代11新增（4 条）

| 编号 | 场景 | 前置条件 | 操作步骤 | 预期结果 | 清理操作 |
|------|------|----------|----------|----------|----------|
| **SM-AE01** | 教师注册自动创建自测学生 | 新用户注册教师 | 1. 新用户微信登录<br>2. 选择教师角色<br>3. 完成注册 | 1. 自动创建自测学生用户（teacher_{user_id}_test）<br>2. 返回自测学生信息<br>3. 不创建教师分身 | 删除自测学生账号 |
| **SM-AE02** | 获取自测学生信息 | 教师已注册 | 1. 进入个人中心<br>2. 点击"自测学生"<br>3. 查看自测学生信息 | 1. 显示用户名、昵称、密码提示<br>2. 显示已加入的班级列表<br>3. 显示创建时间 | 无需清理 |
| **SM-AE03** | 自测学生自动加入班级 | 教师已创建班级 | 1. 教师创建新班级<br>2. 查看班级成员列表 | 1. 自测学生自动加入班级<br>2. 成员状态为已审批<br>3. 无需教师手动审批 | 从班级移除自测学生 |
| **SM-AE04** | 重置自测学生数据 | 教师已有自测学生对话数据 | 1. 进入自测学生管理页<br>2. 点击"重置数据"<br>3. 确认重置 | 1. 对话记录已清空<br>2. 记忆数据已清空<br>3. 师生关系和班级成员关系保留 | 无需清理 |

---

## 模块 AF：向量召回优化 - 迭代11新增（2 条）

| 编号 | 场景 | 前置条件 | 操作步骤 | 预期结果 | 清理操作 |
|------|------|----------|----------|----------|----------|
| **SM-AF01** | 知识库向量召回100条 | 教师已上传知识库文档（≥100条） | 1. 学生发起对话<br>2. 后端日志查看召回数量<br>3. 查看返回结果 | 1. 向量召回 100 条<br>2. 置信度阈值过滤后 ≤20 条<br>3. scope 过滤后 ≤5 条<br>4. 日志可验证 | 无需清理 |
| **SM-AF02** | 知识库 scope=global 生效 | 教师已创建≥2个班级，上传 scope=global 文档 | 1. 学生与班级A分身对话<br>2. 学生与班级B分身对话<br>3. 验证 global 知识库对两个分身都生效 | 1. 班级A分身能引用 global 知识<br>2. 班级B分身能引用 global 知识<br>3. 对话内容正确 | 无需清理 |

---

## 迭代11 测试依赖链

```
SM-A02（教师注册V11）→ SM-AE01（自测学生创建）
        ↓
SM-AD01（创建班级+分身）→ SM-AD03（分身列表）→ SM-AD05（is_public设置）
        ↓                       ↓
SM-AE03（自测学生自动加入班级）  SM-AD04（已删除接口404）
        ↓
SM-AE02（获取自测学生信息）
        ↓
SM-AE04（重置自测学生数据）

SM-AD02（禁止独立创建分身）← 独立执行

SM-AF01（向量召回100条）← 依赖知识库数据
SM-AF02（scope=global生效）← 依赖多班级+知识库数据
```

---

## 迭代11 冒烟测试文件组织

```
tests/integration/
├── smoke_v11_test.go          # 迭代11冒烟测试（模块 AD/AE/AF）
└── integration_v11_test.go    # 迭代11集成测试（IT-601~IT-612）
```

---

## 清理机制设计

每个冒烟测试用例必须实现以下清理机制：

### 1. 测试数据清理函数

```go
// cleanupIteration11 清理迭代11测试数据
func cleanupIteration11(t *testing.T, token string) {
    t.Helper()
    
    // 1. 获取所有班级
    resp, body, _ := doRequest("GET", "/api/classes", nil, token)
    // ... 解析班级列表
    
    // 2. 删除每个班级（同时删除对应分身）
    for _, classID := range classIDs {
        doRequest("DELETE", fmt.Sprintf("/api/classes/%d", classID), nil, token)
    }
    
    // 3. 获取自测学生信息
    resp, body, _ = doRequest("GET", "/api/test-student", nil, token)
    // ... 如果存在，删除自测学生
    
    t.Logf("✅ 清理完成: 删除 %d 个班级，%d 个分身", len(classIDs), len(classIDs))
}
```

### 2. 每个用例独立清理

```go
func TestSmoke_AD01_CreateClassWithPersona(t *testing.T) {
    // Setup: 确保测试环境干净
    cleanupIteration11(t, teacherToken)
    
    // 执行测试
    // ...
    
    // Teardown: 清理测试数据
    cleanupIteration11(t, teacherToken)
}
```

### 3. 异常日志捕获机制

```go
// captureTestError 捕获测试异常并提取日志
func captureTestError(t *testing.T, operation string, err error, resp *http.Response, body []byte) {
    if err != nil || resp.StatusCode >= 400 {
        t.Errorf("❌ %s 失败: %v", operation, err)
        
        // 从响应体提取错误信息
        var apiResp apiResponse
        json.Unmarshal(body, &apiResp)
        
        // 保存错误日志到文件，供开发Agent分析
        errorLog := fmt.Sprintf(`
[%s] 测试失败
操作: %s
HTTP状态: %d
错误码: %d
错误信息: %s
响应体: %s
`, time.Now().Format(time.RFC3339), operation, resp.StatusCode, apiResp.Code, apiResp.Message, string(body))
        
        os.WriteFile(fmt.Sprintf("test_errors/%s_%d.log", operation, time.Now().Unix()), []byte(errorLog), 0644)
        
        // 输出到控制台（供CI Agent捕获）
        t.Logf("🔍 错误日志已保存，路径: test_errors/%s_%d.log", operation, time.Now().Unix())
    }
}
```

---

## 用例数量汇总

| 迭代 | 新增模块 | 新增用例数 |
|------|---------|-----------|
| 迭代11 | 模块 AD（班级绑定分身） | 5 条 |
| 迭代11 | 模块 AE（自测学生） | 4 条 |
| 迭代11 | 模块 AF（向量召回优化） | 2 条 |
| **合计** | **3 个模块** | **11 条** |

---

**文档版本**: v1.5.0
**创建日期**: 2026-04-04
**适用迭代**: V2.0 迭代11
**关联文档**:
- [迭代11需求文档](./iteration11_requirements.md)
- [迭代11接口规范](./iteration11_api_spec.md)
- [冒烟测试计划](./smoke_test_plan.md)
