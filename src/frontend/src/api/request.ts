import Taro from '@tarojs/taro'
import { getToken, removeToken } from '../utils/storage'
import { removeUserInfo } from '../utils/storage'
import { BASE_URL, ERROR_CODES } from '../utils/constants'

/** 请求配置选项 */
interface RequestOptions {
  url: string
  method: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: any
  header?: Record<string, string>
}

/** 统一响应格式 */
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

/** 分页响应数据 */
export interface PaginatedData<T = any> {
  items: T[]
  total: number
  page: number
  page_size: number
}

/**
 * 统一网络请求封装
 * - 自动附加 Authorization: Bearer token
 * - 统一错误处理：401 → 清除 token → 跳转登录页
 * - 网络错误 → Toast 提示
 * - 业务错误 → Toast 提示 message
 */
export async function request<T = any>(options: RequestOptions): Promise<ApiResponse<T>> {
  const token = getToken()

  const header: Record<string, string> = {
    'Content-Type': 'application/json',
    ...options.header,
  }

  if (token) {
    header['Authorization'] = `Bearer ${token}`
  }

  try {
    const response = await Taro.request({
      url: `${BASE_URL}${options.url}`,
      method: options.method,
      data: options.data,
      header,
      timeout: 30000, // 30秒超时
    })

    const result = response.data as ApiResponse<T>

    // Token 无效或过期 → 清除登录态 → 跳转登录页
    if (result.code === ERROR_CODES.UNAUTHORIZED || result.code === ERROR_CODES.TOKEN_EXPIRED) {
      removeToken()
      removeUserInfo()
      Taro.redirectTo({ url: '/pages/login/index' })
      throw new Error('登录已过期，请重新登录')
    }

    // 业务错误 → Toast 提示
    if (result.code !== ERROR_CODES.SUCCESS) {
      Taro.showToast({ title: result.message || '请求失败', icon: 'none' })
      throw new Error(result.message)
    }

    return result
  } catch (error) {
    // 网络异常（Taro.request 本身抛出的错误）
    if (error instanceof Error) {
      console.error('[Request] 请求失败:', options.url, error.message)
      if (error.message.includes('timeout')) {
        Taro.showToast({ title: '请求超时，请稍后重试', icon: 'none' })
      } else if (error.message.includes('request:fail')) {
        Taro.showToast({ title: '网络异常，请检查网络连接', icon: 'none' })
      }
    }
    throw error
  }
}
