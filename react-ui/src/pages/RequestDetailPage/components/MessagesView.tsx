import React, { useMemo } from 'react';
import { Card, Typography, Tag, Empty, Divider, Space, Alert, Button, Dropdown, message } from 'antd';
import { UserOutlined, RobotOutlined, SettingOutlined, DownloadOutlined, FileTextOutlined, FileMarkdownOutlined, CodeOutlined } from '@ant-design/icons';
import type { MenuProps } from 'antd';

const { Text, Paragraph } = Typography;

interface Message {
  role: string;
  content: string | Array<any>;
}

interface MessagesViewProps {
  requestBody: any;
}

const tryBase64Decode = (str: string): string => {
  if (!str || typeof str !== 'string') {
    return str;
  }
  
  const trimmed = str.trim();
  
  if (!/^[A-Za-z0-9+/=]+$/.test(trimmed) || trimmed.length < 10) {
    return str;
  }
  
  try {
    const binaryString = atob(trimmed);
    const bytes = new Uint8Array(binaryString.length);
    for (let i = 0; i < binaryString.length; i++) {
      bytes[i] = binaryString.charCodeAt(i);
    }
    const decoder = new TextDecoder('utf-8');
    const decoded = decoder.decode(bytes);
    
    if (decoded && decoded.length > 0) {
      return decoded;
    }
  } catch (e) {
    console.log('Base64 decode failed:', e);
  }
  
  return str;
};

const parseRequestBody = (body: any): { model: string; messages: Message[]; rawBody: any } => {
  let parsed = body;

  if (typeof body === 'string') {
    const decodedBody = tryBase64Decode(body);
    try {
      parsed = JSON.parse(decodedBody);
    } catch (e) {
      console.error('Failed to parse request body string:', e);
      return { model: '未知模型', messages: [], rawBody: body };
    }
  }

  if (!parsed || typeof parsed !== 'object') {
    return { model: '未知模型', messages: [], rawBody: body };
  }

  const model = parsed.model || '未知模型';
  const messages = Array.isArray(parsed.messages) ? parsed.messages : [];

  return { model, messages, rawBody: parsed };
};

const getRoleName = (role: string): string => {
  const names: Record<string, string> = {
    user: '用户',
    assistant: '助手',
    system: '系统',
    tool: '工具',
    function: '函数'
  };
  return names[role] || role;
};

const getContentText = (content: string | Array<any>): string => {
  if (typeof content === 'string') {
    return tryBase64Decode(content);
  }

  if (Array.isArray(content)) {
    return content.map((part) => {
      if (part.type === 'text') {
        return tryBase64Decode(part.text || '');
      }
      if (part.type === 'image_url') {
        return `[图片: ${part.image_url?.url?.substring(0, 50) || '未知'}...]`;
      }
      return `[${part.type || '未知类型'}]`;
    }).join('\n');
  }

  return '无法解析的内容格式';
};

const exportAsMarkdown = (model: string, messages: Message[]) => {
  const lines: string[] = [];
  
  lines.push(`# 对话记录`);
  lines.push(``);
  lines.push(`**模型**: ${model}`);
  lines.push(`**消息数量**: ${messages.length}`);
  lines.push(`**导出时间**: ${new Date().toLocaleString('zh-CN')}`);
  lines.push(``);
  lines.push(`---`);
  lines.push(``);

  messages.forEach((msg, index) => {
    const roleName = getRoleName(msg.role);
    const content = getContentText(msg.content);
    
    lines.push(`## ${index + 1}. ${roleName}`);
    lines.push(``);
    lines.push(content);
    lines.push(``);
    lines.push(`---`);
    lines.push(``);
  });

  return lines.join('\n');
};

const exportAsText = (model: string, messages: Message[]) => {
  const lines: string[] = [];
  
  lines.push(`对话记录`);
  lines.push(`模型: ${model}`);
  lines.push(`消息数量: ${messages.length}`);
  lines.push(`导出时间: ${new Date().toLocaleString('zh-CN')}`);
  lines.push(``);
  lines.push(`${'='.repeat(50)}`);
  lines.push(``);

  messages.forEach((msg, index) => {
    const roleName = getRoleName(msg.role);
    const content = getContentText(msg.content);
    
    lines.push(`【${roleName}】`);
    lines.push(content);
    lines.push(``);
    lines.push(`${'-'.repeat(50)}`);
    lines.push(``);
  });

  return lines.join('\n');
};

const exportAsJSON = (model: string, messages: Message[], rawBody: any) => {
  const exportData = {
    model,
    messages: messages.map(msg => ({
      role: msg.role,
      content: getContentText(msg.content)
    })),
    exportedAt: new Date().toISOString(),
    rawBody
  };
  
  return JSON.stringify(exportData, null, 2);
};

const downloadFile = (content: string, filename: string, mimeType: string) => {
  const blob = new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
};

