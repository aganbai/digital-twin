import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import {
  getClassList,
  getClassDetail,
  createClass,
  updateClass,
  deleteClass,
  addStudentToClass,
  removeStudentFromClass,
  getClassStudents,
  type CreateClassParams,
  type UpdateClassParams,
  type CurriculumConfig,
} from '../class'
import { server } from '@/__tests__/mocks/server'
import { http, HttpResponse } from 'msw'

// Mock auth to always return token
vi.mock('@/utils/auth', () => ({
  getToken: vi.fn(() => 'mock-token'),
  clearAuthInfo: vi.fn(),
}))

describe('Class API Module', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getClassList', () => {
    it('should fetch class list successfully', async () => {
      const result = await getClassList()

      expect(result.code).toBe(0)
      expect(result.data).toHaveLength(2)
      expect(result.data[0].name).toBe('三年级数学班')
    })

    it('should handle 500 server error', async () => {
      server.use(
        http.get('/api/classes', () => {
          return HttpResponse.json(
            { code: 50001, message: '数据库服务不可用' },
            { status: 500 }
          )
        })
      )

      const result = await getClassList()
      expect(result.code).toBe(50001)
    })

    it('should handle network error', async () => {
      server.use(
        http.get('/api/classes', () => {
          return HttpResponse.error()
        })
      )

      await expect(getClassList()).rejects.toThrow()
    })
  })

  describe('getClassDetail', () => {
    it('should fetch class detail with curriculum config', async () => {
      const result = await getClassDetail(1)

      expect(result.code).toBe(0)
      expect(result.data.id).toBe(1)
      expect(result.data.curriculum_config).toBeDefined()
      expect(result.data.curriculum_config?.grade_level).toBe('primary_lower')
      expect(result.data.curriculum_config?.subjects).toContain('数学')
    })

    it('should handle class not found (404)', async () => {
      const result = await getClassDetail(9999)
      expect(result.code).toBe(40017)
    })

    it('should handle permission denied (403)', async () => {
      const result = await getClassDetail(8888)
      expect(result.code).toBe(40018)
    })
  })

  describe('createClass', () => {
    const validCreateParams: CreateClassParams = {
      name: '新建班级',
      description: '测试班级描述',
      persona_nickname: '张老师',
      persona_school: '实验小学',
      persona_description: '10年教学经验',
      is_public: true,
      curriculum_config: {
        grade_level: 'primary_lower',
        grade: '二年级',
        subjects: ['语文', '数学'],
        textbook_versions: ['人教版'],
        current_progress: '第一单元',
      },
    }

    it('should create class with curriculum config successfully', async () => {
      const result = await createClass(validCreateParams)

      expect(result.code).toBe(0)
      expect(result.data.id).toBeDefined()
      expect(result.data.name).toBe('新建班级')
      expect(result.data.token).toBeDefined()
    })

    it('should create class without curriculum config', async () => {
      const paramsWithoutConfig: CreateClassParams = {
        name: '临时班级',
        persona_nickname: '李老师',
        persona_school: '实验中学',
        persona_description: '临时班级',
        is_public: false,
      }

      const result = await createClass(paramsWithoutConfig)
      expect(result.code).toBe(0)
      expect(result.data.id).toBe(124)
    })

    it('should handle invalid grade level error', async () => {
      server.use(
        http.post('/api/classes', async () => {
          return HttpResponse.json(
            { code: 40041, message: '无效的学段类型' },
            { status: 400 }
          )
        })
      )

      const result = await createClass({
        ...validCreateParams,
        curriculum_config: {
          grade_level: 'invalid_level',
        },
      } as CreateClassParams)

      expect(result.code).toBe(40041)
    })

    it('should handle duplicate class name (409)', async () => {
      server.use(
        http.post('/api/classes', async () => {
          return HttpResponse.json(
            { code: 40030, message: '该班级名称已存在' },
            { status: 409 }
          )
        })
      )

      const result = await createClass(validCreateParams)
      expect(result.code).toBe(40030)
    })

    it('should handle missing required fields (400)', async () => {
      server.use(
        http.post('/api/classes', async () => {
          return HttpResponse.json(
            { code: 40004, message: '请求参数无效' },
            { status: 400 }
          )
        })
      )

      const result = await createClass({
        name: '测试班级',
      } as CreateClassParams)

      expect(result.code).toBe(40004)
    })
  })

  describe('updateClass', () => {
    const validUpdateParams: UpdateClassParams = {
      name: '更新后的班级名',
      description: '更新后的描述',
      is_public: false,
      curriculum_config: {
        grade_level: 'junior',
        grade: '七年级',
        subjects: ['英语'],
        textbook_versions: ['外研版'],
        current_progress: '第二单元',
      },
    }

    it('should update class with curriculum config successfully', async () => {
      server.use(
        http.put('/api/classes/1', async () => {
          return HttpResponse.json({
            code: 0,
            message: 'success',
            data: {
              id: 1,
              name: '更新后的班级名',
              description: '更新后的描述',
              persona_id: 456,
            },
          })
        })
      )

      const result = await updateClass(1, validUpdateParams)
      expect(result.code).toBe(0)
      expect(result.data.id).toBe(1)
    })

    it('should update class without curriculum config', async () => {
      server.use(
        http.put('/api/classes/1', async () => {
          return HttpResponse.json({
            code: 0,
            data: {
              id: 1,
              name: '仅更新名称',
              persona_id: 456,
            },
          })
        })
      )

      const result = await updateClass(1, { name: '仅更新名称' })
      expect(result.code).toBe(0)
      expect(result.data.name).toBe('仅更新名称')
    })

    it('should handle class not found error (404)', async () => {
      const result = await updateClass(9999, validUpdateParams)
      expect(result.code).toBe(40017)
    })

    it('should handle permission denied error (403)', async () => {
      const result = await updateClass(8888, validUpdateParams)
      expect(result.code).toBe(40018)
    })

    it('should handle duplicate name conflict (409)', async () => {
      server.use(
        http.put('/api/classes/1', async () => {
          return HttpResponse.json(
            { code: 40016, message: '班级名称已存在' },
            { status: 409 }
          )
        })
      )

      const result = await updateClass(1, { name: '已存在的班级名' })
      expect(result.code).toBe(40016)
    })
  })

  describe('deleteClass', () => {
    it('should delete class successfully', async () => {
      server.use(
        http.delete('/api/teacher/classes/1', () => {
          return HttpResponse.json({
            code: 0,
            message: 'success',
            data: null,
          })
        })
      )

      const result = await deleteClass(1)
      expect(result.code).toBe(0)
    })

    it('should handle non-existent class deletion (404)', async () => {
      const result = await deleteClass(9999)
      expect(result.code).toBe(40017)
    })
  })

  describe('Class Student Management', () => {
    it('should add student to class successfully', async () => {
      server.use(
        http.post('/api/teacher/classes/1/students', async () => {
          return HttpResponse.json({
            code: 0,
            message: 'success',
            data: null,
          })
        })
      )

      const result = await addStudentToClass(1, 123)
      expect(result.code).toBe(0)
    })

    it('should remove student from class successfully', async () => {
      server.use(
        http.delete('/api/teacher/classes/1/students/123', () => {
          return HttpResponse.json({
            code: 0,
            message: 'success',
            data: null,
          })
        })
      )

      const result = await removeStudentFromClass(1, 123)
      expect(result.code).toBe(0)
    })

    it('should get class students list', async () => {
      server.use(
        http.get('/api/teacher/classes/1/students', () => {
          return HttpResponse.json({
            code: 0,
            message: 'success',
            data: [
              { id: 1, name: '学生A', student_no: '2024001' },
              { id: 2, name: '学生B', student_no: '2024002' },
            ],
          })
        })
      )

      const result = await getClassStudents(1)
      expect(result.code).toBe(0)
      expect(result.data).toHaveLength(2)
    })
  })
})

