import { get, post } from '@/utils/request'

/** 微信 H5 登录 URL 响应 */
export interface WxH5LoginUrlResponse {
  login_url: string
  state: string
}

/** 微信 H5 登录回调响应 */
export interface WxH5CallbackResponse {
  user_id: number
  token: string
  role: string
  nickname: string
  avatar?: string
  is_new_user: boolean
  expires_at: string
}

/**
 * 获取微信 H5 登录 URL
 */
export function getWxH5LoginUrl(redirectUri: string) {
  return get<WxH5LoginUrlResponse>('/api/auth/wx-h5-login-url', { redirect_uri: redirectUri })
}

/**
 * 微信 H5 登录回调
 */
export function wxH5Callback(code: string, state: string) {
  return post<WxH5CallbackResponse>('/api/auth/wx-h5-callback', { code, state })
}

/**
 * 补全用户信息
 */
export function completeProfile(role: string, nickname: string) {
  return post('/api/auth/complete-profile', { role, nickname })
}
