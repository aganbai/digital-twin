/** 获取存储的 Token */
export function getToken(): string | null {
  return localStorage.getItem('token')
}

/** 保存 Token */
export function setToken(token: string): void {
  localStorage.setItem('token', token)
}

/** 获取存储的用户信息 */
export function getUserInfo(): any {
  const info = localStorage.getItem('userInfo')
  return info ? JSON.parse(info) : null
}

/** 保存用户信息 */
export function setUserInfo(info: any): void {
  localStorage.setItem('userInfo', JSON.stringify(info))
}

/** 保存认证信息 */
export function saveAuthInfo(token: string, userInfo: any): void {
  setToken(token)
  setUserInfo(userInfo)
}

/** 清除认证信息 */
export function clearAuthInfo(): void {
  localStorage.removeItem('token')
  localStorage.removeItem('userInfo')
}

/** 检查是否已登录 */
export function isLoggedIn(): boolean {
  return !!getToken()
}

/** 处理微信回调参数 */
export function handleWxCallback(): { code: string; state: string } | null {
  const urlParams = new URLSearchParams(window.location.search)
  const code = urlParams.get('code')
  const state = urlParams.get('state')
  if (code && state) {
    return { code, state }
  }
  return null
}

/** 清除 URL 中的微信授权参数 */
export function clearAuthParams(): void {
  const url = new URL(window.location.href)
  url.searchParams.delete('code')
  url.searchParams.delete('state')
  window.history.replaceState({}, '', url.toString())
}
