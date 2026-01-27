const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: false });
  const context = await browser.newContext();
  const page = await context.newPage();

  try {
    console.log('=== MITM OpenAI Server 完整注册登录测试 ===');
    
    // 生成随机用户名
    const timestamp = Date.now();
    const randomUsername = 'testuser_' + Math.floor(Math.random() * 10000) + '_' + timestamp;
    const password = 'testpassword123';
    
    console.log('1. 正在访问注册页面...');
    await page.goto('http://localhost:8081/ui/register');
    await page.waitForLoadState('networkidle');
    
    console.log('2. 正在填写注册表单...');
    console.log('用户名:', randomUsername);
    
    // 填写注册表单
    await page.fill('input[placeholder="用户名（至少3个字符）"]', randomUsername);
    await page.fill('input[placeholder="密码（至少6个字符）"]', password);
    await page.fill('input[placeholder="确认密码"]', password);
    
    // 提交表单
    console.log('3. 正在提交注册表单...');
    await page.click('button[type="submit"]');
    
    // 等待注册完成并跳转
    await page.waitForTimeout(3000);
    
    // 检查是否跳转到登录页面
    const currentUrl = page.url();
    console.log('注册后页面URL:', currentUrl);
    
    if (currentUrl.includes('/login')) {
      console.log('✅ 注册成功并跳转到了登录页面');
      
      // 现在尝试使用新账号登录
      console.log('4. 正在尝试使用新账号登录...');
      
      // 填写登录表单
      await page.fill('input[placeholder="用户名"]', randomUsername);
      await page.fill('input[placeholder="密码"]', password);
      
      // 提交登录表单
      await page.click('button[type="submit"]');
      
      // 等待登录完成
      await page.waitForTimeout(3000);
      
      // 检查登录是否成功
      const afterLoginUrl = page.url();
      console.log('登录后页面URL:', afterLoginUrl);
      
      if (afterLoginUrl.includes('/requests') || afterLoginUrl.includes('/dashboard')) {
        console.log('✅ 登录成功！');
      } else if (afterLoginUrl.includes('/login')) {
        // 检查是否有错误消息
        const errorMessage = await page.locator('.ant-alert-error').first();
        if (await errorMessage.isVisible()) {
          const errorText = await errorMessage.textContent();
          console.log('❌ 登录失败:', errorText.trim());
        } else {
          console.log('⚠️ 登录后仍在登录页面，但未显示错误消息');
        }
      } else {
        console.log('⚠️ 登录后跳转到了未知页面:', afterLoginUrl);
      }
    } else {
      console.log('❌ 注册后未跳转到登录页面');
      
      // 检查是否有错误消息
      const errorMessage = await page.locator('.ant-alert-error').first();
      if (await errorMessage.isVisible()) {
        const errorText = await errorMessage.textContent();
        console.log('注册错误消息:', errorText.trim());
      }
    }
    
    // 等待一段时间以便观察结果
    await page.waitForTimeout(5000);
    
  } catch (error) {
    console.error('❌ 测试过程中出现错误:', error);
  } finally {
    console.log('5. 测试完成，关闭浏览器');
    await browser.close();
  }
})();