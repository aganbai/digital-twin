<template>
  <div class="discover-page">
    <van-nav-bar title="发现" />
    <div class="discover-content">
      <van-cell-group inset title="功能入口">
        <van-cell title="我的班级" is-link icon="cluster" to="/my-class" />
        <van-cell title="课程列表" is-link icon="orders-o" to="/courses" />
        <van-cell title="学习资料" is-link icon="description" to="/materials" />
      </van-cell-group>
      <van-cell-group inset title="推荐内容" style="margin-top: 15px;">
        <van-card v-for="item in recommendList" :key="item.id" :title="item.title" :desc="item.desc" :thumb="item.thumb">
          <template #footer>
            <van-button size="small">查看详情</van-button>
          </template>
        </van-card>
      </van-cell-group>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const recommendList = ref<any[]>([])

onMounted(() => {
  // TODO: 加载推荐内容
  recommendList.value = [
    { id: 1, title: '高等数学精讲', desc: '大学数学必修课程', thumb: 'https://picsum.photos/100/100?random=1' },
    { id: 2, title: '英语口语提升', desc: '每日一句英语练习', thumb: 'https://picsum.photos/100/100?random=2' },
  ]
})
</script>

<style scoped lang="scss">
.discover-page {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  .discover-content { flex: 1; overflow-y: auto; padding: 10px; }
}
</style>
