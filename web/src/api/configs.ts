import apiClient from './client'
import type { Config, ConfigPayload, TestConfigResponse } from '../types/config'

export async function fetchConfigs() {
  const { data } = await apiClient.get<Config[]>('/configs')
  return data
}

export async function createConfig(payload: ConfigPayload) {
  const { data } = await apiClient.post<Config>('/configs', payload)
  return data
}

export async function updateConfig(id: number, payload: ConfigPayload) {
  const { data } = await apiClient.put<Config>(`/configs/${id}`, payload)
  return data
}

export async function deleteConfig(id: number) {
  const { data } = await apiClient.delete<{ message: string }>(`/configs/${id}`)
  return data
}

export async function testConfig(id: number) {
  const { data } = await apiClient.post<TestConfigResponse>(`/configs/${id}/test`)
  return data
}
