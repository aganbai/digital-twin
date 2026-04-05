/**
 * 冒烟测试钩子模块 (V2.0)
 *
 * 提供完整的测试生命周期管理：
 * - 环境初始化与重置
 * - 测试数据准备与隔离
 * - 测试结果收集与报告
 * - 测试后清理
 *
 * 使用方式：
 * ```javascript
 * const { SmokeTestHooks } = require('./smoke-test-hooks')
 * const hooks = new SmokeTestHooks()
 *
 * beforeAll(() => hooks.beforeAll(), 180000)
 * afterAll(() => hooks.afterAll())
 * beforeEach(() => hooks.beforeEach())
 * afterEach(() => hooks.afterEach())
 * ```
 */

const automator = require('miniprogram-automator')
const http = require('http')
const fs = require('path')
const path = require('path')

// ========== 配置常量 ==========
const CONFIG = {
  DEVTOOLS_PATH: '/Applications/wechatwebdevtools.app/Contents/MacOS/cli',
  PROJECT_PATH: path.resolve(__dirname, '../'),
  API_BASE: 'http://localhost:8080',
  PYTHON_API_BASE: 'http://localhost:8000',
  SCREENSHOT_DIR: path.resolve(__dirname, '../e2e/screenshots-smoke'),
  RESULT_DIR: path.resolve(__dirname, '../e2e/results-smoke'),
  TIMEOUT_PER_CASE: 60000,
  TIMEOUT_BEFORE_ALL: 180000,

  // 测试账号 Mock Code
  TEST_ACCOUNTS: {
    teacherA: { code: 'smoke_teacher_001', nickname: '冒烟测试教师A', school: '测试学校', role: 'teacher' },
    teacherB: { code: 'smoke_teacher_002', nickname: '冒烟测试教师B', school: '测试学校', role: 'teacher' },
    studentA: { code: 'smoke_student_001', nickname: '冒烟测试学生A', role: 'student' },
    studentB: { code: 'smoke_student_002', nickname: '冒烟测试学生B', role: 'student' },
  },

  // 测试班级配置
  TEST_CLASS: {
    name: '冒烟测试班级',
    description: '自动化冒烟测试专用班级',
  },
}

// ========== 工具函数 ==========

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

/**
 * API 请求封装
 */
function apiRequest(method, urlPath, data, token, baseUrl = CONFIG.API_BASE) {
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => reject(new Error('请求超时(30s)')), 30000)
    const url = new URL(urlPath, baseUrl)
    const postData = data ? JSON.stringify(data) : ''

    const options = {
      hostname: url.hostname,
      port: url.port || (url.protocol === 'https:' ? 443 : 80),
      path: url.pathname + url.search,
      method,
      headers: { 'Content-Type': 'application/json' },
    }
    if (postData) options.headers['Content-Length'] = Buffer.byteLength(postData)
    if (token) options.headers['Authorization'] = `Bearer ${token}`

    const req = http.request(options, (res) => {
      let body = ''
      res.on('data', (chunk) => (body += chunk))
      res.on('end', () => {
        clearTimeout(timeout)
        try {
          const parsed = JSON.parse(body)
          parsed._httpStatus = res.statusCode
          resolve(parsed)
        } catch (e) {
          resolve({ code: -1, _httpStatus: res.statusCode, message: body.trim(), _raw: body })
        }
      })
    })
    req.on('error', (err) => { clearTimeout(timeout); reject(err) })
    if (postData) req.write(postData)
    req.end()
  })
}

// ========== SmokeTestHooks 类 ==========

class SmokeTestHooks {
  constructor(options = {}) {
    this.config = { ...CONFIG, ...options }

    // 运行时状态
    this.miniProgram = null
    this.jsExceptions = []
    this.results = []

    // 测试数据状态
    this.state = {
      // 教师 A
      teacherAToken: '',
      teacherAPersonaId: 0,
      teacherAUserId: 0,

      // 教师 B（权限隔离测试）
      teacherBToken: '',
      teacherBPersonaId: 0,

      // 学生 A
      studentAToken: '',
      studentAPersonaId: 0,
      studentAUserId: 0,

      // 学生 B（班级加入测试）
      studentBToken: '',
      studentBPersonaId: 0,

      // 班级
      testClassId: 0,

      // 分享码
      shareCode: '',

      // 会话
      testSessionId: '',
    }

    // 当前测试用例信息
    this.currentCase = null
    this.caseStartTime = null
  }

