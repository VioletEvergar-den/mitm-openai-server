# 假的OpenAPI服务器

这是一个简单的假OpenAPI服务器，用于接收和记录传入的请求数据。服务器符合OpenAPI标准，但总是返回固定的成功响应，并将请求详细信息记录到本地文件数据库中。

## 功能特点

- 完全符合OpenAPI 3.0规范
- 捕获并记录所有API请求的详细信息
- 提供REST API接口查询已记录的请求
- 可配置的数据存储目录
- 模块化设计，易于扩展

## 安装

确保你已安装Go（版本1.20或更高）。

```bash
# 克隆仓库
git clone https://github.com/llm-sec/fake-openapi-server.git
cd fake-openapi-server

# 安装依赖
go mod tidy

# 构建
go build -o fake-openapi-server ./cmd/server
```

## 运行服务器

```bash
# 使用默认设置运行
./fake-openapi-server

# 指定端口和数据目录
./fake-openapi-server -port 9000 -data /path/to/data/directory
```

## API端点

- `GET /health` - 健康检查
- `GET /openapi.json`, `GET /v1/openapi.json`, `GET /v2/openapi.json` - OpenAPI规范
- `GET /api/requests` - 获取所有记录的请求
- `GET /api/requests/:id` - 获取特定ID的请求
- `ANY /v1/*path`, `ANY /v2/*path` - 捕获所有API请求并记录

## 示例用法

向服务器发送请求:

```bash
# 发送POST请求
curl -X POST http://localhost:8080/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name": "张三", "email": "zhangsan@example.com"}'

# 发送GET请求
curl http://localhost:8080/v1/users/123

# 查看已记录的请求
curl http://localhost:8080/api/requests
```

## 查看OpenAPI规范

访问 http://localhost:8080/openapi.json 查看API规范。

## 数据存储

请求数据以JSON格式存储在指定的数据目录中，每个请求一个文件，文件名为请求ID。

## 许可证

MIT

## 作者

LLM-SEC团队 