import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Card, Typography, Button, Tag, Space, Tooltip } from 'antd';
import { EyeOutlined, ReloadOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import Layout from '../../components/Layout';
import { Request } from '../../types';
import { apiService, utils } from '../../services/api';
import { useNotification } from '../../components/Notification';
import './RequestsPage.css';

const { Title } = Typography;

const RequestsPage: React.FC = () => {
  const [requests, setRequests] = useState<Request[]>([]);
  const [loading, setLoading] = useState(true);
  const { addNotification } = useNotification();
  const navigate = useNavigate();

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

  // 定义表格列
  const columns: ColumnsType<Request> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      ellipsis: true,
      width: 220,
    },
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      render: (timestamp) => utils.formatDateTime(timestamp),
      width: 180,
    },
    {
      title: '方法',
      dataIndex: 'method',
      key: 'method',
      render: (method) => {
        const color = 
          method === 'GET' ? 'green' : 
          method === 'POST' ? 'blue' : 
          method === 'PUT' ? 'orange' : 
          method === 'DELETE' ? 'red' : 'default';
        
        return <Tag color={color}>{method}</Tag>;
      },
      width: 100,
    },
    {
      title: '路径',
      dataIndex: 'path',
      key: 'path',
      ellipsis: {
        showTitle: false,
      },
      render: (path) => (
        <Tooltip placement="topLeft" title={path}>
          {path}
        </Tooltip>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_, record) => (
        <Button 
          type="primary" 
          icon={<EyeOutlined />} 
          onClick={() => navigate(`/requests/${record.id}`)}
        >
          查看
        </Button>
      ),
    },
  ];

  // 表格标题区域
  const title = () => (
    <Space style={{ display: 'flex', justifyContent: 'space-between' }}>
      <Title level={4} style={{ margin: 0 }}>捕获的请求列表</Title>
      <Button 
        type="primary" 
        icon={<ReloadOutlined />} 
        onClick={loadRequests}
        loading={loading}
      >
        刷新
      </Button>
    </Space>
  );

  return (
    <Layout title="请求列表">
      <Card>
        <Table 
          columns={columns} 
          dataSource={requests} 
          rowKey="id"
          loading={loading}
          pagination={{ 
            pageSize: 10,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
          title={title}
          locale={{ 
            emptyText: '暂无捕获的请求数据' 
          }}
        />
      </Card>
    </Layout>
  );
};

export default RequestsPage; 