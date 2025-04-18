// 全局配置
const config = {
    apiBasePath: '/ui/api',
    serverApiBasePath: '/',
    pages: {
        login: 'login.html',
        requests: 'index.html',
        requestDetail: 'request-detail.html',
        guide: 'guide.html',
        settings: 'settings.html'
    }
};

// 工具函数
const utils = {
    // 获取URL参数
    getUrlParam: function(name) {
        const urlParams = new URLSearchParams(window.location.search);
        return urlParams.get(name);
    },
    
    // 格式化日期时间
    formatDateTime: function(dateString) {
        const date = new Date(dateString);
        return date.toLocaleString();
    },
    
    // 截断长字符串
    truncate: function(str, length = 50) {
        if (!str) return '';
        return str.length > length ? str.substr(0, length) + '...' : str;
    },
    
    // 保存基本认证凭证
    saveAuth: function(username, password) {
        const auth = btoa(`${username}:${password}`);
        localStorage.setItem('auth', auth);
    },
    
    // 获取基本认证凭证
    getAuth: function() {
        return localStorage.getItem('auth');
    },
    
    // 清除认证凭证
    clearAuth: function() {
        localStorage.removeItem('auth');
    },
    
    // 创建认证请求头
    createAuthHeaders: function() {
        const auth = this.getAuth();
        return auth ? { 'Authorization': `Basic ${auth}` } : {};
    },
    
    // 检查是否已认证
    isAuthenticated: function() {
        return !!this.getAuth();
    },
    
    // 重定向到登录页
    redirectToLogin: function() {
        window.location.href = config.pages.login;
    },
    
    // 格式化文件大小
    formatFileSize: function(bytes) {
        if (bytes === 0) return '0 Bytes';
        
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    },
    
    // 显示通知消息
    showNotification: function(message, type = 'success') {
        // 创建通知元素
        const notification = document.createElement('div');
        notification.className = `alert alert-${type}`;
        notification.textContent = message;
        
        // 查找或创建通知容器
        let container = document.querySelector('.notification-container');
        if (!container) {
            container = document.createElement('div');
            container.className = 'notification-container';
            container.style.position = 'fixed';
            container.style.top = '20px';
            container.style.right = '20px';
            container.style.zIndex = '9999';
            document.body.appendChild(container);
        }
        
        // 添加通知到容器
        container.appendChild(notification);
        
        // 3秒后自动删除
        setTimeout(() => {
            notification.remove();
            if (container.childNodes.length === 0) {
                container.remove();
            }
        }, 3000);
    },
    
    // API请求
    api: {
        // 获取所有请求列表
        getRequests: async function() {
            try {
                const headers = utils.createAuthHeaders();
                const response = await fetch(`${config.apiBasePath}/requests`, { 
                    headers 
                });
                
                if (response.status === 401) {
                    utils.clearAuth();
                    utils.redirectToLogin();
                    return null;
                }
                
                return await response.json();
            } catch (error) {
                console.error('获取请求列表失败:', error);
                return null;
            }
        },
        
        // 获取单个请求详情
        getRequestById: async function(id) {
            try {
                const headers = utils.createAuthHeaders();
                const response = await fetch(`${config.apiBasePath}/requests/${id}`, { 
                    headers 
                });
                
                if (response.status === 401) {
                    utils.clearAuth();
                    utils.redirectToLogin();
                    return null;
                }
                
                if (response.status === 404) {
                    return null;
                }
                
                return await response.json();
            } catch (error) {
                console.error('获取请求详情失败:', error);
                return null;
            }
        },
        
        // 删除请求
        deleteRequest: async function(id) {
            try {
                const headers = utils.createAuthHeaders();
                const response = await fetch(`${config.apiBasePath}/requests/${id}`, { 
                    method: 'DELETE',
                    headers 
                });
                
                if (response.status === 401) {
                    utils.clearAuth();
                    utils.redirectToLogin();
                    return false;
                }
                
                return response.status === 200;
            } catch (error) {
                console.error('删除请求失败:', error);
                return false;
            }
        },
        
        // 获取服务器信息
        getServerInfo: async function() {
            try {
                const headers = utils.createAuthHeaders();
                const response = await fetch(`${config.apiBasePath}/server-info`, { 
                    headers 
                });
                
                if (response.status === 401) {
                    utils.clearAuth();
                    utils.redirectToLogin();
                    return null;
                }
                
                return await response.json();
            } catch (error) {
                console.error('获取服务器信息失败:', error);
                return null;
            }
        },
        
        // 获取存储统计信息
        getStorageStats: async function() {
            try {
                const headers = utils.createAuthHeaders();
                const response = await fetch(`${config.apiBasePath}/storage-stats`, { 
                    headers 
                });
                
                if (response.status === 401) {
                    utils.clearAuth();
                    utils.redirectToLogin();
                    return null;
                }
                
                return await response.json();
            } catch (error) {
                console.error('获取存储统计信息失败:', error);
                return null;
            }
        },
        
        // 清空所有请求
        clearAllRequests: async function() {
            try {
                const headers = utils.createAuthHeaders();
                const response = await fetch(`${config.apiBasePath}/requests`, { 
                    method: 'DELETE',
                    headers 
                });
                
                if (response.status === 401) {
                    utils.clearAuth();
                    utils.redirectToLogin();
                    return false;
                }
                
                return response.status === 200;
            } catch (error) {
                console.error('清空请求失败:', error);
                return false;
            }
        },
        
        // 导出请求数据
        exportRequests: async function() {
            try {
                const headers = utils.createAuthHeaders();
                const response = await fetch(`${config.apiBasePath}/export`, { 
                    headers 
                });
                
                if (response.status === 401) {
                    utils.clearAuth();
                    utils.redirectToLogin();
                    return null;
                }
                
                return await response.json();
            } catch (error) {
                console.error('导出请求数据失败:', error);
                return null;
            }
        },
        
        // 获取代理配置
        getProxyConfig: async function() {
            try {
                const headers = utils.createAuthHeaders();
                const response = await fetch(`${config.apiBasePath}/proxy-config`, { 
                    headers 
                });
                
                if (response.status === 401) {
                    utils.clearAuth();
                    utils.redirectToLogin();
                    return null;
                }
                
                return await response.json();
            } catch (error) {
                console.error('获取代理配置失败:', error);
                return null;
            }
        },
        
        // 更新代理配置
        updateProxyConfig: async function(config) {
            try {
                const headers = {
                    ...utils.createAuthHeaders(),
                    'Content-Type': 'application/json'
                };
                
                const response = await fetch(`${config.apiBasePath}/proxy-config`, { 
                    method: 'POST',
                    headers,
                    body: JSON.stringify(config)
                });
                
                if (response.status === 401) {
                    utils.clearAuth();
                    utils.redirectToLogin();
                    return false;
                }
                
                return await response.json();
            } catch (error) {
                console.error('更新代理配置失败:', error);
                return false;
            }
        }
    }
};

