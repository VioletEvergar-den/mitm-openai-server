import React from 'react';
import { Spin } from 'antd';
import './LoadingState.css';

const LoadingState: React.FC = () => {
  return (
    <div className="loading-container">
      <Spin size="large" tip="加载请求详情..." />
    </div>
  );
};

export default LoadingState; 