<template>
  <div class="role-select-container">
    <div class="role-card">
      <h1>选择您的角色</h1>
      <p>请选择您在系统中的身份</p>
      <div class="role-options">
        <div class="role-option" :class="{ active: selectedRole === 'teacher' }" @click="selectedRole = 'teacher'">
          <el-icon :size="48"><User /></el-icon>
          <span>我是教师</span>
        </div>
        <div class="role-option" :class="{ active: selectedRole === 'student' }" @click="selectedRole = 'student'">
          <el-icon :size="48"><Reading /></el-icon>
          <span>我是学生</span>
        </div>
      </div>
      <div class="nickname-input">
        <el-input v-model="nickname" placeholder="请输入您的昵称" size="large" />
      </div>
      <el-button type="primary" size="large" :disabled="!selectedRole || !nickname" :loading="loading" @click="handleSubmit">
        确认
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Reading } from '@element-plus/icons-vue'
import { completeProfile } from '@/api/auth'

const router = useRouter()
const selectedRole = ref<'teacher' | 'student' | ''>('')
const nickname = ref('')
const loading = ref(false)

async function handleSubmit() {
  if (!selectedRole.value || !nickname.value) return
  loading.value = true
  try {
    await completeProfile(selectedRole.value, nickname.value)
    ElMessage.success('角色设置成功')
    if (selectedRole.value === 'teacher') {
      router.push('/home')
    } else {
      // 学生跳转到学生端
      window.location.href = '/h5-student/'
    }
  } catch (e: any) {
    ElMessage.error(e.message || '设置失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
.role-select-container {
  width: 100vw;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
.role-card {
  width: 400px;
  background: #fff;
  border-radius: 12px;
  padding: 40px;
  text-align: center;
  h1 { font-size: 24px; color: #333; margin-bottom: 10px; }
  p { color: #909399; margin-bottom: 30px; }
}
.role-options {
  display: flex;
  justify-content: center;
  gap: 20px;
  margin-bottom: 30px;
}
.role-option {
  width: 140px;
  height: 140px;
  border: 2px solid #dcdfe6;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s;
  span { margin-top: 10px; font-size: 14px; color: #606266; }
  &:hover { border-color: #409eff; }
  &.active { border-color: #409eff; background: #ecf5ff; }
}
.nickname-input { margin-bottom: 20px; }
</style>