// 页面初始化
document.addEventListener('DOMContentLoaded', function() {
    const path = window.location.pathname;
    
    // 如果不是登录页但未认证，则重定向到登录页
    if (!path.includes(config.pages.login) && !utils.isAuthenticated()) {
        utils.redirectToLogin();
        return;
    }
    
    // 根据页面路径初始化不同的功能
    if (path.includes(config.pages.login)) {
        initLoginPage();
    } else if (path.includes(config.pages.requests) || path === '/ui/' || path === '/ui') {
        initRequestsPage();
    } else if (path.includes(config.pages.requestDetail)) {
        initRequestDetailPage();
    } else if (path.includes(config.pages.guide)) {
        initGuidePage();
    } else if (path.includes(config.pages.settings)) {
        initSettingsPage();
    }
    
    // 初始化导航链接
    initNavigation();
});

// 初始化导航栏
function initNavigation() {
    const logoutBtn = document.querySelector('.logout-btn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', function(e) {
            e.preventDefault();
            utils.clearAuth();
            utils.redirectToLogin();
        });
    }
    
    // 高亮当前页面的导航链接
    const path = window.location.pathname;
    const navLinks = document.querySelectorAll('.navbar-link');
    
    navLinks.forEach(link => {
        const href = link.getAttribute('href');
        if (path.includes(href) || (href === 'index.html' && (path === '/ui/' || path === '/ui'))) {
            link.classList.add('active');
        }
    });
}

// 初始化登录页面
function initLoginPage() {
    const loginForm = document.querySelector('#login-form');
    const errorAlert = document.querySelector('#error-alert');
    
    if (loginForm) {
        loginForm.addEventListener('submit', function(e) {
            e.preventDefault();
            
            const username = document.querySelector('#username').value;
            const password = document.querySelector('#password').value;
            
            if (!username || !password) {
                errorAlert.textContent = '请输入用户名和密码';
                errorAlert.style.display = 'block';
                return;
            }
            
            utils.saveAuth(username, password);
            
            // 测试认证是否有效
            utils.api.getRequests().then(data => {
                if (data !== null) {
                    window.location.href = config.pages.requests;
                } else {
                    errorAlert.textContent = '认证失败，请检查用户名和密码';
                    errorAlert.style.display = 'block';
                    utils.clearAuth();
                }
            });
        });
    }
}

