import Taro from '@tarojs/taro'

const TOKEN_KEY = 'token'
const USER_INFO_KEY = 'userInfo'

/** 获取 Token */
export function getToken(): string {
  return Taro.getStorageSync(TOKEN_KEY) || ''
}

/** 设置 Token */
export function setToken(token: string): void {
  Taro.setStorageSync(TOKEN_KEY, token)
}

/** 移除 Token */
export function removeToken(): void {
  Taro.removeStorageSync(TOKEN_KEY)
}

/** 用户信息类型 */
export interface StoredUserInfo {
  id: number
  nickname: string
  role: string
}

/** 获取用户信息 */
export function getUserInfo(): StoredUserInfo | null {
  const info = Taro.getStorageSync(USER_INFO_KEY)
  return info ? info : null
}

/** 设置用户信息 */
export function setUserInfo(info: StoredUserInfo): void {
  Taro.setStorageSync(USER_INFO_KEY, info)
}

/** 移除用户信息 */
export function removeUserInfo(): void {
  Taro.removeStorageSync(USER_INFO_KEY)
}
