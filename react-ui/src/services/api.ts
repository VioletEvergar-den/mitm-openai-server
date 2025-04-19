import axios from 'axios';
import { Request, StorageStats, ProxyConfig } from '../types';

// API基础配置
const API = axios.create({
  baseURL: '/ui/api',
  headers: {
    'Content-Type': 'application/json'
  }
});

// 拦截器添加认证信息
API.interceptors.request.use(config => {
  const auth = localStorage.getItem('auth');
  if (auth) {
    config.headers.Authorization = `Basic ${auth}`;
  }
  return config;
});

// 拦截器处理认证失败
API.interceptors.response.use(
  response => response,
  error => {
    if (error.response && error.response.status === 401) {
      localStorage.removeItem('auth');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// 工具函数
export const utils = {
  saveAuth: (username: string, password: string, expiryDays = 365) => {
    const auth = btoa(`${username}:${password}`);
    localStorage.setItem('auth', auth);
    
    // 设置cookie有效期
    const expiryDate = new Date();
    expiryDate.setDate(expiryDate.getDate() + expiryDays);
    
    document.cookie = `auth=${auth}; expires=${expiryDate.toUTCString()}; path=/; SameSite=Strict`;
  },
  
  clearAuth: () => {
    localStorage.removeItem('auth');
    document.cookie = 'auth=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/; SameSite=Strict';
  },
  
  isAuthenticated: () => {
    return !!localStorage.getItem('auth');
  },
  
  formatDateTime: (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString();
  },
  
  truncate: (str: string, length = 50) => {
    if (!str) return '';
    return str.length > length ? str.substring(0, length) + '...' : str;
  },
  
  formatFileSize: (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }
};

// API方法
export const apiService = {
  // 认证
  login: async (username: string, password: string) => {
    try {
      // 设置认证头
      utils.saveAuth(username, password);
      
      // 测试认证是否成功
      await API.get('/server-info');
      return true;
    } catch (error) {
      utils.clearAuth();
      return false;
    }
  },
  
  // 获取请求列表
  getRequests: async (): Promise<Request[]> => {
    try {
      const response = await API.get('/requests');
      // 处理 StandardResponse 格式
      if (response.data && response.data.data) {
        // 如果data.requests存在且是数组，返回请求列表
        if (response.data.data.requests && Array.isArray(response.data.data.requests)) {
          return response.data.data.requests;
        }
        // 向后兼容: 如果data本身是数组，直接返回
        if (Array.isArray(response.data.data)) {
          return response.data.data;
        }
      }
      // 兼容处理，如果response.data本身是数组，则直接返回
      if (Array.isArray(response.data)) {
        return response.data;
      }
      return [];
    } catch (error) {
      console.error('获取请求列表失败:', error);
      return [];
    }
  },
  
  // 获取请求详情
  getRequestById: async (id: string): Promise<Request | null> => {
    try {
      const response = await API.get(`/requests/${id}`);
      // 处理 StandardResponse 格式
      if (response.data && response.data.data) {
        return response.data.data;
      }
      // 兼容处理，如果response.data本身就是请求对象
      if (response.data && response.data.id) {
        return response.data;
      }
      return null;
    } catch (error) {
      console.error(`获取请求详情失败 ID=${id}:`, error);
      return null;
    }
  },
  
  // 删除请求
  deleteRequest: async (id: string): Promise<boolean> => {
    try {
      await API.delete(`/requests/${id}`);
      return true;
    } catch (error) {
      console.error(`删除请求失败 ID=${id}:`, error);
      return false;
    }
  },
  
  // 获取存储统计
  getStorageStats: async (): Promise<StorageStats | null> => {
    try {
      const response = await API.get('/storage-stats');
      // 处理 StandardResponse 格式
      if (response.data && response.data.data) {
        return response.data.data;
      }
      // 兼容处理
      if (response.data && typeof response.data === 'object') {
        return response.data;
      }
      return null;
    } catch (error) {
      console.error('获取存储统计失败:', error);
      return null;
    }
  },
  
  // 清空所有请求
  clearAllRequests: async (): Promise<boolean> => {
    try {
      await API.delete('/requests');
      return true;
    } catch (error) {
      console.error('清空请求失败:', error);
      return false;
    }
  },
  
  // 导出请求数据
  exportRequests: async (): Promise<void> => {
    try {
      const response = await API.get('/export', {
        responseType: 'blob'
      });
      
      // 创建下载链接
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `requests-export-${Date.now()}.jsonl`);
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } catch (error) {
      console.error('导出请求失败:', error);
    }
  },
  
  // 获取代理配置
  getProxyConfig: async (): Promise<ProxyConfig | null> => {
    try {
      const response = await API.get('/proxy-config');
      // 处理 StandardResponse 格式
      if (response.data && response.data.data) {
        return response.data.data;
      }
      // 兼容处理
      if (response.data && typeof response.data === 'object') {
        return response.data;
      }
      return null;
    } catch (error) {
      console.error('获取代理配置失败:', error);
      return null;
    }
  },
  
  // 保存代理配置
  saveProxyConfig: async (config: ProxyConfig): Promise<boolean> => {
    try {
      await API.post('/proxy-config', config);
      return true;
    } catch (error) {
      console.error('保存代理配置失败:', error);
      return false;
    }
  }
}; 