// 请求类型定义
export interface Request {
  id: string;
  timestamp: string;
  method: string;
  path: string;
  ip: string;
  headers: Record<string, string>;
  query: Record<string, string>;
  body?: any;
  response?: {
    statusCode: number;
    headers: Record<string, string>;
    body: any;
    latency?: number;
  };
}

// 存储统计信息
export interface StorageStats {
  totalRequests: number;
  totalSizeBytes: number;
  firstRequestTime?: string;
  lastRequestTime?: string;
}

// 代理配置
export interface ProxyConfig {
  enabled: boolean;
  targetURL: string;
  authType: 'none' | 'basic' | 'token';
  username?: string;
  password?: string;
  token?: string;
  storagePath?: string;
}

// 认证状态
export interface AuthState {
  isAuthenticated: boolean;
  username?: string;
}

// 通知类型
export interface Notification {
  id: string;
  message: string;
  type: 'success' | 'danger';
} 