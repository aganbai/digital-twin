import request from '@/utils/request'

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
