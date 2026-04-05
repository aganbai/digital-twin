<template>
  <div class="personas-container">
    <el-card>
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span>班级分身管理</span>
          <!-- 迭代11：移除"创建分身"按钮，分身随班级创建 -->
          <el-tooltip content="分身随班级创建，请前往班级管理创建班级" placement="left">
            <el-button type="info" size="small" disabled>创建分身</el-button>
          </el-tooltip>
        </div>
      </template>

      <!-- 空状态提示 -->
      <el-empty v-if="personas.length === 0 && !loading" description="暂无班级分身">
        <template #description>
          <p>暂无班级分身</p>
          <p style="color: #909399; font-size: 12px;">请先创建班级，分身将随班级自动创建</p>
        </template>
      </el-empty>

      <el-table :data="personas" style="width: 100%" v-loading="loading">
        <el-table-column prop="nickname" label="分身昵称" min-width="120" />
        <el-table-column prop="school" label="学校" min-width="120" />
        <el-table-column label="绑定班级" min-width="140">
          <template #default="{ row }">
            <el-tag v-if="row.bound_class_name" type="success" size="small">
              {{ row.bound_class_name }}
            </el-tag>
            <span v-else style="color: #909399;">-</span>
          </template>
        </el-table-column>
        <el-table-column label="公开状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_public ? 'success' : 'info'" size="small">
              {{ row.is_public ? '已公开' : '未公开' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="student_count" label="学生数" width="90">
          <template #default="{ row }">
            {{ row.student_count || 0 }}
          </template>
        </el-table-column>
        <el-table-column prop="document_count" label="文档数" width="90">
          <template #default="{ row }">
            {{ row.document_count || 0 }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="170">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" link @click="handleEnterDashboard(row)">
              进入管理
            </el-button>
            <el-button
              v-if="row.bound_class_id"
              type="success"
              size="small"
              link
              @click="handleViewClass(row)"
            >
              班级详情
            </el-button>
            <el-button
              :type="row.is_public ? 'warning' : 'success'"
              size="small"
              link
              @click="handleToggleVisibility(row)"
            >
              {{ row.is_public ? '设为私有' : '设为公开' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { getPersonaList, setVisibility, getPersonaDashboard } from '@/api/persona'
import type { Persona } from '@/api/persona'

const router = useRouter()
const personas = ref<Persona[]>([])
const loading = ref(false)

/** 格式化日期 */
function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

/** 获取分身列表 */
async function fetchPersonas() {
  loading.value = true
  try {
    const res = await getPersonaList()
    // 只显示教师分身
    const teacherPersonas = (res.data || []).filter((p: Persona) => p.role === 'teacher')

    // 获取每个分身的统计数据
    const enrichedPersonas = await Promise.all(
      teacherPersonas.map(async (p: Persona) => {
        try {
          const dashRes = await getPersonaDashboard(p.id)
          return {
            ...p,
            student_count: dashRes.data.stats?.total_students || 0,
            document_count: dashRes.data.stats?.total_documents || 0
          }
        } catch {
          return { ...p, student_count: 0, document_count: 0 }
        }
      })
    )

    personas.value = enrichedPersonas
  } catch (error) {
    console.error('获取分身列表失败:', error)
    ElMessage.error('获取分身列表失败')
  } finally {
    loading.value = false
  }
}

/** 进入分身仪表盘 */
function handleEnterDashboard(row: Persona) {
  // 存储当前分身信息到 localStorage
  localStorage.setItem('currentPersona', JSON.stringify(row))
  router.push('/dashboard')
}

/** 查看班级详情 */
function handleViewClass(row: Persona) {
  if (row.bound_class_id) {
    router.push(`/classes/${row.bound_class_id}`)
  }
}

/** 切换公开/私有状态 */
async function handleToggleVisibility(row: Persona) {
  try {
    const newIsPublic = !row.is_public
    await setVisibility(row.id, newIsPublic)
    // 更新本地状态
    const index = personas.value.findIndex(p => p.id === row.id)
    if (index !== -1) {
      personas.value[index] = { ...personas.value[index], is_public: newIsPublic }
    }
    ElMessage.success(newIsPublic ? '已设为公开' : '已设为私有')
  } catch (error) {
    console.error('设置公开状态失败:', error)
    ElMessage.error('设置失败')
  }
}

onMounted(() => {
  fetchPersonas()
})
</script>

<style scoped lang="scss">
.personas-container {
  height: 100%;
  padding: 20px;
}
</style>
