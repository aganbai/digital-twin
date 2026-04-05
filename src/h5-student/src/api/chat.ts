import { get, post } from '@/utils/request'

/** 对话信息 */
export interface Conversation {
  id: number
  title: string
  created_at: string
  updated_at: string
}

/** 消息信息 */
export interface Message {
  id: number
  conversation_id: number
  role: 'user' | 'assistant'
  content: string
  created_at: string
}

/**
 * 获取对话历史列表
 */
export function getConversationList() {
  return get<Conversation[]>('/api/student/conversations')
}

/**
 * 获取对话消息
 */
export function getMessages(conversationId: number) {
  return get<Message[]>(`/api/student/conversations/${conversationId}/messages`)
}

/**
 * 发送消息
 */
export function sendMessage(conversationId: number, content: string) {
  return post<Message>(`/api/student/conversations/${conversationId}/messages`, { content })
}

/**
 * 创建新对话
 */
export function createConversation() {
  return post<Conversation>('/api/student/conversations')
}

/**
 * 删除对话
 */
export function deleteConversation(conversationId: number) {
  return post(`/api/student/conversations/${conversationId}/delete`)
}
