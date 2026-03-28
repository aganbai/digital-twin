import { request, PaginatedData } from './request'

/** 教师信息 */
export interface Teacher {
  id: number
  username: string
  nickname: string
  role: string
  document_count: number
  created_at: string
}

/**
 * 获取教师列表
 * @param page - 页码，默认 1
 * @param pageSize - 每页数量，默认 20
 */
export function getTeachers(page = 1, pageSize = 20) {
  return request<PaginatedData<Teacher>>({
    url: `/api/teachers?page=${page}&page_size=${pageSize}`,
    method: 'GET',
  })
}
