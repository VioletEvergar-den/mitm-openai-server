


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

- 页面配色和风格看上去很丑，你引入AntDesign了吗？用AntDesign的布局和组件美化一下页面 

- 详情页要支持导出当前对话

# 列表页
- 列表页应该支持让用户能够选择分页大小，默认分页为20条每页
- 排序条件不支持让用户选择，就默认按照时间倒序排序；
- 然后就是ID列有点窄了，调整宽一些，要能够展示得开uuid；
- 然后就是在路径列后面加一列，默认展示响应中的choices的最后一个元素的message中的content，如果展示不开的注意省略号






































































