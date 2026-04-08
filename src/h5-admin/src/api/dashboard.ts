import { get } from '@/utils/request'

/** 系统总览数据 */
export interface OverviewData {
  total_users: number
  teacher_count: number
  student_count: number
  admin_count: number
  total_conversations: number
  total_messages: number
  knowledge_count: number
  class_count: number
  today_active_users: number
  today_new_users: number
  today_messages: number
}

/** 用户统计数据 */
export interface UserStatsData {
  role_distribution: { role: string; count: number }[]
  register_trend: { date: string; count: number }[]
  activity_distribution: { level: string; count: number }[]
}

/** 对话统计数据 */
export interface ChatStatsData {
  daily_count: { date: string; count: number }[]
  avg_rounds: number
  hourly_distribution: { hour: number; count: number }[]
}

/** 知识库统计数据 */
export interface KnowledgeStatsData {
  type_distribution: { type: string; count: number }[]
  growth_trend: { date: string; count: number }[]
}

/** 活跃用户 */
export interface ActiveUser {
  user_id: number
  nickname: string
  role: string
  avatar?: string
  message_count: number
  last_active_at: string
}

/**
 * 获取系统总览数据
 */
export function getOverview() {
  return get<OverviewData>('/api/admin/dashboard/overview')
}

/**
 * 获取用户统计
 * @param days - 天数（7/30/90）
 */
export function getUserStats(days: number = 30) {
  return get<UserStatsData>('/api/admin/dashboard/user-stats', { days })
}

/**
 * 获取对话统计
 * @param days - 天数（7/30/90）
 */
export function getChatStats(days: number = 30) {
  return get<ChatStatsData>('/api/admin/dashboard/chat-stats', { days })
}

/**
 * 获取知识库统计
 * @param days - 天数（7/30/90）
 */
export function getKnowledgeStats(days: number = 30) {
  return get<KnowledgeStatsData>('/api/admin/dashboard/knowledge-stats', { days })
}

/**
 * 获取活跃用户排行
 * @param days - 天数
 * @param limit - 返回条数
 */
export function getActiveUsers(days: number = 7, limit: number = 10) {
  return get<ActiveUser[]>('/api/admin/dashboard/active-users', { days, limit })
}
