# 冒烟测试执行指南（Smoke Test Skill）

> **触发条件**：ci_test_agent 在 Phase 3c 执行冒烟测试时加载本文件
> **适用项目**：digital-twin（数字分身教学助手）
> **配套文档**：`docs/smoke-test-cases.md`（用例全集，唯一来源）
> **历史陷阱**：`docs/known_traps.md`（统一陷阱清单，预检时必须查阅）
> **E2E 框架**：Minium（小程序端）+ Playwright（H5 端）
> **测试脚本**：`tests/e2e/smoke_minium.py`（小程序）、`tests/e2e/smoke_playwright.py`（H5）

---

## 0. 环境门禁（Environment Gate）

> ⚠️ **最高优先级。在任何测试代码编写/执行之前，必须先通过本节全部检查项。**
> ⚠️ **本门禁的结果决定后续所有测试是否能执行——不允许在环境未就绪时跳过任何用例。**

### 0.1 门禁规则

```
┌─────────────────────────────────────────────────────┐
│  门禁状态 = 全部 PASS → 进入 §1 概述，开始执行测试    │
│  门禁状态 = 任一 FAIL → 终止执行，报告环境故障         │
│                       → 不允许跳过、不允许降级          │
│                       → 必须修复环境后重新走门禁        │
└─────────────────────────────────────────────────────┘
```

**核心原则**：
- **环境不可用 ≠ 用例可跳过** —— 环境不可用 = 整批测试无法执行，属于前置条件未满足
- **跳过的唯一合法场景**：某用例的**上游依赖用例失败**（标记为 ⏭️ 受阻，见 §4.3）
- **E2E 因环境不可用而跳过是严禁的** —— 这会导致回归盲区

### 0.2 强制检查项（逐项执行，任一FAIL则终止）

#### Check 0: 小程序编译产物（dist/app.json）— 最优先，后续 Minium/E2E 依赖此产物

```bash
# 检查 dist/app.json 是否存在（project.config.json 配置了 miniprogramRoot="dist/"）
FRONTEND_DIR="${PROJECT_PATH}/src/frontend"
if [ -f "$FRONTEND_DIR/dist/app.json" ]; then
  echo "✅ 编译产物存在: dist/app.json"
else
  echo "❌ dist/app.json 不存在，执行自动 build..."
  cd "$FRONTEND_DIR" && /Users/aganbai/local/nodejs/bin/node ./node_modules/@tarojs/cli/bin/taro build --type weapp 2>&1
  if [ -f "$FRONTEND_DIR/dist/app.json" ]; then
    echo "✅ 自动 build 成功，编译产物已生成"
  else
    echo "❌ 自动 build 失败，请手动检查前端代码错误"
    exit 1
  fi
fi
```

> **为什么最优先？** 微信开发者工具报错 `dist/app.json is not found` 就是缺产物导致的。
> **历史教训**：曾因缺此检查导致 Minium/E2E 全部阻塞 ~15 分钟。
>
> **shell 环境限制备忘**：`npm`/`npx` 可能不可用（exit 254），已验证的命令：
> ```bash
> /Users/aganbai/local/nodejs/bin/node ./node_modules/@tarojs/cli/bin/taro build --type weapp
> ```

| 检查项 | 期望 | 失败处理 |
|--------|------|----------|
| `dist/app.json` 存在 | 文件存在且非空 | 自动执行 build |
| 自动 build 失败 | N/A | 报告编译错误详情，终止门禁 |

#### Check 1: 后端 API 服务

```bash
# 必须返回 {"status":"success"}
curl -s http://localhost:8080/api/system/health | python3 -c "
import sys,json
d=json.load(sys.stdin)
assert d.get('success')==True or d.get('status')=='success', f'健康检查失败: {d}'
print('✅ 后端API服务正常')
"
```

| 检查项 | 期望 | 失败处理 |
|--------|------|----------|
| HTTP状态码 | 200 | 启动后端服务：`./dt-server` 或 `go run ./src/cmd/server/` |
| 响应体含 success/status | true/success | 检查端口占用、数据库连接 |

#### Check 2: Minium 连接检查（微信官方 Python 测试框架）

```bash
# 通过 Python 检查 minium 是否可用：
python3 -c "import minium; print('Minium 版本:', minium.__version__)"
# 期望：无报错，输出版本号
```

| 检查项 | 期望 | 失败处理 |
|--------|------|----------|
| `minium` 包已安装 | import 成功 | `pip install minium` |
| 微信开发者工具服务端口开启 | 可连接 | 开发者工具 → 安全 → 服务端口：开启 |
| 项目路径正确 | MiniTest 初始化成功 | 检查 project_path 配置 |

> **Minium 不可用时，小程序 E2E 的 ~25 条用例全部无法执行。这不是"跳过"的理由，而是整个测试批次需要等待环境就绪。**

**Check 2 失败诊断流程**（按顺序执行，不要无限重试）：

