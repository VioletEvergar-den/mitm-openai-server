# MITM OpenAI Server

<div align="center">

**一个简单易用的 OpenAI API 请求监控与调试工具**

[![Go](https://img.shields.io/badge/Go-1.19+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18+-61DAFB?style=flat&logo=react)](https://react.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

[功能介绍](#-功能介绍) • [快速开始](#-快速开始) • [使用指南](#-使用指南) • [常见问题](#-常见问题)

</div>

---

## 📖 这是什么？

**MITM OpenAI Server** 是一个"中间人"服务器，可以帮你：

- 🎯 **拦截** OpenAI API 的请求和响应
- 📊 **查看** 完整的请求内容（参数、头部、响应体等）
- 🔧 **调试** 你的 AI 应用程序
- 📝 **记录** 所有 API 调用历史

### 🤔 通俗理解

想象你在打电话（你的程序调用 OpenAI API），这个工具就像一个"电话录音机"：

```
你的程序 → [MITM Server 录音] → OpenAI API
                ↓
            保存记录，可以回放查看
```

---

## ✨ 功能介绍

| 功能 | 说明 |
|------|------|
| 📡 **请求拦截** | 捕获所有发往 OpenAI API 的请求 |
| 📋 **响应记录** | 完整保存 API 返回的内容 |
| 🖥️ **Web 界面** | 可视化查看请求历史，支持语法高亮 |
| 🔐 **多用户支持** | 每个用户的数据相互隔离，互不干扰 |
| 🔄 **代理模式** | 转发请求到真实 API，同时记录 |
| 🎭 **模拟模式** | 返回模拟数据，无需真实 API 密钥 |
| 📤 **数据导出** | 支持导出请求数据 |

---

## 🚀 快速开始

### 第一步：准备环境

你需要安装两个软件：

#### 1. Go 语言环境

**Windows:**
```powershell
# 下载安装包：https://go.dev/dl/
# 或使用 winget 安装
winget install GoLang.Go
```

**Mac:**
```bash
brew install go
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt update
sudo apt install golang-go
```

#### 2. Node.js 环境

**Windows:**
```powershell
# 下载安装包：https://nodejs.org/
# 或使用 winget 安装
winget install OpenJS.NodeJS.LTS
```

**Mac:**
```bash
brew install node
```

**Linux (Ubuntu/Debian):**
```bash
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt install nodejs
```

#### ✅ 验证安装

打开终端，运行以下命令确认安装成功：

```bash
go version      # 应显示 go version go1.x.x ...
node -v         # 应显示 v18.x.x 或更高
npm -v          # 应显示 9.x.x 或更高
```

---

### 第二步：下载项目

```bash
# 克隆项目到本地
git clone https://github.com/VioletEvergar-den/mitm-openai-server.git

# 进入项目目录
cd mitm-openai-server
```

---

### 第三步：启动服务器

我们提供了一个智能启动脚本，自动完成所有工作：

**Linux/Mac:**
```bash
# 添加执行权限
chmod +x start.sh

# 启动服务器
./start.sh
```

**Windows:**
```powershell
# 需要在 Git Bash 或 WSL 中运行
# 或者手动执行以下步骤
```

<details>
<summary>📦 手动构建步骤（Windows 用户）</summary>

```powershell
# 1. 构建前端
cd react-ui
npm install
npm run build
cd ..

# 2. 构建后端
go build -o mitm-openai-server.exe .

# 3. 启动服务器
.\mitm-openai-server.exe server --port 8081
```

</details>

---

### 第四步：访问 Web 界面

启动成功后，你会看到类似这样的输出：

```
┌─────────────────────────────────────────────────┐
│            MITM OpenAI Server 已启动            │
│              (多用户数据库版本)                  │
└─────────────────────────────────────────────────┘
登录地址: http://localhost:8081/ui/login
用户名:   root
密码:     jKYh21n2ruTQ
```

1. 打开浏览器访问 `http://localhost:8081/ui/login`
2. 输入显示的用户名和密码
3. 登录成功后即可查看请求历史

---

## 📚 使用指南

### 两种工作模式

#### 🎭 模拟模式（Mock Mode）

不需要真实的 OpenAI API 密钥，服务器会返回模拟的响应数据。

**适用场景：**
- 开发测试阶段
- 学习 OpenAI API 格式
- 演示和教学

**启动方式：**
```bash
./mitm-openai-server server --port 8081
```

#### 🔄 代理模式（Proxy Mode）

将请求转发到真实的 OpenAI API，同时记录所有请求和响应。

**适用场景：**
- 调试生产环境问题
- 分析 API 调用情况
- 监控 API 使用量

**启动方式：**
```bash
./mitm-openai-server server --port 8081 \
  --proxy-mode \
  --target-url https://api.openai.com \
  --target-auth-type token \
  --target-token sk-your-api-key
```

---

### 配置你的程序

将你的程序（如 ChatGPT-Next-Web、LobeChat 等）的 API 地址改为本服务器：

**原来：**
```
API 地址: https://api.openai.com
API 密钥: sk-xxxxx
```

**改为：**
```
API 地址: http://localhost:8081
API 密钥: 任意值（模拟模式）或真实密钥（代理模式）
```

---

### Web 界面功能

| 页面 | 功能 |
|------|------|
| 📋 **请求列表** | 查看所有拦截到的请求，支持分页 |
| 📄 **请求详情** | 查看完整的请求/响应内容，支持语法高亮 |
| ⚙️ **设置** | 管理用户账户和系统配置 |
| 📖 **使用指南** | 查看帮助文档 |

---

### 启动脚本参数

```bash
./start.sh [选项]

选项:
  -h, --help       显示帮助信息
  -p, --port       指定端口 (默认: 8081)
  -v, --verbose    显示详细日志
  -s, --skip-build 跳过构建，快速启动
  -c, --clean      清理数据后启动
  --install-deps   自动安装缺失的依赖

示例:
  ./start.sh                    # 默认启动
  ./start.sh -p 3000            # 使用端口 3000
  ./start.sh --install-deps     # 自动安装依赖
  ./start.sh -s                 # 跳过构建快速启动
```

---

## 🔧 常见问题

### Q1: 启动时提示 "未找到 Go"

**原因：** 没有安装 Go 或者 Go 没有加入系统环境变量。

**解决：**
1. 确认已安装 Go：`go version`
2. 如果已安装但找不到，重启终端或重启电脑
3. 或使用 `./start.sh --install-deps` 自动安装

---

### Q2: 启动时提示 "数据目录无写权限"

**原因：** 数据目录权限不足。

**解决：**
```bash
# Linux/Mac
sudo chown -R $(whoami):$(whoami) ./data

# 或者让脚本自动修复
./start.sh  # 脚本会自动尝试修复权限
```

---

### Q3: 前端构建失败

**原因：** npm 依赖安装问题。

**解决：**
```bash
cd react-ui
rm -rf node_modules package-lock.json
npm install --legacy-peer-deps
npm run build
```

---

### Q4: 端口被占用

**原因：** 8081 端口已被其他程序使用。

**解决：**
```bash
# 使用其他端口启动
./start.sh -p 8082

# 或者查找并关闭占用端口的程序
# Linux/Mac
lsof -i :8081
kill -9 <PID>

# Windows
netstat -ano | findstr :8081
taskkill /PID <PID> /F
```

---

### Q5: 登录密码忘记了

**解决：**
查看 `data/login.json` 文件，里面保存了用户名和密码。

```bash
cat data/login.json
```

或者删除该文件，重启服务器会自动生成新密码。

---

## 📁 项目结构

```
mitm-openai-server/
├── cmd/                    # 命令行入口
│   └── server.go           # 服务器启动命令
├── pkg/                    # 核心代码
│   ├── api/                # API 处理器
│   ├── openai/             # OpenAI 接口实现
│   ├── storage/            # 数据存储
│   └── server/             # HTTP 服务器
├── react-ui/               # 前端界面
│   └── src/                # React 源码
├── data/                   # 数据目录（自动创建）
│   ├── mitm_server.db      # SQLite 数据库
│   └── login.json          # 登录凭据
├── start.sh                # 智能启动脚本
└── README.md               # 说明文档
```

---

## 🛡️ 安全提示

- ⚠️ 本工具仅供开发和测试使用，不建议在生产环境使用
- ⚠️ 请勿将服务器暴露在公网，除非你了解相关风险
- ⚠️ API 密钥等敏感信息会保存在本地数据库中，请注意保护

---

## 📄 许可证

本项目采用 [MIT 许可证](LICENSE)。

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

---

<div align="center">

**如果这个项目对你有帮助，请给一个 ⭐ Star 支持一下！**

Made with ❤️ by [VioletEvergar-den](https://github.com/VioletEvergar-den)

</div>
