import { request, PaginatedData } from './request'

/** 师生关系项（教师视角） */
export interface RelationItemTeacher {
  id: number
  student_id: number
  student_persona_id?: number
  student_nickname: string
  status: 'pending' | 'approved' | 'rejected'
  initiated_by: 'teacher' | 'student'
  is_active?: boolean
  created_at: string
}

/** 师生关系项（学生视角） */
export interface RelationItemStudent {
  id: number
  teacher_id: number
  teacher_persona_id?: number
  teacher_nickname: string
  teacher_school: string
  teacher_description: string
  status: 'pending' | 'approved' | 'rejected'
  initiated_by: 'teacher' | 'student'
  created_at: string
}

/** 关系操作响应 */
export interface RelationResponse {
  id: number
  teacher_id: number
  teacher_persona_id?: number
  student_id: number
  student_persona_id?: number
  status: string
  initiated_by: string
  created_at: string
  updated_at?: string
}

/** 获取关系列表参数 */
export interface GetRelationsParams {
  status?: string
  page?: number
  page_size?: number
}

/**
 * 教师邀请学生
 * @param studentId - 学生 ID
 */
export function inviteStudent(studentId: number) {
  return request<RelationResponse>({
    url: '/api/relations/invite',
    method: 'POST',
    data: { student_id: studentId },
  })
}

/**
 * 学生申请使用教师分身
 * @param teacherId - 教师用户 ID
 * @param teacherPersonaId - 教师分身 ID（广场申请时传入）
 */
export function applyTeacher(teacherId: number, teacherPersonaId?: number) {
  const data: Record<string, number> = { teacher_id: teacherId }
  if (teacherPersonaId) {
    data.teacher_persona_id = teacherPersonaId
  }
  return request<RelationResponse>({
    url: '/api/relations/apply',
    method: 'POST',
    data,
  })
}

/**
 * 教师审批同意
 * @param id - 关系记录 ID
 */
export function approveRelation(id: number) {
  return request<RelationResponse>({
    url: `/api/relations/${id}/approve`,
    method: 'PUT',
  })
}

/**
 * 教师审批拒绝
 * @param id - 关系记录 ID
 */
export function rejectRelation(id: number) {
  return request<RelationResponse>({
    url: `/api/relations/${id}/reject`,
    method: 'PUT',
  })
}

/**
 * 获取师生关系列表
 * @param params - 查询参数
 */
export function getRelations(params: GetRelationsParams = {}) {
  const query = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null) {
      query.append(key, String(value))
    }
  })
  const queryStr = query.toString()
  return request<PaginatedData<RelationItemTeacher | RelationItemStudent>>({
    url: `/api/relations${queryStr ? '?' + queryStr : ''}`,
    method: 'GET',
  })
}

/**
 * 启停学生访问权限
 * @param relationId - 关系记录 ID
 * @param isActive - 是否启用
 */
export function toggleRelation(relationId: number, isActive: boolean) {
  return request<{ message: string }>({
    url: `/api/relations/${relationId}/toggle`,
    method: 'PUT',
    data: { is_active: isActive },
  })
}
