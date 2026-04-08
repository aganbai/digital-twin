# V2.0 迭代12 UI规范文档

> 本文档定义迭代12涉及的所有UI变更的详细规范

---

## 一、页面结构与布局

### 1.1 聊天页 (pages/chat) 新增区域

```
┌─────────────────────────────────────────────────────────────┐
│  [会话入口] ← ← 返回  ┃  [头像]  王老师  ┃      [历史] →  │  ← 导航栏
├─────────────────────────────────────────────────────────────┤
│                                                              │
│                        消息列表区域                           │
│                     (原有内容保持不变)                        │
│                                                              │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│  🎤  [输入框......]  😊  +  [发送/停止]                       │  ← 输入栏
└─────────────────────────────────────────────────────────────┘
      └━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┘
                          (状态切换)
┌─────────────────────────────────────────────────────────────┐
│  遮罩层 (rgba(0,0,0,0.5))                                    │
│  ┌─────────────────────────────────┐ ← 侧边栏 (75%宽度)       │
│  │  ┃  + 新会话                     │                         │  ← 侧边栏
│  │  ┃ ━━━━━━━━━━━━━━━━━━━━━━━━━━━   │                         │
│  │  ┃  会话列表                     │                         │
│  │  ┃  • [当前] 关于微积分      15条│                         │
│  │  ┃  • 线性代数讨论        8条 │                         │
│  │  ┃  • 概率论问题          5条 │                         │
│  │  ┃                              │                         │
│  └─────────────────────────────────┘                         │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 导航栏布局调整

| 位置 | 元素 | 状态 | 说明 |
|------|------|------|------|
| 左侧 | [会话入口] | 新增 | 汉堡菜单图标（三横线）☰ |
| 左侧 | ← 返回 | 原有 | 保持不变 |
| 中间 | [头像] + 教师名 | 原有 | 保持不变 |
| 右侧 | [历史] | 原有 | 保持不变 |

### 1.3 输入栏布局调整

| 位置 | 元素 | 状态 A | 状态 B | 说明 |
|------|------|--------|--------|------|
| 左侧 | 语音按钮 | 显示 | 显示 | 保持不变 |
| 中间 | 输入框 | 显示 | 禁用 | 流式进行中禁用 |
| 中间 | Emoji按钮 | 显示 | 禁用 | 流式进行中禁用 |
| 右侧 | +号按钮 | 显示 | 禁用 | 流式进行中禁用 |
| 最右 | 发送按钮 | 显示 | 隐藏 | 正常状态 |
| 最右 | 停止按钮 | 隐藏 | 显示 | 流式进行中 |

---

## 二、交互流程设计

### 2.1 流式中断交互流程

```
┌─────────────────────────────────────────────────────────────┐
│  用户发送消息                                                 │
│         ↓                                                     │
│  [发送按钮] → 变为 [发送中...] 禁用                           │
│         ↓                                                     │
│  AI开始流式回复                                               │
│         ↓                                                     │
│  [停止生成按钮] 出现（红色背景）                              │
│         ↓                                                     │
│  ┌──────────────┐  ┌──────────────┐                         │
│  │ 用户不操作    │  │ 点击停止按钮  │                         │
│  └──────────────┘  └──────────────┘                         │
│         ↓                  ↓                                 │
│  正常完成流程         立即中断流程                             │
│  ↓                   ↓                                       │
│  消息转为完整         调用abort()                              │
│  ↓                   ↓                                       │
│  恢复正常状态         消息转为完整                              │
│                       ↓                                       │
│                       恢复正常状态                              │
│         ↓                  ↓                                 │
│  [发送按钮] 显示       [发送按钮] 显示                         │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 会话列表侧边栏交互流程

```
┌─────────────────────────────────────────────────────────────┐
│  用户点击左上角 [会话入口] 图标                              │
│         ↓                                                     │
│  遮罩层淡入 (fade-in, 200ms)                                 │
│         ↓                                                     │
│  侧边栏从左侧滑入 (slide-in-left, 300ms)                     │
│         ↓                                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ 显示会话列表  │  │ 点击历史会话  │  │ 点击新会话   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│         ↓                  ↓                  ↓            │
│  等待用户操作      加载历史消息       清空消息列表             │
│                       ↓                  ↓                    │
│                 更新session_id      重置session_id           │
│                       ↓                  ↓                    │
│                 切换会话成功        新会话就绪                │
│                       ↓                  ↓                    │
│              ┌────────────────────────────────┐              │
│              │   侧边栏滑出 + 遮罩层淡出        │              │
│              └────────────────────────────────┘              │
└─────────────────────────────────────────────────────────────┘
```

