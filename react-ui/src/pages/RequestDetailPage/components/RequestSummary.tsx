import React from 'react';
import { Descriptions, Tag, Typography, Space } from 'antd';
import { ClockCircleOutlined } from '@ant-design/icons';

interface RequestSummaryProps {
  id: string;
  timestamp: string | number;
  latency?: number | null;
  statusCode?: number;
  host: string;
  getStatusColor: (status: number) => string;
}

const RequestSummary: React.FC<RequestSummaryProps> = ({
  id,
  timestamp,
  latency,
  statusCode = 200,
  host,
  getStatusColor
}) => {
  return (
    <Descriptions bordered column={{ xxl: 4, xl: 3, lg: 3, md: 2, sm: 1, xs: 1 }} style={{ marginTop: 24 }}>
      <Descriptions.Item label="请求ID">{id}</Descriptions.Item>
      <Descriptions.Item label="时间">
        <Space>
          {new Date(timestamp).toLocaleString()}
          {latency != null && (
            <Tag color="blue" icon={<ClockCircleOutlined />}>
              {latency}ms
            </Tag>
          )}
        </Space>
      </Descriptions.Item>
      <Descriptions.Item label="状态码">
        <Tag color={getStatusColor(statusCode)}>
          {statusCode}
        </Tag>
      </Descriptions.Item>
      <Descriptions.Item label="Host">{host}</Descriptions.Item>
    </Descriptions>
  );
};

export default RequestSummary; 