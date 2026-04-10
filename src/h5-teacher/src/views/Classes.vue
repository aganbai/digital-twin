<template>
  <div class="classes-container">
    <el-card v-loading="loading">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span>班级管理</span>
          <el-button type="primary" size="small" @click="showCreateDialog">创建班级</el-button>
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

      <!-- 空状态 -->
      <el-empty
        v-if="!loading && !error && classes.length === 0"
        description="暂无班级"
      >
        <el-button type="primary" @click="showCreateDialog">创建班级</el-button>
      </el-empty>

      <el-table v-if="classes.length > 0" :data="classes" style="width: 100%">
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

        <!-- 教材配置区域（折叠面板） -->
        <el-divider content-position="left">📚 教材配置（可选）</el-divider>

        <div class="curriculum-config-section">
          <div
            class="curriculum-header"
            @click="createCurriculumExpanded = !createCurriculumExpanded"
          >
            <span class="curriculum-title">
              {{ createCurriculumExpanded ? '📖 教材配置已展开' : '📖 点击展开配置教材信息' }}
            </span>
            <el-icon class="curriculum-arrow" :class="{ 'is-expanded': createCurriculumExpanded }">
              <ArrowDown />
            </el-icon>
          </div>

          <el-collapse-transition>
            <div v-show="createCurriculumExpanded" class="curriculum-content">
              <!-- 学段选择 -->
              <el-form-item label="学段">
                <el-select
                  v-model="createCurriculumConfig.grade_level"
                  placeholder="请选择学段"
                  clearable
                  style="width: 100%"
                  @change="onCreateGradeLevelChange"
                >
                  <el-option
                    v-for="opt in GRADE_LEVEL_OPTIONS"
                    :key="opt.value"
                    :label="opt.label"
                    :value="opt.value"
                  />
                </el-select>
              </el-form-item>

              <!-- 年级选择（K12学段显示） -->
              <el-form-item label="年级" v-if="showCreateGradeSelector">
                <el-select
                  v-model="createCurriculumConfig.grade"
                  placeholder="请选择年级"
                  clearable
                  style="width: 100%"
                  :disabled="!createCurriculumConfig.grade_level"
                >
                  <el-option
                    v-for="grade in createGradeOptions"
                    :key="grade"
                    :label="grade"
                    :value="grade"
                  />
                </el-select>
              </el-form-item>

              <!-- 学科选择 -->
              <el-form-item label="学科" v-if="createCurriculumConfig.grade_level">
                <el-select
                  v-model="createCurriculumConfig.subjects"
                  multiple
                  placeholder="请选择学科"
                  style="width: 100%"
                >
                  <el-option
                    v-for="subject in createSubjectOptions"
                    :key="subject"
                    :label="subject"
                    :value="subject"
                  />
                </el-select>
              </el-form-item>

              <!-- 教材版本（非成人学段显示） -->
              <el-form-item
                label="教材版本"
                v-if="createCurriculumConfig.grade_level && !showCreateCustomTextbooks && createCurriculumConfig.grade_level !== 'adult_life' && createCurriculumConfig.grade_level !== 'adult_professional'"
              >
                <el-select
                  v-model="createCurriculumConfig.textbook_versions"
                  multiple
                  placeholder="请选择教材版本"
                  style="width: 100%"
                >
                  <el-option
                    v-for="version in TEXTBOOK_VERSIONS"
                    :key="version"
                    :label="version"
                    :value="version"
                  />
                </el-select>
              </el-form-item>

              <!-- 自定义教材（大学及以上显示） -->
              <el-form-item label="自定义教材" v-if="showCreateCustomTextbooks">
                <div class="custom-textbook-input">
                  <el-input
                    v-model="customTextbookInput"
                    placeholder="请输入教材名称，按回车或点击添加"
                    maxlength="50"
                    @keyup.enter="addCreateCustomTextbook"
                  />
                  <el-button type="primary" @click="addCreateCustomTextbook">添加</el-button>
                </div>
                <div class="custom-textbook-tags" v-if="createCurriculumConfig.custom_textbooks?.length">
                  <el-tag
                    v-for="(book, index) in createCurriculumConfig.custom_textbooks"
                    :key="index"
                    closable
                    @close="removeCreateCustomTextbook(index)"
                    class="custom-textbook-tag"
                  >
                    {{ book }}
                  </el-tag>
                </div>
              </el-form-item>

              <!-- 教学进度 -->
              <el-form-item label="教学进度" v-if="createCurriculumConfig.grade_level">
                <el-input
                  v-model="createCurriculumConfig.current_progress"
                  placeholder="如：第三单元 乘法初步"
                  maxlength="100"
                  clearable
                />
              </el-form-item>
            </div>
          </el-collapse-transition>
        </div>
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

        <!-- 教材配置区域（折叠面板） -->
        <el-divider content-position="left">📚 教材配置（可选）</el-divider>

        <div class="curriculum-config-section">
          <div
            class="curriculum-header"
            @click="editCurriculumExpanded = !editCurriculumExpanded"
          >
            <span class="curriculum-title">
              {{ editCurriculumExpanded ? '📖 教材配置已展开' : (editCurriculumConfig.grade_level ? '📖 已配置教材信息，点击修改' : '📖 点击展开配置教材信息') }}
            </span>
            <el-icon class="curriculum-arrow" :class="{ 'is-expanded': editCurriculumExpanded }">
              <ArrowDown />
            </el-icon>
          </div>

          <el-collapse-transition>
            <div v-show="editCurriculumExpanded" class="curriculum-content">
              <!-- 学段选择 -->
              <el-form-item label="学段">
                <el-select
                  v-model="editCurriculumConfig.grade_level"
                  placeholder="请选择学段"
                  clearable
                  style="width: 100%"
                  @change="onEditGradeLevelChange"
                >
                  <el-option
                    v-for="opt in GRADE_LEVEL_OPTIONS"
                    :key="opt.value"
                    :label="opt.label"
                    :value="opt.value"
                  />
                </el-select>
              </el-form-item>

              <!-- 年级选择（K12学段显示） -->
              <el-form-item label="年级" v-if="showEditGradeSelector">
                <el-select
                  v-model="editCurriculumConfig.grade"
                  placeholder="请选择年级"
                  clearable
                  style="width: 100%"
                  :disabled="!editCurriculumConfig.grade_level"
                >
                  <el-option
                    v-for="grade in editGradeOptions"
                    :key="grade"
                    :label="grade"
                    :value="grade"
                  />
                </el-select>
              </el-form-item>

              <!-- 学科选择 -->
              <el-form-item label="学科" v-if="editCurriculumConfig.grade_level">
                <el-select
                  v-model="editCurriculumConfig.subjects"
                  multiple
                  placeholder="请选择学科"
                  style="width: 100%"
                >
                  <el-option
                    v-for="subject in editSubjectOptions"
                    :key="subject"
                    :label="subject"
                    :value="subject"
                  />
                </el-select>
              </el-form-item>

              <!-- 教材版本（非成人学段显示） -->
              <el-form-item
                label="教材版本"
                v-if="editCurriculumConfig.grade_level && !showEditCustomTextbooks && editCurriculumConfig.grade_level !== 'adult_life' && editCurriculumConfig.grade_level !== 'adult_professional'"
              >
                <el-select
                  v-model="editCurriculumConfig.textbook_versions"
                  multiple
                  placeholder="请选择教材版本"
                  style="width: 100%"
                >
                  <el-option
                    v-for="version in TEXTBOOK_VERSIONS"
                    :key="version"
                    :label="version"
                    :value="version"
                  />
                </el-select>
              </el-form-item>

              <!-- 自定义教材（大学及以上显示） -->
              <el-form-item label="自定义教材" v-if="showEditCustomTextbooks">
                <div class="custom-textbook-input">
                  <el-input
                    v-model="editCustomTextbookInput"
                    placeholder="请输入教材名称，按回车或点击添加"
                    maxlength="50"
                    @keyup.enter="addEditCustomTextbook"
                  />
                  <el-button type="primary" @click="addEditCustomTextbook">添加</el-button>
                </div>
                <div class="custom-textbook-tags" v-if="editCurriculumConfig.custom_textbooks?.length">
                  <el-tag
                    v-for="(book, index) in editCurriculumConfig.custom_textbooks"
                    :key="index"
                    closable
                    @close="removeEditCustomTextbook(index)"
                    class="custom-textbook-tag"
                  >
                    {{ book }}
                  </el-tag>
                </div>
              </el-form-item>

              <!-- 教学进度 -->
              <el-form-item label="教学进度" v-if="editCurriculumConfig.grade_level">
                <el-input
                  v-model="editCurriculumConfig.current_progress"
                  placeholder="如：第三单元 乘法初步"
                  maxlength="100"
                  clearable
                />
              </el-form-item>
            </div>
          </el-collapse-transition>
        </div>
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
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { ArrowDown } from '@element-plus/icons-vue'
import type { CurriculumConfig } from '@/api/class'
import { getClassList, createClass, updateClass, deleteClass } from '@/api/class'