  // ========== 核心钩子方法 ==========

  /**
   * beforeAll: 全局初始化
   * - 检查环境依赖
   * - 启动微信开发者工具
   * - 准备测试数据
   */
  async beforeAll() {
    console.log('\n╔══════════════════════════════════════════════════════════╗')
    console.log('║              冒烟测试 - 环境初始化                        ║')
    console.log('╚══════════════════════════════════════════════════════════╝')
    console.log(`开始时间: ${new Date().toISOString()}`)

    // 1. 检查微信开发者工具 CLI
    if (!fs.existsSync(this.config.DEVTOOLS_PATH)) {
      throw new Error(`微信开发者工具 CLI 不存在: ${this.config.DEVTOOLS_PATH}`)
    }
    console.log('✅ 微信开发者工具 CLI 存在')

    // 2. 检查后端服务健康状态
    const healthResp = await apiRequest('GET', '/api/system/health')
    if (healthResp._httpStatus !== 200) {
      throw new Error(`后端服务不可用: HTTP ${healthResp._httpStatus}`)
    }
    console.log(`✅ 后端服务运行中 (${this.config.API_BASE})`)

    // 3. 检查 Python 知识服务（可选）
    try {
      const pythonHealth = await apiRequest('GET', '/health', null, null, this.config.PYTHON_API_BASE)
      if (pythonHealth._httpStatus === 200) {
        console.log(`✅ Python 知识服务运行中 (${this.config.PYTHON_API_BASE})`)
      }
    } catch (e) {
      console.log(`⚠️ Python 知识服务未启动（部分测试可能失败）`)
    }

    // 4. 启动微信开发者工具
    console.log('🔧 启动微信开发者工具...')
    this.miniProgram = await automator.launch({
      cliPath: this.config.DEVTOOLS_PATH,
      projectPath: this.config.PROJECT_PATH,
      timeout: 120000,
    })
    console.log('✅ 开发者工具启动成功')

    // 5. 监听 JS 异常
    this.miniProgram.on('exception', (msg) => {
      this.jsExceptions.push({
        time: new Date().toISOString(),
        case: this.currentCase,
        message: msg,
      })
      console.log(`  ⚠️ JS异常: ${msg}`)
    })

    // 6. 创建截图和结果目录
    if (!fs.existsSync(this.config.SCREENSHOT_DIR)) {
      fs.mkdirSync(this.config.SCREENSHOT_DIR, { recursive: true })
    }
    if (!fs.existsSync(this.config.RESULT_DIR)) {
      fs.mkdirSync(this.config.RESULT_DIR, { recursive: true })
    }
    console.log(`📁 截图目录: ${this.config.SCREENSHOT_DIR}`)
    console.log(`📁 结果目录: ${this.config.RESULT_DIR}`)

    // 7. 准备测试数据
    await this.prepareTestData()

    console.log('\n✅ 环境初始化完成\n')
  }

  /**
   * afterAll: 全局清理
   * - 关闭微信开发者工具
   * - 输出测试报告
   * - 清理测试数据（可选）
   */
  async afterAll() {
    console.log('\n╔══════════════════════════════════════════════════════════╗')
    console.log('║              冒烟测试 - 全局清理                          ║')
    console.log('╚══════════════════════════════════════════════════════════╝')

    // 1. 清理测试数据（可选）
    await this.cleanupTestData()

    // 2. 关闭微信开发者工具
    if (this.miniProgram) {
      await this.miniProgram.close()
      console.log('✅ 开发者工具已关闭')
    }

    // 3. 输出测试报告
    await this.generateReport()

    console.log(`结束时间: ${new Date().toISOString()}`)
  }

  /**
   * beforeEach: 每个测试用例前的准备
   * - 记录用例开始时间
   * - 清除小程序 Storage（可选）
   */
  async beforeEach() {
    this.caseStartTime = Date.now()
  }

