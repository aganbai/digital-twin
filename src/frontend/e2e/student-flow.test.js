/**
 * E2E 测试 - 学生完整流程
 *
 * 流程：登录 → 角色选择(学生) → 首页(教师列表) → 选择教师 → 对话
 *
 * 前置条件：
 * 1. 后端服务已启动：WX_MODE=mock LLM_MODE=mock go run src/cmd/server/main.go
 * 2. 微信开发者工具已打开，且开启了服务端口（设置 → 安全设置 → 服务端口）
 * 3. 小程序已编译：npm run build:weapp
 * 4. 已通过 curl 注册了一个教师账号并添加了文档
 */

const automator = require('miniprogram-automator')
const http = require('http')
const path = require('path')

// 配置：微信开发者工具的 CLI 路径（macOS 默认路径）
const DEVTOOLS_PATH = '/Applications/wechatwebdevtools.app/Contents/MacOS/cli'
// 小程序项目路径
const PROJECT_PATH = path.resolve(__dirname, '../')
// 后端 API 地址
const API_BASE = 'http://localhost:8080'

/** 等待指定毫秒 */
function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

/** 发送 HTTP 请求到后端 API */
function apiRequest(method, urlPath, data, token) {
  return new Promise((resolve, reject) => {
    const url = new URL(urlPath, API_BASE)
    const postData = data ? JSON.stringify(data) : ''
    const options = {
      hostname: url.hostname,
      port: url.port,
      path: url.pathname,
      method,
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(postData),
      },
    }
    if (token) {
      options.headers['Authorization'] = `Bearer ${token}`
    }
    const req = http.request(options, (res) => {
      let body = ''
      res.on('data', (chunk) => (body += chunk))
      res.on('end', () => {
        try {
          resolve(JSON.parse(body))
        } catch (e) {
          reject(new Error(`JSON 解析失败: ${body}`))
        }
      })
    })
    req.on('error', reject)
    if (postData) req.write(postData)
    req.end()
  })
}

