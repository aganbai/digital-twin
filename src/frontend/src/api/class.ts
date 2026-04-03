import { request } from './request'

/** 班级信息 */
export interface ClassInfo {
  id: number
  name: string
  description?: string
  member_count: number
  is_active?: boolean
  created_at: string
}

/** 班级成员 */
export interface ClassMember {
  id: number
  student_persona_id: number
  student_nickname: string
  joined_at: string
}

/** 班级成员分页响应（兼容后端 SuccessPage 格式） */
export interface ClassMemberPageData {
  items?: ClassMember[]
  total?: number
  page?: number
  page_size?: number
}

/**
 * 创建班级
 * @param name - 班级名称
 * @param description - 班级描述（可选）
 */
export function createClass(name: string, description?: string) {
  const data: Record<string, string> = { name }
  if (description) data.description = description
  return request<ClassInfo>({
    url: '/api/classes',
    method: 'POST',
    data,
  })
}

/**
 * 获取班级列表
 */
export function getClasses() {
  return request<ClassInfo[]>({
    url: '/api/classes',
    method: 'GET',
  })
}

/**
 * 更新班级信息
 * @param id - 班级 ID
 * @param data - 更新字段
 */
export function updateClass(
  id: number,
  data: { name?: string; description?: string },
) {
  return request<ClassInfo>({
    url: `/api/classes/${id}`,
    method: 'PUT',
    data,
  })
}

/**
 * 删除班级
 * @param id - 班级 ID
 */
export function deleteClass(id: number) {
  return request<{ message: string }>({
    url: `/api/classes/${id}`,
    method: 'DELETE',
  })
}

/**
 * 获取班级成员列表
 * @param classId - 班级 ID
 */
export function getClassMembers(classId: number) {
  return request<ClassMemberPageData | ClassMember[]>({
    url: `/api/classes/${classId}/members`,
    method: 'GET',
  })
}

/**
 * 添加班级成员
 * @param classId - 班级 ID
 * @param studentPersonaId - 学生分身 ID
 */
export function addClassMember(classId: number, studentPersonaId: number) {
  return request<ClassMember>({
    url: `/api/classes/${classId}/members`,
    method: 'POST',
    data: { student_persona_id: studentPersonaId },
  })
}

/**
 * 移除班级成员
 * @param classId - 班级 ID
 * @param memberId - 成员记录 ID
 */
export function removeClassMember(classId: number, memberId: number) {
  return request<{ message: string }>({
    url: `/api/classes/${classId}/members/${memberId}`,
    method: 'DELETE',
  })
}

/**
 * 启停班级
 * @param classId - 班级 ID
 * @param isActive - 是否启用
 */
export function toggleClass(classId: number, isActive: boolean) {
  return request<{ message: string }>({
    url: `/api/classes/${classId}/toggle`,
    method: 'PUT',
    data: { is_active: isActive },
  })
}

// ===================== 迭代8 新增 =====================

/** V8 班级创建请求体 */
export interface CreateClassV8Params {
  teacher_display_name: string
  subject: string
  age_group: string[]
  name: string
  description?: string
}

/** V8 班级创建响应 */
export interface ClassInfoV8 {
  id: number
  teacher_display_name: string
  subject: string
  age_group: string[]
  name: string
  description?: string
  member_count: number
  is_active?: boolean
  created_at: string
}

/** 班级分享信息 */
export interface ClassShareInfo {
  share_link: string
  invite_code: string
  qr_code_url: string
}

/** 待审批加入请求 */
export interface JoinRequestItem {
  id: number
  student_avatar?: string
  student_nickname: string
  request_time: string
  status: string
  /** 学生详细信息（审批详情中使用） */
  age?: number
  gender?: string
  family_info?: string
}

/**
 * V8 创建班级
 */
export function createClassV8(params: CreateClassV8Params) {
  return request<ClassInfoV8>({
    url: '/api/classes/v8',
    method: 'POST',
    data: params,
  })
}

/**
 * 获取班级分享信息
 * @param classId - 班级 ID
 */
export function getClassShareInfo(classId: number) {
  return request<ClassShareInfo>({
    url: `/api/classes/${classId}/share-info`,
    method: 'GET',
  })
}

/**
 * 获取待审批加入请求列表
 */
export function getPendingJoinRequests() {
  return request<JoinRequestItem[]>({
    url: '/api/join-requests/pending',
    method: 'GET',
  })
}

/** 审批通过请求体 */
export interface ApproveJoinParams {
  teacher_evaluation?: string
  age?: number
  gender?: string
  family_info?: string
}

/**
 * 审批通过加入请求
 * @param id - 加入请求 ID
 * @param params - 审批信息
 */
export function approveJoinRequest(id: number, params: ApproveJoinParams) {
  return request<{ message: string }>({
    url: `/api/join-requests/${id}/approve`,
    method: 'PUT',
    data: params,
  })
}

/**
 * 拒绝加入请求
 * @param id - 加入请求 ID
 */
export function rejectJoinRequest(id: number) {
  return request<{ message: string }>({
    url: `/api/join-requests/${id}/reject`,
    method: 'PUT',
  })
}
