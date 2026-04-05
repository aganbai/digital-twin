/**
 * SM-AD01: 教师创建班级同步创建分身
 * 验证：教师创建班级时自动创建绑定分身
 * 输出数据：{ classId, personaId } 供后续用例使用
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
  id: 'SM-AD01',
  name: '教师创建班级同步创建分身',
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
  let teacherToken = null;
  let teacherId = null;

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
    await screenshot(miniProgram, 'SM-AD01_01_login_page');

    // 3. 执行 Mock 登录（教师角色）
    console.log('🔐 执行 Mock 登录（教师角色）...');
    const loginResult = await miniProgram.callWxMethod('login');
    
    const loginResp = await fetch(`${CONFIG.backendUrl}/api/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        code: loginResult.code,
        role: 'teacher',
        nickname: '测试教师AD01',
        avatar_url: 'https://example.com/avatar.png'
      })
    });
    const loginData = await loginResp.json();
    
    if (loginResp.status !== 200 || loginData.code !== 0) {
      throw new Error(`登录失败: ${loginData.message}`);
    }
    
    teacherToken = loginData.data.token;
    teacherId = loginData.data.user_id;
    console.log('✅ 登录成功，教师ID:', teacherId);
    
    // 保存教师信息供后续用例使用
    RESULT.data.teacherId = teacherId;
    RESULT.data.teacherToken = teacherToken;

    // 4. 通过页面导航进入班级创建页
    console.log('\n🧪 通过页面导航进入班级创建页...');
    
    // 等待页面加载
    const page = await miniProgram.currentPage();
    console.log('  当前页面:', page.path);
    
    // 点击"创建班级"按钮（假设在首页或个人中心有入口）
    // 注意：这里需要根据实际页面结构调整选择器
    console.log('  导航到班级创建页...');
    await miniProgram.navigateTo('/pages/class-create/index');
    await new Promise(r => setTimeout(r, 2000));
    await screenshot(miniProgram, 'SM-AD01_02_class_create_page');

    // 5. 填写班级创建表单
    console.log('\n📝 填写班级创建表单...');
    
    const className = `测试班级_${Date.now()}`;
    const personaNickname = '张老师';
    const school = '测试大学';
    const description = '测试教师分身';
    
    // 通过 API 创建班级（验证核心功能）
    console.log('  通过 API 创建班级...');
    const createResp = await fetch(`${CONFIG.backendUrl}/api/classes`, {
      method: 'POST',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${teacherToken}`
      },
      body: JSON.stringify({
        name: className,
        persona_nickname: personaNickname,
        school: school,
        description: description,
        is_public: true
      })
    });
    
    const createData = await createResp.json();
    console.log('  创建响应:', JSON.stringify(createData, null, 2));

    // 6. 验证创建结果
    if (createResp.status !== 200 || createData.code !== 0) {
      throw new Error(`创建班级失败: ${createData.message}`);
    }

    const classId = createData.data.class_id;
    const personaId = createData.data.persona_id;
    
    console.log('\n✅ 班级创建成功');
    console.log('  班级ID:', classId);
    console.log('  分身ID:', personaId);
    
    // 保存数据供后续用例使用
    RESULT.data.classId = classId;
    RESULT.data.personaId = personaId;
    RESULT.data.className = className;
    RESULT.data.personaNickname = personaNickname;

    // 7. 验证分身绑定关系
    console.log('\n🔍 验证分身绑定关系...');
    
    const personaResp = await fetch(`${CONFIG.backendUrl}/api/personas`, {
      headers: { 'Authorization': `Bearer ${teacherToken}` }
    });
    const personaData = await personaResp.json();
    
    if (personaData.code === 0 && personaData.data) {
      const persona = personaData.data.find(p => p.id === personaId);
      if (persona) {
        console.log('  分身信息:', JSON.stringify(persona, null, 2));
        
        // 验证绑定关系
        if (persona.bound_class_id === classId) {
          console.log('  ✅ 分身正确绑定到班级');
        } else {
          console.log(`  ❌ 分身绑定错误: 期望 ${classId}, 实际 ${persona.bound_class_id}`);
          throw new Error('分身绑定关系验证失败');
        }
        
        if (persona.is_public === true) {
          console.log('  ✅ is_public 默认为 true');
        } else {
          console.log(`  ⚠️ is_public 不是 true: ${persona.is_public}`);
        }
      } else {
        throw new Error('未找到创建的分身');
      }
    } else {
      throw new Error('获取分身列表失败');
    }

    // 8. 截图记录
    await screenshot(miniProgram, 'SM-AD01_03_create_success');

    console.log('\n✅ SM-AD01 通过: 教师创建班级同步创建分身成功');
    RESULT.status = 'passed';

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
    const dataPath = `/Users/aganbai/Desktop/WorkSpace/digital-twin/test_results/smoke_v11_ad01_data.json`;
    fs.writeFileSync(dataPath, JSON.stringify(RESULT.data, null, 2));
    console.log(`📄 用例数据已保存: ${dataPath}`);

    process.exit(RESULT.status === 'passed' ? 0 : 1);
  }
}

runTest();