// ===== 常量定义 =====
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

// 年级选项映射
const GRADE_OPTIONS_MAP: Record<string, string[]> = {
  preschool: ['幼儿园大班', '学前'],
  primary_lower: ['一年级', '二年级', '三年级'],
  primary_upper: ['四年级', '五年级', '六年级'],
  junior: ['七年级', '八年级', '九年级'],
  senior: ['高一', '高二', '高三'],
  university: ['大一', '大二', '大三', '大四', '研究生', '博士']
}

// K12学科选项
const K12_SUBJECTS = ['语文', '数学', '英语', '物理', '化学', '生物', '历史', '地理', '政治', '音乐', '美术', '体育', '信息技术']

// 成人生活技能课程
const ADULT_LIFE_CATEGORIES = ['中餐', '西餐', '烘焙', '力量训练', '有氧运动', '瑜伽', '手工', '园艺', '摄影', '绘画']

// 成人职业培训课程
const ADULT_PROFESSIONAL_CATEGORIES = ['编程', '设计', '会计', '法律', '医学', '教育', '管理', '营销', '外语', '考证培训']

// 教材版本选项
const TEXTBOOK_VERSIONS = ['人教版', '北师大版', '苏教版', '沪教版', '部编版', '外研版', '浙教版', '冀教版']

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

