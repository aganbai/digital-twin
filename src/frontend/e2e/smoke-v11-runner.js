#!/usr/bin/env node
/**
 * Phase 3c: 迭代11 端到端冒烟验证 - 主执行脚本
 * 模块 AD: 班级绑定分身 (5条)
 * 模块 AE: 自测学生 (4条)
 * 模块 AF: 向量召回优化 (2条)
 * 
 * 执行规则:
 * 1. 每个用例独立 sub agent（独立进程）
 * 2. 页面跳转必须通过导航元素触发（R17）
 * 3. 环境不可用时严禁降级执行（R17）
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const CONFIG = {
  projectPath: '/Users/aganbai/Desktop/WorkSpace/digital-twin/src/frontend',
  backendUrl: 'http://localhost:8080',
  resultsDir: '/Users/aganbai/Desktop/WorkSpace/digital-twin/test_results'
};

// 用例定义
const TEST_CASES = {
  // 第一批：无依赖，可并行
  batch1: [
    { id: 'SM-AD01', name: '教师创建班级同步创建分身', file: 'smoke-v11-ad01.js', depends: [], critical: true },
    { id: 'SM-AD02', name: '教师禁止独立创建分身', file: 'smoke-v11-ad02.js', depends: [], critical: false },
    { id: 'SM-AD04', name: '已删除的接口返回404', file: 'smoke-v11-ad04.js', depends: [], critical: false },
    { id: 'SM-AE01', name: '教师注册自动创建自测学生', file: 'smoke-v11-ae01.js', depends: [], critical: true }
  ],
  // 第二批：依赖第一批
  batch2: [
    { id: 'SM-AD03', name: '分身列表展示班级信息', file: 'smoke-v11-ad03.js', depends: ['SM-AD01'], critical: false },
    { id: 'SM-AD05', name: '班级 is_public 设置', file: 'smoke-v11-ad05.js', depends: ['SM-AD01'], critical: false },
    { id: 'SM-AE02', name: '获取自测学生信息', file: 'smoke-v11-ae02.js', depends: ['SM-AE01'], critical: false },
    { id: 'SM-AE03', name: '自测学生自动加入班级', file: 'smoke-v11-ae03.js', depends: ['SM-AD01', 'SM-AE01'], critical: false },
    { id: 'SM-AE04', name: '重置自测学生数据', file: 'smoke-v11-ae04.js', depends: ['SM-AE01'], critical: false }
  ],
  // 第三批：需外部环境（Python服务）
  batch3: [
    { id: 'SM-AF01', name: '知识库向量召回100条', file: 'smoke-v11-af01.js', depends: [], critical: false, skipReason: '需要Python向量服务' },
    { id: 'SM-AF02', name: '知识库 scope=global 生效', file: 'smoke-v11-af02.js', depends: [], critical: false, skipReason: '需要Python向量服务' }
  ]
};

// 全局结果
const FINAL_RESULT = {
  startTime: new Date().toISOString(),
  endTime: null,
  env: {},
  batches: {},
  summary: {
    total: 0,
    passed: 0,
    failed: 0,
    skipped: 0
  },
  details: []
};

// 环境检查
function checkEnvironment() {
  console.log('\n🔍 第一步：环境可用性检查\n');
  console.log('========================================');
  
  let allPassed = true;
  
  // 1. 检查 miniprogram-automator
  try {
    require('miniprogram-automator');
    FINAL_RESULT.env.automator = '✅ 已安装';
    console.log('✅ miniprogram-automator 已安装');
  } catch (e) {
    FINAL_RESULT.env.automator = '❌ 未安装: ' + e.message;
    console.log('❌ miniprogram-automator 未安装');
    allPassed = false;
  }
  
  // 2. 检查微信开发者工具
  if (fs.existsSync('/Applications/wechatwebdevtools.app')) {
    FINAL_RESULT.env.devtools = '✅ 已安装';
    console.log('✅ 微信开发者工具已安装');
  } else {
    FINAL_RESULT.env.devtools = '❌ 未安装';
    console.log('❌ 微信开发者工具未安装');
    allPassed = false;
  }
  
  // 3. 检查 CLI 工具
  const cliPath = '/Applications/wechatwebdevtools.app/Contents/MacOS/cli';
  if (fs.existsSync(cliPath)) {
    FINAL_RESULT.env.cli = '✅ CLI 工具存在';
    console.log('✅ CLI 工具存在');
  } else {
    FINAL_RESULT.env.cli = '❌ CLI 工具不存在';
    console.log('❌ CLI 工具不存在');
    allPassed = false;
  }
  
  // 4. 检查小程序项目路径
  if (fs.existsSync(CONFIG.projectPath)) {
    FINAL_RESULT.env.projectPath = '✅ 项目路径存在';
    console.log('✅ 小程序项目路径存在');
  } else {
    FINAL_RESULT.env.projectPath = '❌ 项目路径不存在';
    console.log('❌ 小程序项目路径不存在');
    allPassed = false;
  }
  
  // 5. 检查后端服务
  try {
    const response = execSync(`curl -s ${CONFIG.backendUrl}/api/system/health`, { encoding: 'utf8', timeout: 5000 });
    const data = JSON.parse(response);
    if (data.code === 0 && data.data.status === 'running') {
      FINAL_RESULT.env.backend = '✅ 后端服务运行中';
      console.log('✅ 后端服务运行中');
    } else {
      FINAL_RESULT.env.backend = '❌ 后端服务异常';
      console.log('❌ 后端服务异常');
      allPassed = false;
    }
  } catch (e) {
    FINAL_RESULT.env.backend = '❌ 后端服务未启动: ' + e.message;
    console.log('❌ 后端服务未启动');
    allPassed = false;
  }
  
  console.log('========================================\n');
  
  if (!allPassed) {
    console.error('❌ 环境检查失败，无法继续执行（R17：严禁降级执行）');
    console.log('\n请检查以下配置：');
    console.log('1. 确保已安装 miniprogram-automator: cd src/frontend && npm install');
    console.log('2. 确保已安装微信开发者工具');
    console.log('3. 确保微信开发者工具已开启「安全设置」中的「服务端口」');
    console.log('4. 确保后端服务已启动: cd src/backend && go run cmd/server/main.go');
    process.exit(1);
  }
  
  console.log('✅ 环境检查全部通过！\n');
  return true;
}

// 执行单个用例
function runTestCase(testCase) {
  console.log(`\n🧪 执行用例: ${testCase.id} - ${testCase.name}`);
  console.log('----------------------------------------');
  
  const testFile = path.join(CONFIG.projectPath, 'e2e', testCase.file);
  
  // 检查测试文件是否存在
  if (!fs.existsSync(testFile)) {
    console.log(`⚠️ 测试文件不存在: ${testCase.file}`);
    return {
      id: testCase.id,
      status: 'skipped',
      error: '测试文件不存在'
    };
  }
  
  try {
    // 执行测试脚本
    const output = execSync(`node "${testFile}"`, {
      encoding: 'utf8',
      timeout: 120000, // 2分钟超时
      cwd: CONFIG.projectPath
    });
    
    console.log(output);
    
    // 读取结果文件
    const resultPath = path.join(CONFIG.resultsDir, `${testCase.id}_result.json`);
    if (fs.existsSync(resultPath)) {
      const result = JSON.parse(fs.readFileSync(resultPath, 'utf8'));
      return result;
    } else {
      return {
        id: testCase.id,
        status: 'failed',
        error: '未找到结果文件'
      };
    }
  } catch (error) {
    console.log(error.stdout || error.message);
    
    // 尝试读取结果文件
    const resultPath = path.join(CONFIG.resultsDir, `${testCase.id}_result.json`);
    if (fs.existsSync(resultPath)) {
      const result = JSON.parse(fs.readFileSync(resultPath, 'utf8'));
      return result;
    }
    
    return {
      id: testCase.id,
      status: 'failed',
      error: error.message
    };
  }
}

// 执行批次
function runBatch(batchName, testCases) {
  console.log(`\n========================================`);
  console.log(`  执行批次: ${batchName}`);
  console.log(`========================================`);
  
  const results = [];
  
  for (const testCase of testCases) {
    // 检查是否需要跳过
    if (testCase.skipReason) {
      console.log(`\n⏸️ ${testCase.id}: ${testCase.name}`);
      console.log(`   跳过原因: ${testCase.skipReason}`);
      results.push({
        id: testCase.id,
        name: testCase.name,
        status: 'skipped',
        error: testCase.skipReason
      });
      FINAL_RESULT.summary.skipped++;
      continue;
    }
    
    const result = runTestCase(testCase);
    results.push(result);
    
    // 更新统计
    FINAL_RESULT.summary.total++;
    if (result.status === 'passed') {
      FINAL_RESULT.summary.passed++;
    } else if (result.status === 'failed') {
      FINAL_RESULT.summary.failed++;
    } else if (result.status === 'skipped') {
      FINAL_RESULT.summary.skipped++;
    }
    
    // 如果是关键用例且失败，停止执行
    if (testCase.critical && result.status === 'failed') {
      console.log(`\n❌ 关键用例 ${testCase.id} 失败，停止后续执行`);
      break;
    }
  }
  
  FINAL_RESULT.batches[batchName] = results;
  return results;
}

// 生成最终报告
function generateReport() {
  FINAL_RESULT.endTime = new Date().toISOString();
  
  console.log('\n\n');
  console.log('╔════════════════════════════════════════════════════════════╗');
  console.log('║       Phase 3c: 迭代11 端到端冒烟验证 - 最终报告           ║');
  console.log('╚════════════════════════════════════════════════════════════╝');
  
  console.log('\n【环境检查结果】');
  Object.entries(FINAL_RESULT.env).forEach(([key, value]) => {
    console.log(`  ${key}: ${value}`);
  });
  
  console.log('\n【批次执行结果】');
  Object.entries(FINAL_RESULT.batches).forEach(([batchName, results]) => {
    console.log(`\n  ${batchName}:`);
    results.forEach(r => {
      const icon = r.status === 'passed' ? '✅' : r.status === 'failed' ? '❌' : '⏸️';
      console.log(`    ${icon} ${r.id}: ${r.name}`);
      if (r.error) console.log(`       原因: ${r.error}`);
    });
  });
  
  console.log('\n【执行统计】');
  console.log(`  开始时间: ${FINAL_RESULT.startTime}`);
  console.log(`  结束时间: ${FINAL_RESULT.endTime}`);
  console.log(`  总用例数: ${FINAL_RESULT.summary.total}`);
  console.log(`  ✅ 通过: ${FINAL_RESULT.summary.passed}`);
  console.log(`  ❌ 失败: ${FINAL_RESULT.summary.failed}`);
  console.log(`  ⏸️ 跳过: ${FINAL_RESULT.summary.skipped}`);
  
  console.log('\n╔════════════════════════════════════════════════════════════╗');
  if (FINAL_RESULT.summary.failed === 0) {
    console.log('║                    ✅ 所有用例通过                         ║');
  } else {
    console.log('║                    ❌ 存在失败用例                         ║');
  }
  console.log('╚════════════════════════════════════════════════════════════╝\n');
  
  // 保存完整报告
  const reportPath = path.join(CONFIG.resultsDir, 'smoke_v11_final_report.json');
  fs.writeFileSync(reportPath, JSON.stringify(FINAL_RESULT, null, 2));
  console.log(`📄 完整报告已保存: ${reportPath}`);
}

// 主函数
function main() {
  console.log('╔════════════════════════════════════════════════════════════╗');
  console.log('║     Phase 3c: 迭代11 端到端冒烟验证                        ║');
  console.log('║     使用 miniprogram-automator SDK                       ║');
  console.log('╚════════════════════════════════════════════════════════════╝');
  
  // 创建结果目录
  if (!fs.existsSync(CONFIG.resultsDir)) {
    fs.mkdirSync(CONFIG.resultsDir, { recursive: true });
  }
  
  // 第一步：环境检查
  checkEnvironment();
  
  // 执行各批次
  console.log('\n🚀 开始执行冒烟测试用例...\n');
  
  // 第一批：无依赖
  runBatch('第一批（无依赖）', TEST_CASES.batch1);
  
  // 检查第一批关键用例是否通过
  const batch1Results = FINAL_RESULT.batches['第一批（无依赖）'] || [];
  const criticalFailed = batch1Results.some(r => {
    const tc = TEST_CASES.batch1.find(t => t.id === r.id);
    return tc && tc.critical && r.status === 'failed';
  });
  
  if (!criticalFailed) {
    // 第二批：依赖第一批
    runBatch('第二批（依赖第一批）', TEST_CASES.batch2);
    
    // 第三批：需外部环境
    runBatch('第三批（需外部环境）', TEST_CASES.batch3);
  } else {
    console.log('\n❌ 第一批关键用例失败，跳过后续批次');
    // 标记第二批和第三批为跳过
    TEST_CASES.batch2.forEach(tc => {
      FINAL_RESULT.summary.total++;
      FINAL_RESULT.summary.skipped++;
    });
    TEST_CASES.batch3.forEach(tc => {
      FINAL_RESULT.summary.total++;
      FINAL_RESULT.summary.skipped++;
    });
  }
  
  // 生成报告
  generateReport();
  
  // 返回退出码
  process.exit(FINAL_RESULT.summary.failed > 0 ? 1 : 0);
}

// 运行
main();