```
Step 1: pip show minium → 确认是否已安装
        ├─ 未安装 → pip install minium
        └─ 已安装 → 进入 Step 2

Step 2: 确认开发者工具安全设置
        ├─ 安全 → 服务端口：已开启 → 进入 Step 3
        └─ 未开启 → 手动开启后重试

Step 3: 确认项目编译产物
        ├─ dist/app.json 存在 → 进入 Step 4
        └─ 不存在 → 执行 taro build --type weapp

Step 4: 尝试初始化 MiniTest
        from minium import MiniTest
        mini = MiniTest(project_path="...", cli_path="...")
        ├─ 成功 → ✅ PASS
        └─ 失败 → Step 5

Step 5（最终诊断，输出报告后终止门禁）：

   输出以下诊断报告：
   
   ## 🚫 Minium连接失败诊断
   
   | 排查项 | 结果 |
   |--------|------|
   | Python版本 | {version} |
   | minium包状态 | {pip show结果} |
   | IDE服务端口 | {开启/关闭} |
   | CLI工具路径 | {存在/不存在} |
   | 项目路径 | {path} |
   | 编译产物(dist/app.json) | {存在/不存在} |
   | 错误信息 | {exception详情} |
   
   **根因**: {最可能的根因}
   **修复方法**: {具体操作步骤}
   **归属**: 环境配置(需人工操作) / 依赖问题(需pip install) / 其他
```

#### Check 3: H5 服务三端检查

```bash
# 动态扫描 H5 端口（不写死，自动发现）
# 策略：先扫常见端口范围，再匹配进程名确认端身份

echo "===动态扫描H5服务端口==="
# 方法1：从vite进程获取实际端口
H5_PORTS_JSON=$(ps aux | grep -E "node.*vite.*h5" | grep -v grep | while read line; do
  pid=$(echo $line | awk '{print $2}')
  # 通过lsof找该进程监听的端口
  port=$(lsof -i -P -n -p $pid 2>/dev/null | grep LISTEN | awk '{print $9}' | cut -d: -f2)
  # 从命令行判断是哪个端
  if echo "$line" | grep -q "h5-admin"; then
    echo "{\"endpoint\":\"admin\",\"port\":$port,\"pid\":$pid}"
  elif echo "$line" | grep -q "h5-teacher"; then
    echo "{\"endpoint\":\"teacher\",\"port\":$port,\"pid\":$pid}"
  elif echo "$line" | grep -q "h5-student"; then
    echo "{\"endpoint\":\"student\",\"port\":$port,\"pid\":$pid}"
  fi
done)

# 如果上述方法未找到，回退到常见端口探测（3000-5180）
if [ -z "$H5_PORTS_JSON" ]; then
  for port in $(seq 3000 3010) $(seq 5173 5180); do
    code=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$port --max-time 1 2>/dev/null)
    [ "$code" = "200" ] || [ "$code" = "302" ] && echo "发现H5服务: port=$port (HTTP $code)"
  done
fi

echo "$H5_PORTS_JSON"
```

| 检查项 | 期望 | 失败处理 |
|--------|------|----------|
| 管理员 H5 可达 | HTTP 200/302 | `cd src/h5-admin && npm run dev` |
| 教师 H5 可达 | HTTP 200/302 | `cd src/h5-teacher && npm run dev` |
| 学生 H5 可达 | HTTP 200/302 | `cd src/h5-student && npm run dev` |

> **已知端口参考**（仅供参考，以实际扫描结果为准）：
> - admin: 通常 5173, teacher: 通常 5174, student: 通常 3002
> - **每次测试前必须重新扫描，不可依赖硬编码值**

> **H5 构建产物备选方案**：如果 dev server 未启动但有 dist/ 目录，可用任意静态文件服务器托管 dist/。

#### Check 4: Python 知识服务（可选，仅对话类用例依赖）

```bash
curl -s http://localhost:8100/api/v1/health --max-time 3 | head -5
# 期望 HTTP 200
```

| 检查项 | 期望 | 失败处理 |
|--------|------|----------|
| HTTP 状态码 | 200 | 启动知识服务；如不可用则 E-01~E-10 对话类 API 标记为 ⏭️ 受阻（非跳过） |

#### Check 5: Go 编译验证

```bash
cd /path/to/digital-twin && go build ./...
# 期望无错误输出（exit code 0）
```

### 0.3 门禁结果报告模板

```markdown
## 🚦 环境门禁结果

**检查时间**: {datetime}
**检查人**: ci_test_agent

| # | 检查项 | 状态 | 详情 |
|---|--------|------|------|
| 0 | 小程序编译产物 (dist/app.json) | ✅/❌ | {存在/缺失，如自动build则记录耗时} |
| 1 | 后端 API 服务 (8080/8082) | ✅/❌ | {HTTP状态码或错误信息} |
| 2 | Minium 连接 | ✅/❌ | {版本信息或错误} |
| 3 | 管理员 H5 | ✅/❌/⚪ N/A | {实际端口 + HTTP状态码} |
| 4 | 教师 H5 | ✅/❌/⚪ N/A | {实际端口 + HTTP状态码} |
| 5 | 学生 H5 | ✅/❌/⚪ N/A | {实际端口 + HTTP状态码} |
| 6 | 知识服务 (8100) | ✅/❌/⚪ N/A | {HTTP状态码} |
| 7 | Go 编译 | ✅/❌ | {编译输出摘要} |

**门禁结论**: PASS(全部通过) / FAIL({失败项列表})
```

**结论 = FAIL 时**：
1. 输出上述门禁报告
2. **立即终止，不进入 §1**
3. 明确告知用户哪些环境需要修复
4. 用户修复后重新从本门禁开始

