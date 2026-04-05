<template>
  <div class="classes-container">
    <el-card>
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span>班级管理</span>
          <el-button type="primary" size="small" @click="showCreateDialog">创建班级</el-button>
        </div>
      </template>
      <el-table :data="classes" style="width: 100%">
        <el-table-column prop="name" label="班级名称" />
        <el-table-column prop="persona_nickname" label="分身昵称" width="120" />
        <el-table-column label="公开状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_public ? 'success' : 'info'" size="small">
              {{ row.is_public ? '公开' : '私密' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="student_count" label="学生数" width="100" />
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column label="操作" width="250">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="$router.push(`/class/${row.id}`)">详情</el-button>
            <el-button type="default" size="small" @click="showEditDialog(row)">编辑</el-button>
            <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建班级弹窗 -->
    <el-dialog v-model="createDialogVisible" title="创建班级" width="600px" :close-on-click-modal="false">
      <el-form :model="createForm" :rules="createRules" ref="createFormRef" label-width="100px">
        <!-- 分身信息区域 -->
        <el-divider content-position="left">分身信息</el-divider>
        <el-alert type="info" :closable="false" style="margin-bottom: 20px;">
          创建班级时会同步创建该班级专属的分身
        </el-alert>
        
        <el-form-item label="分身昵称" prop="persona_nickname">
          <el-input v-model="createForm.persona_nickname" placeholder="请输入分身昵称（如：王老师）" maxlength="30" show-word-limit />
        </el-form-item>
        
        <el-form-item label="学校名称" prop="persona_school">
          <el-input v-model="createForm.persona_school" placeholder="请输入学校名称" maxlength="50" show-word-limit />
        </el-form-item>
        
        <el-form-item label="分身描述" prop="persona_description">
          <el-input v-model="createForm.persona_description" type="textarea" :rows="3" placeholder="请输入分身描述（教学风格、擅长领域等）" maxlength="200" show-word-limit />
        </el-form-item>

        <!-- 班级信息区域 -->
        <el-divider content-position="left">班级信息</el-divider>
        
        <el-form-item label="班级名称" prop="name">
          <el-input v-model="createForm.name" placeholder="请输入班级名称" maxlength="50" show-word-limit />
        </el-form-item>
        
        <el-form-item label="班级描述">
          <el-input v-model="createForm.description" type="textarea" :rows="3" placeholder="请输入班级描述（可选）" maxlength="200" show-word-limit />
        </el-form-item>
        
        <el-form-item label="公开班级">
          <div style="display: flex; flex-direction: column; gap: 8px;">
            <el-switch v-model="createForm.is_public" active-text="公开" inactive-text="私密" />
            <span style="font-size: 12px; color: #909399;">
              公开班级对所有学生可见，非公开班级仅限受邀学生加入
            </span>
            <div :style="{
              marginTop: '8px',
              padding: '8px 12px',
              borderRadius: '4px',
              backgroundColor: createForm.is_public ? '#f6ffed' : '#fff7e6',
              borderLeft: createForm.is_public ? '3px solid #52c41a' : '3px solid #faad14'
            }">
              <span :style="{
                fontSize: '12px',
                color: createForm.is_public ? '#52c41a' : '#d48806'
              }">
                {{ createForm.is_public ? '当前班级公开，所有学生可见' : '当前班级私密，仅受邀学生可加入' }}
              </span>
            </div>
          </div>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :loading="createLoading">确定创建</el-button>
      </template>
    </el-dialog>

    <!-- 编辑班级弹窗 -->
    <el-dialog v-model="editDialogVisible" title="编辑班级" width="500px" :close-on-click-modal="false">
      <el-form :model="editForm" :rules="editRules" ref="editFormRef" label-width="100px">
        <el-form-item label="班级名称" prop="name">
          <el-input v-model="editForm.name" placeholder="请输入班级名称" maxlength="50" show-word-limit />
        </el-form-item>
        
        <el-form-item label="班级描述">
          <el-input v-model="editForm.description" type="textarea" :rows="3" placeholder="请输入班级描述（可选）" maxlength="200" show-word-limit />
        </el-form-item>
        
        <el-form-item label="公开班级">
          <div style="display: flex; flex-direction: column; gap: 8px;">
            <el-switch v-model="editForm.is_public" active-text="公开" inactive-text="私密" />
            <span style="font-size: 12px; color: #909399;">
              公开班级对所有学生可见，非公开班级仅限受邀学生加入
            </span>
            <div :style="{
              marginTop: '8px',
              padding: '8px 12px',
              borderRadius: '4px',
              backgroundColor: editForm.is_public ? '#f6ffed' : '#fff7e6',
              borderLeft: editForm.is_public ? '3px solid #52c41a' : '3px solid #faad14'
            }">
              <span :style="{
                fontSize: '12px',
                color: editForm.is_public ? '#52c41a' : '#d48806'
              }">
                {{ editForm.is_public ? '当前班级公开，所有学生可见' : '当前班级私密，仅受邀学生可加入' }}
              </span>
            </div>
          </div>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleEdit" :loading="editLoading">保存</el-button>
      </template>
    </el-dialog>

    <!-- 创建成功弹窗 -->
    <el-dialog v-model="successDialogVisible" title="创建成功" width="500px" :close-on-click-modal="false">
      <div class="success-content">
        <el-result icon="success" title="班级创建成功" sub-title="已同步创建班级专属分身">
          <template #extra>
            <!-- 分身信息卡片 -->
            <el-card shadow="never" class="persona-info-card">
              <template #header>
                <span style="color: #409eff; font-weight: 600;">班级分身信息</span>
              </template>
              <el-descriptions :column="1" border size="small">
                <el-descriptions-item label="分身昵称">{{ createdClassInfo?.persona_nickname }}</el-descriptions-item>
                <el-descriptions-item label="分身ID">{{ createdClassInfo?.persona_id }}</el-descriptions-item>
                <el-descriptions-item label="所属学校">{{ createdClassInfo?.persona_school }}</el-descriptions-item>
              </el-descriptions>
            </el-card>

            <!-- 分享信息 -->
            <el-card shadow="never" class="share-info-card" style="margin-top: 16px;">
              <template #header>
                <span>邀请学生</span>
              </template>
              <el-form label-width="80px">
                <el-form-item label="分享链接">
                  <el-input :model-value="createdClassInfo?.share_url" readonly>
                    <template #append>
                      <el-button @click="copyShareUrl">复制</el-button>
                    </template>
                  </el-input>
                </el-form-item>
                <el-form-item label="分享码">
                  <div style="display: flex; align-items: center; gap: 12px;">
                    <span class="share-code">{{ createdClassInfo?.share_code }}</span>
                    <el-button size="small" @click="copyShareCode">复制</el-button>
                  </div>
                </el-form-item>
              </el-form>
            </el-card>

            <el-alert type="warning" :closable="false" style="margin-top: 16px;">
              💡 将分享链接或分享码发给学生，即可邀请他们加入班级
            </el-alert>
          </template>
        </el-result>
      </div>
      
      <template #footer>
        <el-button type="primary" @click="successDialogVisible = false">完成</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'

