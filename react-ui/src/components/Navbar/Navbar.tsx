import React, { useState, useEffect } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { Menu, Layout, Typography, Button, Space } from 'antd';
import { 
  HomeOutlined, 
  SettingOutlined, 
  BookOutlined, 
  LogoutOutlined,
  GithubOutlined
} from '@ant-design/icons';
import { useAuth } from '../../contexts/AuthContext';
import './Navbar.css';

const { Header } = Layout;
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
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: () => logout()
    },
    {
      key: 'github',
      icon: <GithubOutlined />,
      label: <a 
              href="https://github.com/llm-sec/mitm-openai-server" 
              target="_blank" 
              rel="noopener noreferrer" 
      >
        GitHub
      </a>,
    }
  ];

  return (
    <div className="navbar-container">
      <div className="logo">
        <Link to="/">
          <Title level={4} style={{ margin: 0, color: 'white' }}>中间人OpenAI API服务器</Title>
        </Link>
      </div>
      <Menu 
        theme="dark"
        mode="horizontal"
        selectedKeys={[current]}
        items={menuItems}
        style={{ flex: 1, minWidth: 600, justifyContent: 'flex-end' }}
      />
    </div>
  );
};

export default Navbar; 