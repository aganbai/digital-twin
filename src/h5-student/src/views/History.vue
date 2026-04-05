<template>
  <div class="history-page">
    <van-nav-bar title="历史记录" />
    <van-search v-model="searchText" placeholder="搜索对话" />
    <div class="history-list">
      <van-cell-group inset>
        <van-cell v-for="item in filteredHistory" :key="item.id" :title="item.title" :label="item.time" is-link @click="$router.push(`/chat?id=${item.id}`)">
          <template #icon><van-icon name="chat-o" style="margin-right: 8px;" /></template>
        </van-cell>
      </van-cell-group>
      <van-empty v-if="!filteredHistory.length" description="暂无历史记录" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

const searchText = ref('')
const history = ref<any[]>([])

const filteredHistory = computed(() => {
  if (!searchText.value) return history.value
  return history.value.filter(h => h.title.includes(searchText.value))
})

onMounted(() => {
  // TODO: 加载历史数据
  history.value = [
    { id: 1, title: '关于数学问题的讨论', time: '2024-01-15 10:30' },
    { id: 2, title: '英语学习咨询', time: '2024-01-14 15:20' },
  ]
})
</script>

<style scoped lang="scss">
.history-page {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  .history-list { flex: 1; overflow-y: auto; padding: 10px; }
}
</style>
