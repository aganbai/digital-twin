import { get, post, put, del } from '@/utils/request'
import type { PaginatedData } from '@/utils/request'

/** 知识条目 */
export interface Knowledge {
  id: number
  title: string
  content?: string
  type: 'document' | 'qa'
  status: string
  created_at: string
}

/** 问答条目 */
export interface QAItem {
  id: number
  question: string
  answer: string
  created_at: string
}

/**
 * 获取知识库列表
 */
export function getKnowledgeList(type?: string) {
  return get<Knowledge[]>('/api/teacher/knowledge', { type })
}

/**
 * 上传知识文档
 */
export function uploadKnowledge(formData: FormData) {
  return post('/api/teacher/knowledge/upload', formData)
}

/**
 * 创建问答条目
 */
export function createQA(question: string, answer: string) {
  return post<QAItem>('/api/teacher/knowledge/qa', { question, answer })
}

/**
 * 更新问答条目
 */
export function updateQA(id: number, question: string, answer: string) {
  return put(`/api/teacher/knowledge/qa/${id}`, { question, answer })
}

/**
 * 删除知识条目
 */
export function deleteKnowledge(id: number) {
  return del(`/api/teacher/knowledge/${id}`)
}