  /**
   * afterEach: 每个测试用例后的处理
   * - 收集测试结果
   * - 截图保存（失败时）
   */
  async afterEach(testCase) {
    const duration = Date.now() - this.caseStartTime
    console.log(`    耗时: ${(duration / 1000).toFixed(2)}s`)
  }

  // ========== 数据准备方法 ==========

  /**
   * 准备测试数据
   * - 创建/获取测试账号
   * - 建立师生关系
   * - 创建测试班级
   */
  async prepareTestData() {
    console.log('\n📦 准备测试数据...')

    // 1. 教师A 登录/注册
    await this._prepareTeacher('A')

    // 2. 教师B 登录/注册（权限隔离测试）
    await this._prepareTeacher('B')

    // 3. 学生A 登录/注册
    await this._prepareStudent('A')

    // 4. 学生B 登录/注册
    await this._prepareStudent('B')

    // 5. 建立师生关系（学生A -> 教师A）
    await this._establishTeacherStudentRelation()

    // 6. 创建测试班级并添加学生
    await this._createTestClass()

    // 7. 创建初始聊天记录
    await this._createInitialChat()

    console.log('✅ 测试数据准备完成')
  }

  /**
   * 准备教师账号
   */
  async _prepareTeacher(suffix) {
    const account = this.config.TEST_ACCOUNTS[`teacher${suffix}`]
    console.log(`  准备教师${suffix}账号...`)

    const loginRes = await apiRequest('POST', '/api/auth/wx-login', { code: account.code })

    if (loginRes.code !== 0 || !loginRes.data?.token) {
      throw new Error(`教师${suffix}登录失败: ${loginRes.message}`)
    }

    this.state[`teacher${suffix}Token`] = loginRes.data.token
    this.state[`teacher${suffix}UserId`] = loginRes.data.user_id

    // 检查或创建教师分身
    const currentPersona = loginRes.data.current_persona
    const personas = loginRes.data.personas || []

    if (currentPersona?.role === 'teacher') {
      this.state[`teacher${suffix}PersonaId`] = currentPersona.id
      console.log(`    ✅ 教师${suffix}分身: ${currentPersona.id} (已有)`)
    } else {
      // 查找现有教师分身
      const teacherPersonas = personas.filter(p => p.role === 'teacher')
      if (teacherPersonas.length > 0) {
        teacherPersonas.sort((a, b) => b.id - a.id)
        const personaId = teacherPersonas[0].id

        // 切换到教师分身
        const switchRes = await apiRequest(
          'PUT',
          `/api/personas/${personaId}/switch`,
          null,
          this.state[`teacher${suffix}Token`]
        )

        if (switchRes.code === 0) {
          this.state[`teacher${suffix}Token`] = switchRes.data.token
          this.state[`teacher${suffix}PersonaId`] = personaId
          console.log(`    ✅ 教师${suffix}分身: ${personaId} (切换)`)
        }
      } else {
        // 创建新教师分身
        const timestamp = Date.now()
        const createRes = await apiRequest(
          'POST',
          '/api/auth/complete-profile',
          {
            role: 'teacher',
            nickname: `${account.nickname}_${timestamp}`,
            school: account.school,
            description: '冒烟测试教师分身',
          },
          this.state[`teacher${suffix}Token`]
        )

        if (createRes.code === 0 && createRes.data?.persona_id) {
          this.state[`teacher${suffix}Token`] = createRes.data.token
          this.state[`teacher${suffix}PersonaId`] = createRes.data.persona_id
          console.log(`    ✅ 教师${suffix}分身: ${createRes.data.persona_id} (新建)`)
        } else {
          throw new Error(`教师${suffix}分身创建失败: ${createRes.message}`)
        }
      }
    }
  }

