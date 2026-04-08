# 沙盒一键部署指南

## 快速开始

### 方式一：使用 Shell 脚本（推荐）

```bash
# 完整部署流程
./deployments/scripts/deploy-sandbox.sh --host 192.168.1.100 full

# 仅构建镜像（本地）
./deployments/scripts/deploy-sandbox.sh build

# 仅上传到沙盒
./deployments/scripts/deploy-sandbox.sh --host 192.168.1.100 upload

# 仅重启服务
./deployments/scripts/deploy-sandbox.sh --host 192.168.1.100 restart

# 查看服务状态
./deployments/scripts/deploy-sandbox.sh --host 192.168.1.100 status

# 查看服务日志
./deployments/scripts/deploy-sandbox.sh --host 192.168.1.100 logs
```

### 方式二：使用 Python 模块（适合编排任务）

```bash
# 完整部署流程
python3 deployments/scripts/deploy_sandbox.py full --host 192.168.1.100

# 仅构建镜像
python3 deployments/scripts/deploy_sandbox.py build

# 仅重启服务
python3 deployments/scripts/deploy_sandbox.py restart --host 192.168.1.100
```

### 方式三：在编排任务中调用

```python
from pathlib import Path
from deployments.scripts.deploy_sandbox import SandboxDeployer, SandboxConfig

# 创建配置
config = SandboxConfig(
    host="192.168.1.100",
    user="root",
    port=22,
    deploy_dir="/opt/digital-twin",
    skip_tests=False
)

# 创建部署器
deployer = SandboxDeployer(
    project_root=Path("/Users/aganbai/Desktop/WorkSpace/digital-twin"),
    config=config
)

# 执行完整部署
success = deployer.deploy_full()
if success:
    print("部署成功！")
else:
    print("部署失败！")
```

## 编排任务集成

### Phase 3b 集成测试阶段

在 `ci_test_agent` 的集成测试阶段，可以调用部署脚本：

```python
# 示例：集成测试前的部署步骤
def deploy_to_sandbox():
    config = SandboxConfig(
        host=os.getenv("SANDBOX_HOST"),
        user=os.getenv("SANDBOX_USER", "root"),
        deploy_dir=os.getenv("SANDBOX_DEPLOY_DIR", "/opt/digital-twin")
    )
    
    deployer = SandboxDeployer(project_root, config)
    
    # 构建镜像
    if not deployer.build_images():
        raise Exception("构建失败")
    
    # 上传到沙盒
    if not deployer.upload_to_sandbox():
        raise Exception("上传失败")
    
    # 重启服务
    if not deployer.restart_services():
        raise Exception("服务启动失败")
    
    return True
```

### 环境变量配置

在编排任务的环境变量中配置以下参数：

```bash
# 沙盒服务器配置
SANDBOX_HOST=192.168.1.100
SANDBOX_USER=root
SANDBOX_PORT=22
SANDBOX_DEPLOY_DIR=/opt/digital-twin

# 镜像仓库（可选）
IMAGE_REGISTRY=registry.example.com/digital-twin

# 测试配置
SKIP_TESTS=false
```

## 部署流程详解

### 1. 构建阶段（build）

- 检查 Docker 环境
- 检查环境变量配置
- 构建 Backend 镜像
- 构建 Knowledge 镜像
- 拉取 Nginx 镜像
- 运行单元测试（可选）

### 2. 上传阶段（upload）

- 测试 SSH 连接
- 创建远程部署目录
- 上传配置文件（docker-compose.yml, .env.production, nginx.conf）
- 导出镜像为 tar 文件
- 上传镜像到沙盒服务器
- 在沙盒服务器加载镜像

### 3. 重启阶段（restart）

- 停止旧服务
- 启动新服务
- 等待服务就绪
- 执行健康检查
- 确认所有服务正常

### 4. 健康检查

检查以下端点：
- Backend: `http://localhost:8080/api/system/health`
- Knowledge: `http://localhost:8100/api/v1/health`
- Nginx: `http://localhost:80/health`

