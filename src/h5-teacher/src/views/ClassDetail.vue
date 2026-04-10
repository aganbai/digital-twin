<template>
  <div class="class-detail-container">
    <!-- 班级信息卡片 -->
    <el-card v-loading="loading">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span>{{ classInfo.name || '班级详情' }}</span>
          <div style="display: flex; gap: 8px;">
            <el-button type="default" size="small" @click="$router.back()">返回</el-button>
            <el-button type="primary" size="small" @click="showAddStudentDialog">添加学生</el-button>
          </div>
        </div>
      </template>

      <!-- 错误状态 -->
      <el-alert
        v-if="error"
        :title="error"
        type="error"
        :closable="true"
        @close="error = ''"
        show-icon
        style="margin-bottom: 16px;"
      />

      <div v-if="!loading && !error && classInfo.id">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="班级名称">{{ classInfo.name }}</el-descriptions-item>
          <el-descriptions-item label="学生数">{{ classInfo.student_count }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDate(classInfo.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="公开状态">
            <el-tag :type="classInfo.is_public ? 'success' : 'info'" size="small">
              {{ classInfo.is_public ? '公开' : '私密' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">{{ classInfo.description || '-' }}</el-descriptions-item>
        </el-descriptions>

        <!-- 教材配置信息 -->
        <template v-if="classInfo.curriculum_config">
          <el-divider content-position="left">📚 教材配置</el-divider>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="学段">
              {{ getGradeLevelLabel(classInfo.curriculum_config.grade_level) }}
            </el-descriptions-item>
            <el-descriptions-item label="年级">
              {{ classInfo.curriculum_config.grade || '-' }}
            </el-descriptions-item>
            <el-descriptions-item label="学科" :span="2">
              <el-tag
                v-for="subject in classInfo.curriculum_config.subjects"
                :key="subject"
                size="small"
                style="margin-right: 8px;"
              >
                {{ subject }}
              </el-tag>
              <span v-if="!classInfo.curriculum_config.subjects?.length">-</span>
            </el-descriptions-item>
            <el-descriptions-item label="教材版本" :span="2">
              <el-tag
                v-for="version in classInfo.curriculum_config.textbook_versions"
                :key="version"
                type="info"
                size="small"
                style="margin-right: 8px;"
              >
                {{ version }}
              </el-tag>
              <span v-if="!classInfo.curriculum_config.textbook_versions?.length">-</span>
            </el-descriptions-item>
            <el-descriptions-item label="自定义教材" :span="2" v-if="classInfo.curriculum_config.custom_textbooks?.length">
              <el-tag
                v-for="book in classInfo.curriculum_config.custom_textbooks"
                :key="book"
                type="warning"
                size="small"
                style="margin-right: 8px;"
              >
                {{ book }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="教学进度" :span="2">
              {{ classInfo.curriculum_config.current_progress || '-' }}
            </el-descriptions-item>
          </el-descriptions>
        </template>
      </div>

      <!-- 空状态 -->
      <el-empty
        v-if="!loading && !error && !classInfo.id"
        description="班级不存在或已被删除"
      />
    </el-card>

    <!-- 学生列表卡片 -->
    <el-card style="margin-top: 20px;" v-loading="studentsLoading">
      <template #header>
        <span>学生列表</span>
      </template>

      <!-- 空状态 -->
      <el-empty
        v-if="!studentsLoading && students.length === 0"
        description="暂无学生"
      >
        <el-button type="primary" @click="showAddStudentDialog">添加学生</el-button>
      </el-empty>

      <el-table v-if="students.length > 0" :data="students" style="width: 100%">
        <el-table-column prop="nickname" label="昵称" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
              {{ row.status === 'active' ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_active_at" label="最后活跃" width="180">
          <template #default="{ row }">
            {{ formatDate(row.last_active_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button type="danger" size="small" @click="handleRemoveStudent(row)">移除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加学生弹窗 -->
    <el-dialog v-model="addStudentDialogVisible" title="添加学生" width="500px" :close-on-click-modal="false">
      <el-form :model="addStudentForm" label-width="80px">
        <el-form-item label="学生ID">
          <el-input v-model="addStudentForm.student_id" placeholder="请输入学生ID或手机号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addStudentDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddStudent" :loading="addStudentLoading">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getClassDetail, addStudentToClass, removeStudentFromClass, getClassStudents } from '@/api/class'
import type { ClassInfo } from '@/api/class'

const route = useRoute()
const classId = ref<number>(Number(route.params.id))

const classInfo = ref<Partial<ClassInfo>>({})
const loading = ref(false)
const error = ref('')

const students = ref<any[]>([])
const studentsLoading = ref(false)
const studentsError = ref('')

const addStudentDialogVisible = ref(false)
const addStudentForm = ref({ student_id: '' })
const addStudentLoading = ref(false)

// 学段选项
const GRADE_LEVEL_OPTIONS = [
  { value: 'preschool', label: '学前班' },
  { value: 'primary_lower', label: '小学低年级' },
  { value: 'primary_upper', label: '小学高年级' },
  { value: 'junior', label: '初中' },
  { value: 'senior', label: '高中' },
  { value: 'university', label: '大学及以上' },
  { value: 'adult_life', label: '成人生活技能' },
  { value: 'adult_professional', label: '成人职业培训' }
]

function getGradeLevelLabel(value?: string): string {
  if (!value) return '-'
  const option = GRADE_LEVEL_OPTIONS.find(opt => opt.value === value)
  return option?.label || value
}

function formatDate(date?: string): string {
  if (!date) return '-'
  const d = new Date(date)
  return d.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

function showAddStudentDialog() {
  addStudentForm.value = { student_id: '' }
  addStudentDialogVisible.value = true
}

async function loadClassDetail() {
  loading.value = true
  error.value = ''
  try {
    const result = await getClassDetail(classId.value)
    if (result.code === 0) {
      classInfo.value = result.data
    } else {
      error.value = result.message || '加载班级详情失败'
    }
  } catch (e: any) {
    error.value = '加载班级详情失败，请重试'
    console.error('加载班级详情失败:', e)
  } finally {
    loading.value = false
  }
}

async function loadStudents() {
  studentsLoading.value = true
  studentsError.value = ''
  try {
    const result = await getClassStudents(classId.value)
    if (result.code === 0) {
      students.value = Array.isArray(result.data) ? result.data : []
    } else {
      studentsError.value = result.message || '加载学生列表失败'
    }
  } catch (e: any) {
    studentsError.value = '加载学生列表失败，请重试'
    console.error('加载学生列表失败:', e)
  } finally {
    studentsLoading.value = false
  }
}

async function handleAddStudent() {
  if (!addStudentForm.value.student_id) {
    ElMessage.warning('请输入学生ID')
    return
  }
  addStudentLoading.value = true
  try {
    const studentId = Number(addStudentForm.value.student_id)
    const result = await addStudentToClass(classId.value, studentId)
    if (result.code === 0) {
      ElMessage.success('添加成功')
      addStudentDialogVisible.value = false
      loadStudents()
      loadClassDetail()
    } else {
      ElMessage.error(result.message || '添加失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '添加失败，请重试')
  } finally {
    addStudentLoading.value = false
  }
}

async function handleRemoveStudent(row: any) {
  try {
    await ElMessageBox.confirm(`确定要将 "${row.nickname}" 移出班级吗？`, '提示', { type: 'warning' })
    const result = await removeStudentFromClass(classId.value, row.id)
    if (result.code === 0) {
      ElMessage.success('移除成功')
      loadStudents()
      loadClassDetail()
    } else {
      ElMessage.error(result.message || '移除失败')
    }
  } catch (e: any) {
    if (e !== 'cancel' && !e?.toString().includes('cancel')) {
      ElMessage.error(e?.response?.data?.message || '移除失败，请重试')
    }
  }
}

onMounted(() => {
  loadClassDetail()
  loadStudents()
})
</script>

<style scoped lang="scss">
.class-detail-container {
  height: 100%;
}
</style>
