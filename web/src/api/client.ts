import axios from 'axios'

const TOKEN_KEY = 'llm_relay_token'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

const apiClient = axios.create({
  baseURL: '/api',
  timeout: 15000,
})

apiClient.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

let onUnauthorized: (() => void) | null = null

export function setOnUnauthorized(handler: () => void): void {
  onUnauthorized = handler
}

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      clearToken()
      if (onUnauthorized) {
        onUnauthorized()
      }
    }
    const message =
      error.response?.data?.error ??
      error.response?.data?.message ??
      error.message ??
      '请求失败'

    return Promise.reject(new Error(message))
  },
)

export default apiClient
