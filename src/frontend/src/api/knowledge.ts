import { request, PaginatedData } from './request'

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

/** 删除文档响应 */
export interface DeleteDocumentResponse {
  document_id: number
  deleted: boolean
}

/**
 * 获取文档列表
 * @param page - 页码，默认 1
 * @param pageSize - 每页数量，默认 20
 */
export function getDocuments(page = 1, pageSize = 20) {
  return request<PaginatedData<Document>>({
    url: `/api/documents?page=${page}&page_size=${pageSize}`,
    method: 'GET',
  })
}

/**
 * 添加知识文档
 * @param title - 文档标题
 * @param content - 文档内容
 * @param tags - 分类标签
 */
export function addDocument(title: string, content: string, tags: string[] = []) {
  return request<AddDocumentResponse>({
    url: '/api/documents',
    method: 'POST',
    data: { title, content, tags },
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
