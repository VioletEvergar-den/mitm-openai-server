import React, { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { apiService } from '../../services/api';
import './RequestDetailPage.css';
import { Card, Button, Tabs, Typography, Tag, Space, Divider, Descriptions, Spin, Result, message, Tooltip, Modal } from 'antd';
import { ArrowLeftOutlined, ArrowRightOutlined, DeleteOutlined, ExclamationCircleOutlined, ClockCircleOutlined, CodeOutlined, CopyOutlined } from '@ant-design/icons';
import JSONPretty from 'react-json-pretty';
import 'react-json-pretty/themes/monikai.css';
import Layout from '../../components/Layout';

const { Title, Text } = Typography;
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
        console.log('请求详情数据:', requestData); // 添加调试输出
        
        // 处理数据字段映射
        if (requestData) {
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
            // 状态码可能在 statusCode 或 response.statusCode
            statusCode: requestData.statusCode || 
                       (requestData.response ? requestData.response.statusCode : null),
            // 延迟可能在 latency 或 response.latency
            latency: requestData.latency || 
                    (requestData.response ? requestData.response.latency : null),
            // Host可能在不同的字段中
            host: requestData.host || 
                 (requestData.requestHeaders && requestData.requestHeaders.host) ||
                 (requestData.headers && requestData.headers.host) || '',
          };
          setRequest(adaptedRequest);
          console.log('适配后的请求数据:', adaptedRequest); // 添加调试输出
        }
        
        // 获取导航ID
        try {
          const { prevId, nextId } = await apiService.getRequestNavigationIds(id);
          setPrevRequestId(prevId);
          setNextRequestId(nextId);
        } catch (navErr) {
          console.error('获取导航ID失败:', navErr);
          // 导航失败不阻止页面显示
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
      // 避免在输入框中触发
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return;
      }
      
      // 左方向键导航到上一个请求
      if (e.key === 'ArrowLeft' && prevRequestId) {
        navigateToPrevious();
      }
      // 右方向键导航到下一个请求
      else if (e.key === 'ArrowRight' && nextRequestId) {
        navigateToNext();
      }
    };

    // 添加键盘事件监听
    window.addEventListener('keydown', handleKeyDown);

    // 清理函数
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [prevRequestId, nextRequestId, navigateToPrevious, navigateToNext]); // 添加导航函数到依赖列表

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

  // 生成卡片额外操作区域（带复制按钮）
  const renderCardExtra = (title: string, content: any) => (
    <Tooltip title={`复制${title}内容到剪贴板`} placement="left">
      <Button 
        type="text" 
        icon={<CopyOutlined />} 
        onClick={() => copyToClipboard(content)}
        size="small"
        title={`复制${title}`}
      />
    </Tooltip>
  );

  if (loading) {
    return (
      <Layout>
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '400px' }}>
          <Spin size="large" tip="加载请求详情..." />
        </div>
      </Layout>
    );
  }

  if (error || !request) {
    return (
      <Layout>
        <Result
          status="error"
          title="请求加载失败"
          subTitle={error || "无法找到请求详情"}
          extra={
            <Button type="primary" onClick={() => navigate('/requests')}>
              返回请求列表
            </Button>
          }
        />
      </Layout>
    );
  }

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

  return (
    <Layout title="请求详情">
      <div className="request-detail-container">
        <Card
          title={
            <Space>
              <Tag color={getMethodColor(request.method)}>{request.method}</Tag>
              <Text ellipsis style={{ maxWidth: '500px' }}>
                {request.path}
              </Text>
            </Space>
          }
          extra={
            <Space>
              <Tooltip title={prevRequestId ? "查看上一个请求" : "没有上一个请求"}>
                <Button 
                  icon={<ArrowLeftOutlined />} 
                  onClick={navigateToPrevious}
                  disabled={!prevRequestId}
                  type="text"
                />
              </Tooltip>
              <Tooltip title={nextRequestId ? "查看下一个请求" : "没有下一个请求"}>
                <Button 
                  icon={<ArrowRightOutlined />} 
                  onClick={navigateToNext}
                  disabled={!nextRequestId}
                  type="text"
                />
              </Tooltip>
              <Tooltip title="删除此请求">
                <Button 
                  icon={<DeleteOutlined />} 
                  onClick={handleDelete}
                  danger
                  type="text"
                />
              </Tooltip>
            </Space>
          }
        >
          <Descriptions bordered column={{ xxl: 4, xl: 3, lg: 3, md: 2, sm: 1, xs: 1 }}>
            <Descriptions.Item label="请求ID">{request.id}</Descriptions.Item>
            <Descriptions.Item label="时间">
              {new Date(request.timestamp).toLocaleString()}
              {request.latency && (
                <Tag color="blue" className="latency-badge">
                  <ClockCircleOutlined /> {request.latency}ms
                </Tag>
              )}
            </Descriptions.Item>
            <Descriptions.Item label="状态码">
              {request.statusCode && (
                <Tag color={getStatusColor(request.statusCode)}>
                  {request.statusCode}
                </Tag>
              )}
            </Descriptions.Item>
            <Descriptions.Item label="Host">{request.host}</Descriptions.Item>
          </Descriptions>
          
          <Divider />

          <Tabs defaultActiveKey="1">
            <TabPane tab="请求" key="1">
              <Card 
                size="small" 
                title="请求头"
                extra={renderCardExtra('请求头', request.requestHeaders)}
              >
                {request.requestHeaders && Object.keys(request.requestHeaders).length > 0 ? (
                  <div className="json-viewer">
                    <JSONPretty 
                      id="json-pretty-headers"
                      data={request.requestHeaders}
                      theme={jsonPrettyTheme}
                    />
                  </div>
                ) : (
                  <Text type="secondary">无请求头</Text>
                )}
              </Card>
              
              {request.queryParams && Object.keys(request.queryParams).length > 0 && (
                <Card 
                  size="small" 
                  title="查询参数" 
                  style={{ marginTop: 16 }}
                  extra={renderCardExtra('查询参数', request.queryParams)}
                >
                  <div className="json-viewer">
                    <JSONPretty 
                      id="json-pretty-params"
                      data={request.queryParams}
                      theme={jsonPrettyTheme}
                    />
                  </div>
                </Card>
              )}
              
              {request.requestBody && (
                <Card 
                  size="small" 
                  title="请求体" 
                  style={{ marginTop: 16 }}
                  extra={renderCardExtra('请求体', request.requestBody)}
                >
                  <div className="json-viewer">
                    <JSONPretty 
                      id="json-pretty-request-body"
                      data={request.requestBody}
                      theme={jsonPrettyTheme}
                    />
                  </div>
                </Card>
              )}
            </TabPane>
            
            <TabPane tab="响应" key="2">
              <Card 
                size="small" 
                title="响应头"
                extra={renderCardExtra('响应头', request.responseHeaders)}
              >
                {request.responseHeaders && Object.keys(request.responseHeaders).length > 0 ? (
                  <div className="json-viewer">
                    <JSONPretty 
                      id="json-pretty-response-headers"
                      data={request.responseHeaders}
                      theme={jsonPrettyTheme}
                    />
                  </div>
                ) : (
                  <Text type="secondary">无响应头</Text>
                )}
              </Card>
              
              {request.responseBody && (
                <Card 
                  size="small" 
                  title="响应体" 
                  style={{ marginTop: 16 }}
                  extra={renderCardExtra('响应体', request.responseBody)}
                >
                  <div className="json-viewer">
                    <JSONPretty 
                      id="json-pretty-response-body"
                      data={request.responseBody}
                      theme={jsonPrettyTheme}
                    />
                  </div>
                </Card>
              )}
            </TabPane>
          </Tabs>
        </Card>
      </div>
    </Layout>
  );
};

export default RequestDetailPage; 