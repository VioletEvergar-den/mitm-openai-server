import React, { useState, useEffect } from 'react';
import {
  Input,
  Button,
  Card,
  Typography,
  Space,
  Row,
  Col,
  List,
  Avatar,
  Spin,
  Alert,
  Select,
  Divider,
  message
} from 'antd';
import { 
  CopyOutlined, 
  UserOutlined, 
  ApiOutlined, 
  KeyOutlined,
  CheckCircleOutlined,
  SendOutlined,
  BookOutlined
} from '@ant-design/icons';
import Layout from '../../components/Layout/Layout';
import { useNotification } from '../../components/Notification';

const { Title, Paragraph, Text } = Typography;

const AVAILABLE_MODELS = [
  { value: 'gpt-5.4', label: 'GPT-5.4' },
  { value: 'gpt-5.4-pro', label: 'GPT-5.4 Pro' },
  { value: 'gpt-5.2', label: 'GPT-5.2' },
  { value: 'gpt-5.2-pro', label: 'GPT-5.2 Pro' },
  { value: 'gpt-4', label: 'GPT-4' },
  { value: 'gpt-4-32k', label: 'GPT-4 32K' },
  { value: 'gpt-3.5-turbo', label: 'GPT-3.5 Turbo' },
  { value: 'gpt-3.5-turbo-16k', label: 'GPT-3.5 Turbo 16K' },
  { value: 'glm-5', label: 'GLM-5' },
  { value: 'glm-5-turbo', label: 'GLM-5 Turbo' },
  { value: 'kimi-k2.5', label: 'Kimi K2.5' },
  { value: 'qwen3-max', label: 'Qwen3-Max' },
  { value: 'qwen3.5-plus', label: 'Qwen3.5-Plus' },
  { value: 'qwen3.5-flash', label: 'Qwen3.5-Flash' },
  { value: 'qwen3.5-27b', label: 'Qwen3.5-27B' },
  { value: 'minimax-m2.7', label: 'MiniMax M2.7' },
  { value: 'minimax-m2.5', label: 'MiniMax M2.5' },
  { value: 'minimax-m2.5-lightning', label: 'MiniMax M2.5 Lightning' },
];

