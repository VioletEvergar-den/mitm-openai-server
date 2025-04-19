import React from 'react';
import { Descriptions, Tag, Typography } from 'antd';
import { ClockCircleOutlined } from '@ant-design/icons';
import './RequestSummary.css';

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
    <Descriptions bordered column={{ xxl: 4, xl: 3, lg: 3, md: 2, sm: 1, xs: 1 }}>
      <Descriptions.Item label="请求ID">{id}</Descriptions.Item>
      <Descriptions.Item label="时间">
        {new Date(timestamp).toLocaleString()}
        {latency && (
          <Tag color="blue" className="latency-badge">
            <ClockCircleOutlined /> {latency}ms
          </Tag>
        )}
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