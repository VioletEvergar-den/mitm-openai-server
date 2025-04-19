import React from 'react';
import { Result, Button } from 'antd';
import { useNavigate } from 'react-router-dom';
import './ErrorState.css';

interface ErrorStateProps {
  errorMessage?: string;
}

const ErrorState: React.FC<ErrorStateProps> = ({ errorMessage }) => {
  const navigate = useNavigate();
  
  return (
    <Result
      status="error"
      title="请求加载失败"
      subTitle={errorMessage || "无法找到请求详情"}
      extra={
        <Button type="primary" onClick={() => navigate('/requests')}>
          返回请求列表
        </Button>
      }
    />
  );
};

export default ErrorState; 