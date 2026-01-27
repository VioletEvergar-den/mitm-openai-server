const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: false });
  const context = await browser.newContext();
  const page = await context.newPage();

  try {
    console.log('=== MITM OpenAI Server 注册功能测试 ===');
    
    // 访问注册页面
    console.log('1. 正在访问注册页面...');
    const response = await page.goto('http://localhost:8081/ui/register');
    console.log('页面响应状态:', response.status());
    
    // 等待页面加载
    await page.waitForLoadState('networkidle');
    console.log('页面加载完成');
    
    // 检查页面元素
    const title = await page.title();
    console.log('页面标题:', title);
    
    // 检查是否存在注册表单元素
    const usernameInput = await page.$('input[placeholder="用户名（至少3个字符）"]');
    const passwordInput = await page.$('input[placeholder="密码（至少6个字符）"]');
    const confirmPasswordInput = await page.$('input[placeholder="确认密码"]');
    const submitButton = await page.$('button[type="submit"]');
    
    console.log('用户名输入框存在:', !!usernameInput);
    console.log('密码输入框存在:', !!passwordInput);
    console.log('确认密码输入框存在:', !!confirmPasswordInput);
    console.log('提交按钮存在:', !!submitButton);
    
    // 生成随机用户名
    const timestamp = Date.now();
    const randomUsername = 'testuser_' + Math.floor(Math.random() * 10000) + '_' + timestamp;
    const password = 'testpassword123';
    
    console.log('2. 正在填写注册表单...');
    console.log('用户名:', randomUsername);
    console.log('密码: testpassword123');
    
    // 填写注册表单
    await page.fill('input[placeholder="用户名（至少3个字符）"]', randomUsername);
    await page.fill('input[placeholder="密码（至少6个字符）"]', password);
    await page.fill('input[placeholder="确认密码"]', password);
    
    // 检查输入值
    const usernameValue = await page.inputValue('input[placeholder="用户名（至少3个字符）"]');
    const passwordValue = await page.inputValue('input[placeholder="密码（至少6个字符）"]');
    console.log('输入的用户名:', usernameValue);
    console.log('输入的密码长度:', passwordValue.length);
    
    // 提交表单
    console.log('3. 正在提交注册表单...');
    await page.click('button[type="submit"]');
    
    // 等待网络请求完成
    console.log('等待网络请求完成...');
    await page.waitForLoadState('networkidle');
    
    // 检查网络请求
    page.on('response', response => {
      if (response.url().includes('/register')) {
        console.log('注册API响应状态:', response.status());
        console.log('注册API响应URL:', response.url());
      }
    });
    
    // 等待一段时间以便观察结果
    await page.waitForTimeout(5000);
    
    // 检查页面内容
    const pageContent = await page.content();
    console.log('页面内容长度:', pageContent.length);
    
    // 检查是否注册成功
    const successMessage = await page.locator('.ant-alert-success').first();
    const errorMessage = await page.locator('.ant-alert-error').first();
    
    if (await successMessage.isVisible()) {
      console.log('✅ 注册成功！');
      const successText = await successMessage.textContent();
      console.log('成功消息:', successText.trim());
    } else if (await errorMessage.isVisible()) {
      console.log('❌ 注册失败！');
      const errorText = await errorMessage.textContent();
      console.log('错误消息:', errorText.trim());
    } else {
      console.log('⚠️  未找到明确的成功或失败消息');
      // 检查是否跳转到了登录页面
      const currentUrl = page.url();
      console.log('当前页面URL:', currentUrl);
      if (currentUrl.includes('/login')) {
        console.log('✅ 注册成功并跳转到了登录页面');
      }
    }
    
    // 检查页面上是否有特定文本
    const hasSuccessText = await page.getByText('注册成功').isVisible();
    const hasErrorText = await page.getByText('注册失败').isVisible();
    const hasLoginText = await page.getByText('立即登录').isVisible();
    
    console.log('页面包含"注册成功":', hasSuccessText);
    console.log('页面包含"注册失败":', hasErrorText);
    console.log('页面包含"立即登录":', hasLoginText);
    
  } catch (error) {
    console.error('❌ 测试过程中出现错误:', error);
  } finally {
    console.log('4. 测试完成，关闭浏览器');
    await browser.close();
  }
})();