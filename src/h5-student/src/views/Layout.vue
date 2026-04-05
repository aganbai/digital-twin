<template>
  <div class="layout-page">
    <router-view />
    <van-tabbar v-model="activeTab" route>
      <van-tabbar-item to="/chat" icon="chat-o">对话</van-tabbar-item>
      <van-tabbar-item to="/history" icon="clock-o">历史</van-tabbar-item>
      <van-tabbar-item to="/discover" icon="apps-o">发现</van-tabbar-item>
      <van-tabbar-item to="/profile" icon="user-o">我的</van-tabbar-item>
    </van-tabbar>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const activeTab = ref(0)

watch(() => route.path, (path) => {
  const tabs: Record<string, number> = { '/chat': 0, '/history': 1, '/discover': 2, '/profile': 3 }
  activeTab.value = tabs[path] ?? 0
}, { immediate: true })
</script>

<style scoped lang="scss">
.layout-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
}
</style>
