import React, { useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import './styles'; // 导入所有样式
import { Tabs, Divider, Card } from 'antd';
import Layout from '../../components/Layout';
import {
  RequestHeader,
  RequestSummary,
  RequestContent,
  ResponseContent,
  LoadingState,
  ErrorState
} from './components';
import { useNavigationKeys, useRequestDetail } from './hooks';
import { 
  getMethodColor, 
  getStatusColor, 
  jsonPrettyTheme, 
  copyToClipboard,
  confirmDeleteRequest 
} from './utils';

const { TabPane } = Tabs;

const RequestDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  
  // 使用自定义Hook获取请求数据
  const { request, loading, error, prevRequestId, nextRequestId } = useRequestDetail(id);

  // 导航回调函数
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

  // 使用键盘导航钩子
  useNavigationKeys({
    prevAction: navigateToPrevious,
    nextAction: navigateToNext,
    hasPrev: !!prevRequestId,
    hasNext: !!nextRequestId
  });

  // 删除请求处理
  const handleDelete = useCallback(() => {
    if (!id) return;
    confirmDeleteRequest(id, () => navigate('/requests'));
  }, [id, navigate]);

  // 加载中状态
  if (loading) {
    return (
      <Layout>
        <LoadingState />
      </Layout>
    );
  }

  // 错误状态
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
            requestData={request}
            requestId={request.id}
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