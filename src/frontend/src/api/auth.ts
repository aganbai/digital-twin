import Taro from '@tarojs/taro'
import { request } from './request'

/** 分身基础信息（登录返回） */
export interface LoginPersona {
  id: number
  role: 'teacher' | 'student'
  nickname: string
  school?: string
  description?: string
  is_active: boolean
}

/** 微信登录响应 */
export interface WxLoginResponse {
  user_id: number
  token: string
  role: string
  nickname: string
  is_new_user: boolean
  expires_at: string
  personas?: LoginPersona[]
  default_persona_id?: number
}

/** 补全信息响应 */
export interface CompleteProfileResponse {
  user_id: number
  role: string
  nickname: string
  school?: string
  description?: string
  token?: string
  expires_at?: string
}

/**
 * 微信登录
 * 调用 wx.login 获取 code，发送到后端换取 token
 */
export async function wxLogin() {
  const { code } = await Taro.login()
  return request<WxLoginResponse>({
    url: '/api/auth/wx-login',
    method: 'POST',
    data: { code },
  })
}

/**
 * 新用户补全信息（角色 + 昵称 + 教师额外信息）
 * @param role - 角色：teacher / student
 * @param nickname - 昵称
 * @param school - 学校名称（教师必填）
 * @param description - 分身描述（教师必填）
 */
export function completeProfile(
  role: string,
  nickname: string,
  school?: string,
  description?: string,
) {
  const data: Record<string, string> = { role, nickname }
  if (school !== undefined) data.school = school
  if (description !== undefined) data.description = description
  return request<CompleteProfileResponse>({
    url: '/api/auth/complete-profile',
    method: 'POST',
    data,
  })
}