// 初始化请求列表页面
function initRequestsPage() {
    const requestsTable = document.querySelector('#requests-table');
    const noDataMessage = document.querySelector('#no-data-message');
    
    utils.api.getRequests().then(requests => {
        if (!requests || requests.length === 0) {
            if (noDataMessage) {
                noDataMessage.style.display = 'block';
            }
            if (requestsTable) {
                requestsTable.style.display = 'none';
            }
            return;
        }
        
        if (noDataMessage) {
            noDataMessage.style.display = 'none';
        }
        
        if (requestsTable) {
            const tbody = requestsTable.querySelector('tbody');
            tbody.innerHTML = '';
            
            requests.forEach(request => {
                const tr = document.createElement('tr');
                tr.innerHTML = `
                    <td>${request.id}</td>
                    <td>${utils.formatDateTime(request.timestamp)}</td>
                    <td>${request.method}</td>
                    <td>${request.path}</td>
                    <td>
                        <a href="${config.pages.requestDetail}?id=${request.id}" class="btn btn-primary btn-sm">查看</a>
                    </td>
                `;
                tbody.appendChild(tr);
            });
            
            requestsTable.style.display = 'table';
        }
    });
}

// 初始化请求详情页面
function initRequestDetailPage() {
    const requestId = utils.getUrlParam('id');
    const detailContainer = document.querySelector('#request-detail');
    const notFoundMessage = document.querySelector('#not-found-message');
    const deleteBtn = document.querySelector('#delete-btn');
    
    if (!requestId) {
        window.location.href = config.pages.requests;
        return;
    }
    
    utils.api.getRequestById(requestId).then(request => {
        if (!request) {
            if (detailContainer) {
                detailContainer.style.display = 'none';
            }
            if (notFoundMessage) {
                notFoundMessage.style.display = 'block';
            }
            return;
        }
        
        if (detailContainer) {
            detailContainer.style.display = 'block';
        }
        if (notFoundMessage) {
            notFoundMessage.style.display = 'none';
        }
        
        // 填充请求详情
        const idElement = document.querySelector('#request-id');
        const timestampElement = document.querySelector('#request-timestamp');
        const methodElement = document.querySelector('#request-method');
        const pathElement = document.querySelector('#request-path');
        const ipElement = document.querySelector('#request-ip');
        const headersElement = document.querySelector('#request-headers');
        const queryElement = document.querySelector('#request-query');
        const bodyElement = document.querySelector('#request-body');
        
        if (idElement) idElement.textContent = request.id;
        if (timestampElement) timestampElement.textContent = utils.formatDateTime(request.timestamp);
        if (methodElement) methodElement.textContent = request.method;
        if (pathElement) pathElement.textContent = request.path;
        if (ipElement) ipElement.textContent = request.ip_address;
        
        // 填充请求头
        if (headersElement && request.headers) {
            let headersHtml = '';
            for (const [key, value] of Object.entries(request.headers)) {
                headersHtml += `<div><strong>${key}:</strong> ${value}</div>`;
            }
            headersElement.innerHTML = headersHtml || '<div>无请求头</div>';
        }
        
        // 填充查询参数
        if (queryElement && request.query) {
            let queryHtml = '';
            for (const [key, value] of Object.entries(request.query)) {
                queryHtml += `<div><strong>${key}:</strong> ${value}</div>`;
            }
            queryElement.innerHTML = queryHtml || '<div>无查询参数</div>';
        }
        
        // 填充请求体
        if (bodyElement && request.body) {
            bodyElement.textContent = JSON.stringify(request.body, null, 2);
        } else if (bodyElement) {
            bodyElement.textContent = '无请求体';
        }
    });
    
    // 删除按钮事件
    if (deleteBtn) {
        deleteBtn.addEventListener('click', function() {
            if (confirm('确定要删除这个请求记录吗？')) {
                utils.api.deleteRequest(requestId).then(success => {
                    if (success) {
                        window.location.href = config.pages.requests;
                    } else {
                        alert('删除失败');
                    }
                });
            }
        });
    }
}

// 初始化指南页面
function initGuidePage() {
    const serverInfoContainer = document.querySelector('#server-info');
    
    if (serverInfoContainer) {
        utils.api.getServerInfo().then(info => {
            if (!info) return;
            
            const versionElement = document.querySelector('#server-version');
            const authElement = document.querySelector('#server-auth');
            const urlElement = document.querySelector('#server-url');
            
            if (versionElement) versionElement.textContent = info.version;
            
            if (authElement) {
                let authText = '未启用';
                if (info.auth && info.auth.enabled) {
                    authText = `已启用 (${info.auth.type})`;
                }
                authElement.textContent = authText;
            }
            
            if (urlElement) {
                const url = `${window.location.protocol}//${window.location.host}${info.openApiPath}`;
                urlElement.textContent = url;
                
                const urlLink = document.querySelector('#server-url-link');
                if (urlLink) {
                    urlLink.href = url;
                }
            }
        });
    }
}

