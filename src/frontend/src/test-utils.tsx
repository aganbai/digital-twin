import { http, HttpResponse } from 'msw'
import { server } from './__mocks__/server'
import Taro from '@tarojs/taro'
import { CURRICULUM_ERRORS, CURRICULUM_CONSTANTS } from './__tests__/fixtures/curriculum-fixtures'
export { server, http, HttpResponse }

/**
 * 设置 API 响应模拟
 * @param urlPattern URL 匹配模式
 * @param responseData 响应数据
 * @param statusCode HTTP 状态码，默认 200
 */
export function mockApiResponse(
  urlPattern: string,
  responseData: { code: number; message: string; data?: any },
  statusCode: number = 200
) {
  server.use(
    http.get(urlPattern, () => {
      return HttpResponse.json(responseData, { status: statusCode })
    }),
    http.post(urlPattern, () => {
      return HttpResponse.json(responseData, { status: statusCode })
    }),
    http.put(urlPattern, () => {
      return HttpResponse.json(responseData, { status: statusCode })
    }),
    http.delete(urlPattern, () => {
      return HttpResponse.json(responseData, { status: statusCode })
    })
  )
}

/**
 * 设置网络错误模拟
 * @param urlPattern URL 匹配模式
 */
export function mockNetworkError(urlPattern: string) {
  server.use(
    http.get(urlPattern, () => {
      return new Response(null, { status: 0 })
    }),
    http.post(urlPattern, () => {
      return new Response(null, { status: 0 })
    }),
    http.put(urlPattern, () => {
      return new Response(null, { status: 0 })
    }),
    http.delete(urlPattern, () => {
      return new Response(null, { status: 0 })
    })
  )
}

/**
 * 设置超时错误模拟
 * @param urlPattern URL 匹配模式
 * @param delayMs 延迟时间（毫秒）
 */
export function mockTimeoutError(urlPattern: string, delayMs: number = 31000) {
  server.use(
    http.get(urlPattern, async () => {
      await new Promise((resolve) => setTimeout(resolve, delayMs))
      return HttpResponse.json({ code: -1, message: 'timeout' }, { status: 408 })
    }),
    http.post(urlPattern, async () => {
      await new Promise((resolve) => setTimeout(resolve, delayMs))
      return HttpResponse.json({ code: -1, message: 'timeout' }, { status: 408 })
    }),
    http.put(urlPattern, async () => {
      await new Promise((resolve) => setTimeout(resolve, delayMs))
      return HttpResponse.json({ code: -1, message: 'timeout' }, { status: 408 })
    }),
    http.delete(urlPattern, async () => {
      await new Promise((resolve) => setTimeout(resolve, delayMs))
      return HttpResponse.json({ code: -1, message: 'timeout' }, { status: 408 })
    })
  )
}

/**
 * 清空所有 Taro Mock 调用记录
 */
export function clearTaroMocks() {
  ;(Taro.showToast as jest.Mock).mockClear()
  ;(Taro.showModal as jest.Mock).mockClear()
  ;(Taro.navigateBack as jest.Mock).mockClear()
  ;(Taro.redirectTo as jest.Mock).mockClear()
  ;(Taro.navigateTo as jest.Mock).mockClear()
  ;(Taro.setClipboardData as jest.Mock).mockClear()
  ;(Taro.getCurrentInstance as jest.Mock)?.mockClear()
  ;(Taro.request as jest.Mock)?.mockClear()
}

/**
 * 获取最后一次 showToast 调用参数
 */
export function getLastToastCall() {
  const calls = (Taro.showToast as jest.Mock).mock.calls
  return calls[calls.length - 1]?.[0] || null
}

/**
 * 等待指定时间
 * @param ms 毫秒数
 */
export function wait(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

/**
 * 等待 React 状态更新完成
 */
export function waitForStateUpdate(): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, 0))
}

export { CURRICULUM_ERRORS, CURRICULUM_CONSTANTS }
