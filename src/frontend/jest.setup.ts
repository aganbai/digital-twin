import '@testing-library/jest-dom'
import { TextEncoder, TextDecoder } from 'util'

// Polyfill for browser APIs
if (typeof global.TextEncoder === 'undefined') {
  global.TextEncoder = TextEncoder
}
if (typeof global.TextDecoder === 'undefined') {
  global.TextDecoder = TextDecoder as any
}

// Polyfill for fetch APIs used by MSW
if (typeof global.Response === 'undefined') {
  global.Response = class Response {
    constructor(body: any, init?: any) {}
  } as any
}
if (typeof global.Request === 'undefined') {
  global.Request = class Request {
    constructor(input: any, init?: any) {}
  } as any
}
if (typeof global.Headers === 'undefined') {
  global.Headers = class Headers {
    constructor(init?: any) {}
    get(name: string) { return null }
    set(name: string, value: string) {}
  } as any
}

// Mock Taro
jest.mock('@tarojs/taro', () => {
  const mockTaro = {
    request: jest.fn(),
    getStorageSync: jest.fn(),
    setStorageSync: jest.fn(),
    removeStorageSync: jest.fn(),
    showToast: jest.fn(),
    showModal: jest.fn(),
    redirectTo: jest.fn(),
    navigateTo: jest.fn(),
    navigateBack: jest.fn(),
    useRouter: jest.fn(() => ({ params: {} })),
    getCurrentInstance: jest.fn(() => ({
      router: { params: { class_id: '123' } },
    })),
    switchTab: jest.fn(),
    setClipboardData: jest.fn(),
    useDidShow: jest.fn(() => {}),
    useDidHide: jest.fn(() => {}),
    useReady: jest.fn(() => {}),
    useLoad: jest.fn(() => {}),
    useUnload: jest.fn(() => {}),
  }
  return {
    __esModule: true,
    default: mockTaro,
    ...mockTaro,
  }
})

// Suppress console.log during tests (keep console.error and console.warn)
const originalConsoleLog = console.log
console.log = (...args: any[]) => {
  if (typeof args[0] === 'string' && args[0].includes('[Profile]')) {
    return
  }
  originalConsoleLog(...args)
}

// Reset all mocks after each test
afterEach(() => {
  jest.clearAllMocks()
})