// 班级列表
const classes = ref<any[]>([])

// 创建班级弹窗
const createDialogVisible = ref(false)
const createLoading = ref(false)
const createFormRef = ref<FormInstance>()

// 创建成功弹窗
const successDialogVisible = ref(false)
const createdClassInfo = ref<any>(null)

// 编辑班级弹窗
const editDialogVisible = ref(false)
const editLoading = ref(false)
const editFormRef = ref<FormInstance>()
const editingClassId = ref<number>(0)

// 编辑表单
const editForm = reactive({
  name: '',
  description: '',
  is_public: true
})

// 编辑表单验证规则
const editRules: FormRules = {
  name: [
    { required: true, message: '请输入班级名称', trigger: 'blur' },
    { max: 50, message: '班级名称最长50个字符', trigger: 'blur' }
  ]
}

// 创建表单
const createForm = reactive({
  name: '',
  description: '',
  persona_nickname: '',
  persona_school: '',
  persona_description: '',
  is_public: true
})

// 表单验证规则
const createRules: FormRules = {
  name: [
    { required: true, message: '请输入班级名称', trigger: 'blur' },
    { max: 50, message: '班级名称最长50个字符', trigger: 'blur' }
  ],
  persona_nickname: [
    { required: true, message: '请输入分身昵称', trigger: 'blur' },
    { max: 30, message: '分身昵称最长30个字符', trigger: 'blur' }
  ],
  persona_school: [
    { required: true, message: '请输入学校名称', trigger: 'blur' },
    { max: 50, message: '学校名称最长50个字符', trigger: 'blur' }
  ],
  persona_description: [
    { required: true, message: '请输入分身描述', trigger: 'blur' },
    { max: 200, message: '分身描述最长200个字符', trigger: 'blur' }
  ]
}