const MessagesView: React.FC<MessagesViewProps> = ({ requestBody }) => {
  const { model, messages, rawBody } = useMemo(() => parseRequestBody(requestBody), [requestBody]);

  const handleExport = (format: 'markdown' | 'text' | 'json') => {
    if (messages.length === 0) {
      message.warning('没有可导出的对话消息');
      return;
    }

    const timestamp = new Date().toISOString().replace(/[:.]/g, '-').substring(0, 19);
    let content: string;
    let filename: string;
    let mimeType: string;

    switch (format) {
      case 'markdown':
        content = exportAsMarkdown(model, messages);
        filename = `对话记录_${timestamp}.md`;
        mimeType = 'text/markdown;charset=utf-8';
        break;
      case 'text':
        content = exportAsText(model, messages);
        filename = `对话记录_${timestamp}.txt`;
        mimeType = 'text/plain;charset=utf-8';
        break;
      case 'json':
        content = exportAsJSON(model, messages, rawBody);
        filename = `对话记录_${timestamp}.json`;
        mimeType = 'application/json;charset=utf-8';
        break;
    }

    downloadFile(content, filename, mimeType);
    message.success(`已导出为 ${format.toUpperCase()} 格式`);
  };

  const exportMenuItems: MenuProps['items'] = [
    {
      key: 'markdown',
      icon: <FileMarkdownOutlined />,
      label: 'Markdown 格式',
      onClick: () => handleExport('markdown')
    },
    {
      key: 'text',
      icon: <FileTextOutlined />,
      label: '纯文本格式',
      onClick: () => handleExport('text')
    },
    {
      key: 'json',
      icon: <CodeOutlined />,
      label: 'JSON 格式',
      onClick: () => handleExport('json')
    }
  ];

  if (!requestBody) {
    return null;
  }

  if (messages.length === 0) {
    return (
      <Card
        size="small"
        title={
          <Space>
            <Text strong>请求体数据</Text>
            <Text type="secondary">(未识别为对话格式)</Text>
          </Space>
        }
        style={{ marginBottom: 16 }}
      >
        <Alert
          message="无法解析为对话消息格式"
          description="请求体不是标准的 OpenAI Chat 格式，请查看下方的原始请求体"
          type="info"
          showIcon
          style={{ marginBottom: 12 }}
        />
        <Paragraph
          style={{
            margin: 0,
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
            backgroundColor: '#f5f5f5',
            padding: 12,
            borderRadius: 4,
            fontFamily: 'monospace',
            fontSize: 12
          }}
        >
          {typeof rawBody === 'object' ? JSON.stringify(rawBody, null, 2) : String(rawBody)}
        </Paragraph>
      </Card>
    );
  }

  const getRoleIcon = (role: string) => {
    switch (role) {
      case 'user':
        return <UserOutlined style={{ color: '#1890ff' }} />;
      case 'assistant':
        return <RobotOutlined style={{ color: '#52c41a' }} />;
      case 'system':
        return <SettingOutlined style={{ color: '#faad14' }} />;
      default:
        return null;
    }
  };

  const getRoleTag = (role: string) => {
    const colors: Record<string, string> = {
      user: 'blue',
      assistant: 'green',
      system: 'gold',
      tool: 'purple',
      function: 'cyan'
    };
    return <Tag color={colors[role] || 'default'}>{role}</Tag>;
  };

  const renderContent = (content: string | Array<any>) => {
    if (typeof content === 'string') {
      const decodedContent = tryBase64Decode(content);
      return (
        <Paragraph
          style={{
            margin: 0,
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
            fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif'
          }}
        >
          {decodedContent || <Text type="secondary">（空内容）</Text>}
        </Paragraph>
      );
    }

    if (Array.isArray(content)) {
      return (
        <Space direction="vertical" style={{ width: '100%' }}>
          {content.map((part, idx) => {
            if (part.type === 'text') {
              const decodedText = tryBase64Decode(part.text || '');
              return (
                <Paragraph
                  key={idx}
                  style={{
                    margin: 0,
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-word'
                  }}
                >
                  {decodedText}
                </Paragraph>
              );
            }
            if (part.type === 'image_url') {
              return (
                <div key={idx} style={{ marginTop: 8 }}>
                  <Tag color="purple">图片</Tag>
                  {part.image_url?.url && (
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {part.image_url.url.substring(0, 50)}...
                    </Text>
                  )}
                </div>
              );
            }
            return (
              <Tag key={idx} color="default">
                {part.type || '未知类型'}
              </Tag>
            );
          })}
        </Space>
      );
    }

    return <Text type="secondary">无法解析的内容格式</Text>;
  };

  return (
    <Card
      size="small"
      title={
        <Space>
          <Text strong>对话消息</Text>
          <Tag color="blue">{model}</Tag>
          <Text type="secondary">({messages.length} 条消息)</Text>
        </Space>
      }
      extra={
        <Dropdown menu={{ items: exportMenuItems }} placement="bottomRight">
          <Button type="primary" icon={<DownloadOutlined />} size="small">
            导出对话
          </Button>
        </Dropdown>
      }
      style={{ marginBottom: 16 }}
    >
      {messages.map((msg, index) => (
        <div key={index}>
          <div style={{ display: 'flex', alignItems: 'flex-start', marginBottom: 8 }}>
            <div style={{ marginRight: 12, marginTop: 4 }}>
              {getRoleIcon(msg.role)}
            </div>
            <div style={{ flex: 1 }}>
              <div style={{ marginBottom: 4 }}>
                {getRoleTag(msg.role)}
              </div>
              {renderContent(msg.content)}
            </div>
          </div>
          {index < messages.length - 1 && <Divider style={{ margin: '12px 0' }} />}
        </div>
      ))}
    </Card>
  );
};

export default MessagesView;
