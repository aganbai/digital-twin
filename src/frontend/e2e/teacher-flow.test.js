/**
 * E2E 测试 - 教师完整流程
 *
 * 流程：登录 → 角色选择(教师) → 知识库管理 → 添加文档
 *
 * 前置条件：
 * 1. 后端服务已启动：WX_MODE=mock LLM_MODE=mock go run src/cmd/server/main.go
 * 2. 微信开发者工具已打开，且开启了服务端口（设置 → 安全设置 → 服务端口）
 * 3. 小程序已编译：npm run build:weapp
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

describe('教师流程 E2E 测试', () => {
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

  test('1. 登录页 - 微信登录', async () => {
    page = await miniProgram.reLaunch('/pages/login/index')
    await sleep(2000)

    // 验证登录页
    expect(page.path).toBe('pages/login/index')

    // 点击登录按钮
    const loginBtn = await page.$('.login__btn')
    expect(loginBtn).toBeTruthy()
    await loginBtn.tap()
    await sleep(3000)

    // 登录后跳转
    page = await miniProgram.currentPage()
    console.log('登录后跳转到:', page.path)
  })

  test('2. 角色选择页 - 选择教师角色', async () => {
    page = await miniProgram.currentPage()
    if (page.path !== 'pages/role-select/index') {
      console.log('已是老用户，跳过角色选择')
      return
    }

    // 点击教师卡片（第一个角色卡片）
    const cards = await page.$$('.role-select__card')
    expect(cards.length).toBe(2)
    await cards[0].tap() // 教师是第一个
    await sleep(500)

    // 输入昵称
    const input = await page.$('.role-select__input')
    await input.input('E2E测试教师')
    await sleep(500)

    // 点击"开始使用"
    const submitBtn = await page.$('.role-select__btn')
    await submitBtn.tap()
    await sleep(3000)

    // 应跳转到知识库管理页
    page = await miniProgram.currentPage()
    console.log('角色选择后跳转到:', page.path)
    expect(page.path).toBe('pages/knowledge/index')
  })

  test('3. 知识库管理页 - 验证页面渲染', async () => {
    // 确保在知识库页面
    page = await miniProgram.reLaunch('/pages/knowledge/index')
    await sleep(3000)

    expect(page.path).toBe('pages/knowledge/index')

    // 验证标题
    const title = await page.$('.knowledge-page__title')
    expect(title).toBeTruthy()
    const titleText = await title.text()
    expect(titleText).toContain('知识库')

    // 验证 FAB 添加按钮存在
    const fab = await page.$('.knowledge-page__fab')
    expect(fab).toBeTruthy()

    // 检查文档列表或空状态
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

  test('4. 添加文档 - 填写表单并提交', async () => {
    // 确保在知识库页面
    page = await miniProgram.reLaunch('/pages/knowledge/index')
    await sleep(2000)

    // 点击 FAB 添加按钮
    const fab = await page.$('.knowledge-page__fab')
    expect(fab).toBeTruthy()
    await fab.tap()
    await sleep(2000)

    // 验证跳转到添加文档页
    page = await miniProgram.currentPage()
    console.log('当前页面:', page.path)
    expect(page.path).toBe('pages/knowledge/add')

    // 输入标题
    const titleInput = await page.$('.knowledge-add-page__title-input')
    expect(titleInput).toBeTruthy()
    await titleInput.input('E2E测试文档')
    await sleep(500)

    // 输入内容
    const textarea = await page.$('.knowledge-add-page__textarea')
    expect(textarea).toBeTruthy()
    await textarea.input('这是一篇通过E2E自动化测试添加的文档。Python是一种高级编程语言，广泛用于数据分析和人工智能领域。')
    await sleep(500)

    // 点击提交按钮
    const submitBtn = await page.$('.knowledge-add-page__submit')
    expect(submitBtn).toBeTruthy()
    await submitBtn.tap()
    await sleep(3000)

    // 提交成功后应返回知识库列表页
    page = await miniProgram.currentPage()
    console.log('提交后页面:', page.path)

    // 可能还在添加页（等待 navigateBack），也可能已返回
    if (page.path === 'pages/knowledge/index') {
      console.log('✅ 已返回知识库列表')

      // 验证新文档出现在列表中
      await sleep(2000)
      const items = await page.$$('.knowledge-page__item')
      console.log(`知识库中现有 ${items.length} 篇文档`)
      expect(items.length).toBeGreaterThan(0)
    } else {
      // 等待 navigateBack
      await sleep(2000)
      page = await miniProgram.currentPage()
      console.log('等待后页面:', page.path)
    }

    console.log('✅ 添加文档测试通过')
  })
})
