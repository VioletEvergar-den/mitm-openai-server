const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: false });
  const context = await browser.newContext();
  const page = await context.newPage();

  try {
    // 访问注册页面
    console.log('正在访问注册页面...');
    await page.goto('http://localhost:8081/ui/register');
    
    // 等待页面加载
    await page.waitForLoadState('networkidle');
    
    // 生成随机用户名
    const randomUsername = 'testuser_' + Math.floor(Math.random() * 10000);
    const password = 'testpassword123';
    
    console.log('正在填写注册表单...');
    // 填写注册表单
    await page.fill('input[placeholder="用户名（至少3个字符）"]', randomUsername);
    await page.fill('input[placeholder="密码（至少6个字符）"]', password);
    await page.fill('input[placeholder="确认密码"]', password);
    
    // 提交表单
    console.log('正在提交注册表单...');
    await page.click('button[type="submit"]');
    
    // 等待响应
    await page.waitForTimeout(3000);
    
    // 检查是否注册成功
    const successMessage = await page.locator('.ant-alert-success').first();
    const errorMessage = await page.locator('.ant-alert-error').first();
    
    if (await successMessage.isVisible()) {
      console.log('✅ 注册成功！');
      const successText = await successMessage.textContent();
      console.log('成功消息:', successText);
    } else if (await errorMessage.isVisible()) {
      console.log('❌ 注册失败！');
      const errorText = await errorMessage.textContent();
      console.log('错误消息:', errorText);
    } else {
      console.log('⚠️  未找到明确的成功或失败消息');
      // 检查是否跳转到了登录页面
      const currentUrl = page.url();
      if (currentUrl.includes('/login')) {
        console.log('✅ 注册成功并跳转到了登录页面');
      } else {
        console.log('当前页面URL:', currentUrl);
      }
    }
    
    // 等待一段时间以便观察结果
    await page.waitForTimeout(5000);
    
  } catch (error) {
    console.error('测试过程中出现错误:', error);
  } finally {
    await browser.close();
  }
})();