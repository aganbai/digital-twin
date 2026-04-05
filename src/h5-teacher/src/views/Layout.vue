<template>
  <div class="layout-container">
    <el-aside :width="isCollapse ? '64px' : '220px'" class="sidebar">
      <div class="logo">
        <span v-show="!isCollapse" class="logo-text">教师工作台</span>
      </div>
      <el-menu :default-active="activeMenu" class="sidebar-menu" :collapse="isCollapse" :collapse-transition="false" router>
        <el-menu-item index="/chat-list">
          <el-icon><ChatDotSquare /></el-icon>
          <template #title>聊天列表</template>
        </el-menu-item>
        <el-menu-item index="/classes">
          <el-icon><School /></el-icon>
          <template #title>学生管理</template>
        </el-menu-item>
        <el-menu-item index="/knowledge">
          <el-icon><Collection /></el-icon>
          <template #title>知识库</template>
        </el-menu-item>
        <el-menu-item index="/profile">
          <el-icon><User /></el-icon>
          <template #title>我的</template>
        </el-menu-item>
        <el-menu-item index="/courses">
          <el-icon><Document /></el-icon>
          <template #title>课程管理</template>
        </el-menu-item>
        <el-menu-item index="/personas">
          <el-icon><Avatar /></el-icon>
          <template #title>分身管理</template>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container class="main-container">
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="toggleCollapse">
            <Expand v-if="isCollapse" />
            <Fold v-else />
          </el-icon>
          <span class="page-title">{{ pageTitle }}</span>
        </div>
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-avatar :size="32" :src="userStore.userInfo?.avatar">{{ userStore.nickname?.charAt(0) }}</el-avatar>
              <span class="username">{{ userStore.nickname }}</span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { ChatDotSquare, School, Collection, Document, Avatar, User, Expand, Fold, ArrowDown } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const isCollapse = ref(false)

const activeMenu = computed(() => route.path)
const pageTitle = computed(() => (route.meta.title as string) || '教师工作台')

function toggleCollapse() { isCollapse.value = !isCollapse.value }

function handleCommand(command: string) {
  if (command === 'logout') {
    ElMessageBox.confirm('确定要退出登录吗？', '提示', { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' })
      .then(() => { userStore.logout(); router.push('/login') })
      .catch(() => {})
  }
}
</script>

<style scoped lang="scss">
.layout-container { display: flex; height: 100vh; overflow: hidden; }
.sidebar {
  background: #304156;
  transition: width 0.3s;
  overflow: hidden;
  .logo {
    height: 60px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #263445;
    .logo-text { color: #fff; font-size: 16px; font-weight: bold; }
  }
  .sidebar-menu {
    border: none;
    background: transparent;
    &:not(.el-menu--collapse) { width: 220px; }
    .el-menu-item { color: #bfcbd9; &:hover { background-color: #263445; } &.is-active { color: #409eff; background-color: #263445; } }
  }
}
.main-container { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
.header {
  height: 60px;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
  .header-left { display: flex; align-items: center; .collapse-btn { font-size: 20px; cursor: pointer; margin-right: 15px; color: #606266; &:hover { color: #409eff; } } .page-title { font-size: 16px; font-weight: 500; color: #303133; } }
  .header-right { .user-info { display: flex; align-items: center; cursor: pointer; .username { margin: 0 8px; color: #606266; } } }
}
.main-content { flex: 1; overflow: auto; background: #f5f7fa; padding: 20px; }
</style>
