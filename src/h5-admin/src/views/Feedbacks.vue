<template>
  <div class="feedbacks-container">
    <el-card class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="请选择状态" clearable>
            <el-option label="待处理" value="pending" />
            <el-option label="处理中" value="processing" />
            <el-option label="已解决" value="resolved" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
    <el-card class="table-card">
      <el-table :data="feedbackList" v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="user_nickname" label="用户昵称" width="120" />
        <el-table-column prop="user_role" label="用户角色" width="100">
          <template #default="{ row }">
            <el-tag :type="row.user_role === 'teacher' ? 'primary' : 'success'" size="small">
              {{ row.user_role === 'teacher' ? '教师' : '学生' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="content" label="反馈内容" min-width="200" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusTagType(row.status)" size="small">{{ getStatusName(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="提交时间" width="180">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleViewDetail(row)">查看</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        class="pagination"
        @size-change="loadFeedbacks"
        @current-change="loadFeedbacks"
      />
    </el-card>
    <el-dialog v-model="detailDialogVisible" title="反馈详情" width="600px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="用户昵称">{{ currentFeedback?.user_nickname }}</el-descriptions-item>
        <el-descriptions-item label="用户角色">{{ getRoleName(currentFeedback?.user_role || '') }}</el-descriptions-item>
        <el-descriptions-item label="反馈内容">{{ currentFeedback?.content }}</el-descriptions-item>
        <el-descriptions-item label="图片" v-if="currentFeedback?.images?.length">
          <el-image
            v-for="(img, index) in currentFeedback?.images"
            :key="index"
            :src="img"
            :preview-src-list="currentFeedback?.images"
            style="width: 100px; height: 100px; margin-right: 10px"
            fit="cover"
          />
        </el-descriptions-item>
        <el-descriptions-item label="提交时间">{{ formatDate(currentFeedback?.created_at || '') }}</el-descriptions-item>
        <el-descriptions-item label="状态">{{ getStatusName(currentFeedback?.status || '') }}</el-descriptions-item>
        <el-descriptions-item label="回复" v-if="currentFeedback?.reply">{{ currentFeedback?.reply }}</el-descriptions-item>
      </el-descriptions>
      <el-divider />
      <el-form :model="replyForm" label-width="80px">
        <el-form-item label="更新状态">
          <el-select v-model="replyForm.status" placeholder="请选择状态">
            <el-option label="待处理" value="pending" />
            <el-option label="处理中" value="processing" />
            <el-option label="已解决" value="resolved" />
          </el-select>
        </el-form-item>
        <el-form-item label="回复内容">
          <el-input v-model="replyForm.reply" type="textarea" :rows="3" placeholder="请输入回复内容" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="detailDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveReply" :loading="saveLoading">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'
import { getFeedbackList, updateFeedback } from '@/api/feedback'
import type { Feedback } from '@/api/feedback'

const searchForm = reactive({ status: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const feedbackList = ref<Feedback[]>([])
const loading = ref(false)
const detailDialogVisible = ref(false)
const currentFeedback = ref<Feedback | null>(null)
const replyForm = reactive({ status: '', reply: '' })
const saveLoading = ref(false)

function formatDate(date: string) { return dayjs(date).format('YYYY-MM-DD HH:mm:ss') }
function getStatusTagType(status: string) {
  const types: Record<string, string> = { pending: 'warning', processing: 'primary', resolved: 'success' }
  return types[status] || 'info'
}
function getStatusName(status: string) {
  const names: Record<string, string> = { pending: '待处理', processing: '处理中', resolved: '已解决' }
  return names[status] || status
}
function getRoleName(role: string) {
  const names: Record<string, string> = { teacher: '教师', student: '学生' }
  return names[role] || role
}

async function loadFeedbacks() {
  loading.value = true
  try {
    const res = await getFeedbackList({ page: pagination.page, page_size: pagination.pageSize, ...searchForm })
    feedbackList.value = res.data.items
    pagination.total = res.data.total
  } catch (e) {
    console.error('加载反馈列表失败', e)
  } finally {
    loading.value = false
  }
}

function handleSearch() { pagination.page = 1; loadFeedbacks() }
function handleReset() { searchForm.status = ''; handleSearch() }

function handleViewDetail(feedback: Feedback) {
  currentFeedback.value = feedback
  replyForm.status = feedback.status
  replyForm.reply = feedback.reply || ''
  detailDialogVisible.value = true
}

async function handleSaveReply() {
  if (!currentFeedback.value) return
  saveLoading.value = true
  try {
    await updateFeedback(currentFeedback.value.id, replyForm.status, replyForm.reply)
    ElMessage.success('保存成功')
    detailDialogVisible.value = false
    loadFeedbacks()
  } catch (e) {
    console.error('保存失败', e)
  } finally {
    saveLoading.value = false
  }
}

onMounted(() => loadFeedbacks())
</script>

<style scoped lang="scss">
.feedbacks-container {
  .search-card { margin-bottom: 20px; }
  .table-card {
    .pagination { margin-top: 20px; display: flex; justify-content: flex-end; }
  }
}
</style>