### 2.3 指令执行交互流程

```
┌─────────────────────────────────────────────────────────────┐
│  用户在输入框输入 #新会话                                     │
│         ↓                                                     │
│  用户点击 [发送] 按钮                                         │
│         ↓                                                     │
│  前端检测到指令                                               │
│         ↓                                                     │
│  执行指令逻辑                                                 │
│  ┃ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━                         │
│  ┃ clearMessages()                                           │
│  ┃ setSessionId('')                                          │
│  ┃ Taro.showToast('已创建新会话')                            │
│  ┃ setInputValue('')                                         │
│         ↓                                                     │
│  指令执行完毕，输入框恢复                                     │
│         ↓                                                     │
│  可直接发送消息创建新会话                                     │
└─────────────────────────────────────────────────────────────┘
```

---

## 三、组件设计

### 3.1 SessionSidebar 组件

#### 组件位置
`src/frontend/src/components/SessionSidebar/index.tsx`

#### 组件属性 (Props)

```typescript
interface SessionSidebarProps {
  /** 是否显示侧边栏 */
  visible: boolean
  /** 关闭回调 */
  onClose: () => void
  /** 教师分身ID */
  teacherPersonaId: number
  /** 当前会话ID（用于高亮） */
  currentSessionId?: string
  /** 会话选择回调 */
  onSessionSelect: (sessionId: string) => void
  /** 新增会话回调 */
  onCreateNewSession: () => void
  /** 会话列表数据 */
  sessions: SessionInfo[]
  /** 加载状态 */
  loading: boolean
}
```

#### 组件状态 (State)

```typescript
interface SessionSidebarState {
  /** 是否收起（可选功能） */
  collapsed: boolean
}
```

#### 组件事件

| 事件 | 参数 | 说明 |
|------|------|------|
|onClose()| - | 关闭侧边栏 |
|onSessionSelect(sessionId: string)| sessionId | 选择会话 |
|onCreateNewSession()| - | 创建新会话 |

#### 组件结构

```
SessionSidebar
├── MaskView (遮罩层)
└── SidebarView (侧边栏容器)
    ├── Header
    │   ├── CloseButton
    │   └── Title (会话列表)
    ├── NewSessionButton (+ 新会话)
    ├── Divider
    ├── SessionList
    │   └── SessionItem (循环)
    │       ├── CurrentBadge (当前)
    │       ├── Title
    │       ├── LastMessage
    │       └── MetaInfo (时间/数量)
    └── EmptyView (空状态)
```

---

### 3.2 SessionItem 组件

#### 组件位置
`src/frontend/src/components/SessionSidebar/SessionItem.tsx`

#### 组件属性 (Props)

```typescript
interface SessionItemProps {
  /** 会话信息 */
  session: SessionInfo
  /** 是否为当前会话 */
  isCurrent: boolean
  /** 点击回调 */
  onClick: () => void
}

interface SessionInfo {
  session_id: string
  title: string
  last_message: string
  message_count: number
  updated_at: string
}
```

#### 组件状态
无（纯展示组件）

#### 组件事件

| 事件 | 参数 | 说明 |
|------|------|------|
|onClick()| - | 点击会话项 |

---

### 3.3 StopButton 组件

#### 组件位置
`src/frontend/src/components/StopButton/index.tsx`

#### 组件属性 (Props)

```typescript
interface StopButtonProps {
  /** 是否显示 */
  visible: boolean
  /** 加载状态 */
  loading: boolean
  /** 禁用状态 */
  disabled: boolean
  /** 点击回调 */
  onClick: () => void
}
```

#### 组件状态
无（纯展示组件）

#### 组件样式

| 状态 | 背景色 | 文字色 | 其他 |
|------|--------|--------|------|
| 正常 | #FF4D4F | #FFFFFF | scale: 1 |
| 按下 | #FF7875 | #FFFFFF | scale: 0.95 |
| 禁用 | #D9D9D9 | #8C8C8C | opacity: 0.6 |

---

## 四、样式规范 (Styles)

### 4.1 色彩规范 (Colors)

