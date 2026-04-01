#!/usr/bin/env node
/**
 * E2E 环境冒烟检查脚本
 *
 * 验证 miniprogram-automator 能否成功连接微信开发者工具
 * 用法: node e2e/smoke-check.js
 */
const automator = require('miniprogram-automator')
const path = require('path')

const DEVTOOLS_PATH = '/Applications/wechatwebdevtools.app/Contents/MacOS/cli'
const PROJECT_PATH = path.resolve(__dirname, '../')

async function main() {
  console.log('========================================')
  console.log('  E2E 环境冒烟检查')
  console.log('========================================')
  console.log('')

  // 1. 检查 CLI 路径
  const fs = require('fs')
  if (!fs.existsSync(DEVTOOLS_PATH)) {
    console.log('❌ 微信开发者工具 CLI 不存在:', DEVTOOLS_PATH)
    process.exit(1)
  }
  console.log('✅ 微信开发者工具 CLI 已找到')

  // 2. 检查 dist 目录
  const distPath = path.join(PROJECT_PATH, 'dist', 'app.json')
  if (!fs.existsSync(distPath)) {
    console.log('❌ 小程序未编译，dist/app.json 不存在')
    console.log('   请先运行: npm run build:weapp')
    process.exit(1)
  }
  console.log('✅ 小程序已编译 (dist/app.json 存在)')

  // 3. 尝试连接微信开发者工具
  console.log('')
  console.log('🔗 正在连接微信开发者工具...')
  let miniProgram = null
  try {
    miniProgram = await automator.launch({
      cliPath: DEVTOOLS_PATH,
      projectPath: PROJECT_PATH,
      timeout: 60000,
    })
    console.log('✅ 成功连接微信开发者工具')

    // 4. 获取当前页面
    const page = await miniProgram.currentPage()
    console.log('✅ 当前页面:', page.path)

    // 5. 获取系统信息
    const systemInfo = await miniProgram.callWxMethod('getSystemInfoSync')
    console.log('✅ 系统信息:', JSON.stringify({
      platform: systemInfo.platform,
      model: systemInfo.model,
      SDKVersion: systemInfo.SDKVersion,
    }))

    // 6. 尝试导航到登录页
    await miniProgram.navigateTo('/pages/login/index')
    await new Promise(r => setTimeout(r, 2000))
    const loginPage = await miniProgram.currentPage()
    console.log('✅ 导航测试:', loginPage.path)

    console.log('')
    console.log('========================================')
    console.log('  ✅ E2E 环境检查全部通过！')
    console.log('========================================')
    console.log('')
    console.log('可以运行以下命令执行 E2E 测试:')
    console.log('  npm run test:e2e           # 全量测试')
    console.log('  npm run test:e2e:teacher   # 教师流程')
    console.log('  npm run test:e2e:student   # 学生流程')
    console.log('')

  } catch (err) {
    console.log('')
    console.log('❌ 连接失败:', err.message)
    console.log('')
    console.log('常见问题排查:')
    console.log('  1. 确认微信开发者工具已打开')
    console.log('  2. 确认已在 设置 → 安全设置 中开启"服务端口"')
    console.log('  3. 确认开发者工具中已打开本项目')
    console.log('  4. 尝试重启微信开发者工具')
    console.log('')
    process.exit(1)
  } finally {
    if (miniProgram) {
      try {
        await miniProgram.close()
      } catch (e) {
        // 忽略关闭错误
      }
    }
  }
}

main()
