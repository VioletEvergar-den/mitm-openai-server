import React, { useState, useEffect } from 'react';
import {
  Form,
  Input,
  Button,
  Switch,
  Select,
  Card,
  Typography,
  Spin,
  Space,
  Row,
  Col,
  Divider,
  Table,
  Popconfirm,
  Modal,
  Tag,
  Alert,
  Tooltip
} from 'antd';
import { 
  PlusOutlined, 
  DeleteOutlined, 
  SyncOutlined, 
  CheckCircleOutlined, 
  ExclamationCircleOutlined,
  SettingOutlined,
  CloudServerOutlined,
  ApiOutlined,
  SafetyOutlined,
  InfoCircleOutlined,
  ThunderboltOutlined
} from '@ant-design/icons';
import Layout from '../../components/Layout/Layout';
import { ProxyConfig } from '../../types';
import { apiService } from '../../services/api';
import { useNotification } from '../../components/Notification/Notification';

const { Title, Paragraph, Text, Link } = Typography;
const { Option } = Select;

interface ModelMappingItem {
  key: string;
  customModel: string;
  actualModel: string;
  provider?: string;
}

interface AIProvider {
  name: string;
  baseUrl: string;
  description: string;
  models: { customModel: string; actualModel: string; description: string }[];
}

const AI_PROVIDERS: AIProvider[] = [
  {
    name: 'OpenAI',
    baseUrl: 'https://api.openai.com/v1',
    description: 'OpenAI 官方 API',
    models: [
      { customModel: 'gpt-4o', actualModel: 'gpt-4o', description: 'GPT-4 Optimized' },
      { customModel: 'gpt-4-turbo', actualModel: 'gpt-4-turbo', description: 'GPT-4 Turbo' },
      { customModel: 'gpt-4', actualModel: 'gpt-4', description: 'GPT-4' },
      { customModel: 'gpt-3.5-turbo', actualModel: 'gpt-3.5-turbo', description: 'GPT-3.5 Turbo' },
    ]
  },
  {
    name: '飞牛云',
    baseUrl: 'https://api.qnaigc.com/v1',
    description: '飞牛云 AI 服务',
    models: [
      { customModel: 'gpt-4o', actualModel: 'gpt-4o', description: 'GPT-4o' },
      { customModel: 'gpt-4-turbo', actualModel: 'gpt-4-turbo', description: 'GPT-4 Turbo' },
      { customModel: 'claude-3-opus', actualModel: 'claude-3-opus-20240229', description: 'Claude 3 Opus' },
      { customModel: 'claude-3-sonnet', actualModel: 'claude-3-sonnet-20240229', description: 'Claude 3 Sonnet' },
    ]
  },
  {
    name: '硅基流动',
    baseUrl: 'https://api.siliconflow.cn/v1',
    description: '硅基流动 AI 服务',
    models: [
      { customModel: 'qwen-turbo', actualModel: 'Qwen/Qwen2.5-7B-Instruct', description: 'Qwen 2.5 7B' },
      { customModel: 'qwen-plus', actualModel: 'Qwen/Qwen2.5-72B-Instruct', description: 'Qwen 2.5 72B' },
      { customModel: 'deepseek-v3', actualModel: 'deepseek-ai/DeepSeek-V3', description: 'DeepSeek V3' },
      { customModel: 'yi-large', actualModel: '01-ai/Yi-1.5-34B-Chat', description: 'Yi 1.5 34B' },
    ]
  },
  {
    name: 'DeepSeek',
    baseUrl: 'https://api.deepseek.com',
    description: 'DeepSeek 官方 API',
    models: [
      { customModel: 'deepseek-chat', actualModel: 'deepseek-chat', description: 'DeepSeek Chat' },
      { customModel: 'deepseek-coder', actualModel: 'deepseek-coder', description: 'DeepSeek Coder' },
    ]
  },
  {
    name: '智谱AI',
    baseUrl: 'https://open.bigmodel.cn/api/paas/v4',
    description: '智谱 AI GLM 系列',
    models: [
      { customModel: 'glm-4', actualModel: 'glm-4', description: 'GLM-4' },
      { customModel: 'glm-4-flash', actualModel: 'glm-4-flash', description: 'GLM-4 Flash' },
      { customModel: 'glm-3-turbo', actualModel: 'glm-3-turbo', description: 'GLM-3 Turbo' },
    ]
  },
  {
    name: '月之暗面',
    baseUrl: 'https://api.moonshot.cn/v1',
    description: 'Moonshot Kimi 系列',
    models: [
      { customModel: 'moonshot-v1-8k', actualModel: 'moonshot-v1-8k', description: 'Moonshot V1 8K' },
      { customModel: 'moonshot-v1-32k', actualModel: 'moonshot-v1-32k', description: 'Moonshot V1 32K' },
      { customModel: 'moonshot-v1-128k', actualModel: 'moonshot-v1-128k', description: 'Moonshot V1 128K' },
    ]
  },
  {
    name: '阿里云百炼',
    baseUrl: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
    description: '阿里云通义千问系列',
    models: [
      { customModel: 'qwen-turbo', actualModel: 'qwen-turbo', description: '通义千问 Turbo' },
      { customModel: 'qwen-plus', actualModel: 'qwen-plus', description: '通义千问 Plus' },
      { customModel: 'qwen-max', actualModel: 'qwen-max', description: '通义千问 Max' },
    ]
  },
  {
    name: 'MiniMax',
    baseUrl: 'https://api.minimax.chat/v1',
    description: 'MiniMax AI 服务',
    models: [
      { customModel: 'MiniMax-M2.5', actualModel: 'minimax/minimax-m2.5', description: 'MiniMax M2.5' },
      { customModel: 'abab6.5-chat', actualModel: 'abab6.5-chat', description: 'ABAB 6.5 Chat' },
      { customModel: 'abab5.5-chat', actualModel: 'abab5.5-chat', description: 'ABAB 5.5 Chat' },
    ]
  },
  {
    name: 'Anthropic',
    baseUrl: 'https://api.anthropic.com/v1',
    description: 'Anthropic Claude 系列',
    models: [
      { customModel: 'claude-3-opus', actualModel: 'claude-3-opus-20240229', description: 'Claude 3 Opus' },
      { customModel: 'claude-3-sonnet', actualModel: 'claude-3-sonnet-20240229', description: 'Claude 3 Sonnet' },
      { customModel: 'claude-3-haiku', actualModel: 'claude-3-haiku-20240307', description: 'Claude 3 Haiku' },
    ]
  },
  {
    name: 'Google AI',
    baseUrl: 'https://generativelanguage.googleapis.com/v1beta',
    description: 'Google Gemini 系列',
    models: [
      { customModel: 'gemini-pro', actualModel: 'gemini-pro', description: 'Gemini Pro' },
      { customModel: 'gemini-1.5-pro', actualModel: 'gemini-1.5-pro', description: 'Gemini 1.5 Pro' },
      { customModel: 'gemini-1.5-flash', actualModel: 'gemini-1.5-flash', description: 'Gemini 1.5 Flash' },
    ]
  },
];

