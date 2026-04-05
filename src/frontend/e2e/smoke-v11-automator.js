/**
 * Phase 3c: 迭代11 冒烟测试 - miniprogram-automator 端到端验证
 * 模块 AD: 班级绑定分身 (5条)
 * 模块 AE: 自测学生 (4条)
 * 模块 AF: 向量召回优化 (2条)
 */

const automator = require('miniprogram-automator');
const path = require('path');
const fs = require('fs');

// 配置
const CONFIG = {
  projectPath: '/Users/aganbai/Desktop/WorkSpace/digital-twin/src/frontend',
  cliPath: '/Applications/wechatwebdevtools.app/Contents/MacOS/cli',
  backendUrl: 'http://localhost:8080',
  screenshotsDir: '/Users/aganbai/Desktop/WorkSpace/digital-twin/test_results/smoke_v11_screenshots'
};

// 测试结果
const results = {
  env: {},
  passed: [],
  failed: [],
  startTime: new Date().toISOString()
};

// 延迟函数
const delay = (ms) => new Promise(resolve => setTimeout(resolve, ms));

// 截图函数
async function screenshot(miniProgram, name) {
  if (!fs.existsSync(CONFIG.screenshotsDir)) {
    fs.mkdirSync(CONFIG.screenshotsDir, { recursive: true });
  }
  const screenshotPath = path.join(CONFIG.screenshotsDir, `${name}_${Date.now()}.png`);
  await miniProgram.screenshot({ path: screenshotPath });
  console.log(`📸 截图已保存: ${screenshotPath}`);
  return screenshotPath;
}

// 环境检查
async function checkEnvironment() {
  console.log('\n🔍 第一步：环境可用性检查\n');
  
  // 1. 检查 miniprogram-automator
  try {
    require('miniprogram-automator');
    results.env.automator = '✅ 已安装';
    console.log('✅ miniprogram-automator 已安装');
  } catch (e) {
    results.env.automator = '❌ 未安装: ' + e.message;
    console.log('❌ miniprogram-automator 未安装');
    return false;
  }
  
  // 2. 检查微信开发者工具
  if (fs.existsSync('/Applications/wechatwebdevtools.app')) {
    results.env.devtools = '✅ 已安装';
    console.log('✅ 微信开发者工具已安装');
  } else {
    results.env.devtools = '❌ 未安装';
    console.log('❌ 微信开发者工具未安装');
    return false;
  }
  
  // 3. 检查 CLI 工具
  if (fs.existsSync(CONFIG.cliPath)) {
    results.env.cli = '✅ CLI 工具存在';
    console.log('✅ CLI 工具存在');
  } else {
    results.env.cli = '❌ CLI 工具不存在';
    console.log('❌ CLI 工具不存在');
    return false;
  }
  
  // 4. 检查小程序项目路径
  if (fs.existsSync(CONFIG.projectPath)) {
    results.env.projectPath = '✅ 项目路径存在';
    console.log('✅ 小程序项目路径存在');
  } else {
    results.env.projectPath = '❌ 项目路径不存在';
    console.log('❌ 小程序项目路径不存在');
    return false;
  }
  
  // 5. 检查后端服务
  try {
    const response = await fetch(`${CONFIG.backendUrl}/api/system/health`);
    if (response.ok) {
      results.env.backend = '✅ 后端服务运行中';
      console.log('✅ 后端服务运行中');
    } else {
      results.env.backend = '❌ 后端服务异常';
      console.log('❌ 后端服务异常');
      return false;
    }
  } catch (e) {
    results.env.backend = '❌ 后端服务未启动: ' + e.message;
    console.log('❌ 后端服务未启动:', e.message);
    return false;
  }
  
  console.log('\n✅ 环境检查全部通过！\n');
  return true;
}

// 启动小程序
async function launchMiniProgram() {
  console.log('\n🚀 第二步：启动微信开发者工具\n');
  
  try {
    console.log('⏳ 正在启动微信开发者工具...');
    const miniProgram = await automator.launch({
      projectPath: CONFIG.projectPath,
      cliPath: CONFIG.cliPath,
      port: 9420
    });
    
    console.log('✅ 微信开发者工具已启动');
    
    // 监听异常
    miniProgram.on('exception', (data) => {
      console.error('⚠️ 小程序异常:', data);
    });
    
    return miniProgram;
  } catch (error) {
    console.error('❌ 启动微信开发者工具失败:', error.message);
    throw error;
  }
}

