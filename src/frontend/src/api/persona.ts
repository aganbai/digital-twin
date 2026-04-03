import { request, PaginatedData } from './request'

/** 分身信息 */
export interface Persona {
  id: number
  role: 'teacher' | 'student'
  nickname: string
  school?: string
  description?: string
  is_active: boolean
  is_public?: boolean
  created_at: string
}

/** 分身列表响应 */
export interface PersonaListResponse {
  personas: Persona[]
  default_persona_id: number
}

/** 切换分身响应（后端返回扁平结构） */
export interface SwitchPersonaResponse {
  token: string
  persona_id: number
  role: 'teacher' | 'student'
  nickname: string
  school?: string
  description?: string
  expires_at: string
}

/**
 * 创建分身
 * @param role - 角色：teacher / student
 * @param nickname - 昵称
 * @param school - 学校（教师可选）
 * @param description - 描述（教师可选）
 */
export function createPersona(
  role: string,
  nickname: string,
  school?: string,
  description?: string,
) {
  const data: Record<string, string> = { role, nickname }
  if (school !== undefined) data.school = school
  if (description !== undefined) data.description = description
  return request<Persona>({
    url: '/api/personas',
    method: 'POST',
    data,
  })
}

/**
 * 获取分身列表
 */
export function getPersonas() {
  return request<PersonaListResponse>({
    url: '/api/personas',
    method: 'GET',
  })
}

/**
 * 编辑分身
 * @param id - 分身 ID
 * @param data - 更新字段
 */
export function updatePersona(
  id: number,
  data: { nickname?: string; school?: string; description?: string },
) {
  return request<Persona>({
    url: `/api/personas/${id}`,
    method: 'PUT',
    data,
  })
}

/**
 * 激活分身
 * @param id - 分身 ID
 */
export function activatePersona(id: number) {
  return request<{ message: string }>({
    url: `/api/personas/${id}/activate`,
    method: 'PUT',
  })
}

/**
 * 停用分身
 * @param id - 分身 ID
 */
export function deactivatePersona(id: number) {
  return request<{ message: string }>({
    url: `/api/personas/${id}/deactivate`,
    method: 'PUT',
  })
}

/**
 * 切换分身
 * @param id - 分身 ID
 */
export function switchPersona(id: number) {
  return request<SwitchPersonaResponse>({
    url: `/api/personas/${id}/switch`,
    method: 'PUT',
  })
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
 * 获取教师仪表盘聚合数据
 * @param personaId - 分身 ID
 */
export function getPersonaDashboard(personaId: number) {
  return request<DashboardData>({
    url: `/api/personas/${personaId}/dashboard`,
    method: 'GET',
  })
}

/** 广场中的教师分身 */
export interface MarketplacePersona {
  id: number
  nickname: string
  school: string
  description: string
  student_count: number
  document_count: number
  /** 当前学生对该教师的申请状态：空字符串（未申请）/ pending（申请中） */
  application_status: '' | 'pending' | 'approved'
}

/** 广场列表查询参数 */
export interface MarketplaceParams {
  keyword?: string
  page?: number
  page_size?: number
}

/** 设置公开/私有响应 */
export interface SetVisibilityResponse {
  id: number
  nickname: string
  is_public: boolean
  updated_at: string
}

/** 搜索学生结果 */
export interface StudentSearchResult {
  persona_id: number
  user_id: number
  nickname: string
  created_at: string
}

/**
 * 获取分身广场列表（公开的教师分身）
 * @param params - 查询参数
 */
export function getMarketplace(params?: MarketplaceParams) {
  const query = new URLSearchParams()
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        query.append(key, String(value))
      }
    })
  }
  const qs = query.toString()
  return request<PaginatedData<MarketplacePersona>>({
    url: `/api/personas/marketplace${qs ? `?${qs}` : ''}`,
    method: 'GET',
  })
}

/**
 * 设置分身公开/私有
 * @param id - 分身 ID
 * @param isPublic - 是否公开到广场
 */
export function setVisibility(id: number, isPublic: boolean) {
  return request<SetVisibilityResponse>({
    url: `/api/personas/${id}/visibility`,
    method: 'PUT',
    data: { is_public: isPublic },
  })
}

/**
 * 搜索已注册学生
 * @param keyword - 搜索关键词（匹配学生昵称，至少 2 个字符）
 * @param page - 页码
 * @param pageSize - 每页数量
 */
export function searchStudents(keyword: string, page = 1, pageSize = 20) {
  return request<PaginatedData<StudentSearchResult>>({
    url: `/api/students/search?keyword=${encodeURIComponent(keyword)}&page=${page}&page_size=${pageSize}`,
    method: 'GET',
  })
}
