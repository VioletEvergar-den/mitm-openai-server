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
  
  // 保存状态
  const [saving, setSaving] = useState(false);
  
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
    setSaving(true);
    
    try {
      const success = await apiService.saveProxyConfig(proxyConfig);
      
      if (success) {
        addNotification('代理设置已保存', 'success');
      } else {
        addNotification('保存代理设置失败', 'danger');
      }
    } catch (error) {
      console.error('保存代理设置失败:', error);
      addNotification('保存代理设置失败', 'danger');
    } finally {
      setSaving(false);
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
        addNotification('所有请求数据已清空', 'success');
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

  // 处理存储路径设置
  const handleSavePath = async () => {
    setSaving(true);
    
    try {
      const success = await apiService.saveProxyConfig(proxyConfig);
      
      if (success) {
        addNotification('存储路径设置已保存', 'success');
      } else {
        addNotification('保存存储路径设置失败', 'danger');
      }
    } catch (error) {
      console.error('保存存储路径设置失败:', error);
      addNotification('保存存储路径设置失败', 'danger');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Layout title="系统设置">
      <div className="settings-container">
        <div className="settings-header">
          <h1>系统设置</h1>
          <p className="settings-subtitle">配置代理服务器和管理捕获的数据</p>
        </div>
        
        {/* 代理服务器配置 */}
        <div className="settings-card">
          <div className="settings-card-header">
            <h2>代理服务器配置</h2>
            {!loading.proxy && (
              <div className="toggle-switch">
                  <input 
                    type="checkbox" 
                    id="proxy-enabled" 
                    name="enabled"
                    checked={proxyConfig.enabled}
                    onChange={handleChange}
                  />
                <label htmlFor="proxy-enabled">
                  <span className="toggle-track">
                    <span className="toggle-indicator"></span>
                  </span>
                  <span className="toggle-label">{proxyConfig.enabled ? '已启用' : '已禁用'}</span>
                </label>
              </div>
            )}
          </div>
          
          {loading.proxy ? (
            <div className="loading-indicator">
              <div className="spinner"></div>
              <p>正在加载代理配置...</p>
            </div>
          ) : (
            <form id="proxy-form" className="settings-form" onSubmit={handleProxySubmit}>
              <div className="form-info">
                {proxyConfig.enabled ? (
                  <p>代理模式已启用，请求将转发到目标API服务器</p>
                ) : (
                  <p>代理模式已禁用，服务器将返回模拟数据</p>
                )}
              </div>
              
              <div className="form-group">
                <label htmlFor="target-url">目标API服务器地址</label>
                <div className="input-container">
                <input 
                  type="url" 
                  id="target-url" 
                  name="targetURL"
                    placeholder="https://api.openai.com" 
                  className="form-control"
                  value={proxyConfig.targetURL}
                  onChange={handleChange}
                    disabled={!proxyConfig.enabled}
                />
                  <span className="input-icon">🔗</span>
                </div>
                <div className="form-hint">填写目标OpenAI API服务器完整地址</div>
              </div>
              
              <div className="form-group">
                <label htmlFor="auth-type">认证类型</label>
                <div className="select-container">
                <select 
                  id="auth-type" 
                  name="authType" 
                  className="form-control"
                  value={proxyConfig.authType}
                  onChange={handleChange}
                    disabled={!proxyConfig.enabled}
                >
                  <option value="none">无认证</option>
                  <option value="basic">基本认证 (用户名/密码)</option>
                  <option value="token">令牌认证</option>
                </select>
                  <span className="select-arrow">▼</span>
                </div>
              </div>
              
              {proxyConfig.authType === 'basic' && (
                <div className="auth-fields">
                  <div className="form-group">
                    <label htmlFor="username">用户名</label>
                    <div className="input-container">
                    <input 
                      type="text" 
                      id="username" 
                      name="username" 
                      className="form-control"
                      value={proxyConfig.username || ''}
                      onChange={handleChange}
                        disabled={!proxyConfig.enabled}
                    />
                      <span className="input-icon">👤</span>
                    </div>
                  </div>
                  <div className="form-group">
                    <label htmlFor="password">密码</label>
                    <div className="input-container">
                    <input 
                      type="password" 
                      id="password" 
                      name="password" 
                      className="form-control"
                      value={proxyConfig.password || ''}
                      onChange={handleChange}
                        disabled={!proxyConfig.enabled}
                    />
                      <span className="input-icon">🔒</span>
                    </div>
                  </div>
                </div>
              )}
              
              {proxyConfig.authType === 'token' && (
                <div className="auth-fields">
                  <div className="form-group">
                    <label htmlFor="token">访问令牌</label>
                    <div className="input-container">
                    <input 
                      type="password" 
                      id="token" 
                      name="token" 
                      className="form-control"
                      value={proxyConfig.token || ''}
                      onChange={handleChange}
                        disabled={!proxyConfig.enabled}
                    />
                      <span className="input-icon">🔑</span>
                    </div>
                    <div className="form-hint">例如：sk-xxxxxxxxxxxxxxxxxxxx</div>
                  </div>
                </div>
              )}
              
              <div className="form-actions">
                <button 
                  type="submit" 
                  className="btn-primary"
                  disabled={saving}
                >
                  {saving ? (
                    <>
                      <div className="btn-spinner"></div>
                      <span>保存中...</span>
                    </>
                  ) : '保存配置'}
                </button>
              </div>
            </form>
          )}
        </div>
        
        {/* 存储管理 */}
        <div className="settings-card">
          <div className="settings-card-header">
            <h2>数据存储管理</h2>
          </div>
          
          {loading.storage ? (
            <div className="loading-indicator">
              <div className="spinner"></div>
            <p>正在加载存储信息...</p>
            </div>
          ) : storageStats ? (
            <div className="storage-content">
              <div className="form-group">
                <label htmlFor="storagePath">数据存储路径</label>
                <div className="input-container">
                  <input 
                    type="text" 
                    id="storagePath" 
                    name="storagePath" 
                    className="form-control"
                    value={proxyConfig.storagePath || ''}
                    onChange={handleChange}
                    placeholder="请输入本地存储路径，留空使用默认路径"
                  />
                  <span className="input-icon">📁</span>
                </div>
                <div className="form-hint">设置后需要重启服务器才能生效。留空则使用当前目录下的data文件夹</div>
                <button 
                  onClick={() => handleSavePath()}
                  className="btn-primary"
                  disabled={saving}
                  style={{marginTop: '10px'}}
                >
                  保存路径设置
                </button>
              </div>
              
              <div className="stats-container">
                <div className="stat-card">
                  <div className="stat-icon">📊</div>
                  <div className="stat-value">{storageStats.totalRequests}</div>
                  <div className="stat-label">捕获的请求</div>
                </div>
                
                <div className="stat-card">
                  <div className="stat-icon">💾</div>
                  <div className="stat-value">{utils.formatFileSize(storageStats.totalSizeBytes)}</div>
                  <div className="stat-label">总数据大小</div>
                </div>
                
                {storageStats.firstRequestTime && (
                  <div className="stat-card">
                    <div className="stat-icon">🕒</div>
                    <div className="stat-value">{new Date(storageStats.firstRequestTime).toLocaleDateString()}</div>
                    <div className="stat-label">首次捕获时间</div>
                  </div>
                )}
                
                {storageStats.lastRequestTime && (
                  <div className="stat-card">
                    <div className="stat-icon">🕘</div>
                    <div className="stat-value">{new Date(storageStats.lastRequestTime).toLocaleDateString()}</div>
                    <div className="stat-label">最近捕获时间</div>
                  </div>
                )}
              </div>
              
              <div className="storage-actions">
                <button 
                  onClick={handleExportData} 
                  className="btn-secondary"
                  disabled={storageStats.totalRequests === 0}
                >
                  <span className="btn-icon">📤</span>
                  导出数据
                </button>
                <button 
                  onClick={handleClearData} 
                  className="btn-danger"
                  disabled={storageStats.totalRequests === 0}
                >
                  <span className="btn-icon">🗑️</span>
                  清空所有数据
                </button>
              </div>
            </div>
          ) : (
            <div className="error-state">
            <p>无法获取存储信息</p>
              <button onClick={loadStorageStats} className="btn-secondary">重试</button>
            </div>
          )}
        </div>
      </div>
    </Layout>
  );
};

export default SettingsPage; 