import React, { createContext, useContext, useState, useEffect } from 'react';
import { AuthState } from '../types';
import { utils } from '../services/api';

interface AuthContextType {
  isAuthenticated: boolean;
  login: (username: string, password: string, expiryDays?: number) => void;
  logout: () => void;
}

const defaultAuthContext: AuthContextType = {
  isAuthenticated: false,
  login: () => {},
  logout: () => {}
};

// 创建上下文
const AuthContext = createContext<AuthContextType>(defaultAuthContext);

// 上下文提供者组件
export const AuthProvider: React.FC<{children: React.ReactNode}> = ({ children }) => {
  const [authState, setAuthState] = useState<AuthState>({
    isAuthenticated: utils.isAuthenticated()
  });

  // 登录方法
  const login = (username: string, password: string, expiryDays = 1) => {
    utils.saveAuth(username, password, expiryDays);
    setAuthState({ isAuthenticated: true, username });
  };

  // 退出登录方法
  const logout = () => {
    utils.clearAuth();
    setAuthState({ isAuthenticated: false });
    window.location.href = '/login';
  };

  // 初始化时检查认证状态
  useEffect(() => {
    setAuthState({
      isAuthenticated: utils.isAuthenticated()
    });
  }, []);

  return (
    <AuthContext.Provider value={{
      isAuthenticated: authState.isAuthenticated,
      login,
      logout
    }}>
      {children}
    </AuthContext.Provider>
  );
};

// 使用上下文的钩子
export const useAuth = () => useContext(AuthContext); 