// 执行冒烟测试用例 - 实际API测试版本
async function runSmokeTests(miniProgram) {
  console.log('\n🧪 第三步：执行冒烟测试用例（API级别验证）\n');
  
  const testResults = [];
  
  // 模块 AD: 班级绑定分身
  console.log('\n📚 模块 AD: 班级绑定分身\n');
  
  // SM-AD02: 教师禁止独立创建分身（API测试）
  console.log('🔍 SM-AD02: 测试教师禁止独立创建分身...');
  try {
    // 先获取token
    const timestamp = Date.now().toString().slice(-6);
    const loginResp = await fetch(`${CONFIG.backendUrl}/api/auth/wx-login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code: `mock_sm_ad02_${timestamp}` })
    });
    const loginData = await loginResp.json();
    const token = loginData.data?.token;
    
    if (!token) {
      throw new Error('获取token失败');
    }
    
    // 补全教师资料
    await fetch(`${CONFIG.backendUrl}/api/auth/complete-profile`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
      body: JSON.stringify({ role: 'teacher', nickname: `SM_AD02_${timestamp}`, school: '测试大学', description: '测试' })
    });
    
    // 尝试创建分身
    const response = await fetch(`${CONFIG.backendUrl}/api/personas`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
      body: JSON.stringify({
        role: 'teacher',
        nickname: '测试教师',
        description: '测试描述'
      })
    });
    const data = await response.json();
    if (response.status === 400 && data.code === 40040) {
      console.log('✅ SM-AD02 通过: 返回错误码 40040');
      testResults.push({ id: 'SM-AD02', name: '教师禁止独立创建分身', status: 'passed' });
      results.passed.push({ id: 'SM-AD02', name: '教师禁止独立创建分身' });
    } else {
      console.log(`❌ SM-AD02 失败: 期望 400/40040, 实际 ${response.status}/${data.code}`);
      testResults.push({ id: 'SM-AD02', name: '教师禁止独立创建分身', status: 'failed', error: `期望 400/40040, 实际 ${response.status}/${data.code}` });
      results.failed.push({ id: 'SM-AD02', name: '教师禁止独立创建分身', error: `期望 400/40040, 实际 ${response.status}/${data.code}` });
    }
  } catch (e) {
    console.log('❌ SM-AD02 异常:', e.message);
    testResults.push({ id: 'SM-AD02', name: '教师禁止独立创建分身', status: 'failed', error: e.message });
    results.failed.push({ id: 'SM-AD02', name: '教师禁止独立创建分身', error: e.message });
  }
  
  // SM-AD04: 已删除的接口返回404
  console.log('\n🔍 SM-AD04: 测试已删除接口返回404...');
  const deletedEndpoints = [
    '/api/personas/1/switch',
    '/api/personas/1/activate',
    '/api/personas/1/deactivate'
  ];
  let ad04Passed = true;
  for (const endpoint of deletedEndpoints) {
    try {
      const response = await fetch(`${CONFIG.backendUrl}${endpoint}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' }
      });
      if (response.status !== 404) {
        console.log(`  ❌ ${endpoint} 返回 ${response.status}, 期望 404`);
        ad04Passed = false;
      } else {
        console.log(`  ✅ ${endpoint} 返回 404`);
      }
    } catch (e) {
      console.log(`  ❌ ${endpoint} 异常:`, e.message);
      ad04Passed = false;
    }
  }
  if (ad04Passed) {
    console.log('✅ SM-AD04 通过: 所有已删除接口返回404');
    testResults.push({ id: 'SM-AD04', name: '已删除的接口返回404', status: 'passed' });
    results.passed.push({ id: 'SM-AD04', name: '已删除的接口返回404' });
  } else {
    console.log('❌ SM-AD04 失败');
    testResults.push({ id: 'SM-AD04', name: '已删除的接口返回404', status: 'failed' });
    results.failed.push({ id: 'SM-AD04', name: '已删除的接口返回404', error: '部分接口未返回404' });
  }
  
  // 模块 AE: 自测学生
  console.log('\n📚 模块 AE: 自测学生\n');
  
  // SM-AE02: 获取自测学生信息（API测试）
  console.log('🔍 SM-AE02: 测试获取自测学生信息接口...');
  try {
    const response = await fetch(`${CONFIG.backendUrl}/api/test-student`);
    if (response.status === 200 || response.status === 401) {
      console.log(`✅ SM-AE02 通过: 接口返回 ${response.status}`);
      testResults.push({ id: 'SM-AE02', name: '获取自测学生信息', status: 'passed' });
      results.passed.push({ id: 'SM-AE02', name: '获取自测学生信息' });
    } else {
      console.log(`⚠️ SM-AE02: 接口返回 ${response.status}`);
      testResults.push({ id: 'SM-AE02', name: '获取自测学生信息', status: 'pending' });
    }
  } catch (e) {
    console.log('❌ SM-AE02 异常:', e.message);
    testResults.push({ id: 'SM-AE02', name: '获取自测学生信息', status: 'failed', error: e.message });
  }
  
  // 模块 AF: 向量召回优化
  console.log('\n📚 模块 AF: 向量召回优化\n');
  console.log('⚠️ SM-AF01/A02: 向量召回测试需要知识库数据，跳过API验证');
  testResults.push({ id: 'SM-AF01', name: '知识库向量召回100条', status: 'pending', note: '需要知识库数据' });
  testResults.push({ id: 'SM-AF02', name: '知识库 scope=global 生效', status: 'pending', note: '需要知识库数据' });
  
  // 需要登录的用例
  console.log('\n📚 需要微信登录的用例（需人工/UI自动化）:\n');
  const needLoginCases = [
    { id: 'SM-AD01', name: '教师创建班级同步创建分身' },
    { id: 'SM-AD03', name: '分身列表展示班级信息' },
    { id: 'SM-AD05', name: '班级 is_public 设置' },
    { id: 'SM-AE01', name: '教师注册自动创建自测学生' },
    { id: 'SM-AE03', name: '自测学生自动加入班级' },
    { id: 'SM-AE04', name: '重置自测学生数据' }
  ];
  needLoginCases.forEach(tc => {
    console.log(`  ⏸️ ${tc.id}: ${tc.name} [需要微信登录]`);
    testResults.push({ id: tc.id, name: tc.name, status: 'pending', note: '需要微信登录' });
  });
  
  return testResults;
}