| 用途 | 变量名 | 色值 | 说明 |
|------|--------|------|------|
| 停止生成按钮 | --color-stop-btn | #FF4D4F | 红色 |
| 停止生成按钮(按下) | --color-stop-btn-active | #FF7875 | 浅红 |
| 停止生成按钮(禁用) | --color-stop-btn-disabled | #D9D9D9 | 灰色 |
| 新会话按钮 | --color-new-session | #1890FF | 品牌蓝 |
| 新会话按钮(按下) | --color-new-session-active | #40A9FF | 浅蓝 |
| 当前会话标识 | --color-current-session | #52C41A | 绿色 |
| 侧边栏遮罩 | --color-sidebar-mask | rgba(0,0,0,0.5) | 半透明黑 |
| 侧边栏背景 | --color-sidebar-bg | #FFFFFF | 白色 |
| 侧边栏边框 | --color-sidebar-border | #F0F0F0 | 浅灰 |
| 会话项背景 | --color-session-bg | #FFFFFF | 白色 |
| 会话项背景(悬停) | --color-session-bg-hover | #FAFAFA | 浅灰 |
| 会话项背景(当前) | --color-session-bg-active | #F6FFED | 浅绿底 |

### 4.2 字体规范 (Typography)

| 用途 | 变量名 | 大小 | 字重 | 行高 | 说明 |
|------|--------|------|------|------|------|
| 侧边栏标题 | --font-sidebar-title | 18px | 600 | 28px | |
| 会话项标题 | --font-session-title | 16px | 500 | 24px | |
| 会话项副标题 | --font-session-subtitle | 12px | 400 | 18px | |
| 当前会话标签 | --font-current-badge | 10px | 500 | 14px | |
| 按钮文字 | --font-btn-text | 16px | 500 | 24px | |

### 4.3 间距规范 (Spacing)

| 用途 | 变量名 | 值 | 说明 |
|------|--------|-----|------|
| 侧边栏外边距 | --spacing-sidebar-padding | 16px | 左右内边距 |
| 会话项内边距 | --spacing-session-padding | 12px 16px | |
| 新会话按钮高度 | --spacing-new-btn-height | 48px | |
| 会话项间距 | --spacing-session-gap | 4px | |
| 侧边栏顶部留白 | --spacing-sidebar-top | 8px | |

### 4.4 圆角规范 (Border Radius)

| 用途 | 变量名 | 值 | 说明 |
|------|--------|-----|------|
| 侧边栏右上角 | --radius-sidebar-tr | 0 | 与导航栏对齐 |
| 侧边栏右下角 | --radius-sidebar-br | 12px | |
| 会话项圆角 | --radius-session-item | 8px | |
| 按钮圆角 | --radius-btn | 8px | |

### 4.5 阴影规范 (Shadows)

| 用途 | 变量名 | 值 | 说明 |
|------|--------|-----|------|
| 侧边栏阴影 | --shadow-sidebar | 0 4px 12px rgba(0,0,0,0.1) | |

### 4.6 动画规范 (Animations)

| 用途 | 持续时间 | 缓动函数 | 说明 |
|------|---------|---------|------|
| 遮罩淡入/淡出 | 200ms | ease-in-out | |
| 侧边栏滑入/滑出 | 300ms | cubic-bezier(0.4, 0, 0.2, 1) | |
| 按钮缩放 | 150ms | ease-out | |

---

## 五、SCSS 样式示例

### 5.1 SessionSidebar 样式

