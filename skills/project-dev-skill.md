# Digital-Twin 迭代开发 Skill

> **通用核心规则**请加载：`../../skills/shared/iteration-dev/core.md`（~15KB，每次必须加载）
> **通用详细参考**按需查阅：`../../skills/shared/iteration-dev/reference.md`（~75KB，按需读取对应章节）
> **项目技术规范**请参考：`skills/project-tech-spec.md`

---

## 📁 项目配置（动态加载）

以下配置已从本文件剥离，由 dt-orchest 动态加载：

| 配置项 | 文件路径 | 说明 |
|--------|----------|------|
| 项目概述 | `docs/project_config.yaml` | 项目名称、技术栈、核心路径 |
| 已知陷阱 | `docs/known_traps.md` | 历史迭代踩坑点 |
| 冒烟测试 | `docs/testing/smoke_test_plan.md` | 端到端测试用例 |
| 模块划分 | `docs/iterations/v{X}.0/iteration{N}_modules.yaml` | 每次迭代动态生成 |

---

## 1. 模块划分原则

| 原则 | 说明 |
|------|------|
| **后端优先级** | P0（核心）> P1（重要）> P2（次要） |
| **依赖层级** | Layer 0 无依赖，Layer N 依赖 Layer N-1 的模块 |
| **前端模块** | 无强依赖，基于后端接口文档并行开发 |
| **测试用例** | 每个后端模块至少对应 1 个集成测试用例 |
| **测试分批** | 每批 ≤ 5 个用例，按数据依赖分组 |

---

## 2. 执行指令模板

> **使用说明**：以下模板中的占位符由 dt-orchest 动态替换。
> - `{VERSION}` → 版本号（如 `2.0`）
> - `{ITERATION}` → 迭代号（如 `8`）
> - `{PROJECT_NAME}` → 项目名称（从 project_config.yaml 加载）
> - `{MODULES}` → 模块列表（从 iteration{N}_modules.yaml 加载）

### 2.1 启动后端开发 Agent

```
@dev_backend_agent

## 任务：第{ITERATION}迭代后端开发

### 输入文档
- 需求文档: docs/iterations/v{VERSION}/iteration{ITERATION}_requirements.md
- 接口规范: docs/iterations/v{VERSION}/iteration{ITERATION}_api_spec.md
- 项目配置: docs/project_config.yaml
- 项目技术规范: skills/project-tech-spec.md
- 通用核心Skill: ../../skills/shared/iteration-dev/core.md

### 环境保护规则
⚠️ 请严格遵守项目技术规范第2节"环境保护规则"，禁止修改项目配置中标记的受保护文件。

### 模块划分
{MODULES}

### 开发规则
1. **每个模块必须拉起独立的 sub agent，严禁合并多个模块到一个 sub agent 中**（通用Skill R4）
2. 每个模块需要2个独立的 sub agent：1个负责开发代码，1个负责编写并执行单元测试 + 代码Review（通用Skill R1）
3. 同一层级的开发 sub agent 尽量并行拉起，开发完成后再并行拉起对应的单测&Review sub agent
4. **sub agent 的 prompt 只包含结论和指令，不包含编排者的推理过程**（通用Skill R5）
5. 单测不通过或Review发现问题时，反馈给开发 sub agent 修复，最多3轮
6. 所有模块完成后，执行 `go build ./...` 确认整体编译通过
7. 执行 `go test ./...` 确认所有单元测试通过
```

### 2.2 启动集成测试 Agent