// 打开创建弹窗
function showCreateDialog() {
  createForm.name = ''
  createForm.description = ''
  createForm.persona_nickname = ''
  createForm.persona_school = ''
  createForm.persona_description = ''
  createForm.is_public = true
  createDialogVisible.value = true
}

// 打开编辑弹窗
function showEditDialog(row: any) {
  editingClassId.value = row.id
  editForm.name = row.name || ''
  editForm.description = row.description || ''
  editForm.is_public = row.is_public !== undefined ? row.is_public : true
  editDialogVisible.value = true
}

// 编辑班级
async function handleEdit() {
  if (!editFormRef.value) return
  
  try {
    await editFormRef.value.validate()
  } catch {
    return
  }
  
  editLoading.value = true
  try {
    const response = await fetch(`/api/classes/${editingClassId.value}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify({
        name: editForm.name,
        description: editForm.description || undefined,
        is_public: editForm.is_public
      })
    })
    
    const result = await response.json()
    
    if (result.code === 0) {
      editDialogVisible.value = false
      ElMessage.success('保存成功')
      loadClasses() // 刷新列表
    } else {
      ElMessage.error(result.message || '保存失败')
    }
  } catch (e) {
    ElMessage.error('保存失败，请重试')
  } finally {
    editLoading.value = false
  }
}

// 创建班级
async function handleCreate() {
  if (!createFormRef.value) return
  
  try {
    await createFormRef.value.validate()
  } catch {
    return
  }
  
  createLoading.value = true
  try {
    const response = await fetch('/api/classes', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify({
        name: createForm.name,
        description: createForm.description || undefined,
        persona_nickname: createForm.persona_nickname,
        persona_school: createForm.persona_school,
        persona_description: createForm.persona_description,
        is_public: createForm.is_public
      })
    })
    
    const result = await response.json()
    
    if (result.code === 0) {
      createDialogVisible.value = false
      createdClassInfo.value = result.data
      successDialogVisible.value = true
      loadClasses() // 刷新列表
    } else {
      ElMessage.error(result.message || '创建失败')
    }
  } catch (e) {
    ElMessage.error('创建失败，请重试')
  } finally {
    createLoading.value = false
  }
}

// 复制分享链接
function copyShareUrl() {
  if (!createdClassInfo.value?.share_url) return
  navigator.clipboard.writeText(createdClassInfo.value.share_url)
  ElMessage.success('链接已复制')
}

// 复制分享码
function copyShareCode() {
  if (!createdClassInfo.value?.share_code) return
  navigator.clipboard.writeText(createdClassInfo.value.share_code)
  ElMessage.success('分享码已复制')
}

// 删除班级
async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`确定要删除班级 "${row.name}" 吗？`, '提示', { type: 'warning' })
    // TODO: 调用删除API
    ElMessage.success('删除成功')
    loadClasses()
  } catch (e) {}
}

// 加载班级列表
async function loadClasses() {
  try {
    const response = await fetch('/api/classes', {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    const result = await response.json()
    if (result.code === 0) {
      classes.value = Array.isArray(result.data) ? result.data : (result.data?.classes || [])
    }
  } catch (e) {
    console.error('加载班级列表失败:', e)
  }
}

onMounted(() => {
  loadClasses()
})
</script>

<style scoped lang="scss">
.classes-container { 
  height: 100%; 
}

.success-content {
  text-align: left;
}

.persona-info-card {
  :deep(.el-card__header) {
    padding: 12px 16px;
    background-color: #f5f7fa;
  }
}

.share-info-card {
  :deep(.el-card__header) {
    padding: 12px 16px;
    background-color: #f5f7fa;
  }
}

.share-code {
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
  letter-spacing: 4px;
}
</style>
