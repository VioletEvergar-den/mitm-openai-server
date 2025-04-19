import React from 'react';
import { Button, Tooltip } from 'antd';
import { ExportOutlined } from '@ant-design/icons';
import { exportAsJson } from '../utils';
import './ExportButton.css';

interface ExportButtonProps {
  data: any;
  requestId: string;
}

const ExportButton: React.FC<ExportButtonProps> = ({ data, requestId }) => {
  const handleExport = () => {
    // 导出文件名使用请求ID和当前时间
    const filename = `request_${requestId}_${new Date().getTime()}`;
    exportAsJson(data, filename);
  };

  return (
    <Tooltip title="导出为JSON">
      <Button 
        type="primary"
        icon={<ExportOutlined />}
        onClick={handleExport}
        className="export-button"
      >
        导出
      </Button>
    </Tooltip>
  );
};

export default ExportButton; 