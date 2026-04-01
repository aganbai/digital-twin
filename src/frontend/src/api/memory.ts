import { request, PaginatedData } from './request'

/** 记忆层级类型 */
export type MemoryLayer = 'core' | 'episodic' | 'archived'

/** 记忆层级标签映射 */
export const MEMORY_LAYER_LABELS: Record<MemoryLayer, string> = {
  core: '核心记忆',
  episodic: '情景记忆',
  archived: '已归档',
}

/** 记忆条目 */
export interface Memory {
  id: number
  memory_type: string
  memory_layer: MemoryLayer
  content: string
  importance: number
  last_accessed: string
  created_at: string
  updated_at: string
}

/** 摘要合并响应 */
export interface SummarizeResponse {
  summarized_count: number
  new_core_memories: Memory[]
  archived_count: number
}

/**
 * 获取学生记忆列表
 * @param teacherPersonaId - 教师分身 ID
 * @param studentPersonaId - 学生分身 ID
 * @param layer - 记忆层级筛选（可选）
 * @param page - 页码，默认 1
 * @param pageSize - 每页数量，默认 20
 */
export function getMemories(
  teacherPersonaId: number,
  studentPersonaId: number,
  layer?: MemoryLayer,
  page = 1,
  pageSize = 20,
) {
  const query = new URLSearchParams()
  query.append('teacher_persona_id', String(teacherPersonaId))
  query.append('student_persona_id', String(studentPersonaId))
  if (layer) {
    query.append('layer', layer)
  }
  query.append('page', String(page))
  query.append('page_size', String(pageSize))

  return request<PaginatedData<Memory>>({
    url: `/api/memories?${query.toString()}`,
    method: 'GET',
  })
}

/**
 * 编辑记忆
 * @param id - 记忆 ID
 * @param data - 编辑内容
 */
export function updateMemory(
  id: number,
  data: { content: string; importance?: number; memory_layer?: MemoryLayer },
) {
  return request<Memory>({
    url: `/api/memories/${id}`,
    method: 'PUT',
    data,
  })
}

/**
 * 删除记忆
 * @param id - 记忆 ID
 */
export function deleteMemory(id: number) {
  return request<null>({
    url: `/api/memories/${id}`,
    method: 'DELETE',
  })
}

/**
 * 触发记忆摘要合并
 * @param teacherPersonaId - 教师分身 ID
 * @param studentPersonaId - 学生分身 ID
 */
export function summarizeMemories(teacherPersonaId: number, studentPersonaId: number) {
  return request<SummarizeResponse>({
    url: '/api/memories/summarize',
    method: 'POST',
    data: {
      teacher_persona_id: teacherPersonaId,
      student_persona_id: studentPersonaId,
    },
  })
}
