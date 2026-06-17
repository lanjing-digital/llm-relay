import {
  Button,
  Card,
  ConfigProvider,
  Layout,
  message,
  Popconfirm,
  Space,
  Table,
  Tag,
  Typography,
} from 'antd'
import { DeleteOutlined, PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import zhCN from 'antd/locale/zh_CN'
import { useCallback, useEffect, useState } from 'react'
import type { ColumnsType } from 'antd/es/table'
import './App.css'
import ConfigFormModal from './components/config-form-modal'
import {
  createConfig,
  deleteConfig,
  fetchConfigs,
  testConfig,
  updateConfig,
} from './api/configs'
import type { Config, ConfigPayload } from './types/config'

function App() {
  const { Title, Paragraph, Text } = Typography
  const [configs, setConfigs] = useState<Config[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [editingConfig, setEditingConfig] = useState<Config | null>(null)
  const [testingId, setTestingId] = useState<number | null>(null)
  const [deletingId, setDeletingId] = useState<number | null>(null)
  const [messageApi, contextHolder] = message.useMessage()

  const loadConfigs = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchConfigs()
      setConfigs(data)
    } catch (error) {
      const description =
        error instanceof Error ? error.message : '加载配置列表失败'
      void messageApi.error(description)
    } finally {
      setLoading(false)
    }
  }, [messageApi])

  useEffect(() => {
    void Promise.resolve().then(loadConfigs)
  }, [loadConfigs])

  const handleCreate = () => {
    setEditingConfig(null)
    setModalOpen(true)
  }

  const handleEdit = (record: Config) => {
    setEditingConfig(record)
    setModalOpen(true)
  }

  const handleSubmit = async (values: ConfigPayload) => {
    setSubmitting(true)
    try {
      if (editingConfig) {
        await updateConfig(editingConfig.id, values)
        void messageApi.success('配置已更新')
      } else {
        await createConfig(values)
        void messageApi.success('配置已创建')
      }
      setModalOpen(false)
      setEditingConfig(null)
      await loadConfigs()
    } catch (error) {
      const description =
        error instanceof Error ? error.message : '保存配置失败'
      void messageApi.error(description)
    } finally {
      setSubmitting(false)
    }
  }

  const handleDelete = async (record: Config) => {
    setDeletingId(record.id)
    try {
      await deleteConfig(record.id)
      void messageApi.success(`已删除配置「${record.name}」`)
      await loadConfigs()
    } catch (error) {
      const description =
        error instanceof Error ? error.message : '删除配置失败'
      void messageApi.error(description)
    } finally {
      setDeletingId(null)
    }
  }

  const handleTest = async (record: Config) => {
    setTestingId(record.id)
    try {
      const result = await testConfig(record.id)
      const content = result.message ?? result.detail ?? '测试完成'
      void messageApi.success(`配置「${record.name}」测试成功：${content}`)
    } catch (error) {
      const description =
        error instanceof Error ? error.message : '连通性测试失败'
      void messageApi.error(`配置「${record.name}」测试失败：${description}`)
    } finally {
      setTestingId(null)
    }
  }

  const columns: ColumnsType<Config> = [
    {
      title: '配置名称',
      dataIndex: 'name',
      key: 'name',
      render: (_, record) => (
        <Space direction="vertical" size={2}>
          <Text strong>{record.name}</Text>
          <Text type="secondary">{record.external_model}</Text>
        </Space>
      ),
    },
    {
      title: '目标模型',
      dataIndex: 'target_model',
      key: 'target_model',
      render: (value: string) => <Tag color="blue">{value}</Tag>,
    },
    {
      title: '目标地址',
      dataIndex: 'target_base_url',
      key: 'target_base_url',
      ellipsis: true,
      render: (value: string) => (
        <Text copyable={{ text: value }} ellipsis={{ tooltip: value }}>
          {value}
        </Text>
      ),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 180,
      render: (value: string) =>
        value ? new Date(value).toLocaleString('zh-CN') : '-',
    },
    {
      title: '操作',
      key: 'actions',
      width: 280,
      render: (_, record) => (
        <Space wrap>
          <Button type="link" onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Button
            type="link"
            loading={testingId === record.id}
            onClick={() => void handleTest(record)}
          >
            连通性测试
          </Button>
          <Popconfirm
            title="删除配置"
            description={`确定删除「${record.name}」吗？`}
            okText="删除"
            cancelText="取消"
            okButtonProps={{ danger: true, loading: deletingId === record.id }}
            onConfirm={() => void handleDelete(record)}
          >
            <Button danger type="link" icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <ConfigProvider locale={zhCN}>
      {contextHolder}
      <Layout className="app-shell">
        <Layout.Header className="app-header">
          <div>
            <Title level={3}>LLM Relay 管理后台</Title>
            <Paragraph>
              管理外部模型到真实目标模型与 API 线路的映射配置。
            </Paragraph>
          </div>
          <Space>
            <Button
              icon={<ReloadOutlined />}
              onClick={() => void loadConfigs()}
              loading={loading}
            >
              刷新
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
              新建配置
            </Button>
          </Space>
        </Layout.Header>

        <Layout.Content className="app-content">
          <Card className="content-card" bordered={false}>
            <div className="card-heading">
              <div>
                <Title level={4}>配置列表</Title>
                <Paragraph>
                  当前共 {configs.length} 条配置，调用方请求中的 `model` 将匹配到
                  `external_model`。
                </Paragraph>
              </div>
            </div>

            <Table<Config>
              rowKey="id"
              columns={columns}
              dataSource={configs}
              loading={loading}
              pagination={{ pageSize: 8, showSizeChanger: false }}
              scroll={{ x: 960 }}
            />
          </Card>
        </Layout.Content>
      </Layout>

      <ConfigFormModal
        open={modalOpen}
        confirmLoading={submitting}
        initialValues={editingConfig}
        onCancel={() => {
          setModalOpen(false)
          setEditingConfig(null)
        }}
        onSubmit={handleSubmit}
      />
    </ConfigProvider>
  )
}

export default App
