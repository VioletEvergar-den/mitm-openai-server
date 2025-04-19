# MITM OpenAI Server 前端界面

这是MITM OpenAI Server的React前端界面，用于监控和分析OpenAI API请求。

## 开发指南

### 开发环境设置

1. 安装依赖：
   ```bash
   npm install
   ```

2. 启动开发服务器：
   ```bash
   npm run dev
   ```
   
   开发服务器默认在 `http://localhost:5173` 启动，并提供热重载功能。

### 开发模式

开发模式下，前端和后端分开运行：

1. 在一个终端启动后端服务器：
   ```bash
   cd ..
   ./scripts/build.sh --dev  # 开发模式构建后端
   ./mitm-openai-server server
   ```

2. 在另一个终端启动前端开发服务器：
   ```bash
   cd react-ui
   npm run dev
   ```

前端开发服务器会自动将API请求代理到后端服务器（localhost:8080）。

## 构建指南

### 生产环境构建

要构建生产环境版本，您有两种选择：

1. 使用构建脚本（推荐）：
   ```bash
   cd ..
   ./scripts/build.sh --prod
   ```
   此脚本会自动构建前端和后端，并将前端资源正确部署。

2. 手动构建：
   ```bash
   npm run build
   ```
   构建后的文件将位于 `dist` 目录中。

## 项目结构

- `src/` - 源代码目录
  - `components/` - React组件
  - `pages/` - 页面组件
  - `services/` - API服务
  - `contexts/` - React上下文
  - `styles/` - 样式文件
  - `types/` - TypeScript类型定义

## 技术栈

- React 18
- TypeScript
- Vite
- React Router
- Axios 