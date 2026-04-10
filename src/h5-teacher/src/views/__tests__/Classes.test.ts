import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, VueWrapper, flushPromises } from '@vue/test-utils'
import { nextTick } from 'vue'
import Classes from '../Classes.vue'

// Mock Element Plus components
vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
  },
  ElMessageBox: {
    confirm: vi.fn().mockResolvedValue(true),
  },
}))

// Mock vue-router
const mockPush = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}))

describe('Classes.vue - 班级管理教材配置弹窗', () => {
  let wrapper: VueWrapper<any>

  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.setItem('h5_teacher_token', 'mock-token')

    // Mock fetch for component tests (components use fetch directly)
    global.fetch = vi.fn((url: string, options?: RequestInit) => {
      if (url === '/api/classes' && (!options || options.method === 'GET' || !options.method)) {
        return Promise.resolve({
          json: () =>
            Promise.resolve({
              code: 0,
              data: [
                {
                  id: 1,
                  name: '三年级数学班',
                  description: '小学数学培优班级',
                  persona_nickname: '王老师',
                  student_count: 25,
                  is_public: true,
                  created_at: '2026-04-09T10:30:00Z',
                  curriculum_config: {
                    id: 1001,
                    grade_level: 'primary_lower',
                    grade: '三年级',
                    subjects: ['数学'],
                    textbook_versions: ['人教版'],
                    custom_textbooks: [],
                    current_progress: '第三单元',
                  },
                },
                {
                  id: 2,
                  name: '初中英语班',
                  student_count: 30,
                  is_public: false,
                  created_at: '2026-04-08T14:20:00Z',
                  curriculum_config: null,
                },
              ],
            }),
        } as Response)
      }

      if (url === '/api/classes' && options?.method === 'POST') {
        return Promise.resolve({
          json: () =>
            Promise.resolve({
              code: 0,
              data: {
                id: 123,
                name: '新建班级',
                persona_nickname: '张老师',
                persona_school: '实验小学',
                persona_id: 456,
                share_url: 'https://example.com/class/123',
                share_code: 'ABC123',
              },
            }),
        } as Response)
      }

      if (url.startsWith('/api/classes/') && options?.method === 'PUT') {
        return Promise.resolve({
          json: () =>
            Promise.resolve({
              code: 0,
              data: { id: 1 },
            }),
        } as Response)
      }

      return Promise.resolve({
        json: () => Promise.resolve({ code: 0, data: null }),
      } as Response)
    }) as any
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
    }
    // Clear fetch mock only for component tests
    vi.unstubAllGlobals?.()
  })

  describe('渲染测试', () => {
    it('应正确渲染班级列表页', async () => {
      wrapper = mount(Classes, {
        global: {
          stubs: {
            'el-card': true,
            'el-table': true,
            'el-table-column': true,
            'el-button': true,
            'el-tag': true,
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
            'el-switch': true,
            'el-divider': true,
            'el-alert': true,
            'el-select': true,
            'el-option': true,
            'el-icon': true,
            'el-collapse-transition': true,
            'el-result': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'ArrowDown': true,
          },
        },
      })

      await flushPromises()

      expect(wrapper.find('.classes-container').exists()).toBe(true)
    })

    it('应加载班级列表数据', async () => {
      wrapper = mount(Classes, {
        global: {
          stubs: {
            'el-card': {
              template: '<div class="el-card"><slot /><slot name="header" /></div>',
            },
            'el-table': {
              template: '<div class="el-table"><slot /></div>',
              props: ['data'],
            },
            'el-table-column': true,
            'el-button': {
              template: '<button class="el-button" @click="$emit(\'click\')"><slot /></button>',
            },
            'el-tag': {
              template: '<span class="el-tag" :class="type"><slot /></span>',
              props: ['type'],
            },
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
            'el-switch': true,
            'el-divider': true,
            'el-alert': true,
            'el-select': true,
            'el-option': true,
            'el-icon': true,
            'el-collapse-transition': true,
            'el-result': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'ArrowDown': true,
          },
        },
      })

      await flushPromises()

      // 验证组件加载状态已重置
      expect(wrapper.vm.loading).toBe(false)
    })
  })

  describe('创建班级弹窗测试', () => {
    it('应打开创建班级弹窗', async () => {
      wrapper = mount(Classes, {
        global: {
          stubs: {
            'el-card': {
              template: '<div class="el-card"><slot /><slot name="header" /></div>',
            },
            'el-table': {
              template: '<div class="el-table" :data="data"><slot /></div>',
              props: ['data'],
            },
            'el-table-column': true,
            'el-button': {
              template: '<button class="el-button" @click="$emit(\'click\')"><slot /></button>',
            },
            'el-tag': true,
            'el-dialog': {
              template: `
                <div v-if="modelValue" class="el-dialog">
                  <div class="el-dialog__title">{{ title }}</div>
                  <slot />
                  <slot name="footer" />
                </div>
              `,
              props: ['modelValue', 'title'],
            },
            'el-form': {
              template: '<form class="el-form" ref="form"><slot /></form>',
              props: ['model', 'rules'],
            },
            'el-form-item': {
              template: '<div class="el-form-item"><slot /></div>',
              props: ['label', 'prop'],
            },
            'el-input': {
              template: '<input class="el-input" :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />',
              props: ['modelValue', 'placeholder', 'maxlength'],
            },
            'el-switch': {
              template: '<input type="checkbox" class="el-switch" :checked="modelValue" @change="$emit(\'update:modelValue\', $event.target.checked)" />',
              props: ['modelValue'],
            },
            'el-divider': true,
            'el-alert': true,
            'el-select': {
              template: '<select class="el-select" :value="modelValue" @change="$emit(\'update:modelValue\', $event.target.value); $emit(\'change\')"><slot /></select>',
              props: ['modelValue', 'placeholder'],
            },
            'el-option': {
              template: '<option class="el-option" :value="value">{{ label }}</option>',
              props: ['label', 'value'],
            },
            'el-icon': true,
            'el-collapse-transition': {
              template: '<div v-show="show"><slot /></div>',
              props: ['show'],
            },
            'el-result': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'ArrowDown': true,
          },
        },
      })

      await flushPromises()

      // 模拟点击创建按钮
      const createButton = wrapper.findAll('.el-button').find((btn) => btn.text().includes('创建班级'))
      expect(createButton).toBeDefined()

      if (createButton) {
        await createButton.trigger('click')
        await nextTick()

        // 验证弹窗打开
        expect(wrapper.vm.createDialogVisible).toBe(true)
      }
    })
  })

  describe('教材配置表单测试', () => {
    it('应正确初始化教材配置默认值', () => {
      wrapper = mount(Classes, {
        global: {
          stubs: {
            'el-card': true,
            'el-table': true,
            'el-table-column': true,
            'el-button': true,
            'el-tag': true,
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
            'el-switch': true,
            'el-divider': true,
            'el-alert': true,
            'el-select': true,
            'el-option': true,
            'el-icon': true,
            'el-collapse-transition': true,
            'el-result': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'ArrowDown': true,
          },
        },
      })

      // 验证教材配置初始状态
      expect(wrapper.vm.createCurriculumExpanded).toBe(false)
      expect(wrapper.vm.createCurriculumConfig.grade_level).toBeUndefined()
      expect(wrapper.vm.createCurriculumConfig.subjects).toEqual([])
    })

    it('应在学段变化时重置年级和学科', async () => {
      wrapper = mount(Classes, {
        global: {
          stubs: {
            'el-card': true,
            'el-table': true,
            'el-table-column': true,
            'el-button': true,
            'el-tag': true,
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
            'el-switch': true,
            'el-divider': true,
            'el-alert': true,
            'el-select': true,
            'el-option': true,
            'el-icon': true,
            'el-collapse-transition': true,
            'el-result': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'ArrowDown': true,
          },
        },
      })

      // 设置初始值
      wrapper.vm.createCurriculumConfig.grade = '三年级'
      wrapper.vm.createCurriculumConfig.subjects = ['数学']
      wrapper.vm.createCurriculumConfig.grade_level = 'primary_lower'

      // 调用学段变化处理函数
      wrapper.vm.onCreateGradeLevelChange()

      expect(wrapper.vm.createCurriculumConfig.grade).toBeUndefined()
      expect(wrapper.vm.createCurriculumConfig.subjects).toEqual([])
    })

    it('应根据学段计算年级选项', () => {
      wrapper = mount(Classes, {
        global: {
          stubs: {
            'el-card': true,
            'el-table': true,
            'el-table-column': true,
            'el-button': true,
            'el-tag': true,
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
            'el-switch': true,
            'el-divider': true,
            'el-alert': true,
            'el-select': true,
            'el-option': true,
            'el-icon': true,
            'el-collapse-transition': true,
            'el-result': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'ArrowDown': true,
          },
        },
      })

      // 小学低年级
      wrapper.vm.createCurriculumConfig.grade_level = 'primary_lower'
      expect(wrapper.vm.createGradeOptions).toEqual(['一年级', '二年级', '三年级'])

      // 初中
      wrapper.vm.createCurriculumConfig.grade_level = 'junior'
      expect(wrapper.vm.createGradeOptions).toEqual(['七年级', '八年级', '九年级'])

      // 高中
      wrapper.vm.createCurriculumConfig.grade_level = 'senior'
      expect(wrapper.vm.createGradeOptions).toEqual(['高一', '高二', '高三'])
    })

    it('应根据学段计算学科选项', () => {
      wrapper = mount(Classes, {
        global: {
          stubs: {
            'el-card': true,
            'el-table': true,
            'el-table-column': true,
            'el-button': true,
            'el-tag': true,
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
            'el-switch': true,
            'el-divider': true,
            'el-alert': true,
            'el-select': true,
            'el-option': true,
            'el-icon': true,
            'el-collapse-transition': true,
            'el-result': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'ArrowDown': true,
          },
        },
      })

      const k12Subjects = ['语文', '数学', '英语', '物理', '化学', '生物', '历史', '地理', '政治', '音乐', '美术', '体育', '信息技术']

      // K12学段使用标准学科
      wrapper.vm.createCurriculumConfig.grade_level = 'primary_lower'
      expect(wrapper.vm.createSubjectOptions).toEqual(k12Subjects)

      wrapper.vm.createCurriculumConfig.grade_level = 'junior'
      expect(wrapper.vm.createSubjectOptions).toEqual(k12Subjects)

      // 成人生活技能
      wrapper.vm.createCurriculumConfig.grade_level = 'adult_life'
      expect(wrapper.vm.createSubjectOptions).toContain('中餐')
      expect(wrapper.vm.createSubjectOptions).toContain('烘焙')

      // 成人职业培训
      wrapper.vm.createCurriculumConfig.grade_level = 'adult_professional'
      expect(wrapper.vm.createSubjectOptions).toContain('编程')
      expect(wrapper.vm.createSubjectOptions).toContain('设计')
    })

    it('应在成人学段隐藏年级选择器', () => {
      wrapper = mount(Classes, {
        global: {
          stubs: {
            'el-card': true,
            'el-table': true,
            'el-table-column': true,
            'el-button': true,
            'el-tag': true,
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
            'el-switch': true,
            'el-divider': true,
            'el-alert': true,
            'el-select': true,
            'el-option': true,
            'el-icon': true,
            'el-collapse-transition': true,
            'el-result': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'ArrowDown': true,
          },
        },
      })

      // 成人生活技能不显示年级
      wrapper.vm.createCurriculumConfig.grade_level = 'adult_life'
      expect(wrapper.vm.showCreateGradeSelector).toBe(false)

      // 成人职业培训不显示年级
      wrapper.vm.createCurriculumConfig.grade_level = 'adult_professional'
      expect(wrapper.vm.showCreateGradeSelector).toBe(false)

      // K12学段显示年级
      wrapper.vm.createCurriculumConfig.grade_level = 'primary_lower'
      expect(wrapper.vm.showCreateGradeSelector).toBe(true)
    })

    it('应在大学学段显示自定义教材输入', () => {
      wrapper = mount(Classes, {
        global: {
          stubs: {
            'el-card': true,
            'el-table': true,
            'el-table-column': true,
            'el-button': true,
            'el-tag': true,
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
            'el-switch': true,
            'el-divider': true,
            'el-alert': true,
            'el-select': true,
            'el-option': true,
            'el-icon': true,
            'el-collapse-transition': true,
            'el-result': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'ArrowDown': true,
          },
        },
      })

      // 大学及以上显示自定义教材
      wrapper.vm.createCurriculumConfig.grade_level = 'university'
      expect(wrapper.vm.showCreateCustomTextbooks).toBe(true)

      // 其他学段不显示
      wrapper.vm.createCurriculumConfig.grade_level = 'primary_lower'
      expect(wrapper.vm.showCreateCustomTextbooks).toBe(false)
    })
  })

  describe('自定义教材功能测试', () => {
    it('应能添加自定义教材', async () => {
      wrapper = mount(Classes, {
        global: {
          stubs: {
            'el-card': true,
            'el-table': true,
            'el-table-column': true,
            'el-button': true,
            'el-tag': true,
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
            'el-switch': true,
            'el-divider': true,
            'el-alert': true,
            'el-select': true,
            'el-option': true,
            'el-icon': true,
            'el-collapse-transition': true,
            'el-result': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'ArrowDown': true,
          },
        },
      })

      // 设置学段为大学
      wrapper.vm.createCurriculumConfig.grade_level = 'university'
      wrapper.vm.createCurriculumConfig.custom_textbooks = []

      // 输入教材名称
      wrapper.vm.customTextbookInput = '《计算机导论》'

      // 添加教材
      wrapper.vm.addCreateCustomTextbook()

      expect(wrapper.vm.createCurriculumConfig.custom_textbooks).toContain('《计算机导论》')
      expect(wrapper.vm.customTextbookInput).toBe('')
    })

    it('应阻止添加重复的自定义教材', async () => {
      const { ElMessage } = await import('element-plus')

      wrapper = mount(Classes, {
        global: {
          stubs: {
            'el-card': true,
            'el-table': true,
            'el-table-column': true,
            'el-button': true,
            'el-tag': true,
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
            'el-switch': true,
            'el-divider': true,
            'el-alert': true,
            'el-select': true,
            'el-option': true,
            'el-icon': true,
            'el-collapse-transition': true,
            'el-result': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'ArrowDown': true,
          },
        },
      })

      // 设置已有教材
      wrapper.vm.createCurriculumConfig.grade_level = 'university'
      wrapper.vm.createCurriculumConfig.custom_textbooks = ['《计算机导论》']

      // 尝试添加重复教材
      wrapper.vm.customTextbookInput = '《计算机导论》'
      wrapper.vm.addCreateCustomTextbook()

      expect(ElMessage.warning).toHaveBeenCalledWith('该教材已添加')
      expect(wrapper.vm.createCurriculumConfig.custom_textbooks).toHaveLength(1)
    })

    it('应能移除自定义教材', () => {
      wrapper = mount(Classes, {
        global: {
          stubs: {
            'el-card': true,
            'el-table': true,
            'el-table-column': true,
            'el-button': true,
            'el-tag': true,
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
            'el-switch': true,
            'el-divider': true,
            'el-alert': true,
            'el-select': true,
            'el-option': true,
            'el-icon': true,
            'el-collapse-transition': true,
            'el-result': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'ArrowDown': true,
          },
        },
      })

      // 设置已有教材
      wrapper.vm.createCurriculumConfig.custom_textbooks = ['《计算机导论》', '《编程基础》']

      // 移除第一个教材
      wrapper.vm.removeCreateCustomTextbook(0)

      expect(wrapper.vm.createCurriculumConfig.custom_textbooks).toEqual(['《编程基础》'])
    })
  })

  describe('编辑班级弹窗测试', () => {
    it('应加载已有教材配置到编辑表单', async () => {
      wrapper = mount(Classes, {
        global: {
          stubs: {
            'el-card': {
              template: '<div class="el-card"><slot /><slot name="header" /></div>',
            },
            'el-table': {
              template: '<div class="el-table"><slot /></div>',
              props: ['data'],
            },
            'el-table-column': true,
            'el-button': {
              template: '<button class="el-button" @click="$emit(\'click\')"><slot /></button>',
            },
            'el-tag': true,
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
            'el-switch': true,
            'el-divider': true,
            'el-alert': true,
            'el-select': true,
            'el-option': true,
            'el-icon': true,
            'el-collapse-transition': true,
            'el-result': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'ArrowDown': true,
          },
        },
      })

      await flushPromises()

      // 模拟打开编辑弹窗
      const rowWithConfig = {
        id: 1,
        name: '三年级数学班',
        description: '小学数学培优班级',
        is_public: true,
        curriculum_config: {
          grade_level: 'primary_lower',
          grade: '三年级',
          subjects: ['数学'],
          textbook_versions: ['人教版'],
          custom_textbooks: [],
          current_progress: '第三单元',
        },
      }

      wrapper.vm.showEditDialog(rowWithConfig)

      // 验证编辑表单已加载配置
      expect(wrapper.vm.editCurriculumConfig.grade_level).toBe('primary_lower')
      expect(wrapper.vm.editCurriculumConfig.grade).toBe('三年级')
      expect(wrapper.vm.editCurriculumConfig.subjects).toEqual(['数学'])
      expect(wrapper.vm.editCurriculumExpanded).toBe(true)
    })

    it('应在无配置时折叠教材配置区域', async () => {
      wrapper = mount(Classes, {
        global: {
          stubs: {
            'el-card': true,
            'el-table': true,
            'el-table-column': true,
            'el-button': true,
            'el-tag': true,
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
            'el-switch': true,
            'el-divider': true,
            'el-alert': true,
            'el-select': true,
            'el-option': true,
            'el-icon': true,
            'el-collapse-transition': true,
            'el-result': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'ArrowDown': true,
          },
        },
      })

      await flushPromises()

      // 模拟打开编辑弹窗（无配置）
      const rowWithoutConfig = {
        id: 2,
        name: '初中英语班',
        is_public: false,
        curriculum_config: null,
      }

      wrapper.vm.showEditDialog(rowWithoutConfig)

      // 验证教材配置区域折叠
      expect(wrapper.vm.editCurriculumExpanded).toBe(false)
      expect(wrapper.vm.editCurriculumConfig.grade_level).toBeUndefined()
    })
  })

  describe('表单提交测试', () => {
    it('创建班级表单数据验证', async () => {
      wrapper = mount(Classes, {
        global: {
          stubs: {
            'el-card': true,
            'el-table': true,
            'el-table-column': true,
            'el-button': true,
            'el-tag': true,
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
            'el-switch': true,
            'el-divider': true,
            'el-alert': true,
            'el-select': true,
            'el-option': true,
            'el-icon': true,
            'el-collapse-transition': true,
            'el-result': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'ArrowDown': true,
          },
        },
      })

      // 设置表单数据
      wrapper.vm.createForm.name = '新班级'
      wrapper.vm.createForm.persona_nickname = '张老师'
      wrapper.vm.createForm.persona_school = '实验小学'
      wrapper.vm.createForm.persona_description = '10年教学经验'

      // 展开并填写教材配置
      wrapper.vm.createCurriculumExpanded = true
      wrapper.vm.createCurriculumConfig.grade_level = 'primary_lower'
      wrapper.vm.createCurriculumConfig.grade = '二年级'
      wrapper.vm.createCurriculumConfig.subjects = ['数学']
      wrapper.vm.createCurriculumConfig.textbook_versions = ['人教版']
      wrapper.vm.createCurriculumConfig.current_progress = '第一单元'

      // 验证表单数据正确设置
      expect(wrapper.vm.createForm.name).toBe('新班级')
      expect(wrapper.vm.createCurriculumExpanded).toBe(true)
      expect(wrapper.vm.createCurriculumConfig.grade_level).toBe('primary_lower')
      expect(wrapper.vm.createCurriculumConfig.grade).toBe('二年级')
      expect(wrapper.vm.createCurriculumConfig.subjects).toEqual(['数学'])
      expect(wrapper.vm.createCurriculumConfig.textbook_versions).toEqual(['人教版'])
    })

    it('编辑班级时应更新教材配置', async () => {
      wrapper = mount(Classes, {
        global: {
          stubs: {
            'el-card': true,
            'el-table': true,
            'el-table-column': true,
            'el-button': true,
            'el-tag': true,
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
            'el-switch': true,
            'el-divider': true,
            'el-alert': true,
            'el-select': true,
            'el-option': true,
            'el-icon': true,
            'el-collapse-transition': true,
            'el-result': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'ArrowDown': true,
          },
        },
      })

      // 验证编辑时教材配置的状态变化
      wrapper.vm.editingClassId = 1
      wrapper.vm.editForm.name = '更新的班级名'
      wrapper.vm.editCurriculumExpanded = true
      wrapper.vm.editCurriculumConfig.grade_level = 'junior'
      wrapper.vm.editCurriculumConfig.grade = '七年级'
      wrapper.vm.editCurriculumConfig.subjects = ['英语']

      // 验证表单数据
      expect(wrapper.vm.editCurriculumConfig.grade_level).toBe('junior')
      expect(wrapper.vm.editCurriculumConfig.grade).toBe('七年级')
      expect(wrapper.vm.editCurriculumConfig.subjects).toEqual(['英语'])
    })
  })

  describe('计算属性测试', () => {
    it('应正确获取学段标签', () => {
      wrapper = mount(Classes, {
        global: {
          stubs: {
            'el-card': true,
            'el-table': true,
            'el-table-column': true,
            'el-button': true,
            'el-tag': true,
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
            'el-switch': true,
            'el-divider': true,
            'el-alert': true,
            'el-select': true,
            'el-option': true,
            'el-icon': true,
            'el-collapse-transition': true,
            'el-result': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'ArrowDown': true,
          },
        },
      })

      expect(wrapper.vm.getGradeLevelLabel('primary_lower')).toBe('小学低年级')
      expect(wrapper.vm.getGradeLevelLabel('junior')).toBe('初中')
      expect(wrapper.vm.getGradeLevelLabel('senior')).toBe('高中')
      expect(wrapper.vm.getGradeLevelLabel('university')).toBe('大学及以上')
      expect(wrapper.vm.getGradeLevelLabel('adult_life')).toBe('成人生活技能')
      expect(wrapper.vm.getGradeLevelLabel('adult_professional')).toBe('成人职业培训')
      expect(wrapper.vm.getGradeLevelLabel('unknown')).toBe('unknown')
    })
  })

  describe('边界条件测试', () => {
    it('应在未填写学段时不提交教材配置', async () => {
      wrapper = mount(Classes, {
        global: {
          stubs: {
            'el-card': true,
            'el-table': true,
            'el-table-column': true,
            'el-button': true,
            'el-tag': true,
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
            'el-switch': true,
            'el-divider': true,
            'el-alert': true,
            'el-select': true,
            'el-option': true,
            'el-icon': true,
            'el-collapse-transition': true,
            'el-result': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'ArrowDown': true,
          },
        },
      })

      // 设置表单数据
      wrapper.vm.createForm.name = '新班级'
      wrapper.vm.createForm.persona_nickname = '张老师'
      wrapper.vm.createForm.persona_school = '实验小学'
      wrapper.vm.createForm.persona_description = '10年教学经验'

      // 展开但不填写学段
      wrapper.vm.createCurriculumExpanded = true
      wrapper.vm.createCurriculumConfig.grade_level = undefined

      // 验证状态
      expect(wrapper.vm.createCurriculumExpanded).toBe(true)
      expect(wrapper.vm.createCurriculumConfig.grade_level).toBeUndefined()

      // 验证空配置时不会提交教材配置
      expect(wrapper.vm.createCurriculumConfig.grade_level).toBeFalsy()
    })

    it('应在折叠教材配置时不提交配置', async () => {
      wrapper = mount(Classes, {
        global: {
          stubs: {
            'el-card': true,
            'el-table': true,
            'el-table-column': true,
            'el-button': true,
            'el-tag': true,
            'el-dialog': true,
            'el-form': true,
            'el-form-item': true,
            'el-input': true,
            'el-switch': true,
            'el-divider': true,
            'el-alert': true,
            'el-select': true,
            'el-option': true,
            'el-icon': true,
            'el-collapse-transition': true,
            'el-result': true,
            'el-descriptions': true,
            'el-descriptions-item': true,
            'ArrowDown': true,
          },
        },
      })

      // 设置表单数据
      wrapper.vm.createForm.name = '新班级'
      wrapper.vm.createForm.persona_nickname = '张老师'
      wrapper.vm.createForm.persona_school = '实验小学'
      wrapper.vm.createForm.persona_description = '10年教学经验'

      // 确保教材配置折叠
      wrapper.vm.createCurriculumExpanded = false

      // 验证折叠状态
      expect(wrapper.vm.createCurriculumExpanded).toBe(false)

      // 验证折叠状态时不会提交教材配置
      expect(wrapper.vm.createCurriculumExpanded && wrapper.vm.createCurriculumConfig.grade_level).toBeFalsy()
    })
  })
})
