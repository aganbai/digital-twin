import Taro from '@tarojs/taro'
import { request } from './request'

/** 微信登录响应 */
export interface WxLoginResponse {
  user_id: number
  token: string
  role: string
  nickname: string
  is_new_user: boolean
  expires_at: string
}

/** 补全信息响应 */
export interface CompleteProfileResponse {
  user_id: number
  role: string
  nickname: string
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
 * 新用户补全信息（角色 + 昵称）
 * @param role - 角色：teacher / student
 * @param nickname - 昵称
 */
export function completeProfile(role: string, nickname: string) {
  return request<CompleteProfileResponse>({
    url: '/api/auth/complete-profile',
    method: 'POST',
    data: { role, nickname },
  })
}
