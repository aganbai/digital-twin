<template>
  <div class="chat-list-container">
    <el-card>
      <template #header>
        <div class="header-wrapper">
          <span>学生消息</span>
          <span class="subtitle">共 {{ classes.length }} 个班级</span>
        </div>
      </template>

      <!-- 搜索框 -->
      <el-input
        v-model="searchText"
        placeholder="搜索学生姓名"
        prefix-icon="Search"
        style="margin-bottom: 20px;"
        clearable
      />

      <!-- 加载中 -->
      <div v-if="loading" class="loading-wrapper">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>加载中...</span>
      </div>

      <!-- 班级列表 -->
      <div v-else-if="filteredClasses.length > 0" class="class-list">
        <div
          v-for="cls in filteredClasses"
          :key="cls.class_id"
          class="class-item"
        >
          <!-- 班级头部 -->
          <div class="class-header" @click="toggleClass(cls.class_id)">
            <div class="class-info">
              <el-icon v-if="cls.is_pinned" class="pin-icon"><Star /></el-icon>
              <span class="class-name">{{ cls.class_name }}</span>
              <el-tag v-if="cls.subject" size="small" type="info">{{ cls.subject }}</el-tag>
            </div>
            <div class="class-meta">
              <span class="student-count">{{ cls.students?.length || 0 }} 名学生</span>
              <el-icon class="expand-icon" :class="{ expanded: expandedClasses[cls.class_id] }">
                <ArrowDown />
              </el-icon>
            </div>
          </div>

          <!-- 学生列表（可展开/收起） -->
          <el-collapse-transition>
            <div v-show="expandedClasses[cls.class_id]" class="student-list">
              <div
                v-for="student in cls.students"
                :key="student.student_persona_id"
                class="student-item"
                @click="handleStudentClick(student)"
              >
                <div class="student-avatar">
                  {{ student.student_nickname.charAt(0).toUpperCase() }}
                </div>
                <div class="student-info">
                  <div class="student-top">
                    <div class="student-name-row">
                      <el-icon v-if="student.is_pinned" class="pin-icon-small"><Star /></el-icon>
                      <span class="student-name">{{ student.student_nickname }}</span>
                    </div>
                    <span v-if="student.last_message_time" class="student-time">
                      {{ formatTime(student.last_message_time) }}
                    </span>
                  </div>
                  <div class="student-bottom">
                    <span class="student-message">
                      {{ student.last_message || '暂无消息' }}
                    </span>
                    <el-badge
                      v-if="student.unread_count > 0"
                      :value="student.unread_count > 99 ? '99+' : student.unread_count"
                      class="unread-badge"
                    />
                  </div>
                </div>
              </div>

              <!-- 空状态 -->
              <div v-if="!cls.students || cls.students.length === 0" class="empty-students">
                <span>该班级暂无学生</span>
              </div>
            </div>
          </el-collapse-transition>
        </div>
      </div>

      <!-- 空状态 -->
      <el-empty v-else description="暂无班级聊天记录">
        <template #description>
          <p>创建班级并添加学生后即可查看</p>
        </template>
      </el-empty>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Loading, Star, ArrowDown } from '@element-plus/icons-vue'
import { getTeacherChatList, type TeacherChatStudent, type TeacherChatClassItem, type TeacherChatListResponse } from '@/api/chat'

// ========== 状态 ==========
const router = useRouter()
const searchText = ref('')
const loading = ref(false)
const classes = ref<TeacherChatClassItem[]>([])
const expandedClasses = ref<Record<number, boolean>>({})

// ========== 计算属性 ==========
const filteredClasses = computed(() => {
  if (!searchText.value) return classes.value
  
  // 搜索时展开所有匹配的班级
  const result = classes.value.map(cls => {
    const filteredStudents = cls.students?.filter(s =>
      s.student_nickname.includes(searchText.value)
    ) || []
    
    return {
      ...cls,
      students: filteredStudents
    }
  }).filter(cls => cls.students.length > 0)
  
  // 自动展开匹配的班级
  result.forEach(cls => {
    if (!expandedClasses.value[cls.class_id]) {
      expandedClasses.value[cls.class_id] = true
    }
  })
  
  return result
})

// ========== 方法 ==========
/** 加载教师端聊天列表 */
async function fetchTeacherList() {
  loading.value = true
  try {
    const res = await getTeacherChatList()
    
    // 置顶的班级排在前面
    const sorted = (res.data.classes || []).sort((a, b) => {
      if (a.is_pinned && !b.is_pinned) return -1
      if (!a.is_pinned && b.is_pinned) return 1
      return 0
    })
    
    // 每个班级内，置顶学生排前面
    sorted.forEach(cls => {
      if (cls.students) {
        cls.students.sort((a, b) => {
          if (a.is_pinned && !b.is_pinned) return -1
          if (!a.is_pinned && b.is_pinned) return 1
          return 0
        })
      }
    })
    
    classes.value = sorted
    
    // 默认展开第一个班级
    if (sorted.length > 0) {
      expandedClasses.value[sorted[0].class_id] = true
    }
  } catch (error) {
    console.error('获取教师聊天列表失败:', error)
    ElMessage.error('获取聊天列表失败')
  } finally {
    loading.value = false
  }
}

