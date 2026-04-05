/**
 * SM-AD03: 分身列表展示班级信息
 * 前置依赖: SM-AD01（需要已创建的班级和分身）
 * 验证：分身列表中显示 bound_class_id, bound_class_name, is_public
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
  id: 'SM-AD03',
  name: '分身列表展示班级信息',
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
    const ad01DataPath = '/Users/aganbai/Desktop/WorkSpace/digital-twin/test_results/smoke_v11_ad01_data.json';
    
    let ad01Data = null;
    if (fs.existsSync(ad01DataPath)) {
      ad01Data = JSON.parse(fs.readFileSync(ad01DataPath, 'utf8'));
      console.log('✅ 读取到 SM-AD01 数据');
      console.log('  教师ID:', ad01Data.teacherId);
      console.log('  班级ID:', ad01Data.classId);
      console.log('  分身ID:', ad01Data.personaId);
    } else {
      console.log('⚠️ 未找到 SM-AD01 数据，将创建新的测试数据');
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
    await screenshot(miniProgram, 'SM-AD03_01_login_page');

    // 4. 登录（使用已有token或新登录）
    let teacherToken;
    let teacherId;
    
    if (ad01Data && ad01Data.teacherToken) {
      teacherToken = ad01Data.teacherToken;
      teacherId = ad01Data.teacherId;
      console.log('✅ 使用 SM-AD01 的登录凭证');
    } else {
      console.log('🔐 执行新的 Mock 登录...');
      const loginResult = await miniProgram.callWxMethod('login');
      
      const loginResp = await fetch(`${CONFIG.backendUrl}/api/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          code: loginResult.code,
          role: 'teacher',
          nickname: '测试教师AD03',
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

    // 5. 如果没有前置数据，先创建班级
    let classId, personaId, className;
    if (ad01Data && ad01Data.classId) {
      classId = ad01Data.classId;
      personaId = ad01Data.personaId;
      className = ad01Data.className;
      console.log('\n📋 使用已有测试数据');
    } else {
      console.log('\n🏫 创建测试班级...');
      const createResp = await fetch(`${CONFIG.backendUrl}/api/classes`, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${teacherToken}`
        },
        body: JSON.stringify({
          name: `测试班级_AD03_${Date.now()}`,
          persona_nickname: '测试老师',
          school: '测试大学',
          description: '测试描述',
          is_public: true
        })
      });
      
      const createData = await createResp.json();
      if (createResp.status !== 200 || createData.code !== 0) {
        throw new Error(`创建班级失败: ${createData.message}`);
      }
      
      classId = createData.data.class_id;
      personaId = createData.data.persona_id;
      className = `测试班级_AD03_${Date.now()}`;
      console.log('✅ 班级创建成功');
    }

    // 6. 导航到个人中心/分身列表页
    console.log('\n🧪 导航到分身列表页...');
    await miniProgram.navigateTo('/pages/profile/index');
    await new Promise(r => setTimeout(r, 2000));
    await screenshot(miniProgram, 'SM-AD03_02_profile_page');

    // 7. 获取分身列表数据
    console.log('\n🔍 获取分身列表数据...');
    const personaResp = await fetch(`${CONFIG.backendUrl}/api/personas`, {
      headers: { 'Authorization': `Bearer ${teacherToken}` }
    });
    
    const personaData = await personaResp.json();
    
    if (personaResp.status !== 200 || personaData.code !== 0) {
      throw new Error(`获取分身列表失败: ${personaData.message}`);
    }

    const personas = personaData.data || [];
    console.log(`  找到 ${personas.length} 个分身`);

    // 8. 验证分身列表展示班级信息
    console.log('\n✅ 验证分身列表字段...');
    
    let foundTargetPersona = false;
    let allFieldsValid = true;
    
    for (const persona of personas) {
      console.log(`\n  分身: ${persona.nickname} (ID: ${persona.id})`);
      
      // 检查关键字段
      const hasBoundClassId = persona.bound_class_id !== undefined;
      const hasBoundClassName = persona.bound_class_name !== undefined;
      const hasIsPublic = persona.is_public !== undefined;
      
      console.log(`    bound_class_id: ${hasBoundClassId ? '✅' : '❌'} ${persona.bound_class_id}`);
      console.log(`    bound_class_name: ${hasBoundClassName ? '✅' : '❌'} ${persona.bound_class_name}`);
      console.log(`    is_public: ${hasIsPublic ? '✅' : '❌'} ${persona.is_public}`);
      
      if (persona.id === personaId) {
        foundTargetPersona = true;
        
        if (persona.bound_class_id !== classId) {
          console.log(`    ❌ bound_class_id 不匹配: 期望 ${classId}, 实际 ${persona.bound_class_id}`);
          allFieldsValid = false;
        }
        
        if (persona.bound_class_name !== className) {
          console.log(`    ⚠️ bound_class_name 不匹配: 期望 ${className}, 实际 ${persona.bound_class_name}`);
        }
        
        if (persona.is_public !== true) {
          console.log(`    ⚠️ is_public 不是 true: ${persona.is_public}`);
        }
      }
      
      if (!hasBoundClassId || !hasBoundClassName || !hasIsPublic) {
        allFieldsValid = false;
      }
    }

    await screenshot(miniProgram, 'SM-AD03_03_persona_list');

    // 9. 验证结果
    if (foundTargetPersona && allFieldsValid) {
      console.log('\n✅ SM-AD03 通过: 分身列表正确展示班级信息');
      RESULT.status = 'passed';
    } else {
      console.log('\n❌ SM-AD03 失败');
      if (!foundTargetPersona) {
        console.log('  未找到目标分身');
      }
      if (!allFieldsValid) {
        console.log('  部分字段验证失败');
      }
      RESULT.status = 'failed';
      RESULT.error = '分身列表字段验证失败';
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
