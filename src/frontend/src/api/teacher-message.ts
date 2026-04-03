import { request, PaginatedData } from './request'

/** 教师推送消息 */
export interface TeacherMessage {
  id: number
  teacher_id: number
  target_type: 'class' | 'student'
  target_id: number
  content: string
  status: string
  created_at: string
}

/** 推送消息请求 */
export interface PushMessageParams {
  target_type: 'class' | 'student'
  target_id: number
  content: string
  persona_id: number
}

/** 推送历史查询参数 */
export interface MessageHistoryParams {
  persona_id?: number
  page?: number
  page_size?: number
}

/** 推送消息响应 */
export interface PushMessageResponse {
  message_id: number
  target_count: number
  success_count: number
  failed_count: number
}

/** 推送消息 */
export function pushTeacherMessage(params: PushMessageParams) {
  return request<PushMessageResponse>({
    url: '/api/teacher-messages',
    method: 'POST',
    data: params,
  })
}

/** 获取推送历史 */
export function getTeacherMessageHistory(page = 1, pageSize = 20, personaId?: number) {
  let url = `/api/teacher-messages/history?page=${page}&page_size=${pageSize}`
  if (personaId) url += `&persona_id=${personaId}`
  return request<PaginatedData<TeacherMessage>>({
    url,
    method: 'GET',
  })
}

/** sendTeacherMessage 别名（兼容） */
export const sendTeacherMessage = pushTeacherMessage
