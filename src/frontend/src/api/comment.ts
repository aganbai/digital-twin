import { request, PaginatedData } from './request'

/** 评语项 */
export interface CommentItem {
  id: number
  teacher_id: number
  teacher_nickname: string
  student_id: number
  student_nickname: string
  content: string
  progress_summary: string
  created_at: string
}

/** 创建评语请求 */
export interface CreateCommentData {
  student_id: number
  content: string
  progress_summary?: string
}

/** 获取评语列表参数 */
export interface GetCommentsParams {
  student_id?: number
  teacher_id?: number
  page?: number
  page_size?: number
}

/**
 * 教师写评语
 * @param data - 评语数据
 */
export function createComment(data: CreateCommentData) {
  return request<CommentItem>({
    url: '/api/comments',
    method: 'POST',
    data,
  })
}

/**
 * 获取评语列表
 * @param params - 查询参数
 */
export function getComments(params: GetCommentsParams = {}) {
  const query = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null) {
      query.append(key, String(value))
    }
  })
  const queryStr = query.toString()
  return request<PaginatedData<CommentItem>>({
    url: `/api/comments${queryStr ? '?' + queryStr : ''}`,
    method: 'GET',
  })
}
