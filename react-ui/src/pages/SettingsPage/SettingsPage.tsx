import React, { useState, useEffect } from 'react';
import Layout from '../../components/Layout/Layout';
import { ProxyConfig, StorageStats } from '../../types';
import { apiService, utils } from '../../services/api';
import { useNotification } from '../../components/Notification/Notification';
import './SettingsPage.css';

const SettingsPage: React.FC = () => {
  // 代理设置
  const [proxyConfig, setProxyConfig] = useState<ProxyConfig>({
    enabled: false,
    targetURL: '',
    authType: 'none'
  });
  
  // 存储统计信息
  const [storageStats, setStorageStats] = useState<StorageStats | null>(null);
  
  // 加载状态
  const [loading, setLoading] = useState({
    proxy: true,
    storage: true
  });
  
  const { addNotification } = useNotification();

  // 加载代理配置
  const loadProxyConfig = async () => {
    setLoading(prev => ({ ...prev, proxy: true }));
    try {
      const config = await apiService.getProxyConfig();
      if (config) {
        setProxyConfig(config);
      }
    } catch (error) {
      console.error('加载代理配置失败:', error);
      addNotification('加载代理配置失败', 'danger');
    } finally {
      setLoading(prev => ({ ...prev, proxy: false }));
    }
  };

  // 加载存储统计
  const loadStorageStats = async () => {
    setLoading(prev => ({ ...prev, storage: true }));
    try {
      const stats = await apiService.getStorageStats();
      if (stats) {
        setStorageStats(stats);
      }
    } catch (error) {
      console.error('加载存储统计失败:', error);
      addNotification('加载存储统计失败', 'danger');
    } finally {
      setLoading(prev => ({ ...prev, storage: false }));
    }
  };

  // 首次加载
  useEffect(() => {
    loadProxyConfig();
    loadStorageStats();
  }, []);

  // 处理代理配置表单提交
  const handleProxySubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    try {
      const success = await apiService.saveProxyConfig(proxyConfig);
      
      if (success) {
        addNotification('代理设置已保存');
      } else {
        addNotification('保存代理设置失败', 'danger');
      }
    } catch (error) {
      console.error('保存代理设置失败:', error);
      addNotification('保存代理设置失败', 'danger');
    }
  };

  // 处理字段变更
  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target as HTMLInputElement;
    
    setProxyConfig(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? (e.target as HTMLInputElement).checked : value
    }));
  };

  // 处理清空所有请求
  const handleClearData = async () => {
    if (!window.confirm('确定要清空所有请求数据吗？此操作不可恢复。')) {
      return;
    }
    
    try {
      const success = await apiService.clearAllRequests();
      
      if (success) {
        addNotification('所有请求数据已清空');
        loadStorageStats();
      } else {
        addNotification('清空请求数据失败', 'danger');
      }
    } catch (error) {
      console.error('清空请求数据失败:', error);
      addNotification('清空请求数据失败', 'danger');
    }
  };

  // 处理导出数据
  const handleExportData = () => {
    apiService.exportRequests();
  };

  return (
    <Layout title="系统设置">
      <div className="card">
        <h2 className="card-title">系统设置</h2>
        
        {/* 代理服务器配置 */}
        <div className="settings-section">
          <h3>代理服务器配置</h3>
          {loading.proxy ? (
            <p>正在加载代理配置...</p>
          ) : (
            <form id="proxy-form" className="settings-form" onSubmit={handleProxySubmit}>
              <div className="form-group">
                <label htmlFor="proxy-enabled">
                  <input 
                    type="checkbox" 
                    id="proxy-enabled" 
                    name="enabled"
                    checked={proxyConfig.enabled}
                    onChange={handleChange}
                  />
                  启用代理模式
                </label>
                <p className="form-hint">启用代理模式后，将转发请求到下面配置的目标API服务器</p>
              </div>
              
              <div className="form-group">
                <label htmlFor="target-url">目标API服务器地址</label>
                <input 
                  type="url" 
                  id="target-url" 
                  name="targetURL"
                  placeholder="https://api.example.com" 
                  className="form-control"
                  value={proxyConfig.targetURL}
                  onChange={handleChange}
                />
              </div>
              
              <div className="form-group">
                <label htmlFor="auth-type">认证类型</label>
                <select 
                  id="auth-type" 
                  name="authType" 
                  className="form-control"
                  value={proxyConfig.authType}
                  onChange={handleChange}
                >
                  <option value="none">无认证</option>
                  <option value="basic">基本认证 (用户名/密码)</option>
                  <option value="token">令牌认证</option>
                </select>
              </div>
              
              {proxyConfig.authType === 'basic' && (
                <div className="auth-fields">
                  <div className="form-group">
                    <label htmlFor="username">用户名</label>
                    <input 
                      type="text" 
                      id="username" 
                      name="username" 
                      className="form-control"
                      value={proxyConfig.username || ''}
                      onChange={handleChange}
                    />
                  </div>
                  <div className="form-group">
                    <label htmlFor="password">密码</label>
                    <input 
                      type="password" 
                      id="password" 
                      name="password" 
                      className="form-control"
                      value={proxyConfig.password || ''}
                      onChange={handleChange}
                    />
                  </div>
                </div>
              )}
              
              {proxyConfig.authType === 'token' && (
                <div className="auth-fields">
                  <div className="form-group">
                    <label htmlFor="token">访问令牌</label>
                    <input 
                      type="password" 
                      id="token" 
                      name="token" 
                      className="form-control"
                      value={proxyConfig.token || ''}
                      onChange={handleChange}
                    />
                  </div>
                </div>
              )}
              
              <div className="form-group">
                <button type="submit" className="btn btn-primary">保存配置</button>
              </div>
            </form>
          )}
        </div>
        
        {/* 存储管理 */}
        <div className="settings-section">
          <h3>存储管理</h3>
          {loading.storage ? (
            <p>正在加载存储信息...</p>
          ) : storageStats ? (
            <div>
              <div className="settings-info">
                <p>共捕获了 <strong>{storageStats.totalRequests}</strong> 个请求</p>
                <p>总数据大小: <strong>{utils.formatFileSize(storageStats.totalSizeBytes)}</strong></p>
                {storageStats.firstRequestTime && (
                  <p>首次请求时间: <strong>{utils.formatDateTime(storageStats.firstRequestTime)}</strong></p>
                )}
                {storageStats.lastRequestTime && (
                  <p>最近请求时间: <strong>{utils.formatDateTime(storageStats.lastRequestTime)}</strong></p>
                )}
              </div>
              
              <div className="settings-actions">
                <button onClick={handleExportData} className="btn">导出数据 (JSONL)</button>
                <button onClick={handleClearData} className="btn btn-danger">清空所有请求数据</button>
              </div>
            </div>
          ) : (
            <p>无法获取存储信息</p>
          )}
        </div>
      </div>
    </Layout>
  );
};

export default SettingsPage; 