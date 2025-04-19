import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import './Navbar.css';

const Navbar: React.FC = () => {
  const { isAuthenticated, logout } = useAuth();
  const location = useLocation();

  // 如果未认证，不显示导航栏
  if (!isAuthenticated) {
    return null;
  }

  return (
    <nav className="navbar">
      <div className="container navbar-container">
        <Link to="/" className="navbar-brand">中间人OpenAI API服务器</Link>
        <ul className="navbar-menu">
          <li className="navbar-item">
            <Link 
              to="/" 
              className={`navbar-link ${location.pathname === '/' ? 'active' : ''}`}
            >
              请求列表
            </Link>
          </li>
          <li className="navbar-item">
            <Link 
              to="/settings" 
              className={`navbar-link ${location.pathname === '/settings' ? 'active' : ''}`}
            >
              系统设置
            </Link>
          </li>
          <li className="navbar-item">
            <Link 
              to="/guide" 
              className={`navbar-link ${location.pathname === '/guide' ? 'active' : ''}`}
            >
              配置指南
            </Link>
          </li>
          <li className="navbar-item">
            <button 
              onClick={logout} 
              className="navbar-link" 
              style={{ background: 'none', border: 'none', cursor: 'pointer' }}
            >
              退出登录
            </button>
          </li>
        </ul>
      </div>
    </nav>
  );
};

export default Navbar; 