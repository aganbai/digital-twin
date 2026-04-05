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
  const token = ref<string | null>(localStorage.getItem('token'))
  const userInfo = ref<UserInfo | null>(getUserInfo())

  const isLoggedIn = computed(() => !!token.value)
  const isStudent = computed(() => userInfo.value?.role === 'student')
  const nickname = computed(() => userInfo.value?.nickname || '未登录')

  function setUserInfo(info: UserInfo) {
    userInfo.value = info
    localStorage.setItem('userInfo', JSON.stringify(info))
  }

  function setToken(t: string) {
    token.value = t
    localStorage.setItem('token', t)
  }

  function logout() {
    token.value = null
    userInfo.value = null
    clearAuthInfo()
  }

  return { token, userInfo, isLoggedIn, isStudent, nickname, setUserInfo, setToken, logout }
})
