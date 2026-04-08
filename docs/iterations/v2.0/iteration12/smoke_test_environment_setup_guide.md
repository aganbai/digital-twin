# V2.0 迭代12 冒烟测试环境配置指南

**生成时间**: 2026-04-08  
**项目版本**: V2.0 迭代12  
**测试类型**: Phase 3c 端到端冒烟测试

## 1. 当前环境状态

### 1.1 阻塞性问题

| 问题项 | 状态 | 影响 | 解决方案 |
|--------|------|------|----------|
| 微信开发者工具 CLI | ❌ 不可用 | 无法启动小程序 | 安装并配置微信开发者工具 |
| 后端服务 | ❌ 未启动 | API 调用失败 | 启动后端服务 (端口 8080) |
| 后端启动入口 | ✅ 已找到 | - | 使用 `src/cmd/server/main.go` |

### 1.2 环境依赖状态

| 依赖项 | 状态 | 版本 | 检查命令 |
|--------|------|------|----------|
| 微信开发者工具 | ❌ 未安装 | - | `cli --version` |
| Go 后端服务 | ❌ 未启动 | 1.21+ | `go run src/cmd/server/main.go` |
| 测试框架 | ✅ 已安装 | miniprogram-automator@0.12.1 | `npm list miniprogram-automator` |
| 前端依赖 | ✅ 完整 | Taro 3.6.31 | `npm list` |

## 2. 环境配置步骤

### 2.1 微信开发者工具配置

#### 步骤 1: 安装微信开发者工具
```bash
# 下载微信开发者工具
# 官方下载地址: https://developers.weixin.qq.com/miniprogram/dev/devtools/download.html

# macOS 安装后验证
open /Applications/wechatwebdevtools.app
```

#### 步骤 2: 开启 CLI 功能
1. 打开微信开发者工具
2. 进入设置 → 安全设置
3. 开启 "服务端口" 和 "命令行调用" 功能
4. 设置服务端口号（默认 9420）

#### 步骤 3: 验证 CLI 可用性
```bash
# 检查 CLI 是否可用
cli --version

# 如果命令不存在，可能需要配置 PATH
# macOS 默认路径: /Applications/wechatwebdevtools.app/Contents/MacOS/cli
```

### 2.2 后端服务启动

#### 步骤 1: 准备环境变量
```bash
# 复制环境变量模板
cp .env.example .env

# 编辑 .env 文件，配置必要的环境变量
# 至少需要配置 JWT_SECRET
JWT_SECRET="your-32-character-jwt-secret-key-here"
SERVER_PORT="8080"
LLM_MODE="mock"  # 测试环境使用 mock 模式
```

#### 步骤 2: 启动后端服务
```bash
# 进入项目根目录
cd /Users/aganbai/Desktop/WorkSpace/digital-twin

# 启动后端服务
go run src/cmd/server/main.go

# 验证服务状态
curl http://localhost:8080/api/system/health
```

#### 步骤 3: 验证服务健康状态
期望响应:
```json
{
  "status": "healthy",
  "timestamp": "2026-04-08T10:00:00Z",
  "version": "1.1.0"
}
```

### 2.3 前端小程序构建

#### 步骤 1: 构建小程序
```bash
# 进入前端目录
cd src/frontend

# 安装依赖（如果未安装）
npm install

# 构建小程序到 dist 目录
npm run build:weapp
```

#### 步骤 2: 导入到微信开发者工具
1. 打开微信开发者工具
2. 导入项目，选择 `src/frontend/dist` 目录
3. 设置 AppID（可使用测试号）
4. 编译并预览

## 3. 冒烟测试执行流程

### 3.1 环境就绪检查清单

在执行冒烟测试前，请确认以下条件：

- [ ] 微信开发者工具已安装并开启 CLI 功能
- [ ] 后端服务在 localhost:8080 正常运行
- [ ] 小程序已构建并导入开发者工具
- [ ] 测试框架 miniprogram-automator 可用

### 3.2 测试执行顺序

```yaml
执行阶段:
  1. 启动后端服务并验证健康状态
  2. 启动微信开发者工具并连接
  3. 执行核心功能测试 (Part A: 新用户引导流程)
  4. 执行老用户功能测试 (Part B: 核心操作)
  5. 执行异常场景测试 (Part C: 边界条件)
```

### 3.3 测试用例优先级

**P0 核心功能** (必须通过):
- SMOKE-A-001: 新用户首次进入聊天页测试
- SMOKE-A-002: 新用户会话列表功能测试  
- SMOKE-B-001: 老用户会话切换功能测试
- SMOKE-B-002: 老用户流式中断压力测试

## 4. 故障排除

### 4.1 常见问题

#### 微信开发者工具连接失败
```bash
# 检查服务端口是否开启
netstat -an | grep 9420

# 重启开发者工具服务
cli -r
```

#### 后端服务启动失败
```bash
# 检查端口占用
lsof -i :8080

# 检查 Go 依赖
go mod tidy
go mod download
```

#### 小程序编译错误
```bash
# 清理缓存并重新构建
npm run clean
npm run build:weapp
```

### 4.2 日志检查

#### 后端服务日志
```bash
# 查看实时日志
tail -f logs/server.log

# 检查错误日志
grep "ERROR" logs/server.log
```

#### 小程序日志
- 在微信开发者工具中查看 Console 面板
- 检查 Network 面板的 API 调用状态

## 5. 测试报告生成

### 5.1 测试产物

环境配置完成后，冒烟测试将生成：

1. **测试报告**: `docs/iterations/v2.0/iteration12/smoke_report.md`
2. **测试截图**: `docs/iterations/v2.0/iteration12/screenshots/`
3. **执行日志**: `docs/iterations/v2.0/iteration12/test_logs/`

### 5.2 报告内容

- 执行概要（总用例数、通过数、失败数）
- 环境信息（框架版本、服务状态）
- 用例详情（每个用例的执行结果）
- 失败分析（错误信息和修复建议）
- 性能指标（响应时间、内存使用）

## 6. 下一步行动

### 立即行动
1. 安装并配置微信开发者工具
2. 启动后端服务 (端口 8080)
3. 构建小程序并导入开发者工具

### 测试执行
环境配置完成后，重新执行 Phase 3c 冒烟测试：
```bash
# 执行冒烟测试脚本
npm run test:smoke
```

---

**文档版本**: 1.0  
**最后更新**: 2026-04-08  
**维护者**: Claude Code Agent