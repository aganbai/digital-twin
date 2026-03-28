import { request } from './request'

/** 用户统计信息 */
export interface UserStats {
  conversation_count: number
  memory_count: number
  document_count?: number
}

/** 用户详情 */
export interface UserProfile {
  id: number
  username: string
  nickname: string
  role: string
  email: string
  created_at: string
  stats: UserStats
}

/**
 * 获取当前用户信息
 */
export function getUserProfile() {
  return request<UserProfile>({
    url: '/api/user/profile',
    method: 'GET',
  })
}
