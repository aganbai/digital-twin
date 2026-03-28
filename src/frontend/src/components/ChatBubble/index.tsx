import { View, Text } from '@tarojs/components'
import './index.scss'

interface ChatBubbleProps {
  /** 消息角色 */
  role: 'user' | 'assistant'
  /** 消息内容 */
  content: string
  /** 时间戳（格式化后的字符串） */
  timestamp?: string
  /** 教师昵称（用于生成头像首字母） */
  teacherName?: string
  /** 是否发送失败 */
  failed?: boolean
  /** 重试回调 */
  onRetry?: () => void
}

export default function ChatBubble({
  role,
  content,
  timestamp,
  teacherName,
  failed,
  onRetry,
}: ChatBubbleProps) {
  /** 获取教师昵称的首字母（用于头像占位） */
  const getInitial = (name?: string): string => {
    if (!name) return 'T'
    return name.charAt(0).toUpperCase()
  }

  return (
    <View className='chat-bubble-wrapper'>
      {/* 时间戳 */}
      {timestamp && (
        <View className='chat-bubble__timestamp'>
          <Text className='chat-bubble__timestamp-text'>{timestamp}</Text>
        </View>
      )}

      {/* 消息行 */}
      <View className={`chat-bubble chat-bubble--${role}`}>
        {/* AI 回复：左侧头像 */}
        {role === 'assistant' && (
          <View className='chat-bubble__avatar'>
            <Text className='chat-bubble__avatar-text'>
              {getInitial(teacherName)}
            </Text>
          </View>
        )}

        {/* 气泡内容 */}
        <View className={`chat-bubble__content chat-bubble__content--${role}`}>
          <Text className={`chat-bubble__text chat-bubble__text--${role}`}>
            {content}
          </Text>
        </View>
      </View>

      {/* 发送失败提示 */}
      {failed && (
        <View className='chat-bubble__failed' onClick={onRetry}>
          <Text className='chat-bubble__failed-text'>发送失败，点击重试</Text>
        </View>
      )}
    </View>
  )
}
