<template>
  <div class="logs-container">
    <el-card class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="用户ID">
          <el-input v-model="searchForm.user_id" placeholder="请输入用户ID" clearable />
        </el-form-item>
        <el-form-item label="操作类型">
          <el-select v-model="searchForm.action" placeholder="请选择操作类型" clearable>
            <el-option label="用户登录" value="user.login" />
            <el-option label="发送消息" value="chat.send_message" />
            <el-option label="创建班级" value="class.create" />
            <el-option label="上传知识" value="knowledge.upload" />
            <el-option label="创建分身" value="persona.create" />
          </el-select>
        </el-form-item>
        <el-form-item label="平台">
          <el-select v-model="searchForm.platform" placeholder="请选择平台" clearable>
            <el-option label="小程序" value="miniapp" />
            <el-option label="H5" value="h5" />
            <el-option label="API" value="api" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button type="success" @click="handleExport">导出CSV</el-button>
        </el-form-item>
      </el-form>
    </el-card>
    <el-card class="table-card">
      <el-table :data="logList" v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="user_nickname" label="用户昵称" width="120" />
        <el-table-column prop="user_role" label="角色" width="80">
          <template #default="{ row }">
            <el-tag :type="row.user_role === 'teacher' ? 'primary' : 'success'" size="small">
              {{ row.user_role === 'teacher' ? '教师' : '学生' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="action" label="操作类型" width="150" />
        <el-table-column prop="resource" label="资源类型" width="100" />
        <el-table-column prop="detail" label="详情" min-width="200" show-overflow-tooltip />
        <el-table-column prop="platform" label="平台" width="80">
          <template #default="{ row }">
            <el-tag size="small">{{ getPlatformName(row.platform) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status_code" label="状态码" width="80" />
        <el-table-column prop="duration_ms" label="耗时(ms)" width="90" />
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :page-sizes="[20, 50, 100, 200]"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        class="pagination"
        @size-change="loadLogs"
        @current-change="loadLogs"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, watch } from 'vue'
import dayjs from 'dayjs'
import { getLogList, exportLogs } from '@/api/log'
import type { OperationLog } from '@/api/log'

const searchForm = reactive({ user_id: '', action: '', platform: '', start_date: '', end_date: '' })
const dateRange = ref<string[]>([])
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const logList = ref<OperationLog[]>([])
const loading = ref(false)

watch(dateRange, (val) => {
  if (val && val.length === 2) {
    searchForm.start_date = val[0]
    searchForm.end_date = val[1]
  } else {
    searchForm.start_date = ''
    searchForm.end_date = ''
  }
})

function formatDate(date: string) { return dayjs(date).format('YYYY-MM-DD HH:mm:ss') }
function getPlatformName(platform: string) {
  const names: Record<string, string> = { miniapp: '小程序', h5: 'H5', api: 'API' }
  return names[platform] || platform
}

async function loadLogs() {
  loading.value = true
  try {
    const params: any = { page: pagination.page, page_size: pagination.pageSize }
    if (searchForm.user_id) params.user_id = parseInt(searchForm.user_id)
    if (searchForm.action) params.action = searchForm.action
    if (searchForm.platform) params.platform = searchForm.platform
    if (searchForm.start_date) params.start_date = searchForm.start_date
    if (searchForm.end_date) params.end_date = searchForm.end_date
    
    const res = await getLogList(params)
    logList.value = res.data.items
    pagination.total = res.data.total
  } catch (e) {
    console.error('加载日志列表失败', e)
  } finally {
    loading.value = false
  }
}

function handleSearch() { pagination.page = 1; loadLogs() }
function handleReset() {
  searchForm.user_id = ''
  searchForm.action = ''
  searchForm.platform = ''
  dateRange.value = []
  handleSearch()
}

function handleExport() {
  const params: any = {}
  if (searchForm.user_id) params.user_id = parseInt(searchForm.user_id)
  if (searchForm.action) params.action = searchForm.action
  if (searchForm.platform) params.platform = searchForm.platform
  if (searchForm.start_date) params.start_date = searchForm.start_date
  if (searchForm.end_date) params.end_date = searchForm.end_date
  
  const url = exportLogs(params)
  window.open(url, '_blank')
}

onMounted(() => loadLogs())
</script>

<style scoped lang="scss">
.logs-container {
  .search-card { margin-bottom: 20px; }
  .table-card {
    .pagination { margin-top: 20px; display: flex; justify-content: flex-end; }
  }
}
</style>
