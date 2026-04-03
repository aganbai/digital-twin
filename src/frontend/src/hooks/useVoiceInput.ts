import { useState, useRef, useCallback } from 'react'
import Taro from '@tarojs/taro'

/** 语音输入状态 */
export type VoiceState = 'idle' | 'recording' | 'recognizing'

/** 语音输入 Hook 返回值 */
export interface UseVoiceInputReturn {
  /** 当前状态 */
  voiceState: VoiceState
  /** 识别结果文本 */
  recognizedText: string
  /** 录音时长（秒） */
  duration: number
  /** 开始录音 */
  startRecording: () => void
  /** 停止录音 */
  stopRecording: () => void
  /** 取消录音 */
  cancelRecording: () => void
}

/**
 * 语音输入 Hook
 * 语音功能暂停使用，待认证完成后开放
 * 当前直接返回不可用状态
 */
export function useVoiceInput(): UseVoiceInputReturn {
  const [voiceState] = useState<VoiceState>('idle')
  const [recognizedText] = useState('')
  const [duration] = useState(0)

  /** 停止录音 - 空实现 */
  const stopRecording = useCallback(() => {
    // 语音功能暂停使用
  }, [])

  /** 开始录音 */
  const startRecording = useCallback(() => {
    // 语音功能暂停使用，提示用户
    Taro.showToast({ title: '语音功能暂未开放', icon: 'none' })
  }, [])

  /** 取消录音 */
  const cancelRecording = useCallback(() => {
    // 语音功能暂停使用
  }, [])

  return {
    voiceState,
    recognizedText,
    duration,
    startRecording,
    stopRecording,
    cancelRecording,
  }
}
