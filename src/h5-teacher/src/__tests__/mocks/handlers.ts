import { http, HttpResponse } from 'msw'

// 模拟班级数据
export const mockClasses = [
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
      textbook_versions: ['人教版', '北师大版'],
      custom_textbooks: [],
      current_progress: '第三单元 乘法初步',
    },
  },
  {
    id: 2,
    name: '初中英语班',
    description: '初中英语提高班',
    persona_nickname: '李老师',
    student_count: 30,
    is_public: false,
    created_at: '2026-04-08T14:20:00Z',
    curriculum_config: null,
  },
]

// 模拟班级详情
export const mockClassDetail = {
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
    textbook_versions: ['人教版'],
    custom_textbooks: ['《小学奥数启蒙》'],
    current_progress: '第三单元',
  },
}

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

// Helper: verify auth header
const verifyAuth = (request: Request) => {
  const authHeader = request.headers.get('Authorization')
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return HttpResponse.json(
      { code: 40101, message: '未登录或 token 已过期' },
      { status: 401 }
    )
  }
  return null
}

// API Mock Handlers
export const handlers = [
  // 获取班级列表
  http.get('/api/classes', ({ request }) => {
    const authError = verifyAuth(request)
    if (authError) return authError

    return HttpResponse.json({
      code: 0,
      message: 'success',
      data: mockClasses,
    })
  }),

  // 获取班级详情
  http.get('/api/classes/:id', ({ params, request }) => {
    const authError = verifyAuth(request)
    if (authError) return authError

    const id = Number(params.id)
    const classInfo = mockClasses.find((c) => c.id === id)

    if (!classInfo) {
      return HttpResponse.json(
        { code: 40017, message: '班级不存在' },
        { status: 404 }
      )
    }

    return HttpResponse.json({
      code: 0,
      message: 'success',
      data: id === 1 ? { ...mockClassDetail } : { ...classInfo, persona_id: 456, teacher_id: 789, is_active: true, updated_at: classInfo.created_at },
    })
  }),

  // 创建班级
  http.post('/api/classes', async ({ request }) => {
    const authError = verifyAuth(request)
    if (authError) return authError

    const body = (await request.json()) as any

    // 参数校验
    if (!body.name || !body.persona_nickname || !body.persona_school || !body.persona_description) {
      return HttpResponse.json(
        { code: 40004, message: '请求参数无效' },
        { status: 400 }
      )
    }

    // 模拟班级名称已存在
    if (body.name === '三年级数学班') {
      return HttpResponse.json(
        { code: 40030, message: '该班级名称已存在' },
        { status: 409 }
      )
    }

    // 模拟无效学段类型
    if (body.curriculum_config?.grade_level) {
      if (!validGradeLevels.includes(body.curriculum_config.grade_level)) {
        return HttpResponse.json(
          { code: 40041, message: '无效的学段类型' },
          { status: 400 }
        )
      }
    }

    const newClass = {
      id: 123,
      name: body.name,
      description: body.description || '',
      is_public: body.is_public !== false,
      persona_id: 456,
      persona_nickname: body.persona_nickname,
      persona_school: body.persona_school,
      persona_description: body.persona_description,
      teacher_id: 789,
      student_count: 0,
      created_at: '2026-04-09T10:30:00Z',
      share_url: 'https://example.com/class/123',
      share_code: 'ABC123',
      token: 'mock-jwt-token-xyz',
    }

    return HttpResponse.json(
      {
        code: 0,
        message: 'success',
        data: newClass,
      },
      { status: 200 }
    )
  }),

  // 更新班级
  http.put('/api/classes/:id', async ({ params, request }) => {
    const authError = verifyAuth(request)
    if (authError) return authError

    const id = Number(params.id)
    const body = (await request.json()) as any

    // 模拟班级不存在
    if (id === 9999) {
      return HttpResponse.json(
        { code: 40017, message: '班级不存在' },
        { status: 404 }
      )
    }

    // 模拟无权限操作
    if (id === 8888) {
      return HttpResponse.json(
        { code: 40018, message: '无权操作此班级' },
        { status: 403 }
      )
    }

    // 模拟班级名称已存在
    if (body.name) {
      const existingClass = mockClasses.find((c) => c.name === body.name && c.id !== id)
      if (existingClass) {
        return HttpResponse.json(
          { code: 40016, message: '班级名称已存在' },
          { status: 409 }
        )
      }
    }

    // 验证学段类型
    if (body.curriculum_config?.grade_level) {
      if (!validGradeLevels.includes(body.curriculum_config.grade_level)) {
        return HttpResponse.json(
          { code: 40041, message: '无效的学段类型' },
          { status: 400 }
        )
      }
    }

    return HttpResponse.json({
      code: 0,
      message: 'success',
      data: {
        id,
        name: body.name || mockClasses[0].name,
        description: body.description !== undefined ? body.description : mockClasses[0].description,
        persona_id: 456,
      },
    })
  }),

  // 删除班级
  http.delete('/api/teacher/classes/:id', ({ params, request }) => {
    const authError = verifyAuth(request)
    if (authError) return authError

    const id = Number(params.id)

    // 模拟班级不存在
    if (id === 9999) {
      return HttpResponse.json(
        { code: 40017, message: '班级不存在' },
        { status: 404 }
      )
    }

    return HttpResponse.json({
      code: 0,
      message: 'success',
      data: null,
    })
  }),

  // 添加学生到班级
  http.post('/api/teacher/classes/:id/students', async ({ params, request }) => {
    const authError = verifyAuth(request)
    if (authError) return authError

    return HttpResponse.json({
      code: 0,
      message: 'success',
      data: null,
    })
  }),

  // 从班级移除学生
  http.delete('/api/teacher/classes/:classId/students/:studentId', ({ params, request }) => {
    const authError = verifyAuth(request)
    if (authError) return authError

    return HttpResponse.json({
      code: 0,
      message: 'success',
      data: null,
    })
  }),

  // 获取班级学生列表
  http.get('/api/teacher/classes/:id/students', ({ params, request }) => {
    const authError = verifyAuth(request)
    if (authError) return authError

    return HttpResponse.json({
      code: 0,
      message: 'success',
      data: [
        { id: 1, nickname: '学生A', status: 'active', last_active_at: '2026-04-09T10:30:00Z' },
        { id: 2, nickname: '学生B', status: 'inactive', last_active_at: '2026-04-08T14:20:00Z' },
      ],
    })
  }),
]

// 错误场景 handlers - 用于覆盖特定错误测试
export const errorHandlers = {
  // 服务器错误
  serverError: http.get('/api/classes', () => {
    return HttpResponse.json(
      { code: 50001, message: '数据库服务不可用' },
      { status: 500 }
    )
  }),

  // 401 错误
  unauthorized: http.get('/api/classes', () => {
    return HttpResponse.json(
      { code: 40101, message: '未登录或 token 已过期' },
      { status: 401 }
    )
  }),

  // 网络错误
  networkError: http.get('/api/classes', () => {
    return HttpResponse.error()
  }),
}
