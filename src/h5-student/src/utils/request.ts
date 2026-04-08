import axios, { type AxiosRequestConfig, type AxiosResponse } from 'axios'
import { showToast } from 'vant'
import { getToken, clearAuthInfo } from './auth'

// 创建 axios 实例
const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse) => response.data,
  (error) => {
    const status = error.response?.status
    if (status === 401) {
      clearAuthInfo()
      window.location.href = '/login'
      return Promise.reject(error)
    }
    if (status === 403) {
      showToast('没有权限访问')
      return Promise.reject(error)
    }
    const message = error.response?.data?.message || error.message || '请求失败'
    showToast(message)
    return Promise.reject(error)
  }
)

// 封装请求方法
export interface PaginatedData<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export function get<T>(url: string, params?: object, config?: AxiosRequestConfig): Promise<{ data: T }> {
  return request.get(url, { params, ...config })
}

export function post<T>(url: string, data?: object, config?: AxiosRequestConfig): Promise<{ data: T }> {
  return request.post(url, data, config)
}

export function put<T>(url: string, data?: object, config?: AxiosRequestConfig): Promise<{ data: T }> {
  return request.put(url, data, config)
}

export function del<T>(url: string, config?: AxiosRequestConfig): Promise<{ data: T }> {
  return request.delete(url, config)
}

export default request
