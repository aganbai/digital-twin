import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

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

/** 错误码枚举 */
export enum ErrorCode {
  SUCCESS = 0,
  UNAUTHORIZED = 401,
  FORBIDDEN = 403,
  NOT_FOUND = 404,
  SERVER_ERROR = 500,
}

// 创建 axios 实例
const request: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const { data } = response

    // 业务错误处理
    if (data.code !== ErrorCode.SUCCESS) {
      ElMessage.error(data.message || '请求失败')
      return Promise.reject(new Error(data.message))
    }

    return response
  },
  (error) => {
    const { response } = error

    if (response) {
      switch (response.status) {
        case ErrorCode.UNAUTHORIZED:
          // Token 无效或过期，清除登录态并跳转登录页
          localStorage.removeItem('token')
          localStorage.removeItem('userInfo')
          ElMessage.error('登录已过期，请重新登录')
          router.push('/login')
          break
        case ErrorCode.FORBIDDEN:
          ElMessage.error(response.data?.message || '您没有权限执行此操作')
          break
        default:
          ElMessage.error(response.data?.message || '服务器错误')
      }
    } else {
      ElMessage.error('网络异常，请检查网络连接')
    }

    return Promise.reject(error)
  }
)

/**
 * GET 请求
 */
export function get<T = any>(url: string, params?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
  return request.get(url, { params, ...config }).then((res) => res.data)
}

/**
 * POST 请求
 */
export function post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
  return request.post(url, data, config).then((res) => res.data)
}

/**
 * PUT 请求
 */
export function put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
  return request.put(url, data, config).then((res) => res.data)
}

/**
 * DELETE 请求
 */
export function del<T = any>(url: string, params?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
  return request.delete(url, { params, ...config }).then((res) => res.data)
}

export default request
