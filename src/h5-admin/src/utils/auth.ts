/**
 * 微信 H5 登录相关工具函数
 */

/** 微信授权回调处理 */
export function handleWxCallback(): { code: string; state: string } | null {
  const urlParams = new URLSearchParams(window.location.search)
  const code = urlParams.get('code')
  const state = urlParams.get('state') || ''

  if (code) {
    return { code, state }
  }
  return null
}

/** 清除 URL 中的授权参数 */
export function clearAuthParams() {
  const url = new URL(window.location.href)
  url.searchParams.delete('code')
  url.searchParams.delete('state')
  window.history.replaceState({}, document.title, url.toString())
}

/** 保存登录信息 */
export function saveAuthInfo(token: string, userInfo: any) {
  localStorage.setItem('token', token)
  localStorage.setItem('userInfo', JSON.stringify(userInfo))
}

/** 清除登录信息 */
export function clearAuthInfo() {
  localStorage.removeItem('token')
  localStorage.removeItem('userInfo')
}

/** 获取 token */
export function getToken(): string | null {
  return localStorage.getItem('token')
}

/** 获取用户信息 */
export function getUserInfo(): any {
  const userInfo = localStorage.getItem('userInfo')
  return userInfo ? JSON.parse(userInfo) : null
}

/** 检查是否已登录 */
export function isLoggedIn(): boolean {
  return !!getToken()
}

/** 获取当前页面 URL（用于微信回调） */
export function getCurrentUrl(): string {
  return window.location.origin + window.location.pathname
}
