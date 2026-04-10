/**
 * MSW (Mock Service Worker) Server Configuration
 * 用于模拟 API 请求
 */

import { setupServer } from 'msw/node'
import { http, HttpResponse } from 'msw'

// API 基础 URL
const BASE_URL = 'http://localhost:8080'

// Mock 数据定义
export const mockUserProfile = {
  id: 1,
  username: 'test_user',
  nickname: '测试用户',
  role: 'teacher',
  email: 'test@example.com',
  created_at: '2026-01-01T00:00:00Z',
  stats: {
    conversation_count: 10,
    memory_count: 5,
    document_count: 20,
  },
}

export const mockStudentProfile = {
  id: 2,
  username: 'student_user',
  nickname: '学生用户',
  role: 'student',
  email: 'student@example.com',
  created_at: '2026-01-01T00:00:00Z',
  stats: {
    conversation_count: 15,
    memory_count: 8,
  },
}

// Mock 班级数据
export const mockClassDetail = {
  id: 123,
  name: '三年级数学班',
  description: '小学数学培优班级',
  is_public: true,
  persona_id: 456,
  persona_nickname: '王老师',
  persona_school: '实验小学',
  persona_description: '10年数学教学经验',
  student_count: 30,
  created_at: '2026-04-09T10:30:00Z',
  curriculum_config: null,
}

export const mockClassWithCurriculum = {
  id: 124,
  name: '三年级数学班（有配置）',
  description: '已配置教材的班级',
  is_public: true,
  persona_id: 457,
  persona_nickname: '李老师',
  persona_school: '实验中学',
  persona_description: '数学特级教师',
  student_count: 25,
  created_at: '2026-04-09T10:30:00Z',
  curriculum_config: {
    id: 1001,
    grade_level: 'primary_lower',
    grade: '三年级',
    subjects: ['数学', '语文'],
    textbook_versions: ['人教版', '北师大版'],
    custom_textbooks: ['《小学奥数启蒙》'],
    current_progress: '第三单元 乘法初步',
  },
}

export const mockCreateClassV11Response = {
  id: 123,
  name: '三年级数学班',
  description: '小学数学培优班级',
  is_public: true,
  persona_id: 456,
  persona_nickname: '王老师',
  persona_school: '实验小学',
  persona_description: '10年数学教学经验，专注小学奥数',
  share_code: 'ABC123',
  share_url: 'http://example.com/share/ABC123',
  created_at: '2026-04-09T10:30:00Z',
}

