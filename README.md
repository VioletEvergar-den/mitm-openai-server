# 中间人OpenAPI服务器

这是一个中间人OpenAPI服务器，用于接收、转发和记录API请求数据。服务器可以工作在两种模式下：
1. 独立模式：返回模拟的数据响应（类似假的OpenAPI服务器）
2. 代理模式：将请求转发到真实的API服务，同时记录请求和响应

## 功能特点

- 独立模式：提供模拟的API响应
- 代理模式：充当中间人，转发请求到真实API
- 完全符合OpenAPI 3.0规范
- 捕获并记录所有API请求和响应的详细信息
- 提供REST API接口查询已记录的请求
- 可配置的数据存储目录
- 支持对目标API的各种认证方式（无认证、基本认证、令牌认证）
- 包含友好的Web UI界面查看请求历史
- 模块化设计，易于扩展

## 安装

确保你已安装Go（版本1.20或更高）。

```bash
# 克隆仓库
git clone https://github.com/llm-sec/mitm-openai-server.git
cd mitm-openai-server

# 安装依赖
go mod tidy

# 构建
go build
```

## 运行服务器

### 独立模式（模拟响应）

```bash
# 使用默认设置运行
./mitm-openapi-server api

# 指定端口和数据目录
./mitm-openapi-server api --port 9000 --data ./my-data
```

### 代理模式（转发到真实API）

```bash
# 基本代理模式
./mitm-openapi-server api --proxy-mode --target-url https://api.example.com

# 带认证的代理模式
./mitm-openapi-server api --proxy-mode --target-url https://api.example.com \
  --target-auth-type basic --target-username user --target-password pass

# 使用令牌认证
./mitm-openapi-server api --proxy-mode --target-url https://api.example.com \
  --target-auth-type token --target-token your-api-token
```

## 管理请求记录

```bash
# 列出所有记录的请求
./mitm-openapi-server requests list

# 查看特定ID的请求详情
./mitm-openapi-server requests get <request-id>

# 删除特定ID的请求
./mitm-openapi-server requests delete <request-id>

# 删除所有请求
./mitm-openapi-server requests delete --all
```

## API端点

- `GET /health` - 健康检查
- `GET /openapi.json`, `GET /v1/openapi.json`, `GET /v2/openapi.json` - OpenAPI规范
- `GET /api/requests` - 获取所有记录的请求
- `GET /api/requests/:id` - 获取特定ID的请求
- `ANY /v1/*path`, `ANY /v2/*path` - 捕获所有API请求并记录
- `GET /ui/` - Web界面查看请求历史

## Web UI界面

访问 http://localhost:8080/ui/ 查看请求历史的Web界面。默认情况下，会自动生成UI访问凭证。

## 数据存储

请求数据以JSON格式存储在指定的数据目录中，每个请求一个文件，文件名为请求ID。

## 许可证

MIT

## 作者

LLM-SEC团队 