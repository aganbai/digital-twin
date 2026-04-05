<template>
  <div class="class-detail-container">
    <el-card>
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span>{{ classInfo.name }}</span>
          <el-button type="primary" size="small" @click="showAddStudentDialog">添加学生</el-button>
        </div>
      </template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="班级名称">{{ classInfo.name }}</el-descriptions-item>
        <el-descriptions-item label="学生数">{{ classInfo.student_count }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ classInfo.created_at }}</el-descriptions-item>
        <el-descriptions-item label="描述">{{ classInfo.description || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-card>
    <el-card style="margin-top: 20px;">
      <template #header><span>学生列表</span></template>
      <el-table :data="students" style="width: 100%">
        <el-table-column prop="nickname" label="昵称" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">{{ row.status === 'active' ? '正常' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_active_at" label="最后活跃" width="180" />
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button type="danger" size="small" @click="handleRemoveStudent(row)">移除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="addStudentDialogVisible" title="添加学生" width="500px">
      <el-form :model="addStudentForm" label-width="80px">
        <el-form-item label="学生ID"><el-input v-model="addStudentForm.student_id" placeholder="请输入学生ID或手机号" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addStudentDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddStudent">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

const route = useRoute()
const classInfo = ref<any>({ name: '', student_count: 0, created_at: '', description: '' })
const students = ref<any[]>([])
const addStudentDialogVisible = ref(false)
const addStudentForm = ref({ student_id: '' })

function showAddStudentDialog() { addStudentForm.value = { student_id: '' }; addStudentDialogVisible.value = true }

async function handleAddStudent() {
  if (!addStudentForm.value.student_id) { ElMessage.warning('请输入学生ID'); return }
  // TODO: 添加学生
  ElMessage.success('添加成功')
  addStudentDialogVisible.value = false
}

async function handleRemoveStudent(row: any) {
  try {
    await ElMessageBox.confirm(`确定要将 "${row.nickname}" 移出班级吗？`, '提示', { type: 'warning' })
    // TODO: 移除学生
    ElMessage.success('移除成功')
  } catch (e) {}
}

onMounted(() => {
  // TODO: 加载班级详情
})
</script>

<style scoped lang="scss">
.class-detail-container { height: 100%; }
</style>
