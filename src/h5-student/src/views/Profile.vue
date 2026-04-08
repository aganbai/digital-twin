<template>
  <div class="profile-page">
    <van-nav-bar title="我的" />
    <div class="profile-content">
      <van-cell-group inset>
        <van-cell :title="userStore.nickname" :label="'ID: ' + (userStore.userInfo?.id || '-')" center>
          <template #icon>
            <van-image v-if="userStore.userInfo?.avatar" :src="userStore.userInfo.avatar" round width="50" height="50" style="margin-right: 10px;" />
            <van-icon v-else name="user-circle-o" size="50" style="margin-right: 10px;" />
          </template>
        </van-cell>
      </van-cell-group>
      <van-cell-group inset title="功能菜单" style="margin-top: 15px;">
        <van-cell title="个人资料" is-link icon="edit" />
        <van-cell title="意见反馈" is-link icon="chat-o" />
        <van-cell title="关于我们" is-link icon="info-o" />
        <van-cell title="退出登录" is-link icon="revoke" @click="handleLogout" />
      </van-cell-group>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { showConfirmDialog, showToast } from 'vant'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

function handleLogout() {
  showConfirmDialog({ title: '提示', message: '确定要退出登录吗？' })
    .then(() => { userStore.logout(); router.push('/login') })
    .catch(() => {})
}
</script>

<style scoped lang="scss">
.profile-page {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  .profile-content { flex: 1; overflow-y: auto; padding: 10px; }
}
</style>