---

## 1. 概述

**执行角色**：ci_test_agent（严禁修改代码，发现问题反馈开发 Agent）

**执行时机**：集成测试全部通过 + 代码 Review 通过后 + **§0 环境门禁 PASS**

**前端形态（三端）**：
- 小程序端：Taro + React，微信开发者工具执行
- H5 管理员端：Vue3 + Element Plus（`src/h5-admin/`），浏览器执行
- H5 教师端：Vue3 + Element Plus（`src/h5-teacher/`），浏览器执行
- H5 学生端：Vue3 + Vant4（`src/h5-student/`），浏览器执行

**核心策略**：
1. **预防优于修复**：编写测试代码前强制预检，消灭高频失败根源
2. **三层分层验证**：API → 小程序 E2E → H5 E2E，前层全通后才进下层
3. **标准化修复循环**：失败后按固定流程修复，每条用例最多 3 轮

---

## 2. 强制预检清单（编写/执行测试代码前必须完成）

> ⚠️ **必须逐项完成。跳过预检是冒烟失败的首要原因。**
> ⚠️ **陷阱详情**：先查阅 `skills/smoke-test-traps.md`，P0 陷阱优先核查。

### 2.1 小程序页面路径验证

```bash
grep "pages/目标页面/index" src/frontend/src/app.config.ts
```

**当前已注册小程序页面（37 个）**：
```
login, role-select, persona-select, home, chat, history,
knowledge/index, knowledge/add, knowledge/preview,
memories, profile, teacher-students, my-teachers,
student-detail, share-join, class-create, class-detail,
class-edit, persona-overview, student-chat-history,
discover, share-manage, memory-manage, teacher-message,
curriculum-config, feedback, feedback-manage, student-batch,
student-profile, approval-manage, approval-detail,
chat-list, my-comments, course-publish, course-list, test-student
```

**已知错误路径（直接 404，见 traps.md T2）**：
- `pages/teacher-dashboard/index` → 应为 `pages/teacher-students/index`
- `pages/student-dashboard/index` → 应为 `pages/home/index`
- `pages/class-share/index` → 不存在，功能在 `pages/class-detail/index`

### 2.2 H5 路由路径验证

H5 三端均为 Vue Router SPA，测试前须确认路由已注册：

```bash
# 管理员端路由
grep "path:" src/h5-admin/src/router/index.ts

# 教师端路由
grep "path:" src/h5-teacher/src/router/index.ts

# 学生端路由
grep "path:" src/h5-student/src/router/index.ts
```

**各端已知核心路由（测试前确认）**：

| 端 | 路由路径 | 页面 |
|----|---------|------|
| 管理员 | `/login` `/dashboard` `/users` `/feedbacks` `/logs` | 登录、仪表盘、用户管理、反馈管理、日志 |
| 教师 | `/login` `/chat-list` `/class/*` `/knowledge` `/profile` | 各功能页 |
| 学生 | `/login` `/home` `/chat` `/discover` `/profile` | 各功能页 |

### 2.3 CSS 选择器验证（小程序端）

每个选择器在写入测试代码前，必须通过 Minium 确认存在：
```
mini_test.app.reLaunch('pages/目标页面/index')
element = mini_test.app.get_element('.目标选择器')  # 或 page.get_element()
```

> 选择器猜测陷阱 → 见 `smoke-test-traps.md` T4

### 2.4 API 端点可达性验证

```bash
# 确认路由已注册
grep "目标路径" src/backend/api/router.go

# 确认端点可达（需后端已启动）
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/目标路径
```

**关键端点速查（按功能分组）**：

| 分组 | 端点 | 状态 |
|------|------|------|
| 认证 | `POST /api/auth/wx-login`、`POST /api/auth/complete-profile`、`GET /api/auth/wx-h5-login-url`、`POST /api/auth/wx-h5-callback` | ✅ 活跃 |
| 班级 | `POST /api/classes`（含分身信息）、`GET /api/classes`、`PUT /api/classes/:id` | ✅ 活跃（迭代11重构） |
| 分身 | `GET /api/personas`（含 bound_class_id）、`PUT /api/personas/:id` | ✅ 活跃 |
| 自测学生 | `GET /api/test-student`、`POST /api/test-student/reset` | ✅ 活跃（迭代11新增） |
| 对话 | `POST /api/chat`、`POST /api/chat/stream` | ✅ 活跃 |
| 管理员 | `GET /api/admin/dashboard/overview`、`GET /api/admin/users`、`GET /api/admin/logs` | ✅ 活跃（迭代10新增） |
| H5 专属 | `POST /api/upload/h5`、`GET /api/platform/config` | ✅ 活跃（迭代10新增） |
| **已废弃（返回 404）** | `PUT /api/personas/:id/switch`、`PUT /api/personas/:id/activate`、`PUT /api/personas/:id/deactivate` | ❌ 迭代11删除 |

> 路由未注册陷阱 → 见 `smoke-test-traps.md` T3

### 2.5 JWT 与认证验证

**小程序端**（wx-login 流程）：
```bash
WX_MODE=mock go test -v -run "TestSmoke_Login" ./tests/integration/
```

