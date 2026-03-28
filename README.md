# 🎓 数字分身教育平台 (Digital Twin)

> 为教育行业提供 AI 数字分身解决方案，让教师创建智能助教，为学生提供 7x24 小时个性化苏格拉底式教学服务。

## 🛠️ 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| 后端框架 | Go + Gin | 高性能 HTTP 服务 |
| 架构模式 | Harness 插件化 | 配置驱动、管道编排 |
| 关系数据库 | SQLite | 轻量级，V2.0 可扩展至 MySQL |
| 向量数据库 | Chroma DB | 语义检索 |
| AI 模型 | OpenAI GPT 系列 | 对话生成 + 文本向量化 |
| 前端框架 | Taro + Vant Weapp | 跨端小程序（V1.0 后期） |

## 📁 项目目录结构

```
digital-twin/
├── README.md                    # 📌 本文件 - 项目说明
├── go.mod                       # Go 模块定义
├── configs/                     # ⚙️ 配置文件
│   ├── harness.yaml             #   Harness 主配置（插件+管道+系统）
│   └── environments/            #   环境覆盖配置（dev/test/prod）
├── docs/                        # 📚 项目文档
│   ├── design/                  #   总体设计（跨版本）
│   │   ├── requirements.md      #     总需求规格说明书
│   │   ├── architecture.md      #     Harness 架构设计
│   │   └── development_plan.md  #     开发计划总览
│   ├── iterations/              #   迭代文档（按版本）
│   │   ├── v1.0/                #     V1.0 MVP 核心功能
│   │   ├── v2.0/                #     V2.0 增强功能
│   │   └── v3.0/                #     V3.0 扩展功能
│   └── guides/                  #   开发指南
│       └── go_environment_setup.md
├── src/                         # 💻 源代码
│   ├── cmd/                     #   程序入口
│   │   └── server/              #     HTTP 服务入口 main.go
│   ├── harness/                 #   🔧 Harness 核心框架
│   │   ├── core/                #     核心接口定义（Plugin/Pipeline）
│   │   ├── config/              #     配置管理（YAML 解析）
│   │   └── manager/             #     管理器（插件注册/管道编排）
│   ├── plugins/                 #   🔌 业务插件
│   │   ├── auth/                #     认证授权插件
│   │   ├── knowledge/           #     知识库检索插件
│   │   ├── memory/              #     记忆管理插件
│   │   └── dialogue/            #     苏格拉底对话插件
│   ├── backend/                 #   🌐 后端服务
│   │   ├── api/                 #     HTTP 路由 + Handler
│   │   └── database/            #     数据库连接 + Repository
│   └── frontend/                #   📱 前端小程序（V1.0 后期）
├── deployments/                 #   🚀 部署配置
│   ├── docker/                  #     Docker 相关
│   └── scripts/                 #     部署脚本
├── tests/                       #   🧪 测试
│   ├── integration/             #     集成测试
│   └── fixtures/                #     测试数据
└── data/                        #   💾 运行时数据（.gitignore）
    ├── digital-twin.db          #     SQLite 数据库文件
    └── uploads/                 #     上传文件存储
```

## 🚀 版本规划

| 版本 | 周期 | 目标 | 状态 |
|------|------|------|------|
| **V1.0** | 2026-Q2 (4-6月) | MVP：Harness 框架 + 4 核心插件 + 全链路对话 | 🔄 开发中 |
| **V2.0** | 2026-Q3 (7-8月) | 增强：记忆衰减 + 数据分析 + 留言 + 导出 | ⏳ 待规划 |
| **V3.0** | 2026-Q4 (9-12月) | 扩展：多租户 + 插件市场 + API 开放 + 多端 | ⏳ 待规划 |

### V1.0 迭代计划

| 迭代 | 周期 | 内容 | 状态 |
|------|------|------|------|
| Sprint 1 | 2周 | 后端核心框架 + 全链路验证 | 🔄 进行中 |
| Sprint 2 | 2周 | 前端基础界面 + 前后端联调 | ⏳ 待开始 |
| Sprint 3 | 2周 | 功能完善 + 测试 + 部署 | ⏳ 待开始 |

## 🏃 快速开始

```bash
# 1. 环境准备
export JWT_SECRET="your-secret-key-at-least-32-chars"
export OPENAI_API_KEY="sk-your-api-key"

# 2. 安装依赖
go mod tidy

# 3. 启动 Chroma DB
docker run -d -p 8000:8000 chromadb/chroma:latest

# 4. 启动服务
go run src/cmd/server/main.go

# 5. 测试
curl http://localhost:8080/api/system/health
```

## 📖 文档导航

- **总需求**：[docs/design/requirements.md](docs/design/requirements.md)
- **架构设计**：[docs/design/architecture.md](docs/design/architecture.md)
- **V1.0 Sprint1 需求**：[docs/iterations/v1.0/iteration1_requirements.md](docs/iterations/v1.0/iteration1_requirements.md)
- **V1.0 Sprint1 API 规范**：[docs/iterations/v1.0/iteration1_api_spec.md](docs/iterations/v1.0/iteration1_api_spec.md)
- **Go 环境搭建**：[docs/guides/go_environment_setup.md](docs/guides/go_environment_setup.md)