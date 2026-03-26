import React, { useState, useEffect } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { Typography, Dropdown, Avatar, Space, Switch, Tooltip } from 'antd';
import { 
  HomeOutlined, 
  SettingOutlined, 
  BookOutlined,
  LogoutOutlined,
  GithubOutlined,
  UserOutlined,
  FileTextOutlined,
  SunOutlined,
  MoonOutlined
} from '@ant-design/icons';
import { useAuth } from '../../contexts/AuthContext';
import { useTheme } from '../../contexts/ThemeContext';
import './Navbar.css';

const { Title } = Typography;

const Navbar: React.FC = () => {
  const { isAuthenticated, logout } = useAuth();
  const { mode, toggleTheme } = useTheme();
  const location = useLocation();
  const navigate = useNavigate();
  const [current, setCurrent] = useState('');

  useEffect(() => {
    const pathname = location.pathname;
    if (pathname === '/') {
      setCurrent('home');
    } else if (pathname.startsWith('/guide')) {
      setCurrent('guide');
    } else if (pathname.startsWith('/logs')) {
      setCurrent('logs');
    } else if (pathname.startsWith('/settings')) {
      setCurrent('settings');
    } else if (pathname.startsWith('/requests/')) {
      setCurrent('home');
    }
  }, [location]);

  if (!isAuthenticated) {
    return null;
  }

  const handleMenuClick = (e: {key: string}) => {
    setCurrent(e.key);
    if (e.key === 'logout') {
      logout();
    }
  };

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
      label: '配置教程',
      onClick: () => navigate('/guide')
    },
    {
      key: 'logs',
      icon: <FileTextOutlined />,
      label: '实时日志',
      onClick: () => navigate('/logs')
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '系统设置',
      onClick: () => navigate('/settings')
    }
  ];

  const userMenuItems = [
    {
      key: 'theme',
      label: (
        <Space>
          <span>主题模式</span>
          <Switch
            checked={mode === 'dark'}
            onChange={toggleTheme}
            checkedChildren={<MoonOutlined />}
            unCheckedChildren={<SunOutlined />}
            size="small"
          />
        </Space>
      ),
      onClick: (e: any) => {
        if (e.domEvent) {
          e.domEvent.stopPropagation();
        }
      }
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: () => logout()
    }
  ];

  const renderGithubLink = () => (
    <a 
      href="https://github.com/llm-sec/mitm-openai-server" 
      target="_blank" 
      rel="noopener noreferrer"
      className="github-link"
    >
      <span className="navbar-icon"><GithubOutlined /></span>
      <span>GitHub</span>
    </a>
  );

  const renderUserAvatar = () => (
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
          backgroundColor: '#1890ff',
          marginRight: 0
        }}
      />
    </Dropdown>
  );

  return (
    <nav className="navbar">
      <div className="navbar-container">
        <div className="navbar-left">
          <Link to="/" className="navbar-brand">
            <Title level={4}>OpenAI API 代理</Title>
          </Link>
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
        </div>
        
        <div className="navbar-right">
          <Tooltip title={mode === 'dark' ? '切换到亮色模式' : '切换到暗色模式'}>
            <Switch
              checked={mode === 'dark'}
              onChange={toggleTheme}
              checkedChildren={<MoonOutlined />}
              unCheckedChildren={<SunOutlined />}
              className="theme-switch"
            />
          </Tooltip>
          <div className="github-container">
            {renderGithubLink()}
          </div>
          <div className="avatar-container">
            {renderUserAvatar()}
          </div>
        </div>
      </div>
    </nav>
  );
};

export default Navbar;
