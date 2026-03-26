import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Table,
  Button,
  Input,
  Space,
  Tag,
  Select,
  message,
  Tooltip,
  Switch,
  InputNumber,
  Modal,
  Row,
  Col,
  Typography,
  Card,
  Statistic
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  SearchOutlined,
  ReloadOutlined,
  ClockCircleOutlined,
  DeleteOutlined,
  ExclamationCircleOutlined,
  ApiOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined
} from '@ant-design/icons';
import { RequestRecord } from '../../types/RequestRecord';
import { RequestService } from '../../services/RequestService';
import Layout from '../../components/Layout/Layout';

const { Title, Text } = Typography;
const { Option } = Select;

const getStoredPageSize = (): number => {
  const storedSize = localStorage.getItem('requestsPageSize');
  return storedSize ? parseInt(storedSize, 10) : 20;
};

const getStoredAutoRefresh = (): boolean => {
  const storedAutoRefresh = localStorage.getItem('requestsAutoRefresh');
  return storedAutoRefresh === null ? true : storedAutoRefresh === 'true';
};

const getStoredRefreshInterval = (): number => {
  const storedInterval = localStorage.getItem('requestsRefreshInterval');
  return storedInterval ? parseInt(storedInterval, 10) : 3;
};

const RequestsPage: React.FC = () => {
  const navigate = useNavigate();
  const searchInput = useRef<any>(null);
  const [searchText, setSearchText] = useState<string>('');
  const [searchedColumn, setSearchedColumn] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(false);
  const [requestRecords, setRequestRecords] = useState<RequestRecord[]>([]);
  const [pageSize, setPageSize] = useState<number>(getStoredPageSize());
  const [autoRefresh, setAutoRefresh] = useState<boolean>(getStoredAutoRefresh());
  const [refreshInterval, setRefreshInterval] = useState<number>(getStoredRefreshInterval());
  const refreshTimerRef = useRef<NodeJS.Timeout | null>(null);

  const loadRequests = async () => {
    try {
      setLoading(true);
      const data = await RequestService.getRequests();
      setRequestRecords(data);
    } catch (error) {
      console.error('Failed to load requests:', error);
      message.error('加载请求数据失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadRequests();
  }, []);

  useEffect(() => {
    if (refreshTimerRef.current) {
      clearInterval(refreshTimerRef.current);
      refreshTimerRef.current = null;
    }

    if (autoRefresh && refreshInterval > 0) {
      refreshTimerRef.current = setInterval(() => {
        loadRequests();
      }, refreshInterval * 1000);
    }

    return () => {
      if (refreshTimerRef.current) {
        clearInterval(refreshTimerRef.current);
      }
    };
  }, [autoRefresh, refreshInterval]);

  const handleSearch = (
    selectedKeys: string[],
    confirm: Function,
    dataIndex: string
  ) => {
    confirm();
    setSearchText(selectedKeys[0]);
    setSearchedColumn(dataIndex);
  };

  const handleReset = (clearFilters: Function) => {
    clearFilters();
    setSearchText('');
  };

  const getColumnSearchProps = (dataIndex: string) => ({
    filterDropdown: ({
      setSelectedKeys,
      selectedKeys,
      confirm,
      clearFilters,
    }: any) => (
      <div style={{ padding: 8 }}>
        <Input
          ref={searchInput}
          placeholder={`搜索 ${dataIndex}`}
          value={selectedKeys[0]}
          onChange={(e) =>
            setSelectedKeys(e.target.value ? [e.target.value] : [])
          }
          onPressEnter={() => handleSearch(selectedKeys, confirm, dataIndex)}
          style={{ marginBottom: 8, display: 'block' }}
        />
        <Space>
          <Button
            type="primary"
            onClick={() => handleSearch(selectedKeys, confirm, dataIndex)}
            icon={<SearchOutlined />}
            size="small"
            style={{ width: 90 }}
          >
            搜索
          </Button>
          <Button
            onClick={() => handleReset(clearFilters)}
            size="small"
            style={{ width: 90 }}
          >
            重置
          </Button>
        </Space>
      </div>
    ),
    filterIcon: (filtered: boolean) => (
      <SearchOutlined style={{ color: filtered ? '#1890ff' : undefined }} />
    ),
    onFilter: (value: any, record: RequestRecord) =>
      record[dataIndex as keyof RequestRecord]
        ?.toString()
        .toLowerCase()
        .includes(value.toString().toLowerCase()) || false,
    filterDropdownProps: {
      onOpenChange: (visible: boolean) => {
        if (visible) {
          setTimeout(() => searchInput.current?.select(), 100);
        }
      }
    },
    render: (text: string) =>
      searchedColumn === dataIndex ? (
        <Tooltip title={text}>{text}</Tooltip>
      ) : (
        <Tooltip title={text}>{text}</Tooltip>
      ),
  });

  const handlePageSizeChange = (value: number) => {
    setPageSize(value);
    localStorage.setItem('requestsPageSize', value.toString());
  };

  const handleAutoRefreshChange = (checked: boolean) => {
    setAutoRefresh(checked);
    localStorage.setItem('requestsAutoRefresh', checked.toString());
  };

  const handleRefreshIntervalChange = (value: number | null) => {
    if (value !== null && value > 0) {
      setRefreshInterval(value);
      localStorage.setItem('requestsRefreshInterval', value.toString());
    }
  };

  const handleDeleteAll = () => {
    Modal.confirm({
      title: '确认清空',
      icon: <ExclamationCircleOutlined />,
      content: '确定要清空所有请求记录吗？此操作不可恢复。',
      okText: '确认',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await RequestService.deleteAllRequests();
          message.success('所有请求记录已删除');
          setRequestRecords([]);
        } catch (error) {
          console.error('Failed to delete all requests:', error);
          message.error('删除请求记录失败');
        }
      }
    });
  };

  const successCount = requestRecords.filter(r => r.statusCode && r.statusCode >= 200 && r.statusCode < 300).length;
  const errorCount = requestRecords.filter(r => r.statusCode && r.statusCode >= 400).length;

  const columns: ColumnsType<RequestRecord> = [
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 180,
      render: (timestamp: string) => {
        if (!timestamp) {
          return <span>-</span>;
        }
        const date = new Date(timestamp);
        return (
          <Tooltip title={date.toLocaleString()}>
            <Text style={{ fontSize: 13 }}>{date.toLocaleString()}</Text>
          </Tooltip>
        );
      },
      sorter: (a, b) => {
        if (!a.timestamp || !b.timestamp) {
          return 0;
        }
        const dateA = new Date(a.timestamp).getTime();
        const dateB = new Date(b.timestamp).getTime();
        return dateA - dateB;
      },
      defaultSortOrder: 'descend',
      showSorterTooltip: false,
    },
    {
      title: '方法',
      dataIndex: 'method',
      key: 'method',
      width: 90,
      filters: [
        { text: 'GET', value: 'GET' },
        { text: 'POST', value: 'POST' },
        { text: 'PUT', value: 'PUT' },
        { text: 'DELETE', value: 'DELETE' },
      ],
      onFilter: (value, record) => record.method === value,
      render: (method: string) => {
        if (!method) {
          return <Tag color="default">未知</Tag>;
        }
        
        let color = 'default';
        switch (method.toUpperCase()) {
          case 'GET':
            color = 'success';
            break;
          case 'POST':
            color = 'processing';
            break;
          case 'PUT':
            color = 'warning';
            break;
          case 'DELETE':
            color = 'error';
            break;
        }
        return <Tag color={color} style={{ minWidth: 60, textAlign: 'center' }}>{method}</Tag>;
      },
    },
    {
      title: '路径',
      dataIndex: 'path',
      key: 'path',
      ellipsis: true,
      ...getColumnSearchProps('path'),
      render: (path: string) => (
        <Tooltip title={path}>
          <Text code style={{ fontSize: 13 }}>{path}</Text>
        </Tooltip>
      ),
    },
    {
      title: '消息内容预览',
      dataIndex: 'body',
      key: 'body',
      ellipsis: true,
      width: 300,
      render: (_, record) => {
        try {
          const body = record.body;
          
          if (!body) {
            return <Text type="secondary">-</Text>;
          }
          
          let content = '';
          
          if (typeof body === 'string') {
            content = body;
          } else if (body.messages && Array.isArray(body.messages) && body.messages.length > 0) {
            const lastMessage = body.messages[body.messages.length - 1];
            content = lastMessage.content || '';
          } else if (body.prompt) {
            content = body.prompt;
          } else {
            content = JSON.stringify(body);
          }
          
          return (
            <Tooltip title={content}>
              <Text style={{ fontSize: 13 }} ellipsis>{content}</Text>
            </Tooltip>
          );
        } catch (error) {
          return <Text type="secondary">解析错误</Text>;
        }
      },
    },
    {
      title: '状态',
      dataIndex: 'statusCode',
      key: 'statusCode',
      width: 90,
      filters: [
        { text: '成功 (2xx)', value: '2xx' },
        { text: '客户端错误 (4xx)', value: '4xx' },
        { text: '服务器错误 (5xx)', value: '5xx' },
      ],
      onFilter: (value, record) => {
        const code = record.statusCode;
        if (value === '2xx') return code >= 200 && code < 300;
        if (value === '4xx') return code >= 400 && code < 500;
        if (value === '5xx') return code >= 500;
        return true;
      },
      render: (statusCode: number) => {
        if (!statusCode) {
          return <Tag color="default">-</Tag>;
        }
        
        let color = 'default';
        let icon = null;
        
        if (statusCode >= 200 && statusCode < 300) {
          color = 'success';
          icon = <CheckCircleOutlined />;
        } else if (statusCode >= 400 && statusCode < 500) {
          color = 'warning';
          icon = <CloseCircleOutlined />;
        } else if (statusCode >= 500) {
          color = 'error';
          icon = <CloseCircleOutlined />;
        }
        
        return (
          <Tag color={color} icon={icon} style={{ minWidth: 50, textAlign: 'center' }}>
            {statusCode}
          </Tag>
        );
      },
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_, record) => (
        <Button 
          type="link" 
          size="small" 
          onClick={(e) => {
            e.stopPropagation();
            navigate(`/requests/${record.id}`);
          }}
        >
          详情
        </Button>
      ),
    },
  ];

  return (
    <Layout title="请求列表">
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Row justify="space-between" align="middle" gutter={[16, 16]}>
          <Col>
            <Title level={3} style={{ margin: 0 }}>
              <ApiOutlined style={{ marginRight: 8 }} />
              请求列表
            </Title>
          </Col>
          <Col>
            <Space>
              <Button
                type="primary"
                icon={<ReloadOutlined spin={loading} />}
                loading={loading}
                onClick={loadRequests}
              >
                刷新
              </Button>
              <Button
                danger
                icon={<DeleteOutlined />}
                onClick={handleDeleteAll}
              >
                清空
              </Button>
            </Space>
          </Col>
        </Row>

        <Row gutter={16}>
          <Col xs={24} sm={8}>
            <Card size="small" bordered={false} style={{ background: 'transparent' }}>
              <Statistic 
                title="总请求数" 
                value={requestRecords.length} 
                prefix={<ApiOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={8}>
            <Card size="small" bordered={false} style={{ background: 'transparent' }}>
              <Statistic 
                title="成功请求" 
                value={successCount} 
                valueStyle={{ color: '#52c41a' }}
                prefix={<CheckCircleOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={8}>
            <Card size="small" bordered={false} style={{ background: 'transparent' }}>
              <Statistic 
                title="失败请求" 
                value={errorCount} 
                valueStyle={{ color: errorCount > 0 ? '#ff4d4f' : undefined }}
                prefix={<CloseCircleOutlined />}
              />
            </Card>
          </Col>
        </Row>

        <Card size="small" bordered={false} style={{ background: 'transparent' }}>
          <Row justify="space-between" align="middle" wrap gutter={[16, 16]}>
            <Col>
              <Space>
                <Switch
                  checkedChildren="自动"
                  unCheckedChildren="手动"
                  checked={autoRefresh}
                  onChange={handleAutoRefreshChange}
                />
                {autoRefresh && (
                  <Space>
                    <ClockCircleOutlined />
                    <InputNumber
                      min={1}
                      max={60}
                      value={refreshInterval}
                      onChange={handleRefreshIntervalChange}
                      addonAfter="秒"
                      size="small"
                    />
                  </Space>
                )}
              </Space>
            </Col>
            <Col>
              <Space>
                <Text type="secondary">每页显示:</Text>
                <Select 
                  value={pageSize} 
                  onChange={handlePageSizeChange}
                  size="small"
                  style={{ width: 80 }}
                >
                  <Option value={10}>10</Option>
                  <Option value={20}>20</Option>
                  <Option value={50}>50</Option>
                  <Option value={100}>100</Option>
                </Select>
              </Space>
            </Col>
          </Row>
        </Card>

        <Table
          columns={columns}
          dataSource={requestRecords}
          rowKey="id"
          pagination={{ 
            pageSize: pageSize,
            showSizeChanger: true,
            pageSizeOptions: ['10', '20', '50', '100'],
            onShowSizeChange: (_, size) => handlePageSizeChange(size),
            showTotal: (total) => `共 ${total} 条记录`,
            showQuickJumper: true
          }}
          loading={loading}
          onRow={(record) => ({
            onClick: () => navigate(`/requests/${record.id}`),
            style: { cursor: 'pointer' }
          })}
          showSorterTooltip={false}
          tableLayout="fixed"
          size="middle"
        />
      </Space>
    </Layout>
  );
};

export default RequestsPage;
