# 📚 文档中心

## 文档目录结构

```
docs/
├── README.md                          # 📌 本文件 - 文档索引
├── design/                            # 🏗️ 总体设计（跨版本）
│   ├── requirements.md                #   总需求规格说明书
│   ├── architecture.md                #   Harness 架构设计
│   └── development_plan.md            #   开发计划总览
├── iterations/                        # 🔄 迭代文档（按版本）
│   ├── v1.0/                          #   V1.0 - MVP 核心功能
│   │   ├── iteration1_requirements.md #     Sprint1 需求规格
│   │   ├── iteration1_api_spec.md     #     Sprint1 API 接口规范
│   │   └── task_tracking.md           #     任务跟踪
│   ├── v2.0/                          #   V2.0 - 增强功能（待规划）
│   └── v3.0/                          #   V3.0 - 扩展功能（待规划）
└── guides/                            # 📖 开发指南
    └── go_environment_setup.md        #   Go 环境搭建指南
```

## 🏗️ 总体设计

| 文档 | 说明 | 状态 |
|------|------|------|
| [requirements.md](design/requirements.md) | 总需求规格说明书，定义全部功能需求和非功能需求 | ✅ 已完成 |
| [architecture.md](design/architecture.md) | Harness 插件化架构设计，核心接口和模式定义 | ✅ 已完成 |
| [development_plan.md](design/development_plan.md) | 开发计划总览，版本规划和里程碑 | ✅ 已完成 |

## 🔄 迭代文档

### V1.0 - MVP 核心功能（2026-Q2）

> 目标：Harness 框架 + 4 个核心插件 + 全链路跑通

| 文档 | 说明 | 状态 |
|------|------|------|
| [iteration1_requirements.md](iterations/v1.0/iteration1_requirements.md) | Sprint1 后端核心框架需求，8 个模块拆解 | ✅ 已完成 |
| [iteration1_api_spec.md](iterations/v1.0/iteration1_api_spec.md) | Sprint1 API 接口规范，12 个端点详细定义 | ✅ 已完成 |
| [task_tracking.md](iterations/v1.0/task_tracking.md) | 任务跟踪和进度管理 | 🔄 进行中 |

### V2.0 - 增强功能（2026-Q3）

> 目标：记忆衰减 + 数据分析 + 留言系统 + 导出功能

| 文档 | 说明 | 状态 |
|------|------|------|
| 待规划 | - | ⏳ 待开始 |

### V3.0 - 扩展功能（2026-Q4）

> 目标：多租户 + 插件市场 + API 开放 + 多端适配

| 文档 | 说明 | 状态 |
|------|------|------|
| 待规划 | - | ⏳ 待开始 |

## 📖 开发指南

| 文档 | 说明 | 状态 |
|------|------|------|
| [go_environment_setup.md](guides/go_environment_setup.md) | Go 开发环境搭建（Go 1.21 + 依赖配置） | ✅ 已完成 |
