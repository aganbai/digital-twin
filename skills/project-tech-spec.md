# Digital-Twin 项目技术规范

> 本文档包含 digital-twin 项目特定的技术注意事项、环境配置和编码规范。
> 通用的迭代开发流程请参考：`../../skills/shared/iteration-dev/core.md`（核心规则）

---

## 1. 环境准备（项目特定）

### 1.1 环境检查清单

| 检查项 | 验证命令 | 期望结果 |
|--------|----------|----------|
| Go 编译器 | `go version` | go 1.22+ |
| SQLite 可用 | `go build ./...`（含 modernc.org/sqlite） | 编译通过 |
| Go 依赖完整 | `go mod tidy && go build ./...` | 无缺失依赖 |
| 环境变量配置 | `cat .env` | JWT_SECRET、LLM相关变量已设置 |
| LLM API 可用 | `curl` 测试 LLM 接口 | 返回正常响应 |
| 端口可用 | `lsof -i :8080` | 端口未被占用（或可配置其他端口） |
| data 目录 | `ls -la data/` | 目录存在且可写 |
| 磁盘空间 | `df -h .` | 剩余空间 > 1GB |

### 1.2 依赖统一初始化

```bash
cd /path/to/digital-twin
go get github.com/gin-gonic/gin
go get gopkg.in/yaml.v3
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
go get modernc.org/sqlite
go get github.com/google/uuid
go mod tidy
```

### 1.3 目录结构预创建

```bash
mkdir -p src/harness/config
mkdir -p src/backend/database
mkdir -p src/plugins/auth
mkdir -p src/plugins/knowledge
mkdir -p src/plugins/memory
mkdir -p src/plugins/dialogue
mkdir -p src/backend/api
mkdir -p src/cmd/server
mkdir -p tests/integration
mkdir -p data
```

### 1.4 环境快照

```bash
cp go.mod go.mod.snapshot
cp go.sum go.sum.snapshot
```

---

## 2. 环境保护规则（项目特定）

### 2.1 禁止的破坏性操作

| ❌ 禁止操作 | 说明 |
|-------------|------|
| 修改 `.env` 中的 API Key | 环境变量由主 Agent 统一管理 |
| 修改 `go.mod` 的 module 名称 | 模块路径 `digital-twin` 不可变 |
| 删除 `configs/` 目录下的配置文件 | 配置文件只允许读取，不允许删除 |
| 修改 `src/harness/core/plugin.go` | 核心接口定义已冻结，不可修改 |
| 删除或覆盖 `data/` 目录下的数据库文件 | 测试数据可追加，不可删除 |
| 修改端口号（除非端口冲突） | 默认 8080，所有 Agent 统一使用 |
| 执行 `go mod init` | 模块已初始化，禁止重新初始化 |

### 2.2 环境恢复机制

```bash
# 恢复 go.mod
cp go.mod.snapshot go.mod
cp go.sum.snapshot go.sum
go mod tidy

# 清理测试数据
rm -f data/test_*.db

# 重新验证环境
go build ./...
```

---

## 3. 技术注意事项（所有 Agent 必读）

### 3.1 插件 Data 累积规则

**关键**：当前 `pipeline.go` 的 Execute 逻辑中，每个插件的输出 Data **完全替换**下一个插件的输入 Data：

```go
currentInput = &core.PluginInput{
    Data: output.Data,  // 完全用输出替换
}
```

因此，**每个插件的 Execute 方法必须将上游传入的 Data 字段 merge 到自己的输出 Data 中**，否则上游数据会丢失。

示例：
```go
func (p *AuthPlugin) Execute(input *core.PluginInput) (*core.PluginOutput, error) {
    outputData := make(map[string]interface{})
    for k, v := range input.Data {
        outputData[k] = v  // 保留上游所有字段
    }
    outputData["user_id"] = userID      // 添加本插件的输出
    outputData["user_role"] = userRole
    
    return &core.PluginOutput{Data: outputData}, nil
}
```

### 3.2 UserContext 传递修复

当前 `pipeline.go` 中 `UserContext` 始终使用初始值，auth 插件填充的 UserContext 不会传递给下游。**必须修复此问题**：

```go
// 修复前（当前代码）
currentInput = &core.PluginInput{
    UserContext: input.UserContext,  // 始终用原始的
}

// 修复后
currentInput = &core.PluginInput{
    UserContext: output.UserContext,  // 使用插件更新后的
    Data:        output.Data,
}
```

### 3.3 配置文件环境变量替换

`harness.yaml` 中大量使用 `${VAR_NAME:-default}` 语法，但 Go 的 `yaml.v3` **不会自动做环境变量替换**。

**配置管理模块必须实现环境变量替换逻辑**：
```go
// 解析 ${VAR_NAME:-default} 格式
// 1. 先读取 yaml 文件为字符串
// 2. 用正则替换 ${...} 为环境变量值
// 3. 再用 yaml.Unmarshal 解析
```

