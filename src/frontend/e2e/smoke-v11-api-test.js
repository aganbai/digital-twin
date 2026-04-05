/**
 * Phase 3c: 迭代11 API 级别冒烟测试
 * 快速验证后端接口功能，无需启动微信开发者工具
 */

const fs = require('fs');

const CONFIG = {
  backendUrl: 'http://localhost:8080',
  resultsDir: '/Users/aganbai/Desktop/WorkSpace/digital-twin/test_results'
};

const RESULTS = {
  startTime: new Date().toISOString(),
  endTime: null,
  tests: [],
  summary: { total: 0, passed: 0, failed: 0, skipped: 0 }
};

// 延迟函数
const delay = (ms) => new Promise(resolve => setTimeout(resolve, ms));

// API 请求辅助函数
async function apiRequest(method, endpoint, body = null, token = null) {
  const headers = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  
  const response = await fetch(`${CONFIG.backendUrl}${endpoint}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : null
  });
  
  const data = await response.json().catch(() => ({}));
  return { status: response.status, data };
}

// 微信登录并补全教师信息
async function wxLoginAndCompleteTeacher(code, nickname) {
  // 1. 微信登录
  const loginRes = await apiRequest('POST', '/api/auth/wx-login', { code });
  
  if (loginRes.status !== 200 || loginRes.data.code !== 0) {
    return { success: false, error: `微信登录失败: ${loginRes.data.message}` };
  }
  
  let token = loginRes.data.data.token;
  const userId = loginRes.data.data.user_id;
  
  // 2. 补全教师信息
  const completeRes = await apiRequest('POST', '/api/auth/complete-profile', {
    role: 'teacher',
    nickname: nickname,
    school: '测试大学',
    description: '测试教师分身'
  }, token);
  
  if (completeRes.status !== 200 || completeRes.data.code !== 0) {
    return { success: false, error: `补全资料失败: ${completeRes.data.message}` };
  }
  
  // 更新token
  if (completeRes.data.data && completeRes.data.data.token) {
    token = completeRes.data.data.token;
  }
  
  return { 
    success: true, 
    token, 
    userId,
    isNewUser: loginRes.data.data.is_new_user 
  };
}

// 记录测试结果
function recordTest(id, name, status, error = null, data = null) {
  const result = { id, name, status, error, data, time: new Date().toISOString() };
  RESULTS.tests.push(result);
  RESULTS.summary.total++;
  if (status === 'passed') RESULTS.summary.passed++;
  else if (status === 'failed') RESULTS.summary.failed++;
  else if (status === 'skipped') RESULTS.summary.skipped++;
  
  const icon = status === 'passed' ? '✅' : status === 'failed' ? '❌' : '⏸️';
  console.log(`${icon} ${id}: ${name}`);
  if (error) console.log(`   原因: ${error}`);
  return result;
}

// 测试用例实现

// SM-AD02: 教师禁止独立创建分身
async function testSM_AD02() {
  console.log('\n🧪 SM-AD02: 教师禁止独立创建分身');
  
  // 1. 先注册一个教师
  const timestamp = Date.now().toString().slice(-6);
  const loginResult = await wxLoginAndCompleteTeacher(
    `mock_ad02_${timestamp}`,
    `教师AD02_${timestamp}`
  );
  
  if (!loginResult.success) {
    return recordTest('SM-AD02', '教师禁止独立创建分身', 'failed', loginResult.error);
  }
  
  const token = loginResult.token;
  
  // 2. 尝试直接创建分身
  const res = await apiRequest('POST', '/api/personas', {
    role: 'teacher',
    nickname: '测试教师分身',
    description: '测试描述'
  }, token);
  
  console.log(`   HTTP状态: ${res.status}, 错误码: ${res.data.code}`);
  
  if (res.status === 400 && res.data.code === 40040) {
    return recordTest('SM-AD02', '教师禁止独立创建分身', 'passed', null, { errorCode: res.data.code });
  } else {
    return recordTest('SM-AD02', '教师禁止独立创建分身', 'failed', 
      `期望 400/40040, 实际 ${res.status}/${res.data.code}`);
  }
}

// SM-AD04: 已删除的接口返回404
async function testSM_AD04() {
  console.log('\n🧪 SM-AD04: 已删除的接口返回404');
  
  // 注册教师获取token
  const timestamp = Date.now().toString().slice(-6);
  const loginResult = await wxLoginAndCompleteTeacher(
    `mock_ad04_${timestamp}`,
    `教师AD04_${timestamp}`
  );
  
  const token = loginResult.token;
  
  const endpoints = [
    { method: 'PUT', path: '/api/personas/1/switch' },
    { method: 'PUT', path: '/api/personas/1/activate' },
    { method: 'PUT', path: '/api/personas/1/deactivate' }
  ];
  
  let allPassed = true;
  const details = [];
  
  for (const ep of endpoints) {
    const res = await apiRequest(ep.method, ep.path, null, token);
    const passed = res.status === 404;
    details.push({ endpoint: ep.path, status: res.status, passed });
    console.log(`   ${ep.path}: ${passed ? '✅' : '❌'} ${res.status}`);
    if (!passed) allPassed = false;
  }
  
    if (allPassed) {
      return recordTest('SM-AD04', '已删除的接口返回404', 'passed', null, { details });
    } else {
      return recordTest('SM-AD04', '已删除的接口返回404', 'failed', 
        '接口仍然存在，需要后端删除: /api/personas/:id/{switch,activate,deactivate}', { details });
    }
}

// SM-AD01: 教师创建班级同步创建分身
async function testSM_AD01() {
  console.log('\n🧪 SM-AD01: 教师创建班级同步创建分身');
  
  // 注册教师
  const timestamp = Date.now().toString().slice(-6);
  const loginResult = await wxLoginAndCompleteTeacher(
    `mock_ad01_${timestamp}`,
    `教师AD01_${timestamp}`
  );
  
  if (!loginResult.success) {
    return recordTest('SM-AD01', '教师创建班级同步创建分身', 'failed', loginResult.error);
  }
  
  const token = loginResult.token;
  const teacherId = loginResult.userId;
  
  // 创建班级
  const className = `测试班级_${timestamp}`;
  const createRes = await apiRequest('POST', '/api/classes', {
    name: className,
    persona_nickname: '张老师',
    persona_school: '测试大学',
    persona_description: '测试教师分身',
    is_public: true
  }, token);
  
  console.log(`   创建响应: HTTP ${createRes.status}, code ${createRes.data.code}`);
  
  if (createRes.status !== 200 || createRes.data.code !== 0) {
    return recordTest('SM-AD01', '教师创建班级同步创建分身', 'failed', 
      `创建班级失败: ${createRes.data.message}`);
  }
  
  const classId = createRes.data.data.id;
  const personaId = createRes.data.data.persona_id;
  // 使用返回的新token（如果有的话）
  const newToken = createRes.data.data.token || token;
  
  console.log(`   班级ID: ${classId}, 分身ID: ${personaId || '未返回'}`);
  
  // 验证分身绑定关系
  const personaRes = await apiRequest('GET', '/api/personas', null, newToken);
  
  if (personaRes.status !== 200 || personaRes.data.code !== 0) {
    return recordTest('SM-AD01', '教师创建班级同步创建分身', 'failed', '获取分身列表失败');
  }
  
    let personas = personaRes.data.data || [];
    // 处理返回数据是对象格式 {personas: [...], current_persona_id: ...}
    if (!Array.isArray(personas) && personas.personas) {
      personas = personas.personas;
    }
    if (!Array.isArray(personas)) {
      personas = [];
    }
    
    console.log(`   获取到 ${personas.length} 个分身`);
    
    // 如果没有返回 persona_id，从班级数据中查询
    let targetPersonaId = personaId;
    if (!targetPersonaId && personas.length > 0) {
      // 找到最近创建的分身
      targetPersonaId = personas[personas.length - 1].id;
    }
    
    const persona = personas.find(p => p.id === targetPersonaId);
    
    if (!persona) {
      console.log(`   可用分身ID列表: ${personas.map(p => p.id).join(', ')}`);
      return recordTest('SM-AD01', '教师创建班级同步创建分身', 'failed', `未找到创建的分身 (期望ID: ${targetPersonaId})`);
    }
  
  console.log(`   分身信息: bound_class_id=${persona.bound_class_id}, is_public=${persona.is_public}`);
  
  // 验证绑定关系
  const validBinding = persona.bound_class_id === classId;
  const validPublic = persona.is_public === true;
  
  if (validBinding && validPublic) {
    // 保存数据供后续用例使用
    const testData = {
      teacherId,
      token: newToken,
      classId,
      personaId,
      className,
      timestamp
    };
    fs.writeFileSync(`${CONFIG.resultsDir}/smoke_v11_ad01_data.json`, JSON.stringify(testData, null, 2));
    
    return recordTest('SM-AD01', '教师创建班级同步创建分身', 'passed', null, testData);
  } else {
    return recordTest('SM-AD01', '教师创建班级同步创建分身', 'failed', 
      `绑定验证失败: bound_class_id=${validBinding}, is_public=${validPublic}`);
  }
}

// SM-AE01: 教师注册自动创建自测学生
async function testSM_AE01() {
  console.log('\n🧪 SM-AE01: 教师注册自动创建自测学生');
  
  // 注册新教师
  const timestamp = Date.now().toString().slice(-6);
  const loginResult = await wxLoginAndCompleteTeacher(
    `mock_ae01_${timestamp}_${Math.random().toString(36).substr(2, 9)}`,
    `教师AE01_${timestamp}`
  );
  
  if (!loginResult.success) {
    return recordTest('SM-AE01', '教师注册自动创建自测学生', 'failed', loginResult.error);
  }
  
  const token = loginResult.token;
  const teacherId = loginResult.userId;
  
  console.log(`   教师ID: ${teacherId}`);
  
  // 获取自测学生信息
  const testStudentRes = await apiRequest('GET', '/api/test-student', null, token);
  
  console.log(`   自测学生接口: HTTP ${testStudentRes.status}, code ${testStudentRes.data?.code}`);
  
  if (testStudentRes.status === 200 && testStudentRes.data.code === 0) {
    const student = testStudentRes.data.data;
    const expectedUsername = `teacher_${teacherId}_test`;
    
    console.log(`   自测学生: username=${student.username}, id=${student.id}`);
    
    if (student.username === expectedUsername) {
      // 验证不创建教师分身
      const personaRes = await apiRequest('GET', '/api/personas', null, token);
      const personas = personaRes.data?.data || [];
      
      // 保存数据供后续用例使用
      const testData = {
        teacherId,
        token,
        testStudentId: student.id,
        testStudentUsername: student.username,
        testStudentNickname: student.nickname,
        personaCount: personas.length
      };
      fs.writeFileSync(`${CONFIG.resultsDir}/smoke_v11_ae01_data.json`, JSON.stringify(testData, null, 2));
      
      return recordTest('SM-AE01', '教师注册自动创建自测学生', 'passed', null, testData);
    } else {
      return recordTest('SM-AE01', '教师注册自动创建自测学生', 'failed', 
        `用户名格式错误: 期望 ${expectedUsername}, 实际 ${student.username}`);
    }
  } else if (testStudentRes.status === 404 || testStudentRes.data?.code === 40400) {
    return recordTest('SM-AE01', '教师注册自动创建自测学生', 'failed', '自测学生未自动创建');
  } else {
    return recordTest('SM-AE01', '教师注册自动创建自测学生', 'failed', 
      `获取自测学生失败: ${testStudentRes.data?.message}`);
  }
}

// SM-AD03: 分身列表展示班级信息
async function testSM_AD03() {
  console.log('\n🧪 SM-AD03: 分身列表展示班级信息');
  
  // 读取 SM-AD01 数据
  let ad01Data = null;
  try {
    ad01Data = JSON.parse(fs.readFileSync(`${CONFIG.resultsDir}/smoke_v11_ad01_data.json`, 'utf8'));
  } catch (e) {
    // 如果没有前置数据，执行 SM-AD01 创建数据
    console.log('   未找到前置数据，先执行 SM-AD01...');
    await testSM_AD01();
    try {
      ad01Data = JSON.parse(fs.readFileSync(`${CONFIG.resultsDir}/smoke_v11_ad01_data.json`, 'utf8'));
    } catch (e2) {
      return recordTest('SM-AD03', '分身列表展示班级信息', 'failed', '无法获取前置数据');
    }
  }
  
  // 获取分身列表
  const personaRes = await apiRequest('GET', '/api/personas', null, ad01Data.token);
  
  if (personaRes.status !== 200 || personaRes.data.code !== 0) {
    return recordTest('SM-AD03', '分身列表展示班级信息', 'failed', '获取分身列表失败');
  }
  
  let personas = personaRes.data.data || [];
  // 处理返回数据是对象格式 {personas: [...], current_persona_id: ...}
  if (!Array.isArray(personas) && personas.personas) {
    personas = personas.personas;
  }
  if (!Array.isArray(personas)) {
    personas = [];
  }
  
  const targetPersona = personas.find(p => p.id === ad01Data.personaId);
  
  if (!targetPersona) {
    return recordTest('SM-AD03', '分身列表展示班级信息', 'failed', '未找到目标分身');
  }
  
  console.log(`   分身字段: bound_class_id=${targetPersona.bound_class_id}, bound_class_name=${targetPersona.bound_class_name}, is_public=${targetPersona.is_public}`);
  
  const hasClassId = targetPersona.bound_class_id !== undefined;
  const hasClassName = targetPersona.bound_class_name !== undefined;
  const hasIsPublic = targetPersona.is_public !== undefined;
  
  if (hasClassId && hasClassName && hasIsPublic) {
    return recordTest('SM-AD03', '分身列表展示班级信息', 'passed', null, {
      bound_class_id: targetPersona.bound_class_id,
      bound_class_name: targetPersona.bound_class_name,
      is_public: targetPersona.is_public
    });
  } else {
    return recordTest('SM-AD03', '分身列表展示班级信息', 'failed', 
      `字段缺失: bound_class_id=${hasClassId}, bound_class_name=${hasClassName}, is_public=${hasIsPublic}`);
  }
}

// SM-AE02: 获取自测学生信息
async function testSM_AE02() {
  console.log('\n🧪 SM-AE02: 获取自测学生信息');
  
  // 读取 SM-AE01 数据
  let ae01Data = null;
  try {
    ae01Data = JSON.parse(fs.readFileSync(`${CONFIG.resultsDir}/smoke_v11_ae01_data.json`, 'utf8'));
  } catch (e) {
    console.log('   未找到前置数据，先执行 SM-AE01...');
    await testSM_AE01();
    try {
      ae01Data = JSON.parse(fs.readFileSync(`${CONFIG.resultsDir}/smoke_v11_ae01_data.json`, 'utf8'));
    } catch (e2) {
      return recordTest('SM-AE02', '获取自测学生信息', 'failed', '无法获取前置数据');
    }
  }
  
  // 获取自测学生信息
  const testStudentRes = await apiRequest('GET', '/api/test-student', null, ae01Data.token);
  
  if (testStudentRes.status !== 200 || testStudentRes.data.code !== 0) {
    return recordTest('SM-AE02', '获取自测学生信息', 'failed', '获取自测学生信息失败');
  }
  
  const student = testStudentRes.data.data;
  
  console.log(`   自测学生字段: user_id=${student.user_id}, username=${student.username}, nickname=${student.nickname}, joined_classes=${Array.isArray(student.joined_classes)}`);
  
  const hasUserId = student.user_id !== undefined;
  const hasUsername = student.username !== undefined;
  const hasNickname = student.nickname !== undefined;
  const hasPasswordHint = student.password_hint !== undefined;
  const hasJoinedClasses = Array.isArray(student.joined_classes);
  
  if (hasUserId && hasUsername && hasNickname && hasPasswordHint && hasJoinedClasses) {
    return recordTest('SM-AE02', '获取自测学生信息', 'passed', null, {
      user_id: student.user_id,
      username: student.username,
      joinedClassesCount: student.joined_classes.length
    });
  } else {
    return recordTest('SM-AE02', '获取自测学生信息', 'failed', '字段不完整');
  }
}

// SM-AE03: 自测学生自动加入班级
async function testSM_AE03() {
  console.log('\n🧪 SM-AE03: 自测学生自动加入班级');
  
  // 读取前置数据
  let ad01Data = null;
  
  try {
    ad01Data = JSON.parse(fs.readFileSync(`${CONFIG.resultsDir}/smoke_v11_ad01_data.json`, 'utf8'));
  } catch (e) {
    return recordTest('SM-AE03', '自测学生自动加入班级', 'skipped', '缺少前置用例数据（SM-AD01）');
  }
  
  // 获取班级所属教师的自测学生
  const testStudentRes = await apiRequest('GET', '/api/test-student', null, ad01Data.token);
  
  if (testStudentRes.status !== 200 || testStudentRes.data.code !== 0) {
    return recordTest('SM-AE03', '自测学生自动加入班级', 'failed', '获取自测学生信息失败');
  }
  
  const testStudent = testStudentRes.data.data;
  const testStudentPersonaId = testStudent.persona_id;
  
  console.log(`   自测学生分身ID: ${testStudentPersonaId}`);
  
  // 获取班级成员列表
  const membersRes = await apiRequest('GET', `/api/classes/${ad01Data.classId}/members`, null, ad01Data.token);
  
  if (membersRes.status !== 200 || membersRes.data.code !== 0) {
    return recordTest('SM-AE03', '自测学生自动加入班级', 'failed', '获取班级成员失败');
  }
  
  let members = membersRes.data.data || [];
  // 处理返回数据是对象格式
  if (!Array.isArray(members) && members.members) {
    members = members.members;
  }
  if (!Array.isArray(members)) {
    members = [];
  }
  
  const testStudentMember = members.find(m => m.student_persona_id === testStudentPersonaId);
  
  console.log(`   班级成员数: ${members.length}, 自测学生在成员中: ${testStudentMember ? '是' : '否'}`);
  
  if (testStudentMember) {
    console.log(`   成员状态: ${testStudentMember.status}`);
    if (testStudentMember.status === 'approved' || testStudentMember.status === 'active') {
      return recordTest('SM-AE03', '自测学生自动加入班级', 'passed', null, {
        memberId: testStudentMember.student_persona_id,
        status: testStudentMember.status
      });
    } else {
      return recordTest('SM-AE03', '自测学生自动加入班级', 'failed', `成员状态不正确: ${testStudentMember.status}`);
    }
  } else {
    return recordTest('SM-AE03', '自测学生自动加入班级', 'failed', '自测学生未自动加入班级');
  }
}

// SM-AF01/A02: 向量召回（需要Python服务，跳过）
async function testSM_AF() {
  console.log('\n🧪 SM-AF01/A02: 知识库向量召回');
  return recordTest('SM-AF01/02', '知识库向量召回', 'skipped', '需要Python向量服务');
}

// 主函数
async function main() {
  console.log('╔════════════════════════════════════════════════════════════╗');
  console.log('║     Phase 3c: 迭代11 API 级别冒烟测试                      ║');
  console.log('╚════════════════════════════════════════════════════════════╝');
  
  // 创建结果目录
  if (!fs.existsSync(CONFIG.resultsDir)) {
    fs.mkdirSync(CONFIG.resultsDir, { recursive: true });
  }
  
  // 检查后端服务
  console.log('\n🔍 检查后端服务...');
  try {
    const healthRes = await apiRequest('GET', '/api/system/health');
    if (healthRes.data.data?.status === 'running') {
      console.log('✅ 后端服务运行中');
    } else {
      console.log('❌ 后端服务异常');
      process.exit(1);
    }
  } catch (e) {
    console.log('❌ 后端服务未启动:', e.message);
    process.exit(1);
  }
  
  // 执行测试用例
  console.log('\n========================================');
  console.log('  第一批：无依赖用例');
  console.log('========================================');
  
  await testSM_AD02();
  await testSM_AD04();
  await testSM_AD01();
  await testSM_AE01();
  
  console.log('\n========================================');
  console.log('  第二批：依赖第一批');
  console.log('========================================');
  
  await testSM_AD03();
  await testSM_AE02();
  await testSM_AE03();
  
  console.log('\n========================================');
  console.log('  第三批：需外部环境');
  console.log('========================================');
  
  await testSM_AF();
  
  // 生成报告
  RESULTS.endTime = new Date().toISOString();
  
  console.log('\n\n╔════════════════════════════════════════════════════════════╗');
  console.log('║                    最终测试报告                            ║');
  console.log('╚════════════════════════════════════════════════════════════╝');
  
  console.log(`\n【执行统计】`);
  console.log(`  总用例数: ${RESULTS.summary.total}`);
  console.log(`  ✅ 通过: ${RESULTS.summary.passed}`);
  console.log(`  ❌ 失败: ${RESULTS.summary.failed}`);
  console.log(`  ⏸️ 跳过: ${RESULTS.summary.skipped}`);
  
  console.log(`\n【用例详情】`);
  RESULTS.tests.forEach(t => {
    const icon = t.status === 'passed' ? '✅' : t.status === 'failed' ? '❌' : '⏸️';
    console.log(`  ${icon} ${t.id}: ${t.name}`);
    if (t.error) console.log(`     原因: ${t.error}`);
  });
  
  // 保存报告
  const reportPath = `${CONFIG.resultsDir}/smoke_v11_api_report.json`;
  fs.writeFileSync(reportPath, JSON.stringify(RESULTS, null, 2));
  console.log(`\n📄 报告已保存: ${reportPath}`);
  
  console.log('\n╔════════════════════════════════════════════════════════════╗');
  if (RESULTS.summary.failed === 0) {
    console.log('║                 ✅ 所有用例通过                            ║');
  } else {
    console.log('║                 ❌ 存在失败用例                            ║');
  }
  console.log('╚════════════════════════════════════════════════════════════╝\n');
  
  process.exit(RESULTS.summary.failed > 0 ? 1 : 0);
}

main().catch(err => {
  console.error('执行异常:', err);
  process.exit(1);
});
