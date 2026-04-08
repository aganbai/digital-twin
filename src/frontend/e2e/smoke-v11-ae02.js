/**
 * SM-AE02: 获取自测学生信息
 * 前置依赖: SM-AE01（需要已创建的自测学生）
 * 验证：获取自测学生信息接口返回正确的用户信息和班级列表
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
  id: 'SM-AE02',
  name: '获取自测学生信息',
  status: 'pending',
  error: null,
  screenshots: [],
  data: {},
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
    // 1. 检查前置依赖数据
    console.log('🔍 检查前置依赖数据...');
    const ae01DataPath = '/Users/aganbai/Desktop/WorkSpace/digital-twin/test_results/smoke_v11_ae01_data.json';
    
    let ae01Data = null;
    if (fs.existsSync(ae01DataPath)) {
      ae01Data = JSON.parse(fs.readFileSync(ae01DataPath, 'utf8'));
      console.log('✅ 读取到 SM-AE01 数据');
      console.log('  教师ID:', ae01Data.teacherId);
      console.log('  自测学生ID:', ae01Data.testStudentId);
    } else {
      console.log('⚠️ 未找到 SM-AE01 数据，将创建新的测试数据');
    }

    // 2. 启动微信开发者工具
    console.log('\n🚀 启动微信开发者工具...');
    miniProgram = await automator.launch({
      projectPath: CONFIG.projectPath,
      cliPath: CONFIG.cliPath,
      port: 9420
    });
    console.log('✅ 微信开发者工具已启动');

    // 3. 打开小程序首页
    console.log('📱 打开小程序首页...');
    await miniProgram.reLaunch('/pages/login/index');
    await new Promise(r => setTimeout(r, 2000));
    await screenshot(miniProgram, 'SM-AE02_01_login_page');

    // 4. 登录
    let teacherToken;
    let teacherId;
    
    if (ae01Data && ae01Data.teacherToken) {
      teacherToken = ae01Data.teacherToken;
      teacherId = ae01Data.teacherId;
      console.log('✅ 使用 SM-AE01 的登录凭证');
    } else {
      console.log('🔐 执行新的 Mock 登录...');
      const timestamp = Date.now();
      const loginResp = await fetch(`${CONFIG.backendUrl}/api/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          code: `mock_code_ae02_${timestamp}`,
          role: 'teacher',
          nickname: `测试教师AE02_${timestamp}`,
          avatar_url: 'https://example.com/avatar.png'
        })
      });
      const loginData = await loginResp.json();
      
      if (loginResp.status !== 200 || loginData.code !== 0) {
        throw new Error(`登录失败: ${loginData.message}`);
      }
      
      teacherToken = loginData.data.token;
      teacherId = loginData.data.user_id;
      console.log('✅ 登录成功');
    }

    // 5. 导航到个人中心
    console.log('\n🧪 导航到个人中心...');
    await miniProgram.navigateTo('/pages/profile/index');
    await new Promise(r => setTimeout(r, 2000));
    await screenshot(miniProgram, 'SM-AE02_02_profile_page');

    // 6. 获取自测学生信息
    console.log('\n🔍 获取自测学生信息...');
    const testStudentResp = await fetch(`${CONFIG.backendUrl}/api/test-student`, {
      headers: { 'Authorization': `Bearer ${teacherToken}` }
    });
    
    const testStudentData = await testStudentResp.json();
    console.log('  接口响应:', JSON.stringify(testStudentData, null, 2));

    // 7. 验证自测学生信息
    if (testStudentResp.status === 200 && testStudentData.code === 0) {
      const student = testStudentData.data;
      
      console.log('\n✅ 验证自测学生信息字段...');
      
      // 检查关键字段
      const checks = [
        { field: 'id', value: student.id, required: true },
        { field: 'username', value: student.username, required: true },
        { field: 'nickname', value: student.nickname, required: true },
        { field: 'password_hint', value: student.password_hint, required: true },
        { field: 'classes', value: student.classes, required: true, type: 'array' },
        { field: 'created_at', value: student.created_at, required: true }
      ];
      
      let allValid = true;
      for (const check of checks) {
        const hasField = check.value !== undefined && check.value !== null;
        const valid = hasField && (check.type !== 'array' || Array.isArray(check.value));
        console.log(`  ${check.field}: ${valid ? '✅' : '❌'} ${hasField ? check.value : '缺失'}`);
        if (!valid) allValid = false;
      }
      
      // 验证用户名格式
      const expectedUsername = `teacher_${teacherId}_test`;
      if (student.username === expectedUsername) {
        console.log(`  username格式: ✅ ${student.username}`);
      } else {
        console.log(`  username格式: ⚠️ 期望 ${expectedUsername}, 实际 ${student.username}`);
      }
      
      // 验证班级列表
      if (Array.isArray(student.classes)) {
        console.log(`  已加入班级数: ${student.classes.length}`);
        for (const cls of student.classes) {
          console.log(`    - ${cls.name || cls.class_name} (ID: ${cls.id || cls.class_id})`);
        }
      }
      
      await screenshot(miniProgram, 'SM-AE02_03_test_student_info');
      
      if (allValid) {
        console.log('\n✅ SM-AE02 通过: 自测学生信息获取成功');
        RESULT.status = 'passed';
        RESULT.data.testStudentId = student.id;
        RESULT.data.testStudentUsername = student.username;
        RESULT.data.classes = student.classes;
      } else {
        console.log('\n❌ SM-AE02 失败: 部分字段缺失');
        RESULT.status = 'failed';
        RESULT.error = '自测学生信息字段不完整';
      }
      
    } else if (testStudentResp.status === 404 || testStudentData.code === 40400) {
      console.log('\n❌ SM-AE02 失败: 未找到自测学生');
      RESULT.status = 'failed';
      RESULT.error = '自测学生不存在';
    } else {
      console.log('\n❌ SM-AE02 失败: 获取自测学生信息失败');
      RESULT.status = 'failed';
      RESULT.error = `获取失败: ${testStudentData.message}`;
    }

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
    console.log('截图路径:', RESULT.screenshots.join(', ') || '无');
    console.log('========================================\n');

    // 保存结果
    const resultPath = `/Users/aganbai/Desktop/WorkSpace/digital-twin/test_results/${RESULT.id}_result.json`;
    fs.writeFileSync(resultPath, JSON.stringify(RESULT, null, 2));
    console.log(`📄 结果已保存: ${resultPath}`);

    process.exit(RESULT.status === 'passed' ? 0 : 1);
  }
}

runTest();
