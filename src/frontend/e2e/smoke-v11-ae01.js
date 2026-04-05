/**
 * SM-AE01: 教师注册自动创建自测学生
 * 验证：新教师注册时自动创建自测学生账号
 * 输出数据：{ teacherId, testStudentId, testStudentUsername } 供后续用例使用
 */

const automator = require('miniprogram-automator');
const path = require('path');
const fs = require('fs');

const CONFIG = {
  projectPath: '/Users/aganbai/Desktop/WorkSpace/digital-twin/src/frontend',
  cliPath: '/Applications/wechatwebdevtools.app/Contents/MacOS/cli',
  backendUrl: 'http://localhost:8080',
  screenshotsDir: '/Users/aganbai/Desktop/WorkSpace/digital-twin/test_results/smoke_v11_screenshots'
};

const RESULT = {
  id: 'SM-AE01',
  name: '教师注册自动创建自测学生',
  status: 'pending',
  error: null,
  screenshots: [],
  data: {}, // 供后续用例使用
  startTime: new Date().toISOString(),
  endTime: null
};

async function screenshot(miniProgram, name) {
  if (!fs.existsSync(CONFIG.screenshotsDir)) {
    fs.mkdirSync(CONFIG.screenshotsDir, { recursive: true });
  }
  const screenshotPath = path.join(CONFIG.screenshotsDir, `${name}_${Date.now()}.png`);
  await miniProgram.screenshot({ path: screenshotPath });
  console.log(`📸 截图: ${screenshotPath}`);
  RESULT.screenshots.push(screenshotPath);
  return screenshotPath;
}

