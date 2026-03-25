import React, { useMemo } from 'react';
import { Space } from 'antd';
import JsonCardView from './JsonCardView';

interface ResponseContentProps {
  responseHeaders: any;
  responseBody?: any;
  copyToClipboard: (text: any) => Promise<boolean>;
  jsonPrettyTheme?: any;
}

const ResponseContent: React.FC<ResponseContentProps> = ({
  responseHeaders,
  responseBody,
  copyToClipboard,
  jsonPrettyTheme
}) => {
  // 确保响应头正确显示
  const formattedHeaders = useMemo(() => {
    if (!responseHeaders) return {};
    
    // 如果是字符串，尝试解析为JSON
    if (typeof responseHeaders === 'string') {
      try {
        return JSON.parse(responseHeaders);
      } catch {
        return { raw: responseHeaders };
      }
    }
    
    return responseHeaders;
  }, [responseHeaders]);

  // 确保响应体正确显示
  const formattedBody = useMemo(() => {
    if (!responseBody) return null;
    
    // 如果是字符串，尝试解析为JSON
    if (typeof responseBody === 'string') {
      try {
        return JSON.parse(responseBody);
      } catch {
        // 如果不是有效的JSON，直接返回原始字符串
        return responseBody;
      }
    }
    
    return responseBody;
  }, [responseBody]);

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <JsonCardView
        title="响应头"
        data={formattedHeaders}
        copyToClipboard={copyToClipboard}
        jsonPrettyTheme={jsonPrettyTheme}
      />
      
      {formattedBody && (
        <JsonCardView
          title="响应体"
          data={formattedBody}
          copyToClipboard={copyToClipboard}
          jsonPrettyTheme={jsonPrettyTheme}
        />
      )}
    </Space>
  );
};

export default ResponseContent; 