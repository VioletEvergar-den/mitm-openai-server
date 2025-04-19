import React, { useState } from 'react';
import Layout from '../../components/Layout/Layout';
import './GuidePage.css';

const GuidePage: React.FC = () => {
  const [showToken, setShowToken] = useState(false);
  // 使用正确的后端服务器地址和端口
  const backendUrl = window.location.hostname === 'localhost' ? 
    'http://localhost:8080' : window.location.origin;
  const apiUrl = '/v1';
  const fullApiUrl = backendUrl + apiUrl; // 完整的API地址，包括/v1路径
  const token = 'Bearer mt-mitm-server-token'; // 实际应该从后端API获取

  // 对话状态
  const [messages, setMessages] = useState<Array<{role: string, content: string}>>([]);
  const [inputMessage, setInputMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  // 发送消息到MITM OpenAI Server
  const sendMessage = async () => {
    if (!inputMessage.trim()) return;
    
    // 添加用户消息到历史
    const userMessage = { role: 'user', content: inputMessage };
    const updatedMessages = [...messages, userMessage];
    setMessages(updatedMessages);
    setInputMessage('');
    setIsLoading(true);
    
    try {
      // 构建请求数据，完全符合OpenAI API规范
      const requestData = {
        model: 'gpt-3.5-turbo',
        messages: updatedMessages,
        temperature: 0.7
      };
      
      // 发送请求到MITM OpenAI Server
      // 这里使用MITM UI API的chat接口，而不是直接请求OpenAI API
      const response = await fetch(`/ui/api/chat`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token
        },
        credentials: 'include',
        body: JSON.stringify(requestData)
      });
      
      if (!response.ok) {
        throw new Error(`请求失败: ${response.status}`);
      }
      
      const data = await response.json();
      
      // 添加助手响应到历史，处理标准OpenAI API响应格式
      if (data.choices && data.choices.length > 0) {
        const assistantMessage = data.choices[0].message;
        setMessages([...updatedMessages, assistantMessage]);
      }
    } catch (error) {
      console.error('请求MITM OpenAI API失败:', error);
      // 添加错误消息
      setMessages([
        ...updatedMessages, 
        { 
          role: 'assistant', 
          content: `发生错误: ${error instanceof Error ? error.message : '未知错误'}`
        }
      ]);
    } finally {
      setIsLoading(false);
    }
  };

  // 处理按回车键发送
  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  return (
    <Layout title="OpenAI API配置教程">
      <div className="simple-guide-container">
        <div className="guide-card">
          <h2>OpenAI API配置指南</h2>
          <p className="guide-description">
            要使用此中间人服务器，您需要将OpenAI客户端配置为使用以下API地址和认证Token。
            这样您的所有API请求将通过此服务器，并被记录下来供后续分析。
          </p>

          <div className="config-section">
            <h3>1. API地址</h3>
            <div className="config-item">
              <div className="input-with-copy">
                <input 
                  type="text" 
                  value={fullApiUrl} 
                  readOnly 
                  className="config-input"
                />
                <button 
                  className="copy-button"
                  onClick={() => copyToClipboard(fullApiUrl)}
                  title="复制到剪贴板"
                >
                  复制
                </button>
              </div>
              <p className="config-hint">在您的OpenAI客户端中，将baseURL设置为上述地址</p>
            </div>
          </div>

          <div className="config-section">
            <h3>2. 认证Token</h3>
            <div className="config-item">
              <div className="input-with-copy">
                <input 
                  type={showToken ? "text" : "password"}
                  value={token} 
                  readOnly 
                  className="config-input"
                />
                <button 
                  className="button-secondary"
                  onClick={() => setShowToken(!showToken)}
                >
                  {showToken ? "隐藏" : "显示"}
                </button>
                <button 
                  className="copy-button"
                  onClick={() => copyToClipboard(token)}
                  title="复制到剪贴板"
                >
                  复制
                </button>
              </div>
              <p className="config-hint">在您的OpenAI客户端中，使用此Token作为API密钥</p>
            </div>
          </div>

          {/* 添加测试提示信息 */}
          <div className="config-section">
            <h3>3. 验证API连接</h3>
            <div className="config-item">
              <p className="config-hint">您可以使用以下命令测试API连接：</p>
              <div className="code-block">
                <code>
                  curl -X GET {fullApiUrl}/models -H "Authorization: {token}"
                </code>
                <button 
                  className="copy-button"
                  onClick={() => copyToClipboard(`curl -X GET ${fullApiUrl}/models -H "Authorization: ${token}"`)}
                  title="复制到剪贴板"
                >
                  复制
                </button>
              </div>
              <p className="config-hint">或使用下方的测试对话框直接测试聊天功能</p>
            </div>
          </div>

          {/* 测试对话框 */}
          <div className="chat-test-section">
            <h3>测试对话</h3>
            <p className="chat-description">
              在下方与MITM OpenAI Server进行对话测试，验证服务器是否正常工作。
            </p>
            
            <div className="chat-container">
              <div className="messages-container">
                {messages.length === 0 ? (
                  <div className="empty-chat">
                    <p>开始与AI助手对话吧</p>
                  </div>
                ) : (
                  messages.map((msg, index) => (
                    <div 
                      key={index} 
                      className={`message ${msg.role === 'user' ? 'user-message' : 'assistant-message'}`}
                    >
                      <div className="message-role">{msg.role === 'user' ? '用户' : 'AI助手'}</div>
                      <div className="message-content">{msg.content}</div>
                    </div>
                  ))
                )}
                {isLoading && (
                  <div className="message assistant-message loading">
                    <div className="message-role">AI助手</div>
                    <div className="message-content">
                      <div className="typing-indicator">
                        <span></span>
                        <span></span>
                        <span></span>
                      </div>
                    </div>
                  </div>
                )}
        </div>
        
              <div className="chat-input-container">
                <textarea 
                  className="chat-input"
                  value={inputMessage}
                  onChange={(e) => setInputMessage(e.target.value)}
                  onKeyPress={handleKeyPress}
                  placeholder="输入消息..."
                  disabled={isLoading}
                  rows={2}
                />
                <button 
                  className="send-button"
                  onClick={sendMessage}
                  disabled={isLoading || !inputMessage.trim()}
                >
                  {isLoading ? '发送中...' : '发送'}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
};

export default GuidePage; 