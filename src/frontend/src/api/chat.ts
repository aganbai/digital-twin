import { request, PaginatedData } from './request'

/** 对话消息 */
export interface Conversation {
  id: number
  session_id: string
  role: 'user' | 'assistant'
  content: string
  created_at: string
}

/** 发送消息响应 */
export interface ChatResponse {
  reply: string
  session_id: string
  conversation_id: number
  token_usage: {
    prompt_tokens: number
    completion_tokens: number
    total_tokens: number
  }
  pipeline_duration_ms: number
}

/** 会话摘要 */
export interface Session {
  session_id: string
  teacher_id: number
  teacher_nickname: string
  last_message: string
  last_message_role: string
  message_count: number
  updated_at: string
}

/** 获取对话历史参数 */
export interface GetConversationsParams {
  teacher_id?: number
  session_id?: string
  page?: number
  page_size?: number
}

/**
 * 发送对话消息
 * @param message - 消息内容
 * @param teacherId - 教师 ID
 * @param sessionId - 会话 ID（可选，不传则新建会话）
 */
export function sendMessage(message: string, teacherId: number, sessionId?: string) {
  return request<ChatResponse>({
    url: '/api/chat',
    method: 'POST',
    data: {
      message,
      teacher_id: teacherId,
      session_id: sessionId,
    },
  })
}

/**
 * 获取对话历史
 * @param params - 查询参数
 */
export function getConversations(params: GetConversationsParams) {
  const query = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null) {
      query.append(key, String(value))
    }
  })
  return request<PaginatedData<Conversation>>({
    url: `/api/conversations?${query.toString()}`,
    method: 'GET',
  })
}

/**
 * 获取会话列表
 * @param page - 页码，默认 1
 * @param pageSize - 每页数量，默认 20
 */
export function getSessions(page = 1, pageSize = 20) {
  return request<PaginatedData<Session>>({
    url: `/api/conversations/sessions?page=${page}&page_size=${pageSize}`,
    method: 'GET',
  })
}