describe('CurriculumConfig Type Check', () => {
  it('should have correct interface structure', () => {
    const config: CurriculumConfig = {
      id: 1,
      grade_level: 'primary_lower',
      grade: '三年级',
      subjects: ['数学', '语文'],
      textbook_versions: ['人教版'],
      custom_textbooks: [],
      current_progress: '第三单元',
    }

    expect(config.grade_level).toBe('primary_lower')
    expect(config.subjects).toHaveLength(2)
    expect(config.textbook_versions).toContain('人教版')
  })

  it('should allow partial config for updates', () => {
    const partialConfig: CurriculumConfig = {
      grade_level: 'junior',
    }

    expect(partialConfig.grade_level).toBe('junior')
    expect(partialConfig.subjects).toBeUndefined()
  })

  it('should support university level with custom textbooks', () => {
    const universityConfig: CurriculumConfig = {
      grade_level: 'university',
      grade: '大一',
      subjects: ['计算机基础'],
      textbook_versions: [],
      custom_textbooks: ['《计算机导论》', '《编程基础》'],
      current_progress: '第5章',
    }

    expect(universityConfig.grade_level).toBe('university')
    expect(universityConfig.custom_textbooks).toHaveLength(2)
  })

  it('should support adult life categories', () => {
    const adultLifeConfig: CurriculumConfig = {
      grade_level: 'adult_life',
      subjects: ['中餐', '烘焙'],
    }

    expect(adultLifeConfig.grade_level).toBe('adult_life')
  })

  it('should support adult professional categories', () => {
    const adultProConfig: CurriculumConfig = {
      grade_level: 'adult_professional',
      subjects: ['编程', '设计'],
    }

    expect(adultProConfig.grade_level).toBe('adult_professional')
  })

  it('should support preschool level', () => {
    const preschoolConfig: CurriculumConfig = {
      grade_level: 'preschool',
      grade: '幼儿园大班',
      subjects: ['手工', '音乐'],
    }

    expect(preschoolConfig.grade_level).toBe('preschool')
  })

  it('should support null curriculum_config', () => {
    const nullConfig: CurriculumConfig | null = null
    expect(nullConfig).toBeNull()
  })
})

