import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getUserInfo, clearAuthInfo } from '@/utils/auth'

export interface UserInfo {
  id: number
  role: string
  nickname: string
  avatar?: string
  status: string
}

export const useUserStore = defineStore('user', () => {
  // 状态
  const token = ref<string | null>(localStorage.getItem('token'))
  const userInfo = ref<UserInfo | null>(getUserInfo())

  // 计算属性
  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => userInfo.value?.role === 'admin')
  const nickname = computed(() => userInfo.value?.nickname || '未登录')

  // 设置用户信息
  function setUserInfo(info: UserInfo) {
    userInfo.value = info
    localStorage.setItem('userInfo', JSON.stringify(info))
  }

  // 设置 token
  function setToken(t: string) {
    token.value = t
    localStorage.setItem('token', t)
  }

  // 登出
  function logout() {
    token.value = null
    userInfo.value = null
    clearAuthInfo()
  }

  return {
    token,
    userInfo,
    isLoggedIn,
    isAdmin,
    nickname,
    setUserInfo,
    setToken,
    logout,
  }
})
