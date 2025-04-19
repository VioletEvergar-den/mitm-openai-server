import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import Layout from '../../components/Layout';
import { Request } from '../../types';
import { apiService, utils } from '../../services/api';
import { useNotification } from '../../components/Notification';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import './RequestDetailPage.css';
import '../../styles/components/Table';

const RequestDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { addNotification } = useNotification();
  
  const [request, setRequest] = useState<Request | null>(null);
  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);

  // 加载请求详情
  const loadRequestDetail = async () => {
    if (!id) return;
    
    setLoading(true);
    try {
      const data = await apiService.getRequestById(id);
      
      if (data) {
        setRequest(data);
      } else {
        setNotFound(true);
      }
    } catch (error) {
      console.error('加载请求详情失败:', error);
      addNotification('加载请求详情失败', 'danger');
    } finally {
      setLoading(false);
    }
  };

  // 删除请求
  const handleDelete = async () => {
    if (!id || !window.confirm('确定要删除此请求吗？')) {
      return;
    }
    
    try {
      const success = await apiService.deleteRequest(id);
      
      if (success) {
        addNotification('请求已成功删除');
        navigate('/');
      } else {
        addNotification('删除请求失败', 'danger');
      }
    } catch (error) {
      console.error('删除请求失败:', error);
      addNotification('删除请求失败', 'danger');
    }
  };

  // 首次加载
  useEffect(() => {
    loadRequestDetail();
  }, [id]);

  // 未找到请求的显示
  if (notFound) {
    return (
      <Layout title="请求未找到">
        <div className="card">
          <h2 className="card-title">未找到请求</h2>
          <p>请求不存在或已被删除。</p>
          <button
            onClick={() => navigate('/')}
            className="btn btn-primary"
          >
            返回请求列表
          </button>
        </div>
      </Layout>
    );
  }

  // 加载中的显示
  if (loading) {
    return (
      <Layout title="加载中">
        <div className="card">
          <p>正在加载请求详情...</p>
        </div>
      </Layout>
    );
  }

  // 请求详情显示
  return (
    <Layout title="请求详情">
      <div className="card">
        <div className="detail-header">
          <h2 className="detail-title">请求详情</h2>
          <div>
            <button
              onClick={handleDelete}
              className="btn btn-danger"
            >
              删除请求
            </button>
            <button
              onClick={() => navigate('/')}
              className="btn btn-primary"
              style={{ marginLeft: '10px' }}
            >
              返回列表
            </button>
          </div>
        </div>

        <div className="detail-section">
          <h3 className="detail-section-title">基本信息</h3>
          <table className="table">
            <tbody>
              <tr>
                <th>ID</th>
                <td>{request?.id}</td>
              </tr>
              <tr>
                <th>时间戳</th>
                <td>{request ? utils.formatDateTime(request.timestamp) : ''}</td>
              </tr>
              <tr>
                <th>请求方法</th>
                <td>{request?.method}</td>
              </tr>
              <tr>
                <th>请求路径</th>
                <td>{request?.path}</td>
              </tr>
              <tr>
                <th>IP地址</th>
                <td>{request?.ip}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div className="detail-section">
          <h3 className="detail-section-title">请求头</h3>
          <div className="json-viewer">
            <SyntaxHighlighter language="json" style={vscDarkPlus}>
              {request?.headers ? JSON.stringify(request.headers, null, 2) : ''}
            </SyntaxHighlighter>
          </div>
        </div>

        <div className="detail-section">
          <h3 className="detail-section-title">查询参数</h3>
          <div className="json-viewer">
            <SyntaxHighlighter language="json" style={vscDarkPlus}>
              {request?.query ? JSON.stringify(request.query, null, 2) : ''}
            </SyntaxHighlighter>
          </div>
        </div>

        <div className="detail-section">
          <h3 className="detail-section-title">请求体</h3>
          <div className="json-viewer">
            <SyntaxHighlighter language="json" style={vscDarkPlus}>
              {request?.body ? JSON.stringify(request.body, null, 2) : ''}
            </SyntaxHighlighter>
          </div>
        </div>

        {request?.response && (
          <div className="detail-section">
            <h3 className="detail-section-title">响应内容</h3>
            <div className="response-detail">
              <div className="response-header">
                <h4>状态码: <span className={`status-${Math.floor(((request.response.statusCode || 200) / 100))}xx`}>{request.response.statusCode || 200}</span></h4>
                {request.response.latency && (
                  <p>响应时间: {request.response.latency} ms</p>
                )}
              </div>
              
              {request.response.headers && Object.keys(request.response.headers).length > 0 && (
                <div className="response-section">
                  <h4>响应头</h4>
                  <div className="json-viewer">
                    <SyntaxHighlighter language="json" style={vscDarkPlus}>
                      {JSON.stringify(request.response.headers, null, 2)}
                    </SyntaxHighlighter>
                  </div>
                </div>
              )}
              
              <div className="response-section">
                <h4>响应体</h4>
                <div className="json-viewer">
                  <SyntaxHighlighter language="json" style={vscDarkPlus}>
                    {request.response.body ? JSON.stringify(request.response.body, null, 2) : ''}
                  </SyntaxHighlighter>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </Layout>
  );
};

export default RequestDetailPage; 