import React, { useState, useEffect } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { Menu, Typography, Dropdown, Avatar, Button, Space } from 'antd';
import { 
  HomeOutlined, 
  SettingOutlined, 
  BookOutlined,
  LogoutOutlined,
  GithubOutlined,
  UserOutlined
} from '@ant-design/icons';
import { useAuth } from '../../contexts/AuthContext';
import './Navbar.css';

const { Title } = Typography;

const Navbar: React.FC = () => {
  const { isAuthenticated, logout } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();
  const [current, setCurrent] = useState('');

  // 根据当前路径更新选中的菜单项
  useEffect(() => {
    const pathname = location.pathname;
    if (pathname === '/') {
      setCurrent('home');
    } else if (pathname.startsWith('/guide')) {
      setCurrent('guide');
    } else if (pathname.startsWith('/settings')) {
      setCurrent('settings');
    } else if (pathname.startsWith('/requests/')) {
      setCurrent('home'); // 查看请求详情时，仍然高亮"首页"
    }
  }, [location]);

  // 如果未认证，不显示导航栏
  if (!isAuthenticated) {
    return null;
  }

  // 处理菜单点击
  const handleMenuClick = (e: {key: string}) => {
    setCurrent(e.key);
    if (e.key === 'logout') {
      logout();
    }
  };

  // 菜单项定义
  const menuItems = [
    {
      key: 'home',
      icon: <HomeOutlined />,
      label: '请求列表',
      onClick: () => navigate('/')
    },
    {
      key: 'guide',
      icon: <BookOutlined />,
      label: 'OpenAI API配置教程',
      onClick: () => navigate('/guide')
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '系统设置',
      onClick: () => navigate('/settings')
    }
  ];

  // 用户下拉菜单
  const userMenuItems = [
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: () => logout()
    }
  ];

  const renderUserSection = () => {
    if (isAuthenticated) {
      return (
        <Space size={16}>
          <a 
            href="https://github.com/llm-sec/mitm-openai-server" 
            target="_blank" 
            rel="noopener noreferrer"
            className="navbar-link github-link"
          >
            <span className="navbar-icon"><GithubOutlined /></span>
            <span>GitHub</span>
          </a>
          <Dropdown 
            menu={{ items: userMenuItems }} 
            placement="bottomRight"
            trigger={['click']}
          >
            <Avatar
              className="user-avatar"
              icon={<UserOutlined />}
              style={{ 
                cursor: 'pointer',
                backgroundColor: '#1890ff'
              }}
            />
          </Dropdown>
        </Space>
      );
    }

    return (
      <Space size={8}>
        <Button type="text" onClick={() => navigate('/login')} className="auth-button">
          登录
        </Button>
        <Button type="primary" onClick={() => navigate('/register')} className="auth-button">
          注册
        </Button>
        <a 
          href="https://github.com/llm-sec/mitm-openai-server" 
          target="_blank" 
          rel="noopener noreferrer"
          className="navbar-link github-link"
        >
          <span className="navbar-icon"><GithubOutlined /></span>
          <span>GitHub</span>
        </a>
      </Space>
    );
  };

  return (
    <nav className="navbar">
      <div className="navbar-content">
        <Link to="/" className="navbar-brand">
          <Title level={4}>中间人OpenAI API服务器</Title>
        </Link>
        <div className="navbar-right">
          <ul className="navbar-menu">
            {menuItems.map(item => (
              <li key={item.key} className="navbar-item">
                <a
                  className={`navbar-link ${current === item.key ? 'active' : ''}`}
                  onClick={item.onClick}
                >
                  <span className="navbar-icon">{item.icon}</span>
                  <span>{item.label}</span>
                </a>
              </li>
            ))}
          </ul>
          {renderUserSection()}
        </div>
      </div>
    </nav>
  );
};

export default Navbar; 