  /**
   * 准备学生账号
   */
  async _prepareStudent(suffix) {
    const account = this.config.TEST_ACCOUNTS[`student${suffix}`]
    console.log(`  准备学生${suffix}账号...`)

    const loginRes = await apiRequest('POST', '/api/auth/wx-login', { code: account.code })

    if (loginRes.code !== 0 || !loginRes.data?.token) {
      throw new Error(`学生${suffix}登录失败: ${loginRes.message}`)
    }

    this.state[`student${suffix}Token`] = loginRes.data.token
    this.state[`student${suffix}UserId`] = loginRes.data.user_id

    // 检查或创建学生分身
    const currentPersona = loginRes.data.current_persona
    const personas = loginRes.data.personas || []

    if (currentPersona?.role === 'student') {
      this.state[`student${suffix}PersonaId`] = currentPersona.id
      console.log(`    ✅ 学生${suffix}分身: ${currentPersona.id} (已有)`)
    } else {
      // 查找现有学生分身
      const studentPersonas = personas.filter(p => p.role === 'student')
      if (studentPersonas.length > 0) {
        studentPersonas.sort((a, b) => b.id - a.id)
        const personaId = studentPersonas[0].id

        // 切换到学生分身
        const switchRes = await apiRequest(
          'PUT',
          `/api/personas/${personaId}/switch`,
          null,
          this.state[`student${suffix}Token`]
        )

        if (switchRes.code === 0) {
          this.state[`student${suffix}Token`] = switchRes.data.token
          this.state[`student${suffix}PersonaId`] = personaId
          console.log(`    ✅ 学生${suffix}分身: ${personaId} (切换)`)
        }
      } else {
        // 创建新学生分身
        const timestamp = Date.now()
        const createRes = await apiRequest(
          'POST',
          '/api/auth/complete-profile',
          {
            role: 'student',
            nickname: `${account.nickname}_${timestamp}`,
          },
          this.state[`student${suffix}Token`]
        )

        if (createRes.code === 0 && createRes.data?.persona_id) {
          this.state[`student${suffix}Token`] = createRes.data.token
          this.state[`student${suffix}PersonaId`] = createRes.data.persona_id
          console.log(`    ✅ 学生${suffix}分身: ${createRes.data.persona_id} (新建)`)
        } else {
          throw new Error(`学生${suffix}分身创建失败: ${createRes.message}`)
        }
      }
    }
  }

  /**
   * 建立师生关系
   */
  async _establishTeacherStudentRelation() {
    console.log('  建立师生关系...')

    // 检查是否已有关系
    const relationsRes = await apiRequest(
      'GET',
      '/api/relations?status=approved',
      null,
      this.state.teacherAToken
    )

    if (relationsRes.code === 0 && relationsRes.data?.items?.length > 0) {
      const hasRelation = relationsRes.data.items.some(
        r => r.student_persona_id === this.state.studentAPersonaId
      )
      if (hasRelation) {
        console.log('    ✅ 师生关系已存在')
        return
      }
    }

    // 创建分享码
    const shareRes = await apiRequest(
      'POST',
      '/api/shares',
      { persona_id: this.state.teacherAPersonaId },
      this.state.teacherAToken
    )

    if (shareRes.code !== 0 || !shareRes.data?.share_code) {
      throw new Error(`创建分享码失败: ${shareRes.message}`)
    }

    this.state.shareCode = shareRes.data.share_code

    // 学生A 加入
    const joinRes = await apiRequest(
      'POST',
      `/api/shares/${this.state.shareCode}/join`,
      { student_persona_id: this.state.studentAPersonaId },
      this.state.studentAToken
    )

    if (joinRes.code !== 0 && !joinRes.message?.includes('已加入')) {
      // 可能需要审批
      const pendingRes = await apiRequest(
        'GET',
        '/api/relations?status=pending',
        null,
        this.state.teacherAToken
      )

      if (pendingRes.code === 0 && pendingRes.data?.items?.length > 0) {
        for (const item of pendingRes.data.items) {
          await apiRequest(
            'PUT',
            `/api/relations/${item.id}/approve`,
            null,
            this.state.teacherAToken
          )
        }
      }
    }

    console.log('    ✅ 师生关系已建立')
  }

