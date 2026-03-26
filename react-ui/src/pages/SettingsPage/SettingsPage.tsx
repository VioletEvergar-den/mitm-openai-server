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
  Tag
} from 'antd';
import { PlusOutlined, DeleteOutlined, SyncOutlined, CheckCircleOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
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
}

const SettingsPage: React.FC = () => {
  const [form] = Form.useForm();
  const [proxyConfig, setProxyConfig] = useState<ProxyConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [modelMappings, setModelMappings] = useState<ModelMappingItem[]>([]);
  const [newCustomModel, setNewCustomModel] = useState('');
  const [newActualModel, setNewActualModel] = useState('');
  const [updateInfo, setUpdateInfo] = useState<{hasUpdate: boolean, currentVersion: string, latestVersion: string} | null>(null);
  const [checkingUpdate, setCheckingUpdate] = useState(false);
  const [updating, setUpdating] = useState(false);
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
                actualModel: actual
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
      actualModel: newActualModel.trim()
    };
    setModelMappings([...modelMappings, newItem]);
    setNewCustomModel('');
    setNewActualModel('');
  };

  const handleDeleteMapping = (key: string) => {
    setModelMappings(modelMappings.filter(item => item.key !== key));
  };

  const mappingColumns = [
    {
      title: '自定义模型名',
      dataIndex: 'customModel',
      key: 'customModel',
      width: '45%',
    },
    {
      title: '实际模型ID',
      dataIndex: 'actualModel',
      key: 'actualModel',
      width: '45%',
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
      <Title level={2}>系统设置</Title>
      <Paragraph>配置代理服务器以将请求转发到实际的API端点。</Paragraph>
      <Spin spinning={loading}>
        <Card>
          <Form
            form={form}
            layout="vertical"
            onFinish={handleFinish}
            initialValues={proxyConfig || undefined}
          >
            <Form.Item label="代理模式" name="enabled" valuePropName="checked">
              <Switch />
            </Form.Item>
            
            <Paragraph type="secondary">
              {isProxyEnabled ? '代理模式已启用，请求将转发到目标API服务器。' : '代理模式已禁用，服务器将返回模拟数据。'}
            </Paragraph>

            <Form.Item label="目标API服务器地址" name="targetURL">
              <Input placeholder="例如: https://api.openai.com" disabled={!isProxyEnabled} />
            </Form.Item>
            
            <Form.Item label="快捷配置">
              <Space>
                <Button onClick={() => handleQuickConfig('https://api.openai.com')} disabled={!isProxyEnabled}>OpenAI</Button>
                <Button onClick={() => handleQuickConfig('https://api.deepseek.com')} disabled={!isProxyEnabled}>Deepseek</Button>
              </Space>
              <Paragraph type="secondary" style={{ marginTop: 8 }}>
                <Link href="https://platform.openai.com/signup" target="_blank">OpenAI API Key</Link> | <Link href="https://platform.deepseek.com/usage" target="_blank">Deepseek API Key</Link>
              </Paragraph>
            </Form.Item>

            <Form.Item label="认证类型" name="authType">
              <Select disabled={!isProxyEnabled}>
                <Option value="none">无认证</Option>
                <Option value="basic">基本认证 (用户名/密码)</Option>
                <Option value="token">令牌认证 (Bearer Token)</Option>
              </Select>
            </Form.Item>

            <Form.Item noStyle shouldUpdate={(prev, curr) => prev.authType !== curr.authType}>
              {({ getFieldValue }) => {
                const authType = getFieldValue('authType');
                if (authType === 'basic') {
                  return (
                    <Row gutter={16}>
                      <Col span={12}>
                        <Form.Item label="用户名" name="username">
                          <Input placeholder="用户名" disabled={!isProxyEnabled} />
                        </Form.Item>
                      </Col>
                      <Col span={12}>
                        <Form.Item label="密码" name="password">
                          <Input.Password placeholder="密码" disabled={!isProxyEnabled} />
                        </Form.Item>
                      </Col>
                    </Row>
                  );
                }
                if (authType === 'token') {
                  return (
                    <Form.Item label="访问令牌" name="token">
                      <Input.Password placeholder="例如: sk-xxxxxxxx" disabled={!isProxyEnabled} />
                    </Form.Item>
                  );
                }
                return null;
              }}
            </Form.Item>

            <Divider>Model ID 映射</Divider>
            
            <Paragraph type="secondary">
              配置模型名称映射，代理转发时将请求中的模型名替换为实际的模型ID。
              例如：将 <Text code>gpt-5.4</Text> 映射到 <Text code>abab6.5-chat</Text>
            </Paragraph>

            <div style={{ marginBottom: 16 }}>
              <Space.Compact style={{ width: '100%' }}>
                <Input
                  placeholder="自定义模型名 (如: gpt-5.4)"
                  value={newCustomModel}
                  onChange={(e) => setNewCustomModel(e.target.value)}
                  disabled={!isProxyEnabled}
                  style={{ width: '40%' }}
                />
                <Input
                  placeholder="实际模型ID (如: abab6.5-chat)"
                  value={newActualModel}
                  onChange={(e) => setNewActualModel(e.target.value)}
                  disabled={!isProxyEnabled}
                  style={{ width: '40%' }}
                />
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
              <Button type="primary" htmlType="submit" loading={saving}>
                保存配置
              </Button>
            </Form.Item>
          </Form>
        </Card>

        <Card style={{ marginTop: 16 }}>
          <Title level={4}>系统更新</Title>
          <Paragraph type="secondary">
            检查并更新到最新版本。更新后需要重启服务器。
          </Paragraph>
          
          <Space direction="vertical" style={{ width: '100%' }}>
            {updateInfo && (
              <div style={{ marginBottom: 16 }}>
                <Space>
                  <Text strong>当前版本:</Text>
                  <Tag color="blue">{updateInfo.currentVersion}</Tag>
                  <Text strong>最新版本:</Text>
                  <Tag color={updateInfo.hasUpdate ? 'green' : 'blue'}>
                    {updateInfo.latestVersion}
                  </Tag>
                  {updateInfo.hasUpdate ? (
                    <Tag color="orange" icon={<ExclamationCircleOutlined />}>有新版本</Tag>
                  ) : (
                    <Tag color="green" icon={<CheckCircleOutlined />}>已是最新</Tag>
                  )}
                </Space>
              </div>
            )}
            
            <Space>
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
      </Spin>
    </Layout>
  );
};

export default SettingsPage; 