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
  const token = ref<string | null>(localStorage.getItem('h5_teacher_token'))
  const userInfo = ref<UserInfo | null>(getUserInfo())

  const isLoggedIn = computed(() => !!token.value)
  const isTeacher = computed(() => userInfo.value?.role === 'teacher')
  const nickname = computed(() => userInfo.value?.nickname || '未登录')

  function setUserInfo(info: UserInfo) {
    userInfo.value = info
    localStorage.setItem('h5_teacher_user_info', JSON.stringify(info))
  }

  function setToken(t: string) {
    token.value = t
    localStorage.setItem('h5_teacher_token', t)
  }

  function logout() {
    token.value = null
    userInfo.value = null
    clearAuthInfo()
  }

  return {
    token,
    userInfo,
    isLoggedIn,
    isTeacher,
    nickname,
    setUserInfo,
    setToken,
    logout,
  }
})
