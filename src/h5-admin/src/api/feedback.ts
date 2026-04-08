import { get, put } from '@/utils/request'
import type { PaginatedData } from '@/utils/request'

/** 反馈信息 */
export interface Feedback {
  id: number
  user_id: number
  user_nickname: string
  user_role: string
  content: string
  images?: string[]
  status: 'pending' | 'processing' | 'resolved'
  reply?: string
  created_at: string
  updated_at: string
}

/** 反馈列表查询参数 */
export interface FeedbackListParams {
  page?: number
  page_size?: number
  status?: string
  user_id?: number
}

/**
 * 获取反馈列表
 */
export function getFeedbackList(params: FeedbackListParams) {
  return get<PaginatedData<Feedback>>('/api/admin/feedbacks', params)
}

/**
 * 更新反馈状态
 * @param feedbackId - 反馈ID
 * @param status - 状态
 * @param reply - 回复内容
 */
export function updateFeedback(feedbackId: number, status: string, reply?: string) {
  return put(`/api/admin/feedbacks/${feedbackId}`, { status, reply })
}
