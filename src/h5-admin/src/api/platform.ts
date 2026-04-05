import { get } from '@/utils/request'

/** 平台配置 */
export interface PlatformConfig {
  platform: string
  voice_input: boolean
  wx_subscribe: boolean
  file_upload: boolean
  share_to_wx: boolean
  camera: boolean
  max_file_size: number
  allowed_file_types: string[]
}

/**
 * 获取平台配置
 * @param platform - 平台标识（h5）
 */
export function getPlatformConfig(platform: string = 'h5') {
  return get<PlatformConfig>('/api/platform/config', { platform })
}
