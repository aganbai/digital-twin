<template>
  <div class="login-page">
    <div class="login-header">
      <h1>数字分身</h1>
      <p>学生端</p>
    </div>
    <div class="login-content">
      <van-loading v-if="loading" size="40">登录中...</van-loading>
      <template v-else-if="error">
        <van-icon name="warning-o" size="40" color="#ee0a24" />
        <p class="error-text">{{ error }}</p>
        <van-button type="primary" block @click="handleLogin">重新登录</van-button>
      </template>
      <template v-else>
        <van-button type="primary" block @click="handleLogin">微信授权登录</van-button>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { getWxH5LoginUrl, wxH5Callback } from '@/api/auth'
import { saveAuthInfo, handleWxCallback, clearAuthParams, isLoggedIn } from '@/utils/auth'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)
const error = ref('')

async function handleLogin() {
  loading.value = true
  error.value = ''
  try {
    const redirectUri = window.location.origin + window.location.pathname
    const res = await getWxH5LoginUrl(redirectUri)
    window.location.href = res.data.login_url
  } catch (e: any) {
    error.value = e.message || '获取授权链接失败'
    loading.value = false
  }
}

async function handleCallback(code: string, state: string) {
  loading.value = true
  try {
    const res = await wxH5Callback(code, state)
    saveAuthInfo(res.data.token, { id: res.data.user_id, role: res.data.role, nickname: res.data.nickname, avatar: res.data.avatar })
    userStore.setToken(res.data.token)
    userStore.setUserInfo({ id: res.data.user_id, role: res.data.role, nickname: res.data.nickname, avatar: res.data.avatar, status: 'active' })
    clearAuthParams()
    showToast('登录成功')
    if (res.data.role !== 'student') {
      error.value = '您不是学生，无法访问此页面'
      userStore.logout()
      loading.value = false
      return
    }
    router.push('/chat')
  } catch (e: any) {
    error.value = e.message || '登录失败'
    loading.value = false
  }
}

onMounted(() => {
  if (isLoggedIn()) {
    router.push('/chat')
    return
  }
  const callback = handleWxCallback()
  if (callback) {
    handleCallback(callback.code, callback.state)
  }
})
</script>

<style scoped lang="scss">
.login-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  .login-header {
    text-align: center;
    margin-bottom: 40px;
    h1 { color: #fff; font-size: 28px; margin-bottom: 10px; }
    p { color: rgba(255, 255, 255, 0.8); font-size: 14px; }
  }
  .login-content {
    width: 100%;
    max-width: 300px;
    text-align: center;
    background: #fff;
    border-radius: 12px;
    padding: 30px 20px;
    .error-text { color: #ee0a24; margin: 15px 0; }
  }
}
</style>
