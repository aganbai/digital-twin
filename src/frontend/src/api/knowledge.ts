import Taro from '@tarojs/taro'
import { request, PaginatedData } from './request'
import { getToken } from '../utils/storage'
import { BASE_URL } from '../utils/constants'

/** 知识文档 */
export interface Document {
  id: number
  title: string
  doc_type: string
  tags: string[]
  status: string
  created_at: string
  updated_at: string
}

/** 添加文档响应 */
export interface AddDocumentResponse {
  document_id: number
  title: string
  chunks_count: number
  status: string
}

/** 预览响应 */
export interface PreviewResponse {
  preview_id: string
  title: string
  /** LLM 自动生成的标题（生成失败时为空字符串） */
  llm_title: string
  /** LLM 自动生成的摘要（生成失败时为空字符串） */
  llm_summary: string
  tags: string
  total_chars: number
  chunks: Array<{ index: number; content: string; char_count: number }>
  chunk_count: number
  doc_type?: string
  source_url?: string
}

/** 确认入库响应 */
export interface ConfirmResponse {
  document_id: number
  title: string
  chunks_count: number
  status: string
}

/** 删除文档响应 */
export interface DeleteDocumentResponse {
  document_id: number
  deleted: boolean
}

/**
 * 获取文档列表
 * @param page - 页码，默认 1
 * @param pageSize - 每页数量，默认 20
 * @param scope - 文档范围：global / class / student（可选）
 * @param scopeId - 范围 ID（可选）
 */
export function getDocuments(page = 1, pageSize = 20, scope?: string, scopeId?: number) {
  let url = `/api/documents?page=${page}&page_size=${pageSize}`
  if (scope) url += `&scope=${scope}`
  if (scopeId !== undefined) url += `&scope_id=${scopeId}`
  return request<PaginatedData<Document>>({
    url,
    method: 'GET',
  })
}

/**
 * 添加知识文档
 * @param title - 文档标题
 * @param content - 文档内容
 * @param tags - 分类标签
 * @param scope - 文档范围：global / class / student（可选）
 * @param scopeId - 范围 ID（可选）
 */
export function addDocument(
  title: string,
  content: string,
  tags: string[] = [],
  scope?: string,
  scopeId?: number,
  scopeIds?: number[],
) {
  const data: Record<string, any> = { title, content, tags }
  if (scope) data.scope = scope
  if (scopeId !== undefined) data.scope_id = scopeId
  if (scopeIds && scopeIds.length > 0) data.scope_ids = scopeIds
  return request<AddDocumentResponse>({
    url: '/api/documents',
    method: 'POST',
    data,
  })
}

/**
 * 删除文档
 * @param id - 文档 ID
 */
export function deleteDocument(id: number) {
  return request<DeleteDocumentResponse>({
    url: `/api/documents/${id}`,
    method: 'DELETE',
  })
}

/** 文件上传响应 */
export interface UploadDocumentResponse {
  document_id: number
  title: string
  doc_type: string
  chunks_count: number
  file_size: number
  status: string
}

/** URL 导入响应 */
export interface ImportUrlResponse {
  document_id: number
  title: string
  doc_type: string
  chunks_count: number
  content_length: number
  source_url: string
  status: string
}

/**
 * 上传文档文件（multipart/form-data）
 * @param filePath - 本地文件路径
 * @param title - 文档标题（可选）
 * @param tags - 标签（逗号分隔，可选）
 * @param scope - 文档范围（可选）
 * @param scopeId - 范围 ID（可选）
 */
export function uploadDocument(
  filePath: string,
  title?: string,
  tags?: string,
  scope?: string,
  scopeId?: number,
  scopeIds?: number[],
) {
  const token = getToken()

  const formData: Record<string, string> = {}
  if (title) formData.title = title
  if (tags) formData.tags = tags
  if (scope) formData.scope = scope
  if (scopeId !== undefined) formData.scope_id = String(scopeId)
  if (scopeIds && scopeIds.length > 0) formData.scope_ids = JSON.stringify(scopeIds)

  return new Promise<UploadDocumentResponse>((resolve, reject) => {
    Taro.uploadFile({
      url: `${BASE_URL}/api/documents/upload`,
      filePath,
      name: 'file',
      formData,
      header: {
        Authorization: `Bearer ${token}`,
      },
      success: (res) => {
        try {
          const data = JSON.parse(res.data)
          if (data.code === 0) {
            resolve(data.data)
          } else {
            Taro.showToast({ title: data.message || '上传失败', icon: 'none' })
            reject(new Error(data.message))
          }
        } catch {
          reject(new Error('解析响应失败'))
        }
      },
      fail: (err) => {
        Taro.showToast({ title: '网络异常', icon: 'none' })
        reject(err)
      },
    })
  })
}

/**
 * URL 导入文档
 * @param url - 目标网页 URL
 * @param title - 文档标题（可选）
 * @param tags - 标签（逗号分隔，可选）
 * @param scope - 文档范围（可选）
 * @param scopeId - 范围 ID（可选）
 */
