<template>
  <div class="home-container">
    <el-row :gutter="20">
      <el-col :span="6"><div class="stat-card"><div class="stat-title">我的学生</div><div class="stat-value">{{ stats.studentCount }}</div></div></el-col>
      <el-col :span="6"><div class="stat-card"><div class="stat-title">班级数量</div><div class="stat-value">{{ stats.classCount }}</div></div></el-col>
      <el-col :span="6"><div class="stat-card"><div class="stat-title">今日消息</div><div class="stat-value">{{ stats.todayMessages }}</div></div></el-col>
      <el-col :span="6"><div class="stat-card"><div class="stat-title">知识条目</div><div class="stat-value">{{ stats.knowledgeCount }}</div></div></el-col>
    </el-row>
    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="12">
        <el-card>
          <template #header><span>最近对话</span></template>
          <el-table :data="recentChats" style="width: 100%">
            <el-table-column prop="student_name" label="学生" />
            <el-table-column prop="last_message" label="最新消息" show-overflow-tooltip />
            <el-table-column prop="updated_at" label="时间" width="150" />
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header><span>快捷操作</span></template>
          <div class="quick-actions">
            <el-button type="primary" @click="$router.push('/classes')">管理班级</el-button>
            <el-button type="success" @click="$router.push('/knowledge')">添加知识</el-button>
            <el-button type="warning" @click="$router.push('/courses')">发布课程</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const stats = ref({ studentCount: 0, classCount: 0, todayMessages: 0, knowledgeCount: 0 })
const recentChats = ref([])

onMounted(() => {
  // TODO: 加载数据
})
</script>

<style scoped lang="scss">
.home-container {
  .stat-card {
    background: #fff;
    border-radius: 8px;
    padding: 20px;
    text-align: center;
    .stat-title { font-size: 14px; color: #909399; margin-bottom: 10px; }
    .stat-value { font-size: 28px; font-weight: bold; color: #303133; }
  }
  .quick-actions { display: flex; gap: 10px; flex-wrap: wrap; }
}
</style>
