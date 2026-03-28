/**
 * E2E 全局配置 - 共享开发者工具实例
 *
 * 所有测试共享同一个 miniProgram 实例，避免重复启动/关闭开发者工具
 */
const automator = require('miniprogram-automator')
const path = require('path')

const DEVTOOLS_PATH = '/Applications/wechatwebdevtools.app/Contents/MacOS/cli'
const PROJECT_PATH = path.resolve(__dirname, '../')

let miniProgram = null

async function getMiniProgram() {
  if (!miniProgram) {
    miniProgram = await automator.launch({
      cliPath: DEVTOOLS_PATH,
      projectPath: PROJECT_PATH,
      timeout: 120000,
    })
  }
  return miniProgram
}

async function closeMiniProgram() {
  if (miniProgram) {
    await miniProgram.close()
    miniProgram = null
  }
}

/** 等待指定毫秒 */
function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

module.exports = { getMiniProgram, closeMiniProgram, sleep, DEVTOOLS_PATH, PROJECT_PATH }
