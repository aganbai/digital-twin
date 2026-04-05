import { request, PaginatedData } from './request'

/** 会话信息（迭代9增强） */
export interface SessionInfo {
  session_id: string
  title: string | null
  last_message: string
  last_message_role: string
  message_count: number
  is_active: boolean
  updated_at: string
}

/** 创建新会话参数 */
export interface CreateSessionParams {
  teacher_persona_id: number
  student_persona_id?: number
}

/** 创建新会话响应 */
export interface CreateSessionResponse {
  session_id: string
  created_at: string
}

/**
 * 创建新会话（迭代9新增）
 */
export function createNewSessionV9(params: CreateSessionParams) {
  return request<CreateSessionResponse>({
    url: '/api/conversations/sessions',
    method: 'POST',
    data: params,
  })
}

/**
 * 获取会话列表（迭代9增强）
 * @param teacherPersonaId - 教师分身ID
 * @param page - 页码
 * @param pageSize - 每页数量
 */
export function getSessionsV9(teacherPersonaId: number, page = 1, pageSize = 20) {
  return request<PaginatedData<SessionInfo>>({
    url: `/api/conversations/sessions?teacher_persona_id=${teacherPersonaId}&page=${page}&page_size=${pageSize}`,
    method: 'GET',
  })
}

/**
 * 异步生成会话标题
 */
export function generateSessionTitle(sessionId: string) {
  return request<{ message: string }>({
    url: `/api/conversations/sessions/${sessionId}/title`,
    method: 'POST',
  })
}

/** 班级详情（学生查看） */
export interface ClassDetailForStudent {
  id: number
  name: string
  subject: string
  description?: string
  teacher_name: string
  member_count: number
  created_at: string
}

/**
 * 获取班级详情（学生查看）
 */
export function getClassDetail(id: number) {
  return request<ClassDetailForStudent>({
    url: `/api/classes/${id}`,
    method: 'GET',
  })
}

/** 学生详情（教师查看） */
export interface StudentProfileForTeacher {
  id: number
  nickname: string
  age?: number
  gender?: string
  family_info?: string
  teacher_evaluation?: string
  class_name?: string
}

/**
 * 获取学生详情（教师查看）
 */
export function getStudentProfile(id: number) {
  return request<StudentProfileForTeacher>({
    url: `/api/students/${id}/profile`,
    method: 'GET',
  })
}

/**
 * 更新学生评语
 */
export function updateStudentEvaluation(id: number, evaluation: string) {
  return request<{ message: string }>({
    url: `/api/students/${id}/evaluation`,
    method: 'PUT',
    data: { evaluation },
  })
}
