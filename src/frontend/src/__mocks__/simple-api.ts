/**
 * 简化的 API Mock 方案
 * 不使用 MSW，直接模拟 request 函数
 */

import type { ApiResponse } from '../api/request'

// Mock 响应存储
const mockResponses = new Map<string, any>()
const mockErrors = new Map<string, Error>()

/**
 * 设置 Mock 响应
 */
export function setMockResponse(urlPattern: string | RegExp, response: ApiResponse<any>) {
  mockResponses.set(urlPattern.toString(), response)
}

/**
 * 设置 Mock 错误
 */
export function setMockError(urlPattern: string | RegExp, error: Error) {
  mockErrors.set(urlPattern.toString(), error)
}

/**
 * 清除所有 Mock
 */
export function clearMocks() {
  mockResponses.clear()
  mockErrors.clear()
}

/**
 * 模拟 request 函数
 */
export function mockRequest<T>(options: {
  url: string
  method: string
  data?: any
}): Promise<ApiResponse<T>> {
  const key = options.url

  // 检查是否有预设的错误
  for (const [pattern, error] of mockErrors.entries()) {
    if (key.includes(pattern.replace(/\//g, '').replace(/api/, 'api'))) {
      return Promise.reject(error)
    }
  }

  // 检查是否有预设的响应
  for (const [pattern, response] of mockResponses.entries()) {
    if (key.includes(pattern.replace(/\//g, '').replace(/\?.*$/, ''))) {
      return Promise.resolve(response as ApiResponse<T>)
    }
  }

  // 默认成功响应
  return Promise.resolve({
    code: 0,
    message: 'success',
    data: {} as T,
  })
}

// 导出的工具函数
export function createSuccessResponse<T>(data: T): ApiResponse<T> {
  return {
    code: 0,
    message: 'success',
    data,
  }
}

export function createErrorResponse<T = null>(
  code: number,
  message: string
): ApiResponse<T> {
  return {
    code,
    message,
    data: null as T,
  }
}

export { mockResponses, mockErrors }
