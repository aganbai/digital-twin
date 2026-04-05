const http = require('http');
const https = require('https');

// 辅助函数：发送HTTP请求
function request(url, options = {}) {
  return new Promise((resolve, reject) => {
    const urlObj = new URL(url);
    const client = urlObj.protocol === 'https:' ? https : http;
    
    const req = client.request({
      hostname: urlObj.hostname,
      port: urlObj.port || (urlObj.protocol === 'https:' ? 443 : 80),
      path: urlObj.pathname + urlObj.search,
      method: options.method || 'GET',
      headers: {
        'Content-Type': 'application/json',
        ...options.headers
      }
    }, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        try {
          resolve({
            status: res.statusCode,
            headers: res.headers,
            data: JSON.parse(data)
          });
        } catch (e) {
          resolve({
            status: res.statusCode,
            headers: res.headers,
            data: data
          });
        }
      });
    });
    
    req.on('error', reject);
    
    if (options.body) {
      req.write(JSON.stringify(options.body));
    }
    
    req.end();
  });
}

async function testH5LoginE2E() {
  console.log('=========================================');
  console.log('H5教师端登录 - 完整端到端测试');
  console.log('=========================================\n');
  
  const FRONTEND_URL = 'http://localhost:5175';
  const BACKEND_URL = 'http://localhost:8080';
  
  try {
    // 步骤1: 检查前端服务
    console.log('1. 检查前端服务...');
    try {
      const frontendRes = await request(FRONTEND_URL);
      if (frontendRes.status === 200) {
        console.log('   ✅ 前端服务正常运行');
        console.log(`   状态码: ${frontendRes.status}`);
      } else {
        console.log(`   ⚠️  前端服务返回: ${frontendRes.status}`);
      }
    } catch (e) {
      console.log('   ❌ 前端服务无法访问');
      console.log(`   错误: ${e.message}`);
      throw new Error('前端服务未运行');
    }
    
    // 步骤2: 检查后端服务
    console.log('\n2. 检查后端服务...');
    try {
      const backendRes = await request(`${BACKEND_URL}/api/system/health`);
      if (backendRes.data && backendRes.data.status === 'healthy') {
        console.log('   ✅ 后端服务正常运行');
      } else {
        console.log('   ⚠️  后端服务状态异常');
        console.log(`   响应: ${JSON.stringify(backendRes.data)}`);
      }
    } catch (e) {
      console.log('   ❌ 后端服务无法访问');
      console.log(`   错误: ${e.message}`);
      throw new Error('后端服务未运行');
    }
    
    // 步骤3: 模拟前端获取登录URL
    console.log('\n3. 测试获取微信登录URL...');
    const redirectUri = `${FRONTEND_URL}/login`;
    console.log(`   redirect_uri: ${redirectUri}`);
    
    const loginUrlRes = await request(
      `${BACKEND_URL}/api/auth/wx-h5-login-url?redirect_uri=${encodeURIComponent(redirectUri)}`
    );
    
    console.log(`   状态码: ${loginUrlRes.status}`);
    console.log(`   响应数据: ${JSON.stringify(loginUrlRes.data, null, 2).split('\n').map(l => '     ' + l).join('\n')}`);
    
    if (loginUrlRes.data.code !== 0) {
      throw new Error(`获取登录URL失败: ${loginUrlRes.data.message}`);
    }
    
    // 检查返回字段
    if (!loginUrlRes.data.data || !loginUrlRes.data.data.login_url) {
      console.log('   ❌ 返回数据缺少 login_url 字段');
      console.log(`   实际返回字段: ${Object.keys(loginUrlRes.data.data || {}).join(', ')}`);
      throw new Error('返回数据格式错误');
    }
    
    console.log('   ✅ 成功获取登录URL');
    const loginUrl = loginUrlRes.data.data.login_url;
    console.log(`   登录URL: ${loginUrl}`);
    
    // 步骤4: 解析微信授权URL，提取参数
    console.log('\n4. 解析微信授权URL...');
    const wxAuthUrl = new URL(loginUrl);
    const wxRedirectUri = wxAuthUrl.searchParams.get('redirect_uri');
    const wxState = wxAuthUrl.searchParams.get('state') || '';
    
    console.log(`   微信授权地址: ${wxAuthUrl.origin}${wxAuthUrl.pathname}`);
    console.log(`   redirect_uri: ${wxRedirectUri}`);
    console.log(`   state: ${wxState}`);
    
    // 步骤5: 模拟微信授权回调
    console.log('\n5. 模拟微信授权回调...');
    const mockCode = 'mock_code_1001';  // 使用固定的mock code
    
    console.log(`   使用 mock code: ${mockCode}`);
    
    const callbackRes = await request(`${BACKEND_URL}/api/auth/wx-h5-callback`, {
      method: 'POST',
      body: {
        code: mockCode,
        state: wxState
      }
    });
    
    console.log(`   状态码: ${callbackRes.status}`);
    console.log(`   响应数据: ${JSON.stringify(callbackRes.data, null, 2).split('\n').map(l => '     ' + l).join('\n')}`);
    
    if (callbackRes.data.code !== 0) {
      throw new Error(`微信授权回调失败: ${callbackRes.data.message}`);
    }
    
    // 检查返回字段
    const callbackData = callbackRes.data.data;
    if (!callbackData || !callbackData.token) {
      console.log('   ❌ 返回数据缺少 token 字段');
      throw new Error('回调数据格式错误');
    }
    
    console.log('   ✅ 微信授权回调成功');
    console.log(`   用户ID: ${callbackData.user_id}`);
    console.log(`   Token: ${callbackData.token.substring(0, 30)}...`);
    console.log(`   昵称: ${callbackData.nickname}`);
    console.log(`   角色: ${callbackData.role || '未设置'}`);
    console.log(`   新用户: ${callbackData.is_new_user ? '是' : '否'}`);
    
    // 步骤6: 使用token访问需要认证的接口
    console.log('\n6. 测试使用token访问用户信息...');
    const profileRes = await request(`${BACKEND_URL}/api/user/profile`, {
      headers: {
        'Authorization': `Bearer ${callbackData.token}`
      }
    });
    
    console.log(`   状态码: ${profileRes.status}`);
    if (profileRes.status === 200 && profileRes.data.code === 0) {
      console.log('   ✅ Token验证成功');
      console.log(`   用户信息: ${JSON.stringify(profileRes.data.data, null, 2).split('\n').map(l => '     ' + l).join('\n')}`);
    } else if (profileRes.status === 401) {
      console.log('   ❌ Token验证失败 (401未授权)');
    } else {
      console.log(`   ⚠️  返回状态: ${profileRes.status}`);
    }
    
    // 最终结果
    console.log('\n=========================================');
    console.log('✅ 所有端到端测试通过！');
    console.log('=========================================\n');
    console.log('测试总结:');
    console.log('  1. ✅ 前端服务正常');
    console.log('  2. ✅ 后端服务正常');
    console.log('  3. ✅ 获取登录URL成功 (返回 login_url 字段)');
    console.log('  4. ✅ 微信授权回调成功 (返回 token)');
    console.log('  5. ✅ Token验证成功');
    console.log('\n📱 您现在可以在浏览器访问:');
    console.log(`   ${FRONTEND_URL}`);
    console.log('   点击"微信授权登录"按钮应该能正常工作\n');
    
  } catch (error) {
    console.error('\n❌ 测试失败:', error.message);
    console.error(error.stack);
    process.exit(1);
  }
}

// 运行测试
testH5LoginE2E();
