<template>
  <div class="not-found-container">
    <div class="not-found-content">
      <h1>404</h1>
      <p>抱歉，您访问的页面不存在</p>
      <el-button type="primary" @click="goHome">返回首页</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

function goHome() {
  router.push('/')
}
</script>

<style scoped lang="scss">
.not-found-container {
  width: 100%;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  
  .not-found-content {
    text-align: center;
    
    h1 {
      font-size: 120px;
      font-weight: bold;
      color: #409eff;
      margin: 0;
      line-height: 1;
    }
    
    p {
      font-size: 18px;
      color: #606266;
      margin: 20px 0 30px;
    }
  }
}
</style>
