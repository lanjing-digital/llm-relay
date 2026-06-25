import {
  Button,
  Card,
  ConfigProvider,
  Descriptions,
  Drawer,
  Layout,
  message,
  Popconfirm,
  Space,
  Table,
  Tag,
  Tabs,
  Typography,
} from 'antd'
import { DeleteOutlined, PlusOutlined, ReloadOutlined, LogoutOutlined } from '@ant-design/icons'
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
import { fetchLogDetail, fetchLogs } from './api/logs'
import { clearToken, getToken, setToken, setOnUnauthorized } from './api/client'
import type { Config, ConfigPayload } from './types/config'
import type { LogDetail, LogListItem } from './types/log'
import LoginPage from './pages/login'

function App() {
  const { Title, Paragraph, Text } = Typography
  const [authed, setAuthed] = useState(() => getToken() != null)
  const [activeTabKey, setActiveTabKey] = useState<'configs' | 'logs'>('configs')
  const [configs, setConfigs] = useState<Config[]>([])
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [editingConfig, setEditingConfig] = useState<Config | null>(null)
  const [testingId, setTestingId] = useState<number | null>(null)
  const [deletingId, setDeletingId] = useState<number | null>(null)
  const [messageApi, contextHolder] = message.useMessage()

  const [logs, setLogs] = useState<LogListItem[]>([])
  const [logsLoading, setLogsLoading] = useState(false)
  const [logsPage, setLogsPage] = useState(1)
  const [logsPageSize, setLogsPageSize] = useState(20)
  const [logsTotal, setLogsTotal] = useState(0)
  const [logDetailOpen, setLogDetailOpen] = useState(false)
  const [logDetailLoading, setLogDetailLoading] = useState(false)
  const [selectedLog, setSelectedLog] = useState<LogDetail | null>(null)

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

  useEffect(() => {
    setOnUnauthorized(() => {
      clearToken()
      setAuthed(false)
    })
  }, [])

  const loadLogs = useCallback(
    async (page: number, pageSize: number) => {
      setLogsLoading(true)
      try {
        const data = await fetchLogs({ page, pageSize })
        setLogs(data.items)
        setLogsTotal(data.total)
        setLogsPage(page)
        setLogsPageSize(pageSize)
      } catch (error) {
        const description =
          error instanceof Error ? error.message : '加载调用日志失败'
        void messageApi.error(description)
      } finally {
        setLogsLoading(false)
      }
    },
    [messageApi],
  )

  useEffect(() => {
    if (activeTabKey !== 'logs') {
      return
    }
    if (logs.length > 0) {
      return
    }
    void Promise.resolve().then(() => loadLogs(logsPage, logsPageSize))
  }, [activeTabKey, loadLogs, logs.length, logsPage, logsPageSize])

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

  const redactSecrets = (value?: string | null) => {
    if (!value) {
      return '-'
    }

    return value
      .replace(/Authorization:\s*Bearer\s+([^\s"'\\]+)/gi, 'Authorization: Bearer [REDACTED]')
      .replace(/Bearer\s+([^\s"'\\]+)/gi, 'Bearer [REDACTED]')
      .replace(/("api_key"\s*:\s*")([^"]*)(")/gi, '$1[REDACTED]$3')
      .replace(/("target_api_key"\s*:\s*")([^"]*)(")/gi, '$1[REDACTED]$3')
      .replace(/api_key=([^\s&]+)/gi, 'api_key=[REDACTED]')
      .replace(/sk-[A-Za-z0-9]{10,}/g, 'sk-[REDACTED]')
  }

  const prettyJSON = (raw: string): string => {
    try {
      const parsed = JSON.parse(raw)
      return JSON.stringify(parsed, null, 2)
    } catch {
      return raw
    }
  }

  const renderSnippet = (value?: string | null): string => {
    if (!value || value.trim() === '') {
      return '-'
    }
    if (value.trim() === '[stream]') {
      return '[stream] — 该请求为流式响应，响应内容已通过 SSE 实时透传给调用方，未在日志中存储完整内容。'
    }
    return redactSecrets(prettyJSON(value))
  }

  const renderStreamResult = (value?: string | null): { reasoning?: string; content?: string } | null => {
    if (!value) return null
    const reasoningMatch = value.match(/=== reasoning_content ===\n([\s\S]*?)(?=\n=== |$)/)
    const contentMatch = value.match(/=== content ===\n([\s\S]*?)(?=\n=== |$)/)
    if (!reasoningMatch && !contentMatch) return null
    return {
      reasoning: reasoningMatch ? reasoningMatch[1].trim() : undefined,
      content: contentMatch ? contentMatch[1].trim() : undefined,
    }
  }

  const openLogDetail = useCallback(
    async (record: LogListItem) => {
      setLogDetailOpen(true)
      setSelectedLog(null)

      if (
        record.request_snippet != null ||
        record.response_snippet != null ||
        record.error != null ||
        record.id == null
      ) {
        setSelectedLog(record)
        return
      }

      setLogDetailLoading(true)
      try {
        const detail = await fetchLogDetail(record.id)
        setSelectedLog(detail)
      } catch (error) {
        const description =
          error instanceof Error ? error.message : '加载调用详情失败'
        void messageApi.error(description)
      } finally {
        setLogDetailLoading(false)
      }
    },
    [messageApi],
  )

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

  const logColumns: ColumnsType<LogListItem> = [
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (value: string) =>
        value ? new Date(value).toLocaleString('zh-CN') : '-',
    },
    {
      title: '外部模型',
      dataIndex: 'external_model',
      key: 'external_model',
      width: 220,
      ellipsis: true,
    },
    {
      title: '目标模型',
      dataIndex: 'target_model',
      key: 'target_model',
      width: 220,
      ellipsis: true,
      render: (value: string) => <Tag color="blue">{value}</Tag>,
    },
    {
      title: '状态码',
      dataIndex: 'status_code',
      key: 'status_code',
      width: 110,
      render: (value: number) => {
        const color =
          value >= 200 && value < 300 ? 'success' : value >= 400 ? 'error' : 'warning'
        return <Tag color={color}>{value}</Tag>
      },
    },
    {
      title: '耗时(ms)',
      dataIndex: 'duration_ms',
      key: 'duration_ms',
      width: 120,
      render: (value: number) => (Number.isFinite(value) ? value.toLocaleString() : '-'),
    },
    {
      title: '上游地址',
      dataIndex: 'upstream_url',
      key: 'upstream_url',
      ellipsis: true,
      render: (value: string) => (
        <Text copyable={{ text: value }} ellipsis={{ tooltip: value }}>
          {value}
        </Text>
      ),
    },
  ]

  const handleRefresh = () => {
    if (activeTabKey === 'logs') {
      void loadLogs(logsPage, logsPageSize)
      return
    }
    void loadConfigs()
  }

  if (!authed) {
    return (
      <ConfigProvider locale={zhCN}>
        <LoginPage
          onLoginSuccess={(token) => {
            setToken(token)
            setAuthed(true)
          }}
        />
      </ConfigProvider>
    )
  }

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
              onClick={handleRefresh}
              loading={activeTabKey === 'logs' ? logsLoading : loading}
            >
              刷新
            </Button>
            {activeTabKey === 'configs' ? (
              <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
                新建配置
              </Button>
            ) : null}
            <Button
              icon={<LogoutOutlined />}
              onClick={() => {
                clearToken()
                setAuthed(false)
              }}
            >
              登出
            </Button>
          </Space>
        </Layout.Header>

        <Layout.Content className="app-content">
          <Card className="content-card" bordered={false}>
            <Tabs
              activeKey={activeTabKey}
              onChange={(key) => setActiveTabKey(key as 'configs' | 'logs')}
              items={[
                {
                  key: 'configs',
                  label: '配置管理',
                  children: (
                    <>
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
                    </>
                  ),
                },
                {
                  key: 'logs',
                  label: '调用日志',
                  children: (
                    <>
                      <div className="card-heading">
                        <div>
                          <Title level={4}>调用日志</Title>
                          <Paragraph>共 {logsTotal.toLocaleString()} 条记录</Paragraph>
                        </div>
                      </div>

                      <Table<LogListItem>
                        rowKey={(record, index) =>
                          record.id != null ? String(record.id) : `${record.created_at}-${index}`
                        }
                        columns={logColumns}
                        dataSource={logs}
                        loading={logsLoading}
                        pagination={{
                          current: logsPage,
                          pageSize: logsPageSize,
                          total: logsTotal,
                          showSizeChanger: true,
                          pageSizeOptions: ['10', '20'],
                        }}
                        onChange={(pagination) => {
                          const nextPage = pagination.current ?? 1
                          const nextSize = pagination.pageSize ?? logsPageSize
                          void loadLogs(nextPage, nextSize)
                        }}
                        onRow={(record) => ({
                          onClick: () => void openLogDetail(record),
                        })}
                        scroll={{ x: 1080 }}
                      />
                    </>
                  ),
                },
              ]}
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

      <Drawer
        title="调用详情"
        open={logDetailOpen}
        width={760}
        onClose={() => {
          setLogDetailOpen(false)
          setSelectedLog(null)
          setLogDetailLoading(false)
        }}
      >
        {logDetailLoading ? (
          <Paragraph>加载中...</Paragraph>
        ) : selectedLog ? (
          <Space direction="vertical" size="large" style={{ width: '100%' }}>
            <Descriptions
              size="small"
              column={1}
              bordered
              items={[
                {
                  key: 'created_at',
                  label: '时间',
                  children: selectedLog.created_at
                    ? new Date(selectedLog.created_at).toLocaleString('zh-CN')
                    : '-',
                },
                {
                  key: 'external_model',
                  label: 'external_model',
                  children: selectedLog.external_model ?? '-',
                },
                {
                  key: 'target_model',
                  label: 'target_model',
                  children: selectedLog.target_model ?? '-',
                },
                {
                  key: 'status_code',
                  label: 'status_code',
                  children: selectedLog.status_code ?? '-',
                },
                {
                  key: 'duration_ms',
                  label: 'duration_ms',
                  children: selectedLog.duration_ms ?? '-',
                },
                {
                  key: 'upstream_url',
                  label: 'upstream_url',
                  children: selectedLog.upstream_url ? (
                    <Text copyable={{ text: selectedLog.upstream_url }}>
                      {selectedLog.upstream_url}
                    </Text>
                  ) : (
                    '-'
                  ),
                },
              ]}
            />

            <div>
              <Title level={5}>Request</Title>
              <pre style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word', margin: 0, maxHeight: 'none', overflow: 'visible' }}>
                {renderSnippet(selectedLog.request_snippet)}
              </pre>
            </div>

            <div>
              <Title level={5}>Response</Title>
              {(() => {
                const streamResult = renderStreamResult(selectedLog.response_snippet)
                if (streamResult) {
                  return (
                    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                      {streamResult.reasoning != null && (
                        <div>
                          <Tag color="purple">reasoning_content</Tag>
                          <pre style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word', margin: '8px 0 0', padding: '12px', background: '#faf5ff', borderRadius: '6px', maxHeight: 'none', overflow: 'visible' }}>
                            {redactSecrets(streamResult.reasoning)}
                          </pre>
                        </div>
                      )}
                      {streamResult.content != null && (
                        <div>
                          <Tag color="blue">content</Tag>
                          <pre style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word', margin: '8px 0 0', padding: '12px', background: '#f0f7ff', borderRadius: '6px', maxHeight: 'none', overflow: 'visible' }}>
                            {redactSecrets(streamResult.content)}
                          </pre>
                        </div>
                      )}
                    </Space>
                  )
                }
                return (
                  <pre style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word', margin: 0, maxHeight: 'none', overflow: 'visible' }}>
                    {renderSnippet(selectedLog.response_snippet)}
                  </pre>
                )
              })()}
            </div>

            <div>
              <Title level={5}>Error</Title>
              <pre style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word', margin: 0, maxHeight: 'none', overflow: 'visible' }}>
                {redactSecrets(selectedLog.error) || '-'}
              </pre>
            </div>
          </Space>
        ) : (
          <Paragraph>暂无数据</Paragraph>
        )}
      </Drawer>
    </ConfigProvider>
  )
}

export default App
