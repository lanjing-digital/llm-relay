export interface LogListItem {
  id?: number | string
  created_at: string
  external_model: string
  target_model: string
  status_code: number
  duration_ms: number
  upstream_url: string
  request_snippet?: string | null
  response_snippet?: string | null
  error?: string | null
}

export interface LogDetail {
  id?: number | string
  created_at: string
  external_model: string
  target_model: string
  status_code: number
  duration_ms: number
  upstream_url: string
  request_snippet?: string | null
  response_snippet?: string | null
  error?: string | null
}

export interface FetchLogsParams {
  page: number
  pageSize: number
}

export interface FetchLogsResponse {
  items: LogListItem[]
  total: number
}
