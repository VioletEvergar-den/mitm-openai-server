import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, theme } from 'antd';
import LoginPage from './pages/LoginPage';
import RequestsPage from './pages/RequestsPage';
import RequestDetailPage from './pages/RequestDetailPage';
import SettingsPage from './pages/SettingsPage';
import GuidePage from './pages/GuidePage';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { NotificationProvider } from './components/Notification';
import './styles/Global.css';

// 受保护的路由组件
const ProtectedRoute: React.FC<{children: React.ReactNode}> = ({ children }) => {
  const { isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return <Navigate to="/login" />;
  }

  return <>{children}</>;
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
        <Router>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            
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