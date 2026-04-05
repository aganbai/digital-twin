import { get } from '@/utils/request'
import type { PaginatedData } from '@/utils/request'

/** 操作日志 */
export interface OperationLog {
  id: number
  user_id: number
  user_nickname: string
  user_role: string
  persona_id: number
  action: string
  resource: string
  resource_id: string
  detail: string
  ip: string
  user_agent: string
  platform: string
  status_code: number
  duration_ms: number
  created_at: string
}

/** 日志列表查询参数 */
export interface LogListParams {
  page?: number
  page_size?: number
  user_id?: number
  action?: string
  resource?: string
  platform?: string
  start_date?: string
  end_date?: string
}

/** 日志统计数据 */
export interface LogStatsData {
  action_distribution: { action: string; count: number }[]
  platform_distribution: { platform: string; count: number }[]
  hourly_heatmap: { hour: number; count: number }[]
  active_users: { user_id: number; nickname: string; count: number }[]
}

/**
 * 获取日志列表
 */
export function getLogList(params: LogListParams) {
  return get<PaginatedData<OperationLog>>('/api/admin/logs', params)
}

/**
 * 获取日志统计
 * @param days - 天数
 */
export function getLogStats(days: number = 7) {
  return get<LogStatsData>('/api/admin/logs/stats', { days })
}

/**
 * 导出日志
 * @param params - 查询参数
 */
export function exportLogs(params: LogListParams) {
  const query = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null) {
      query.append(key, String(value))
    }
  })
  return `/api/admin/logs/export?${query.toString()}`
}
