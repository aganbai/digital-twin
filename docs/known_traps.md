# 已知陷阱清单

> ci_test_agent 编写集成测试和冒烟测试前必须阅读本清单，避免重复踩坑。开发 Agent 也应参考。
> 
> 本文件由 dt-orchest 动态加载。

---

## P0 陷阱（必须避免，违反直接导致测试失败）

### T1: API 请求参数名与实际 handler 不一致
| 字段 | 内容 |
|------|------|
| 现象 | 集成测试全部失败，需逐个修复 |
| 根因 | 测试用 `username` 但 handler 期望 `phone` |
| 来源 | V2.0-iter4 |
| 防护 | ci_test_agent 以 API 规范为准编写测试，发现不一致则反馈开发 Agent 修改代码 |

### T2: 新增数据库字段遗漏 INSERT SQL 和 handler 响应
| 字段 | 内容 |
|------|------|
| 现象 | 字段写入为空值，响应缺少字段 |
| 根因 | 只改了模型定义和 DDL，遗漏了 INSERT SQL 和 handler 响应 |
| 来源 | V2.0-iter4 |
| 防护 | 开发 Agent 必须先输出「字段影响清单」 |

### T3: JWT user_id 类型断言
| 字段 | 内容 |
|------|------|
| 现象 | API 返回 401 未授权 |
| 根因 | Go JWT 解析后 `user_id` 为 `float64`，但 handler 用 `userID.(int64)` 断言，运行时 panic |
| 来源 | V2.0-iter11 |
| 防护 | `grep -r "userID\.(int64)" src/backend/api/` → 如有结果，必须要求后端修复 |
| 正确做法 | `userID := int64(claims["user_id"].(float64))` |

### T4: 小程序页面路径不存在
| 字段 | 内容 |
|------|------|
| 现象 | E2E 测试 navigateTo 无响应或跳转到错误页 |
| 根因 | 测试脚本引用了不存在的页面路径（如 `pages/teacher-dashboard/index`） |
| 来源 | V2.0-iter7 |
| 防护 | `grep "目标路径" src/frontend/src/app.config.ts` → 不存在则不能使用 |
| 已知错误路径 | `pages/teacher-dashboard/index`（应为 `pages/teacher-students/index`）、`pages/student-dashboard/index`（应为 `pages/home/index`）、`pages/class-share/index`（不存在，功能在 `pages/class-detail/index` 内） |

### T5: 接口路由未注册就跑冒烟
| 字段 | 内容 |
|------|------|
| 现象 | API 返回 404 |
| 根因 | 功能已实现但路由未在 `router.go` 中注册 |
| 来源 | V2.0-iter11 |
| 防护 | `grep "目标路径" src/backend/api/router.go` → 不存在则告知后端补注册 |

---

## P1 陷阱（应当避免，大概率导致失败）

### T6: 测试用例间存在隐式数据依赖
| 字段 | 内容 |
|------|------|
| 现象 | 用例执行顺序变化时随机失败 |
| 根因 | IT-03 依赖 IT-01 创建的用户，但未显式声明 |
| 来源 | V2.0-iter4 |
| 防护 | 用例规划表增加「数据依赖」列，显式声明 |

### T7: 集成测试一次性编写过多用例
| 字段 | 内容 |
|------|------|
| 现象 | 后半部分用例参数错误率高 |
| 根因 | 上下文过长导致后半部分用例质量下降 |
| 来源 | V2.0-iter4 |
| 防护 | 分批编写执行，每批 ≤ 5 个用例 |

### T8: 测试数据库未隔离
| 字段 | 内容 |
|------|------|
| 现象 | 并发测试随机失败 |
| 根因 | 多个测试用例操作同一张表导致数据冲突 |
| 来源 | V2.0-iter4 |
| 防护 | 每个测试函数使用独立的数据库文件或事务回滚 |

### T9: 前端模块在迭代执行中被遗漏
| 字段 | 内容 |
|------|------|
| 现象 | 前端功能缺失，需要回退补开发 |
| 根因 | 主 Agent 只开发了后端就进入了 Phase 3 |
| 来源 | V2.0-iter4 |
| 防护 | Phase 2 结束前必须对照需求文档确认后端+前端模块全部已开发 |