describe('Grade Level Options Validation', () => {
  const validGradeLevels = [
    'preschool',
    'primary_lower',
    'primary_upper',
    'junior',
    'senior',
    'university',
    'adult_life',
    'adult_professional',
  ]

  it('should accept all valid grade levels', () => {
    validGradeLevels.forEach((level) => {
      const config: CurriculumConfig = { grade_level: level }
      expect(config.grade_level).toBe(level)
    })
  })

  it('should validate grade level values match API spec', () => {
    const gradeLevelLabels: Record<string, string> = {
      preschool: '学前班',
      primary_lower: '小学低年级',
      primary_upper: '小学高年级',
      junior: '初中',
      senior: '高中',
      university: '大学及以上',
      adult_life: '成人生活技能',
      adult_professional: '成人职业培训',
    }

    validGradeLevels.forEach((level) => {
      expect(gradeLevelLabels[level]).toBeDefined()
    })
  })
})

describe('API Integration Scenarios', () => {
  it('should handle complete class creation flow with config', async () => {
    const createParams: CreateClassParams = {
      name: '三年级数学班',
      description: '小学数学培优班级',
      persona_nickname: '王老师',
      persona_school: '实验小学',
      persona_description: '10年数学教学经验',
      is_public: true,
      curriculum_config: {
        grade_level: 'primary_lower',
        grade: '三年级',
        subjects: ['数学'],
        textbook_versions: ['人教版', '北师大版'],
        custom_textbooks: ['《小学奥数启蒙》'],
        current_progress: '第三单元 乘法初步',
      },
    }

    const result = await createClass(createParams)

    expect(result.code).toBe(0)
    expect(result.data).toBeDefined()
  })

  it('should handle class update flow with config modification', async () => {
    server.use(
      http.put('/api/classes/1', async () => {
        return HttpResponse.json({
          code: 0,
          message: 'success',
          data: { id: 1 },
        })
      })
    )

    const updateParams: UpdateClassParams = {
      name: '三年级数学班（已更新）',
      description: '更新后的描述',
      is_public: false,
      curriculum_config: {
        grade_level: 'primary_lower',
        grade: '三年级',
        subjects: ['数学', '奥数'],
        textbook_versions: ['人教版'],
        custom_textbooks: ['《小学奥数进阶》'],
        current_progress: '第四单元',
      },
    }

    const result = await updateClass(1, updateParams)
    expect(result.code).toBe(0)
  })

  it('should handle class detail fetch with curriculum config', async () => {
    const result = await getClassDetail(1)

    expect(result.code).toBe(0)
    expect(result.data.curriculum_config).toBeDefined()
    expect(result.data.curriculum_config?.subjects).toBeInstanceOf(Array)
    expect(result.data.curriculum_config?.textbook_versions).toBeInstanceOf(Array)
  })

  it('should handle university level curriculum config', async () => {
    const universityParams: CreateClassParams = {
      name: '计算机导论班',
      persona_nickname: '张教授',
      persona_school: '理工大学',
      persona_description: '计算机科学教授',
      curriculum_config: {
        grade_level: 'university',
        grade: '大一',
        subjects: ['计算机基础', '编程入门'],
        custom_textbooks: ['《计算机导论》', '《Python编程》'],
        current_progress: '第3章 数据类型',
      },
    }

    const result = await createClass(universityParams)
    expect(result.code).toBe(0)
  })

  it('should handle adult life skills curriculum config', async () => {
    const adultParams: CreateClassParams = {
      name: '烹饪兴趣班',
      persona_nickname: '厨师长',
      persona_school: '烹饪学校',
      persona_description: '资深中餐厨师',
      curriculum_config: {
        grade_level: 'adult_life',
        subjects: ['中餐', '烘焙'],
        current_progress: '学习制作红烧肉',
      },
    }

    const result = await createClass(adultParams)
    expect(result.code).toBe(0)
  })

  it('should handle adult professional curriculum config', async () => {
    const proParams: CreateClassParams = {
      name: '前端开发训练营',
      persona_nickname: '工程师',
      persona_school: '技术学院',
      persona_description: '资深前端工程师',
      curriculum_config: {
        grade_level: 'adult_professional',
        subjects: ['编程', '设计'],
        current_progress: 'Vue3 Composition API',
      },
    }

    const result = await createClass(proParams)
    expect(result.code).toBe(0)
  })
})
