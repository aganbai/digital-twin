import { get, put } from '@/utils/request'
import type { PaginatedData } from '@/utils/request'

/** 用户信息 */
export interface User {
  id: number
  openid: string
  role: 'teacher' | 'student' | 'admin'
  nickname: string
  avatar?: string
  status: 'active' | 'disabled'
  created_at: string
  updated_at: string
  last_active_at?: string
}

/** 用户列表查询参数 */
export interface UserListParams {
  page?: number
  page_size?: number
  nickname?: string
  role?: string
  status?: string
}

/**
 * 获取用户列表
 */
export function getUserList(params: UserListParams) {
  return get<PaginatedData<User>>('/api/admin/users', params)
}

/**
 * 修改用户角色
 * @param userId - 用户ID
 * @param role - 新角色
 */
export function updateUserRole(userId: number, role: string) {
  return put(`/api/admin/users/${userId}/role`, { role })
}

/**
 * 启用/禁用用户
 * @param userId - 用户ID
 * @param status - 状态
 */
export function updateUserStatus(userId: number, status: 'active' | 'disabled') {
  return put(`/api/admin/users/${userId}/status`, { status })
}
