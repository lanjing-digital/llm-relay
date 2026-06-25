import apiClient from './client'
import type {
  FetchLogsParams,
  FetchLogsResponse,
  LogDetail,
  LogListItem,
} from '../types/log'

type LogsListBackendResponse =
  | {
      items: LogListItem[]
      total: number
    }
  | {
      data: LogListItem[]
      total: number
    }
  | {
      logs: LogListItem[]
      total: number
    }

export async function fetchLogs(params: FetchLogsParams): Promise<FetchLogsResponse> {
  const { data } = await apiClient.get<LogsListBackendResponse | LogListItem[]>('/logs', {
    params: {
      page: params.page,
      page_size: params.pageSize,
      pageSize: params.pageSize,
    },
  })

  if (Array.isArray(data)) {
    return { items: data, total: data.length }
  }

  if ('items' in data) {
    return { items: data.items ?? [], total: data.total ?? 0 }
  }

  if ('data' in data) {
    return { items: data.data ?? [], total: data.total ?? 0 }
  }

  return { items: data.logs ?? [], total: data.total ?? 0 }
}

export async function fetchLogDetail(id: number | string): Promise<LogDetail> {
  const { data } = await apiClient.get<LogDetail>(`/logs/${id}`)
  return data
}

