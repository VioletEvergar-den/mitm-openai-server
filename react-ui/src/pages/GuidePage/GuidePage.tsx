import React from 'react';
import Layout from '../../components/Layout/Layout';
import './GuidePage.css';

const GuidePage: React.FC = () => {
  return (
    <Layout title="配置指南">
      <div className="card">
        <h2 className="card-title">配置指南</h2>
        
        <div className="guide-section">
          <h3>服务器配置</h3>
          <p>
            使用此中间人OpenAI API服务器，您需要先进行服务器配置。在"系统设置"页面中，您可以配置代理模式和目标API服务器地址。
          </p>
          <p>
            启用代理模式后，服务器将转发API请求到目标API服务器，并记录请求和响应内容。
          </p>
        </div>
        
        <div className="guide-section">
          <h3>客户端配置</h3>
          <p>
            您需要修改您的OpenAI客户端应用程序的配置，使其将API请求发送到此中间人服务器，而不是直接发送到OpenAI官方API。
          </p>
          
          <h4>NodeJS示例</h4>
          <pre className="code-block">
{`// 使用OpenAI官方SDK
const { Configuration, OpenAIApi } = require("openai");

const configuration = new Configuration({
  apiKey: "your-api-key",
  // 修改基础URL指向中间人服务器
  basePath: "http://localhost:8080/v1",
});

const openai = new OpenAIApi(configuration);

async function main() {
  const response = await openai.createChatCompletion({
    model: "gpt-3.5-turbo",
    messages: [
      { role: "system", content: "你是一个有帮助的助手。" },
      { role: "user", content: "你好，请问你是谁？" }
    ],
  });
  
  console.log(response.data);
}

main();`}
          </pre>
          
          <h4>Python示例</h4>
          <pre className="code-block">
{`# 使用OpenAI官方SDK
import openai

# 配置API密钥
openai.api_key = "your-api-key"
# 修改基础URL指向中间人服务器
openai.api_base = "http://localhost:8080/v1"

# 发送请求
response = openai.ChatCompletion.create(
    model="gpt-3.5-turbo",
    messages=[
        {"role": "system", "content": "你是一个有帮助的助手。"},
        {"role": "user", "content": "你好，请问你是谁？"}
    ]
)

print(response)`}
          </pre>
        </div>
        
        <div className="guide-section">
          <h3>查看和分析请求</h3>
          <p>
            所有经过中间人服务器的API请求都会被记录。您可以在"请求列表"页面中查看所有请求，并点击"查看详情"按钮查看请求的详细信息。
          </p>
          <p>
            请求详情页面显示了请求的完整信息，包括请求头、请求参数和请求体。
          </p>
          <p>
            您可以使用这些信息来了解API的使用情况、调试应用程序问题，或者学习OpenAI API的用法。
          </p>
        </div>
        
        <div className="guide-section">
          <h3>数据管理</h3>
          <p>
            在"系统设置"页面的"存储管理"部分，您可以导出所有请求数据或者清空请求记录。
          </p>
          <p>
            导出的数据格式为JSONL（每行一个JSON对象），您可以使用文本编辑器或者数据分析工具打开和处理。
          </p>
        </div>
      </div>
    </Layout>
  );
};

export default GuidePage; 