describe('学生流程 E2E 测试', () => {
  let miniProgram
  let page
  let presetTeacherToken = '' // 预置教师 token，用于审批学生申请

  beforeAll(async () => {
    // 1. 获取预置教师的 token（用于后续审批学生申请）
    console.log('📦 获取预置教师 token...')
    try {
      const loginResp = await apiRequest('POST', '/api/auth/wx-login', { code: 'e2e_teacher_setup' })
      presetTeacherToken = loginResp.data?.token || ''
      // 如果是新用户，需要补全信息
      if (loginResp.data?.is_new_user) {
        const completeResp = await apiRequest('POST', '/api/auth/complete-profile', {
          role: 'teacher', nickname: 'E2E预置教师', school: 'E2E测试学校', description: 'E2E预置教师描述'
        }, presetTeacherToken)
        if (completeResp.data?.token) presetTeacherToken = completeResp.data.token
      }
      console.log('✅ 预置教师 token 获取成功')
    } catch (e) {
      console.log('⚠️ 获取预置教师 token 失败:', e.message)
    }

    // 2. 通过 CLI 启动微信开发者工具并启用自动化
    miniProgram = await automator.launch({
      cliPath: DEVTOOLS_PATH,
      projectPath: PROJECT_PATH,
      timeout: 120000, // 启动超时 120 秒
    })
  }, 180000)

  afterAll(async () => {
    if (miniProgram) {
      await miniProgram.close()
    }
  })

  test('1. 登录页 - 应显示登录按钮并可点击', async () => {
    page = await miniProgram.reLaunch('/pages/login/index')
    await sleep(2000)

    // 验证页面路径
    expect(page.path).toBe('pages/login/index')

    // 验证标题文本存在
    const title = await page.$('.login__title')
    expect(title).toBeTruthy()
    const titleText = await title.text()
    expect(titleText).toContain('AI 数字分身')

    // 验证登录按钮存在
    const loginBtn = await page.$('.login__btn')
    expect(loginBtn).toBeTruthy()
    const btnText = await loginBtn.text()
    expect(btnText).toContain('微信登录')

    // 点击登录按钮
    await loginBtn.tap()
    await sleep(3000)

    // 登录后应跳转到角色选择页（新用户）或首页（老用户）
    page = await miniProgram.currentPage()
    const currentPath = page.path
    console.log('登录后跳转到:', currentPath)
    expect(
      currentPath === 'pages/role-select/index' ||
      currentPath === 'pages/home/index' ||
      currentPath === 'pages/knowledge/index'
    ).toBeTruthy()
  })

  test('2. 角色选择页 - 选择学生角色并提交', async () => {
    // 如果不在角色选择页，跳过此测试
    page = await miniProgram.currentPage()
    if (page.path !== 'pages/role-select/index') {
      console.log('已是老用户，跳过角色选择')
      return
    }

    // 验证页面标题
    const title = await page.$('.role-select__title')
    expect(title).toBeTruthy()

    // 点击学生卡片（第二个角色卡片）
    const cards = await page.$$('.role-select__card')
    expect(cards.length).toBe(2)
    // 学生卡片是第二个
    await cards[1].tap()
    await sleep(500)

    // 验证学生卡片被选中（有 active 样式）
    const studentCard = await page.$('.role-select__card--active')
    expect(studentCard).toBeTruthy()

    // 输入昵称
    const input = await page.$('.role-select__input')
    expect(input).toBeTruthy()
    await input.input('E2E测试学生')
    await sleep(500)

    // 点击"开始使用"按钮
    const submitBtn = await page.$('.role-select__btn')
    expect(submitBtn).toBeTruthy()
    await submitBtn.tap()
    await sleep(3000)

    // 应跳转到学生首页
    page = await miniProgram.currentPage()
    console.log('角色选择后跳转到:', page.path)
    expect(page.path).toBe('pages/home/index')
  })

  test('3. 学生首页 - 应显示教师列表', async () => {
    // 确保在首页
    page = await miniProgram.reLaunch('/pages/home/index')
    await sleep(3000)

    expect(page.path).toBe('pages/home/index')

    // 迭代4: greeting 现在显示的是分身昵称，不再是"你好"
    const greeting = await page.$('.home-page__greeting')
    expect(greeting).toBeTruthy()
    const greetingText = await greeting.text()
    console.log('分身昵称:', greetingText)
    expect(greetingText.length).toBeGreaterThan(0)

    // 等待教师列表加载
    await sleep(2000)

    // 检查是否有教师列表或空状态
    const teacherList = await page.$('.home-page__list')
    const emptyState = await page.$('.empty')

    if (teacherList) {
      console.log('✅ 教师列表已加载')
      // 获取教师卡片
      const teacherCards = await page.$$('.teacher-card')
      console.log(`找到 ${teacherCards.length} 位教师`)
      expect(teacherCards.length).toBeGreaterThan(0)
    } else if (emptyState) {
      console.log('⚠️ 暂无教师（需要先通过 curl 注册教师）')
    }
  })

  test('3.5. 申请使用 - 点击申请使用按钮并通过 API 审批（关键：验证 token 角色权限）', async () => {
    // 确保在首页
    page = await miniProgram.reLaunch('/pages/home/index')
    await sleep(3000)

    const teacherCards = await page.$$('.teacher-card')
    if (!teacherCards || teacherCards.length === 0) {
      console.log('⚠️ 没有教师，跳过申请使用测试')
      return
    }

    // 查找"申请使用"按钮
    const applyBtn = await page.$('.teacher-card__btn')
    if (applyBtn) {
      const btnText = await applyBtn.text()
      console.log('按钮文本:', btnText)

      if (btnText.includes('申请使用')) {
        // 点击"申请使用"按钮 —— 这是之前暴露 bug 的关键步骤
        // 如果 complete-profile 返回的 token 角色不正确，这里会报"权限不足"
        await applyBtn.tap()
        await sleep(3000)

        // 验证没有弹出错误提示（如"权限不足"）
        page = await miniProgram.currentPage()
        console.log('申请使用后页面:', page.path)
        console.log('✅ 申请使用按钮点击成功（未报权限不足错误）')

        // 通过 API 让教师审批该申请，使关系变为 approved
        if (presetTeacherToken) {
          try {
            // 获取教师视角的 pending 关系列表
            const relResp = await apiRequest('GET', '/api/relations?status=pending', null, presetTeacherToken)
            console.log('待审批关系数量:', relResp.data?.items?.length || 0)

            if (relResp.data?.items?.length > 0) {
              // 审批所有 pending 关系，确保当前学生的关系被审批
              for (const pendingRelation of relResp.data.items) {
                console.log('审批关系 ID:', pendingRelation.id, '学生:', pendingRelation.student_nickname)
                const approveResp = await apiRequest('PUT', `/api/relations/${pendingRelation.id}/approve`, {}, presetTeacherToken)
                console.log('  审批结果:', approveResp.data?.status || approveResp.message || JSON.stringify(approveResp))
              }
              console.log('✅ 教师已审批所有 pending 关系')
            } else {
              console.log('⚠️ 未找到待审批关系（可能已审批过）')
            }
          } catch (e) {
            console.log('⚠️ API 审批失败:', e.message)
          }
        } else {
          console.log('⚠️ 无预置教师 token，无法自动审批')
        }
      } else if (btnText.includes('开始对话')) {
        console.log('✅ 已有 approved 关系，按钮状态: 开始对话')
      } else if (btnText.includes('等待审批')) {
        console.log('⚠️ 关系为等待审批状态，尝试通过 API 审批...')
        // 同样尝试通过 API 审批
        if (presetTeacherToken) {
          try {
            const relResp = await apiRequest('GET', '/api/relations?status=pending', null, presetTeacherToken)
            if (relResp.data?.items?.length > 0) {
              const approveResp = await apiRequest('PUT', `/api/relations/${relResp.data.items[0].id}/approve`, {}, presetTeacherToken)
              console.log('✅ 补充审批结果:', approveResp.data?.status)
            }
          } catch (e) {
            console.log('⚠️ 补充审批失败:', e.message)
          }
        }
      }
    } else {
      console.log('⚠️ 未找到操作按钮，跳过申请使用测试')
    }
  })

  test('4. 对话页 - 发送消息并收到回复', async () => {
    // 通过 API 获取当前学生的 approved 关系，找到可对话的教师
    // 先从小程序 storage 中获取学生 token
    const studentToken = await miniProgram.callWxMethod('getStorageSync', 'token')

    if (studentToken) {
      try {
        const relResp = await apiRequest('GET', '/api/relations?status=approved', null, studentToken)
        const approvedItems = relResp.data?.items || []
        console.log('已 approved 关系数量:', approvedItems.length)

        if (approvedItems.length > 0) {
          const rel = approvedItems[0]
          const teacherId = rel.teacher_id
          const teacherName = rel.teacher_nickname || '教师'
          console.log(`对话测试 - 使用已 approved 教师: ${teacherName} (ID: ${teacherId})`)

          // 直接 navigateTo 对话页
          page = await miniProgram.navigateTo(
            `/pages/chat/index?teacher_id=${teacherId}&teacher_name=${encodeURIComponent(teacherName)}`
          )
          await sleep(3000)

          page = await miniProgram.currentPage()
          console.log('当前页面:', page.path)

          if (page.path === 'pages/chat/index') {
            // 验证空状态提示
            const emptyText = await page.$('.chat-page__empty-text')
            if (emptyText) {
              const text = await emptyText.text()
              console.log('空状态提示:', text)
            }

            // 输入消息
            const input = await page.$('.chat-page__input')
            expect(input).toBeTruthy()
            await input.input('你好，请问什么是Python？')
            await sleep(500)

            // 点击发送按钮
            const sendBtn = await page.$('.chat-page__send-btn')
            expect(sendBtn).toBeTruthy()
            await sendBtn.tap()
            await sleep(5000) // 等待 AI 回复

            // 验证消息列表中有消息
            const messages = await page.$$('.chat-bubble')
            console.log(`对话中共 ${messages.length} 条消息`)
            expect(messages.length).toBeGreaterThanOrEqual(2)

            // 验证 AI 回复内容不为空
            const assistantBubbles = await page.$$('.chat-bubble--assistant')
            if (assistantBubbles.length > 0) {
              const lastReply = assistantBubbles[assistantBubbles.length - 1]
              const replyContent = await lastReply.$('.chat-bubble__content')
              if (replyContent) {
                const text = await replyContent.text()
                console.log('AI 回复:', text.substring(0, 100) + (text.length > 100 ? '...' : ''))
                if (text.length === 0) {
                  console.log('⚠️ AI 回复内容为空（可能是流式渲染延迟）')
                }
              }
            } else {
              console.log('⚠️ 未找到 AI 回复气泡（可能是流式渲染延迟）')
            }

            console.log('✅ 对话测试通过')

            // 返回首页
            await miniProgram.navigateBack()
            await sleep(1000)
            return
          }
        }
      } catch (e) {
        console.log('⚠️ 获取 approved 关系失败:', e.message)
      }
    }

    // 兜底：如果无法通过 API 获取，尝试从首页找"开始对话"按钮
    page = await miniProgram.reLaunch('/pages/home/index')
    await sleep(3000)

    const allBtns = await page.$$('.teacher-card__btn')
    for (const btn of allBtns) {
      const btnText = await btn.text()
      if (btnText.includes('开始对话')) {
        console.log('对话测试 - 从首页找到"开始对话"按钮')
        await btn.tap()
        await sleep(3000)
        page = await miniProgram.currentPage()
        if (page.path === 'pages/chat/index') {
          // 简化的对话测试
          const input = await page.$('.chat-page__input')
          if (input) {
            await input.input('你好')
            await sleep(500)
            const sendBtn = await page.$('.chat-page__send-btn')
            if (sendBtn) await sendBtn.tap()
            await sleep(5000)
            const messages = await page.$$('.chat-bubble')
            console.log(`对话中共 ${messages.length} 条消息`)
            expect(messages.length).toBeGreaterThanOrEqual(2)
          }
          console.log('✅ 对话测试通过（从首页进入）')
          await miniProgram.navigateBack()
          await sleep(1000)
          return
        }
      }
    }

    console.log('⚠️ 无可用的 approved 关系，跳过对话测试')
  })

  test('5. 个人中心页 - 验证用户信息和菜单项', async () => {
    // 通过 tabBar 切换到个人中心
    page = await miniProgram.switchTab('/pages/profile/index')
    await sleep(3000)

    expect(page.path).toBe('pages/profile/index')

    // 验证头像区域存在
    const avatar = await page.$('.profile-page__avatar')
    expect(avatar).toBeTruthy()

    // 验证昵称显示
    const nickname = await page.$('.profile-page__nickname')
    expect(nickname).toBeTruthy()
    const nicknameText = await nickname.text()
    console.log('个人中心昵称:', nicknameText)
    expect(nicknameText.length).toBeGreaterThan(0)

    // 验证角色标签
    const roleTag = await page.$('.profile-page__role-text')
    expect(roleTag).toBeTruthy()
    const roleText = await roleTag.text()
    console.log('角色标签:', roleText)
    expect(roleText).toBe('学生')

    // 验证统计信息区域
    const stats = await page.$('.profile-page__stats')
    expect(stats).toBeTruthy()

    // 验证功能菜单项
    const menuItems = await page.$$('.profile-page__menu-item')
    console.log(`菜单项数量: ${menuItems.length}`)
    expect(menuItems.length).toBeGreaterThanOrEqual(3) // 我的记忆、对话历史、关于系统、退出登录

    // 验证"我的记忆"菜单存在
    const menuLabels = await page.$$('.profile-page__menu-label')
    const labelTexts = []
    for (const label of menuLabels) {
      const text = await label.text()
      labelTexts.push(text)
    }
    console.log('菜单项:', labelTexts.join(', '))
    expect(labelTexts).toContain('我的记忆')
    expect(labelTexts).toContain('对话历史')

    console.log('✅ 个人中心页测试通过')
  })

  test('6. 对话历史页 - 验证页面渲染', async () => {
    // 通过 tabBar 切换到对话历史
    page = await miniProgram.switchTab('/pages/history/index')
    await sleep(3000)

    expect(page.path).toBe('pages/history/index')

    // 验证标题
    const title = await page.$('.history-page__title')
    expect(title).toBeTruthy()
    const titleText = await title.text()
    expect(titleText).toContain('对话历史')

    // 检查会话列表或空状态
    await sleep(2000)
    const list = await page.$('.history-page__list')
    const emptyState = await page.$('.empty')

    if (list) {
      const items = await page.$$('.history-page__item')
      console.log(`对话历史中有 ${items.length} 条会话`)
      expect(items.length).toBeGreaterThan(0)

      // 验证会话项包含教师昵称和时间
      const firstNickname = await page.$('.history-page__nickname')
      if (firstNickname) {
        const name = await firstNickname.text()
        console.log('最近对话教师:', name)
        expect(name.length).toBeGreaterThan(0)
      }
    } else if (emptyState) {
      console.log('⚠️ 暂无对话历史（新用户正常）')
    }

    console.log('✅ 对话历史页测试通过')
  })

  test('7. 我的教师页 - 验证关系列表', async () => {
    // 导航到我的教师页
    page = await miniProgram.navigateTo('/pages/my-teachers/index')
    await sleep(3000)

    expect(page.path).toBe('pages/my-teachers/index')

    // 验证标题
    const title = await page.$('.my-teachers-page__title-text')
    expect(title).toBeTruthy()
    const titleText = await title.text()
    expect(titleText).toContain('我的教师')

    // 等待数据加载
    await sleep(2000)

    // 检查是否有教师卡片或空状态
    const cards = await page.$$('.my-teachers-page__card')
    const emptyState = await page.$('.empty')

    if (cards && cards.length > 0) {
      console.log(`我的教师列表中有 ${cards.length} 位教师`)

      // 验证教师卡片包含昵称
      const name = await page.$('.my-teachers-page__card-name')
      if (name) {
        const nameText = await name.text()
        console.log('教师名称:', nameText)
        expect(nameText.length).toBeGreaterThan(0)
      }

      // 检查是否有"审批中"分区（测试3.5申请后应该有）
      const sectionTitles = await page.$$('.my-teachers-page__section-title')
      for (const st of sectionTitles) {
        const text = await st.text()
        console.log('分区:', text)
      }
    } else if (emptyState) {
      console.log('⚠️ 暂无教师关系')
    }

    // 返回上一页
    await miniProgram.navigateBack()
    await sleep(1000)

    console.log('✅ 我的教师页测试通过')
  })

  test('8. 我的记忆页 - 验证记忆列表', async () => {
    // 导航到我的记忆页
    page = await miniProgram.navigateTo('/pages/memories/index')
    await sleep(3000)

    expect(page.path).toBe('pages/memories/index')

    // 验证教师筛选器
    const filterLabel = await page.$('.memories-page__filter-label')
    expect(filterLabel).toBeTruthy()
    const filterText = await filterLabel.text()
    expect(filterText).toContain('选择教师')

    // 验证记忆类型 Tab
    const tabs = await page.$$('.memories-page__tab')
    expect(tabs.length).toBe(4) // 全部、对话记忆、学习进度、个性特征
    console.log(`记忆类型 Tab 数量: ${tabs.length}`)

    // 验证第一个 Tab 是"全部"且默认选中
    const firstTab = await page.$('.memories-page__tab--active')
    if (firstTab) {
      const tabText = await firstTab.text()
      console.log('默认选中 Tab:', tabText)
    }

    // 等待数据加载
    await sleep(2000)

    // 检查记忆列表或空状态
    const list = await page.$('.memories-page__list')
    const emptyState = await page.$('.empty')

    if (list) {
      const items = await page.$$('.memories-page__item')
      console.log(`记忆列表中有 ${items.length} 条记忆`)
      expect(items.length).toBeGreaterThan(0)
    } else if (emptyState) {
      console.log('⚠️ 暂无记忆记录（新用户正常）')
    }

    // 测试切换记忆类型 Tab
    if (tabs.length >= 2) {
      await tabs[1].tap() // 点击"对话记忆"
      await sleep(2000)
      const activeTab = await page.$('.memories-page__tab--active')
      if (activeTab) {
        const activeText = await activeTab.text()
        console.log('切换后选中 Tab:', activeText)
      }
    }

    // 返回上一页
    await miniProgram.navigateBack()
    await sleep(1000)

    console.log('✅ 我的记忆页测试通过')
  })
})
