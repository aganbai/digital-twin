import Taro from '@tarojs/taro'
import { request, PaginatedData } from './request'
import { BASE_URL } from '../utils/constants'

/** 对话消息 */
export interface Conversation {
  id: number
  session_id: string
  role: 'user' | 'assistant'
  content: string
  /** 消息发送者类型：student / ai / teacher */
  sender_type?: 'student' | 'ai' | 'teacher' | ''
  /** 引用的消息 ID（0 表示非引用回复） */
  reply_to_id?: number
  /** 引用的消息内容摘要 */
  reply_to_content?: string
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
  teacher_persona_id?: number
  teacher_nickname: string
  last_message: string
  last_message_role: string
  message_count: number
  updated_at: string
}

/** 获取对话历史参数 */
export interface GetConversationsParams {
  teacher_id?: number
  teacher_persona_id?: number
  session_id?: string
  page?: number
  page_size?: number
}

/**
 * 发送对话消息
 * @param message - 消息内容
 * @param teacherPersonaId - 教师分身 ID
 * @param sessionId - 会话 ID（可选，不传则新建会话）
 * @param attachment - 附件信息（可选）
 */
export function sendMessage(
  message: string,
  teacherPersonaId: number,
  sessionId?: string,
  attachment?: { url: string; type: string; name: string },
) {
  return request<ChatResponse>({
    url: '/api/chat',
    method: 'POST',
    data: {
      message,
      teacher_persona_id: teacherPersonaId,
      session_id: sessionId,
      ...(attachment ? {
        attachment_url: attachment.url,
        attachment_type: attachment.type,
        attachment_name: attachment.name,
      } : {}),
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

/** SSE 流式事件类型 */
export interface StreamStartEvent {
  type: 'start'
  session_id: string
}

export interface StreamDeltaEvent {
  type: 'delta'
  content: string
}

export interface StreamDoneEvent {
  type: 'done'
  conversation_id: number
  token_usage: {
    prompt_tokens: number
    completion_tokens: number
    total_tokens: number
  }
}

export interface StreamErrorEvent {
  type: 'error'
  code: number
  message: string
}

export type StreamEvent = StreamStartEvent | StreamDeltaEvent | StreamDoneEvent | StreamErrorEvent

/** 流式对话回调 */
export interface ChatStreamCallbacks {
  onStart?: (sessionId: string) => void
  onDelta?: (content: string) => void
  onDone?: (conversationId: number) => void
  onError?: (code: number, message: string) => void
}

/**
 * SSE 流式对话
 * 使用微信小程序 enableChunked 接收流式数据
 * @param message - 消息内容
 * @param teacherPersonaId - 教师分身 ID
 * @param callbacks - 流式回调
 * @param sessionId - 会话 ID（可选）
 * @returns 返回 RequestTask，可用于中断请求
 */
export function chatStream(
  message: string,
  teacherPersonaId: number,
  callbacks: ChatStreamCallbacks,
  sessionId?: string,
  attachment?: { url: string; type: string; name: string },
) {
  const token = Taro.getStorageSync('token') || ''

  let buffer = '' // 用于缓存不完整的 SSE 数据

  const requestTask = Taro.request({
    url: `${BASE_URL}/api/chat/stream`,
    method: 'POST',
    data: {
      message,
      teacher_persona_id: teacherPersonaId,
      session_id: sessionId,
      ...(attachment ? {
        attachment_url: attachment.url,
        attachment_type: attachment.type,
        attachment_name: attachment.name,
      } : {}),
    },
    header: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    enableChunked: true,
    responseType: 'text',
    success: (res) => {
      // 非流式响应（错误情况）
      if (res.statusCode !== 200) {
        const data = typeof res.data === 'string' ? JSON.parse(res.data) : res.data
        callbacks.onError?.(data.code || res.statusCode, data.message || '请求失败')
      }
    },
    fail: () => {
      callbacks.onError?.(0, '网络异常，请检查网络连接')
    },
  })

  // 监听分块数据
  requestTask.onChunkReceived?.((response: { data: ArrayBuffer }) => {
    try {
      // 将 ArrayBuffer 转为字符串
      const text = arrayBufferToString(response.data)
      buffer += text

      // 按行解析 SSE 数据
      const lines = buffer.split('\n')
      buffer = lines.pop() || '' // 最后一行可能不完整，保留在 buffer 中

      for (const line of lines) {
        const trimmed = line.trim()
        if (!trimmed || !trimmed.startsWith('data: ')) continue

        const jsonStr = trimmed.slice(6) // 去掉 "data: " 前缀
        if (!jsonStr) continue

        try {
          const event: StreamEvent = JSON.parse(jsonStr)
          switch (event.type) {
            case 'start':
              callbacks.onStart?.(event.session_id)
              break
            case 'delta':
              callbacks.onDelta?.(event.content)
              break
            case 'done':
              callbacks.onDone?.(event.conversation_id)
              break
            case 'error':
              callbacks.onError?.(event.code, event.message)
              break
          }
        } catch {
          // 忽略解析失败的行
        }
      }
    } catch {
      // 忽略解码错误
    }
  })

  return requestTask
}

/**
 * ArrayBuffer 转字符串
 * 微信小程序环境下的兼容实现
 */
function arrayBufferToString(buffer: ArrayBuffer): string {
  const uint8Array = new Uint8Array(buffer)
  let result = ''
  // 使用简单的逐字节解码（UTF-8 兼容 ASCII 部分）
  for (let i = 0; i < uint8Array.length; i++) {
    result += String.fromCharCode(uint8Array[i])
  }
  try {
    return decodeURIComponent(escape(result))
  } catch {
    return result
  }
}

/** 教师真人回复请求参数 */
export interface TeacherReplyParams {
  student_persona_id: number
  session_id: string
  content: string
  reply_to_id?: number
}

/** 教师真人回复响应 */
export interface TeacherReplyResponse {
  conversation_id: number
  sender_type: 'teacher'
  reply_to_id: number
  reply_to_content: string
  takeover_status: string
  created_at: string
}

/** 接管状态响应 */
export interface TakeoverStatusResponse {
  is_taken_over: boolean
  teacher_persona_id: number
  teacher_nickname: string
  started_at: string
}

/** 学生对话记录响应 */
export interface StudentConversationsResponse {
  student_persona_id: number
  student_nickname: string
  session_id: string
  takeover_status: 'active' | 'ended' | 'none'
  messages: Conversation[]
  total: number
  page: number
  page_size: number
}

/** 退出接管响应 */
export interface EndTakeoverResponse {
  session_id: string
  status: string
  ended_at: string
}

/**
 * 教师真人回复
 * @param params - 回复参数
 */
export function teacherReply(params: TeacherReplyParams) {
  return request<TeacherReplyResponse>({
    url: '/api/chat/teacher-reply',
    method: 'POST',
    data: params,
  })
}

/**
 * 查询接管状态
 * @param sessionId - 会话 ID
 */
export function getTakeoverStatus(sessionId: string) {
  return request<TakeoverStatusResponse>({
    url: `/api/chat/takeover-status?session_id=${encodeURIComponent(sessionId)}`,
    method: 'GET',
  })
}

/**
 * 教师退出接管
 * @param sessionId - 会话 ID
 */
export function endTakeover(sessionId: string) {
  return request<EndTakeoverResponse>({
    url: '/api/chat/end-takeover',
    method: 'POST',
    data: { session_id: sessionId },
  })
}

/**
 * 教师查看学生对话记录
 * @param studentPersonaId - 学生分身 ID
 * @param sessionId - 会话 ID（可选）
 * @param page - 页码
 * @param pageSize - 每页数量
 */
export function getStudentConversations(
  studentPersonaId: number,
  sessionId?: string,
  page = 1,
  pageSize = 50,
) {
  const query = new URLSearchParams()
  if (sessionId) query.append('session_id', sessionId)
  query.append('page', String(page))
  query.append('page_size', String(pageSize))
  return request<StudentConversationsResponse>({
    url: `/api/conversations/student/${studentPersonaId}?${query.toString()}`,
    method: 'GET',
  })
}