// 编辑表单教材配置
const editCurriculumConfig = reactive<CurriculumConfig>({
  grade_level: undefined,
  grade: undefined,
  subjects: [],
  textbook_versions: [],
  custom_textbooks: [],
  current_progress: undefined
})

// 是否展开编辑教材配置
const editCurriculumExpanded = ref(false)

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

// 创建表单教材配置
const createCurriculumConfig = reactive<CurriculumConfig>({
  grade_level: undefined,
  grade: undefined,
  subjects: [],
  textbook_versions: [],
  custom_textbooks: [],
  current_progress: undefined
})

// 是否展开创建教材配置
const createCurriculumExpanded = ref(false)

// 自定义教材输入
const customTextbookInput = ref('')
const editCustomTextbookInput = ref('')

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

// 计算属性：当前年级选项
const createGradeOptions = computed(() => {
  if (!createCurriculumConfig.grade_level) return []
  return GRADE_OPTIONS_MAP[createCurriculumConfig.grade_level] || []
})

const editGradeOptions = computed(() => {
  if (!editCurriculumConfig.grade_level) return []
  return GRADE_OPTIONS_MAP[editCurriculumConfig.grade_level] || []
})

// 计算属性：当前学科选项
const createSubjectOptions = computed(() => {
  if (!createCurriculumConfig.grade_level) return []
  if (createCurriculumConfig.grade_level === 'adult_life') {
    return ADULT_LIFE_CATEGORIES
  }
  if (createCurriculumConfig.grade_level === 'adult_professional') {
    return ADULT_PROFESSIONAL_CATEGORIES
  }
  return K12_SUBJECTS
})

const editSubjectOptions = computed(() => {
  if (!editCurriculumConfig.grade_level) return []
  if (editCurriculumConfig.grade_level === 'adult_life') {
    return ADULT_LIFE_CATEGORIES
  }
  if (editCurriculumConfig.grade_level === 'adult_professional') {
    return ADULT_PROFESSIONAL_CATEGORIES
  }
  return K12_SUBJECTS
})