// 定义 handlers
export const handlers = [
  // 获取用户详情 - 成功响应
  http.get(`${BASE_URL}/api/user/profile`, ({ request }) => {
    const authHeader = request.headers.get('Authorization')

    // 模拟 401 未授权
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return HttpResponse.json(
        {
          code: 40101,
          message: '未登录或登录已过期',
          data: null,
        },
        { status: 401 }
      )
    }

    const token = authHeader.replace('Bearer ', '')

    // 模拟学生角色
    if (token.includes('student')) {
      return HttpResponse.json({
        code: 0,
        message: 'success',
        data: mockStudentProfile,
      })
    }

    // 默认返回教师角色
    return HttpResponse.json({
      code: 0,
      message: 'success',
      data: mockUserProfile,
    })
  }),

  // 获取用户详情 - 服务器错误
  http.get(`${BASE_URL}/api/user/profile-error`, () => {
    return HttpResponse.json(
      {
        code: 50001,
        message: '服务器内部错误',
        data: null,
      },
      { status: 500 }
    )
  }),

  // 获取用户详情 - 网络错误模拟
  http.get(`${BASE_URL}/api/user/profile-network-error`, () => {
    return new Response(null, { status: 0 })
  }),

  // ===== 班级相关 API Mock =====

  // 创建班级 V11 - 成功响应
  http.post(`${BASE_URL}/api/classes`, async ({ request }) => {
    const body = (await request.json()) as any

    // 验证必填字段
    if (!body.name || !body.persona_nickname || !body.persona_school || !body.persona_description) {
      return HttpResponse.json(
        {
          code: 40004,
          message: '请求参数无效',
          data: null,
        },
        { status: 400 }
      )
    }

    // 模拟班级名称重复
    if (body.name === '重复班级名') {
      return HttpResponse.json(
        {
          code: 40030,
          message: '该班级名称已存在',
          data: null,
        },
        { status: 409 }
      )
    }

    return HttpResponse.json({
      code: 0,
      message: 'success',
      data: {
        ...mockCreateClassV11Response,
        name: body.name,
        description: body.description || '',
        persona_nickname: body.persona_nickname,
        persona_school: body.persona_school,
        persona_description: body.persona_description,
        // 如传递了教材配置，在响应中包含
        ...(body.curriculum_config && {
          curriculum_config: body.curriculum_config,
        }),
      },
    })
  }),

  // 获取班级详情 - 成功响应（无教材配置）
  http.get(`${BASE_URL}/api/classes/:id`, ({ params }) => {
    const id = params.id as string

    // 模拟班级不存在
    if (id === '999') {
      return HttpResponse.json(
        {
          code: 40017,
          message: '班级不存在',
          data: null,
        },
        { status: 404 }
      )
    }

    // 返回有教材配置的班级
    if (id === '124') {
      return HttpResponse.json({
        code: 0,
        message: 'success',
        data: mockClassWithCurriculum,
      })
    }

    return HttpResponse.json({
      code: 0,
      message: 'success',
      data: {
        ...mockClassDetail,
        id: Number(id),
      },
    })
  }),

  // 更新班级信息 - 成功响应
  http.put(`${BASE_URL}/api/classes/:id`, async ({ params, request }) => {
    const id = params.id as string
    const body = (await request.json()) as any

    // 模拟班级不存在
    if (id === '999') {
      return HttpResponse.json(
        {
          code: 40017,
          message: '班级不存在',
          data: null,
        },
        { status: 404 }
      )
    }

    // 模拟无权操作
    if (id === '888') {
      return HttpResponse.json(
        {
          code: 40018,
          message: '无权操作此班级',
          data: null,
        },
        { status: 403 }
      )
    }

    // 模拟班级名称重复
    if (body.name === '重复班级名') {
      return HttpResponse.json(
        {
          code: 40016,
          message: '班级名称已存在',
          data: null,
        },
        { status: 409 }
      )
    }

    return HttpResponse.json({
      code: 0,
      message: 'success',
      data: {
        id: Number(id),
        name: body.name || mockClassDetail.name,
        description: body.description !== undefined ? body.description : mockClassDetail.description,
        persona_id: mockClassDetail.persona_id,
        // 如传递了教材配置，在响应中包含
        ...(body.curriculum_config && {
          curriculum_config: body.curriculum_config,
        }),
      },
    })
  }),

  // 获取班级详情 - 服务器错误
  http.get(`${BASE_URL}/api/classes/:id/server-error`, () => {
    return HttpResponse.json(
      {
        code: 50001,
        message: '数据库服务不可用',
        data: null,
      },
      { status: 500 }
    )
  }),

  // 获取班级列表
  http.get(`${BASE_URL}/api/classes`, () => {
    return HttpResponse.json({
      code: 0,
      message: 'success',
      data: [mockClassDetail, mockClassWithCurriculum],
    })
  }),

  // 删除班级
  http.delete(`${BASE_URL}/api/classes/:id`, ({ params }) => {
    const id = params.id as string

    if (id === '999') {
      return HttpResponse.json(
        {
          code: 40017,
          message: '班级不存在',
          data: null,
        },
        { status: 404 }
      )
    }

    return HttpResponse.json({
      code: 0,
      message: 'success',
      data: { message: '班级已删除' },
    })
  }),
]

// 创建 server 实例
export const server = setupServer(...handlers)

// 导出 Mock 数据供测试使用（mockStudentProfile 已在27行定义导出，此处无需重复导出）
