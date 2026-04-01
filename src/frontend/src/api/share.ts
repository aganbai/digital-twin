import { request } from './request'

/** 分享码信息 */
export interface ShareInfo {
  id: number
  share_code: string
  class_id?: number
  expires_at: string
  max_uses: number
  used_count: number
  is_active: boolean
}

/** 分享码详情（学生视角） */
export interface ShareDetail {
  teacher_persona_id: number
  teacher_nickname: string
  teacher_school?: string
  teacher_description?: string
  class_name?: string
  /** 目标学生分身 ID（0 表示不限定） */
  target_student_persona_id?: number
  /** 目标学生昵称（不限定时为空） */
  target_student_nickname?: string
  is_valid: boolean
  /** 迭代6新增：加入状态 */
  join_status?: 'can_join' | 'already_joined' | 'not_target' | 'need_login' | 'need_persona'
}

/** 加入分享响应 */
export interface JoinShareResponse {
  teacher_persona: {
    id: number
    nickname: string
    school?: string
    description?: string
  }
  class?: {
    id: number
    name: string
  }
}

/**
 * 创建分享码
 * @param teacherPersonaId - 教师分身 ID
 * @param classId - 班级 ID（可选）
 * @param expiresHours - 过期时间（小时，可选）
 * @param maxUses - 最大使用次数（可选）
 * @param targetStudentPersonaId - 目标学生分身 ID（可选，0=不限定）
 */
export function createShare(
  teacherPersonaId: number,
  classId?: number,
  expiresHours?: number,
  maxUses?: number,
  targetStudentPersonaId?: number,
) {
  const data: Record<string, any> = { teacher_persona_id: teacherPersonaId }
  if (classId !== undefined) data.class_id = classId
  if (expiresHours !== undefined) data.expires_hours = expiresHours
  if (maxUses !== undefined) data.max_uses = maxUses
  if (targetStudentPersonaId !== undefined && targetStudentPersonaId > 0) {
    data.target_student_persona_id = targetStudentPersonaId
  }
  return request<ShareInfo>({
    url: '/api/shares',
    method: 'POST',
    data,
  })
}

/**
 * 获取分享码列表
 */
export function getShares() {
  return request<ShareInfo[]>({
    url: '/api/shares',
    method: 'GET',
  })
}

/**
 * 获取分享码详情（学生查看）
 * @param code - 分享码
 */
export function getShareInfo(code: string) {
  return request<ShareDetail>({
    url: `/api/shares/${code}/info`,
    method: 'GET',
  })
}

/**
 * 通过分享码加入
 * @param code - 分享码
 * @param studentPersonaId - 学生分身 ID（可选）
 */
export function joinShare(code: string, studentPersonaId?: number) {
  const data: Record<string, any> = {}
  if (studentPersonaId !== undefined) data.student_persona_id = studentPersonaId
  return request<JoinShareResponse>({
    url: `/api/shares/${code}/join`,
    method: 'POST',
    data,
  })
}

/**
 * 停用分享码
 * @param id - 分享码 ID
 */
export function deactivateShare(id: number) {
  return request<{ message: string }>({
    url: `/api/shares/${id}/deactivate`,
    method: 'PUT',
  })
}
