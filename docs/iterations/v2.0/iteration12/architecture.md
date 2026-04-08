# V2.0 迭代12 架构设计文档

> 本文档描述迭代12的模块依赖关系、分层架构、开发顺序和接口设计

---

## 1. 模块依赖关系图

### 1.1 整体架构图

```mermaid
graph TB
    %% 前端模块
    FE001[IT12-FE-001<br/>聊天页中断能力]
    FE002[IT12-FE-002<br/>会话列表侧边栏]
    FE003[IT12-FE-003<br/>聊天页会话入口集成]
    FE004[IT12-FE-004<br/>新会话按钮移动]
    FE005[IT12-FE-005<br/>指令系统]
    
    %% 后端模块
    BE001[IT12-BE-001<br/>SSE流中断接口]
    BE002[IT12-BE-002<br/>指令消息类型支持]
    
    %% 依赖关系
    FE001 --> BE001
    FE003 --> FE002
    FE004 --> FE003
    FE005 --> FE001
    
    %% 数据流向
    FE001 -.->|调用中断接口| BE001
    FE003 -.->|调用会话列表API| BE002
    FE005 -.->|发送指令消息| BE002
    
    %% 样式定义
    classDef frontend fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef backend fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    
    class FE001,FE002,FE003,FE004,FE005 frontend
    class BE001,BE002 backend
```

### 1.2 数据流向图

```mermaid
sequenceDiagram
    participant User as 用户
    participant FE as 前端
    participant BE as 后端
    participant DB as 数据库
    
    %% 中断流程
    User->>FE: 点击停止生成按钮
    FE->>BE: POST /api/chat/stream/:session_id/abort
    BE->>BE: 中断SSE连接
    BE-->>FE: 返回中断成功
    FE->>FE: 重置流式状态
    
    %% 会话切换流程
    User->>FE: 点击会话入口
    FE->>BE: GET /api/sessions?teacher_persona_id=X
    BE->>DB: 查询会话列表
    DB-->>BE: 返回会话数据
    BE-->>FE: 返回会话列表
    FE->>FE: 显示侧边栏
    
    User->>FE: 点击历史会话
    FE->>BE: GET /api/chat/messages?session_id=Y
    BE->>DB: 查询历史消息
    DB-->>BE: 返回消息数据
    BE-->>FE: 返回历史消息
    FE->>FE: 更新消息列表
    
    %% 指令处理流程
    User->>FE: 输入#新会话
    FE->>FE: 识别指令并处理
    FE->>FE: 清空消息列表
    FE->>FE: 重置session_id
```

---

## 2. 分层架构说明

### 2.1 前端分层架构

```mermaid
graph TB
    %% Layer 0: 基础组件层
    L0_StopBtn[StopButton组件]
    L0_SessionItem[SessionItem组件]
    
    %% Layer 1: 业务组件层
    L1_SessionSidebar[SessionSidebar组件<br/>IT12-FE-002]
    L1_CommandHandler[指令处理器<br/>IT12-FE-005]
    
    %% Layer 2: 页面集成层
    L2_ChatPage[聊天页集成<br/>IT12-FE-001, IT12-FE-003, IT12-FE-004]
    
    %% Layer 3: 状态管理层
    L3_Store[Zustand Store<br/>会话状态管理]
    
    %% 依赖关系
    L1_SessionSidebar --> L0_SessionItem
    L2_ChatPage --> L1_SessionSidebar
    L2_ChatPage --> L1_CommandHandler
    L2_ChatPage --> L0_StopBtn
    L2_ChatPage --> L3_Store
    
    classDef layer0 fill:#f5f5f5,stroke:#9e9e9e
    classDef layer1 fill:#e3f2fd,stroke:#1976d2
    classDef layer2 fill:#e8f5e8,stroke:#388e3c
    classDef layer3 fill:#fce4ec,stroke:#c2185b
    
    class L0_StopBtn,L0_SessionItem layer0
    class L1_SessionSidebar,L1_CommandHandler layer1
    class L2_ChatPage layer2
    class L3_Store layer3
```

### 2.2 后端分层架构

