import { useState, useCallback, useRef } from 'react'
import Taro from '@tarojs/taro'
import { View, Text } from '@tarojs/components'
import './index.scss'

/** 语音输入状态 */
export type VoiceState = 'idle' | 'recording' | 'recognizing'

interface VoiceInputProps {
  /** 语音状态变化回调 */
  onStateChange?: (state: VoiceState) => void
  /** 识别结果回调 */
  onResult?: (text: string) => void
  /** 取消回调 */
  onCancel?: () => void
}

/**
 * 语音输入组件
 * 提供按住说话的交互方式
 */
export default function VoiceInput({
  onStateChange,
  onResult,
  onCancel,
}: VoiceInputProps) {
  const [voiceState, setVoiceState] = useState<VoiceState>('idle')
  const [duration, setDuration] = useState(0)
  const recorderManager = useRef<any>(null)
  const timerRef = useRef<any>(null)

  /** 初始化录音管理器 */
  const initRecorder = useCallback(() => {
    if (!recorderManager.current) {
      recorderManager.current = Taro.getRecorderManager()
      
      recorderManager.current.onStop((res: any) => {
        setVoiceState('recognizing')
        onStateChange?.('recognizing')
        
        // 调用微信语音识别API
        // 注意：需要在小程序后台开通语音识别权限
        Taro.showLoading({ title: '识别中...' })
        
        // 这里使用微信同声传译插件或第三方服务
        // 为简化示例，这里模拟识别过程
        setTimeout(() => {
          Taro.hideLoading()
          // 模拟识别结果（实际应调用语音识别API）
          const mockText = '这是语音识别的测试文本'
          setVoiceState('idle')
          onStateChange?.('idle')
          onResult?.(mockText)
        }, 1000)
      })
      
      recorderManager.current.onError((err: any) => {
        console.error('录音失败:', err)
        Taro.hideLoading()
        Taro.showToast({ title: '录音失败', icon: 'none' })
        setVoiceState('idle')
        onStateChange?.('idle')
      })
    }
    return recorderManager.current
  }, [onStateChange, onResult])

  /** 开始录音 */
  const handleTouchStart = useCallback((e: any) => {
    e.preventDefault()
    
    // 检查录音权限
    Taro.authorize({
      scope: 'scope.record',
      success: () => {
        const manager = initRecorder()
        
        manager.start({
          format: 'mp3',
          sampleRate: 16000,
          numberOfChannels: 1,
          encodeBitRate: 48000,
        })
        
        setVoiceState('recording')
        onStateChange?.('recording')
        
        // 开始计时
        let seconds = 0
        timerRef.current = setInterval(() => {
          seconds++
          setDuration(seconds)
          
          // 最长60秒
          if (seconds >= 60) {
            handleTouchEnd()
          }
        }, 1000)
      },
      fail: () => {
        Taro.showModal({
          title: '权限提示',
          content: '需要录音权限才能使用语音输入功能',
          success: (res) => {
            if (res.confirm) {
              Taro.openSetting()
            }
          },
        })
      },
    })
  }, [initRecorder, onStateChange])

  /** 结束录音 */
  const handleTouchEnd = useCallback(() => {
    if (voiceState !== 'recording') return
    
    // 停止计时
    if (timerRef.current) {
      clearInterval(timerRef.current)
      timerRef.current = null
    }
    
    // 停止录音
    const manager = recorderManager.current
    if (manager && duration >= 1) {
      manager.stop()
    } else {
      // 录音时间太短
      if (manager) {
        manager.stop()
      }
      setVoiceState('idle')
      onStateChange?.('idle')
      setDuration(0)
      Taro.showToast({ title: '录音时间太短', icon: 'none' })
    }
  }, [voiceState, duration, onStateChange])

  /** 取消录音 */
  const handleTouchCancel = useCallback(() => {
    // 停止计时
    if (timerRef.current) {
      clearInterval(timerRef.current)
      timerRef.current = null
    }
    
    // 停止录音
    const manager = recorderManager.current
    if (manager) {
      manager.stop()
    }
    
    setVoiceState('idle')
    setDuration(0)
    onStateChange?.('idle')
    onCancel?.()
  }, [onStateChange, onCancel])

  return (
    <View className='voice-input'>
      {voiceState === 'idle' ? (
        <View className='voice-input__idle'>
          <Text className='voice-input__hint'>点击语音图标开始说话</Text>
        </View>
      ) : voiceState === 'recording' ? (
        <View className='voice-input__recording'>
          <View className='voice-input__wave'>
            <View className='voice-input__wave-bar' />
            <View className='voice-input__wave-bar' />
            <View className='voice-input__wave-bar' />
            <View className='voice-input__wave-bar' />
            <View className='voice-input__wave-bar' />
          </View>
          <Text className='voice-input__time'>{duration}s</Text>
          <Text className='voice-input__action'>松开发送，上滑取消</Text>
        </View>
      ) : (
        <View className='voice-input__recognizing'>
          <Text className='voice-input__loading'>识别中...</Text>
        </View>
      )}
    </View>
  )
}

/** 按住说话按钮组件 */
export function VoiceButton({
  onPress,
  onRelease,
  onCancel,
}: {
  onPress?: () => void
  onRelease?: () => void
  onCancel?: () => void
}) {
  return (
    <View
      className='voice-button'
      onTouchStart={(e) => {
        e.preventDefault()
        onPress?.()
      }}
      onTouchEnd={(e) => {
        e.preventDefault()
        onRelease?.()
      }}
      onTouchCancel={(e) => {
        e.preventDefault()
        onCancel?.()
      }}
      onLongPress={(e) => {
        e.preventDefault()
      }}
    >
      <Text className='voice-button__text'>按住说话</Text>
    </View>
  )
}
