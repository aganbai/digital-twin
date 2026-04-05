/**
 * E2E 测试 - 分身流程
 *
 * 覆盖冒烟用例：
 * - SM-B01: 分身选择页 - 多分身切换
 * - SM-B02: 分身概览页 - 查看所有分身统计
 *
 * 前置条件：
 * 1. 后端服务已启动：WX_MODE=mock LLM_MODE=mock go run src/cmd/server/main.go
 * 2. 微信开发者工具已安装
 * 3. 小程序已编译：npm run build:weapp
 */

const automator = require('miniprogram-automator')
const http = require('http')
const path = require('path')

const DEVTOOLS_PATH = '/Applications/wechatwebdevtools.app/Contents/MacOS/cli'
const PROJECT_PATH = path.resolve(__dirname, '../')
const API_BASE = 'http://localhost:8080'

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

function apiRequest(method, urlPath, data, token) {
  return new Promise((resolve, reject) => {
    const url = new URL(urlPath, API_BASE)
    const postData = data ? JSON.stringify(data) : ''
    const options = {
      hostname: url.hostname,
      port: url.port,
      path: url.pathname + (url.search || ''),
      method,
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(postData),
      },
    }
    if (token) options.headers['Authorization'] = `Bearer ${token}`
    const req = http.request(options, (res) => {
      let body = ''
      res.on('data', (chunk) => (body += chunk))
      res.on('end', () => {
        try { resolve(JSON.parse(body)) } catch (e) { reject(new Error(`JSON 解析失败: ${body}`)) }
      })
    })
    req.on('error', reject)
    if (postData) req.write(postData)
    req.end()
  })
}

describe('分身流程 E2E 测试', () => {
  let miniProgram
  let page
  let teacherToken = ''

  beforeAll(async () => {
    // 通过 API 注册教师用户
    console.log('📦 通过 API 注册教师用户...')
    const uniqueId = Date.now()
    const loginResp = await apiRequest('POST', '/api/auth/wx-login', {
      code: 'e2e_persona_flow_' + uniqueId,
    })
    teacherToken = loginResp.data?.token || ''

    if (teacherToken) {
      const completeResp = await apiRequest('POST', '/api/auth/complete-profile', {
        role: 'teacher',
        nickname: 'E2E分身教师' + (uniqueId % 10000),
        school: 'E2E分身测试大学',
        description: 'E2E分身流程测试教师',
      }, teacherToken)
      if (completeResp.data?.token) teacherToken = completeResp.data.token
      console.log('✅ 教师用户注册成功')
    }

    // 启动开发者工具
    miniProgram = await automator.launch({
      cliPath: DEVTOOLS_PATH,
      projectPath: PROJECT_PATH,
      timeout: 120000,
    })

    // 注入教师 token
    await miniProgram.callWxMethod('clearStorage')
    await sleep(500)
    await miniProgram.callWxMethod('setStorageSync', 'token', teacherToken)
    await miniProgram.callWxMethod('setStorageSync', 'userInfo', {
      id: loginResp.data?.user_id || 0,
      nickname: 'E2E分身教师' + (uniqueId % 10000),
      role: 'teacher',
    })
    console.log('✅ 教师 token 已注入')
  }, 180000)

  afterAll(async () => {
    if (miniProgram) await miniProgram.close()
  })

  // SM-B01: 分身选择页 - 多分身切换
  test('SM-B01: 分身选择页 - 多分身切换', async () => {
    page = await miniProgram.reLaunch('/pages/persona-select/index')
    await sleep(3000)

    page = await miniProgram.currentPage()
    expect(page.path).toBe('pages/persona-select/index')

    // 验证分区标题（教师分身 / 学生分身）
    const sections = await page.$$('.persona-select__section')
    console.log(`分身分区数量: ${sections.length}`)
    expect(sections.length).toBeGreaterThanOrEqual(1)

    // 验证分区标题文本
    const sectionTitles = await page.$$('.persona-select__section-title')
    for (const title of sectionTitles) {
      const text = await title.text()
      console.log('分区标题:', text)
    }

    // 验证分身卡片
    const cards = await page.$$('.persona-select__card')
    console.log(`分身卡片数量: ${cards.length}`)
    expect(cards.length).toBeGreaterThanOrEqual(1)

    // 验证卡片包含昵称、学校、描述
    const cardName = await page.$('.persona-select__card-name')
    expect(cardName).toBeTruthy()
    const nameText = await cardName.text()
    console.log('分身昵称:', nameText)
    expect(nameText.length).toBeGreaterThan(0)

    // 点击第一个分身卡片，验证跳转
    await cards[0].tap()
    await sleep(3000)

    page = await miniProgram.currentPage()
    console.log('选择分身后跳转到:', page.path)
    // 教师分身跳转到首页或知识库页
    expect(
      page.path === 'pages/home/index' ||
      page.path === 'pages/knowledge/index'
    ).toBeTruthy()

    console.log('✅ SM-B01 分身选择页测试通过')
  })

  // SM-B02: 分身概览页 - 查看所有分身统计
  test('SM-B02: 分身概览页 - 查看所有分身统计', async () => {
    page = await miniProgram.navigateTo('/pages/persona-overview/index')
    await sleep(3000)

    expect(page.path).toBe('pages/persona-overview/index')

    // 验证标题
    const title = await page.$('.persona-overview__title')
    expect(title).toBeTruthy()
    const titleText = await title.text()
    console.log('概览标题:', titleText)
    // 迭代11：标题改为"我的班级分身"
    expect(titleText).toContain('分身')

    // 验证汇总统计（迭代11：班级数、学生数）
    const summary = await page.$('.persona-overview__summary')
    expect(summary).toBeTruthy()
    const summaryText = await summary.text()
    console.log('汇总统计:', summaryText)
    // 迭代11：统计改为"共 X 个班级"
    expect(summaryText).toContain('个班级')

    // 验证分身卡片列表
    const cards = await page.$$('.persona-overview__card')
    console.log(`分身卡片数量: ${cards.length}`)
    expect(cards.length).toBeGreaterThanOrEqual(1)

    // 验证卡片信息完整：昵称
    const cardName = await page.$('.persona-overview__card-name')
    expect(cardName).toBeTruthy()
    const nameText = await cardName.text()
    console.log('分身名称:', nameText)

    // 验证状态标签（公开/私有）
    const badges = await page.$$('.persona-overview__badge')
    console.log(`状态标签数量: ${badges.length}`)
    expect(badges.length).toBeGreaterThanOrEqual(1)

    // 验证统计数据（迭代11：学生数、文档数）
    const stats = await page.$$('.persona-overview__stat')
    console.log(`统计项数量: ${stats.length}`)
    expect(stats.length).toBeGreaterThanOrEqual(2)

    // 验证"进入管理"按钮
    const enterBtn = await page.$('.persona-overview__card-btn')
    expect(enterBtn).toBeTruthy()
    const enterBtnText = await enterBtn.text()
    console.log('管理按钮:', enterBtnText)
    expect(enterBtnText).toContain('进入管理')

    // 迭代11：验证班级绑定信息展示
    const classInfo = await page.$('.persona-overview__class-info')
    if (classInfo) {
      console.log('✅ 迭代11: 班级绑定信息已展示')
    }

    // 迭代11：不再有"创建新分身"按钮，分身随班级创建
    const createBtn = await page.$('.persona-overview__create-btn')
    expect(createBtn).toBeFalsy()
    console.log('✅ 迭代11: 已移除独立创建分身按钮')

    console.log('✅ SM-B02 分身概览页测试通过')
  })
})