**H5 端**（wx-h5-callback 流程）：
```bash
# H5 Mock 模式（需设置 WX_H5_MOCK_ENABLED=true）
curl -X POST http://localhost:8080/api/auth/wx-h5-callback \
  -d '{"code":"mock_h5_code","state":"test"}' \
  -H "Content-Type: application/json"
```

**验证 token 内容**（两端均适用）：
```js
JSON.parse(atob(token.split('.')[1]))
// 重点检查：user_id 类型为 float64（Go JSON 默认行为）、role 字段值格式
```

> JWT 类型断言陷阱 → T1；role 不一致 → T6

### 2.6 环境就绪检查（三端）

```bash
# 1. 后端服务
curl -s http://localhost:8080/api/system/health     # 期望 200

# 2. Python 知识服务（注意：端口是 8100，不是 8000）
curl -s http://localhost:8100/api/v1/health         # 期望 200

# 3. Go 编译
cd /path/to/digital-twin && go build ./...          # 期望无错误

# 4. 小程序 E2E（需要 Minium 框架）
# → python3 tests/e2e/smoke_v12_minium.py（或 from minium import MiniTest）

# 5. H5 端（需要构建产物）
ls src/h5-admin/dist/    # 管理员 H5
ls src/h5-teacher/dist/  # 教师 H5
ls src/h5-student/dist/  # 学生 H5
# 或本地开发服务（dev server）
curl -s http://localhost:5173  # 管理员 H5 dev（确认实际端口）
```

---

## 3. 分层执行策略（三层）

```
┌─────────────────────────────────────────────────┐
│  层级1：API 验证（Go test，所有端共享）           │
│  层级2：小程序 E2E（Minium Python 框架）         │
│  层级3：H5 E2E（浏览器，三端）                   │
└─────────────────────────────────────────────────┘
前层全部 PASS → 才进入下层
前层有 FAIL → 停止，修复后重跑，不跳层
```

### ⛔ E2E 用例禁止跳过规则

> **这是本节最重要的规则。违反此规则 = 测试结果无效。**

| 场景 | 允许的操作 | 禁止的操作 |
|------|-----------|-----------|
| 微信开发者工具未启动 | **终止测试**，报告环境门禁 FAIL | 标记 E2E 用例为 ⏭️ 跳过 |
| Minium 连接失败 | **终止测试**，要求安装 minium 并开启开发者工具服务端口 | 跳过小程序E2E继续测H5 |
| H5 dev server 未运行 | **终止测试**或启动服务 | 跳过H5-E2E只报API结果 |
| 某页面渲染失败 | 按 §5 修复循环处理 | 跳过该用例测下一个 |
| 上游依赖用例失败 | 下游标记 **⏭️ 受阻**（非"跳过"） | 无 |

**关键区分**：

| 标记 | 含义 | 使用场景 | 计入通过率？ |
|------|------|----------|-------------|
| ✅ PASS | 测试通过 | — | ✅ 是 |
| ❌ FAIL | 测试失败 | 实际结果不符预期 | ❌ 否（算失败） |
| ⏭️ 受阻 | 因上游依赖失败无法执行 | 前置用例❌导致无法准备数据 | ⚪ 不计入(单独统计) |
| 🚫 环境阻塞 | 因环境不可用无法整批执行 | Minium/H5服务未启动 | **整批测试无效** |

**⏭️ 受阻 vs 🚫 环境阻塞的区别**：
- **⏭️ 受阻**：代码层面的问题（某功能bug导致下游无法测），是合理的测试状态
- **🚫 环境阻塞**：基础设施问题（MCP没连、H5没启动），说明测试条件不满足，**必须修好环境后重跑**

### 3.1 层级1：API 验证

**执行方式**：Go test（`WX_MODE=mock LLM_MODE=mock`）

```bash
cd /path/to/digital-twin
go test -v ./tests/integration/ -run "TestSmoke_" -count=1 -timeout 300s
```

**覆盖范围**：后端所有接口逻辑（小程序端和 H5 端共享同一套后端，一次验证全部覆盖）

**通过标准**：全部 PASS（SKIP 允许，FAIL 不允许）

**失败处理**：停止，不进入 E2E 层。问题 100% 归属后端，按 §5 修复循环处理。

### 3.2 层级2：小程序 E2E 验证

**前置条件**：API 层全部通过 + **§0 Check 2 (Minium) PASS**

> ⚠️ 如果 Minium 不可用（§0 门禁已拦截则不会到达此处；若运行中 Minium 断开）：
> - **立即终止本层所有测试**
> - 输出 🚫 环境阻塞报告
> - **不允许逐条标记跳过**

**执行方式**：Minium Python 框架控制微信开发者工具模拟器。
测试脚本入口：`tests/e2e/smoke_v12_minium.py`

**覆盖范围**：`docs/smoke-test-cases.md` 中「验证方式=E2E」且「前端=小程序」的用例

```
执行步骤：
1. MiniTest(project_path, cli_path) → 初始化连接
2. mini_test.app.get_current_page() → 确认当前页面
3. 按用例步骤执行（详见 §6.2）
4. mini_test.capture() → 截图留证
5. mini_test.console_logs() → 检查 console.error
```

**通过标准**：页面渲染正确 + 交互响应正确 + 无 console.error

### 3.3 层级3：H5 E2E 验证

**前置条件**：小程序 E2E 层全部通过（或无小程序用例时：API 层全部通过）+ **§0 Check 3 (H5服务) 对应端 PASS**

