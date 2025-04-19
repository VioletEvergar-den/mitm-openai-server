import React, { createContext, useContext, useState, ReactNode } from 'react';
import { Notification as NotificationType } from '../../types';
import './Notification.css';

// 通知上下文类型
interface NotificationContextType {
  notifications: NotificationType[];
  addNotification: (message: string, type?: 'success' | 'danger') => void;
  removeNotification: (id: string) => void;
}

// 创建上下文
const NotificationContext = createContext<NotificationContextType | undefined>(undefined);

// 通知提供者组件
export const NotificationProvider: React.FC<{children: ReactNode}> = ({ children }) => {
  const [notifications, setNotifications] = useState<NotificationType[]>([]);

  // 添加通知
  const addNotification = (message: string, type: 'success' | 'danger' = 'success') => {
    const id = Date.now().toString();
    const notification: NotificationType = { id, message, type };
    
    setNotifications(prev => [...prev, notification]);
    
    // 3秒后自动移除
    setTimeout(() => {
      removeNotification(id);
    }, 3000);
  };

  // 移除通知
  const removeNotification = (id: string) => {
    setNotifications(prev => prev.filter(notification => notification.id !== id));
  };

  return (
    <NotificationContext.Provider value={{ notifications, addNotification, removeNotification }}>
      {children}
      <NotificationDisplay />
    </NotificationContext.Provider>
  );
};

// 使用通知上下文的钩子
export const useNotification = () => {
  const context = useContext(NotificationContext);
  if (context === undefined) {
    throw new Error('useNotification must be used within a NotificationProvider');
  }
  return context;
};

// 通知显示组件
const NotificationDisplay: React.FC = () => {
  const { notifications, removeNotification } = useContext(NotificationContext)!;

  if (notifications.length === 0) {
    return null;
  }

  return (
    <div className="notification-container">
      {notifications.map(notification => (
        <div 
          key={notification.id} 
          className={`alert alert-${notification.type}`}
          onClick={() => removeNotification(notification.id)}
        >
          {notification.message}
        </div>
      ))}
    </div>
  );
};

export default NotificationDisplay; 