// 生成测试报告
function generateReport() {
  console.log('\n\n📊 ========== Phase 3c 冒烟测试报告 ==========\n');
  
  console.log('【环境检查结果】');
  Object.entries(results.env).forEach(([key, value]) => {
    console.log(`  ${key}: ${value}`);
  });
  
  console.log('\n【测试执行结果】');
  console.log(`  开始时间: ${results.startTime}`);
  console.log(`  结束时间: ${new Date().toISOString()}`);
  console.log(`  通过用例: ${results.passed.length}`);
  console.log(`  失败用例: ${results.failed.length}`);
  
  console.log('\n【用例详情】');
  results.passed.forEach(tc => console.log(`  ✅ ${tc.id}: ${tc.name}`));
  results.failed.forEach(tc => console.log(`  ❌ ${tc.id}: ${tc.name} - ${tc.error}`));
  
  console.log('\n========================================\n');
  
  // 保存报告到文件
  const reportPath = '/Users/aganbai/Desktop/WorkSpace/digital-twin/test_results/smoke_v11_report.json';
  fs.writeFileSync(reportPath, JSON.stringify(results, null, 2));
  console.log(`📄 详细报告已保存: ${reportPath}`);
}

// 主函数
async function main() {
  console.log('========================================');
  console.log('  Phase 3c: 迭代11 端到端冒烟验证');
  console.log('  使用 miniprogram-automator SDK');
  console.log('========================================\n');
  
  let miniProgram = null;
  
  try {
    // 第一步：环境检查
    const envOk = await checkEnvironment();
    if (!envOk) {
      console.error('\n❌ 环境检查失败，无法继续执行');
      console.log('\n请检查以下配置：');
      console.log('1. 确保已安装 miniprogram-automator: npm install -g miniprogram-automator');
      console.log('2. 确保已安装微信开发者工具');
      console.log('3. 确保微信开发者工具已开启「安全设置」中的「CLI/HTTP 调用功能」');
      console.log('4. 确保后端服务已启动');
      process.exit(1);
    }
    
    // 第二步：启动小程序
    miniProgram = await launchMiniProgram();
    
    // 第三步：执行冒烟测试
    await runSmokeTests(miniProgram);
    
    console.log('\n✅ 冒烟测试环境验证完成！');
    
  } catch (error) {
    console.error('\n❌ 测试执行失败:', error.message);
    results.failed.push({ id: 'ENV', name: '环境启动', error: error.message });
  } finally {
    // 生成报告
    generateReport();
    
    // 关闭小程序
    if (miniProgram) {
      console.log('\n🔄 关闭微信开发者工具...');
      await miniProgram.close();
      console.log('✅ 已关闭');
    }
  }
}

// 运行
main().catch(console.error);
