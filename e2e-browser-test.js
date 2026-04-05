const puppeteer = require('puppeteer');

async function testH5Login() {
  console.log('=========================================');
  console.log('H5教师端登录 - 浏览器端到端测试');
  console.log('=========================================\n');

  let browser;
  let page;
  
  try {
    // 启动浏览器
    console.log('1. 启动浏览器...');
    browser = await puppeteer.launch({
      headless: 'new',
      args: ['--no-sandbox', '--disable-setuid-sandbox']
    });
    page = await browser.newPage();
    
    // 监听网络请求
    const requests = [];
    page.on('request', request => {
      if (request.url().includes('/api/auth/')) {
        requests.push({
          url: request.url(),
          method: request.method()
        });
      }
    });

    // 监听响应
    const responses = [];
    page.on('response', async response => {
      if (response.url().includes('/api/auth/')) {
        try {
          const data = await response.json();
          responses.push({
            url: response.url(),
            status: response.status(),
            data: data
          });
        } catch (e) {}
      }
    });

    // 监听控制台错误
    const consoleErrors = [];
    page.on('console', msg => {
      if (msg.type() === 'error') {
        consoleErrors.push(msg.text());
      }
    });

    // 访问登录页面
    console.log('2. 访问登录页面 http://localhost:5175...');
    await page.goto('http://localhost:5175', {
      waitUntil: 'networkidle0',
      timeout: 10000
    });
    console.log('   ✅ 页面加载成功');

    // 截图
    await page.screenshot({ path: '/tmp/h5-login-page.png' });
    console.log('   📸 截图已保存: /tmp/h5-login-page.png');

    // 检查页面标题
    const title = await page.title();
    console.log(`   页面标题: ${title}`);

    // 查找登录按钮
    console.log('\n3. 查找登录按钮...');
    const loginButton = await page.waitForSelector('button:has-text("微信授权登录")', {
      timeout: 5000
    }).catch(() => null);
    
    if (!loginButton) {
      console.log('   ❌ 未找到登录按钮');
      console.log('   页面HTML:', await page.content());
      throw new Error('未找到登录按钮');
    }
    console.log('   ✅ 找到登录按钮');

    // 点击登录按钮
    console.log('\n4. 点击登录按钮...');
    await Promise.all([
      page.waitForNavigation({ waitUntil: 'networkidle0', timeout: 5000 }).catch(() => {}),
      page.click('button:has-text("微信授权登录")')
    ]);
    
    // 等待一下让请求完成
    await page.waitForTimeout(2000);

    // 检查是否跳转到了微信授权页面
    const currentUrl = page.url();
    console.log(`   当前URL: ${currentUrl}`);
    
    if (currentUrl.includes('mock.weixin.com')) {
      console.log('   ✅ 成功跳转到微信授权页面');
      
      // Mock微信授权回调 - 提取redirect_uri并模拟回调
      const urlObj = new URL(currentUrl);
      const redirectUri = urlObj.searchParams.get('redirect_uri');
      const state = urlObj.searchParams.get('state') || '';
      
      if (redirectUri) {
        console.log(`\n5. 模拟微信授权回调...`);
        console.log(`   redirect_uri: ${redirectUri}`);
        
        // 构造回调URL（带上mock code）
        const callbackUrl = `${redirectUri}?code=mock_code_1001&state=${state}`;
        console.log(`   回调URL: ${callbackUrl}`);
        
        // 访问回调URL
        await page.goto(callbackUrl, {
          waitUntil: 'networkidle0',
          timeout: 10000
        });
        
        await page.waitForTimeout(2000);
        
        const afterCallbackUrl = page.url();
        console.log(`   回调后URL: ${afterCallbackUrl}`);
        
        // 截图
        await page.screenshot({ path: '/tmp/h5-after-callback.png' });
        console.log('   📸 回调后截图: /tmp/h5-after-callback.png');
        
        // 检查是否跳转到了首页
        if (afterCallbackUrl.includes('/home')) {
          console.log('   ✅ 登录成功，已跳转到首页');
        } else if (afterCallbackUrl.includes('/role-select')) {
          console.log('   ✅ 登录成功，已跳转到角色选择页');
        } else {
          console.log('   ⚠️  登录流程可能未完成');
        }
      }
    } else if (currentUrl.includes('localhost:5175')) {
      console.log('   ⚠️  仍在登录页面，可能未成功跳转');
    }

    // 输出网络请求信息
    console.log('\n=========================================');
    console.log('网络请求详情:');
    console.log('=========================================');
    console.log('\nAPI请求:');
    requests.forEach((req, i) => {
      console.log(`  ${i + 1}. ${req.method} ${req.url}`);
    });
    
    console.log('\nAPI响应:');
    responses.forEach((res, i) => {
      console.log(`  ${i + 1}. ${res.status} ${res.url}`);
      console.log(`     数据:`, JSON.stringify(res.data, null, 2).split('\n').map(l => '     ' + l).join('\n'));
    });

    if (consoleErrors.length > 0) {
      console.log('\n控制台错误:');
      consoleErrors.forEach((err, i) => {
        console.log(`  ${i + 1}. ${err}`);
      });
    }

    console.log('\n=========================================');
    console.log('✅ 端到端测试完成');
    console.log('=========================================\n');

  } catch (error) {
    console.error('\n❌ 测试失败:', error.message);
    console.error(error.stack);
    
    // 保存错误截图
    if (page) {
      await page.screenshot({ path: '/tmp/h5-error.png' });
      console.log('\n📸 错误截图已保存: /tmp/h5-error.png');
    }
    
    process.exit(1);
  } finally {
    if (browser) {
      await browser.close();
    }
  }
}

// 运行测试
testH5Login();