export function importUrl(
  url: string,
  title?: string,
  tags?: string,
  scope?: string,
  scopeId?: number,
  scopeIds?: number[],
) {
  const data: Record<string, any> = { url }
  if (title) data.title = title
  if (tags) data.tags = tags
  if (scope) data.scope = scope
  if (scopeId !== undefined) data.scope_id = scopeId
  if (scopeIds && scopeIds.length > 0) data.scope_ids = scopeIds
  return request<ImportUrlResponse>({
    url: '/api/documents/import-url',
    method: 'POST',
    data,
  })
}

/**
 * 文本预览
 * @param title - 文档标题
 * @param content - 文档内容
 * @param tags - 标签（可选）
 */
export function previewDocument(title: string, content: string, tags?: string) {
  return request<PreviewResponse>({
    url: '/api/documents/preview',
    method: 'POST',
    data: { title, content, tags },
  })
}

/**
 * 文件上传预览（multipart/form-data）
 * @param filePath - 本地文件路径
 * @param title - 文档标题（可选）
 * @param tags - 标签（可选）
 */
export function previewUpload(filePath: string, title?: string, tags?: string) {
  const token = getToken()

  const formData: Record<string, string> = {}
  if (title) formData.title = title
  if (tags) formData.tags = tags

  return new Promise<PreviewResponse>((resolve, reject) => {
    Taro.uploadFile({
      url: `${BASE_URL}/api/documents/preview-upload`,
      filePath,
      name: 'file',
      formData,
      header: {
        Authorization: `Bearer ${token}`,
      },
      success: (res) => {
        try {
          const data = JSON.parse(res.data)
          if (data.code === 0) {
            resolve(data.data)
          } else {
            Taro.showToast({ title: data.message || '预览失败', icon: 'none' })
            reject(new Error(data.message))
          }
        } catch {
          reject(new Error('解析响应失败'))
        }
      },
      fail: (err) => {
        Taro.showToast({ title: '网络异常', icon: 'none' })
        reject(err)
      },
    })
  })
}

/**
 * URL 导入预览
 * @param url - 目标网页 URL
 * @param title - 文档标题（可选）
 * @param tags - 标签（可选）
 */
export function previewUrl(url: string, title?: string, tags?: string) {
  return request<PreviewResponse>({
    url: '/api/documents/preview-url',
    method: 'POST',
    data: { url, title, tags },
  })
}

/**
 * 确认入库
 * @param previewId - 预览 ID
 * @param title - 文档标题（可选）
 * @param tags - 标签（可选）
 * @param scope - 文档范围（可选）
 * @param scopeIds - 范围 ID 数组（可选）
 * @param summary - 文档摘要（可选，可由 LLM 生成后教师修改）
 */
export function confirmDocument(previewId: string, title?: string, tags?: string, scope?: string, scopeIds?: number[], summary?: string) {
  const data: Record<string, any> = { preview_id: previewId }
  if (title) data.title = title
  if (tags) data.tags = tags
  if (scope) data.scope = scope
  if (scopeIds && scopeIds.length > 0) data.scope_ids = scopeIds
  if (summary !== undefined) data.summary = summary
  return request<ConfirmResponse>({
    url: '/api/documents/confirm',
    method: 'POST',
    data,
  })
}

/** 聊天记录导入响应 */
export interface ImportChatResponse {
  document_id: number
  title: string
  doc_type: 'chat'
  conversation_count: number
  chunks_count: number
  status: string
}

/**
 * 聊天记录 JSON 导入知识库（仅教师）
 * @param filePath - 本地 JSON 文件路径
 * @param personaId - 教师分身 ID
 * @param title - 文档标题（可选）
 * @param tags - 标签（JSON 数组格式字符串，可选）
 * @param scope - 范围（可选）
 * @param scopeIds - 范围 ID 列表（可选）
 */
export function importChat(
  filePath: string,
  personaId: number,
  title?: string,
  tags?: string,
  scope?: string,
  scopeIds?: number[],
) {
  const token = getToken()

  const formData: Record<string, string> = {
    persona_id: String(personaId),
  }
  if (title) formData.title = title
  if (tags) formData.tags = tags
  if (scope) formData.scope = scope
  if (scopeIds && scopeIds.length > 0) formData.scope_ids = JSON.stringify(scopeIds)

  return new Promise<ImportChatResponse>((resolve, reject) => {
    Taro.uploadFile({
      url: `${BASE_URL}/api/documents/import-chat`,
      filePath,
      name: 'file',
      formData,
      header: {
        Authorization: `Bearer ${token}`,
      },
      success: (res) => {
        try {
          const data = JSON.parse(res.data)
          if (data.code === 0) {
            resolve(data.data)
          } else {
            Taro.showToast({ title: data.message || '导入失败', icon: 'none' })
            reject(new Error(data.message))
          }
        } catch {
          reject(new Error('解析响应失败'))
        }
      },
      fail: (err) => {
        Taro.showToast({ title: '网络异常', icon: 'none' })
        reject(err)
      },
    })
  })
}