> ⚠️ 如果目标 H5 端不可访问：
> - **立即终止该端的测试**
> - 输出 🚫 环境阻塞报告
> - 不允许跳过该端继续测其他端（除非该端确实不在本轮测试范围内）

**执行方式**：Playwright 浏览器自动化

**测试脚本入口**：`tests/e2e/smoke_playwright.py`

**覆盖范围**：`docs/smoke-test-cases.md` 中 S/T 模块（管理员H5）和 H5 三端核心用例

**Playwright 执行步骤**：
```python
from playwright.sync_api import sync_playwright

with sync_playwright() as p:
    browser = p.chromium.launch(headless=False)
    context = browser.new_context()
    page = context.new_page()
    
    # 导航到目标 H5 页面
    page.goto(f"http://localhost:{h5_port}/target-path")
    
    # 等待页面加载
    page.wait_for_load_state("networkidle")
    
    # 执行操作步骤
    page.click(".target-selector")
    page.fill("input[name='field']", "value")
    
    # 验证预期结果
    assert page.locator(".result").is_visible()
    
    # 截图留证
    page.screenshot(path="screenshot.png")
```

**覆盖范围**：`docs/smoke-test-cases.md` 中 S/T 模块（管理员H5）和 H5 三端核心用例

**三端执行顺序**：
1. **管理员 H5**（优先，权限最高，问题影响面最大）
2. **教师 H5**（功能最多）
3. **学生 H5**（依赖教师端数据）

**H5 E2E 专项注意**：
- H5 登录走微信 OAuth（`wx-h5-callback`），开发环境须启用 `WX_H5_MOCK_ENABLED=true`
- H5 Storage 与小程序 Storage 完全独立，H5 用 `localStorage`
- H5 无 `switchTab` 概念，导航用 `router.push`
- CORS：H5 请求头须带 `Origin`，确认后端 CORS 配置允许开发域名

> H5 CORS/OAuth 陷阱 → 见 §2.5 H5 端检查项

### 3.4 用例分类速查

参考 `docs/smoke-test-cases.md` 「验证方式」列：

| 层级 | 用例范围 | 数量（参考） |
|------|---------|------------|
| API | A-01~05、B-01/02/05、C-01~06、D-01~06、E-01/02/04/05/06/09/10、F-01~05、G-01/02/G-06、H-01~07、I-01~05、J-01/02、K-01~03/05、L-01/04、M-01~03、N-01~03、O-01/02、P-01~04、Q-01~05、S-01~14、T-01~03、AD-01~04、AE-01~04、AF-01~02 | ~95条 |
| 小程序 E2E | A-06、B-04、C-07/08、D-07/08、E-03/07/08、G-03~05、H-08、I-06、J-03、K-04、L-02/03、M-04、N-04、O-03、R-01/02、AD-05 | ~25条 |
| H5 E2E | S-01~03（H5 登录流程）、H5 管理员仪表盘渲染、H5 教师/学生核心页面渲染 | ~10条 |

**优先级执行顺序**：P0 → P1 → P2（P0 失败则中断整批，与层级无关）

---

## 4. 依赖图与执行编排

### 4.1 主依赖链（迭代11版本）

> 注意：迭代11后，分身不再独立创建，`SM-G01`（创建班级）**同时**产出 classID 和 personaID。

```
【认证链】
SM-A01（学生注册）→ SM-B01（学生分身）
SM-A02（教师注册）→ [自动] 自测学生创建

【班级+分身链】（迭代11核心变更）
SM-A02 → SM-G01（创建班级+分身）→ SM-D01（创建分享码）
                    ↓                        ↓
                SM-AD01（验证班级绑定分身）  SM-H01（学生加入）→ SM-AA01（教师审批）
                    ↓                                              ↓
                SM-AE03（自测学生自动加入）                  SM-D01（学生对话）

【对话链】
SM-D01 → SM-D02（SSE 流式）
SM-D01 → SM-F01（教师介入）
SM-D01 → SM-N05（记忆自动提取）→ SM-N01（查看记忆）

【H5 管理员链】（独立，仅依赖管理员账号）
管理员登录（S-01/S-02）→ S-04（仪表盘）→ S-07（用户管理）→ S-09（禁用用户）→ S-10/S-11（验证禁用效果）
                                      → S-12（日志查询）→ S-14（导出 CSV）
```

### 4.2 独立可并行用例

以下用例无数据依赖，可与主链并行执行：

```
SM-AD02（禁止独立创建分身）  ← 独立执行
SM-AD04（已删除接口返回404） ← 独立执行
SM-T01（Docker 部署）        ← 独立执行
SM-Q01（未登录访问）         ← 独立执行
SM-T-01（H5 平台配置）       ← 独立执行
```

### 4.3 失败隔离规则

```
用例 X 失败 → 仅阻塞 X 的直接下游（标记 ⏭️ 受阻）
             → 不阻塞与 X 无依赖关系的其他用例

示例：
SM-G01 失败 → 阻塞 SM-AD01/AE03/D01/H01...（全部班级相关下游）
            → 不影响 SM-AD02/SM-Q01/S-04（独立用例）

S-09（禁用用户）失败 → 阻塞 S-10/S-11
                    → 不影响 S-04/S-12
```

### 4.4 跨迭代增量合并

