<template>
  <div class="dashboard-container">
    <!-- 系统总览卡片 -->
    <el-row :gutter="20" class="overview-row">
      <el-col :xs="12" :sm="8" :md="6" :lg="4">
        <div class="stat-card">
          <div class="stat-title">总用户数</div>
          <div class="stat-value">{{ overview?.total_users || 0 }}</div>
          <div class="stat-trend up">今日新增: {{ overview?.today_new_users || 0 }}</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6" :lg="4">
        <div class="stat-card">
          <div class="stat-title">教师数</div>
          <div class="stat-value">{{ overview?.teacher_count || 0 }}</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6" :lg="4">
        <div class="stat-card">
          <div class="stat-title">学生数</div>
          <div class="stat-value">{{ overview?.student_count || 0 }}</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6" :lg="4">
        <div class="stat-card">
          <div class="stat-title">总对话数</div>
          <div class="stat-value">{{ overview?.total_conversations || 0 }}</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6" :lg="4">
        <div class="stat-card">
          <div class="stat-title">总消息数</div>
          <div class="stat-value">{{ overview?.total_messages || 0 }}</div>
          <div class="stat-trend up">今日: {{ overview?.today_messages || 0 }}</div>
        </div>
      </el-col>
      <el-col :xs="12" :sm="8" :md="6" :lg="4">
        <div class="stat-card">
          <div class="stat-title">今日活跃</div>
          <div class="stat-value">{{ overview?.today_active_users || 0 }}</div>
        </div>
      </el-col>
    </el-row>

    <!-- 趋势图表 -->
    <el-row :gutter="20" class="charts-row">
      <el-col :xs="24" :lg="16">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>用户趋势</span>
              <el-radio-group v-model="trendDays" size="small">
                <el-radio-button :value="7">7天</el-radio-button>
                <el-radio-button :value="30">30天</el-radio-button>
                <el-radio-button :value="90">90天</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <div ref="userTrendChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :lg="8">
        <el-card class="chart-card">
          <template #header>
            <span>角色分布</span>
          </template>
          <div ref="rolePieChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 对话统计 -->
    <el-row :gutter="20" class="charts-row">
      <el-col :xs="24" :lg="12">
        <el-card class="chart-card">
          <template #header>
            <span>对话时段分布</span>
          </template>
          <div ref="chatHourChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :lg="12">
        <el-card class="chart-card">
          <template #header>
            <span>活跃用户排行</span>
          </template>
          <el-table :data="activeUsers" style="width: 100%" max-height="300">
            <el-table-column prop="nickname" label="用户昵称" />
            <el-table-column prop="role" label="角色" width="80">
              <template #default="{ row }">
                <el-tag :type="row.role === 'teacher' ? 'primary' : 'success'" size="small">
                  {{ row.role === 'teacher' ? '教师' : '学生' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="message_count" label="消息数" width="80" />
            <el-table-column prop="last_active_at" label="最后活跃" width="150">
              <template #default="{ row }">
                {{ formatDate(row.last_active_at) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, nextTick } from 'vue'
import * as echarts from 'echarts'
import dayjs from 'dayjs'
import { getOverview, getUserStats, getChatStats, getActiveUsers } from '@/api/dashboard'
import type { OverviewData, ActiveUser } from '@/api/dashboard'

// 数据
const overview = ref<OverviewData | null>(null)
const activeUsers = ref<ActiveUser[]>([])
const trendDays = ref(30)

// 图表引用
const userTrendChartRef = ref<HTMLElement>()
const rolePieChartRef = ref<HTMLElement>()
const chatHourChartRef = ref<HTMLElement>()

// 图表实例
let userTrendChart: echarts.ECharts | null = null
let rolePieChart: echarts.ECharts | null = null
let chatHourChart: echarts.ECharts | null = null

// 格式化日期
function formatDate(date: string) {
  return dayjs(date).format('MM-DD HH:mm')
}

// 加载概览数据
async function loadOverview() {
  try {
    const res = await getOverview()
    overview.value = res.data
  } catch (e) {
    console.error('加载概览数据失败', e)
  }
}

// 加载用户统计
async function loadUserStats() {
  try {
    const res = await getUserStats(trendDays.value)
    const data = res.data
    
    // 更新用户趋势图
    if (userTrendChart && data.register_trend) {
      userTrendChart.setOption({
        xAxis: {
          type: 'category',
          data: data.register_trend.map((item) => item.date),
        },
        series: [
          {
            name: '注册用户',
            type: 'line',
            smooth: true,
            data: data.register_trend.map((item) => item.count),
          },
        ],
      })
    }
    
    // 更新角色分布饼图
    if (rolePieChart && data.role_distribution) {
      rolePieChart.setOption({
        series: [
          {
            type: 'pie',
            radius: '60%',
            data: data.role_distribution.map((item) => ({
              name: item.role === 'teacher' ? '教师' : item.role === 'student' ? '学生' : '管理员',
              value: item.count,
            })),
          },
        ],
      })
    }
  } catch (e) {
    console.error('加载用户统计失败', e)
  }
}

// 加载对话统计
async function loadChatStats() {
  try {
    const res = await getChatStats(trendDays.value)
    const data = res.data
    
    // 更新对话时段分布图
    if (chatHourChart && data.hourly_distribution) {
      chatHourChart.setOption({
        xAxis: {
          type: 'category',
          data: data.hourly_distribution.map((item) => `${item.hour}:00`),
        },
        series: [
          {
            type: 'bar',
            data: data.hourly_distribution.map((item) => item.count),
          },
        ],
      })
    }
  } catch (e) {
    console.error('加载对话统计失败', e)
  }
}

// 加载活跃用户
async function loadActiveUsers() {
  try {
    const res = await getActiveUsers(7, 10)
    activeUsers.value = res.data
  } catch (e) {
    console.error('加载活跃用户失败', e)
  }
}

// 初始化图表
function initCharts() {
  // 用户趋势图
  if (userTrendChartRef.value) {
    userTrendChart = echarts.init(userTrendChartRef.value)
    userTrendChart.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: { type: 'category', data: [] },
      yAxis: { type: 'value' },
      series: [{ name: '注册用户', type: 'line', smooth: true, data: [] }],
    })
  }
  
  // 角色分布饼图
  if (rolePieChartRef.value) {
    rolePieChart = echarts.init(rolePieChartRef.value)
    rolePieChart.setOption({
      tooltip: { trigger: 'item' },
      legend: { orient: 'vertical', left: 'left' },
      series: [{ type: 'pie', radius: '60%', data: [] }],
    })
  }
  
  // 对话时段分布图
  if (chatHourChartRef.value) {
    chatHourChart = echarts.init(chatHourChartRef.value)
    chatHourChart.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: { type: 'category', data: [] },
      yAxis: { type: 'value' },
      series: [{ type: 'bar', data: [] }],
    })
  }
}

// 监听趋势天数变化
watch(trendDays, () => {
  loadUserStats()
  loadChatStats()
})

onMounted(async () => {
  await nextTick()
  initCharts()
  loadOverview()
  loadUserStats()
  loadChatStats()
  loadActiveUsers()
})
</script>

<style scoped lang="scss">
.dashboard-container {
  .overview-row {
    margin-bottom: 20px;
  }
  
  .charts-row {
    margin-bottom: 20px;
  }
  
  .stat-card {
    background: #fff;
    border-radius: 8px;
    padding: 20px;
    
    .stat-title {
      font-size: 14px;
      color: #909399;
      margin-bottom: 10px;
    }
    
    .stat-value {
      font-size: 28px;
      font-weight: bold;
      color: #303133;
    }
    
    .stat-trend {
      font-size: 12px;
      margin-top: 10px;
      
      &.up {
        color: #67c23a;
      }
      
      &.down {
        color: #f56c6c;
      }
    }
  }
  
  .chart-card {
    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
    
    .chart-container {
      height: 300px;
    }
  }
}
</style>