const SettingsPage: React.FC = () => {
  const [form] = Form.useForm();
  const [proxyConfig, setProxyConfig] = useState<ProxyConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [modelMappings, setModelMappings] = useState<ModelMappingItem[]>([]);
  const [newCustomModel, setNewCustomModel] = useState('');
  const [newActualModel, setNewActualModel] = useState('');
  const [newProvider, setNewProvider] = useState('');
  const [updateInfo, setUpdateInfo] = useState<{hasUpdate: boolean, currentVersion: string, latestVersion: string} | null>(null);
  const [checkingUpdate, setCheckingUpdate] = useState(false);
  const [updating, setUpdating] = useState(false);
  const [selectedProvider, setSelectedProvider] = useState<string>('');
  const { addNotification } = useNotification();
  
  const isProxyEnabled = Form.useWatch('enabled', form);

  useEffect(() => {
    const loadProxyConfig = async () => {
      setLoading(true);
      try {
        const config = await apiService.getProxyConfig();
        if (config) {
          setProxyConfig(config);
          form.setFieldsValue(config);
          if (config.modelMapping) {
            const mappings: ModelMappingItem[] = Object.entries(config.modelMapping).map(
              ([custom, actual], index) => ({
                key: `mapping-${index}`,
                customModel: custom,
                actualModel: actual,
                provider: ''
              })
            );
            setModelMappings(mappings);
          }
        }
      } catch (error) {
        addNotification('加载代理配置失败', 'danger');
      } finally {
        setLoading(false);
      }
    };
    loadProxyConfig();
  }, [form, addNotification]);

  const handleFinish = async (values: ProxyConfig) => {
    setSaving(true);
    try {
      const mappingRecord: Record<string, string> = {};
      modelMappings.forEach(item => {
        if (item.customModel && item.actualModel) {
          mappingRecord[item.customModel] = item.actualModel;
        }
      });
      const configToSave = {
        ...values,
        modelMapping: mappingRecord
      };
      const success = await apiService.saveProxyConfig(configToSave);
      if (success) {
        addNotification('代理设置已保存', 'success');
        setProxyConfig(configToSave);
      } else {
        addNotification('保存代理设置失败', 'danger');
      }
    } catch (error) {
      addNotification('保存代理设置失败', 'danger');
    } finally {
      setSaving(false);
    }
  };

  const handleProviderSelect = (providerName: string) => {
    const provider = AI_PROVIDERS.find(p => p.name === providerName);
    if (provider) {
      form.setFieldsValue({ targetURL: provider.baseUrl });
      setSelectedProvider(providerName);
      addNotification(`已选择 ${provider.name}，地址: ${provider.baseUrl}`, 'success');
    }
  };

  const handleQuickConfig = (url: string) => {
    form.setFieldsValue({ targetURL: url });
  };

  const handleAddMapping = () => {
    if (!newCustomModel.trim() || !newActualModel.trim()) {
      addNotification('请填写完整的模型映射信息', 'danger');
      return;
    }
    const newItem: ModelMappingItem = {
      key: `mapping-${Date.now()}`,
      customModel: newCustomModel.trim(),
      actualModel: newActualModel.trim(),
      provider: newProvider.trim()
    };
    setModelMappings([...modelMappings, newItem]);
    setNewCustomModel('');
    setNewActualModel('');
    setNewProvider('');
  };

  const handleQuickAddMappings = (provider: AIProvider) => {
    const newMappings: ModelMappingItem[] = provider.models.map(model => ({
      key: `mapping-${Date.now()}-${model.customModel}`,
      customModel: model.customModel,
      actualModel: model.actualModel,
      provider: provider.name
    }));
    
    const existingCustomModels = new Set(modelMappings.map(m => m.customModel));
    const uniqueNewMappings = newMappings.filter(m => !existingCustomModels.has(m.customModel));
    
    if (uniqueNewMappings.length === 0) {
      addNotification('所选服务商的模型映射已全部存在', 'info');
      return;
    }
    
    setModelMappings([...modelMappings, ...uniqueNewMappings]);
    addNotification(`已添加 ${provider.name} 的 ${uniqueNewMappings.length} 个模型映射`, 'success');
  };

  const handleDeleteMapping = (key: string) => {
    setModelMappings(modelMappings.filter(item => item.key !== key));
  };

  const getProviderColor = (providerName: string): string => {
    const colors: Record<string, string> = {
      'OpenAI': 'green',
      '飞牛云': 'blue',
      '硅基流动': 'purple',
      'DeepSeek': 'cyan',
      '智谱AI': 'orange',
      '月之暗面': 'magenta',
      '阿里云百炼': 'gold',
      'MiniMax': 'lime',
      'Anthropic': 'volcano',
      'Google AI': 'red',
    };
    return colors[providerName] || 'default';
  };

  const mappingColumns = [
    {
      title: '自定义模型名',
      dataIndex: 'customModel',
      key: 'customModel',
      width: '35%',
    },
    {
      title: '实际模型ID',
      dataIndex: 'actualModel',
      key: 'actualModel',
      width: '40%',
    },
    {
      title: '服务商',
      dataIndex: 'provider',
      key: 'provider',
      width: '15%',
      render: (provider: string) => provider ? (
        <Tag color={getProviderColor(provider)}>{provider}</Tag>
      ) : (
        <Text type="secondary">-</Text>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: '10%',
      render: (_: any, record: ModelMappingItem) => (
        <Popconfirm
          title="确定删除此映射?"
          onConfirm={() => handleDeleteMapping(record.key)}
          okText="确定"
          cancelText="取消"
        >
          <Button type="text" danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];

  const handleCheckUpdate = async () => {
    setCheckingUpdate(true);
    try {
      const info = await apiService.checkUpdate();
      if (info) {
        setUpdateInfo(info);
        if (info.hasUpdate) {
          addNotification(`发现新版本 ${info.latestVersion}`, 'success');
        } else {
          addNotification('当前已是最新版本', 'success');
        }
      } else {
        addNotification('检查更新失败', 'danger');
      }
    } catch (error) {
      addNotification('检查更新失败', 'danger');
    } finally {
      setCheckingUpdate(false);
    }
  };

  const handlePerformUpdate = async (useGit = false) => {
    Modal.confirm({
      title: '确认更新',
      content: useGit 
        ? '将通过 Git 拉取最新代码并重新编译，确定继续吗？' 
        : '将下载最新版本并替换当前程序，更新后需要重启服务器。确定继续吗？',
      okText: '确定',
      cancelText: '取消',
      onOk: async () => {
        setUpdating(true);
        try {
          const result = await apiService.performUpdate(useGit);
          if (result.success) {
            addNotification(result.message, 'success');
            Modal.success({
              title: '更新成功',
              content: '请重启服务器以使用新版本。',
            });
          } else {
            addNotification(result.message, 'danger');
          }
        } catch (error) {
          addNotification('更新失败', 'danger');
        } finally {
          setUpdating(false);
        }
      }
    });
  };

  return (
    <Layout title="系统设置">
      <Spin spinning={loading}>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <Title level={3} style={{ margin: 0 }}>
            <SettingOutlined style={{ marginRight: 8 }} />
            系统设置
          </Title>

          <Card 
            title={
              <Space>
                <CloudServerOutlined />
                <span>代理配置</span>
              </Space>
            }
          >
            <Form
              form={form}
              layout="vertical"
              onFinish={handleFinish}
              initialValues={proxyConfig || undefined}
            >
              <Row gutter={24}>
                <Col span={24}>
                  <Form.Item label="代理模式" name="enabled" valuePropName="checked">
                    <Switch checkedChildren="启用" unCheckedChildren="禁用" />
                  </Form.Item>
                  
                  <Alert
                    type={isProxyEnabled ? 'success' : 'info'}
                    message={isProxyEnabled ? '代理模式已启用，请求将转发到目标API服务器' : '代理模式已禁用，服务器将返回模拟数据'}
                    showIcon
                    style={{ marginBottom: 16 }}
                  />
                </Col>

                <Col xs={24}>
                  <Form.Item label={
                    <Space>
                      <span>选择服务商</span>
                      <Tooltip title="选择服务商后，将自动填充对应的 API 地址">
                        <InfoCircleOutlined style={{ color: '#999' }} />
                      </Tooltip>
                    </Space>
                  }>
                    <Select
                      placeholder="选择 AI 服务商快速配置"
                      onChange={handleProviderSelect}
                      disabled={!isProxyEnabled}
                      allowClear
                      value={selectedProvider || undefined}
                    >
                      {AI_PROVIDERS.map(provider => (
                        <Option key={provider.name} value={provider.name}>
                          <Space>
                            <Tag color={getProviderColor(provider.name)}>{provider.name}</Tag>
                            <Text type="secondary" style={{ fontSize: 12 }}>{provider.description}</Text>
                          </Space>
                        </Option>
                      ))}
                    </Select>
                  </Form.Item>
                </Col>

                <Col xs={24} md={16}>
                  <Form.Item label="目标API服务器地址" name="targetURL">
                    <Input 
                      placeholder="例如: https://api.openai.com/v1" 
                      disabled={!isProxyEnabled}
                      prefix={<ApiOutlined />}
                    />
                  </Form.Item>
                </Col>

                <Col xs={24} md={8}>
                  <Form.Item label="常用地址">
                    <Space wrap>
                      {AI_PROVIDERS.slice(0, 4).map(provider => (
                        <Tooltip key={provider.name} title={provider.baseUrl}>
                          <Button 
                            size="small"
                            onClick={() => handleQuickConfig(provider.baseUrl)} 
                            disabled={!isProxyEnabled}
                          >
                            {provider.name}
                          </Button>
                        </Tooltip>
                      ))}
                    </Space>
                  </Form.Item>
                </Col>

                <Col xs={24} md={8}>
                  <Form.Item label="认证类型" name="authType">
                    <Select disabled={!isProxyEnabled}>
                      <Option value="none">无认证</Option>
                      <Option value="basic">基本认证</Option>
                      <Option value="token">令牌认证</Option>
                    </Select>
                  </Form.Item>
                </Col>

                <Form.Item noStyle shouldUpdate={(prev, curr) => prev.authType !== curr.authType}>
                  {({ getFieldValue }) => {
                    const authType = getFieldValue('authType');
                    if (authType === 'basic') {
                      return (
                        <>
                          <Col xs={24} md={8}>
                            <Form.Item label="用户名" name="username">
                              <Input placeholder="用户名" disabled={!isProxyEnabled} />
                            </Form.Item>
                          </Col>
                          <Col xs={24} md={8}>
                            <Form.Item label="密码" name="password">
                              <Input.Password placeholder="密码" disabled={!isProxyEnabled} />
                            </Form.Item>
                          </Col>
                        </>
                      );
                    }
                    if (authType === 'token') {
                      return (
                        <Col xs={24} md={16}>
                          <Form.Item label="访问令牌" name="token">
                            <Input.Password 
                              placeholder="例如: sk-xxxxxxxx" 
                              disabled={!isProxyEnabled}
                              prefix={<SafetyOutlined />}
                            />
                          </Form.Item>
                        </Col>
                      );
                    }
                    return null;
                  }}
                </Form.Item>
              </Row>

              <Divider orientation="left">
                <Space>
                  <ApiOutlined />
                  Model ID 映射
                </Space>
              </Divider>
              
              <Paragraph type="secondary">
                配置模型名称映射，代理转发时将请求中的模型名替换为实际的模型ID。
                例如：将 <Text code>gpt-5.4</Text> 映射到 <Text code>abab6.5-chat</Text>
              </Paragraph>

              <Card size="small" style={{ marginBottom: 16, background: '#fafafa' }}>
                <Space direction="vertical" style={{ width: '100%' }}>
                  <Text strong>
                    <ThunderboltOutlined style={{ marginRight: 4, color: '#1890ff' }} />
                    快捷添加模型映射
                  </Text>
                  <Space wrap>
                    {AI_PROVIDERS.map(provider => (
                      <Tooltip 
                        key={provider.name} 
                        title={
                          <div>
                            <div><strong>{provider.name}</strong> - {provider.description}</div>
                            <div style={{ marginTop: 4 }}>
                              {provider.models.slice(0, 3).map(m => (
                                <div key={m.customModel}>
                                  <Text code style={{ color: '#fff' }}>{m.customModel}</Text>
                                  {' → '}
                                  <Text code style={{ color: '#fff' }}>{m.actualModel}</Text>
                                </div>
                              ))}
                              {provider.models.length > 3 && <Text style={{ color: '#fff' }}>...</Text>}
                            </div>
                          </div>
                        }
                      >
                        <Button 
                          size="small"
                          onClick={() => handleQuickAddMappings(provider)}
                          disabled={!isProxyEnabled}
                        >
                          <Tag color={getProviderColor(provider.name)} style={{ margin: 0 }}>
                            {provider.name}
                          </Tag>
                        </Button>
                      </Tooltip>
                    ))}
                  </Space>
                </Space>
              </Card>

              <div style={{ marginBottom: 16 }}>
                <Space.Compact style={{ width: '100%' }}>
                  <Input
                    placeholder="自定义模型名 (如: gpt-5.4)"
                    value={newCustomModel}
                    onChange={(e) => setNewCustomModel(e.target.value)}
                    disabled={!isProxyEnabled}
                    style={{ width: '30%' }}
                  />
                  <Input
                    placeholder="实际模型ID (如: abab6.5-chat)"
                    value={newActualModel}
                    onChange={(e) => setNewActualModel(e.target.value)}
                    disabled={!isProxyEnabled}
                    style={{ width: '35%' }}
                  />
                  <Select
                    placeholder="服务商 (可选)"
                    value={newProvider || undefined}
                    onChange={setNewProvider}
                    disabled={!isProxyEnabled}
                    style={{ width: '20%' }}
                    allowClear
                  >
                    {AI_PROVIDERS.map(provider => (
                      <Option key={provider.name} value={provider.name}>
                        <Tag color={getProviderColor(provider.name)} style={{ margin: 0 }}>
                          {provider.name}
                        </Tag>
                      </Option>
                    ))}
                  </Select>
                  <Button
                    type="primary"
                    icon={<PlusOutlined />}
                    onClick={handleAddMapping}
                    disabled={!isProxyEnabled}
                  >
                    添加
                  </Button>
                </Space.Compact>
              </div>

              {modelMappings.length > 0 && (
                <Table
                  dataSource={modelMappings}
                  columns={mappingColumns}
                  rowKey="key"
                  pagination={false}
                  size="small"
                  bordered
                />
              )}
              
              <Divider />
              
              <Form.Item>
                <Button type="primary" htmlType="submit" loading={saving} size="large">
                  保存配置
                </Button>
              </Form.Item>
            </Form>
          </Card>

          <Card 
            title={
              <Space>
                <SyncOutlined spin={checkingUpdate} />
                <span>系统更新</span>
              </Space>
            }
          >
            <Paragraph type="secondary">
              检查并更新到最新版本。更新后需要重启服务器。
            </Paragraph>
            
            <Space direction="vertical" style={{ width: '100%' }} size="large">
              {updateInfo && (
                <Card size="small" bordered={false} style={{ background: 'transparent' }}>
                  <Space wrap size="large">
                    <div>
                      <Text type="secondary">当前版本:</Text>
                      <Tag color="blue" style={{ marginLeft: 8 }}>{updateInfo.currentVersion}</Tag>
                    </div>
                    <div>
                      <Text type="secondary">最新版本:</Text>
                      <Tag color={updateInfo.hasUpdate ? 'green' : 'blue'} style={{ marginLeft: 8 }}>
                        {updateInfo.latestVersion}
                      </Tag>
                    </div>
                    {updateInfo.hasUpdate ? (
                      <Tag color="orange" icon={<ExclamationCircleOutlined />}>有新版本可用</Tag>
                    ) : (
                      <Tag color="green" icon={<CheckCircleOutlined />}>已是最新版本</Tag>
                    )}
                  </Space>
                </Card>
              )}
              
              <Space wrap>
                <Button
                  icon={<SyncOutlined spin={checkingUpdate} />}
                  onClick={handleCheckUpdate}
                  loading={checkingUpdate}
                >
                  检查更新
                </Button>
                
                {updateInfo?.hasUpdate && (
                  <>
                    <Button
                      type="primary"
                      onClick={() => handlePerformUpdate(false)}
                      loading={updating}
                    >
                      下载更新
                    </Button>
                    <Button
                      onClick={() => handlePerformUpdate(true)}
                      loading={updating}
                    >
                      Git 更新
                    </Button>
                  </>
                )}
              </Space>
            </Space>
          </Card>
        </Space>
      </Spin>
    </Layout>
  );
};

export default SettingsPage;
