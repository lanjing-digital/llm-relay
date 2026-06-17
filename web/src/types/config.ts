export interface Config {
  id: number
  name: string
  external_model: string
  target_model: string
  target_base_url: string
  target_api_key: string
  created_at: string
  updated_at: string
}

export interface ConfigPayload {
  name: string
  external_model: string
  target_model: string
  target_base_url: string
  target_api_key: string
}

export interface TestConfigResponse {
  message?: string
  error?: string
  detail?: string
}