迭代增量文件（如 `smoke_test_plan_v*.md`）存在时：
1. 读取增量文件的依赖图
2. 将增量用例追加到对应主链（按依赖标注）
3. 独立用例直接追加到可并行组

---

## 5. 失败修复循环

> **目标成功率**：3 轮内修复率 ≥ 90%

### 5.1 失败报告标准格式

每条失败用例**立即**输出，不得遗漏字段：

```markdown
## ❌ {用例编号}: {用例名称}

**失败层级**: API / 小程序E2E / H5-E2E（管理员/教师/学生）
**失败步骤**: 第 N 步 —— {步骤描述}
**期望结果**: {预期值}
**实际结果**: {实际值}
**HTTP 状态码**: {code}（API 层必填）
**错误码 / 错误信息**: {error_code} / {message}
**截图路径**: {path}（E2E 必填）
**控制台日志**: {关键错误行}（E2E 必填）
**相关代码**: {file:line}

**问题归属**: 后端 / 小程序前端 / H5前端 / 联调 / 测试脚本
**归属判定依据**: {具体判定逻辑，见 §5.2}
**修复方向**: {1~2句话描述修复切入点}
```

### 5.2 问题归属判定规则

| 现象 | 归属 | 转交 Agent |
|------|------|------------|
| API 返回 4xx/5xx | 后端 | dev_backend_agent |
| API 返回 401（JWT 类型断言）| 后端 | dev_backend_agent |
| API 返回 404（路由未注册）| 后端 | dev_backend_agent |
| API 正常，小程序页面不渲染/元素不存在 | 小程序前端 | dev_frontend_agent |
| API 正常，H5 页面不渲染/元素不存在 | H5前端 | dev_h5_agent |
| H5 请求被 CORS 拒绝（Console 报 CORS error）| 后端（CORS 配置）| dev_backend_agent |
| H5 OAuth 回调失败（wx-h5-callback 返回错误）| 后端（OAuth 逻辑）| dev_backend_agent |
| H5 登录后 token 无法持久化（刷新丢失）| H5前端（localStorage 存储）| dev_h5_agent |
| API 正常，前后端数据格式不一致 | 联调 | 双方 Agent |
| 小程序页面路径不存在 / CSS 选择器不匹配 | 测试脚本 | ci_test_agent 自行修正 |
| H5 路由不存在 | 测试脚本 | ci_test_agent 自行修正 |
| Storage userInfo 字段缺失（小程序）| 小程序前端 | dev_frontend_agent |

**快速四步定位法**（适用于所有层级）：
```
Step 1: 直接 curl API → 失败 → 后端问题
Step 2: API 正常 → 检查 console 报错 → 有报错 → 对应端前端问题
Step 3: 无报错 → 检查请求/响应数据字段格式差异 → 联调问题
Step 4: 数据也正常 → 检查选择器/路由路径 → 测试脚本自身问题
```

**H5 端专项排查（Step 1.5，在 Step 1 和 Step 2 之间执行）**：
```
H5 失败时，先检查：
a. 是否有 CORS 错误（浏览器 console: "Access-Control-Allow-Origin"）
b. OAuth token 是否正确存储（检查 localStorage.getItem('token')）
c. 请求 Origin 头是否在后端白名单中
```

### 5.3 修复循环主流程

```
每条失败用例最多 3 轮，超过则人工介入
```

**Round 1 步骤**：

```
1. [ci_test_agent] 输出 §5.1 标准失败报告（含失败层级字段）
2. [ci_test_agent] 按 §5.2 判定归属
3. [ci_test_agent → 主 Agent] 发送失败报告，注明归属和修复方向
4. [主 Agent] 使用 §5.4 修复 Prompt 模板，转交对应开发 Agent
5. [开发 Agent] 执行修复，输出修复摘要
6. [ci_test_agent] 收到修复完成通知后：
   a. 重新执行 §2 对应预检项（确认根因覆盖）
   b. 重新执行该条失败用例
   c. 重新执行该用例的所有下游依赖用例
7. 通过 → 标记 ✅；失败 → 进入 Round 2
```

**Round 2 额外步骤**：
```
- 对比 Round 1 和 Round 2 失败信息差异：
  - 完全相同 → 修复无效，要求重新分析根因，提供代码行号+变量值
  - 有变化 → 方向正确但不完整，提供更精确定位（具体代码行/变量值）
- H5 端特别：提供完整的 Network 请求/响应截图（Request Headers + Response Body）
```

**Round 3 额外步骤**：
```
- 提供可能的修复代码片段（仅供参考）
- 检查是否为环境/配置问题（非代码问题）：
  - API 层：检查 go.sum / 依赖版本
  - 小程序 E2E：检查模拟器版本/微信开发者工具版本
  - H5 E2E：检查 CORS 配置文件 / Nginx 反向代理配置
```

**超过 3 轮**：
```
1. 标记该用例为「需人工介入」，记录全部 3 轮失败详情
2. 跳过该用例及其下游依赖用例（标记为 ⏭️ 受阻）
3. 继续执行无依赖的其他用例
4. 最终报告中单独汇总人工介入项
```

### 5.4 修复 Prompt 模板