async function runTest() {
  console.log('\n========================================');
  console.log(`  用例: ${RESULT.id} - ${RESULT.name}`);
  console.log('========================================\n');

  let miniProgram = null;

  try {
    // 1. 启动微信开发者工具
    console.log('🚀 启动微信开发者工具...');
    miniProgram = await automator.launch({
      projectPath: CONFIG.projectPath,
      cliPath: CONFIG.cliPath,
      port: 9420
    });
    console.log('✅ 微信开发者工具已启动');

    // 2. 打开小程序首页
    console.log('📱 打开小程序首页...');
    await miniProgram.reLaunch('/pages/login/index');
    await new Promise(r => setTimeout(r, 2000));
    await screenshot(miniProgram, 'SM-AE01_01_login_page');

    // 3. 执行 Mock 登录（新教师注册）
    console.log('🔐 执行 Mock 登录（新教师注册）...');
    const loginResult = await miniProgram.callWxMethod('login');
    
    // 使用时间戳确保是新用户
    const timestamp = Date.now();
    
    const loginResp = await fetch(`${CONFIG.backendUrl}/api/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        code: `mock_code_${timestamp}_${Math.random().toString(36).substr(2, 9)}`,
        role: 'teacher',
        nickname: `测试教师AE01_${timestamp}`,
        avatar_url: 'https://example.com/avatar.png'
      })
    });
    const loginData = await loginResp.json();
    
    if (loginResp.status !== 200 || loginData.code !== 0) {
      throw new Error(`登录失败: ${loginData.message}`);
    }
    
    const teacherToken = loginData.data.token;
    const teacherId = loginData.data.user_id;
    
    console.log('✅ 教师注册成功');
    console.log('  教师ID:', teacherId);
    console.log('  教师昵称:', loginData.data.nickname);
    
    // 保存教师信息
    RESULT.data.teacherId = teacherId;
    RESULT.data.teacherToken = teacherToken;

    // 4. 验证自测学生是否自动创建
    console.log('\n🔍 验证自测学生自动创建...');
    
    const testStudentResp = await fetch(`${CONFIG.backendUrl}/api/test-student`, {
      headers: { 'Authorization': `Bearer ${teacherToken}` }
    });
    
    const testStudentData = await testStudentResp.json();
    console.log('  自测学生接口响应:', JSON.stringify(testStudentData, null, 2));

    // 5. 验证自测学生信息
    if (testStudentResp.status === 200 && testStudentData.code === 0) {
      const testStudent = testStudentData.data;
      
      console.log('\n✅ 自测学生信息获取成功');
      console.log('  自测学生ID:', testStudent.id);
      console.log('  用户名:', testStudent.username);
      console.log('  昵称:', testStudent.nickname);
      
      // 验证用户名格式: teacher_{user_id}_test
      const expectedUsername = `teacher_${teacherId}_test`;
      if (testStudent.username === expectedUsername) {
        console.log('  ✅ 用户名格式正确');
      } else {
        console.log(`  ❌ 用户名格式错误: 期望 ${expectedUsername}, 实际 ${testStudent.username}`);
        throw new Error('自测学生用户名格式验证失败');
      }
      
      // 验证不创建教师分身
      const personaResp = await fetch(`${CONFIG.backendUrl}/api/personas`, {
        headers: { 'Authorization': `Bearer ${teacherToken}` }
      });
      const personaData = await personaResp.json();
      
      if (personaData.code === 0) {
        const personas = personaData.data || [];
        if (personas.length === 0) {
          console.log('  ✅ 未创建教师分身（符合预期）');
        } else {
          console.log(`  ⚠️ 创建了 ${personas.length} 个分身（可能不符合预期）`);
          console.log('  分身列表:', JSON.stringify(personas, null, 2));
        }
      }
      
      // 保存自测学生信息供后续用例使用
      RESULT.data.testStudentId = testStudent.id;
      RESULT.data.testStudentUsername = testStudent.username;
      RESULT.data.testStudentNickname = testStudent.nickname;
      RESULT.data.testStudentPassword = testStudent.password_hint || '密码提示未返回';
      
      console.log('\n✅ SM-AE01 通过: 教师注册自动创建自测学生成功');
      RESULT.status = 'passed';
      
    } else if (testStudentResp.status === 404 || testStudentData.code === 40400) {
      console.log('\n❌ SM-AE01 失败: 自测学生未自动创建');
      RESULT.status = 'failed';
      RESULT.error = '自测学生未自动创建';
    } else {
      console.log('\n❌ SM-AE01 失败: 获取自测学生信息失败');
      RESULT.status = 'failed';
      RESULT.error = `获取自测学生失败: ${testStudentData.message}`;
    }

    await screenshot(miniProgram, `SM-AE01_02_result_${RESULT.status}`);

  } catch (error) {
    console.error('\n❌ 测试执行异常:', error.message);
    RESULT.status = 'failed';
    RESULT.error = error.message;
  } finally {
    RESULT.endTime = new Date().toISOString();
    
    if (miniProgram) {
      console.log('\n🔄 关闭微信开发者工具...');
      await miniProgram.close();
    }

    // 输出结果
    console.log('\n========================================');
    console.log('  测试执行完成');
    console.log('========================================');
    console.log('用例编号:', RESULT.id);
    console.log('用例名称:', RESULT.name);
    console.log('执行结果:', RESULT.status === 'passed' ? '✅ 通过' : RESULT.status === 'failed' ? '❌ 失败' : '⏸️ 跳过');
    if (RESULT.error) console.log('失败原因:', RESULT.error);
    console.log('关键数据:', JSON.stringify(RESULT.data, null, 2));
    console.log('截图路径:', RESULT.screenshots.join(', ') || '无');
    console.log('========================================\n');

    // 保存结果
    const resultPath = `/Users/aganbai/Desktop/WorkSpace/digital-twin/test_results/${RESULT.id}_result.json`;
    fs.writeFileSync(resultPath, JSON.stringify(RESULT, null, 2));
    console.log(`📄 结果已保存: ${resultPath}`);
    
    // 同时保存数据供后续用例使用
    const dataPath = `/Users/aganbai/Desktop/WorkSpace/digital-twin/test_results/smoke_v11_ae01_data.json`;
    fs.writeFileSync(dataPath, JSON.stringify(RESULT.data, null, 2));
    console.log(`📄 用例数据已保存: ${dataPath}`);

    process.exit(RESULT.status === 'passed' ? 0 : 1);
  }
}

runTest();