  /**
   * 创建测试班级
   */
  async _createTestClass() {
    console.log('  创建测试班级...')

    // 检查是否已有班级
    const classListRes = await apiRequest(
      'GET',
      '/api/classes',
      null,
      this.state.teacherAToken
    )

    const classes = Array.isArray(classListRes.data)
      ? classListRes.data
      : (classListRes.data?.classes || [])

    let existingClass = classes.find(c => c.name === this.config.TEST_CLASS.name)

    if (existingClass) {
      this.state.testClassId = existingClass.id
      console.log(`    ✅ 班级已存在: ${existingClass.id}`)
    } else {
      // 创建新班级
      const createRes = await apiRequest(
        'POST',
        '/api/classes',
        {
          name: this.config.TEST_CLASS.name,
          description: this.config.TEST_CLASS.description,
        },
        this.state.teacherAToken
      )

      if (createRes.code === 0 && createRes.data?.id) {
        this.state.testClassId = createRes.data.id
        console.log(`    ✅ 班级已创建: ${createRes.data.id}`)
      } else {
        throw new Error(`创建班级失败: ${createRes.message}`)
      }
    }

    // 添加学生A到班级
    if (this.state.testClassId && this.state.studentAPersonaId) {
      const addRes = await apiRequest(
        'POST',
        `/api/classes/${this.state.testClassId}/members`,
        { student_persona_id: this.state.studentAPersonaId },
        this.state.teacherAToken
      )

      if (addRes.code === 0 || addRes.message?.includes('已在班级')) {
        console.log('    ✅ 学生A已加入班级')
      } else {
        console.log(`    ⚠️ 添加学生到班级: ${addRes.message}`)
      }
    }
  }

  /**
   * 创建初始聊天记录
   */
  async _createInitialChat() {
    console.log('  创建初始聊天记录...')

    if (!this.state.studentAToken || !this.state.teacherAPersonaId) {
      console.log('    ⚠️ 跳过创建聊天记录：缺少学生token或教师persona')
      return
    }

    const chatRes = await apiRequest(
      'POST',
      '/api/chat',
      {
        teacher_persona_id: this.state.teacherAPersonaId,
        message: '冒烟测试初始消息',
      },
      this.state.studentAToken
    )

    if (chatRes.code === 0) {
      this.state.testSessionId = chatRes.data?.session_id || ''
      console.log('    ✅ 初始聊天记录已创建')
    } else {
      console.log(`    ⚠️ 创建聊天记录: ${chatRes.message}`)
    }
  }

  // ========== 数据清理方法 ==========

  /**
   * 清理测试数据
   */
  async cleanupTestData() {
    console.log('\n🧹 清理测试数据...')

    // 注意：这里可以选择性地清理测试数据
    // 由于测试账号是固定的，可以选择保留数据以便下次测试使用

    // 1. 清理测试班级（可选）
    // if (this.state.testClassId) {
    //   await apiRequest('DELETE', `/api/classes/${this.state.testClassId}`, null, this.state.teacherAToken)
    // }

    // 2. 清理测试文档（可选）
    // ...

    console.log('✅ 数据清理完成（保留测试数据供下次使用）')
  }

  // ========== 辅助方法 ==========

  /**
   * 记录测试结果
   */
  recordResult(caseId, status, detail) {
    const result = {
      caseId,
      status,
      detail,
      timestamp: new Date().toISOString(),
      duration: this.caseStartTime ? Date.now() - this.caseStartTime : 0,
    }
    this.results.push(result)

    const icon = status === 'PASS' ? '✅' : status === 'FAIL' ? '❌' : '⚠️'
    console.log(`  ${icon} ${caseId}: ${detail}`)
    return result
  }

  /**
   * 截图
   */
  async screenshot(name) {
    try {
      const filePath = path.join(this.config.SCREENSHOT_DIR, `${name}.png`)
      await this.miniProgram.screenshot({ path: filePath })
      return filePath
    } catch (e) {
      console.log(`    📸 截图失败: ${e.message}`)
      return null
    }
  }

  /**
   * 等待元素出现
   */
  async waitForElement(page, selector, timeout = 5000) {
    const start = Date.now()
    while (Date.now() - start < timeout) {
      try {
        const element = await page.$(selector)
        if (element) return element
      } catch (e) {
        // 忽略
      }
      await sleep(200)
    }
    return null
  }

  /**
   * 等待文本出现
   */
  async waitForText(page, selector, expectedText, timeout = 5000) {
    const start = Date.now()
    while (Date.now() - start < timeout) {
      try {
        const element = await page.$(selector)
        if (element) {
          const text = await element.text()
          if (text && text.includes(expectedText)) {
            return element
          }
        }
      } catch (e) {
        // 忽略
      }
      await sleep(200)
    }
    return null
  }

