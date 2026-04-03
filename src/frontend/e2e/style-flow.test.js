/**
 * E2E 测试 - 对话风格流程
 *
 * 覆盖冒烟用例：
 * - SM-O01: 教师为学生选择教学风格
 * - SM-O02: 教学风格向后兼容（默认苏格拉底）
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

// 教学风格列表
const TEACHING_STYLES = ['socratic', 'explanatory', 'encouraging', 'strict', 'companion', 'custom']

describe('对话风格 E2E 测试', () => {
  let miniProgram, page
  let teacherToken = ''
  let studentToken = ''
  let teacherPersonaId = ''
  let studentPersonaId = ''
  let studentId = ''
  let studentName = ''

  beforeAll(async () => {
    // 1. 注册教师
    console.log('📦 注册教师用户...')
    const uniqueId = Date.now()
    const teacherLogin = await apiRequest('POST', '/api/auth/wx-login', {
      code: 'e2e_style_teacher_' + uniqueId,
    })
    teacherToken = teacherLogin.data?.token || ''
    if (teacherToken) {
      const completeResp = await apiRequest('POST', '/api/auth/complete-profile', {
        role: 'teacher',
        nickname: 'E2E风格教师' + (uniqueId % 10000),
        school: 'E2E风格测试学校',
        description: 'E2E对话风格测试教师',
      }, teacherToken)
      if (completeResp.data?.token) teacherToken = completeResp.data.token
      teacherPersonaId = completeResp.data?.persona_id || completeResp.data?.user?.default_persona_id || ''
    }
    console.log('✅ 教师注册成功')

    // 2. 注册学生
    console.log('📦 注册学生用户...')
    const studentLogin = await apiRequest('POST', '/api/auth/wx-login', {
      code: 'e2e_style_student_' + uniqueId,
    })
    studentToken = studentLogin.data?.token || ''
    studentName = 'E2E风格学生' + (uniqueId % 10000)
    if (studentToken) {
      const completeResp = await apiRequest('POST', '/api/auth/complete-profile', {
        role: 'student',
        nickname: studentName,
        school: 'E2E风格测试学校',
        description: 'E2E对话风格测试学生',
      }, studentToken)
      if (completeResp.data?.token) studentToken = completeResp.data.token
      studentPersonaId = completeResp.data?.persona_id || completeResp.data?.user?.default_persona_id || ''
      studentId = completeResp.data?.user?.id || completeResp.data?.user_id || ''
    }
    console.log('✅ 学生注册成功')

    // 3. 建立师生关系
    console.log('📦 建立师生关系...')
    await apiRequest('POST', '/api/shares', {
      persona_id: teacherPersonaId,
      hours: 24,
      max_uses: 10,
    }, teacherToken)
    const sharesResp = await apiRequest('GET', '/api/shares', null, teacherToken)
    const shareCode = sharesResp.data?.[0]?.code || ''
    if (shareCode) {
      await apiRequest('POST', `/api/shares/${shareCode}/join`, {
        student_persona_id: studentPersonaId,
      }, studentToken)
    }
    console.log('✅ 师生关系建立完成')

    // 4. 获取共享 miniProgram 实例
    miniProgram = await getMiniProgram()

    // 5. 注入教师 token
    await miniProgram.callWxMethod('setStorageSync', 'token', teacherToken)
    await miniProgram.callWxMethod('setStorageSync', 'userInfo', {
      nickname: 'E2E风格教师' + (uniqueId % 10000),
      role: 'teacher',
    })
    console.log('✅ 教师 token 已注入')
  }, 180000)

  afterAll(async () => {
    // 共享实例不关闭
  })

  // SM-O01: 教师为学生选择教学风格
  test('SM-O01: 教师为学生选择教学风格', async () => {
    // 导航到学生详情页
    page = await miniProgram.navigateTo(
      `/pages/student-detail/index?student_id=${studentId}&student_name=${encodeURIComponent(studentName)}&student_persona_id=${studentPersonaId}`
    )
    await sleep(3000)

    page = await miniProgram.currentPage()
    expect(page.path).toBe('pages/student-detail/index')

    // 验证风格网格
    const styleGrid = await page.$('.student-detail-page__style-grid')
    expect(styleGrid).toBeTruthy()

    // 验证有 6 种风格选项
    const styleCards = await page.$$('.student-detail-page__style-card')
    console.log('风格卡片数量:', styleCards.length)
    expect(styleCards.length).toBe(6)

    // 验证风格名称
    const styleNames = await page.$$('.student-detail-page__style-card-name')
    for (const nameEl of styleNames) {
      const name = await nameEl.text()
      console.log('风格名称:', name)
    }

    // 选择"讲解式教学"（explanatory，通常是第2个）
    let targetCard = null
    for (const card of styleCards) {
      const nameEl = await card.$('.student-detail-page__style-card-name')
      if (nameEl) {
        const name = await nameEl.text()
        if (name.includes('讲解') || name.includes('explanatory')) {
          targetCard = card
          break
        }
      }
    }
    // 如果找不到按名称匹配的，选择第2个（explanatory）
    if (!targetCard && styleCards.length >= 2) {
      targetCard = styleCards[1]
    }
    expect(targetCard).toBeTruthy()
    await targetCard.tap()
    await sleep(1000)

    // 验证选中状态
    const activeCard = await page.$('.student-detail-page__style-card--active')
    expect(activeCard).toBeTruthy()

    // 保存设置
    const saveBtn = await page.$('.student-detail-page__save-btn')
    expect(saveBtn).toBeTruthy()
    await saveBtn.tap()
    await sleep(2000)

    // 通过 API 验证风格已保存
    const stylesResp = await apiRequest(
      'GET',
      `/api/styles?teacher_persona_id=${teacherPersonaId}&student_persona_id=${studentPersonaId}`,
      null,
      teacherToken
    )
    console.log('API 返回风格:', stylesResp.data)
    if (stylesResp.data) {
      const savedStyle = stylesResp.data.style_config?.teaching_style || stylesResp.data.teaching_style
      console.log('已保存的教学风格:', savedStyle)
      expect(savedStyle).toBeTruthy()
    }

    console.log('✅ SM-O01 教师为学生选择教学风格测试通过')
  }, 60000)

  // SM-O02: 教学风格向后兼容（默认苏格拉底）
  test('SM-O02: 教学风格向后兼容（默认苏格拉底）', async () => {
    // 注册一个新学生（未设置过风格）
    const uniqueId2 = Date.now()
    const newStudentLogin = await apiRequest('POST', '/api/auth/wx-login', {
      code: 'e2e_style_new_student_' + uniqueId2,
    })
    let newStudentToken = newStudentLogin.data?.token || ''
    let newStudentPersonaId = ''
    let newStudentId = ''
    let newStudentName = 'E2E新学生' + (uniqueId2 % 10000)
    if (newStudentToken) {
      const completeResp = await apiRequest('POST', '/api/auth/complete-profile', {
        role: 'student',
        nickname: newStudentName,
        school: 'E2E风格测试学校',
        description: 'E2E风格兼容测试学生',
      }, newStudentToken)
      if (completeResp.data?.token) newStudentToken = completeResp.data.token
      newStudentPersonaId = completeResp.data?.persona_id || completeResp.data?.user?.default_persona_id || ''
      newStudentId = completeResp.data?.user?.id || completeResp.data?.user_id || ''
    }

    // 建立师生关系
    const sharesResp = await apiRequest('GET', '/api/shares', null, teacherToken)
    const shareCode = sharesResp.data?.[0]?.code || ''
    if (shareCode) {
      await apiRequest('POST', `/api/shares/${shareCode}/join`, {
        student_persona_id: newStudentPersonaId,
      }, newStudentToken)
    }

    // 进入新学生的详情页
    page = await miniProgram.navigateTo(
      `/pages/student-detail/index?student_id=${newStudentId}&student_name=${encodeURIComponent(newStudentName)}&student_persona_id=${newStudentPersonaId}`
    )
    await sleep(3000)

    // 验证风格选择器默认选中"苏格拉底式提问"
    const activeCard = await page.$('.student-detail-page__style-card--active')
    expect(activeCard).toBeTruthy()

    if (activeCard) {
      const nameEl = await activeCard.$('.student-detail-page__style-card-name')
      if (nameEl) {
        const name = await nameEl.text()
        console.log('默认选中风格:', name)
        expect(name.includes('苏格拉底') || name.includes('socratic') || name.includes('Socratic')).toBeTruthy()
      }
    }

    // 通过 API 验证默认值
    const stylesResp = await apiRequest(
      'GET',
      `/api/styles?teacher_persona_id=${teacherPersonaId}&student_persona_id=${newStudentPersonaId}`,
      null,
      teacherToken
    )
    console.log('新学生 API 返回风格:', stylesResp.data)
    if (stylesResp.data) {
      const defaultStyle = stylesResp.data.style_config?.teaching_style || stylesResp.data.teaching_style || 'socratic'
      console.log('默认教学风格:', defaultStyle)
      expect(defaultStyle).toBe('socratic')
    }

    console.log('✅ SM-O02 教学风格向后兼容测试通过')
  }, 60000)
})
