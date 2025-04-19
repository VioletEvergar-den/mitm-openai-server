import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { utils } from '../../services/api';
import { 
  Button, 
  Checkbox, 
  Form, 
  Input, 
  Card, 
  Typography, 
  Alert, 
  Space, 
  Layout
} from 'antd';
import { 
  UserOutlined, 
  LockOutlined, 
  EyeTwoTone, 
  EyeInvisibleOutlined, 
  LoginOutlined,
  GithubOutlined
} from '@ant-design/icons';
import Footer from '../../components/Footer';
import './LoginPage.css';

const { Title, Paragraph, Link } = Typography;
const { Content } = Layout;

const LoginPage: React.FC = () => {
  const [form] = Form.useForm();
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [rememberMe, setRememberMe] = useState(true);
  
  const { login } = useAuth();
  const navigate = useNavigate();
  
  // 从本地存储加载之前保存的用户名和密码
  useEffect(() => {
    const savedUsername = localStorage.getItem('mitm_remembered_username');
    const savedPassword = localStorage.getItem('mitm_remembered_password');
    const savedRememberMe = localStorage.getItem('mitm_remember_me');
    
    if (savedRememberMe === 'true') {
      setRememberMe(true);
      
      if (savedUsername && savedPassword) {
        form.setFieldsValue({
          username: savedUsername,
          password: savedPassword,
          remember: true
        });
      }
    } else {
      setRememberMe(false);
    }
  }, [form]);

  const handleSubmit = async (values: { username: string; password: string; remember: boolean }) => {
    const { username, password, remember } = values;
    
    setLoading(true);
    setError('');
    
    try {
      // 使用专门的登录API接口
      const response = await fetch('/ui/api/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ username, password })
      });
      
      const data = await response.json();
      
      if (response.ok && data.status === 'success') {
        // 登录成功
        
        // 保存用户的"记住我"选择
        localStorage.setItem('mitm_remember_me', remember.toString());
        
        // 如果用户选择了"记住我"，保存凭据
        if (remember) {
          localStorage.setItem('mitm_remembered_username', username);
          localStorage.setItem('mitm_remembered_password', password);
        } else {
          // 清除之前可能存在的凭据
          localStorage.removeItem('mitm_remembered_username');
          localStorage.removeItem('mitm_remembered_password');
        }
        
        // 保存登录状态和凭据，设置为一年有效期
        utils.saveAuth(username, password, 365);
        login(username, password);
        navigate('/');
      } else {
        setError(data.message || '用户名或密码不正确');
      }
    } catch (err) {
      setError('登录失败，请稍后重试');
      console.error('登录错误:', err);
    } finally {
      setLoading(false);
    }
  };
  
  return (
    <Layout className="login-layout">
      <Content className="login-content">
    <div className="login-container">
      <div className="login-banner">
            <Typography>
              <Title level={2}>MITM OpenAI Server</Title>
              <Paragraph>监控、拦截和分析AI API请求的工具</Paragraph>
              <Space className="banner-links">
                <Link 
                  href="https://github.com/llm-sec/mitm-openai-server" 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="github-link"
                >
                  <GithubOutlined /> GitHub 项目仓库
                </Link>
              </Space>
            </Typography>
          </div>
          
          <Card 
            className="login-card" 
            bordered={false}
            title={<Title level={3} style={{ textAlign: 'center', margin: 0 }}>登录</Title>}
          >
            {error && (
              <Alert
                message={error}
                type="error"
                showIcon
                style={{ marginBottom: 24 }}
              />
            )}
            
            <Form
              form={form}
              name="login"
              initialValues={{ remember: rememberMe }}
              onFinish={handleSubmit}
              layout="vertical"
              size="large"
            >
              <Form.Item
                name="username"
                label="用户名"
                rules={[{ required: true, message: '请输入用户名' }]}
              >
                <Input 
                  prefix={<UserOutlined />} 
                  placeholder="请输入用户名" 
                  autoComplete="username"
                disabled={loading}
              />
              </Form.Item>
              
              <Form.Item
                name="password"
                label="密码"
                rules={[{ required: true, message: '请输入密码' }]}
              >
                <Input.Password
                  prefix={<LockOutlined />}
                  placeholder="请输入密码"
                  autoComplete="current-password"
            disabled={loading}
                  iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                />
              </Form.Item>
              
              <Form.Item name="remember" valuePropName="checked">
                <Checkbox disabled={loading}>记住登录信息</Checkbox>
              </Form.Item>
              
              <Form.Item>
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={loading}
                  block
                  icon={<LoginOutlined />}
                >
                  {loading ? '登录中...' : '登录'}
                </Button>
              </Form.Item>
            </Form>
          </Card>
      </div>
      </Content>
      <Footer />
    </Layout>
  );
};

export default LoginPage; 