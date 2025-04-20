


- 在website目录下使用React+TypeScript，给这个项目写一个官网，这个官网会使用GitHub pages来部署，所以路由方式要使用hash路由

- 写个GitHub action，在website下的内容有改动的时候，自动部署GitHub pages分支
- 写个GitHub action，在每次提交代码的时候，自动运行所有的单元测试 

# 应用场景
- 安全攻防以dnslog的方式外带提示词
- 开发AI应用辅助调试


# 概念重构
- 把所有涉及到 捕获的请求列表 请求列表 等概念的地方的请求，在页面上的展示的文案都替换为对话，不是请求，而是对话 
- 导航栏上的 中间人OpenAI API服务器，显示为 MITM OpenAI API Server ，这样看起来逼格高一些 



- 为什么Openapi配置教程页面的测试对话，每次对话都会产生两条请求？/v1/chat/completions和	/chat/completions，看上去	/chat/completions是不必要的 


- 详情页要支持导出当前对话，导出的格式为JSON格式
- 详情页要支持返回到列表页 
- 详情页的上一条、下一条，要支持全局快捷键，键盘左键上一条，键盘右键下一条，同时在按钮上把键盘快捷键也显示出来
- 请求标签和响应标签，要能够记得住用户上次选的是哪个标签，下次自动切换到这个标签；










































































