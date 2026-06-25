import { useState } from 'react'
import { Button, Card, Form, Input, message } from 'antd'
import { LockOutlined, UserOutlined } from '@ant-design/icons'
import { login } from '../api/auth'

interface LoginForm {
  username: string
  password: string
}

interface LoginPageProps {
  onLoginSuccess: (token: string) => void
}

function LoginPage({ onLoginSuccess }: LoginPageProps) {
  const [loading, setLoading] = useState(false)
  const [messageApi, contextHolder] = message.useMessage()

  const handleSubmit = async (values: LoginForm) => {
    setLoading(true)
    try {
      const { token } = await login(values.username, values.password)
      onLoginSuccess(token)
    } catch (error) {
      const description = error instanceof Error ? error.message : '登录失败'
      void messageApi.error(description)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      }}
    >
      {contextHolder}
      <Card
        style={{ width: 400, boxShadow: '0 8px 24px rgba(0,0,0,0.15)' }}
        styles={{ body: { padding: '32px 32px 8px' } }}
      >
        <h2 style={{ textAlign: 'center', marginBottom: 32, color: '#333' }}>
          LLM Relay 管理后台
        </h2>
        <Form<LoginForm>
          name="login"
          size="large"
          onFinish={handleSubmit}
          autoComplete="off"
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入账号' }]}
          >
            <Input prefix={<UserOutlined />} placeholder="账号" />
          </Form.Item>
          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0 }}>
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              block
            >
              登录
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}

export default LoginPage
