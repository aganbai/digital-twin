<template>
  <div class="profile-container">
    <el-card>
      <template #header><span>个人中心</span></template>
      <el-descriptions :column="1" border>
        <el-descriptions-item label="头像">
          <el-avatar :size="60" :src="userStore.userInfo?.avatar">{{ userStore.nickname?.charAt(0) }}</el-avatar>
        </el-descriptions-item>
        <el-descriptions-item label="昵称">{{ userStore.nickname }}</el-descriptions-item>
        <el-descriptions-item label="角色">教师</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag type="success" size="small">正常</el-tag>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
    <el-card style="margin-top: 20px;">
      <template #header><span>功能菜单</span></template>
      <el-menu>
        <el-menu-item index="edit-profile"><el-icon><Edit /></el-icon>修改资料</el-menu-item>
        <el-menu-item index="feedback"><el-icon><ChatDotSquare /></el-icon>意见反馈</el-menu-item>
        <el-menu-item index="about"><el-icon><InfoFilled /></el-icon>关于我们</el-menu-item>
        <el-menu-item index="logout" @click="handleLogout"><el-icon><SwitchButton /></el-icon>退出登录</el-menu-item>
      </el-menu>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { Edit, ChatDotSquare, InfoFilled, SwitchButton } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

function handleLogout() {
  ElMessageBox.confirm('确定要退出登录吗？', '提示', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' })
    .then(() => { userStore.logout(); router.push('/login') })
    .catch(() => {})
}
</script>

<style scoped lang="scss">
.profile-container { height: 100%; }
</style>
