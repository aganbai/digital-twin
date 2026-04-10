import { beforeAll, afterAll, afterEach, vi } from 'vitest'
import { server } from './mocks/server'

// Mock Element Plus
vi.mock('element-plus', async () => {
  const actual = await vi.importActual<typeof import('element-plus')>('element-plus')
  return {
    ...actual,
    ElMessage: {
      success: vi.fn(),
      error: vi.fn(),
      warning: vi.fn(),
      info: vi.fn(),
    },
    ElMessageBox: {
      confirm: vi.fn(),
      alert: vi.fn(),
      prompt: vi.fn(),
    },
    ElLoading: {
      service: vi.fn(() => ({ close: vi.fn() })),
    },
  }
})

// Mock window.location for request.ts
Object.defineProperty(window, 'location', {
  writable: true,
  value: {
    href: 'http://localhost/',
    pathname: '/',
  },
})

// Mock localStorage for auth token
Object.defineProperty(window, 'localStorage', {
  value: {
    getItem: vi.fn((key: string) => {
      if (key === 'h5_teacher_token') return 'mock-token'
      return null
    }),
    setItem: vi.fn(),
    removeItem: vi.fn(),
  },
  writable: true,
})

// 建立 MSW 监听
beforeAll(() => server.listen({ onUnhandledRequest: 'bypass' }))

// 每个测试后重置 handlers
afterEach(() => {
  server.resetHandlers()
  vi.clearAllMocks()
})

// 测试完成后关闭
afterAll(() => server.close())
