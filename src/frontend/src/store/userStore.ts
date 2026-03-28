import { create } from 'zustand'
import { getToken, setToken as saveToken, removeToken } from '../utils/storage'
import {
  getUserInfo as getStoredUserInfo,
  setUserInfo as saveUserInfo,
  removeUserInfo,
  StoredUserInfo,
} from '../utils/storage'

/** 用户状态 */
interface UserState {
  /** JWT Token */
  token: string
  /** 用户信息 */
  userInfo: StoredUserInfo | null
  /** 是否已登录 */
  isLoggedIn: boolean

  /** 设置 Token（同时持久化到 Storage） */
  setToken: (token: string) => void
  /** 设置用户信息（同时持久化到 Storage） */
  setUserInfo: (info: StoredUserInfo) => void
  /** 退出登录（清除所有登录态） */
  logout: () => void
}

export const useUserStore = create<UserState>((set) => ({
  // 初始化时从 Storage 恢复
  token: getToken(),
  userInfo: getStoredUserInfo(),
  isLoggedIn: !!getToken(),

  setToken: (token: string) => {
    saveToken(token)
    set({ token, isLoggedIn: !!token })
  },

  setUserInfo: (info: StoredUserInfo) => {
    saveUserInfo(info)
    set({ userInfo: info })
  },

  logout: () => {
    removeToken()
    removeUserInfo()
    set({ token: '', userInfo: null, isLoggedIn: false })
  },
}))
