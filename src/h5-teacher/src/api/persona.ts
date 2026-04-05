import { get, post, put, del } from '@/utils/request'

/** 分身信息（迭代11：班级专属分身） */
export interface Persona {
  id: number
  role: 'teacher' | 'student'
  nickname: string
  school?: string
  description?: string
  is_active: boolean
  is_public?: boolean
  /** 绑定的班级ID */
  bound_class_id?: number
  /** 绑定的班级名称 */
  bound_class_name?: string
  /** 学生数 */
  student_count?: number
  /** 文档数 */
  document_count?: number
  created_at: string
}

/** 分身列表响应 */
export interface PersonaListResponse {
  personas: Persona[]
  current_persona_id: number
}

/** 仪表盘数据 */
export interface DashboardData {
  persona: Persona
  pending_count: number
  classes: Array<{ id: number; name: string; member_count: number; is_active: boolean }>
  latest_share: { share_code: string; class_name: string; used_count: number; max_uses: number; is_active: boolean } | null
  stats: { total_students: number; total_documents: number; total_classes: number }
}

/**
 * 获取分身列表
 */
export function getPersonaList() {
  return get<PersonaListResponse>('/api/personas')
}

/**
 * 获取分身详情
 */
export function getPersonaDetail(personaId: number) {
  return get<Persona>(`/api/personas/${personaId}`)
}

/**
 * 获取教师仪表盘聚合数据
 */
export function getPersonaDashboard(personaId: number) {
  return get<DashboardData>(`/api/personas/${personaId}/dashboard`)
}

/**
 * 设置分身公开/私有
 */
export function setVisibility(id: number, isPublic: boolean) {
  return put<{ id: number; nickname: string; is_public: boolean; updated_at: string }>(
    `/api/personas/${id}/visibility`,
    { is_public: isPublic }
  )
}

/**
 * 更新分身信息
 */
export function updatePersona(personaId: number, params: { nickname?: string; school?: string; description?: string }) {
  return put<Persona>(`/api/personas/${personaId}`, params)
}
