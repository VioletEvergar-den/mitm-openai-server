import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import './styles/App.css';

// 导入简化的登录页面
const LoginPage = () => {
  return (
    <div className="login-container">
      <div className="login-card">
        <h2 className="login-title">登录</h2>
        <form>
          <div className="form-group">
            <label htmlFor="username" className="form-label">用户名</label>
            <input
              type="text"
              id="username"
              className="form-control"
            />
          </div>
          
          <div className="form-group">
            <label htmlFor="password" className="form-label">密码</label>
            <input
              type="password"
              id="password"
              className="form-control"
            />
          </div>
          
          <button 
            type="submit" 
            className="btn btn-primary btn-block"
          >
            登录
          </button>
        </form>
      </div>
    </div>
  );
};

// 简单的主页面
const HomePage = () => {
  return (
    <div className="container content">
      <div className="card">
        <h2 className="card-title">欢迎使用中间人OpenAI API服务器</h2>
        <p>这是主页面。后续将添加更多功能。</p>
      </div>
    </div>
  );
};

const FinalApp: React.FC = () => {
  return (
    <Router>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/" element={<HomePage />} />
      </Routes>
    </Router>
  );
};

export default FinalApp; 