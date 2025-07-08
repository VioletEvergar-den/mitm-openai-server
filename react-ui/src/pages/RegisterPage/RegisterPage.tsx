import React, { useState } from 'react';
import { useNavigate, Link as RouterLink } from 'react-router-dom';
import { 
  Button, 
  Form, 
  Input, 
  Card, 
  Typography, 
  Alert, 
  Layout,
  Row,
  Col
} from 'antd';
import { 
  UserOutlined, 
  LockOutlined, 
  EyeTwoTone, 
  EyeInvisibleOutlined, 
  UserAddOutlined,
  LoginOutlined
} from '@ant-design/icons';
import Footer from '../../components/Footer';

const { Title, Text } = Typography;

const RegisterPage: React.FC = () => {
  const [form] = Form.useForm();
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);
  
  const navigate = useNavigate();

  const handleSubmit = async (values: { username: string; password: string; confirmPassword: string }) => {
    const { username, password, confirmPassword } = values;
    
    // 确认密码验证
    if (password !== confirmPassword) {
      setError('两次输入的密码不一致');
      return;
    }
    
    setLoading(true);
    setError('');
    setSuccess('');
    
    try {
      const response = await fetch('/ui/api/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ 
          username, 
          password
        })
      });
      
      const data = await response.json();
      
      console.log('注册响应数据:', data);
      console.log('注册响应状态:', response.status);
      
      if (response.ok && data.code === 0) {
        // 注册成功
        setSuccess('注册成功！3秒后自动跳转到登录页面...');
        
        // 3秒后跳转到登录页面
        setTimeout(() => {
          navigate('/login');
        }, 3000);
      } else {
        setError(data.msg || '注册失败，请稍后重试');
      }
    } catch (err) {
      setError('注册失败，请稍后重试');
      console.error('注册错误:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Layout style={{ minHeight: '100vh', backgroundColor: '#f0f2f5' }}>
      <Row justify="center" align="middle" style={{ flex: 1, padding: '24px 0' }}>
        <Col xs={22} sm={16} md={12} lg={8} xl={6}>
          <Title level={2} style={{ textAlign: 'center', marginBottom: 32 }}>
            创建新账号
          </Title>
          <Card style={{ boxShadow: '0 4px 12px rgba(0, 0, 0, 0.15)' }}>
            {error && (
              <Alert
                message={error}
                type="error"
                showIcon
                style={{ marginBottom: 24 }}
              />
            )}
            
            {success && (
              <Alert
                message={success}
                type="success"
                showIcon
                style={{ marginBottom: 24 }}
              />
            )}
            
            <Form
              form={form}
              name="register"
              onFinish={handleSubmit}
              layout="vertical"
              size="large"
            >
              <Form.Item
                name="username"
                rules={[
                  { required: true, message: '请输入用户名' },
                  { min: 3, message: '用户名至少需要3个字符' },
                  { pattern: /^[a-zA-Z0-9_-]+$/, message: '用户名只能包含字母、数字、下划线和横线' }
                ]}
              >
                <Input 
                  prefix={<UserOutlined />} 
                  placeholder="用户名（至少3个字符）" 
                  autoComplete="username"
                  disabled={loading}
                />
              </Form.Item>
              <Form.Item
                name="password"
                rules={[
                  { required: true, message: '请输入密码' },
                  { min: 6, message: '密码至少需要6个字符' }
                ]}
              >
                <Input.Password
                  prefix={<LockOutlined />}
                  placeholder="密码（至少6个字符）"
                  autoComplete="new-password"
                  disabled={loading}
                  iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                />
              </Form.Item>
              <Form.Item
                name="confirmPassword"
                dependencies={['password']}
                rules={[
                  { required: true, message: '请确认密码' },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (!value || getFieldValue('password') === value) {
                        return Promise.resolve();
                      }
                      return Promise.reject(new Error('两次输入的密码不一致'));
                    },
                  }),
                ]}
              >
                <Input.Password
                  prefix={<LockOutlined />}
                  placeholder="确认密码"
                  autoComplete="new-password"
                  disabled={loading}
                  iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                />
              </Form.Item>
              <Form.Item>
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={loading}
                  block
                  icon={<UserAddOutlined />}
                >
                  {loading ? '注册中...' : '创建账号'}
                </Button>
              </Form.Item>
              <Text style={{ textAlign: 'center', display: 'block' }}>
                已有账号？ <RouterLink to="/login"><LoginOutlined /> 立即登录</RouterLink>
              </Text>
            </Form>
          </Card>
        </Col>
      </Row>
      <Footer />
    </Layout>
  );
};

export default RegisterPage; 