// 初始化设置页面
function initSettingsPage() {
    // 加载代理配置
    loadProxyConfig();
    
    // 加载存储统计信息
    loadStorageStats();
    
    // 初始化事件监听器
    initSettingsEvents();
}

// 加载代理配置
async function loadProxyConfig() {
    const config = await utils.api.getProxyConfig();
    if (!config) return;
    
    // 填充表单
    document.getElementById('proxy-enabled').checked = config.enabled;
    document.getElementById('target-url').value = config.targetURL || '';
    document.getElementById('auth-type').value = config.authType || 'none';
    document.getElementById('username').value = config.username || '';
    
    // 根据认证类型显示相应字段
    toggleAuthFields(config.authType);
}

// 加载存储统计信息
async function loadStorageStats() {
    const stats = await utils.api.getStorageStats();
    if (!stats) return;
    
    // 创建HTML内容
    let html = `
        <p>存储位置: <strong>${stats.path}</strong></p>
        <p>记录数量: <strong>${stats.count}</strong> 条</p>
        <p>存储大小: <span class="file-size">${utils.formatFileSize(stats.size)}</span></p>
    `;
    
    // 更新DOM
    document.getElementById('storage-stats').innerHTML = html;
}

// 初始化设置页面事件监听器
function initSettingsEvents() {
    // 认证类型切换事件
    const authTypeSelect = document.getElementById('auth-type');
    if (authTypeSelect) {
        authTypeSelect.addEventListener('change', function() {
            toggleAuthFields(this.value);
        });
    }
    
    // 代理配置表单提交
    const proxyForm = document.getElementById('proxy-form');
    if (proxyForm) {
        proxyForm.addEventListener('submit', async function(e) {
            e.preventDefault();
            
            // 获取表单数据
            const formData = new FormData(this);
            const enabled = formData.get('enabled') === 'on';
            const targetURL = formData.get('targetURL');
            const authType = formData.get('authType');
            
            // 验证URL
            if (enabled && !targetURL) {
                utils.showNotification('请输入目标API服务器地址', 'danger');
                return;
            }
            
            // 构建配置对象
            const config = {
                enabled: enabled,
                targetURL: targetURL,
                authType: authType,
                updateAuth: true
            };
            
            // 根据认证类型添加认证信息
            if (authType === 'basic') {
                config.username = formData.get('username');
                config.password = formData.get('password');
            } else if (authType === 'token') {
                config.token = formData.get('token');
            }
            
            // 提交更新
            const result = await utils.api.updateProxyConfig(config);
            if (result && result.success) {
                utils.showNotification('代理配置已成功更新');
            } else {
                utils.showNotification('更新代理配置失败', 'danger');
            }
        });
    }
    
    // 导出数据按钮
    const exportBtn = document.getElementById('export-data');
    if (exportBtn) {
        exportBtn.addEventListener('click', async function() {
            const result = await utils.api.exportRequests();
            if (result && result.success) {
                // 创建下载链接
                utils.showNotification('数据导出成功，正在下载...');
                window.location.href = result.download_url;
            } else {
                utils.showNotification('导出数据失败', 'danger');
            }
        });
    }
    
    // 清空数据按钮
    const clearBtn = document.getElementById('clear-data');
    if (clearBtn) {
        clearBtn.addEventListener('click', async function() {
            if (confirm('确定要删除所有请求数据吗？此操作不可恢复！')) {
                const success = await utils.api.clearAllRequests();
                if (success) {
                    utils.showNotification('所有请求数据已清空');
                    // 刷新统计信息
                    loadStorageStats();
                } else {
                    utils.showNotification('清空数据失败', 'danger');
                }
            }
        });
    }
}

// 根据认证类型切换认证字段的显示
function toggleAuthFields(authType) {
    const basicAuthFields = document.getElementById('basic-auth-fields');
    const tokenAuthFields = document.getElementById('token-auth-fields');
    
    // 隐藏所有认证字段
    basicAuthFields.style.display = 'none';
    tokenAuthFields.style.display = 'none';
    
    // 根据类型显示相应字段
    if (authType === 'basic') {
        basicAuthFields.style.display = 'block';
    } else if (authType === 'token') {
        tokenAuthFields.style.display = 'block';
    }
} 