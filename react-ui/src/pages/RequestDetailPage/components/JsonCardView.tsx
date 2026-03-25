import React, { useMemo, useState } from 'react';
import { Card, Button, Tooltip, Typography, Switch, Space } from 'antd';
import { CopyOutlined, ExpandOutlined, CompressOutlined } from '@ant-design/icons';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';

const { Text } = Typography;

interface JsonCardViewProps {
  title: string;
  data: any;
  copyToClipboard: (text: any) => Promise<boolean>;
  style?: React.CSSProperties;
  jsonPrettyTheme?: any;
}

const detectLanguage = (content: string): string => {
  const trimmed = content.trim();
  if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
    return 'json';
  }
  if (trimmed.startsWith('<')) {
    return 'xml';
  }
  return 'text';
};

const JsonCardView: React.FC<JsonCardViewProps> = ({
  title,
  data,
  copyToClipboard,
  style,
}) => {
  const [wrapText, setWrapText] = useState(true);
  const [isExpanded, setIsExpanded] = useState(false);

  const jsonString = useMemo(() => {
    if (data === null || data === undefined) {
      return '';
    }
    if (typeof data === 'string') {
      return data;
    }
    try {
      return JSON.stringify(data, null, 2);
    } catch (error) {
      console.error('JSON序列化错误:', error);
      return String(data);
    }
  }, [data]);

  const hasData = jsonString && jsonString.length > 0;

  const language = useMemo(() => detectLanguage(jsonString), [jsonString]);

  const renderCardExtra = () => (
    <Space size="small">
      <Tooltip title={wrapText ? '关闭自动换行' : '开启自动换行'} placement="left">
        <Switch
          size="small"
          checked={wrapText}
          onChange={setWrapText}
          checkedChildren="换行"
          unCheckedChildren="滚动"
        />
      </Tooltip>
      <Tooltip title={isExpanded ? '折叠内容' : '展开内容'} placement="left">
        <Button
          type="text"
          icon={isExpanded ? <CompressOutlined /> : <ExpandOutlined />}
          onClick={() => setIsExpanded(!isExpanded)}
          size="small"
        />
      </Tooltip>
      <Tooltip title={`复制${title}内容到剪贴板`} placement="left">
        <Button
          type="text"
          icon={<CopyOutlined />}
          onClick={() => copyToClipboard(jsonString)}
          size="small"
          title={`复制${title}`}
        />
      </Tooltip>
    </Space>
  );

  return (
    <Card
      size="small"
      title={title}
      style={style}
      extra={renderCardExtra()}
      styles={{ body: { padding: 0 } }}
    >
      {hasData ? (
        <div
          style={{
            backgroundColor: '#1e1e1e',
            borderRadius: '0 0 6px 6px',
            overflow: 'auto',
            maxHeight: isExpanded ? '600px' : '200px',
          }}
        >
          <SyntaxHighlighter
            language={language}
            style={vscDarkPlus}
            customStyle={{
              margin: 0,
              padding: '12px 16px',
              backgroundColor: '#1e1e1e',
              whiteSpace: wrapText ? 'pre-wrap' : 'pre',
              wordBreak: wrapText ? 'break-word' : 'normal',
              overflowWrap: wrapText ? 'anywhere' : 'normal',
            }}
            codeTagProps={{
              style: {
                whiteSpace: wrapText ? 'pre-wrap' : 'pre',
                wordBreak: wrapText ? 'break-word' : 'normal',
              }
            }}
            wrapLongLines={wrapText}
            showLineNumbers={true}
            lineNumberStyle={{
              minWidth: '3em',
              paddingRight: '1em',
              color: '#636d83',
              textAlign: 'right',
              userSelect: 'none',
            }}
          >
            {jsonString}
          </SyntaxHighlighter>
        </div>
      ) : (
        <div style={{ padding: '16px 24px' }}>
          <Text type="secondary">无数据</Text>
        </div>
      )}
    </Card>
  );
};

export default JsonCardView;
