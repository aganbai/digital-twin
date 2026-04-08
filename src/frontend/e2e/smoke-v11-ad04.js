/**
 * SM-AD04: 已删除的接口返回404
 * 验证：已删除的分身切换/激活/停用接口返回404
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
  id: 'SM-AD04',
  name: '已删除的接口返回404',
  status: 'pending',
  error: null,
  screenshots: [],
  details: [],
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

async function testDeletedEndpoint(token, method, endpoint) {
  console.log(`\n  测试: ${method} ${endpoint}`);
  try {
    const response = await fetch(`${CONFIG.backendUrl}${endpoint}`, {
      method: method,
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      }
    });
    
    const data = await response.json().catch(() => ({}));
    console.log(`    HTTP状态: ${response.status}`);
    
    const passed = response.status === 404;
    RESULT.details.push({
      endpoint,
      method,
      status: response.status,
      passed
    });
    
    if (passed) {
      console.log(`    ✅ 返回 404，符合预期`);
    } else {
      console.log(`    ❌ 返回 ${response.status}，期望 404`);
    }
    
    return passed;
  } catch (e) {
    console.log(`    ❌ 请求异常: ${e.message}`);
    RESULT.details.push({
      endpoint,
      method,
      error: e.message,
      passed: false
    });
    return false;
  }
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
    await screenshot(miniProgram, 'SM-AD04_01_login_page');

    // 3. 执行 Mock 登录
    console.log('🔐 执行 Mock 登录...');
    const loginResult = await miniProgram.callWxMethod('login');
    
    const loginResp = await fetch(`${CONFIG.backendUrl}/api/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        code: loginResult.code,
        role: 'teacher',
        nickname: '测试教师AD04',
        avatar_url: 'https://example.com/avatar.png'
      })
    });
    const loginData = await loginResp.json();
    
    if (loginResp.status !== 200 || loginData.code !== 0) {
      throw new Error(`登录失败: ${loginData.message}`);
    }
    
    const token = loginData.data.token;
    console.log('✅ 登录成功');

    // 4. 测试已删除的接口
    console.log('\n🧪 测试已删除接口...');
    
    const deletedEndpoints = [
      { method: 'PUT', endpoint: '/api/personas/1/switch' },
      { method: 'PUT', endpoint: '/api/personas/1/activate' },
      { method: 'PUT', endpoint: '/api/personas/1/deactivate' }
    ];

    let allPassed = true;
    for (const { method, endpoint } of deletedEndpoints) {
      const passed = await testDeletedEndpoint(token, method, endpoint);
      if (!passed) allPassed = false;
    }

    // 5. 验证结果
    if (allPassed) {
      console.log('\n✅ SM-AD04 通过: 所有已删除接口均返回404');
      RESULT.status = 'passed';
    } else {
      console.log('\n❌ SM-AD04 失败: 部分接口未返回404');
      RESULT.status = 'failed';
      RESULT.error = '部分已删除接口未返回404';
    }

    await screenshot(miniProgram, `SM-AD04_02_result_${RESULT.status}`);

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
    console.log('测试详情:', JSON.stringify(RESULT.details, null, 2));
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
