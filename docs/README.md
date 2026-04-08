# 📚 文档中心

## 文档目录结构

```
docs/
├── README.md                              # 📌 本文件 - 文档索引
├── project_config.yaml                    # ⚙️ 项目配置（由 dt-orchest 加载）
├── known_traps.md                         # ⚠️ 已知陷阱清单
├── smoke-test-cases.md                    # 🧪 冒烟测试用例全集
│
├── design/                                # 🏗️ 总体设计（跨版本）
│   ├── requirements.md                    #   总需求规格说明书
│   ├── architecture.md                    #   Harness 架构设计
│   └── development_plan.md                #   开发计划总览
│
├── iterations/                            # 🔄 迭代文档（按版本）
│   ├── v1.0/                              #   V1.0 - MVP 核心功能 ✅
│   ├── v2.0/                              #   V2.0 - 单机生产可用版 📋
│   └── v3.0/                              #   V3.0 - 扩展功能（待规划）
│
├── testing/                               # 🧪 测试相关
│   ├── reports/                           #   测试报告归档
│   └── test_data_guide.md                 #   测试数据准备指南
│
├── guides/                                # 📖 开发指南
│   └── go_environment_setup.md            #   Go 环境搭建指南
│
└── archive/                               # 📦 归档文档
    └── iteration-modules-archive.md       #   模块划分历史归档
```

## 🏗️ 总体设计

| 文档 | 说明 | 状态 |
|------|------|------|
| [requirements.md](design/requirements.md) | 总需求规格说明书，定义全部功能需求和非功能需求 | ✅ 已完成 |
| [architecture.md](design/architecture.md) | Harness 插件化架构设计，核心接口和模式定义 | ✅ 已完成 |
| [development_plan.md](design/development_plan.md) | 开发计划总览，版本规划和里程碑 | ✅ 已完成 |

## 🔄 迭代文档

### V1.0 - MVP 核心功能（✅ 已完成）

> 3 个迭代，39 个集成测试全部通过

| 文档 | 说明 | 状态 |
|------|------|------|
| [iteration1_requirements.md](iterations/v1.0/iteration1_requirements.md) | 迭代1：后端核心框架 & 全链路验证（8 个模块） | ✅ |
| [iteration1_api_spec.md](iterations/v1.0/iteration1_api_spec.md) | 迭代1：API 接口规范（12 个端点） | ✅ |
| [iteration2_requirements.md](iterations/v1.0/iteration2_requirements.md) | 迭代2：小程序前端开发 & 后端适配改造 | ✅ |
| [iteration2_api_spec.md](iterations/v1.0/iteration2_api_spec.md) | 迭代2：API 接口规范（5 个新增端点） | ✅ |
| [iteration3_requirements.md](iterations/v1.0/iteration3_requirements.md) | 迭代3：架构补全 & 质量加固（6 个模块） | ✅ |
| [iteration3_api_spec.md](iterations/v1.0/iteration3_api_spec.md) | 迭代3：API 接口变更规范 | ✅ |
| [task_tracking.md](iterations/v1.0/task_tracking.md) | V1.0 总任务跟踪 | ✅ |

### V2.0 - 单机生产可用版（📋 规划中）

> 2 个迭代：迭代1 核心功能开发 + 迭代2 生产就绪与上线

| 文档 | 说明 | 状态 |
|------|------|------|
| [requirements.md](iterations/v2.0/requirements.md) | V2.0 需求规格说明书（2 个迭代） | ✅ 已完成 |
| [iteration1_requirements.md](iterations/v2.0/iteration1_requirements.md) | 迭代1：核心功能开发需求（10 个后端模块 + 8 个前端模块） | ✅ 已完成 |
| [iteration1_api_spec.md](iterations/v2.0/iteration1_api_spec.md) | 迭代1：API 接口规范（17 个新增 + 2 个改造 + 1 个增强） | ✅ 已完成 |
| [task_tracking.md](iterations/v2.0/task_tracking.md) | V2.0 任务跟踪 | 📋 待开发 |

### V3.0 - 扩展功能（待规划）

> 多租户 + 插件市场 + API 开放 + 智能推荐

| 文档 | 说明 | 状态 |
|------|------|------|
| [backlog.md](iterations/v3.0/backlog.md) | 需求备忘（URL 导入已合并到 V2.0） | 📋 |

## 📖 开发指南

| 文档 | 说明 | 状态 |
|------|------|------|
| [go_environment_setup.md](guides/go_environment_setup.md) | Go 开发环境搭建（Go 1.25 + 依赖配置） | ✅ 已完成 |

## 🧪 测试相关

| 文档 | 说明 | 状态 |
|------|------|------|
| [smoke-test-cases.md](smoke-test-cases.md) | 冒烟测试用例全集（124 条） | ✅ 已完成 |
| [known_traps.md](known_traps.md) | 已知陷阱清单（17 条） | ✅ 已完成 |
| [test_data_guide.md](testing/test_data_guide.md) | 测试数据准备指南 | ✅ 已完成 |

### 测试报告归档

| 报告 | 说明 |
|------|------|
| [smoke_test_report_phase3c.md](testing/reports/smoke_test_report_phase3c.md) | Phase 3C 冒烟测试报告 |
| [smoke-test-report-20260405.md](testing/reports/smoke-test-report-20260405.md) | 2026-04-05 冒烟测试报告 |
| [smoke-test-report-v3.2.md](testing/reports/smoke-test-report-v3.2.md) | V3.2 冒烟测试报告 |
| [smoke-test-report-v4.2.md](testing/reports/smoke-test-report-v4.2.md) | V4.2 冒烟测试报告 |
| [smoke-test-report-v4.3.md](testing/reports/smoke-test-report-v4.3.md) | V4.3 冒烟测试报告 |
| [iteration9-smoke-test-report-final.md](testing/reports/iteration9-smoke-test-report-final.md) | 迭代9 最终冒烟测试报告 |

## ⚙️ 项目配置

| 文档 | 说明 |
|------|------|
| [project_config.yaml](project_config.yaml) | 项目配置文件（由 dt-orchest 动态加载） |

## 📦 归档文档

| 文档 | 说明 |
|------|------|
| [iteration-modules-archive.md](archive/iteration-modules-archive.md) | 模块划分历史归档 |
