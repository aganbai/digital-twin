#!/usr/bin/env node

/**
 * 迭代11核心功能快速验证
 * 重点验证：SM-AD02和SM-AD04
 */

const http = require('http');
const crypto = require('crypto');

const BASE_URL = 'http://localhost:8080';

function request(method, path, data, token = null) {
  return new Promise((resolve, reject) => {
    const options = {
      hostname: 'localhost',
      port: 8080,
      path: path,
      method: method,
      headers: { 'Content-Type': 'application/json' }
    };
    
    if (token) {
      options.headers['Authorization'] = `Bearer ${token}`;
    }
    
    const req = http.request(options, (res) => {
      let body = '';
      res.on('data', chunk => body += chunk);
      res.on('end', () => {
        try {
          resolve({ status: res.statusCode, data: body ? JSON.parse(body) : {} });
        } catch (e) {
          resolve({ status: res.statusCode, data: { raw: body } });
        }
      });
    });
    
    req.on('error', reject);
    req.setTimeout(5000, () => {
      req.destroy();
      reject(new Error('请求超时'));
    });
    
    if (data) {
      req.write(JSON.stringify(data));
    }
    req.end();
  });
}

async function main() {
  console.log('=== 迭代11核心功能快速验证 ===\n');
  
  // 使用已知存在的教师账号登录
  const username = '13800138001';
  const password = 'password123';
  
  console.log('步骤1：登录教师账号...');
  const loginResp = await request('POST', '/api/auth/login', {
    username,
    password
  });
  
  console.log(`  登录响应: ${JSON.stringify(loginResp.data)}`);
  
  if (loginResp.status !== 200) {
    console.log('  ✗ 登录失败，尝试注册新账号...');
    
    // 注册新账号
    const uniqueId = crypto.randomBytes(16).toString('hex');
    const newUsername = `test_${uniqueId}`;
    const regResp = await request('POST', '/api/auth/register', {
      username: newUsername,
      password: 'Test@123456',
      role: 'teacher',
      nickname: '测试教师'
    });
    
    if (regResp.status !== 200) {
      console.log('  ✗ 注册失败，无法继续测试');
      return;
    }
    
    console.log('  ✓ 注册成功');
    var token = regResp.data.data?.token;
  } else {
    console.log('  ✓ 登录成功');
    var token = loginResp.data.data?.token;
  }
  
  console.log(`\n步骤2：测试 SM-AD02 - 教师禁止独立创建分身`);
  const uniqueSuffix = Date.now();
  const createPersonaResp = await request('POST', '/api/personas', {
    role: 'teacher',
    nickname: `测试分身_${uniqueSuffix}`,
    school: `测试学校_${uniqueSuffix}`,
    description: '测试描述'
  }, token);
  
  console.log(`  响应状态: ${createPersonaResp.status}`);
  console.log(`  响应数据: ${JSON.stringify(createPersonaResp.data)}`);
  
  const ad02Pass = createPersonaResp.status === 400 && createPersonaResp.data.code === 40040;
  console.log(`  ${ad02Pass ? '✓' : '✗'} SM-AD02: ${ad02Pass ? '通过' : '失败'}`);
  
  console.log(`\n步骤3：测试 SM-AD01 - 创建班级同步创建分身`);
  const createClassResp = await request('POST', '/api/classes', {
    name: `测试班级_${Date.now()}`,
    description: '测试班级描述',
    persona_nickname: '班级专属分身',
    persona_school: '测试学校',
    persona_description: '班级分身描述',
    is_public: true
  }, token);
  
  console.log(`  响应状态: ${createClassResp.status}`);
  console.log(`  响应数据: ${JSON.stringify(createClassResp.data)}`);
  
  if (createClassResp.status === 200 || createClassResp.status === 201) {
    const personaId = createClassResp.data.data?.persona_id;
    console.log(`  ✓ 班级创建成功，persona_id=${personaId}`);
    
    // 验证分身绑定字段
    const personasResp = await request('GET', '/api/personas', null, token);
    const persona = personasResp.data.data?.personas?.find(p => p.id === personaId);
    
    if (persona) {
      console.log(`  分身信息: bound_class_id=${persona.bound_class_id}, is_public=${persona.is_public}`);
      const ad01Pass = persona.bound_class_id && persona.is_public === true;
      console.log(`  ${ad01Pass ? '✓' : '✗'} SM-AD01: ${ad01Pass ? '通过' : '失败'}`);
    } else {
      console.log('  ✗ 未找到对应的分身');
    }
  } else {
    console.log('  ✗ 班级创建失败');
  }
  
  console.log(`\n步骤4：测试 SM-AD04 - 废弃接口返回404`);
  const tests = [
    { path: '/api/personas/1/switch', method: 'POST' },
    { path: '/api/personas/1/activate', method: 'POST' },
    { path: '/api/personas/1/deactivate', method: 'POST' }
  ];
  
  let ad04Pass = true;
  for (const test of tests) {
    const resp = await request(test.method, test.path, null, token);
    const pass = resp.status === 404;
    ad04Pass = ad04Pass && pass;
    console.log(`  ${pass ? '✓' : '✗'} ${test.path}: ${resp.status}`);
  }
  console.log(`  ${ad04Pass ? '✓' : '✗'} SM-AD04: ${ad04Pass ? '通过' : '失败'}`);
  
  console.log('\n=== 测试完成 ===');
}

main().catch(console.error);
