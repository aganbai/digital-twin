<template>
  <div class="knowledge-container">
    <el-card>
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span>知识库</span>
          <el-button type="primary" size="small" @click="showUploadDialog">上传知识</el-button>
        </div>
      </template>
      <el-tabs v-model="activeTab">
        <el-tab-pane label="文档知识" name="document">
          <el-table :data="documents" style="width: 100%">
            <el-table-column prop="title" label="标题" />
            <el-table-column prop="type" label="类型" width="100" />
            <el-table-column prop="created_at" label="上传时间" width="180" />
            <el-table-column label="操作" width="150">
              <template #default="{ row }">
                <el-button type="primary" size="small" link>预览</el-button>
                <el-button type="danger" size="small" link @click="handleDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="问答知识" name="qa">
          <el-table :data="qaList" style="width: 100%">
            <el-table-column prop="question" label="问题" />
            <el-table-column prop="answer" label="答案" show-overflow-tooltip />
            <el-table-column prop="created_at" label="创建时间" width="180" />
            <el-table-column label="操作" width="150">
              <template #default="{ row }">
                <el-button type="primary" size="small" link @click="showEditQADialog(row)">编辑</el-button>
                <el-button type="danger" size="small" link @click="handleDeleteQA(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>
    <el-dialog v-model="uploadDialogVisible" title="上传知识" width="500px">
      <el-upload drag action="#" :auto-upload="false">
        <el-icon class="el-icon--upload"><upload-filled /></el-icon>
        <div class="el-upload__text">拖拽文件到此处或 <em>点击上传</em></div>
        <template #tip><div class="el-upload__tip">支持 PDF、Word、TXT 等格式文件</div></template>
      </el-upload>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">取消</el-button>
        <el-button type="primary">上传</el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="qaDialogVisible" title="编辑问答" width="500px">
      <el-form :model="qaForm" label-width="80px">
        <el-form-item label="问题"><el-input v-model="qaForm.question" type="textarea" :rows="2" /></el-form-item>
        <el-form-item label="答案"><el-input v-model="qaForm.answer" type="textarea" :rows="4" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="qaDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveQA">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'

const activeTab = ref('document')
const documents = ref<any[]>([])
const qaList = ref<any[]>([])
const uploadDialogVisible = ref(false)
const qaDialogVisible = ref(false)
const qaForm = ref({ id: 0, question: '', answer: '' })

function showUploadDialog() { uploadDialogVisible.value = true }
function showEditQADialog(row: any) { qaForm.value = { ...row }; qaDialogVisible.value = true }

async function handleSaveQA() {
  if (!qaForm.value.question || !qaForm.value.answer) { ElMessage.warning('请填写完整'); return }
  // TODO: 保存问答
  ElMessage.success('保存成功')
  qaDialogVisible.value = false
}

async function handleDelete(row: any) {
  try { await ElMessageBox.confirm('确定要删除此文档吗？', '提示', { type: 'warning' }); ElMessage.success('删除成功') } catch (e) {}
}

async function handleDeleteQA(row: any) {
  try { await ElMessageBox.confirm('确定要删除此问答吗？', '提示', { type: 'warning' }); ElMessage.success('删除成功') } catch (e) {}
}

onMounted(() => {
  // TODO: 加载数据
})
</script>

<style scoped lang="scss">
.knowledge-container { height: 100%; }
</style>