```mermaid
graph TB
    %% Layer 0: 基础设施层
    L0_DB[数据库层]
    L0_SSE[SSE流式传输]
    
    %% Layer 1: 服务层
    L1_SessionService[会话服务]
    L1_MessageService[消息服务]
    
    %% Layer 2: 控制器层
    L2_ChatCtrl[聊天控制器<br/>IT12-BE-001]
    L2_SessionCtrl[会话控制器]
    
    %% Layer 3: 路由层
    L3_Router[API路由]
    
    %% 依赖关系
    L2_ChatCtrl --> L1_MessageService
    L2_ChatCtrl --> L0_SSE
    L2_SessionCtrl --> L1_SessionService
    L1_SessionService --> L0_DB
    L1_MessageService --> L0_DB
    L3_Router --> L2_ChatCtrl
    L3_Router --> L2_SessionCtrl
    
    classDef layer0 fill:#f5f5f5,stroke:#9e9e9e
    classDef layer1 fill:#e3f2fd,stroke:#1976d2
    classDef layer2 fill:#e8f5e8,stroke:#388e3c
    classDef layer3 fill:#fce4ec,stroke:#c2185b
    
    class L0_DB,L0_SSE layer0
    class L1_SessionService,L1_MessageService layer1
    class L2_ChatCtrl,L2_SessionCtrl layer2
    class L3_Router layer3
```

---

## 3. 后端/前端模块开发顺序

### 3.1 开发阶段划分

| 阶段 | 模块 | 开发顺序 | 并行性 | 依赖关系 |
|------|------|---------|--------|---------|
| **Phase 1** | IT12-BE-001: SSE流中断接口 | 1 | 可并行 | 无依赖 |
| **Phase 1** | IT12-FE-002: 会话列表侧边栏 | 2 | 可并行 | 无依赖 |
| **Phase 1** | IT12-BE-002: 指令消息类型支持 | 3 | 可并行 | 无依赖 |
| **Phase 2** | IT12-FE-001: 聊天页中断能力 | 4 | 串行 | 依赖IT12-BE-001 |
| **Phase 2** | IT12-FE-003: 聊天页会话入口集成 | 5 | 串行 | 依赖IT12-FE-002 |
| **Phase 3** | IT12-FE-004: 新会话按钮移动 | 6 | 串行 | 依赖IT12-FE-003 |
| **Phase 3** | IT12-FE-005: 指令系统 | 7 | 串行 | 依赖IT12-FE-001 |

### 3.2 并行开发策略

```mermaid
gantt
    title 迭代12开发时间线
    dateFormat  YYYY-MM-DD
    section 后端模块
    IT12-BE-001 :done, be001, 2026-04-07, 2d
    IT12-BE-002 :done, be002, 2026-04-07, 1d
    
    section 前端模块
    IT12-FE-002 :done, fe002, 2026-04-07, 2d
    IT12-FE-001 :active, fe001, after be001, 2d
    IT12-FE-003 :active, fe003, after fe002, 2d
    IT12-FE-004 :fe004, after fe003, 1d
    IT12-FE-005 :fe005, after fe001, 1d
    
    section 集成测试
    后端集成测试 :test1, after be001, 1d
    前端集成测试 :test2, after fe003, 2d
    端到端测试 :test3, after fe005, 1d
```

---

## 4. 接口依赖关系

### 4.1 新增接口列表

| 接口 | 方法 | 路径 | 模块 | 状态 |
|------|------|------|------|------|
| 流式中断接口 | GET | `/api/chat/stream/:session_id/abort` | IT12-BE-001 | 新增 |
| 消息类型扩展 | - | `/api/chat/stream` | IT12-BE-002 | 扩展 |
| 消息类型扩展 | - | `/api/chat` | IT12-BE-002 | 扩展 |

### 4.2 接口调用关系

```mermaid
graph LR
    %% 前端组件
    FE_StopBtn[StopButton组件]
    FE_SessionSidebar[SessionSidebar组件]
    FE_CommandHandler[指令处理器]
    
    %% 后端接口
    BE_AbortAPI[中断接口<br/>/api/chat/stream/:id/abort]
    BE_SessionsAPI[会话列表接口<br/>/api/sessions]
    BE_ChatAPI[聊天接口<br/>/api/chat/stream]
    
    %% 调用关系
    FE_StopBtn --> BE_AbortAPI
    FE_SessionSidebar --> BE_SessionsAPI
    FE_CommandHandler --> BE_ChatAPI
    
    classDef frontend fill:#e1f5fe,stroke:#01579b
    classDef backend fill:#f3e5f5,stroke:#4a148c
    
    class FE_StopBtn,FE_SessionSidebar,FE_CommandHandler frontend
    class BE_AbortAPI,BE_SessionsAPI,BE_ChatAPI backend
```

### 4.3 接口变更影响分析