```scss
// src/frontend/src/components/SessionSidebar/index.scss

.session-sidebar {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 1000;
  pointer-events: none;

  &__mask {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: var(--color-sidebar-mask);
    opacity: 0;
    transition: opacity var(--anim-mask-duration) ease-in-out;
    pointer-events: auto;

    &--visible {
      opacity: 1;
    }
  }

  &__sidebar {
    position: absolute;
    top: 0;
    left: 0;
    width: 75%;
    max-width: 400px;
    height: 100%;
    background-color: var(--color-sidebar-bg);
    transform: translateX(-100%);
    transition: transform var(--anim-sidebar-duration) cubic-bezier(0.4, 0, 0.2, 1);
    pointer-events: auto;
    display: flex;
    flex-direction: column;
    box-shadow: var(--shadow-sidebar);
    border-top-right-radius: 0;
    border-bottom-right-radius: var(--radius-sidebar-br);

    &--visible {
      transform: translateX(0);
    }
  }

  &__header {
    padding: var(--spacing-sidebar-padding) var(--spacing-sidebar-padding) var(--spacing-sidebar-top);
    display: flex;
    align-items: center;
    justify-content: space-between;
    border-bottom: 1px solid var(--color-sidebar-border);
  }

  &__title {
    font-size: var(--font-sidebar-title);
    font-weight: 600;
    color: #262626;
  }

  &__close {
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background-color: #F5F5F5;
    font-size: 18px;
    color: #8C8C8C;
  }

  &__new-btn {
    margin: var(--spacing-new-btn-margin);
    padding: 0 var(--spacing-sidebar-padding);
    height: var(--spacing-new-btn-height);
    background-color: var(--color-new-session);
    color: #FFFFFF;
    border-radius: var(--radius-btn);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: var(--font-btn-text);
    font-weight: 500;
    gap: 8px;
    transition: background-color 150ms ease-out, transform 150ms ease-out;

    &:active {
      background-color: var(--color-new-session-active);
      transform: scale(0.98);
    }
  }

  &__divider {
    height: 1px;
    background-color: var(--color-sidebar-border);
    margin: 0 var(--spacing-sidebar-padding) 8px;
  }

  &__list {
    flex: 1;
    overflow-y: auto;
    padding: 0 var(--spacing-sidebar-padding);
  }

  &__empty {
    padding: 48px 16px;
    text-align: center;
    color: #8C8C8C;
    font-size: 14px;
  }
}
```

### 5.2 SessionItem 样式

```scss
// src/frontend/src/components/SessionSidebar/SessionItem.scss

.session-item {
  padding: var(--spacing-session-padding);
  background-color: var(--color-session-bg);
  border-radius: var(--radius-session-item);
  margin-bottom: var(--spacing-session-gap);
  cursor: pointer;
  transition: background-color 150ms ease-out;
  position: relative;

  &:hover {
    background-color: var(--color-session-bg-hover);
  }

  &:active {
    transform: scale(0.98);
  }

  &--active {
    background-color: var(--color-session-bg-active);

    .session-item__badge {
      display: flex;
    }
  }

  &__badge {
    position: absolute;
    top: 8px;
    right: 8px;
    display: none;
    align-items: center;
    gap: 4px;
    padding: 2px 8px;
    background-color: var(--color-current-session);
    color: #FFFFFF;
    border-radius: 12px;
    font-size: var(--font-current-badge);
    font-weight: 500;
  }

  &__icon {
    font-size: 14px;
    margin-right: 8px;
  }

  &__content {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  &__title {
    font-size: var(--font-session-title);
    font-weight: 500;
    color: #262626;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  &__message {
    font-size: var(--font-session-subtitle);
    color: #8C8C8C;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  &__meta {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 4px;
  }

  &__time {
    font-size: var(--font-session-subtitle);
    color: #BFBFBF;
  }

  &__count {
    font-size: var(--font-session-subtitle);
    color: #BFBFBF;
  }
}
```

### 5.3 StopButton 样式

```scss
// src/frontend/src/components/StopButton/index.scss

.stop-btn {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background-color: var(--color-stop-btn);
  color: #FFFFFF;
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 150ms ease-out, transform 150ms ease-out;
  box-shadow: 0 2px 8px rgba(255, 77, 79, 0.3);

  &--disabled {
    background-color: var(--color-stop-btn-disabled);
    color: #8C8C8C;
    cursor: not-allowed;
    box-shadow: none;
    opacity: 0.6;
  }

  &:not(&--disabled):active {
    background-color: var(--color-stop-btn-active);
    transform: scale(0.95);
  }

  &__icon {
    width: 20px;
    height: 20px;
  }

  &__text {
    margin-left: 4px;
  }
}

.stop-btn-mini {
  width: 48px;
  height: 32px;
  border-radius: var(--radius-btn);
  font-size: 14px;
  box-shadow: none;

  &__icon {
    width: 16px;
    height: 16px;
  }
}
```

---

## 六、UI测试用例

### 6.1 侧边栏动画测试