// 计算属性：是否显示年级选择器
const showCreateGradeSelector = computed(() => {
  return createCurriculumConfig.grade_level !== 'adult_life' &&
         createCurriculumConfig.grade_level !== 'adult_professional'
})

const showEditGradeSelector = computed(() => {
  return editCurriculumConfig.grade_level !== 'adult_life' &&
         editCurriculumConfig.grade_level !== 'adult_professional'
})

// 计算属性：是否显示自定义教材输入（大学及以上）
const showCreateCustomTextbooks = computed(() => {
  return createCurriculumConfig.grade_level === 'university'
})

const showEditCustomTextbooks = computed(() => {
  return editCurriculumConfig.grade_level === 'university'
})

// 打开创建弹窗
function showCreateDialog() {
  createForm.name = ''
  createForm.description = ''
  createForm.persona_nickname = ''
  createForm.persona_school = ''
  createForm.persona_description = ''
  createForm.is_public = true
  // 重置教材配置
  createCurriculumConfig.grade_level = undefined
  createCurriculumConfig.grade = undefined
  createCurriculumConfig.subjects = []
  createCurriculumConfig.textbook_versions = []
  createCurriculumConfig.custom_textbooks = []
  createCurriculumConfig.current_progress = undefined
  createCurriculumExpanded.value = false
  customTextbookInput.value = ''
  createDialogVisible.value = true
}

// 打开编辑弹窗
function showEditDialog(row: any) {
  editingClassId.value = row.id
  editForm.name = row.name || ''
  editForm.description = row.description || ''
  editForm.is_public = row.is_public !== undefined ? row.is_public : true
  // 加载教材配置
  if (row.curriculum_config) {
    editCurriculumConfig.grade_level = row.curriculum_config.grade_level
    editCurriculumConfig.grade = row.curriculum_config.grade
    editCurriculumConfig.subjects = row.curriculum_config.subjects || []
    editCurriculumConfig.textbook_versions = row.curriculum_config.textbook_versions || []
    editCurriculumConfig.custom_textbooks = row.curriculum_config.custom_textbooks || []
    editCurriculumConfig.current_progress = row.curriculum_config.current_progress
    editCurriculumExpanded.value = true
  } else {
    editCurriculumConfig.grade_level = undefined
    editCurriculumConfig.grade = undefined
    editCurriculumConfig.subjects = []
    editCurriculumConfig.textbook_versions = []
    editCurriculumConfig.custom_textbooks = []
    editCurriculumConfig.current_progress = undefined
    editCurriculumExpanded.value = false
  }
  editCustomTextbookInput.value = ''
  editDialogVisible.value = true
}

// 处理创建时学段变化
function onCreateGradeLevelChange() {
  createCurriculumConfig.grade = undefined
  createCurriculumConfig.subjects = []
}

// 处理编辑时学段变化
function onEditGradeLevelChange() {
  editCurriculumConfig.grade = undefined
  editCurriculumConfig.subjects = []
}

// 添加自定义教材（创建）
function addCreateCustomTextbook() {
  const value = customTextbookInput.value.trim()
  if (!value) {
    ElMessage.warning('请输入教材名称')
    return
  }
  if (!createCurriculumConfig.custom_textbooks) {
    createCurriculumConfig.custom_textbooks = []
  }
  if (createCurriculumConfig.custom_textbooks.includes(value)) {
    ElMessage.warning('该教材已添加')
    return
  }
  createCurriculumConfig.custom_textbooks.push(value)
  customTextbookInput.value = ''
}

// 添加自定义教材（编辑）
function addEditCustomTextbook() {
  const value = editCustomTextbookInput.value.trim()
  if (!value) {
    ElMessage.warning('请输入教材名称')
    return
  }
  if (!editCurriculumConfig.custom_textbooks) {
    editCurriculumConfig.custom_textbooks = []
  }
  if (editCurriculumConfig.custom_textbooks.includes(value)) {
    ElMessage.warning('该教材已添加')
    return
  }
  editCurriculumConfig.custom_textbooks.push(value)
  editCustomTextbookInput.value = ''
}

// 移除自定义教材（创建）
function removeCreateCustomTextbook(index: number) {
  createCurriculumConfig.custom_textbooks?.splice(index, 1)
}

