import React, { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { apiService } from '../../services/api';
import './RequestDetailPage.css';
import { Tabs, message, Modal, Divider, Card } from 'antd';
import { ExclamationCircleOutlined } from '@ant-design/icons';
import Layout from '../../components/Layout';
import {
  RequestHeader,
  RequestSummary,
  RequestContent,
  ResponseContent,
  LoadingState,
  ErrorState
} from './components';

const { TabPane } = Tabs;
const { confirm } = Modal;

// 自定义JSON主题
const jsonPrettyTheme = {
  main: 'line-height:1.3;color:#66d9ef;background:#272822;overflow:auto;padding:12px;border-radius:4px;',
  key: 'color:#f92672;',
  string: 'color:#a6e22e;',
  value: 'color:#fd971f;',
  boolean: 'color:#ac81fe;',
};

const RequestDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [request, setRequest] = useState<any>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [prevRequestId, setPrevRequestId] = useState<string | null>(null);
  const [nextRequestId, setNextRequestId] = useState<string | null>(null);

  // 使用useCallback定义导航函数，以便在useEffect依赖中使用
  const navigateToPrevious = useCallback(() => {
    if (prevRequestId) {
      navigate(`/requests/${prevRequestId}`);
    }
  }, [navigate, prevRequestId]);

  const navigateToNext = useCallback(() => {
    if (nextRequestId) {
      navigate(`/requests/${nextRequestId}`);
    }
  }, [navigate, nextRequestId]);

  useEffect(() => {
    const fetchData = async () => {
      if (!id) return;
      
      setLoading(true);
      setError(null);
      
      try {
        // 获取当前请求详情
        const requestData: any = await apiService.getRequestById(id);
        console.log('请求详情数据:', requestData); 
        
        // 处理数据字段映射
        if (requestData) {
          // 尝试从路径或URL中提取HOST
          let extractedHost = '';
          const path = requestData.path || requestData.url || '';
          
          // 尝试从路径/URL提取主机名
          if (path) {
            try {
              // 处理完整URL的情况
              if (path.startsWith('http')) {
                const url = new URL(path);
                extractedHost = url.host;
              } 
              // 处理不带协议的URL
              else if (path.includes('//')) {
                extractedHost = path.split('//')[1].split('/')[0];
              }
              // 如果路径中包含主机名信息
              else if (path.startsWith('/') && path.split('/').length > 1) {
                const possibleHost = path.split('/')[1];
                if (possibleHost.includes('.') || possibleHost.includes(':')) {
                  extractedHost = possibleHost;
                }
              }
            } catch (e) {
              console.error('从URL提取主机名失败:', e);
            }
          }
          
          // 适配不同的字段名称
          const adaptedRequest = {
            ...requestData,
            // 请求头可能使用 headers 或 requestHeaders
            requestHeaders: requestData.requestHeaders || requestData.headers || {},
            // 查询参数可能使用 query 或 queryParams
            queryParams: requestData.queryParams || requestData.query || {},
            // 请求体可能使用 body 或 requestBody
            requestBody: requestData.requestBody || requestData.body || null,
            // 响应头可能在 response.headers 或 responseHeaders
            responseHeaders: requestData.responseHeaders || 
                           (requestData.response ? requestData.response.headers : {}) || {},
            // 响应体可能在 response.body 或 responseBody
            responseBody: requestData.responseBody || 
                         (requestData.response ? requestData.response.body : null) || null,
            // 状态码可能在多个不同字段中
            statusCode: requestData.statusCode || 
                       (requestData.response ? requestData.response.statusCode : null) || 
                       requestData.responseCode || 
                       requestData.status_code || 
                       (requestData.responseHeaders && requestData.responseHeaders['status-code']) ||
                       (requestData.responseBody && requestData.responseBody.status) ||
                       200,
            // 延迟可能在 latency 或 response.latency
            latency: requestData.latency || 
                    (requestData.response ? requestData.response.latency : null) ||
                    requestData.responseTime || null,
            // Host可能在不同的字段中
            host: requestData.host || 
                 (requestData.headers && (requestData.headers.host || requestData.headers.Host)) ||
                 (requestData.requestHeaders && (requestData.requestHeaders.host || requestData.requestHeaders.Host)) ||
                 extractedHost ||
                 (requestData.method === 'POST' && path.includes('/v1/') ? 'api.openai.com' : '') || 
                 'localhost',
          };
          
          setRequest(adaptedRequest);
        }
        
        // 获取导航ID
        try {
          const { prevId, nextId } = await apiService.getRequestNavigationIds(id);
          setPrevRequestId(prevId);
          setNextRequestId(nextId);
        } catch (navErr) {
          console.error('获取导航ID失败:', navErr);
        }
      } catch (err) {
        console.error('Failed to fetch request data:', err);
        setError('无法加载请求详情');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [id]);

  // 添加键盘快捷键
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return;
      }
    
      if (e.key === 'ArrowLeft' && prevRequestId) {
        navigateToPrevious();
      }
      else if (e.key === 'ArrowRight' && nextRequestId) {
        navigateToNext();
      }
    };

    window.addEventListener('keydown', handleKeyDown);

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [prevRequestId, nextRequestId, navigateToPrevious, navigateToNext]);

  const handleDelete = () => {
    if (!id) return;

    confirm({
      title: '确认删除',
      icon: <ExclamationCircleOutlined />,
      content: '确定要删除此请求记录吗？此操作不可撤销。',
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      async onOk() {
        try {
          await apiService.deleteRequest(id);
          message.success('请求已成功删除');
          navigate('/requests');
        } catch (err) {
          console.error('Failed to delete request:', err);
          message.error('删除请求失败');
        }
      },
    });
  };

  // 复制到剪贴板的函数
  const copyToClipboard = (text: any) => {
    if (typeof text !== 'string') {
      text = JSON.stringify(text, null, 2);
    }
    
    navigator.clipboard.writeText(text)
      .then(() => {
        message.success('已复制到剪贴板');
      })
      .catch(err => {
        message.error('复制失败');
        console.error('复制失败:', err);
      });
  };

  // 请求方法对应的颜色
  const getMethodColor = (method: string) => {
    const methodMap: Record<string, string> = {
      GET: 'green',
      POST: 'blue',
      PUT: 'orange',
      DELETE: 'red',
      PATCH: 'purple',
    };
    return methodMap[method] || 'default';
  };

  // 状态码对应的颜色
  const getStatusColor = (status: number) => {
    if (status >= 200 && status < 300) return 'success';
    if (status >= 300 && status < 400) return 'processing';
    if (status >= 400 && status < 500) return 'warning';
    if (status >= 500) return 'error';
    return 'default';
  };

  if (loading) {
    return (
      <Layout>
        <LoadingState />
      </Layout>
    );
  }

  if (error || !request) {
    return (
      <Layout>
        <ErrorState errorMessage={error || undefined} />
      </Layout>
    );
  }

  return (
    <Layout title="请求详情">
      <div className="request-detail-container">
        <Card>
          <RequestHeader
            method={request.method}
            path={request.path}
            prevRequestId={prevRequestId}
            nextRequestId={nextRequestId}
            navigateToPrevious={navigateToPrevious}
            navigateToNext={navigateToNext}
            handleDelete={handleDelete}
            getMethodColor={getMethodColor}
          />

          <RequestSummary
            id={request.id}
            timestamp={request.timestamp}
            latency={request.latency}
            statusCode={request.statusCode}
            host={request.host}
            getStatusColor={getStatusColor}
          />
          
          <Divider />

          <Tabs defaultActiveKey="1">
            <TabPane tab="请求" key="1">
              <RequestContent
                requestHeaders={request.requestHeaders}
                queryParams={request.queryParams}
                requestBody={request.requestBody}
                copyToClipboard={copyToClipboard}
                jsonPrettyTheme={jsonPrettyTheme}
              />
            </TabPane>
            
            <TabPane tab="响应" key="2">
              <ResponseContent
                responseHeaders={request.responseHeaders}
                responseBody={request.responseBody}
                copyToClipboard={copyToClipboard}
                jsonPrettyTheme={jsonPrettyTheme}
              />
            </TabPane>
          </Tabs>
        </Card>
      </div>
    </Layout>
  );
};

export default RequestDetailPage; 