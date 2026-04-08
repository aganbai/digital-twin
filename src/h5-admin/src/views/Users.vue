<template>
  <div class="users-container">
    <el-card class="search-card">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="昵称">
          <el-input v-model="searchForm.nickname" placeholder="请输入昵称" clearable />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="searchForm.role" placeholder="请选择角色" clearable>
            <el-option label="教师" value="teacher" />
            <el-option label="学生" value="student" />
            <el-option label="管理员" value="admin" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="请选择状态" clearable>
            <el-option label="正常" value="active" />
            <el-option label="禁用" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
    <el-card class="table-card">
      <el-table :data="userList" v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="nickname" label="昵称" />
        <el-table-column prop="role" label="角色" width="100">
          <template #default="{ row }">
            <el-tag :type="getRoleTagType(row.role)" size="small">{{ getRoleName(row.role) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
              {{ row.status === 'active' ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="注册时间" width="180">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleEditRole(row)">修改角色</el-button>
            <el-button :type="row.status === 'active' ? 'danger' : 'success'" size="small" @click="handleToggleStatus(row)">
              {{ row.status === 'active' ? '禁用' : '启用' }}
            </el-button>
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
        @size-change="loadUsers"
        @current-change="loadUsers"
      />
    </el-card>
    <el-dialog v-model="roleDialogVisible" title="修改用户角色" width="400px">
      <el-form :model="roleForm" label-width="80px">
        <el-form-item label="用户昵称">
          <el-input :value="currentUser?.nickname" disabled />
        </el-form-item>
        <el-form-item label="新角色">
          <el-select v-model="roleForm.role" placeholder="请选择角色">
            <el-option label="教师" value="teacher" />
            <el-option label="学生" value="student" />
            <el-option label="管理员" value="admin" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveRole" :loading="saveLoading">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import dayjs from 'dayjs'
import { getUserList, updateUserRole, updateUserStatus } from '@/api/user'
import type { User } from '@/api/user'

const searchForm = reactive({ nickname: '', role: '', status: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const userList = ref<User[]>([])
const loading = ref(false)
const roleDialogVisible = ref(false)
const currentUser = ref<User | null>(null)
const roleForm = reactive({ role: '' })
const saveLoading = ref(false)

function formatDate(date: string) { return dayjs(date).format('YYYY-MM-DD HH:mm:ss') }
function getRoleTagType(role: string) {
  const types: Record<string, string> = { admin: 'danger', teacher: 'primary', student: 'success' }
  return types[role] || 'info'
}
function getRoleName(role: string) {
  const names: Record<string, string> = { admin: '管理员', teacher: '教师', student: '学生' }
  return names[role] || role
}

async function loadUsers() {
  loading.value = true
  try {
    const res = await getUserList({ page: pagination.page, page_size: pagination.pageSize, ...searchForm })
    userList.value = res.data.items
    pagination.total = res.data.total
  } catch (e) {
    console.error('加载用户列表失败', e)
  } finally {
    loading.value = false
  }
}

function handleSearch() { pagination.page = 1; loadUsers() }
function handleReset() { searchForm.nickname = ''; searchForm.role = ''; searchForm.status = ''; handleSearch() }
function handleEditRole(user: User) { currentUser.value = user; roleForm.role = user.role; roleDialogVisible.value = true }

async function handleSaveRole() {
  if (!currentUser.value) return
  saveLoading.value = true
  try {
    await updateUserRole(currentUser.value.id, roleForm.role)
    ElMessage.success('修改成功')
    roleDialogVisible.value = false
    loadUsers()
  } catch (e) {
    console.error('修改角色失败', e)
  } finally {
    saveLoading.value = false
  }
}

async function handleToggleStatus(user: User) {
  const newStatus = user.status === 'active' ? 'disabled' : 'active'
  const actionText = newStatus === 'disabled' ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(`确定要${actionText}用户 "${user.nickname}" 吗？`, '提示', {
      confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
    })
    await updateUserStatus(user.id, newStatus)
    ElMessage.success(`${actionText}成功`)
    loadUsers()
  } catch (e) {}
}

onMounted(() => loadUsers())
</script>

<style scoped lang="scss">
.users-container {
  .search-card { margin-bottom: 20px; }
  .table-card {
    .pagination { margin-top: 20px; display: flex; justify-content: flex-end; }
  }
}
</style>
