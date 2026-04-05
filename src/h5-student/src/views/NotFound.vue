<template>
  <div class="not-found-page">
    <van-empty image="error" description="页面不存在">
      <van-button type="primary" @click="$router.push('/')">返回首页</van-button>
    </van-empty>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped lang="scss">
.not-found-page {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
