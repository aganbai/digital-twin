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
const path = require('path')

// 配置：微信开发者工具的 CLI 路径（macOS 默认路径）
const DEVTOOLS_PATH = '/Applications/wechatwebdevtools.app/Contents/MacOS/cli'
// 小程序项目路径
const PROJECT_PATH = path.resolve(__dirname, '../')
/** 等待指定毫秒 */
function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

describe('学生流程 E2E 测试', () => {
  let miniProgram
  let page

  beforeAll(async () => {
    // 通过 CLI 启动微信开发者工具并启用自动化
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

    // 验证问候语
    const greeting = await page.$('.home-page__greeting')
    expect(greeting).toBeTruthy()
    const greetingText = await greeting.text()
    console.log('问候语:', greetingText)
    expect(greetingText).toContain('你好')

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

  test('4. 对话页 - 发送消息并收到回复', async () => {
    // 先确保首页有教师
    page = await miniProgram.reLaunch('/pages/home/index')
    await sleep(3000)

    const teacherCards = await page.$$('.teacher-card')
    if (!teacherCards || teacherCards.length === 0) {
      console.log('⚠️ 没有教师，跳过对话测试')
      return
    }

    // 点击第一个教师的"开始对话"按钮
    const chatBtn = await page.$('.teacher-card__btn')
    if (chatBtn) {
      await chatBtn.tap()
    } else {
      // 直接点击教师卡片
      await teacherCards[0].tap()
    }
    await sleep(3000)

    // 验证跳转到对话页
    page = await miniProgram.currentPage()
    console.log('当前页面:', page.path)
    expect(page.path).toBe('pages/chat/index')

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
    await sleep(5000) // 等待 AI 回复（mock 模式很快，api 模式需要更久）

    // 验证消息列表中有消息
    const messages = await page.$$('.chat-bubble')
    console.log(`对话中共 ${messages.length} 条消息`)
    expect(messages.length).toBeGreaterThanOrEqual(2) // 至少有用户消息和 AI 回复

    // 验证 AI 回复内容不为空
    const assistantBubbles = await page.$$('.chat-bubble--assistant')
    if (assistantBubbles.length > 0) {
      const lastReply = assistantBubbles[assistantBubbles.length - 1]
      const replyContent = await lastReply.$('.chat-bubble__content')
      if (replyContent) {
        const text = await replyContent.text()
        console.log('AI 回复:', text.substring(0, 100) + '...')
        expect(text.length).toBeGreaterThan(0)
      }
    }

    console.log('✅ 对话测试通过')
  })
})