const GuidePage: React.FC = () => {
  const [backendUrl, setBackendUrl] = useState('');
  const [fullApiUrl, setFullApiUrl] = useState('');
  const [token, setToken] = useState('');
  const [messages, setMessages] = useState<Array<{role: string, content: string}>>([]);
  const [inputMessage, setInputMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [selectedModel, setSelectedModel] = useState('gpt-5.4');
  const { addNotification } = useNotification();

  useEffect(() => {
    const origin = window.location.origin;
    setBackendUrl(origin);
    setFullApiUrl(`${origin}/v1`);

    const fetchToken = async () => {
      try {
        const response = await fetch(`${origin}/ui/api/token`);
        const data = await response.json();
        if (data.code === 0 && data.data?.token) {
          setToken(`Bearer ${data.data.token}`);
        }
      } catch (error) {
        addNotification('获取认证Token失败', 'danger');
      }
    };
    fetchToken();
  }, [addNotification]);

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      message.success('已复制到剪贴板');
    } catch (error) {
      message.error('复制失败');
    }
  };

  const sendMessage = async () => {
    if (!inputMessage.trim()) return;
    
    const userMessage = { role: 'user', content: inputMessage };
    const updatedMessages = [...messages, userMessage];
    setMessages(updatedMessages);
    setInputMessage('');
    setIsLoading(true);
    
    try {
      const requestData = { model: selectedModel, messages: updatedMessages, temperature: 0.7 };
      const response = await fetch(`${backendUrl}/v1/chat/completions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': token },
        body: JSON.stringify(requestData)
      });
      
      if (!response.ok) throw new Error(`请求失败: ${response.status}`);
      
      const data = await response.json();
      if (data.choices?.[0]?.message) {
        setMessages([...updatedMessages, data.choices[0].message]);
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '未知错误';
      setMessages([...updatedMessages, { role: 'assistant', content: `发生错误: ${errorMessage}` }]);
    } finally {
      setIsLoading(false);
    }
  };

  const CodeBlock: React.FC<{text: string, language?: string}> = ({ text, language = 'bash' }) => (
    <div 
      style={{ 
        position: 'relative', 
        backgroundColor: '#282c34', 
        padding: '16px', 
        borderRadius: 8, 
        color: '#abb2bf',
        fontFamily: 'Consolas, Monaco, "Courier New", monospace',
        fontSize: 13,
        overflow: 'auto'
      }}
    >
      <code>{text}</code>
      <Button 
        icon={<CopyOutlined />} 
        onClick={() => copyToClipboard(text)} 
        style={{ position: 'absolute', top: 8, right: 8 }}
        size="small"
      />
    </div>
  );

  return (
    <Layout title="配置教程">
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Title level={3} style={{ margin: 0 }}>
          <BookOutlined style={{ marginRight: 8 }} />
          OpenAI API 配置指南
        </Title>
        
        <Alert
          message="快速开始"
          description="将您的OpenAI客户端配置为使用以下API地址和认证Token，所有API请求将通过此服务器并被记录下来供后续分析。"
          type="info"
          showIcon
        />
        
        <Row gutter={[16, 16]}>
          <Col xs={24} lg={12}>
            <Card 
              title={
                <Space>
                  <ApiOutlined />
                  <span>1. API 地址</span>
                </Space>
              }
              size="small"
            >
              <Input.Password 
                value={fullApiUrl} 
                readOnly 
                visibilityToggle={false}
                addonAfter={
                  <Button 
                    icon={<CopyOutlined />} 
                    onClick={() => copyToClipboard(fullApiUrl)}
                    size="small"
                  />
                }
              />
              <Text type="secondary" style={{ marginTop: 8, display: 'block' }}>
                在您的OpenAI客户端中，将 baseURL 设置为上述地址。
              </Text>
            </Card>
          </Col>
          
          <Col xs={24} lg={12}>
            <Card 
              title={
                <Space>
                  <KeyOutlined />
                  <span>2. 认证 Token</span>
                </Space>
              }
              size="small"
            >
              <Input.Password 
                value={token} 
                readOnly 
                visibilityToggle={false}
                addonAfter={
                  <Button 
                    icon={<CopyOutlined />} 
                    onClick={() => copyToClipboard(token)}
                    size="small"
                  />
                }
              />
              <Text type="secondary" style={{ marginTop: 8, display: 'block' }}>
                在您的OpenAI客户端中，使用此 Token 作为 API 密钥。
              </Text>
            </Card>
          </Col>
        </Row>

        <Card 
          title={
            <Space>
              <CheckCircleOutlined />
              <span>3. 验证 API 连接</span>
            </Space>
          }
          size="small"
        >
          <Paragraph>您可以使用以下命令测试API连接：</Paragraph>
          <CodeBlock text={`curl -X GET ${fullApiUrl}/models -H "Authorization: ${token}"`} />
        </Card>

        <Card 
          title={
            <Space>
              <SendOutlined />
              <span>4. 测试对话</span>
            </Space>
          }
          size="small"
        >
          <Space style={{ marginBottom: 16 }}>
            <Text>选择模型：</Text>
            <Select
              value={selectedModel}
              onChange={setSelectedModel}
              options={AVAILABLE_MODELS}
              style={{ width: 200 }}
              placeholder="选择模型"
            />
          </Space>
          
          <div 
            style={{ 
              minHeight: 200, 
              maxHeight: 400, 
              overflowY: 'auto', 
              marginBottom: 16, 
              border: '1px solid #f0f0f0', 
              borderRadius: 8,
              padding: 16 
            }}
          >
            {messages.length === 0 ? (
              <Text type="secondary">发送消息开始测试...</Text>
            ) : (
              <List
                dataSource={messages}
                renderItem={item => (
                  <List.Item style={{ border: 'none', padding: '8px 0' }}>
                    <List.Item.Meta
                      avatar={
                        <Avatar 
                          icon={item.role === 'user' ? <UserOutlined /> : undefined}
                          style={{ 
                            backgroundColor: item.role === 'user' ? '#1890ff' : '#52c41a'
                          }}
                        >
                          {item.role === 'assistant' ? 'AI' : undefined}
                        </Avatar>
                      }
                      title={item.role === 'user' ? '你' : 'AI助手'}
                      description={item.content}
                    />
                  </List.Item>
                )}
              />
            )}
            {isLoading && (
              <div style={{ textAlign: 'center', padding: 16 }}>
                <Spin />
              </div>
            )}
          </div>
          
          <Space.Compact style={{ width: '100%' }}>
            <Input.TextArea 
              rows={2}
              value={inputMessage}
              onChange={e => setInputMessage(e.target.value)}
              placeholder="输入消息..."
              disabled={isLoading}
              onPressEnter={e => { 
                if (!e.shiftKey) { 
                  e.preventDefault(); 
                  sendMessage(); 
                } 
              }}
              style={{ flex: 1 }}
            />
            <Button 
              type="primary" 
              icon={<SendOutlined />}
              onClick={sendMessage} 
              loading={isLoading}
            >
              发送
            </Button>
          </Space.Compact>
        </Card>

        <Divider />

        <Card title="Python 示例代码" size="small">
          <CodeBlock 
            text={`from openai import OpenAI

client = OpenAI(
    api_key="${token.replace('Bearer ', '')}",
    base_url="${fullApiUrl}"
)

response = client.chat.completions.create(
    model="${selectedModel}",
    messages=[
        {"role": "user", "content": "你好"}
    ]
)

print(response.choices[0].message.content)`}
            language="python"
          />
        </Card>
      </Space>
    </Layout>
  );
};

export default GuidePage;