### 3.4 内存向量存储

第一迭代不依赖外部 Chroma DB 服务。**知识库插件的 `chroma_client.go` 应实现 InMemoryVectorStore**：
- 使用简单的关键词匹配或 TF-IDF 代替真正的向量检索
- 不需要 Embedding API 调用
- 数据存储在内存 map 中，服务重启后丢失（第一迭代可接受）

### 3.5 数据库目录自动创建

**数据库层必须在初始化时自动创建 `data/` 目录**：
```go
os.MkdirAll("data", 0755)
```

### 3.6 集成测试方式

**推荐使用 Go 的 `httptest` 包**在测试代码中直接启动 HTTP server：
```go
router := api.SetupRouter(...)
ts := httptest.NewServer(router)
defer ts.Close()
// 用 ts.URL 作为 base URL 发送请求
```

---

## 4. 技术栈约束

| 技术 | 选型 | 说明 |
|------|------|------|
| 语言 | Go 1.22+ | |
| HTTP 框架 | github.com/gin-gonic/gin | |
| 数据库 | modernc.org/sqlite | 纯Go，无需CGO |
| JWT | github.com/golang-jwt/jwt/v5 | |
| YAML | gopkg.in/yaml.v3 | |
| 密码加密 | golang.org/x/crypto/bcrypt | |
| UUID | github.com/google/uuid | |
| 大模型 | OpenAI兼容格式（通义千问 qwen-turbo） | |

---

## 5. 编码规范（项目特定）

### 5.1 并发安全

- 共享资源（如 LLMClient、数据库连接）必须考虑并发安全
- 使用 `sync.Mutex` 或 `sync.RWMutex` 保护共享状态
- 避免在 goroutine 中直接操作未加锁的共享变量

### 5.2 避免硬编码

- 配置值（如默认温度、超时时间）应从配置文件读取，不应硬编码在代码中
- 如需默认值，应定义为常量并在注释中说明来源

### 5.3 路径配置化

- 文件上传目录、数据库路径等应从配置文件读取
- 禁止使用相对路径作为默认值，应使用绝对路径或基于项目根目录的路径

### 5.4 单文件大小控制

- 单个代码文件不应超过 **500 行**
- 超过 500 行时应按职责拆分为多个文件
- 例如：`handlers.go` 应按资源类型拆分为 `auth_handlers.go`、`document_handlers.go`、`chat_handlers.go` 等

### 5.5 API 安全防护

- 所有公开 API 应考虑限流中间件
- 敏感操作（如登录、注册）应有频率限制
- 错误信息不应暴露内部实现细节

### 5.6 错误处理规范

- 所有插件对**未知或无效的 action** 必须返回明确的错误信息，禁止静默忽略
- 错误信息应包含：错误码 + 错误描述 + 建议操作（如适用）
- 插件 Execute 方法中，所有可能失败的操作（数据库查询、外部 API 调用、文件读写等）都必须检查并处理错误
- 禁止使用 `_ = someFunc()` 忽略错误返回值（除非有明确注释说明原因）
- 错误应逐层向上传递，不应在中间层被吞掉；如需包装错误，使用 `fmt.Errorf("context: %w", err)` 保留错误链
- 对外暴露的 API 错误响应必须使用统一的错误格式：

```go
// 统一错误响应格式
type ErrorResponse struct {
    Code    int    `json:"code"`    // 业务错误码
    Message string `json:"message"` // 用户可见的错误描述
}
```

---

## 6. 架构改进建议（待实施）

### S1: 管道（Pipeline）的 action 路由机制需要重新设计

**问题**：管道中所有插件共享同一个 `Data["action"]`，但不同插件需要不同的 action。

**改进方向**（选择其一）：
- **方案A**：管道配置中为每个插件指定独立的 action 参数
- **方案B**：插件在管道模式下不依赖 action 字段，根据自身类型自动判断行为
- **方案C**：在 `PluginInput` 中增加 `Params map[string]interface{}`（每个插件的独立参数），与 `Data`（流转数据）分离

### S2: goroutine 退出机制

**问题**：`harness_manager.go` 中的 `startHealthCheck()` goroutine 没有退出机制，会导致 goroutine 泄漏。

**改进方向**：添加 `stopCh chan struct{}` 或使用 `context.WithCancel`，在 `Stop()` 中关闭。

### S3: 公共工具函数抽取

**问题**：`toInt`、`toInt64`、`toFloat64` 等类型转换函数在多个文件中重复定义。

**改进方向**：抽取到 `src/common/utils/convert.go` 公共包中，所有模块统一引用。

---

**文档版本**: v1.1.0
**创建日期**: 2026-03-29
**适用项目**: digital-twin
**关联通用Skill**: `../../skills/shared/iteration-dev/core.md`（核心规则）+ `reference.md`（详细参考）