### T10: CSS 选择器凭猜测编写（小程序 E2E）
| 字段 | 内容 |
|------|------|
| 现象 | E2E 测试 `page_getElement` 返回空 / `element_tap` 无效 |
| 根因 | 测试脚本直接猜写了 `.chat-page__teacher-avatar`，实际组件无此类名 |
| 来源 | V2.0-iter9 |
| 防护 | 先用 Minium `mini_test.app.get_element()` 查看页面结构，再确定正确选择器 |

### T11: Storage userInfo 格式与前端不一致
| 字段 | 内容 |
|------|------|
| 现象 | API 正常但前端页面显示"教师没有班级"等数据缺失 |
| 根因 | 测试脚本设置的 `Storage.userInfo` 缺少 `persona_id` 等字段，前端读取时为 undefined |
| 来源 | V2.0-iter9 |
| 防护 | 查看 `src/frontend/src/` 中 `getStorageSync('userInfo')` 的所有使用点，确认字段结构 |

### T12: JWT role 字段值与 handler 比较不一致
| 字段 | 内容 |
|------|------|
| 现象 | API 返回意外状态码（如 409/40015 而非 400/40040） |
| 根因 | JWT token 中 `role` 字段的值格式与 handler 中 `c.Get("role")` 的比较字符串不一致，跳过了角色校验分支 |
| 来源 | V2.0-iter11 |
| 防护 | 解码 token：`JSON.parse(atob(token.split('.')[1]))` 检查 role 字段实际值 |

### T13: Python 服务未启动时 Go 后端未做降级处理
| 字段 | 内容 |
|------|------|
| 现象 | 知识库添加/检索全部失败 |
| 根因 | VectorClient 未实现降级逻辑（Python 不可用时返回空结果） |
| 来源 | V2.0-iter5 |
| 防护 | VectorClient 必须实现降级逻辑 |

---

## P2 陷阱（建议避免，偶发性失败）

### T14: 异步加载等待时间不足
| 字段 | 内容 |
|------|------|
| 现象 | 页面元素未出现，`get_element` 返回空 |
| 根因 | 页面数据通过异步请求加载，测试只等了 2500ms 不够 |
| 来源 | V2.0-iter9 |
| 防护 | 异步数据页面等待 ≥ 5000ms，或使用轮询替代固定 sleep |

### T15: Go test 与 JS test 行为差异
| 字段 | 内容 |
|------|------|
| 现象 | Go 集成测试通过，JS E2E 测试失败 |
| 根因 | Go test 在 Mock 模式下直接操作数据库；JS test 通过真实 HTTP 调用，受 JWT 解析、中间件、CORS 等影响 |
| 来源 | V2.0-iter11 |
| 防护 | 严格遵循 API-first 分层策略，API 层失败不进入 E2E 层 |

### T16: 包级变量导致级联跳过
| 字段 | 内容 |
|------|------|
| 现象 | 大量用例被批量 Skip |
| 根因 | 管理员 token 获取失败（包级变量为空）→ 所有依赖管理员权限的用例被跳过 |
| 来源 | V2.0-iter10 |
| 防护 | 每个用例的 Setup 函数独立获取 token，禁止依赖包级共享变量 |

### T17: 主 Agent 直接编写前端代码
| 字段 | 内容 |
|------|------|
| 现象 | 代码质量不可控，缺少单测和 Review |
| 根因 | 主 Agent 未调用 dev_frontent_agent |
| 来源 | V2.0-iter4 |
| 防护 | 主 Agent 禁止直接编写代码 |

---

## 防护检查清单

开发 Agent 在编码前应检查：

- [ ] 数据库字段变更是否已更新：模型定义 + DDL + INSERT SQL + handler 响应
- [ ] API 参数名是否与 handler 一致
- [ ] 外部服务依赖是否有降级处理
- [ ] 新增路由是否已在 router.go 中注册

ci_test_agent 在编写测试前应检查：

- [ ] 测试用例数据依赖是否显式声明
- [ ] 测试数据库是否隔离
- [ ] 是否分批编写（每批 ≤ 5 个用例）
- [ ] 小程序页面路径是否已在 app.config.ts 注册
- [ ] CSS 选择器是否已通过 Minium 验证

---

**维护说明**：每个迭代发现的新陷阱，应在迭代收尾时更新本清单。

**文档版本**: v2.0.0
**更新日期**: 2026-04-06
