# 部署和测试脚本快速入门

本文档提供部署和冒烟测试脚本的快速入门指南。

## 📦 脚本清单

### 部署脚本

| 脚本 | 语言 | 用途 | 路径 |
|------|------|------|------|
| `deploy-sandbox.sh` | Shell | 一键部署到沙盒 | `deployments/scripts/` |
| `deploy_sandbox.py` | Python | 一键部署（编程接口） | `deployments/scripts/` |

### 环境检查脚本

| 脚本 | 语言 | 用途 | 路径 |
|------|------|------|------|
| `check-smoke-env.sh` | Shell | 冒烟环境检查 | `deployments/scripts/` |
| `check_smoke_env.py` | Python | 环境检查（编程接口） | `deployments/scripts/` |

### 编排脚本

| 脚本 | 语言 | 用途 | 路径 |
|------|------|------|------|
| `smoke_test_orchestrator.py` | Python | 完整编排流程 | `deployments/scripts/` |

## 🚀 快速开始

### 1. 部署到沙盒

```bash
# 完整部署流程（编译 → 上传 → 重启）
./deployments/scripts/deploy-sandbox.sh --host 192.168.1.100 full

# 仅构建镜像
./deployments/scripts/deploy-sandbox.sh build

# 查看服务状态
./deployments/scripts/deploy-sandbox.sh --host 192.168.1.100 status
```

### 2. 检查冒烟环境

```bash
# 执行环境检查
./deployments/scripts/check-smoke-env.sh

# 自动修复问题
./deployments/scripts/check-smoke-env.sh --fix

# Python 版本（支持 JSON 输出）
python3 deployments/scripts/check_smoke_env.py --json
```

### 3. 完整编排流程

```bash
# 部署 + 环境检查 + 冒烟测试
python3 deployments/scripts/smoke_test_orchestrator.py --host 192.168.1.100
```

## 🔧 在编排任务中使用

### 方式一：直接调用脚本

```python
import subprocess

# 部署
subprocess.run([
    "./deployments/scripts/deploy-sandbox.sh",
    "--host", "192.168.1.100",
    "full"
])

# 环境检查
subprocess.run([
    "python3", 
    "deployments/scripts/check_smoke_env.py",
    "--fix"
])
```

### 方式二：使用 Python API

```python
from pathlib import Path
from deployments.scripts.deploy_sandbox import SandboxDeployer, SandboxConfig
from deployments.scripts.check_smoke_env import SmokeEnvChecker

# 部署
config = SandboxConfig(host="192.168.1.100")
deployer = SandboxDeployer(Path("."), config)
deployer.deploy_full()

# 环境检查
checker = SmokeEnvChecker(Path("."), auto_fix=True)
if checker.run_all_checks():
    print("环境就绪")
```

### 方式三：使用编排器

```python
from pathlib import Path
from deployments.scripts.smoke_test_orchestrator import SmokeTestOrchestrator

orchestrator = SmokeTestOrchestrator(Path("."))
orchestrator.deploy_and_test(sandbox_host="192.168.1.100")
```

## 📋 环境变量配置

在编排任务中配置以下环境变量：

```bash
# 沙盒服务器
SANDBOX_HOST=192.168.1.100
SANDBOX_USER=root
SANDBOX_PORT=22
SANDBOX_DEPLOY_DIR=/opt/digital-twin

# 镜像仓库（可选）
IMAGE_REGISTRY=registry.example.com/digital-twin
```

## 🔍 故障排查

### 部署失败

1. 查看日志：`/tmp/deploy-sandbox-*.log`
2. 检查 SSH 连接：`ssh root@192.168.1.100`
3. 检查 Docker 环境：`docker ps`

### 环境检查失败

1. 使用 `--fix` 参数自动修复
2. 查看详细报告：`python3 check_smoke_env.py --json`
3. 手动安装缺失依赖

### 测试执行失败

1. 确认服务正常运行
2. 检查测试脚本路径
3. 查看测试输出日志

## 📚 相关文档

- [部署详细文档](./README.md)
- [冒烟测试用例](../docs/smoke-test-cases.md)
- [已知陷阱](../docs/known_traps.md)
- [项目配置](../docs/project_config.yaml)

## 🆘 获取帮助

```bash
# 部署脚本帮助
./deployments/scripts/deploy-sandbox.sh --help

# 环境检查帮助
python3 deployments/scripts/check_smoke_env.py --help

# 编排器帮助
python3 deployments/scripts/smoke_test_orchestrator.py --help
```
