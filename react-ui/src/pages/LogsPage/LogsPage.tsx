import React, { useState, useEffect, useRef } from 'react';
import {
  Card,
  Typography,
  Button,
  Space,
  Tag,
  Select,
  Spin,
  Empty,
  Tooltip,
  message,
  Row,
  Col,
  Switch
} from 'antd';
import {
  PlayCircleOutlined,
  PauseCircleOutlined,
  ClearOutlined,
  DownloadOutlined,
  FilterOutlined,
  FileTextOutlined,
  SyncOutlined
} from '@ant-design/icons';
import Layout from '../../components/Layout/Layout';
import { useTheme } from '../../contexts/ThemeContext';

const { Title, Text, Paragraph } = Typography;
const { Option } = Select;

interface LogEntry {
  timestamp: string;
  level: string;
  message: string;
}

const LogsPage: React.FC = () => {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [isStreaming, setIsStreaming] = useState(false);
  const [levelFilter, setLevelFilter] = useState<string>('all');
  const [loading, setLoading] = useState(true);
  const [autoScroll, setAutoScroll] = useState(true);
  const logContainerRef = useRef<HTMLDivElement>(null);
  const eventSourceRef = useRef<EventSource | null>(null);
  const { mode } = useTheme();

  useEffect(() => {
    fetchLogs();
    return () => {
      stopStreaming();
    };
  }, []);

  useEffect(() => {
    if (autoScroll && logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  const fetchLogs = async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem('auth_token');
      const response = await fetch('/ui/api/logs?count=100', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      const data = await response.json();
      if (data.code === 0 && data.data) {
        setLogs(data.data);
      }
    } catch (error) {
      message.error('获取日志失败');
    } finally {
      setLoading(false);
    }
  };

  const startStreaming = () => {
    const token = localStorage.getItem('auth_token');
    if (!token) {
      message.error('未登录，无法获取日志');
      return;
    }

    const url = `/ui/api/logs/stream?token=${encodeURIComponent(token)}`;
    const eventSource = new EventSource(url);
    
    eventSource.onopen = () => {
      setIsStreaming(true);
      message.success('日志流已连接');
    };

    eventSource.addEventListener('log', (event) => {
      try {
        const entry = JSON.parse(event.data);
        setLogs(prev => {
          const newLogs = [...prev, entry];
          return newLogs.slice(-500);
        });
      } catch (e) {
        console.error('解析日志失败:', e);
      }
    });

    eventSource.addEventListener('connected', (event) => {
      console.log('SSE connected:', event.data);
    });

    eventSource.onerror = (error) => {
      console.error('SSE error:', error);
      setIsStreaming(false);
      message.error('日志流连接断开');
      eventSource.close();
    };

    eventSourceRef.current = eventSource;
  };

  const stopStreaming = () => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    setIsStreaming(false);
  };

  const clearLogs = async () => {
    try {
      const token = localStorage.getItem('auth_token');
      const response = await fetch('/ui/api/logs', {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });
      const data = await response.json();
      if (data.code === 0) {
        setLogs([]);
        message.success('日志已清除');
      } else {
        message.error(data.msg || '清除日志失败');
      }
    } catch (error) {
      message.error('清除日志失败');
    }
  };

  const exportLogs = () => {
    const content = logs
      .map(log => `[${log.timestamp}] [${log.level}] ${log.message}`)
      .join('\n');
    
    const blob = new Blob([content], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `logs_${new Date().toISOString().slice(0, 10)}.txt`;
    a.click();
    URL.revokeObjectURL(url);
    message.success('日志已导出');
  };

  const getLevelColor = (level: string) => {
    switch (level.toUpperCase()) {
      case 'ERROR':
        return 'error';
      case 'WARN':
        return 'warning';
      case 'INFO':
        return 'processing';
      case 'DEBUG':
        return 'default';
      default:
        return 'default';
    }
  };

  const filteredLogs = logs.filter(log => {
    if (levelFilter === 'all') return true;
    return log.level.toUpperCase() === levelFilter.toUpperCase();
  });

  const formatTimestamp = (timestamp: string) => {
    try {
      const date = new Date(timestamp);
      return date.toLocaleTimeString('zh-CN', { 
        hour: '2-digit', 
        minute: '2-digit', 
        second: '2-digit',
        fractionalSecondDigits: 3
      });
    } catch {
      return timestamp;
    }
  };

  const logContainerStyle: React.CSSProperties = {
    backgroundColor: mode === 'dark' ? '#0d1117' : '#f6f8fa',
    color: mode === 'dark' ? '#c9d1d9' : '#24292f',
    padding: 16,
    borderRadius: 8,
    height: 550,
    overflowY: 'auto',
    fontFamily: 'Consolas, Monaco, "Courier New", monospace',
    fontSize: 13,
    lineHeight: 1.7,
    border: mode === 'dark' ? '1px solid #30363d' : '1px solid #d0d7de'
  };

  const errorCount = logs.filter(l => l.level === 'ERROR').length;
  const warnCount = logs.filter(l => l.level === 'WARN').length;
  const infoCount = logs.filter(l => l.level === 'INFO').length;

  return (
    <Layout title="实时日志">
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Row justify="space-between" align="middle" gutter={[16, 16]}>
          <Col>
            <Title level={3} style={{ margin: 0 }}>
              <FileTextOutlined style={{ marginRight: 8 }} />
              实时日志
            </Title>
          </Col>
          <Col>
            <Space>
              {!isStreaming ? (
                <Button 
                  type="primary" 
                  icon={<PlayCircleOutlined />} 
                  onClick={startStreaming}
                >
                  开始实时日志
                </Button>
              ) : (
                <Button 
                  icon={<PauseCircleOutlined />} 
                  onClick={stopStreaming}
                  danger
                >
                  停止实时日志
                </Button>
              )}
            </Space>
          </Col>
        </Row>

        <Row gutter={16}>
          <Col xs={24} sm={8}>
            <Card size="small" bordered={false} style={{ background: 'transparent' }}>
              <Text type="secondary">错误日志</Text>
              <Title level={4} style={{ margin: 0, color: errorCount > 0 ? '#ff4d4f' : undefined }}>
                {errorCount}
              </Title>
            </Card>
          </Col>
          <Col xs={24} sm={8}>
            <Card size="small" bordered={false} style={{ background: 'transparent' }}>
              <Text type="secondary">警告日志</Text>
              <Title level={4} style={{ margin: 0, color: warnCount > 0 ? '#faad14' : undefined }}>
                {warnCount}
              </Title>
            </Card>
          </Col>
          <Col xs={24} sm={8}>
            <Card size="small" bordered={false} style={{ background: 'transparent' }}>
              <Text type="secondary">信息日志</Text>
              <Title level={4} style={{ margin: 0, color: '#1890ff' }}>
                {infoCount}
              </Title>
            </Card>
          </Col>
        </Row>

        <Card size="small" bordered={false} style={{ background: 'transparent' }}>
          <Row justify="space-between" align="middle" wrap gutter={[16, 16]}>
            <Col>
              <Space wrap>
                <Button 
                  icon={<SyncOutlined spin={loading} />}
                  onClick={fetchLogs}
                  loading={loading}
                  size="small"
                >
                  刷新
                </Button>
                <Button 
                  icon={<ClearOutlined />} 
                  onClick={clearLogs}
                  size="small"
                >
                  清除
                </Button>
                <Button 
                  icon={<DownloadOutlined />} 
                  onClick={exportLogs}
                  size="small"
                >
                  导出
                </Button>
              </Space>
            </Col>
            <Col>
              <Space wrap>
                <FilterOutlined />
                <Select 
                  value={levelFilter} 
                  onChange={setLevelFilter}
                  style={{ width: 120 }}
                  size="small"
                >
                  <Option value="all">全部级别</Option>
                  <Option value="ERROR">ERROR</Option>
                  <Option value="WARN">WARN</Option>
                  <Option value="INFO">INFO</Option>
                  <Option value="DEBUG">DEBUG</Option>
                </Select>
                
                <Switch
                  checkedChildren="自动滚动"
                  unCheckedChildren="手动滚动"
                  checked={autoScroll}
                  onChange={setAutoScroll}
                  size="small"
                />
              </Space>
            </Col>
          </Row>
        </Card>

        <div 
          ref={logContainerRef}
          style={logContainerStyle}
        >
          {loading ? (
            <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
              <Spin tip="加载日志中..." />
            </div>
          ) : filteredLogs.length === 0 ? (
            <Empty 
              description="暂无日志" 
              style={{ marginTop: 150 }}
              image={Empty.PRESENTED_IMAGE_SIMPLE}
            />
          ) : (
            filteredLogs.map((log, index) => (
              <div key={index} style={{ marginBottom: 2, display: 'flex', alignItems: 'flex-start' }}>
                <Text 
                  style={{ 
                    color: mode === 'dark' ? '#7ee787' : '#1a7f37', 
                    fontFamily: 'inherit',
                    minWidth: 100,
                    flexShrink: 0
                  }}
                >
                  [{formatTimestamp(log.timestamp)}]
                </Text>
                <Tag 
                  color={getLevelColor(log.level)} 
                  style={{ marginLeft: 8, marginRight: 8, minWidth: 55, textAlign: 'center' }}
                >
                  {log.level}
                </Tag>
                <Text 
                  style={{ 
                    color: log.level === 'ERROR' ? (mode === 'dark' ? '#f85149' : '#cf222e') : 
                           log.level === 'WARN' ? (mode === 'dark' ? '#d29922' : '#9a6700') : 
                           (mode === 'dark' ? '#c9d1d9' : '#24292f'),
                    fontFamily: 'inherit',
                    wordBreak: 'break-all'
                  }}
                >
                  {log.message}
                </Text>
              </div>
            ))
          )}
        </div>

        <div style={{ textAlign: 'right' }}>
          <Space>
            <Tag color={isStreaming ? 'green' : 'default'}>
              {isStreaming ? '实时连接中' : '未连接'}
            </Tag>
            <Text type="secondary">
              共 {filteredLogs.length} 条日志
              {levelFilter !== 'all' && ` (筛选: ${levelFilter})`}
            </Text>
          </Space>
        </div>
      </Space>
    </Layout>
  );
};

export default LogsPage;