/** 切换班级展开/收起 */
function toggleClass(classId: number) {
  expandedClasses.value[classId] = !expandedClasses.value[classId]
}

/** 点击学生进入聊天详情 */
function handleStudentClick(student: TeacherChatStudent) {
  router.push({
    path: `/chat/${student.student_persona_id}`,
    query: {
      student_name: student.student_nickname
    }
  })
}

/** 格式化时间 */
function formatTime(timeStr: string): string {
  if (!timeStr) return ''
  
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  
  // 1小时内显示"刚刚"
  if (diff < 3600000) {
    const minutes = Math.floor(diff / 60000)
    return minutes <= 1 ? '刚刚' : `${minutes}分钟前`
  }
  
  // 今天显示时分
  if (date.toDateString() === now.toDateString()) {
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  }
  
  // 昨天
  const yesterday = new Date(now)
  yesterday.setDate(yesterday.getDate() - 1)
  if (date.toDateString() === yesterday.toDateString()) {
    return '昨天'
  }
  
  // 一周内显示星期几
  if (diff < 7 * 24 * 3600000) {
    const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
    return weekdays[date.getDay()]
  }
  
  // 其他显示日期
  return `${date.getMonth() + 1}/${date.getDate()}`
}

// ========== 生命周期 ==========
onMounted(() => {
  fetchTeacherList()
})
</script>

<style scoped lang="scss">
.chat-list-container {
  height: 100%;
}

.header-wrapper {
  display: flex;
  align-items: center;
  justify-content: space-between;
  
  .subtitle {
    font-size: 14px;
    font-weight: normal;
    color: #909399;
  }
}

.loading-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 0;
  color: #909399;
  
  .el-icon {
    margin-right: 8px;
    font-size: 18px;
  }
}

.class-list {
  .class-item {
    margin-bottom: 16px;
    border: 1px solid #e4e7ed;
    border-radius: 8px;
    overflow: hidden;
    
    &:last-child {
      margin-bottom: 0;
    }
  }
  
  .class-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    background: #f5f7fa;
    cursor: pointer;
    user-select: none;
    transition: background-color 0.3s;
    
    &:hover {
      background: #eef1f6;
    }
    
    .class-info {
      display: flex;
      align-items: center;
      gap: 8px;
      
      .pin-icon {
        color: #e6a23c;
        font-size: 16px;
      }
      
      .class-name {
        font-size: 15px;
        font-weight: 500;
        color: #303133;
      }
    }
    
    .class-meta {
      display: flex;
      align-items: center;
      gap: 12px;
      
      .student-count {
        font-size: 13px;
        color: #909399;
      }
      
      .expand-icon {
        transition: transform 0.3s;
        
        &.expanded {
          transform: rotate(180deg);
        }
      }
    }
  }
  
  .student-list {
    background: #fff;
    
    .student-item {
      display: flex;
      align-items: center;
      padding: 12px 16px;
      border-top: 1px solid #ebeef5;
      cursor: pointer;
      transition: background-color 0.3s;
      
      &:hover {
        background: #f5f7fa;
      }
      
      .student-avatar {
        width: 40px;
        height: 40px;
        border-radius: 50%;
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        color: #fff;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 16px;
        font-weight: 500;
        flex-shrink: 0;
      }
      
      .student-info {
        flex: 1;
        margin-left: 12px;
        min-width: 0;
        
        .student-top {
          display: flex;
          align-items: center;
          justify-content: space-between;
          margin-bottom: 4px;
          
          .student-name-row {
            display: flex;
            align-items: center;
            gap: 4px;
            
            .pin-icon-small {
              color: #e6a23c;
              font-size: 14px;
            }
            
            .student-name {
              font-size: 14px;
              font-weight: 500;
              color: #303133;
            }
          }
          
          .student-time {
            font-size: 12px;
            color: #909399;
            flex-shrink: 0;
          }
        }
        
        .student-bottom {
          display: flex;
          align-items: center;
          justify-content: space-between;
          
          .student-message {
            font-size: 13px;
            color: #606266;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
            flex: 1;
            margin-right: 8px;
          }
          
          .unread-badge {
            flex-shrink: 0;
          }
        }
      }
    }
    
    .empty-students {
      padding: 24px;
      text-align: center;
      color: #909399;
      font-size: 14px;
    }
  }
}
</style>