```
@{agent_name}

## 冒烟测试修复任务（Round {N}）: {用例编号}

### 失败现象
{粘贴 §5.1 的完整失败报告，含失败层级}

### 归属判定
{归属类型} — {判定依据}

### 修复方向
{1~2句话描述修复切入点}

### 约束
1. 仅修复该问题，不做额外重构
2. 修复后确保 `go build ./...` 无错误
3. 如涉及 API 修改：`go test -run "TestSmoke_" ./tests/integration/` 通过
4. 如涉及 H5 修改：确认 CORS 配置和 OAuth 回调逻辑正确
5. 输出修复摘要：修改了哪些文件、修改了什么、为什么这样修复
```

**Round 2/3 附加字段**：
```
### 前序修复记录
Round {N-1} 修复摘要: {摘要}
Round {N-1} 修复后现象: {差异对比，含错误码变化}
```

### 5.5 测试脚本自身问题的修复

归属为「测试脚本」时，ci_test_agent 自行修复：

```
【小程序端】
1. 页面路径错误 → 查阅 §2.1 已注册页面列表，更正路径
2. CSS 选择器错误 → 重新执行 §2.3 预检，用 mini_test.app.get_element() 查找正确选择器

【H5 端】
3. H5 路由路径错误 → 查阅 §2.2 路由列表或 Vue Router 配置，更正路径
4. H5 元素选择器错误 → 用浏览器开发者工具检查 DOM，更正选择器

【通用】
5. API 路径错误 → 查阅 §2.4 已注册端点速查表，更正路径
6. 修复后直接重新执行该用例，不转交开发 Agent
7. 修复耗时 > 5 分钟 → 记录过程并追加到 smoke-test-traps.md
```

---

## 6. 单用例执行 SOP

### 6.1 API 用例执行流程

```
1. 确认前置数据（依赖用例的产出：token、classID、personaID 等）
2. 构建请求（method + path + Authorization header + body）
3. 发送请求，记录响应（status_code + response_body）
4. 逐项验证预期结果（含字段类型检查）
5. 通过 → 输出 ✅ + 产出数据（供下游用例使用）
6. 失败 → 输出 ❌ + §5.1 失败报告（失败层级=API）
```

### 6.2 小程序 E2E 用例执行流程

```
前置检查（每批用例开始前执行一次，非逐条执行）：
  ├─ MiniTest 初始化 → 确认 Minium 连接
  │   └─ 失败 → 🚫 整批环境阻塞，终止（不标记跳过）
  └─ mini_test.app.get_current_page() → 确认模拟器正常响应
      └─ 失败 → 同上

步骤 1: MiniTest(project_path, cli_path) → 连接开发者工具
        失败则检查 IDE 服务端口是否开启，重新初始化

步骤 2: mini_test.app.reLaunch('/pages/目标页面/index') 或 redirectTo / navigateTo / switchTab
        → mini_test.wait(2000) 等待加载（见 §6.4 等待策略）
        → mini_test.app.get_current_page() → 确认路径正确（路径不对立即报告归属测试脚本）

步骤 3: 执行操作步骤
        → page.get_element('.目标元素') 或 mini_test.app.get_element() → 确认元素存在
          （元素不存在 → 归属前端，立即输出失败报告，不继续后续步骤）
        → element.tap() / element.input() → 操作
        → 按 §6.4 策略等待响应

步骤 4: 验证预期结果
        → element.get_text() / element.get_value() → 检查 text/value
        → mini_test.capture(path=...) → 截图（路径记入报告）
        → mini_test.console_logs() → 确认无 console.error

步骤 5: 输出结果
        → 通过: ✅ + 截图路径 + 关键数据
        → 失败: ❌ + §5.1 失败报告（失败层级=小程序E2E）
```

### 6.3 H5 E2E 用例执行流程

```
前置检查（每端开始前执行一次）：
  ├─ curl 目标 H5 端口 → 确认服务可达（HTTP 200/302）
  │   └─ 失败 → 🚫 该端环境阻塞，终止（尝试启动 dev server 后重试一次）
  └─ curl /api/system/health → 确认后端正常
      └─ 失败 → 🚫 整批环境阻塞，终止

步骤 1: 确认 H5 服务可访问
        curl -s -o /dev/null -w "%{http_code}" http://localhost:{h5_port}
        期望 200；否则报告「H5 服务未启动」，归属环境问题

步骤 2: Playwright 浏览器启动
        from playwright.sync_api import sync_playwright
        p = sync_playwright().start()
        browser = p.chromium.launch(headless=False)
        page = browser.new_page()

步骤 3: 导航到目标页面
        page.goto(f"http://localhost:{h5_port}/target-path")
        page.wait_for_load_state("networkidle")

步骤 4: H5 登录流程（如未登录）
        → 开发环境：使用 WX_H5_MOCK_ENABLED=true，直接 POST /api/auth/wx-h5-callback
        → 获取 token，存入 localStorage：page.evaluate("localStorage.setItem('token', '{token}')")
        → 检查 token 是否正确存储：page.evaluate("localStorage.getItem('token')")

步骤 5: 执行操作步骤
        → 定位元素：page.locator(".target-selector")
        → 点击：page.click(".button")
        → 输入：page.fill("input[name='field']", "value")
        → 等待响应（见 §6.4）

步骤 6: 验证预期结果
        → 检查元素可见性：page.locator(".result").is_visible()
        → 检查文本：page.locator(".text").inner_text()
        → 截图留证：page.screenshot(path="screenshot.png")
        → 检查 console 无报错：page.on("console", lambda msg: ...)

步骤 7: 输出结果
        → 通过: ✅ + 截图路径 + 关键数据
        → 失败: ❌ + §5.1 失败报告（失败层级=H5-E2E，注明端：管理员/教师/学生）
```

