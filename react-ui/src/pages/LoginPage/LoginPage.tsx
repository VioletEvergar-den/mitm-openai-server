import React, { useState, useEffect } from 'react';
import { useNavigate, Link as RouterLink } from 'react-router-dom';
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
  Layout,
  Row,
  Col,
  Space
} from 'antd';
import { 
  UserOutlined, 
  LockOutlined, 
  EyeTwoTone, 
  EyeInvisibleOutlined, 
  LoginOutlined,
  GithubOutlined,
  SafetyOutlined
} from '@ant-design/icons';
import Footer from '../../components/Footer';

const { Title, Paragraph, Link, Text } = Typography;

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
      
      // 添加控制台调试输出
      console.log('登录响应数据:', data);
      console.log('登录响应状态:', response.status);
      console.log('data.data 完整内容:', data.data);
      console.log('data.data 类型:', typeof data.data);
      console.log('token是否存在:', !!data.data?.token);
      if (data.data?.token) {
        console.log('token长度:', data.data.token.length);
        console.log('token前10个字符:', data.data.token.substring(0, 10));
      } else {
        console.log('token不存在，data.data?.token 值:', data.data?.token);
      }
      
      if (response.ok && data.code === 0) {
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
        
        // 保存后端返回的token，设置为一年有效期
        if (data.data?.token) {
          utils.saveAuth(data.data.token, 365);
          login(username);
          console.log('已保存token到localStorage');
          console.log('保存的token:', localStorage.getItem('auth_token'));
          navigate('/');
        } else {
          setError('服务器没有返回有效的认证令牌');
        }
      } else {
        setError(data.msg || '用户名或密码不正确');
      }
    } catch (err) {
      setError('登录失败，请稍后重试');
      console.error('登录错误:', err);
    } finally {
      setLoading(false);
    }
  };
  
  return (
    <Layout style={{ minHeight: '100vh', backgroundColor: '#f0f2f5' }}>
      <Row justify="center" align="middle" style={{ minHeight: '100vh' }}>
        <Col xs={22} sm={20} md={16} lg={12} xl={10}>
          <Card style={{ boxShadow: '0 4px 12px rgba(0, 0, 0, 0.15)' }}>
            <Row>
              <Col xs={0} sm={0} md={10} style={{ 
                backgroundColor: '#001529', 
                color: 'white', 
                padding: '48px 32px',
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'center'
              }}>
                <Typography>
                  <Title level={2} style={{ color: 'white' }}>MITM OpenAI Server</Title>
                  <Paragraph style={{ color: 'rgba(255, 255, 255, 0.85)' }}>
                    安全监控与分析OpenAI API请求的强大工具
                  </Paragraph>
                  <Paragraph style={{ color: 'rgba(255, 255, 255, 0.65)' }}>
                    实时拦截、记录和检查AI交互的完整内容
                  </Paragraph>
                  <Space direction="vertical" style={{ marginTop: 24 }}>
                    <Link 
                      href="https://github.com/llm-sec/mitm-openai-server" 
                      target="_blank" 
                      rel="noopener noreferrer"
                      style={{ color: 'white' }}
                    >
                      <GithubOutlined /> GitHub 项目仓库
                    </Link>
                    <Text style={{ color: 'rgba(255, 255, 255, 0.65)' }}>
                      <SafetyOutlined /> 安全可靠的专业工具
                    </Text>
                  </Space>
                </Typography>
              </Col>
              <Col xs={24} sm={24} md={14} style={{ padding: '48px 32px' }}>
                <Title level={3} style={{ textAlign: 'center' }}>系统登录</Title>
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
                    rules={[{ required: true, message: '请输入用户名' }]}
                  >
                    <Input 
                      prefix={<UserOutlined />} 
                      placeholder="用户名" 
                      autoComplete="username"
                      disabled={loading}
                    />
                  </Form.Item>
                  <Form.Item
                    name="password"
                    rules={[{ required: true, message: '请输入密码' }]}
                  >
                    <Input.Password
                      prefix={<LockOutlined />}
                      placeholder="密码"
                      autoComplete="current-password"
                      disabled={loading}
                      iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                    />
                  </Form.Item>
                  <Form.Item>
                    <Form.Item name="remember" valuePropName="checked" noStyle>
                      <Checkbox disabled={loading}>记住我</Checkbox>
                    </Form.Item>
                    <RouterLink to="/forgot-password" style={{ float: 'right' }}>
                      忘记密码
                    </RouterLink>
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
                  <Text style={{ textAlign: 'center', display: 'block' }}>
                    还没有账号？ <RouterLink to="/register">立即注册</RouterLink>
                  </Text>
                </Form>
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>
      <Footer />
    </Layout>
  );
};

export default LoginPage; 