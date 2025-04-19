import React from 'react';
import JsonCardView from './JsonCardView';
import './RequestContent.css';

interface RequestContentProps {
  requestHeaders: any;
  queryParams?: any;
  requestBody?: any;
  copyToClipboard: (text: any) => void;
  jsonPrettyTheme: any;
}

const RequestContent: React.FC<RequestContentProps> = ({
  requestHeaders,
  queryParams,
  requestBody,
  copyToClipboard,
  jsonPrettyTheme
}) => {
  return (
    <div className="request-content">
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
          style={{ marginTop: 16 }}
          jsonPrettyTheme={jsonPrettyTheme}
        />
      )}
      
      {requestBody && (
        <JsonCardView
          title="请求体"
          data={requestBody}
          copyToClipboard={copyToClipboard}
          style={{ marginTop: 16 }}
          jsonPrettyTheme={jsonPrettyTheme}
        />
      )}
    </div>
  );
};

export default RequestContent; 