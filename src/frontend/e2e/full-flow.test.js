/**
 * E2E 完整流程测试
 *
 * 包含学生流程和教师流程，共享同一个开发者工具实例
 *
 * 前置条件：
 * 1. 后端服务已启动
 * 2. 微信开发者工具未打开（测试会自动启动）
 * 3. 小程序已编译：npm run build:weapp
 */

const automator = require('miniprogram-automator')
const path = require('path')

const DEVTOOLS_PATH = '/Applications/wechatwebdevtools.app/Contents/MacOS/cli'
const PROJECT_PATH = path.resolve(__dirname, '../')

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

let miniProgram
let page

beforeAll(async () => {
  miniProgram = await automator.launch({
    cliPath: DEVTOOLS_PATH,
    projectPath: PROJECT_PATH,
    timeout: 120000,
  })
}, 180000)

afterAll(async () => {
  if (miniProgram) {
    await miniProgram.close()
  }
})

// ==================== 学生流程 ====================
describe('学生流程 E2E 测试', () => {
  test('1. 登录页 - 应显示登录按钮并可点击', async () => {
    page = await miniProgram.reLaunch('/pages/login/index')
    await sleep(2000)

    expect(page.path).toBe('pages/login/index')

    const title = await page.$('.login__title')
    expect(title).toBeTruthy()
    const titleText = await title.text()
    expect(titleText).toContain('AI 数字分身')

    const loginBtn = await page.$('.login__btn')
    expect(loginBtn).toBeTruthy()
    const btnText = await loginBtn.text()
    expect(btnText).toContain('微信登录')

    await loginBtn.tap()
    await sleep(3000)

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
    page = await miniProgram.currentPage()
    if (page.path !== 'pages/role-select/index') {
      console.log('已是老用户，跳过角色选择')
      return
    }

    const title = await page.$('.role-select__title')
    expect(title).toBeTruthy()

    const cards = await page.$$('.role-select__card')
    expect(cards.length).toBe(2)
    await cards[1].tap()
    await sleep(500)

    const studentCard = await page.$('.role-select__card--active')
    expect(studentCard).toBeTruthy()

    const input = await page.$('.role-select__input')
    expect(input).toBeTruthy()
    await input.input('E2E测试学生')
    await sleep(500)

    const submitBtn = await page.$('.role-select__btn')
    expect(submitBtn).toBeTruthy()
    await submitBtn.tap()
    await sleep(3000)

    page = await miniProgram.currentPage()
    console.log('角色选择后跳转到:', page.path)
    expect(page.path).toBe('pages/home/index')
  })

  test('3. 学生首页 - 应显示教师列表', async () => {
    page = await miniProgram.reLaunch('/pages/home/index')
    await sleep(3000)

    expect(page.path).toBe('pages/home/index')

    const greeting = await page.$('.home-page__greeting')
    expect(greeting).toBeTruthy()
    const greetingText = await greeting.text()
    console.log('问候语:', greetingText)
    expect(greetingText).toContain('你好')

    await sleep(2000)

    const teacherList = await page.$('.home-page__list')
    const emptyState = await page.$('.empty')

    if (teacherList) {
      console.log('✅ 教师列表已加载')
      const teacherCards = await page.$$('.teacher-card')
      console.log(`找到 ${teacherCards.length} 位教师`)
      expect(teacherCards.length).toBeGreaterThan(0)
    } else if (emptyState) {
      console.log('⚠️ 暂无教师（需要先通过 curl 注册教师）')
    }
  })

  test('4. 对话页 - 发送消息并收到回复', async () => {
    page = await miniProgram.reLaunch('/pages/home/index')
    await sleep(3000)

    const teacherCards = await page.$$('.teacher-card')
    if (!teacherCards || teacherCards.length === 0) {
      console.log('⚠️ 没有教师，跳过对话测试')
      return
    }

    const chatBtn = await page.$('.teacher-card__btn')
    if (chatBtn) {
      await chatBtn.tap()
    } else {
      await teacherCards[0].tap()
    }
    await sleep(3000)

    page = await miniProgram.currentPage()
    console.log('当前页面:', page.path)
    expect(page.path).toBe('pages/chat/index')

    const emptyText = await page.$('.chat-page__empty-text')
    if (emptyText) {
      const text = await emptyText.text()
      console.log('空状态提示:', text)
    }

    const input = await page.$('.chat-page__input')
    expect(input).toBeTruthy()
    await input.input('你好，请问什么是Python？')
    await sleep(500)

    const sendBtn = await page.$('.chat-page__send-btn')
    expect(sendBtn).toBeTruthy()
    await sendBtn.tap()
    await sleep(5000)

    const messages = await page.$$('.chat-bubble')
    console.log(`对话中共 ${messages.length} 条消息`)
    expect(messages.length).toBeGreaterThanOrEqual(2)

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

// ==================== 教师流程 ====================
describe('教师流程 E2E 测试', () => {
  test('5. 登录页 - 教师微信登录', async () => {
    // 清除存储，模拟新用户
    await miniProgram.callWxMethod('clearStorage')
    await sleep(1000)

    page = await miniProgram.reLaunch('/pages/login/index')
    await sleep(2000)

    expect(page.path).toBe('pages/login/index')

    const loginBtn = await page.$('.login__btn')
    expect(loginBtn).toBeTruthy()
    await loginBtn.tap()
    await sleep(3000)

    page = await miniProgram.currentPage()
    console.log('教师登录后跳转到:', page.path)
  })

  test('6. 角色选择页 - 选择教师角色', async () => {
    page = await miniProgram.currentPage()
    if (page.path !== 'pages/role-select/index') {
      console.log('已是老用户，跳过角色选择')
      return
    }

    const cards = await page.$$('.role-select__card')
    expect(cards.length).toBe(2)
    await cards[0].tap() // 教师是第一个
    await sleep(500)

    const input = await page.$('.role-select__input')
    await input.input('E2E测试教师')
    await sleep(500)

    const submitBtn = await page.$('.role-select__btn')
    await submitBtn.tap()
    await sleep(3000)

    page = await miniProgram.currentPage()
    console.log('角色选择后跳转到:', page.path)
    expect(page.path).toBe('pages/knowledge/index')
  })

  test('7. 知识库管理页 - 验证页面渲染', async () => {
    page = await miniProgram.reLaunch('/pages/knowledge/index')
    await sleep(3000)

    expect(page.path).toBe('pages/knowledge/index')

    const title = await page.$('.knowledge-page__title')
    expect(title).toBeTruthy()
    const titleText = await title.text()
    expect(titleText).toContain('知识库')

    const fab = await page.$('.knowledge-page__fab')
    expect(fab).toBeTruthy()

    const docList = await page.$('.knowledge-page__list')
    const emptyState = await page.$('.empty')

    if (docList) {
      const items = await page.$$('.knowledge-page__item')
      console.log(`知识库中有 ${items.length} 篇文档`)
    } else if (emptyState) {
      console.log('知识库为空，准备添加文档')
    }

    console.log('✅ 知识库页面渲染正常')
  })

  test('8. 添加文档 - 填写表单并提交', async () => {
    page = await miniProgram.reLaunch('/pages/knowledge/index')
    await sleep(2000)

    const fab = await page.$('.knowledge-page__fab')
    expect(fab).toBeTruthy()
    await fab.tap()
    await sleep(2000)

    page = await miniProgram.currentPage()
    console.log('当前页面:', page.path)
    expect(page.path).toBe('pages/knowledge/add')

    const titleInput = await page.$('.knowledge-add-page__title-input')
    expect(titleInput).toBeTruthy()
    await titleInput.input('E2E测试文档')
    await sleep(500)

    const textarea = await page.$('.knowledge-add-page__textarea')
    expect(textarea).toBeTruthy()
    await textarea.input('这是一篇通过E2E自动化测试添加的文档。Python是一种高级编程语言，广泛用于数据分析和人工智能领域。')
    await sleep(500)

    const submitBtn = await page.$('.knowledge-add-page__submit')
    expect(submitBtn).toBeTruthy()
    await submitBtn.tap()
    await sleep(3000)

    page = await miniProgram.currentPage()
    console.log('提交后页面:', page.path)

    if (page.path === 'pages/knowledge/index') {
      console.log('✅ 已返回知识库列表')
      await sleep(2000)
      const items = await page.$$('.knowledge-page__item')
      console.log(`知识库中现有 ${items.length} 篇文档`)
      expect(items.length).toBeGreaterThan(0)
    } else {
      await sleep(2000)
      page = await miniProgram.currentPage()
      console.log('等待后页面:', page.path)
    }

    console.log('✅ 添加文档测试通过')
  })
})
