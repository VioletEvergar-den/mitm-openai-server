import React from 'react';
import { Space, Tag, Button, Tooltip, Typography } from 'antd';
import { ArrowLeftOutlined, ArrowRightOutlined, DeleteOutlined } from '@ant-design/icons';
import './RequestHeader.css';

const { Text } = Typography;

interface RequestHeaderProps {
  method: string;
  path: string;
  prevRequestId: string | null;
  nextRequestId: string | null;
  navigateToPrevious: () => void;
  navigateToNext: () => void;
  handleDelete: () => void;
  getMethodColor: (method: string) => string;
}

const RequestHeader: React.FC<RequestHeaderProps> = ({
  method,
  path,
  prevRequestId,
  nextRequestId,
  navigateToPrevious,
  navigateToNext,
  handleDelete,
  getMethodColor
}) => {
  return (
    <div className="detail-header">
      <div className="detail-title">
        <Space>
          <Tag color={getMethodColor(method)}>{method}</Tag>
          <Text ellipsis style={{ maxWidth: '500px' }}>
            {path}
          </Text>
        </Space>
      </div>
      <div className="detail-actions">
        <Space className="navigation-buttons">
          <Tooltip title={prevRequestId ? "查看上一个请求" : "没有上一个请求"}>
            <Button 
              icon={<ArrowLeftOutlined />} 
              onClick={navigateToPrevious}
              disabled={!prevRequestId}
              type="text"
              className="nav-icon"
            />
          </Tooltip>
          <Tooltip title={nextRequestId ? "查看下一个请求" : "没有下一个请求"}>
            <Button 
              icon={<ArrowRightOutlined />} 
              onClick={navigateToNext}
              disabled={!nextRequestId}
              type="text"
              className="nav-icon"
            />
          </Tooltip>
          <Tooltip title="删除此请求">
            <Button 
              icon={<DeleteOutlined />} 
              onClick={handleDelete}
              danger
              type="text"
              className="nav-icon"
            />
          </Tooltip>
        </Space>
      </div>
    </div>
  );
};

export default RequestHeader; 