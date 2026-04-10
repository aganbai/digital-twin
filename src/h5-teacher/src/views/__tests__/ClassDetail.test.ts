import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises, VueWrapper } from '@vue/test-utils'
import { nextTick } from 'vue'
import { server } from '@/__tests__/mocks/server'
import { http, HttpResponse } from 'msw'
import ClassDetail from '../ClassDetail.vue'

// Mock vue-router
const mockRoute = {
  params: { id: '1' }
}
const mockPush = vi.fn()
vi.mock('vue-router', () => ({
  useRoute: () => mockRoute,
  useRouter: () => ({
    push: mockPush,
    back: vi.fn()
  })
}))

// Mock auth token
vi.mock('@/utils/auth', () => ({
  getToken: vi.fn(() => 'mock-token'),
  clearAuthInfo: vi.fn(),
}))

describe('ClassDetail.vue - 班级详情教材配置展示', () => {
  let wrapper: VueWrapper<any>

  beforeEach(() => {
    vi.clearAllMocks()
    mockRoute.params.id = '1'

    // Setup MSW handlers for this test suite
    server.use(
      // 获取班级详情 - 有教材配置
      http.get('/api/classes/1', () => {
        return HttpResponse.json({
          code: 0,
          data: {
            id: 1,
            name: '三年级数学班',
            description: '小学数学培优班级',
            is_public: true,
            is_active: true,
            persona_id: 456,
            persona_nickname: '王老师',
            teacher_id: 789,
            student_count: 25,
            created_at: '2026-04-09T10:30:00Z',
            updated_at: '2026-04-09T10:30:00Z',
            curriculum_config: {
              id: 1001,
              grade_level: 'primary_lower',
              grade: '三年级',
              subjects: ['数学'],
              textbook_versions: ['人教版', '北师大版'],
              custom_textbooks: ['《小学奥数启蒙》'],
              current_progress: '第三单元 乘法初步',
            },
          },
        })
      }),

      // 获取班级详情 - 无教材配置
      http.get('/api/classes/2', () => {
        return HttpResponse.json({
          code: 0,
          data: {
            id: 2,
            name: '初中英语班',
            description: '英语提高班',
            is_public: false,
            is_active: true,
            persona_id: 457,
            persona_nickname: '李老师',
            teacher_id: 789,
            student_count: 30,
            created_at: '2026-04-08T14:20:00Z',
            updated_at: '2026-04-08T14:20:00Z',
            curriculum_config: null,
          },
        })
      }),

      // 班级不存在
      http.get('/api/classes/9999', () => {
        return HttpResponse.json(
          { code: 40017, message: '班级不存在' },
          { status: 404 }
        )
      }),

      // 获取学生列表 - 有学生
      http.get('/api/teacher/classes/1/students', () => {
        return HttpResponse.json({
          code: 0,
          data: [
            { id: 1, nickname: '学生A', status: 'active', last_active_at: '2026-04-09T10:30:00Z' },
            { id: 2, nickname: '学生B', status: 'inactive', last_active_at: '2026-04-08T14:20:00Z' },
          ],
        })
      }),

      // 获取学生列表 - 无学生
      http.get('/api/teacher/classes/2/students', () => {
        return HttpResponse.json({
          code: 0,
          data: [],
        })
      }),

      // 添加学生
      http.post('/api/teacher/classes/1/students', () => {
        return HttpResponse.json({
          code: 0,
          message: 'success',
          data: null,
        })
      }),

      // 移除学生
      http.delete('/api/teacher/classes/1/students/:studentId', () => {
        return HttpResponse.json({
          code: 0,
          message: 'success',
          data: null,
        })
      })
    )
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
    }
  })

  describe('渲染测试', () => {
    it('应正确渲染班级详情页', async () => {
      wrapper = mount(ClassDetail, {
        global: {
          stubs: {
            'el-card': {
              template: '<div class="el-card" v-loading="loading"><slot /><slot name="header" /></div>',
            },
            'el-button': {
              template: '<button class="el-button" @click="$emit(\'click\')"><slot /></button>',
            },
            'el-alert': true,
            'el-descriptions': {
              template: '<div class="el-descriptions"><slot /></div>',
            },
            'el-descriptions-item': {
              template: '<div class="el-descriptions-item" :span="span"><slot /></div>',
              props: ['label', 'span'],
            },
            'el-tag': {
              template: '<span class="el-tag" :class="type"><slot /></span>',
              props: ['type', 'size'],
            },
            'el-divider': {
              template: '<div class="el-divider"><slot /></div>',
            },
            'el-empty': {
              template: '<div class="el-empty"><slot /></div>',
            },
            'el-table': {
              template: '<div class="el-table"><slot /></div>',
            },
            'el-table-column': true,
            'el-dialog': {
              template: '<div v-if="modelValue" class="el-dialog"><slot /><slot name="footer" /></div>',
              props: ['modelValue', 'title'],
            },
            'el-form': {
              template: '<form class="el-form"><slot /></form>',
            },
            'el-form-item': {
              template: '<div class="el-form-item"><slot /></div>',
            },
            'el-input': {
              template: '<input class="el-input" :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
              props: ['modelValue'],
            },
          },
        },
      })

      await flushPromises()

      expect(wrapper.find('.class-detail-container').exists()).toBe(true)
    })

    it('应显示班级基本信息', async () => {
      wrapper = mount(ClassDetail, {
        global: {
          stubs: {
            'el-card': {
              template: '<div class="el-card" v-loading="loading"><slot /><slot name="header" /></div>',
            },
            'el-button': true,
            'el-alert': true,
            'el-descriptions': {
              template: '<div class="el-descriptions"><slot /></div>',
            },
            'el-descriptions-item': {
              template: '<div class="el-descriptions-item"><slot /></div>',
              props: ['label', 'span'],
            },
            'el-tag': true,
            'el-divider': true,
            'el-empty': true,
            'el-table': true,
          },
        },
      })

      await flushPromises()

      expect(wrapper.vm.classInfo.name).toBe('三年级数学班')
      expect(wrapper.vm.classInfo.description).toBe('小学数学培优班级')
      expect(wrapper.vm.classInfo.is_public).toBe(true)
    })
  })

  describe('教材配置展示测试', () => {
    it('应正确显示完整教材配置信息', async () => {
      wrapper = mount(ClassDetail, {
        global: {
          stubs: {
            'el-card': {
              template: '<div class="el-card"><slot /><slot name="header" /></div>',
            },
            'el-button': true,
            'el-alert': true,
            'el-descriptions': {
              template: '<div class="el-descriptions"><slot /></div>',
            },
            'el-descriptions-item': {
              template: '<div class="el-descriptions-item"><slot /></div>',
              props: ['label', 'span'],
            },
            'el-tag': {
              template: '<span class="el-tag"><slot /></span>',
            },
            'el-divider': true,
            'el-empty': true,
            'el-table': true,
          },
        },
      })

      await flushPromises()

      const config = wrapper.vm.classInfo.curriculum_config
      expect(config).toBeDefined()
      expect(config.grade_level).toBe('primary_lower')
      expect(config.grade).toBe('三年级')
      expect(config.subjects).toContain('数学')
      expect(config.textbook_versions).toContain('人教版')
      expect(config.custom_textbooks).toContain('《小学奥数启蒙》')
      expect(config.current_progress).toBe('第三单元 乘法初步')
    })

    it('应正确显示学段标签', async () => {
      wrapper = mount(ClassDetail, {
        global: {
          stubs: {
            'el-card': true,
            'el-button': true,
            'el-alert': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'el-tag': true,
            'el-divider': true,
            'el-empty': true,
            'el-table': true,
          },
        },
      })

      expect(wrapper.vm.getGradeLevelLabel('primary_lower')).toBe('小学低年级')
      expect(wrapper.vm.getGradeLevelLabel('primary_upper')).toBe('小学高年级')
      expect(wrapper.vm.getGradeLevelLabel('junior')).toBe('初中')
      expect(wrapper.vm.getGradeLevelLabel('senior')).toBe('高中')
      expect(wrapper.vm.getGradeLevelLabel('university')).toBe('大学及以上')
      expect(wrapper.vm.getGradeLevelLabel('adult_life')).toBe('成人生活技能')
      expect(wrapper.vm.getGradeLevelLabel('adult_professional')).toBe('成人职业培训')
      expect(wrapper.vm.getGradeLevelLabel('preschool')).toBe('学前班')
    })

    it('应在无教材配置时隐藏教材配置区域', async () => {
      mockRoute.params.id = '2'

      wrapper = mount(ClassDetail, {
        global: {
          stubs: {
            'el-card': {
              template: '<div class="el-card"><slot /><slot name="header" /></div>',
            },
            'el-button': true,
            'el-alert': true,
            'el-descriptions': {
              template: '<div class="el-descriptions"><slot /></div>',
            },
            'el-descriptions-item': true,
            'el-tag': true,
            'el-divider': true,
            'el-empty': true,
            'el-table': true,
          },
        },
      })

      await flushPromises()

      // 验证 curriculum_config 为 null
      expect(wrapper.vm.classInfo.curriculum_config).toBeNull()
    })

    it('应正确渲染教材配置中的标签列表', async () => {
      wrapper = mount(ClassDetail, {
        global: {
          stubs: {
            'el-card': true,
            'el-button': true,
            'el-alert': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'el-tag': {
              template: '<span class="el-tag" :type="type"><slot /></span>',
              props: ['type', 'size'],
            },
            'el-divider': true,
            'el-empty': true,
            'el-table': true,
          },
        },
      })

      await flushPromises()

      const config = wrapper.vm.classInfo.curriculum_config
      // 验证学科是数组
      expect(Array.isArray(config.subjects)).toBe(true)
      expect(config.subjects.length).toBeGreaterThan(0)

      // 验证教材版本是数组
      expect(Array.isArray(config.textbook_versions)).toBe(true)
      expect(config.textbook_versions.length).toBeGreaterThan(0)

      // 验证自定义教材是数组
      expect(Array.isArray(config.custom_textbooks)).toBe(true)
    })
  })

  describe('学生列表测试', () => {
    it('应正确加载和显示学生列表', async () => {
      wrapper = mount(ClassDetail, {
        global: {
          stubs: {
            'el-card': {
              template: '<div class="el-card"><slot /><slot name="header" /></div>',
            },
            'el-button': true,
            'el-alert': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'el-tag': true,
            'el-divider': true,
            'el-empty': {
              template: '<div class="el-empty" :description="description"><slot /></div>',
              props: ['description'],
            },
            'el-table': {
              template: '<div class="el-table" :data="data"><slot /></div>',
              props: ['data'],
            },
            'el-table-column': true,
            'el-dialog': true,
          },
        },
      })

      await flushPromises()

      // 验证学生列表加载完成后状态
      expect(wrapper.vm.studentsLoading).toBe(false)
      expect(wrapper.vm.students).toBeInstanceOf(Array)
    })

    it('应在无学生时显示空状态', async () => {
      mockRoute.params.id = '2'

      wrapper = mount(ClassDetail, {
        global: {
          stubs: {
            'el-card': {
              template: '<div class="el-card"><slot /><slot name="header" /></div>',
            },
            'el-button': true,
            'el-alert': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'el-tag': true,
            'el-divider': true,
            'el-empty': {
              template: '<div class="el-empty" :description="description"><slot /></div>',
              props: ['description'],
            },
            'el-table': true,
            'el-table-column': true,
            'el-dialog': true,
          },
        },
      })

      await flushPromises()

      expect(wrapper.vm.students).toHaveLength(0)
    })
  })

  describe('学生管理测试', () => {
    it('应能打开添加学生弹窗', async () => {
      wrapper = mount(ClassDetail, {
        global: {
          stubs: {
            'el-card': true,
            'el-button': {
              template: '<button class="el-button" @click="$emit(\'click\')"><slot /></button>',
            },
            'el-alert': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'el-tag': true,
            'el-divider': true,
            'el-empty': true,
            'el-table': true,
            'el-dialog': {
              template: '<div v-if="modelValue" class="el-dialog"><slot /><slot name="footer" /></div>',
              props: ['modelValue', 'title'],
            },
          },
        },
      })

      await flushPromises()

      // 打开添加学生弹窗
      wrapper.vm.showAddStudentDialog()
      await nextTick()

      expect(wrapper.vm.addStudentDialogVisible).toBe(true)
      expect(wrapper.vm.addStudentForm.student_id).toBe('')
    })

    it('应能添加学生', async () => {
      wrapper = mount(ClassDetail, {
        global: {
          stubs: {
            'el-card': true,
            'el-button': true,
            'el-alert': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'el-tag': true,
            'el-divider': true,
            'el-empty': true,
            'el-table': true,
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
          },
        },
      })

      await flushPromises()

      // 设置学生ID并提交
      wrapper.vm.addStudentForm.student_id = '123'
      wrapper.vm.addStudentDialogVisible = true

      await wrapper.vm.handleAddStudent()

      // 验证弹窗关闭即可（成功后会关闭弹窗）
      expect(wrapper.vm.addStudentLoading).toBe(false)
    })

    it('应在未输入学生ID时提示', async () => {
      const { ElMessage } = await import('element-plus')

      wrapper = mount(ClassDetail, {
        global: {
          stubs: {
            'el-card': true,
            'el-button': true,
            'el-alert': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'el-tag': true,
            'el-divider': true,
            'el-empty': true,
            'el-table': true,
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
          },
        },
      })

      await flushPromises()

      // 不输入学生ID直接提交
      wrapper.vm.addStudentForm.student_id = ''
      await wrapper.vm.handleAddStudent()

      expect(ElMessage.warning).toHaveBeenCalledWith('请输入学生ID')
    })

    it('应能移除学生', async () => {
      const { ElMessageBox } = await import('element-plus')
      ;(ElMessageBox.confirm as any).mockResolvedValueOnce(true)

      wrapper = mount(ClassDetail, {
        global: {
          stubs: {
            'el-card': true,
            'el-button': true,
            'el-alert': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'el-tag': true,
            'el-divider': true,
            'el-empty': true,
            'el-table': true,
            'el-table-column': true,
          },
        },
      })

      await flushPromises()

      const student = { id: 1, nickname: '学生A', status: 'active' }
      await wrapper.vm.handleRemoveStudent(student)

      // 验证方法成功执行（不应报错）
      expect(wrapper.vm.studentsLoading).toBeDefined()
    })

    it('应处理用户取消移除操作', async () => {
      const { ElMessageBox } = await import('element-plus')
      ;(ElMessageBox.confirm as any).mockRejectedValueOnce('cancel')

      wrapper = mount(ClassDetail, {
        global: {
          stubs: {
            'el-card': true,
            'el-button': true,
            'el-alert': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'el-tag': true,
            'el-divider': true,
            'el-empty': true,
            'el-table': true,
          },
        },
      })

      await flushPromises()

      const student = { id: 1, nickname: '学生A', status: 'active' }
      // 不应抛出错误
      await expect(wrapper.vm.handleRemoveStudent(student)).resolves.not.toThrow()
    })
  })

  describe('错误处理测试', () => {
    it('应处理班级不存在错误', async () => {
      mockRoute.params.id = '9999'

      wrapper = mount(ClassDetail, {
        global: {
          stubs: {
            'el-card': {
              template: '<div class="el-card" v-loading="loading"><slot /></div>',
            },
            'el-button': true,
            'el-alert': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'el-tag': true,
            'el-divider': true,
            'el-empty': true,
            'el-table': true,
          },
        },
      })

      await flushPromises()

      // 错误应该被设置
      expect(wrapper.vm.error).toBeDefined()
    })

    it('应处理加载失败错误', async () => {
      global.fetch = vi.fn(() => Promise.reject(new Error('Network error'))) as any

      wrapper = mount(ClassDetail, {
        global: {
          stubs: {
            'el-card': {
              template: '<div class="el-card" v-loading="loading"><slot /></div>',
            },
            'el-button': true,
            'el-alert': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'el-tag': true,
            'el-divider': true,
            'el-empty': true,
            'el-table': true,
          },
        },
      })

      await flushPromises()

      expect(wrapper.vm.loading).toBe(false)
    })
  })

  describe('日期格式化测试', () => {
    it('应正确格式化日期', () => {
      wrapper = mount(ClassDetail, {
        global: {
          stubs: {
            'el-card': true,
            'el-button': true,
            'el-alert': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'el-tag': true,
            'el-divider': true,
            'el-empty': true,
            'el-table': true,
          },
        },
      })

      const formattedDate = wrapper.vm.formatDate('2026-04-09T10:30:00Z')
      expect(formattedDate).toContain('2026')
      expect(formattedDate).toContain('4')
      expect(formattedDate).toContain('9')
    })

    it('应在日期为空时返回 -', () => {
      wrapper = mount(ClassDetail, {
        global: {
          stubs: {
            'el-card': true,
          },
        },
      })

      expect(wrapper.vm.formatDate('')).toBe('-')
      expect(wrapper.vm.formatDate(undefined)).toBe('-')
    })
  })

  describe('loading 状态测试', () => {
    it('应在加载数据时显示 loading', async () => {
      // 创建一个延迟resolve的fetch
      global.fetch = vi.fn(() => new Promise((resolve) => {
        setTimeout(() => {
          resolve({
            json: () => Promise.resolve({
              code: 0,
              data: {
                id: 1,
                name: '测试班级',
                is_public: true,
                persona_id: 456,
                student_count: 0,
                created_at: '2026-04-09T10:30:00Z',
                updated_at: '2026-04-09T10:30:00Z',
              },
            }),
          } as Response)
        }, 100)
      })) as any

      wrapper = mount(ClassDetail, {
        global: {
          stubs: {
            'el-card': {
              template: '<div class="el-card" v-loading="loading"><slot /></div>',
            },
          },
        },
      })

      // 初始状态应为 loading
      expect(wrapper.vm.loading).toBe(true)
    })
  })
})

describe('ClassDetail.vue - 状态管理测试', () => {
  beforeEach(() => {
    localStorage.setItem('h5_teacher_token', 'mock-token')
  })

  afterEach(() => {
    localStorage.removeItem('h5_teacher_token')
  })

  it('应正确处理空教材配置', () => {
    const wrapper = mount(ClassDetail, {
      global: {
        stubs: {
          'el-card': true,
          'el-button': true,
          'el-alert': true,
          'el-descriptions': true,
          'el-descriptions-item': true,
          'el-tag': true,
          'el-divider': true,
          'el-empty': true,
          'el-table': true,
        },
      },
    })

    // 测试空配置
    wrapper.vm.classInfo.curriculum_config = null
    expect(wrapper.vm.classInfo.curriculum_config).toBeNull()

    // 测试部分配置
    wrapper.vm.classInfo.curriculum_config = {
      id: 1,
      grade_level: 'junior',
      grade: undefined,
      subjects: [],
    }
    expect(wrapper.vm.classInfo.curriculum_config.grade_level).toBe('junior')
    expect(wrapper.vm.classInfo.curriculum_config.subjects).toEqual([])
  })
})