## 返回码说明

| 返回码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1 | 参数错误 |
| 2 | 构建失败 |
| 3 | 上传失败 |
| 4 | 服务启动失败 |
| 5 | 健康检查失败 |

## 故障排查

### 查看日志

```bash
# 查看部署日志
tail -f /tmp/deploy-sandbox-*.log

# 查看沙盒服务日志
./deploy-sandbox.sh --host 192.168.1.100 logs
```

### 常见问题

1. **SSH 连接失败**
   - 检查服务器地址和端口
   - 确认 SSH 密钥已配置
   - 验证用户权限

2. **镜像构建失败**
   - 检查 Dockerfile 语法
   - 确认依赖包可访问
   - 查看构建日志

3. **服务启动失败**
   - 检查端口占用
   - 查看容器日志
   - 验证环境变量配置

4. **健康检查失败**
   - 等待服务完全启动
   - 检查服务依赖
   - 查看服务日志

## 冒烟环境检查

在执行冒烟测试前，需要检查测试环境是否就绪。

### 使用 Shell 脚本

```bash
# 执行环境检查
./deployments/scripts/check-smoke-env.sh

# 自动修复可修复的问题
./deployments/scripts/check-smoke-env.sh --fix

# 显示详细输出
./deployments/scripts/check-smoke-env.sh --verbose
```

### 使用 Python 模块

```bash
# 执行环境检查
python3 deployments/scripts/check_smoke_env.py

# 自动修复
python3 deployments/scripts/check_smoke_env.py --fix

# 输出 JSON 格式报告
python3 deployments/scripts/check_smoke_env.py --json
```

### 在编排任务中调用

```python
from pathlib import Path
from deployments.scripts.check_smoke_env import SmokeEnvChecker

# 创建检查器
checker = SmokeEnvChecker(
    project_root=Path("/path/to/digital-twin"),
    auto_fix=True
)

# 执行检查
if checker.run_all_checks():
    print("环境检查通过，可以执行冒烟测试")
else:
    print("环境检查失败，请先修复问题")
    # 获取详细报告
    report = checker.generate_report()
    print(report)
```

### 检查项目

环境检查脚本会检查以下内容：

#### 1. Python 环境
- Python 3 版本（>= 3.8）
- pip 包管理器

#### 2. Minium 环境
- minium 包安装状态
- minium 依赖包（requests, pytest, allure-pytest）
- Minium 导入测试

#### 3. Playwright 环境
- playwright 包安装状态
- 浏览器安装状态（chromium, firefox, webkit）
- 系统 Chrome 浏览器（可选）

#### 4. 微信开发者工具
- 开发者工具安装状态
- CLI 工具可用性
- 服务端口开启状态
- 小程序编译产物

#### 5. Node.js 环境
- Node.js 版本
- npm 版本
- 前端依赖安装状态
- miniprogram-automator 包

#### 6. 服务端口
- 后端服务（8080）
- Knowledge 服务（8100）
- H5 管理端（5173）
- H5 教师端（5174）
- H5 学生端（5175）

#### 7. 测试文件
- Minium 测试脚本
- Playwright 测试脚本
- 测试用例文档

#### 8. 环境变量
- .env 文件存在性
- JWT_SECRET 配置
- OPENAI_API_KEY 配置

### 检查结果

检查结果分为三类：
- ✅ 通过：环境正常
- ⚠️ 警告：可选组件缺失，不影响核心测试
- ❌ 失败：必需组件缺失，需要修复

## 高级用法

### Dry Run 模式

预览部署步骤而不实际执行：

```bash
./deploy-sandbox.sh --host 192.168.1.100 --dry-run full
```

### 自定义环境变量文件

```bash
./deploy-sandbox.sh --host 192.168.1.100 --env-file .env.staging full
```

### 跳过测试

```bash
./deploy-sandbox.sh --host 192.168.1.100 --skip-tests full
```

### 回滚到上一版本

```bash
./deploy-sandbox.sh --host 192.168.1.100 rollback
```
