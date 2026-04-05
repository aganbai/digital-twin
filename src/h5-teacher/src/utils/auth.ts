// Token 存储 Key
const TOKEN_KEY = 'h5_teacher_token'
const USER_INFO_KEY = 'h5_teacher_user_info'

export interface UserInfo {
  id: number
  role: string
  nickname: string
  avatar?: string
}

/**
 * 获取 Token
 */
export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

/**
 * 保存 Token
 */
export function saveToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

/**
 * 获取用户信息
 */
export function getUserInfo(): UserInfo | null {
  const info = localStorage.getItem(USER_INFO_KEY)
  return info ? JSON.parse(info) : null
}

/**
 * 保存用户信息
 */
export function saveUserInfo(info: UserInfo): void {
  localStorage.setItem(USER_INFO_KEY, JSON.stringify(info))
}

/**
 * 保存认证信息（Token + 用户信息）
 */
export function saveAuthInfo(token: string, userInfo: UserInfo): void {
  saveToken(token)
  saveUserInfo(userInfo)
}

/**
 * 清除认证信息
 */
export function clearAuthInfo(): void {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_INFO_KEY)
}

/**
 * 检查是否已登录
 */
export function isLoggedIn(): boolean {
  return !!getToken()
}

/**
 * 处理微信回调参数
 * 从 URL 中提取 code 和 state 参数
 */
export function handleWxCallback(): { code: string; state: string } | null {
  const urlParams = new URLSearchParams(window.location.search)
  const code = urlParams.get('code')
  const state = urlParams.get('state')
  
  if (code && state) {
    return { code, state }
  }
  return null
}

/**
 * 清除 URL 中的微信授权参数
 */
export function clearAuthParams(): void {
  const url = new URL(window.location.href)
  url.searchParams.delete('code')
  url.searchParams.delete('state')
  window.history.replaceState({}, '', url.toString())
}
