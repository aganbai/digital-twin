#!/usr/bin/env node
/**
 * 迭代9 SM-06 详细调试脚本
 * 用于诊断教师聊天列表问题
 */

const http = require('http')

const API_BASE = 'http://localhost:8080/api'

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
          resolve(JSON.parse(body))
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
  console.log('━━━━━━ SM-06 详细调试开始 ━━━━━━\n')

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
    const payload = parseJWT(teacherToken)
    
    console.log('✅ 教师登录成功')
    console.log('   用户ID:', payload?.user_id)
    console.log('   Token中的persona_id:', payload?.persona_id)
    console.log('   Token中的role:', payload?.role)
    console.log('   登录返回的当前分身:', loginRes.data.current_persona)

    // 2. 查询数据库中的班级数据
    console.log('\n2. 检查班级数据归属...')
    console.log('   查询班级列表API...')
    
    const classesRes = await apiRequest('GET', '/classes', null, teacherToken)
    const classes = classesRes.data || []
    
    console.log(`   班级列表API返回: ${classes.length} 个班级`)
    classes.forEach((cls, idx) => {
      console.log(`   班级${idx + 1}: ID=${cls.id}, 名称=${cls.name}`)
    })

    // 3. 查询教师聊天列表
    console.log('\n3. 查询教师聊天列表API...')
    const chatListRes = await apiRequest('GET', '/chat-list/teacher', null, teacherToken)
    
    if (chatListRes.code !== 0) {
      console.log('❌ 查询失败:', chatListRes.message)
      return
    }
    
    const chatClasses = chatListRes.data?.classes || []
    console.log(`   聊天列表API返回: ${chatClasses.length} 个班级`)
    
    chatClasses.forEach((cls, idx) => {
      console.log(`   班级${idx + 1}: ID=${cls.class_id}, 名称=${cls.class_name}, 学生数=${cls.students?.length || 0}`)
    })

    // 4. 对比分析
    console.log('\n━━━━━━ 问题诊断 ━━━━━━')
    
    if (classes.length > 0 && chatClasses.length > 0) {
      console.log('✅ 数据正常！两个API都返回了班级数据')
      console.log('   问题可能在前端页面的数据加载或渲染逻辑')
    } else if (classes.length > 0 && chatClasses.length === 0) {
      console.log('❌ 问题确认：班级列表API有数据，但聊天列表API返回空')
      console.log('   可能原因：')
      console.log('   1. HandleGetTeacherChatList中的persona_id获取逻辑有问题')
      console.log('   2. 班级查询SQL的WHERE条件有误')
    } else if (classes.length === 0) {
      console.log('⚠️  数据库中没有班级数据')
      console.log('   建议：重新运行数据准备脚本，确保班级创建成功')
    }

    // 5. 输出详细的API响应数据供前端调试
    console.log('\n━━━━━━ API响应详情 ━━━━━━')
    console.log('班级列表API完整响应:')
    console.log(JSON.stringify(classesRes, null, 2))
    console.log('\n聊天列表API完整响应:')
    console.log(JSON.stringify(chatListRes, null, 2))

  } catch (error) {
    console.error('\n❌ 调试过程出错:', error.message)
  }
}

main()
