import Taro from '@tarojs/taro'
import { request } from './request'
import { BASE_URL } from '../utils/constants'

/** 批量任务状态 */
export interface BatchTask {
  task_id: string
  status: 'pending' | 'processing' | 'success' | 'partial_success' | 'failed'
  total_files: number
  success_files: number
  failed_files: number
  result_json?: any
  created_at: string
}

/** 批量上传响应 */
export interface BatchUploadResponse {
  task_id: string
  status: string
  total_files: number
}

/**
 * 批量上传文件（逐个上传到批量接口）
 * 小程序环境下不支持真正的多文件 FormData，
 * 因此先逐个上传文件收集临时路径，再统一提交批量任务
 */
export function batchUploadDocuments(
  filePaths: string[],
  personaId: number,
  knowledgeBaseId?: number,
): Promise<BatchUploadResponse> {
  return new Promise((resolve, reject) => {
    const token = Taro.getStorageSync('token') || ''

    // 使用第一个文件发起上传，其余文件路径通过 formData 传递
    // 小程序 uploadFile 只支持单文件，这里采用逐个上传方案：
    // 将所有文件逐个上传到同一个 batch-upload 接口
    // 后端通过 persona_id 关联同一批次
    const formData: Record<string, string> = {
      persona_id: String(personaId),
      file_count: String(filePaths.length),
    }
    if (knowledgeBaseId) {
      formData['knowledge_base_id'] = String(knowledgeBaseId)
    }

    // 逐个上传文件，收集结果
    let uploadedCount = 0
    let taskId = ''

    const uploadNext = (index: number) => {
      if (index >= filePaths.length) {
        // 全部上传完成
        resolve({
          task_id: taskId,
          status: 'pending',
          total_files: filePaths.length,
        })
        return
      }

      Taro.uploadFile({
        url: `${BASE_URL}/api/documents/batch-upload`,
        filePath: filePaths[index],
        name: 'file',
        formData: {
          ...formData,
          file_index: String(index),
          // 首个文件创建任务，后续文件追加到同一任务
          ...(taskId ? { task_id: taskId } : {}),
        },
        header: {
          Authorization: `Bearer ${token}`,
        },
        success: (res) => {
          if (res.statusCode === 200 || res.statusCode === 202) {
            const data = JSON.parse(res.data)
            if (data.code === 0 || data.task_id) {
              const respData = data.data || data
              if (!taskId && respData.task_id) {
                taskId = respData.task_id
              }
              uploadedCount++
              uploadNext(index + 1)
            } else {
              reject(new Error(data.message || '批量上传失败'))
            }
          } else {
            reject(new Error(`上传失败: HTTP ${res.statusCode}`))
          }
        },
        fail: (err) => {
          reject(new Error(err.errMsg || '批量上传失败'))
        },
      })
    }

    uploadNext(0)
  })
}

/** 查询批量任务状态 */
export function getBatchTaskStatus(taskId: string) {
  return request<BatchTask>({
    url: `/api/batch-tasks/${taskId}`,
    method: 'GET',
  })
}
