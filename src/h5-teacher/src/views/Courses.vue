<template>
  <div class="courses-container">
    <el-card>
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span>课程管理</span>
          <el-button type="primary" size="small" @click="showCreateDialog">发布课程</el-button>
        </div>
      </template>
      <el-table :data="courses" style="width: 100%">
        <el-table-column prop="title" label="课程标题" />
        <el-table-column prop="class_name" label="所属班级" width="150" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'published' ? 'success' : 'info'" size="small">
              {{ row.status === 'published' ? '已发布' : '草稿' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button type="primary" size="small" link @click="handleEdit(row)">编辑</el-button>
            <el-button type="success" size="small" link @click="handlePublish(row)" v-if="row.status !== 'published'">发布</el-button>
            <el-button type="danger" size="small" link @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="createDialogVisible" title="发布课程" width="600px">
      <el-form :model="createForm" label-width="80px">
        <el-form-item label="课程标题"><el-input v-model="createForm.title" placeholder="请输入课程标题" /></el-form-item>
        <el-form-item label="所属班级">
          <el-select v-model="createForm.class_id" placeholder="请选择班级" style="width: 100%;">
            <el-option v-for="c in classes" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="课程内容"><el-input v-model="createForm.content" type="textarea" :rows="6" placeholder="请输入课程内容" /></el-form-item>
        <el-form-item label="附件">
          <el-upload action="#" :auto-upload="false">
            <el-button size="small" type="primary">点击上传</el-button>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const courses = ref<any[]>([])
const classes = ref<any[]>([])
const createDialogVisible = ref(false)
const createForm = ref({ title: '', class_id: '', content: '' })

function showCreateDialog() { createForm.value = { title: '', class_id: '', content: '' }; createDialogVisible.value = true }

async function handleCreate() {
  if (!createForm.value.title || !createForm.value.class_id) { ElMessage.warning('请填写完整'); return }
  // TODO: 创建课程
  ElMessage.success('创建成功')
  createDialogVisible.value = false
}

function handleEdit(row: any) { /* TODO */ }
function handlePublish(row: any) { /* TODO */ ElMessage.success('发布成功') }

async function handleDelete(row: any) {
  try { await ElMessageBox.confirm(`确定要删除课程 "${row.title}" 吗？`, '提示', { type: 'warning' }); ElMessage.success('删除成功') } catch (e) {}
}

onMounted(() => {
  // TODO: 加载数据
})
</script>

<style scoped lang="scss">
.courses-container { height: 100%; }
</style>
