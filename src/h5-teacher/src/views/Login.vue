<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h1>数字分身教师端</h1>
        <p>微信授权登录</p>
      </div>
      <div class="login-content">
        <div v-if="loading" class="loading-wrapper">
          <el-icon class="loading-icon" :size="40"><Loading /></el-icon>
          <p>正在登录中...</p>
        </div>
        <div v-else-if="error" class="error-wrapper">
          <el-icon :size="40" color="#f56c6c"><WarningFilled /></el-icon>
          <p>{{ error }}</p>
          <el-button type="primary" @click="handleLogin">重新登录</el-button>
        </div>
        <div v-else class="login-button-wrapper">
          <el-button type="primary" size="large" @click="handleLogin" :icon="User">微信授权登录</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Loading, WarningFilled } from '@element-plus/icons-vue'
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
    ElMessage.success('登录成功')
    
    if (res.data.is_new_user) {
      router.push('/role-select')
    } else if (res.data.role !== 'teacher') {
      error.value = '您不是教师角色，无法访问此页面'
      userStore.logout()
      loading.value = false
    } else {
      router.push('/home')
    }
  } catch (e: any) {
    error.value = e.message || '登录失败'
    loading.value = false
  }
}

onMounted(() => {
  if (isLoggedIn()) {
    router.push('/home')
    return
  }
  const callback = handleWxCallback()
  if (callback) {
    handleCallback(callback.code, callback.state)
  }
})
</script>

<style scoped lang="scss">
.login-container {
  width: 100vw;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
.login-card {
  width: 400px;
  background: #fff;
  border-radius: 12px;
  padding: 40px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
}
.login-header {
  text-align: center;
  margin-bottom: 40px;
  h1 { font-size: 24px; color: #333; margin-bottom: 10px; }
  p { color: #909399; font-size: 14px; }
}
.login-content { text-align: center; }
.loading-wrapper {
  display: flex; flex-direction: column; align-items: center;
  .loading-icon { animation: spin 1s linear infinite; color: #409eff; }
  p { margin-top: 20px; color: #909399; }
}
.error-wrapper {
  display: flex; flex-direction: column; align-items: center;
  p { margin: 20px 0; color: #f56c6c; }
}
.login-button-wrapper { padding: 20px 0; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
</style>
