import { request } from './request'

// ========== 学生端聊天列表 ==========

/** 学生端老师聊天列表项 */
export interface StudentTeacherChatItem {
  teacher_persona_id: number
  teacher_nickname: string
  teacher_avatar: string
  teacher_school?: string
  subject?: string
  last_message?: string
  last_message_time?: string
  unread_count: number
  is_pinned: boolean
}

/** 学生端聊天列表响应 */
export interface StudentChatListResponse {
  teachers: StudentTeacherChatItem[]
  total: number
}

/**
 * 获取学生端聊天列表（按老师分组）
 */
export function getStudentChatList() {
  return request<StudentChatListResponse>({
    url: '/api/chat-list/student',
    method: 'GET',
  })
}

// ========== 教师端聊天列表 ==========

/** 教师端学生聊天信息 */
export interface TeacherChatStudent {
  student_persona_id: number
  student_nickname: string
  student_avatar: string
  last_message?: string
  last_message_time?: string
  unread_count: number
  is_pinned: boolean
}

/** 教师端班级聊天项 */
export interface TeacherChatClassItem {
  class_id: number
  class_name: string
  subject?: string
  students: TeacherChatStudent[]
  is_pinned: boolean
  pin_time?: string
}

/** 教师端聊天列表响应 */
export interface TeacherChatListResponse {
  classes: TeacherChatClassItem[]
  total: number
}

/**
 * 获取教师端聊天列表（按班级组织）
 */
export function getTeacherChatList() {
  return request<TeacherChatListResponse>({
    url: '/api/chat-list/teacher',
    method: 'GET',
  })
}

// ========== 置顶功能 ==========

/** 置顶请求参数 */
export interface PinChatParams {
  target_type: 'teacher' | 'student' | 'class'
  target_id: number
}

/** 置顶项 */
export interface PinnedChatItem {
  id: number
  target_type: string
  target_id: number
  target_name: string
  avatar?: string
  pinned_at: string
}

/** 置顶列表响应 */
export interface PinnedChatsResponse {
  pins: PinnedChatItem[]
  total: number
}

/**
 * 置顶聊天
 */
export function pinChat(params: PinChatParams) {
  return request<{ pin_id: number; success: boolean }>({
    url: '/api/chat-pins',
    method: 'POST',
    data: params,
  })
}

/**
 * 取消置顶
 * @param targetType - 置顶类型
 * @param targetId - 目标 ID
 */
export function unpinChat(targetType: string, targetId: number) {
  return request<{ success: boolean }>({
    url: `/api/chat-pins/${targetType}/${targetId}`,
    method: 'DELETE',
  })
}

/**
 * 获取置顶列表
 */
export function getPinnedChats() {
  return request<PinnedChatsResponse>({
    url: '/api/chat-pins',
    method: 'GET',
  })
}

// ========== 新建会话 ==========

/** 新建会话请求参数 */
export interface NewSessionParams {
  teacher_persona_id: number
  initial_message?: string
}

/** 新建会话响应 */
export interface NewSessionResponse {
  session_id: string
  created_at: string
}

/**
 * 新建会话
 */
export function createNewSession(params: NewSessionParams) {
  return request<NewSessionResponse>({
    url: '/api/chat/new-session',
    method: 'POST',
    data: params,
  })
}

// ========== 快捷指令 ==========

/** 快捷指令项 */
export interface QuickActionItem {
  id: string
  label: string
  action: string
}

/**
 * 获取快捷指令列表
 * @param teacherPersonaId - 教师分身 ID
 */
export function getQuickActions(teacherPersonaId: number) {
  return request<QuickActionItem[]>({
    url: `/api/chat/quick-actions?teacher_persona_id=${teacherPersonaId}`,
    method: 'GET',
  })
}
