import React from 'react';
import JsonCardView from './JsonCardView';
import './ResponseContent.css';

interface ResponseContentProps {
  responseHeaders: any;
  responseBody?: any;
  copyToClipboard: (text: any) => void;
  jsonPrettyTheme: any;
}

const ResponseContent: React.FC<ResponseContentProps> = ({
  responseHeaders,
  responseBody,
  copyToClipboard,
  jsonPrettyTheme
}) => {
  return (
    <div className="response-content">
      <JsonCardView
        title="响应头"
        data={responseHeaders}
        copyToClipboard={copyToClipboard}
        jsonPrettyTheme={jsonPrettyTheme}
      />
      
      {responseBody && (
        <JsonCardView
          title="响应体"
          data={responseBody}
          copyToClipboard={copyToClipboard}
          style={{ marginTop: 16 }}
          jsonPrettyTheme={jsonPrettyTheme}
        />
      )}
    </div>
  );
};

export default ResponseContent; 