  /**
   * 以学生身份登录小程序
   */
  async loginAsStudent(suffix = 'A') {
    console.log(`  登录学生${suffix}账号...`)

    const token = this.state[`student${suffix}Token`]
    const personaId = this.state[`student${suffix}PersonaId`]

    if (!token) {
      throw new Error(`学生${suffix} Token 未初始化`)
    }

    await this.miniProgram.callWxMethod('setStorageSync', 'token', token)
    await this.miniProgram.callWxMethod('setStorageSync', 'userInfo', {
      id: personaId || 0,
      nickname: this.config.TEST_ACCOUNTS[`student${suffix}`].nickname,
      role: 'student',
    })

    await this.miniProgram.reLaunch('/pages/home/index')
    await sleep(1500)

    console.log(`    ✅ 学生${suffix}登录完成`)
  }

  /**
   * 以教师身份登录小程序
   */
  async loginAsTeacher(suffix = 'A') {
    console.log(`  登录教师${suffix}账号...`)

    const token = this.state[`teacher${suffix}Token`]
    const personaId = this.state[`teacher${suffix}PersonaId`]

    if (!token) {
      throw new Error(`教师${suffix} Token 未初始化`)
    }

    await this.miniProgram.callWxMethod('setStorageSync', 'token', token)
    await this.miniProgram.callWxMethod('setStorageSync', 'userInfo', {
      id: personaId || 0,
      nickname: this.config.TEST_ACCOUNTS[`teacher${suffix}`].nickname,
      role: 'teacher',
    })

    await this.miniProgram.reLaunch('/pages/home/index')
    await sleep(1500)

    console.log(`    ✅ 教师${suffix}登录完成`)
  }

  /**
   * 清除小程序 Storage
   */
  async clearStorage() {
    await this.miniProgram.callWxMethod('clearStorage')
    await sleep(300)
  }

  // ========== 报告生成 ==========

  /**
   * 生成测试报告
   */
  async generateReport() {
    const passCount = this.results.filter(r => r.status === 'PASS').length
    const failCount = this.results.filter(r => r.status === 'FAIL').length
    const warnCount = this.results.filter(r => r.status === 'WARN').length
    const skipCount = this.results.filter(r => r.status === 'SKIP').length
    const total = this.results.length
    const passRate = total > 0 ? ((passCount / total) * 100).toFixed(1) : '0'

    console.log('\n')
    console.log('╔══════════════════════════════════════════════════════════╗')
    console.log('║                    测试结果汇总                           ║')
    console.log('╚══════════════════════════════════════════════════════════╝')
    console.log(`总用例数: ${total}`)
    console.log(`通过: ${passCount} ✅`)
    console.log(`失败: ${failCount} ❌`)
    console.log(`警告: ${warnCount} ⚠️`)
    console.log(`跳过: ${skipCount} ⏭️`)
    console.log(`通过率: ${passRate}%`)

    if (this.jsExceptions.length > 0) {
      console.log(`\nJS异常数量: ${this.jsExceptions.length}`)
    }

    // 写入 JSON 报告
    const reportPath = path.join(this.config.RESULT_DIR, `smoke-report-${Date.now()}.json`)
    const report = {
      timestamp: new Date().toISOString(),
      summary: {
        total,
        pass: passCount,
        fail: failCount,
        warn: warnCount,
        skip: skipCount,
        passRate: `${passRate}%`,
      },
      jsExceptions: this.jsExceptions.length,
      results: this.results,
      state: {
        teacherAPersonaId: this.state.teacherAPersonaId,
        studentAPersonaId: this.state.studentAPersonaId,
        testClassId: this.state.testClassId,
      },
    }

    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2))
    console.log(`\n📄 详细报告已保存: ${reportPath}`)

    // 输出详细结果
    console.log('\n详细结果:')
    this.results.forEach(r => {
      const icon = r.status === 'PASS' ? '✅' : r.status === 'FAIL' ? '❌' : r.status === 'SKIP' ? '⏭️' : '⚠️'
      console.log(`  ${icon} ${r.caseId}: ${r.detail}`)
    })

    return report
  }
}

// ========== 导出 ==========

module.exports = {
  SmokeTestHooks,
  CONFIG,
  apiRequest,
  sleep,
}