| 用例编号 | 测试项 | 测试步骤 | 预期结果 |
|---------|-------|---------|---------|
| UI-001 | 遮罩淡入 | 1. 点击会话入口<br>2. 观察遮罩 | 遮罩颜色从 rgba(0,0,0,0) 渐变到 rgba(0,0,0,0.5)，时长200ms |
| UI-002 | 侧边栏滑入 | 1. 点击会话入口<br>2. 观察侧边栏 | 侧边栏从左侧(-100%) 滑入到 0，时长300ms |
| UI-003 | 遮罩淡出 | 1. 侧边栏打开<br>2. 点击遮罩 | 遮罩颜色从 rgba(0,0,0,0.5) 渐变到 rgba(0,0,0,0)，时长200ms |
| UI-004 | 侧边栏滑出 | 1. 侧边栏打开<br>2. 点击遮罩 | 侧边栏从 0 滑出到左侧(-100%)，时长300ms |

### 6.2 按钮交互测试

| 用例编号 | 测试项 | 测试步骤 | 预期结果 |
|---------|-------|---------|---------|
| UI-005 | 停止按钮显示 | 1. 发送消息<br>2. 流式进行中 | 显示"停止生成"按钮，发送按钮隐藏 |
| UI-006 | 停止按钮按下 | 1. 停止按钮显示<br>2. 长按或点击 | 背景色变为 #FF7875，缩放至 0.95 |
| UI-007 | 停止按钮恢复 | 1. 中断流式<br>2. 恢复状态 | 显示"发送"按钮，停止按钮隐藏 |
| UI-008 | 新会话按钮交互 | 1. 点击新会话按钮 | 背景色变为 #40A9FF，缩放至 0.95，侧边栏关闭 |

### 6.3 响应式测试

| 用例编号 | 测试项 | 测试环境 | 预期结果 |
|---------|-------|---------|---------|
| UI-009 | 小屏手机 | 屏幕宽度 < 320px | 侧边栏占屏幕 85%，新会话按钮文字简化为"+" |
| UI-010 | 标准手机 | 屏幕宽度 375px | 侧边栏占屏幕 75%，新会话按钮显示完整文字 |
| UI-011 | 大屏手机 | 屏幕宽度 > 414px | 侧边栏最大宽度 400px |

### 6.4 可访问性测试

| 用例编号 | 测试项 | 测试步骤 | 预期结果 |
|---------|-------|---------|---------|
| UI-012 | 触摸目标大小 | 1. 测量按钮点击区域 | 所有可点击元素最小 44x44px |
| UI-013 | 颜色对比度 | 1. 检查文字与背景对比 | 所有文字与背景对比度 ≥ 4.5:1 |
| UI-014 | 当前会话标识 | 1. 查看当前会话项 | 当前会话有明确的"当前"绿色标签 |

---

## 七、图标资源

### 7.1 新增图标清单

| 图标名称 | 用途 | SVG路径 | Unicode |
|---------|------|---------|---------|
| menu | 会话入口（汉堡菜单） | M4 6h16v2H4V6zm0 5h16v2H4v-2zm0 5h16v2H4v-2z | ☰ |
| stop | 停止生成（方形） | M6 6h12v12H6V6z | ■ |
| stop-circle | 停止生成（圆形） | M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 14H9V8h2v8zm4 0h-2V8h2v8z | ⏹ |
| current | 当前会话 | M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z | ✓ |
| plus | 新建 | M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z | + |
| close | 关闭 | M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z | ✕ |

---

## 八、设计稿尺寸建议

### 8.1 关键尺寸

| 元素 | 宽度/高度 | 说明 |
|------|----------|------|
| 侧边栏宽度 | 屏幕宽度 × 75% (max 400px) | |
| 新会话按钮高度 | 48px | |
| 会话项内边距 | 12px (上下) 16px (左右) | |
| 停止按钮直径 | 64px | 圆形 |
| 停止按钮(小) | 48px × 32px | 圆角矩形 |
| 遮罩层z-index | 1000 | |
| 侧边栏z-index | 1001 | |

### 8.2 文字截断

| 元素 | 最大宽度 | 截断方式 |
|------|---------|---------|
| 会话标题 | 200px | text-overflow: ellipsis |
| 最后消息 | 200px | text-overflow: ellipsis |

---

**文档版本**: v1.0
**创建日期**: 2026-04-07
**适用迭代**: V2.0 迭代12
**关联文档**：
- [需求文档](./requirements.md)
- [模块配置文档](./modules.yaml)
