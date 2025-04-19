import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import Layout from '../../components/Layout';
import { Request } from '../../types';
import { apiService, utils } from '../../services/api';
import { useNotification } from '../../components/Notification';
import './RequestsPage.css';
import '../../styles/components/Table';

const RequestsPage: React.FC = () => {
  const [requests, setRequests] = useState<Request[]>([]);
  const [loading, setLoading] = useState(true);
  const { addNotification } = useNotification();

  // 加载请求列表
  const loadRequests = async () => {
    setLoading(true);
    try {
      const data = await apiService.getRequests();
      setRequests(data);
    } catch (error) {
      console.error('加载请求列表失败:', error);
      addNotification('加载请求列表失败', 'danger');
    } finally {
      setLoading(false);
    }
  };

  // 首次加载
  useEffect(() => {
    loadRequests();
  }, []);

  return (
    <Layout title="请求列表">
      <div className="card">
        <h2 className="card-title">捕获的请求列表</h2>
        
        {loading ? (
          <p>正在加载请求数据...</p>
        ) : requests.length === 0 ? (
          <p>暂无捕获的请求数据。</p>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>ID</th>
                <th>时间</th>
                <th>方法</th>
                <th>路径</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {requests.map(request => (
                <tr key={request.id}>
                  <td>{request.id}</td>
                  <td>{utils.formatDateTime(request.timestamp)}</td>
                  <td>{request.method}</td>
                  <td>{utils.truncate(request.path, 50)}</td>
                  <td>
                    <Link to={`/requests/${request.id}`} className="btn btn-primary">
                      查看详情
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </Layout>
  );
};

export default RequestsPage; 