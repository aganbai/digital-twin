import { request, ApiResponse, PaginatedData } from './request'

/** 反馈类型 */
export type FeedbackType = 'suggestion' | 'bug' | 'content_issue' | 'other'

/** 反馈状态 */
export type FeedbackStatus = 'pending' | 'reviewed' | 'resolved'

/** 反馈类型选项 */
export const FEEDBACK_TYPES = [
  { value: 'suggestion' as FeedbackType, label: '💡 功能建议', desc: '希望增加的新功能' },
  { value: 'bug' as FeedbackType, label: '🐛 Bug报告', desc: '发现的问题或错误' },
  { value: 'content_issue' as FeedbackType, label: '📝 内容问题', desc: 'AI回答不准确等' },
  { value: 'other' as FeedbackType, label: '💬 其他', desc: '其他意见或建议' },
]

/** 反馈状态选项 */
export const FEEDBACK_STATUSES = [
  { value: '' as string, label: '全部' },
  { value: 'pending' as FeedbackStatus, label: '待处理' },
  { value: 'reviewed' as FeedbackStatus, label: '已查看' },
  { value: 'resolved' as FeedbackStatus, label: '已解决' },
]

/** 反馈项数据结构 */
export interface FeedbackItem {
  id: number
  feedback_type: FeedbackType
  content: string
  status: FeedbackStatus
  context_info?: Record<string, any>
  user_id?: number
  user_nickname?: string
  created_at: string
  updated_at?: string
}

/** 提交反馈 */
export function createFeedback(params: {
  feedback_type: FeedbackType
  content: string
  context_info?: Record<string, any>
}) {
  return request<{ id: number; message: string }>({
    url: '/api/feedbacks',
    method: 'POST',
    data: params,
  })
}

/** 获取反馈列表（教师/管理员） */
export function getFeedbacks(params?: {
  status?: string
  page?: number
  page_size?: number
}) {
  return request<PaginatedData<FeedbackItem>>({
    url: '/api/feedbacks',
    method: 'GET',
    data: params,
  })
}

/** 更新反馈状态（教师/管理员） */
export function updateFeedbackStatus(id: number, status: FeedbackStatus) {
  return request<{ message: string }>({
    url: `/api/feedbacks/${id}/status`,
    method: 'PUT',
    data: { status },
  })
}
