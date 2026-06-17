import { Form, Input, Modal } from 'antd'
import { useEffect } from 'react'
import type { Config, ConfigPayload } from '../types/config'

interface ConfigFormModalProps {
  open: boolean
  confirmLoading: boolean
  initialValues?: Config | null
  onCancel: () => void
  onSubmit: (values: ConfigPayload) => Promise<void>
}

export default function ConfigFormModal({
  open,
  confirmLoading,
  initialValues,
  onCancel,
  onSubmit,
}: ConfigFormModalProps) {
  const [form] = Form.useForm<ConfigPayload>()

  useEffect(() => {
    if (!open) {
      return
    }

    if (initialValues) {
      form.setFieldsValue({
        name: initialValues.name,
        external_model: initialValues.external_model,
        target_model: initialValues.target_model,
        target_base_url: initialValues.target_base_url,
        target_api_key: initialValues.target_api_key,
      })
      return
    }

    form.resetFields()
  }, [form, initialValues, open])

  const handleOk = async () => {
    const values = await form.validateFields()
    await onSubmit(values)
  }

  return (
    <Modal
      destroyOnClose
      open={open}
      title={initialValues ? '编辑配置' : '创建配置'}
      okText={initialValues ? '保存' : '创建'}
      cancelText="取消"
      confirmLoading={confirmLoading}
      onCancel={() => {
        form.resetFields()
        onCancel()
      }}
      onOk={() => void handleOk()}
    >
      <Form<ConfigPayload> form={form} layout="vertical" autoComplete="off">
        <Form.Item
          label="配置名称"
          name="name"
          rules={[{ required: true, message: '请输入配置名称' }]}
        >
          <Input maxLength={80} placeholder="例如：OpenRouter 默认线路" />
        </Form.Item>

        <Form.Item
          label="外部模型名"
          name="external_model"
          rules={[{ required: true, message: '请输入外部模型名' }]}
          extra="调用方请求中的 model 字段，例如 deepseek-chat。"
        >
          <Input maxLength={120} placeholder="deepseek-chat" />
        </Form.Item>

        <Form.Item
          label="目标模型名"
          name="target_model"
          rules={[{ required: true, message: '请输入目标模型名' }]}
        >
          <Input maxLength={120} placeholder="gpt-4.1-mini" />
        </Form.Item>

        <Form.Item
          label="目标 Base URL"
          name="target_base_url"
          rules={[
            { required: true, message: '请输入目标 Base URL' },
            { type: 'url', message: '请输入有效的 URL' },
          ]}
        >
          <Input placeholder="https://api.openai.com/v1" />
        </Form.Item>

        <Form.Item
          label="目标 API Key"
          name="target_api_key"
          rules={[{ required: true, message: '请输入目标 API Key' }]}
        >
          <Input.Password placeholder="sk-..." visibilityToggle />
        </Form.Item>
      </Form>
    </Modal>
  )
}