### 6.4 等待策略

| 场景 | 等待时间 | 适用层级 |
|------|----------|---------|
| 小程序页面导航后 | 2000ms | 小程序 E2E |
| 小程序 TabBar 切换后 | 1500ms | 小程序 E2E |
| 小程序元素操作后 | 1000ms | 小程序 E2E |
| H5 页面导航后 | 3000ms（SPA 渲染较慢）| H5 E2E |
| H5 元素操作后 | 800ms | H5 E2E |
| API 调用后 | 500ms | 所有层级 |
| 异步加载（含 LLM 回复）| 轮询，最长 30s | 所有层级 |
| 含异步数据的页面 | ≥ 5000ms 或轮询 | 所有层级 |

### 6.5 小程序 TabBar 页面（必须用 switchTab）

```
pages/chat-list/index, pages/home/index, pages/discover/index,
pages/teacher-students/index, pages/knowledge/index, pages/profile/index
```

> 误用 navigateTo 会导致页面不响应，归属测试脚本问题。

---

## 7. 冒烟验证报告模板

```markdown
# 迭代{N} Phase 3c 冒烟测试报告

**执行时间**: {datetime}
**测试环境**: localhost / staging
**执行策略**: 三层（API → 小程序E2E[Minium] → H5-E2E）
**Skill版本**: smoke-test-skill.md v4.0

## 🚦 环境门禁结果

| # | 检查项 | 状态 | 详情 |
|---|--------|------|------|
| 1 | 后端 API 服务 | ✅/❌ | {详情} |
| 2 | Minium 连接 | ✅/❌ | {详情} |
| 3 | 管理员 H5 | ✅/❌/⚪N/A | {详情} |
| 4 | 教师 H5 | ✅/❌/⚪N/A | {详情} |
| 5 | 学生 H5 | ✅/❌/⚪N/A | {详情} |
| 6 | 知识服务 | ✅/❌/⚪N/A | {详情} |
| 7 | Go 编译 | ✅/❌ | {详情} |

**门禁结论**: PASS / FAIL({失败项})

> ⚠️ 门禁 FAIL 时以下测试结果无效，需修复环境后重跑

## 测试结果汇总

| 层级 | 总用例 | ✅通过 | ❌失败 | ⏭️受阻 | 🚫环境阻塞 | 通过率 |
|------|--------|--------|--------|--------|-----------|--------|
| API 层 | {n} | {p} | {f} | {b} | - | {rate} |
| 小程序 E2E | {n} | {p} | {f} | {b} | {env_block} | {rate} |
| H5 管理员 E2E | {n} | {p} | {f} | {b} | {env_block} | {rate} |
| H5 教师 E2E | {n} | {p} | {f} | {b} | {env_block} | {rate} |
| H5 学生 E2E | {n} | {p} | {f} | {b} | {env_block} | {rate} |
| **合计（不含环境阻塞）** | **{n}** | **{p}** | **{f}** | **{b}** | **-** | **{rate}** |

> 注：🚫环境阻塞表示该层因基础设施不可用无法执行，不计入通过率计算但必须单独标注

## 用例执行详情

| 用例ID | 场景 | 层级 | 状态 | 修复轮次 | 备注 |
|--------|------|------|------|----------|------|
| SM-xxx | ... | API/小程序/H5-管理员/H5-教师/H5-学生 | ✅/❌/⏭️/🚫 | 0~3/人工 | ... |

## 失败用例详情
{每条失败用例的 §5.1 标准失败报告}

## 修复记录
{每轮修复摘要，含修改文件和改动内容}

## 人工介入项
{超过 3 轮未修复的用例，含全部失败详情}

## 依赖链执行顺序
{实际执行的依赖图，含状态标记（✅/❌/⏭️）}

## 结论
{总体评估 + 遗留问题 + 建议}
```

---

**文档版本**: v4.0.0
---

**文档版本**: v5.0.0
**更新日期**: 2026-04-06

**v5.0 变更（重大）**:
- **E2E 框架标准化**：小程序端使用 Minium，H5 端使用 Playwright
- **陷阱清单统一**：合并 `smoke-test-traps.md` 到 `docs/known_traps.md`
- **用例来源唯一**：`docs/smoke-test-cases.md` 为唯一用例全集，删除重复的 `smoke_test_plan.md`
- **路径配置化**：移除硬编码绝对路径，使用 `${PROJECT_PATH}` 环境变量

**关联文档**:
- 统一陷阱清单: `docs/known_traps.md`
- 用例全集: `docs/smoke-test-cases.md`（唯一来源）
- 接口全量: `docs/iterations/v2.0/api_spec_full.md`
- Minium 官方文档: https://minitest.weixin.qq.com/#/minium/Python/readme
- Playwright 官方文档: https://playwright.dev/python/
- E2E 测试脚本: `tests/e2e/smoke_minium.py`（小程序）、`tests/e2e/smoke_playwright.py`（H5）

... EOF no more lines ...