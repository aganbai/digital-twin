# 环境配置指导文档

**文档版本**: v1.0  
**适用项目**: digital-twin V2.0 迭代12  
**配置目标**: Phase 3c 冒烟测试环境准备

## 1. 微信开发者工具配置

### 1.1 安装微信开发者工具

**下载地址**: https://developers.weixin.qq.com/miniprogram/dev/devtools/download.html

**安装步骤**:
1. 下载对应操作系统的安装包
2. 完成安装并启动开发者工具
3. 使用微信扫码登录

### 1.2 开启 CLI 功能

**图形界面配置**:
1. 打开微信开发者工具
2. 进入 "设置" → "安全设置"
3. 开启 "服务端口" 功能
4. 可选：设置自定义端口号（默认 9420）

**验证 CLI 功能**:
```bash
# 检查 CLI 是否可用
cli --version

# 如果命令不存在，可能需要配置环境变量
# macOS 默认安装路径: /Applications/wechatwebdevtools.app/Contents/MacOS/cli

# 添加别名或软链接
echo 'alias cli="/Applications/wechatwebdevtools.app/Contents/MacOS/cli"' >> ~/.zshrc
source ~/.zshrc
```

### 1.3 项目导入和配置

```bash
# 导入项目到开发者工具
cli -o --project /Users/aganbai/Desktop/WorkSpace/digital-twin/src/frontend

# 编译项目
cli -c --project /Users/aganbai/Desktop/WorkSpace/digital-twin/src/frontend

# 预览项目
cli -p --project /Users/aganbai/Desktop/WorkSpace/digital-twin/src/frontend
```

## 2. 后端服务配置

### 2.1 后端项目结构分析

根据检查，后端目录结构如下：
```
src/backend/
├── api/           # API 接口层
├── database/      # 数据库层
├── data/          # 数据相关
├── knowledge/     # 知识服务
└── debug_*.js     # 调试脚本
```

### 2.2 寻找后端启动方式

**可能的启动方式**:

#### 方案一: Go 语言启动
```bash
# 检查是否有 Go 模块文件
find ./src/backend -name "go.mod" -o -name "main.go"

# 如果存在 Go 项目
cd ./src/backend
go mod tidy
go run main.go
```

#### 方案二: Node.js 启动
```bash
# 检查是否有 package.json
find ./src/backend -name "package.json"

# 如果存在 Node.js 项目
cd ./src/backend
npm install
npm start
```

#### 方案三: Python 启动
```bash
# 检查是否有 Python 入口文件
find ./src/backend -name "*.py" | grep -E "(app|main|server)"

# 如果存在 Python 项目
cd ./src/backend
python app.py
```

### 2.3 后端服务验证

**健康检查端点**:
```bash
# 验证服务是否启动
curl -s http://localhost:3000/api/health

# 预期响应格式（示例）
{
  "status": "healthy",
  "timestamp": "2026-04-07T20:36:00Z",
  "version": "v2.0"
}
```

## 3. 测试环境配置

### 3.1 数据库配置

**检查数据库连接**:
```bash
# 检查数据库配置文件
find ./src/backend -name "*.env" -o -name "config.*" -o -name "database.*"

# 常见的数据库配置位置
# - .env 文件
# - config/database.js
# - database/config.go
```

### 3.2 测试数据准备

**创建测试用户和数据**:
```bash
# 检查是否有测试数据脚本
find ./src/backend -name "*seed*" -o -name "*fixture*" -o -name "*test*data*"

# 执行测试数据初始化
# 通常命令如: npm run seed, go run seed.go, python init_data.py
```

## 4. 冒烟测试执行准备

### 4.1 环境变量配置

创建测试环境配置文件：
```bash
# 创建测试环境配置
echo "NODE_ENV=test" > ./src/frontend/.env.test
echo "API_BASE_URL=http://localhost:3000" >> ./src/frontend/.env.test
echo "WEAPP_DEBUG=true" >> ./src/frontend/.env.test
```

### 4.2 测试脚本验证

**验证现有测试脚本**:
```bash
# 检查现有的端到端测试
ls -la ./e2e/

# 运行简单的测试检查
cd ./src/frontend
npm run test:e2e:check
```

## 5. 故障排除指南

### 5.1 常见问题解决

**微信开发者工具连接失败**:
```bash
# 检查服务端口是否开启
netstat -an | grep 9420

# 重启开发者工具
cli -r --project /path/to/project
```

**后端服务启动失败**:
```bash
# 检查端口占用
lsof -i :3000

# 检查依赖是否完整
# Go: go mod tidy
# Node.js: npm install
# Python: pip install -r requirements.txt
```

**数据库连接失败**:
```bash
# 检查数据库服务状态
# MySQL: sudo systemctl status mysql
# PostgreSQL: sudo systemctl status postgresql
# MongoDB: sudo systemctl status mongod
```

### 5.2 日志和调试

**启用详细日志**:
```bash
# 后端服务日志
cd ./src/backend
# 设置日志级别为 DEBUG

# 前端调试
cd ./src/frontend
# 启用开发者工具调试模式
```

## 6. 配置验证清单

### 6.1 预检清单

- [ ] 微信开发者工具已安装并登录
- [ ] CLI 功能已开启并验证
- [ ] 后端服务启动方式已确认
- [ ] 数据库连接配置正确
- [ ] 测试环境变量已设置
- [ ] 健康检查端点可访问

### 6.2 冒烟测试准备清单

- [ ] 微信开发者工具项目导入成功
- [ ] 后端服务在端口 3000 运行
- [ ] API 健康检查返回 "healthy"
- [ ] 测试用户数据已准备
- [ ] 测试脚本可正常执行

## 7. 后续步骤

完成环境配置后，按以下顺序执行冒烟测试：

1. **环境验证** - 确认所有服务正常运行
2. **核心功能测试** - 执行 Part A 测试用例
3. **完整流程测试** - 执行 Part B 测试用例
4. **异常场景测试** - 执行 Part C 测试用例
5. **测试报告生成** - 汇总结果和截图

---

**文档维护**:  
- 最后更新: 2026-04-07
- 维护者: Claude Code Agent
- 关联文档: [冒烟测试用例](./smoke_tests.yaml)