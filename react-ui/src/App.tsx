import React, { useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, theme } from 'antd';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import RequestsPage from './pages/RequestsPage';
import RequestDetailPage from './pages/RequestDetailPage';
import SettingsPage from './pages/SettingsPage';
import GuidePage from './pages/GuidePage';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { NotificationProvider } from './components/Notification';
import './styles/Global.css';

// 调试认证状态的函数
const DebugAuthStatus = () => {
  useEffect(() => {
    const authToken = localStorage.getItem('auth_token');
    console.log('当前认证状态:', !!authToken);
    if (authToken) {
      console.log('Token长度:', authToken.length);
      console.log('Token前10个字符:', authToken.substring(0, 10) + '...');
      console.log('认证头示例:', `Bearer ${authToken.substring(0, 10)}...`);
      
      // 模拟API请求，打印出实际发送的认证头
      const headers = new Headers();
      if (authToken) {
        headers.append('Authorization', `Bearer ${authToken}`);
      }
      console.log('模拟请求头:', {
        'Content-Type': 'application/json',
        'Authorization': headers.get('Authorization')
      });
    }
  }, []);
  
  return null;
};

// 受保护的路由组件
const ProtectedRoute: React.FC<{children: React.ReactNode}> = ({ children }) => {
  const { isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return <Navigate to="/login" />;
  }

  return (
    <>
      <DebugAuthStatus />
      {children}
    </>
  );
};

// 定义主题配置
const themeConfig = {
  token: {
    colorPrimary: '#1890ff',
    borderRadius: 4,
  },
  algorithm: theme.defaultAlgorithm,
};

const App: React.FC = () => {
  return (
    <ConfigProvider theme={themeConfig}>
    <AuthProvider>
      <NotificationProvider>
        <Router basename="/ui">
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            
            <Route path="/" element={
              <ProtectedRoute>
                <RequestsPage />
              </ProtectedRoute>
            } />

            <Route path="/requests/:id" element={
              <ProtectedRoute>
                <RequestDetailPage />
              </ProtectedRoute>
            } />

            <Route path="/settings" element={
              <ProtectedRoute>
                <SettingsPage />
              </ProtectedRoute>
            } />

            <Route path="/guide" element={
              <ProtectedRoute>
                <GuidePage />
              </ProtectedRoute>
            } />
          </Routes>
        </Router>
      </NotificationProvider>
    </AuthProvider>
    </ConfigProvider>
  );
};

export default App; 