| 接口 | 变更类型 | 影响范围 | 兼容性 |
|------|---------|---------|--------|
| `/api/chat/stream/:session_id/abort` | 新增接口 | 前端中断功能 | 无影响 |
| `/api/chat/stream` | 扩展消息类型 | 指令系统 | 向后兼容 |
| `/api/chat` | 扩展消息类型 | 指令系统 | 向后兼容 |
| `/api/sessions` | 复用接口 | 会话列表 | 无变更 |

---

## 5. 数据库表依赖关系

### 5.1 涉及的数据表

| 表名 | 用途 | 变更类型 | 影响模块 |
|------|------|---------|---------|
| `sessions` | 会话表 | 查询 | IT12-FE-002, IT12-FE-003 |
| `session_titles` | 会话标题表 | 查询 | IT12-FE-002, IT12-FE-003 |
| `messages` | 消息表 | 查询/插入 | IT12-BE-002 |
| `streaming_sessions` | 流式会话表 | 更新 | IT12-BE-001 |

### 5.2 数据流图

```mermaid
graph TB
    %% 数据表
    DB_Sessions[sessions表]
    DB_SessionTitles[session_titles表]
    DB_Messages[messages表]
    DB_StreamingSessions[streaming_sessions表]
    
    %% 业务操作
    OP_GetSessions[获取会话列表]
    OP_GetMessages[获取历史消息]
    OP_InsertMessage[插入消息]
    OP_UpdateStreaming[更新流式状态]
    
    %% 依赖关系
    OP_GetSessions --> DB_Sessions
    OP_GetSessions --> DB_SessionTitles
    OP_GetMessages --> DB_Messages
    OP_InsertMessage --> DB_Messages
    OP_UpdateStreaming --> DB_StreamingSessions
    
    classDef table fill:#fff2cc,stroke:#d6b656
    classDef operation fill:#d5e8d4,stroke:#82b366
    
    class DB_Sessions,DB_SessionTitles,DB_Messages,DB_StreamingSessions table
    class OP_GetSessions,OP_GetMessages,OP_InsertMessage,OP_UpdateStreaming operation
```

### 5.3 数据迁移策略

| 数据操作 | 时机 | 影响 | 回滚方案 |
|---------|------|------|---------|
| 无架构变更 | 部署时 | 无影响 | 无需回滚 |
| 会话列表查询优化 | 开发时 | 性能提升 | 保持原有查询 |
| 消息类型扩展 | 部署时 | 支持新功能 | 保持原有类型 |

---

## 6. 技术风险与应对措施

### 6.1 关键技术风险

| 风险点 | 风险等级 | 影响模块 | 应对措施 |
|--------|---------|---------|---------|
| SSE流中断兼容性 | 中 | IT12-BE-001 | 使用标准AbortController，充分测试 |
| 会话切换状态同步 | 中 | IT12-FE-003 | Zustand统一状态管理，先保存后切换 |
| 指令识别准确性 | 低 | IT12-FE-005 | 严格正则匹配，前后空格处理 |
| 侧边栏动画性能 | 低 | IT12-FE-002 | CSS transform优化，避免重排 |

### 6.2 性能考虑

| 场景 | 性能指标 | 优化措施 |
|------|---------|---------|
| 会话列表加载 | < 500ms | 分页查询，缓存机制 |
| 侧边栏动画 | 60fps | CSS硬件加速，transform优化 |
| 流式中断响应 | < 100ms | 轻量级接口，快速响应 |
| 指令识别 | < 10ms | 前端预检查，正则优化 |

---

## 7. 部署与发布策略

### 7.1 部署顺序

1. **后端部署** (IT12-BE-001, IT12-BE-002)
   - 部署新增接口
   - 验证接口功能
   - 确保向后兼容

2. **前端部署** (IT12-FE-001 ~ IT12-FE-005)
   - 部署前端组件
   - 验证功能集成
   - 性能测试

3. **集成测试**
   - 端到端功能测试
   - 性能基准测试
   - 兼容性验证

### 7.2 回滚方案

| 组件 | 回滚策略 | 影响范围 |
|------|---------|---------|
| 后端接口 | 版本回退 | 前端中断功能暂时不可用 |
| 前端组件 | 版本回退 | 会话管理功能暂时不可用 |
| 数据库 | 无需回滚 | 无架构变更 |

---

**文档版本**: v1.0  
**创建日期**: 2026-04-07  
**适用迭代**: V2.0 迭代12  
**关联文档**:
- [需求文档](./requirements.md)
- [UI规范文档](./ui_spec.md)
- [模块配置文档](./modules.yaml)