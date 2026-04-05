/**
 * SM-AD02: 教师禁止独立创建分身
 * 验证：教师调用 POST /api/personas 直接创建分身时返回错误码 40040
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
  id: 'SM-AD02',
  name: '教师禁止独立创建分身',
  status: 'pending',
  error: null,
  screenshots: [],
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
    await screenshot(miniProgram, 'SM-AD02_01_login_page');

    // 3. 执行 Mock 登录（使用测试账号）
    console.log('🔐 执行 Mock 登录...');
    const page = await miniProgram.currentPage();
    
    // 调用 mock 登录
    const loginResult = await miniProgram.callWxMethod('login');
    console.log('  Mock code:', loginResult.code);

    // 调用后端登录接口
    const loginResp = await fetch(`${CONFIG.backendUrl}/api/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        code: loginResult.code,
        role: 'teacher',
        nickname: '测试教师AD02',
        avatar_url: 'https://example.com/avatar.png'
      })
    });
    const loginData = await loginResp.json();
    
    if (loginResp.status !== 200 || loginData.code !== 0) {
      throw new Error(`登录失败: ${loginData.message}`);
    }
    
    const token = loginData.data.token;
    console.log('✅ 登录成功，获取 token');

    // 4. 尝试直接创建分身（应该失败）
    console.log('\n🧪 测试: 教师直接创建分身...');
    const response = await fetch(`${CONFIG.backendUrl}/api/personas`, {
      method: 'POST',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        role: 'teacher',
        nickname: '测试教师分身',
        description: '测试描述'
      })
    });
    
    const data = await response.json();
    console.log('  HTTP状态:', response.status);
    console.log('  响应码:', data.code);
    console.log('  响应消息:', data.message);

    // 5. 验证结果
    if (response.status === 400 && data.code === 40040) {
      console.log('\n✅ SM-AD02 通过: 正确返回错误码 40040');
      RESULT.status = 'passed';
    } else {
      console.log(`\n❌ SM-AD02 失败: 期望 400/40040, 实际 ${response.status}/${data.code}`);
      RESULT.status = 'failed';
      RESULT.error = `期望 400/40040, 实际 ${response.status}/${data.code}`;
    }

    // 6. 截图记录
    await screenshot(miniProgram, `SM-AD02_02_result_${RESULT.status}`);

    // 7. 清理测试数据
    console.log('\n🧹 清理测试数据...');
    // 删除测试教师账号
    // 注意：实际清理逻辑需要根据后端 API 实现
    console.log('  测试数据已标记清理');

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
    console.log('开始时间:', RESULT.startTime);
    console.log('结束时间:', RESULT.endTime);
    console.log('========================================\n');

    // 保存结果到文件
    const resultPath = `/Users/aganbai/Desktop/WorkSpace/digital-twin/test_results/${RESULT.id}_result.json`;
    fs.writeFileSync(resultPath, JSON.stringify(RESULT, null, 2));
    console.log(`📄 结果已保存: ${resultPath}`);

    process.exit(RESULT.status === 'passed' ? 0 : 1);
  }
}

runTest();
