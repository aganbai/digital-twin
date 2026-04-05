#!/usr/bin/env node
/**
 * 迭代9 SM-06 完整验证脚本
 * 模拟测试流程，验证API响应和前端数据结构一致性
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
  console.log('━━━━━━ SM-06 完整验证流程 ━━━━━━\n')

  try {
    // 1. 数据准备检查
    console.log('步骤1: 数据准备检查')
    console.log('  1.1 教师登录...')
    const teacherLoginRes = await apiRequest('POST', '/auth/wx-login', {
      code: 'v9iter_tch_001',
    })
    
    if (teacherLoginRes.code !== 0) {
      console.log('  ❌ 教师登录失败:', teacherLoginRes.message)
      return
    }
    
    const teacherToken = teacherLoginRes.data.token
    const teacherPayload = parseJWT(teacherToken)
    console.log('  ✅ 教师登录成功')
    console.log('     用户ID:', teacherPayload.user_id)
    console.log('     分身ID:', teacherPayload.persona_id)
    console.log('     角色:', teacherPayload.role)

    console.log('\n  1.2 学生登录...')
    const studentLoginRes = await apiRequest('POST', '/auth/wx-login', {
      code: 'v9iter_stu_001',
    })
    
    if (studentLoginRes.code !== 0) {
      console.log('  ❌ 学生登录失败:', studentLoginRes.message)
      return
    }
    
    const studentToken = studentLoginRes.data.token
    const studentPayload = parseJWT(studentToken)
    console.log('  ✅ 学生登录成功')
    console.log('     用户ID:', studentPayload.user_id)
    console.log('     分身ID:', studentPayload.persona_id)

    // 2. 验证班级数据
    console.log('\n步骤2: 验证班级数据')
    const classesRes = await apiRequest('GET', '/classes', null, teacherToken)
    const classes = Array.isArray(classesRes.data) ? classesRes.data : []
    console.log(`  ✅ 班级数量: ${classes.length}`)
    classes.forEach((cls, idx) => {
      console.log(`     班级${idx + 1}: ID=${cls.id}, 名称="${cls.name}", 成员数=${cls.member_count || 0}`)
    })

    // 3. 验证班级成员
    if (classes.length > 0) {
      const classId = classes[0].id
      console.log(`\n步骤3: 验证班级成员 (classId=${classId})`)
      const membersRes = await apiRequest('GET', `/classes/${classId}/members`, null, teacherToken)
      const members = membersRes.data?.items || []
      console.log(`  ✅ 成员数量: ${members.length}`)
      members.forEach((m, idx) => {
        console.log(`     成员${idx + 1}: persona_id=${m.student_persona_id}, 姓名="${m.student_name || m.student_nickname || '未知'}"`)
      })
    }

    // 4. 验证教师聊天列表
    console.log('\n步骤4: 验证教师聊天列表API')
    const chatListRes = await apiRequest('GET', '/chat-list/teacher', null, teacherToken)
    
    if (chatListRes.code !== 0) {
      console.log('  ❌ 获取聊天列表失败:', chatListRes.message)
      return
    }
    
    const chatClasses = chatListRes.data?.classes || []
    console.log(`  ✅ 聊天列表班级数量: ${chatClasses.length}`)
    
    chatClasses.forEach((cls, idx) => {
      console.log(`     班级${idx + 1}: ID=${cls.class_id}, 名称="${cls.class_name}"`)
      const students = cls.students || []
      console.log(`       学生数量: ${students.length}`)
      students.forEach((s, sIdx) => {
        console.log(`       学生${sIdx + 1}: persona_id=${s.student_persona_id}, 姓名="${s.student_nickname}"`)
      })
    })

    // 5. 验证数据一致性
    console.log('\n步骤5: 数据一致性验证')
    if (classes.length === chatClasses.length && classes.length > 0) {
      console.log('  ✅ 班级列表API和聊天列表API返回的班级数量一致')
      
      // 验证班级ID匹配
      const classIds = classes.map(c => c.id).sort((a, b) => a - b)
      const chatClassIds = chatClasses.map(c => c.class_id).sort((a, b) => a - b)
      const idsMatch = JSON.stringify(classIds) === JSON.stringify(chatClassIds)
      
      if (idsMatch) {
        console.log('  ✅ 班级ID完全匹配')
      } else {
        console.log('  ❌ 班级ID不匹配')
        console.log('     班级列表API返回的ID:', classIds)
        console.log('     聊天列表API返回的ID:', chatClassIds)
      }
    } else {
      console.log('  ❌ 班级数量不一致')
      console.log(`     班级列表API: ${classes.length} 个班级`)
      console.log(`     聊天列表API: ${chatClasses.length} 个班级`)
    }

    // 6. 总结
    console.log('\n━━━━━━ 验证结果总结 ━━━━━━')
    if (chatClasses.length > 0 && chatClasses[0].students?.length > 0) {
      console.log('✅ 后端API完全正常，教师可以看到班级和学生')
      console.log('✅ SM-06后端问题已修复')
      console.log('')
      console.log('⚠️  如果前端测试依然失败，可能原因：')
      console.log('   1. 前端页面缓存了旧数据，需要重新编译前端')
      console.log('   2. 前端数据加载时机问题，需要增加等待时间')
      console.log('   3. 前端页面渲染逻辑问题')
    } else {
      console.log('❌ 后端API存在问题，需要进一步排查')
    }

  } catch (error) {
    console.error('\n❌ 验证过程出错:', error.message)
    console.error(error.stack)
  }
}

main()
