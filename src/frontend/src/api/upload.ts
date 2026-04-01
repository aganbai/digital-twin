import Taro from '@tarojs/taro'
import { BASE_URL } from '../utils/constants'

/** 上传文件响应 */
export interface UploadResponse {
  url: string
  filename: string
  original_name: string
  size: number
  mime_type: string
}

/**
 * 通用文件上传
 * @param filePath - 本地文件路径
 * @param type - 用途标识：assignment / document / general
 */
export function uploadFile(filePath: string, type = 'general'): Promise<{ data: UploadResponse }> {
  return new Promise((resolve, reject) => {
    const token = Taro.getStorageSync('token') || ''
    Taro.uploadFile({
      url: `${BASE_URL}/api/upload`,
      filePath,
      name: 'file',
      formData: { type },
      header: {
        Authorization: `Bearer ${token}`,
      },
      success: (res) => {
        if (res.statusCode === 200) {
          const data = JSON.parse(res.data)
          if (data.code === 0) {
            resolve({ data: data.data })
          } else {
            reject(new Error(data.message || '上传失败'))
          }
        } else {
          reject(new Error(`上传失败: HTTP ${res.statusCode}`))
        }
      },
      fail: (err) => {
        reject(new Error(err.errMsg || '上传失败'))
      },
    })
  })
}
