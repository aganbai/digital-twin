import { request, PaginatedData } from './request'

/** 记忆条目 */
export interface Memory {
  id: number
  memory_type: string
  content: string
  importance: number
  last_accessed: string
  created_at: string
}

/**
 * 获取学生记忆列表
 * @param teacherId - 教师 ID
 * @param memoryType - 记忆类型（可选）
 * @param page - 页码，默认 1
 * @param pageSize - 每页数量，默认 20
 */
export function getMemories(
  teacherId: number,
  memoryType?: string,
  page = 1,
  pageSize = 20,
) {
  const query = new URLSearchParams()
  query.append('teacher_id', String(teacherId))
  if (memoryType) {
    query.append('memory_type', memoryType)
  }
  query.append('page', String(page))
  query.append('page_size', String(pageSize))

  return request<PaginatedData<Memory>>({
    url: `/api/memories?${query.toString()}`,
    method: 'GET',
  })
}
