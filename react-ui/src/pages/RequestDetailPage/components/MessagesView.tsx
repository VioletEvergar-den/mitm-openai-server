import React, { useMemo } from 'react';
import { Card, Typography, Tag, Empty, Divider, Space } from 'antd';
import { UserOutlined, RobotOutlined, SettingOutlined } from '@ant-design/icons';

const { Text, Paragraph } = Typography;

interface Message {
  role: string;
  content: string | Array<any>;
}

interface MessagesViewProps {
  requestBody: any;
}

const parseRequestBody = (body: any): { model: string; messages: Message[] } => {
  let parsed = body;

  if (typeof body === 'string') {
    try {
      parsed = JSON.parse(body);
    } catch (e) {
      console.error('Failed to parse request body string:', e);
      return { model: '未知模型', messages: [] };
    }
  }

  if (!parsed || typeof parsed !== 'object') {
    return { model: '未知模型', messages: [] };
  }

  const model = parsed.model || '未知模型';
  const messages = Array.isArray(parsed.messages) ? parsed.messages : [];

  return { model, messages };
};

const MessagesView: React.FC<MessagesViewProps> = ({ requestBody }) => {
  const { model, messages } = useMemo(() => parseRequestBody(requestBody), [requestBody]);

  if (messages.length === 0) {
    return null;
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
      return (
        <Paragraph
          style={{
            margin: 0,
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
            fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif'
          }}
        >
          {content || <Text type="secondary">（空内容）</Text>}
        </Paragraph>
      );
    }

    if (Array.isArray(content)) {
      return (
        <Space direction="vertical" style={{ width: '100%' }}>
          {content.map((part, idx) => {
            if (part.type === 'text') {
              return (
                <Paragraph
                  key={idx}
                  style={{
                    margin: 0,
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-word'
                  }}
                >
                  {part.text}
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
