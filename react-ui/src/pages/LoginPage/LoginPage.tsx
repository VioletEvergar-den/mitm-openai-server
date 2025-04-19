import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { utils } from '../../services/api';
import './LoginPage.css';

const LoginPage: React.FC = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [rememberMe, setRememberMe] = useState(true);
  const usernameInputRef = useRef<HTMLInputElement>(null);
  
  const { login } = useAuth();
  const navigate = useNavigate();
  
  // 从本地存储加载之前保存的用户名和密码
  useEffect(() => {
    const savedUsername = localStorage.getItem('mitm_remembered_username');
    const savedPassword = localStorage.getItem('mitm_remembered_password');
    const savedRememberMe = localStorage.getItem('mitm_remember_me');
    
    if (savedRememberMe === 'true') {
      setRememberMe(true);
      
      if (savedUsername) {
        setUsername(savedUsername);
      }
      
      if (savedPassword) {
        setPassword(savedPassword);
      }
    } else {
      setRememberMe(false);
    }
    
    // 自动聚焦用户名输入框，如果没有保存的用户名
    if (!savedUsername && usernameInputRef.current) {
      usernameInputRef.current.focus();
    }
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!username || !password) {
      setError('请输入用户名和密码');
      return;
    }
    
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
        localStorage.setItem('mitm_remember_me', rememberMe.toString());
        
        // 如果用户选择了"记住我"，保存凭据
        if (rememberMe) {
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

  const togglePasswordVisibility = () => {
    setShowPassword(!showPassword);
  };
  
  return (
    <div className="login-container">
      <div className="login-banner">
        <div className="banner-content">
          <h1>MITM OpenAI Server</h1>
          <p>监控、拦截和分析AI API请求的工具</p>
        </div>
      </div>
      
      <div className="login-card">
        <h2 className="login-title">登录</h2>
        
        {error && (
          <div className="alert alert-danger">
            <i className="fas fa-exclamation-circle" style={{ marginRight: '8px' }}></i>
            {error}
          </div>
        )}
        
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="username" className="form-label">用户名</label>
            <input
              type="text"
              id="username"
              ref={usernameInputRef}
              className="form-control"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              disabled={loading}
              placeholder="请输入用户名"
              autoComplete="username"
            />
          </div>
          
          <div className="form-group">
            <label htmlFor="password" className="form-label">密码</label>
            <div className="password-input-container">
              <input
                type={showPassword ? "text" : "password"}
                id="password"
                className="form-control"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={loading}
                placeholder="请输入密码"
                autoComplete="current-password"
              />
              <button 
                type="button" 
                className="password-toggle-btn"
                onClick={togglePasswordVisibility}
                aria-label={showPassword ? "隐藏密码" : "显示密码"}
              >
                {showPassword ? 
                  <i className="fas fa-eye-slash"></i> : 
                  <i className="fas fa-eye"></i>
                }
              </button>
            </div>
          </div>
          
          <div className="form-group remember-me">
            <label className="checkbox-container">
              <input
                type="checkbox"
                checked={rememberMe}
                onChange={(e) => setRememberMe(e.target.checked)}
                disabled={loading}
              />
              <span className="checkmark"></span>
              记住登录信息
            </label>
          </div>
          
          <button 
            type="submit" 
            className="btn btn-primary btn-block"
            disabled={loading}
          >
            {loading ? (
              <>
                <i className="fas fa-circle-notch fa-spin" style={{ marginRight: '8px' }}></i>
                登录中...
              </>
            ) : '登录'}
          </button>
        </form>
      </div>
    </div>
  );
};

export default LoginPage; 