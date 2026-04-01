import Taro from '@tarojs/taro'
import { request, PaginatedData } from './request'
import { getToken } from '../utils/storage'
import { BASE_URL } from '../utils/constants'

/** 作业点评 */
export interface AssignmentReview {
  id: number
  reviewer_type: 'ai' | 'teacher'
  reviewer_id: number | null
  content: string
  score: number | null
  created_at: string
}

/** 作业列表项 */
export interface AssignmentListItem {
  id: number
  student_id: number
  student_nickname: string
  teacher_id: number
  teacher_nickname: string
  title: string
  status: 'submitted' | 'reviewed'
  has_file: boolean
  review_count: number
  created_at: string
}

/** 作业详情 */
export interface AssignmentDetail {
  id: number
  student_id: number
  student_nickname: string
  teacher_id: number
  teacher_nickname: string
  title: string
  content: string
  file_path: string
  file_type: string
  status: 'submitted' | 'reviewed'
  created_at: string
  reviews: AssignmentReview[]
}

/** 提交作业响应 */
export interface SubmitAssignmentResponse {
  id: number
  student_id: number
  teacher_id: number
  title: string
  status: string
  created_at: string
}

/** 点评响应 */
export interface ReviewResponse {
  id: number
  assignment_id: number
  reviewer_type: string
  reviewer_id: number | null
  content: string
  score: number | null
  created_at: string
  token_usage?: {
    prompt_tokens: number
    completion_tokens: number
    total_tokens: number
  }
}

/** 获取作业列表参数 */
export interface GetAssignmentsParams {
  teacher_id?: number
  student_id?: number
  status?: string
  page?: number
  page_size?: number
}

/**
 * 学生提交作业（JSON 模式，不含文件）
 * @param data - 作业数据
 */
export function submitAssignment(data: { teacher_id: number; title: string; content?: string }) {
  return request<SubmitAssignmentResponse>({
    url: '/api/assignments',
    method: 'POST',
    data,
  })
}

/**
 * 学生提交作业（含文件上传，multipart/form-data）
 * @param filePath - 本地文件路径
 * @param teacherId - 教师 ID
 * @param title - 作业标题
 * @param content - 文本内容（可选）
 */
export function submitAssignmentWithFile(
  filePath: string,
  teacherId: number,
  title: string,
  content?: string,
) {
  const token = getToken()
  const formData: Record<string, string> = {
    teacher_id: String(teacherId),
    title,
  }
  if (content) {
    formData.content = content
  }
  return new Promise<SubmitAssignmentResponse>((resolve, reject) => {
    Taro.uploadFile({
      url: `${BASE_URL}/api/assignments`,
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
            Taro.showToast({ title: data.message || '提交失败', icon: 'none' })
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
 * 获取作业列表
 * @param params - 查询参数
 */
export function getAssignments(params: GetAssignmentsParams = {}) {
  const query = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null) {
      query.append(key, String(value))
    }
  })
  const queryStr = query.toString()
  return request<PaginatedData<AssignmentListItem>>({
    url: `/api/assignments${queryStr ? '?' + queryStr : ''}`,
    method: 'GET',
  })
}

/**
 * 获取作业详情
 * @param id - 作业 ID
 */
export function getAssignmentDetail(id: number) {
  return request<AssignmentDetail>({
    url: `/api/assignments/${id}`,
    method: 'GET',
  })
}

/**
 * 教师点评作业
 * @param id - 作业 ID
 * @param data - 点评数据
 */
export function reviewAssignment(id: number, data: { content: string; score?: number }) {
  return request<ReviewResponse>({
    url: `/api/assignments/${id}/review`,
    method: 'POST',
    data,
  })
}

/**
 * AI 自动点评作业
 * @param id - 作业 ID
 */
export function aiReviewAssignment(id: number) {
  return request<ReviewResponse>({
    url: `/api/assignments/${id}/ai-review`,
    method: 'POST',
  })
}
