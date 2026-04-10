/**
 * Taro 框架 Mock
 * 用于单元测试中模拟小程序 API
 */

// Mock storage
const mockStorage: Record<string, any> = {}

const Taro = {
  // Storage API
  getStorageSync: jest.fn((key: string) => mockStorage[key] || ''),
  setStorageSync: jest.fn((key: string, data: any) => {
    mockStorage[key] = data
  }),
  removeStorageSync: jest.fn((key: string) => {
    delete mockStorage[key]
  }),

  // Navigation API
  redirectTo: jest.fn(({ url }: { url: string }) => Promise.resolve({ errMsg: 'redirectTo:ok' })),
  navigateTo: jest.fn(({ url }: { url: string }) => Promise.resolve({ errMsg: 'navigateTo:ok' })),
  navigateBack: jest.fn(() => Promise.resolve({ errMsg: 'navigateBack:ok' })),

  // UI API
  showToast: jest.fn(() => Promise.resolve()),
  showModal: jest.fn(() => Promise.resolve({ confirm: false, cancel: true })),
  showLoading: jest.fn(() => Promise.resolve()),
  hideLoading: jest.fn(() => Promise.resolve()),

  // Network API
  request: jest.fn(),

  // Lifecycle hooks (for components)
  useDidShow: jest.fn((callback: () => void) => {
    // Call immediately for testing
    callback()
  }),
  useDidHide: jest.fn(),
  useReady: jest.fn(),
  useLoad: jest.fn((callback: () => void) => {
    callback()
  }),
  useUnload: jest.fn(),

  // Other utilities
  getSystemInfoSync: jest.fn(() => ({
    windowWidth: 375,
    windowHeight: 667,
    screenWidth: 375,
    screenHeight: 667,
    pixelRatio: 2,
    statusBarHeight: 20,
  })),

  // Clear mock storage helper
  __clearStorage: () => {
    Object.keys(mockStorage).forEach((key) => delete mockStorage[key])
  },
}

// Named exports for specific hooks
export const useDidShow = Taro.useDidShow
export const useDidHide = Taro.useDidHide
export const useReady = Taro.useReady
export const useLoad = Taro.useLoad
export const useUnload = Taro.useUnload
export const getSystemInfoSync = Taro.getSystemInfoSync

export default Taro