// 移除自定义教材（编辑）
function removeEditCustomTextbook(index: number) {
  editCurriculumConfig.custom_textbooks?.splice(index, 1)
}

// 获取学段标签
function getGradeLevelLabel(value: string): string {
  const option = GRADE_LEVEL_OPTIONS.find(opt => opt.value === value)
  return option?.label || value
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
    const body: any = {
      name: editForm.name,
      description: editForm.description || undefined,
      is_public: editForm.is_public
    }

    // 如果展开并填写了教材配置，则添加到请求体
    if (editCurriculumExpanded.value && editCurriculumConfig.grade_level) {
      body.curriculum_config = {
        grade_level: editCurriculumConfig.grade_level,
        grade: editCurriculumConfig.grade,
        subjects: editCurriculumConfig.subjects,
        textbook_versions: editCurriculumConfig.textbook_versions,
        custom_textbooks: editCurriculumConfig.custom_textbooks,
        current_progress: editCurriculumConfig.current_progress
      }
    }

    const result = await updateClass(editingClassId.value, body)

    if (result.code === 0) {
      editDialogVisible.value = false
      ElMessage.success('保存成功')
      loadClasses() // 刷新列表
    } else {
      ElMessage.error(result.message || '保存失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '保存失败，请重试')
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
    const body: any = {
      name: createForm.name,
      description: createForm.description || undefined,
      persona_nickname: createForm.persona_nickname,
      persona_school: createForm.persona_school,
      persona_description: createForm.persona_description,
      is_public: createForm.is_public
    }

    // 如果展开并填写了教材配置，则添加到请求体
    if (createCurriculumExpanded.value && createCurriculumConfig.grade_level) {
      body.curriculum_config = {
        grade_level: createCurriculumConfig.grade_level,
        grade: createCurriculumConfig.grade,
        subjects: createCurriculumConfig.subjects,
        textbook_versions: createCurriculumConfig.textbook_versions,
        custom_textbooks: createCurriculumConfig.custom_textbooks,
        current_progress: createCurriculumConfig.current_progress
      }
    }

    const result = await createClass(body)

    if (result.code === 0) {
      createDialogVisible.value = false
      createdClassInfo.value = result.data
      successDialogVisible.value = true
      loadClasses() // 刷新列表
    } else {
      ElMessage.error(result.message || '创建失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || '创建失败，请重试')
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
    const result = await deleteClass(row.id)
    if (result.code === 0) {
      ElMessage.success('删除成功')
      loadClasses()
    } else {
      ElMessage.error(result.message || '删除失败')
    }
  } catch (e: any) {
    // 用户取消操作不报错
    if (e !== 'cancel' && !e?.toString().includes('cancel')) {
      ElMessage.error(e?.response?.data?.message || '删除失败，请重试')
    }
  }
}

// 加载班级列表
const loading = ref(false)
const error = ref('')

async function loadClasses() {
  loading.value = true
  error.value = ''
  try {
    const result = await getClassList()
    if (result.code === 0) {
      classes.value = Array.isArray(result.data) ? result.data : (result.data?.classes || [])
    } else {
      error.value = result.message || '加载失败'
    }
  } catch (e) {
    error.value = '加载班级列表失败，请重试'
    console.error('加载班级列表失败:', e)
  } finally {
    loading.value = false
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

// 教材配置样式
.curriculum-config-section {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background-color: #fafafa;
  margin-bottom: 16px;
}

.curriculum-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  cursor: pointer;
  user-select: none;
  transition: background-color 0.2s;

  &:hover {
    background-color: #f0f2f5;
  }
}

.curriculum-title {
  font-size: 14px;
  color: #409eff;
  font-weight: 500;
}

.curriculum-arrow {
  font-size: 16px;
  color: #909399;
  transition: transform 0.3s;

  &.is-expanded {
    transform: rotate(180deg);
  }
}

.curriculum-content {
  padding: 16px;
  border-top: 1px solid #e4e7ed;
  background-color: #fff;
}

.custom-textbook-input {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
}

.custom-textbook-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.custom-textbook-tag {
  margin-right: 0;
}
</style>
