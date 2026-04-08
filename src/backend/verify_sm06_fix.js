#!/usr/bin/env node
/**
 * 迭代9 SM-06 问题验证脚本
 * 用于验证教师聊天列表API是否能正确返回班级
 */

const http = require('http')

const API_BASE = 'http://localhost:8080/api'

// 辅助函数：发送HTTP请求
function apiRequest(method, path, data, token) {
  return new Promise((resolve, reject) => {
    const url = new URL(API_BASE + path)
    const options = {
      hostname: url.hostname,
      port: url.port || 8080,
      path: url.pathname,
      method: method,
      headers: {
        'Content-Type': 'application/json',
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
          const json = JSON.parse(body)
          resolve(json)
        } catch (e) {
          resolve({ raw: body })
        }
      })
    })

    req.on('error', reject)
    req.setTimeout(10000, () => {
      req.destroy()
      reject(new Error('Request timeout'))
    })

    if (data) {
      req.write(JSON.stringify(data))
    }
    req.end()
  })
}

// 解析JWT token获取payload
function parseJWT(token) {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) return null
    const payload = Buffer.from(parts[1], 'base64').toString('utf8')
    return JSON.parse(payload)
  } catch (e) {
    return null
  }
}

async function main() {
  console.log('━━━━━━ SM-06 验证开始 ━━━━━━\n')

  try {
    // 1. 教师登录
    console.log('1. 教师登录...')
    const loginRes = await apiRequest('POST', '/auth/wx-login', {
      code: 'v9iter_tch_001',
    })

    if (loginRes.code !== 0 || !loginRes.data?.token) {
      console.log('❌ 教师登录失败:', loginRes.message || loginRes)
      return
    }

    const teacherToken = loginRes.data.token
    const teacherPersonaId = loginRes.data.current_persona?.id
    console.log('✅ 教师登录成功')
    console.log('   Token中的persona_id:', parseJWT(teacherToken)?.persona_id)
    console.log('   当前分身ID:', teacherPersonaId)

    // 2. 查询班级列表（用于对比）
    console.log('\n2. 查询班级列表 (GET /api/classes)...')
    const classesRes = await apiRequest('GET', '/classes', null, teacherToken)
    console.log('   返回结果:', JSON.stringify(classesRes, null, 2))

    // 3. 查询教师聊天列表
    console.log('\n3. 查询教师聊天列表 (GET /api/chat-list/teacher)...')
    const chatListRes = await apiRequest('GET', '/chat-list/teacher', null, teacherToken)
    console.log('   返回结果:', JSON.stringify(chatListRes, null, 2))

    // 4. 验证结果
    console.log('\n━━━━━━ 验证结果 ━━━━━━')
    const classes = classesRes.data || []
    const chatClasses = chatListRes.data?.classes || []
    
    console.log(`班级列表API返回: ${classes.length} 个班级`)
    console.log(`聊天列表API返回: ${chatClasses.length} 个班级`)
    
    if (chatClasses.length > 0) {
      console.log('✅ 修复成功！教师聊天列表API能正确返回班级')
    } else if (classes.length > 0 && chatClasses.length === 0) {
      console.log('❌ 问题依然存在：班级列表有数据，但聊天列表返回空')
    } else if (classes.length === 0) {
      console.log('⚠️  教师没有班级数据，需要先运行数据准备脚本')
    }

  } catch (error) {
    console.error('\n❌ 验证过程出错:', error.message)
  }
}

main()
