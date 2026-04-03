/**
 * E2E 测试 - 二维码分享流程
 *
 * 覆盖冒烟用例：
 * - SM-P01: 教师生成分享码二维码
 * - SM-P02: 保存二维码到相册
 *
 * 前置条件：
 * 1. 后端服务已启动：WX_MODE=mock LLM_MODE=mock go run src/cmd/server/main.go
 * 2. 微信开发者工具已安装
 * 3. 小程序已编译：npm run build:weapp
 */

const { getMiniProgram, sleep } = require('./helpers')
const http = require('http')

const API_BASE = 'http://localhost:8080'

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

describe('二维码分享 E2E 测试', () => {
  let miniProgram, page
  let teacherToken = ''
  let teacherPersonaId = ''

  beforeAll(async () => {
    // 1. 注册教师
    console.log('📦 注册教师用户...')
    const uniqueId = Date.now()
    const teacherLogin = await apiRequest('POST', '/api/auth/wx-login', {
      code: 'e2e_qrcode_teacher_' + uniqueId,
    })
    teacherToken = teacherLogin.data?.token || ''
    if (teacherToken) {
      const completeResp = await apiRequest('POST', '/api/auth/complete-profile', {
        role: 'teacher',
        nickname: 'E2E二维码教师' + (uniqueId % 10000),
        school: 'E2E二维码测试学校',
        description: 'E2E二维码分享测试教师',
      }, teacherToken)
      if (completeResp.data?.token) teacherToken = completeResp.data.token
      teacherPersonaId = completeResp.data?.persona_id || completeResp.data?.user?.default_persona_id || ''
    }
    console.log('✅ 教师注册成功')

    // 2. 通过 API 生成分享码
    console.log('📦 生成分享码...')
    await apiRequest('POST', '/api/shares', {
      persona_id: teacherPersonaId,
      hours: 24,
      max_uses: 10,
    }, teacherToken)
    console.log('✅ 分享码已生成')

    // 3. 获取共享 miniProgram 实例
    miniProgram = await getMiniProgram()

    // 4. 注入教师 token
    await miniProgram.callWxMethod('setStorageSync', 'token', teacherToken)
    await miniProgram.callWxMethod('setStorageSync', 'userInfo', {
      nickname: 'E2E二维码教师' + (uniqueId % 10000),
      role: 'teacher',
    })
    console.log('✅ 教师 token 已注入')
  }, 180000)

  afterAll(async () => {
    // 共享实例不关闭
  })

  // SM-P01: 教师生成分享码二维码
  test('SM-P01: 教师生成分享码二维码', async () => {
    // 导航到分享管理页
    page = await miniProgram.navigateTo('/pages/share-manage/index')
    await sleep(3000)

    page = await miniProgram.currentPage()
    expect(page.path).toBe('pages/share-manage/index')

    // 验证标题
    const title = await page.$('.share-manage__title')
    if (title) {
      const titleText = await title.text()
      console.log('页面标题:', titleText)
      expect(titleText).toContain('分享码管理')
    }

    // 验证分享码卡片存在
    const cards = await page.$$('.share-manage__card')
    console.log('分享码卡片数量:', cards.length)
    expect(cards.length).toBeGreaterThanOrEqual(1)

    // 验证分享码文本
    const codeText = await page.$('.share-manage__card-code')
    if (codeText) {
      const code = await codeText.text()
      console.log('分享码:', code)
      expect(code.length).toBeGreaterThan(0)
    }

    // 验证"二维码"按钮存在
    const qrcodeBtn = await page.$('.share-manage__btn--qrcode')
    expect(qrcodeBtn).toBeTruthy()
    const qrcodeBtnText = await qrcodeBtn.text()
    console.log('二维码按钮文本:', qrcodeBtnText)
    expect(qrcodeBtnText).toContain('二维码')

    // 点击"二维码"按钮
    await qrcodeBtn.tap()
    await sleep(2000)

    // 验证二维码区域出现
    const qrcodeArea = await page.$('.share-manage__qrcode-area')
    expect(qrcodeArea).toBeTruthy()

    // 验证 Canvas 存在
    const qrcodeCanvas = await page.$('.share-manage__qrcode-canvas')
    expect(qrcodeCanvas).toBeTruthy()

    // 验证提示文案
    const hint = await page.$('.share-manage__qrcode-hint')
    if (hint) {
      const hintText = await hint.text()
      console.log('二维码提示:', hintText)
      expect(hintText).toContain('学生扫码即可加入')
    }

    // 验证"保存到相册"按钮
    const saveBtn = await page.$('.share-manage__qrcode-save-btn')
    expect(saveBtn).toBeTruthy()
    const saveBtnText = await saveBtn.text()
    console.log('保存按钮文本:', saveBtnText)
    expect(saveBtnText).toContain('保存到相册')

    console.log('✅ SM-P01 教师生成分享码二维码测试通过')
  }, 60000)

  // SM-P02: 保存二维码到相册
  test('SM-P02: 保存二维码到相册', async () => {
    page = await miniProgram.currentPage()

    // 确保二维码已展示（如果从 SM-P01 继续，应该已经展示）
    let qrcodeArea = await page.$('.share-manage__qrcode-area')
    if (!qrcodeArea) {
      // 重新点击二维码按钮
      const qrcodeBtn = await page.$('.share-manage__btn--qrcode')
      if (qrcodeBtn) {
        await qrcodeBtn.tap()
        await sleep(2000)
      }
    }

    // 验证"保存到相册"按钮存在且可点击
    const saveBtn = await page.$('.share-manage__qrcode-save-btn')
    expect(saveBtn).toBeTruthy()

    const saveBtnText = await saveBtn.text()
    console.log('保存到相册按钮:', saveBtnText)
    expect(saveBtnText).toContain('保存到相册')

    // 点击保存按钮（模拟器中可能无法真正保存，验证按钮可点击即可）
    await saveBtn.tap()
    await sleep(2000)

    // 验证按钮仍然存在（未出错）
    const saveBtnAfter = await page.$('.share-manage__qrcode-save-btn')
    expect(saveBtnAfter).toBeTruthy()

    console.log('✅ SM-P02 保存二维码到相册测试通过')
  }, 60000)
})
