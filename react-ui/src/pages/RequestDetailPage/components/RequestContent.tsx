import React from 'react';
import { Space } from 'antd';
import JsonCardView from './JsonCardView';

interface RequestContentProps {
  requestHeaders: any;
  queryParams?: any;
  requestBody?: any;
  copyToClipboard: (text: any) => Promise<boolean>;
  jsonPrettyTheme?: any;
}

const RequestContent: React.FC<RequestContentProps> = ({
  requestHeaders,
  queryParams,
  requestBody,
  copyToClipboard,
  jsonPrettyTheme
}) => {
  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <JsonCardView
        title="请求头"
        data={requestHeaders}
        copyToClipboard={copyToClipboard}
        jsonPrettyTheme={jsonPrettyTheme}
      />
      
      {queryParams && Object.keys(queryParams).length > 0 && (
        <JsonCardView
          title="查询参数"
          data={queryParams}
          copyToClipboard={copyToClipboard}
          jsonPrettyTheme={jsonPrettyTheme}
        />
      )}
      
      {requestBody && (
        <JsonCardView
          title="请求体"
          data={requestBody}
          copyToClipboard={copyToClipboard}
          jsonPrettyTheme={jsonPrettyTheme}
        />
      )}
    </Space>
  );
};

export default RequestContent; 