```
@ci_test_agent

## 任务：第{ITERATION}迭代集成测试 + 迭代级Review

### 输入文档
- 需求文档: docs/iterations/v{VERSION}/iteration{ITERATION}_requirements.md
- 接口规范: docs/iterations/v{VERSION}/iteration{ITERATION}_api_spec.md
- 项目配置: docs/project_config.yaml
- 已知陷阱: docs/known_traps.md
- 项目技术规范: skills/project-tech-spec.md
- 通用核心Skill: ../../skills/shared/iteration-dev/core.md

### 环境保护规则
⚠️ 请严格遵守项目技术规范第2节"环境保护规则"。测试数据库使用独立文件，不要修改主数据库。

### ⚠️ 重要约束：不允许直接修改代码
集成测试 Agent **严禁直接修改任何代码文件**。发现问题时，只能通过反馈机制让对应的开发 Agent 进行修改。

### 测试准备
1. 读取需求文档和接口规范
2. **阅读已知陷阱清单 (docs/known_traps.md)，了解历史踩坑点**
3. 在 tests/integration/ 目录下编写**端到端（E2E）集成测试**用例
4. **分批编写执行**：每批 ≤ 5 个用例，以 API 规范文档为准编写测试

### 测试执行
1. 等待后端开发完成（go build 和 go test 通过）
2. 使用 httptest 在测试代码中直接启动 HTTP server
3. 按顺序执行测试用例（先注册登录，再业务操作）
4. 记录每个用例的通过/失败状态

### 失败处理
1. 测试不通过时，定位问题归属（后端哪个模块）
2. 生成失败报告：用例编号 + 期望结果 + 实际结果 + 错误信息
3. 反馈给 @dev_backend_agent 进行修复（**不要自己修改代码**）
4. 修复后重新执行失败的用例
5. 最多3轮修复循环
```

### 2.3 启动前端开发 Agent

```
@dev_frontent_agent

## 任务：第{ITERATION}迭代前端开发

### 输入文档
- 需求文档: docs/iterations/v{VERSION}/iteration{ITERATION}_requirements.md
- 接口规范: docs/iterations/v{VERSION}/iteration{ITERATION}_api_spec.md
- 项目配置: docs/project_config.yaml
- 项目技术规范: skills/project-tech-spec.md
- 通用核心Skill: ../../skills/shared/iteration-dev/core.md

### 环境保护规则
⚠️ 请严格遵守项目技术规范第2节"环境保护规则"。

### 模块划分
{FRONTEND_MODULES}

### 前端测试技术栈
| 测试层 | 小程序端 (Taro + React) | H5端 (Vue3 + Vite) |
|--------|------------------------|--------------------|
| **单元测试** | Jest | Vitest |
| **组件测试** | Jest + @testing-library/react + MSW | Vitest + @vue/test-utils + MSW |
| **E2E冒烟测试** | Jest + miniprogram-automator | Playwright |
| **API Mock** | MSW (Mock Service Worker) | MSW |

### 开发规则
1. **每个模块必须拉起独立的 sub agent，严禁合并多个模块到一个 sub agent 中**（通用Skill R4）
2. 每个模块需要2个独立的 sub agent：1个负责开发代码，1个负责编写并执行单元测试 + 组件测试 + 代码Review（通用Skill R1 + R20）
3. 同层级的开发 sub agent 尽量并行拉起，开发完成后再并行拉起对应的单测&Review sub agent
4. **sub agent 的 prompt 只包含结论和指令，不包含编排者的推理过程**（通用Skill R5）
5. 单测/组件测试不通过或Review发现问题时，反馈给开发 sub agent 修复，最多3轮
6. 所有模块完成后，执行整体构建检查
```

---

## 3. 已验证的项目特定最佳实践

| 实践 | 说明 |
|------|------|
| httptest 集成测试 | 避免端口占用、进程管理等复杂问题，测试自包含可重复 |
| flattenConfig 适配层 | 当配置结构（嵌套）和消费方接口（扁平 map）不匹配时，用 flatten 适配 |
| 插件 Data merge | 每个插件的 Execute 方法必须 merge 上游 Data，否则上游数据丢失 |

---

**文档版本**: v5.0.0
**更新日期**: 2026-04-06
**变更记录**:
- v5.0.0: 项目配置剥离 — 项目概述迁移到 `docs/project_config.yaml`；已知陷阱迁移到 `docs/known_traps.md`；模块划分已在 v4.0 迁移到 `iteration{N}_modules.yaml`；执行模板支持动态占位符替换
