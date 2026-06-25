import apiClient from './client'

interface LoginResponse {
  token: string
  token_type: string
  expires_in: number
}

export async function login(username: string, password: string): Promise<LoginResponse> {
  const { data } = await apiClient.post<LoginResponse>('/auth/login', {
    username,
    password,
  })
  return data
}
