import { useState, useEffect, useCallback, useRef } from 'react'
import { View, Text, ScrollView, Input } from '@tarojs/components'
import Taro, { useRouter, useDidShow, useUnload } from '@tarojs/taro'
import { sendMessage, getConversations } from '@/api/chat'
import { useChatStore } from '@/store'
import { formatTime } from '@/utils/format'
import ChatBubble from '@/components/ChatBubble'
import './index.scss'

/** 判断两条消息之间是否需要显示时间戳（间隔超过 5 分钟） */
function shouldShowTimestamp(prev?: string, curr?: string): boolean {
  if (!prev || !curr) return true
  const prevTime = new Date(prev).getTime()
  const currTime = new Date(curr).getTime()
  if (isNaN(prevTime) || isNaN(currTime)) return true
  return currTime - prevTime > 5 * 60 * 1000
}

export default function Chat() {
  const router = useRouter()
  const teacherId = Number(router.params.teacher_id) || 0
  const teacherName = decodeURIComponent(router.params.teacher_name || '教师')

  const {
    messages,
    loading,
    sessionId,
    addMessage,
    setLoading,
    setSessionId,
    setMessages,
    clearMessages,
  } = useChatStore()

  /** 输入框内容 */
  const [inputValue, setInputValue] = useState('')
  /** 用于 ScrollView 自动滚动的锚点 ID */
  const [scrollIntoId, setScrollIntoId] = useState('')
  /** 失败消息的索引集合 */
  const [failedIndexes, setFailedIndexes] = useState<Set<number>>(new Set())
  /** 是否已初始化加载 */
  const initialized = useRef(false)

  /** 设置导航栏标题 */
  useDidShow(() => {
    Taro.setNavigationBarTitle({ title: teacherName })
  })

  /** 加载历史消息 */
  const loadHistory = useCallback(async () => {
    try {
      const res = await getConversations({
        teacher_id: teacherId,
        page: 1,
        page_size: 50,
      })
      const items = res.data.items || []
      // 接口返回按时间倒序，需要反转为正序
      const sorted = [...items].reverse()
      const historyMessages = sorted.map((item) => ({
        id: item.id,
        role: item.role as 'user' | 'assistant',
        content: item.content,
        created_at: item.created_at,
      }))
      setMessages(historyMessages)

      // 从历史记录中获取最新的 session_id
      if (sorted.length > 0) {
        const lastItem = items[0] // 倒序中第一条是最新的
        if ((lastItem as any).session_id) {
          setSessionId((lastItem as any).session_id)
        }
      }
    } catch (error) {
      console.error('加载对话历史失败:', error)
    }
  }, [teacherId, setMessages, setSessionId])

  /** 页面初始化 */
  useEffect(() => {
    if (!initialized.current && teacherId) {
      initialized.current = true
      loadHistory()
    }
  }, [teacherId, loadHistory])

  /** 消息列表变化时自动滚动到底部 */
  useEffect(() => {
    if (messages.length > 0) {
      // 使用 setTimeout 确保 DOM 已更新
      setTimeout(() => {
        setScrollIntoId(`msg-${messages.length - 1}`)
      }, 100)
    }
  }, [messages.length])

  /** 页面卸载时清空 chatStore */
  useUnload(() => {
    clearMessages()
  })

  /** 发送消息 */
  const handleSend = useCallback(async () => {
    const text = inputValue.trim()
    if (!text || loading) return

    // 清空输入框
    setInputValue('')

    // 乐观更新：立即添加用户消息
    const userMessage = {
      role: 'user' as const,
      content: text,
      created_at: new Date().toISOString(),
    }
    addMessage(userMessage)
    const userMsgIndex = useChatStore.getState().messages.length - 1

    // 显示 AI 思考中状态
    setLoading(true)

    try {
      const res = await sendMessage(text, teacherId, sessionId || undefined)
      const { reply, session_id } = res.data

      // 更新 sessionId
      if (session_id) {
        setSessionId(session_id)
      }

      // 添加 AI 回复消息
      addMessage({
        role: 'assistant',
        content: reply,
        created_at: new Date().toISOString(),
      })
    } catch (error) {
      console.error('发送消息失败:', error)
      // 标记用户消息为发送失败
      setFailedIndexes((prev) => new Set(prev).add(userMsgIndex))
    } finally {
      setLoading(false)
    }
  }, [inputValue, loading, teacherId, sessionId, addMessage, setLoading, setSessionId])

  /** 重试发送失败的消息 */
  const handleRetry = useCallback(
    async (index: number) => {
      const msg = messages[index]
      if (!msg || msg.role !== 'user') return

      // 移除失败标记
      setFailedIndexes((prev) => {
        const next = new Set(prev)
        next.delete(index)
        return next
      })

      // 重新发送
      setLoading(true)
      try {
        const res = await sendMessage(msg.content, teacherId, sessionId || undefined)
        const { reply, session_id } = res.data

        if (session_id) {
          setSessionId(session_id)
        }

        addMessage({
          role: 'assistant',
          content: reply,
          created_at: new Date().toISOString(),
        })
      } catch (error) {
        console.error('重试发送失败:', error)
        setFailedIndexes((prev) => new Set(prev).add(index))
      } finally {
        setLoading(false)
      }
    },
    [messages, teacherId, sessionId, addMessage, setLoading, setSessionId]
  )

  /** 输入框内容变化 */
  const handleInput = useCallback((e: any) => {
    setInputValue(e.detail.value)
  }, [])

  /** 键盘确认发送 */
  const handleConfirm = useCallback(() => {
    handleSend()
  }, [handleSend])

  /** 是否为空状态（无历史消息） */
  const isEmpty = messages.length === 0 && !loading

  return (
    <View className='chat-page'>
      {/* 消息列表区域 */}
      <ScrollView
        className='chat-page__messages'
        scrollY
        scrollIntoView={scrollIntoId}
        scrollWithAnimation
        enhanced
        showScrollbar={false}
      >
        {/* 空状态 */}
        {isEmpty && (
          <View className='chat-page__empty'>
            <View className='chat-page__empty-icon'>💬</View>
            <Text className='chat-page__empty-text'>
              向 {teacherName} 的数字分身提问吧！
            </Text>
          </View>
        )}

        {/* 消息列表 */}
        {messages.map((msg, index) => {
          const prevMsg = index > 0 ? messages[index - 1] : undefined
          const showTime = shouldShowTimestamp(prevMsg?.created_at, msg.created_at)

          return (
            <View key={`msg-${index}`} id={`msg-${index}`}>
              <ChatBubble
                role={msg.role}
                content={msg.content}
                timestamp={showTime && msg.created_at ? formatTime(msg.created_at) : undefined}
                teacherName={teacherName}
                failed={failedIndexes.has(index)}
                onRetry={() => handleRetry(index)}
              />
            </View>
          )
        })}

        {/* AI 思考中动画 */}
        {loading && (
          <View id='msg-loading' className='chat-page__thinking'>
            <View className='chat-bubble chat-bubble--assistant'>
              <View className='chat-bubble__avatar'>
                <Text className='chat-bubble__avatar-text'>
                  {teacherName.charAt(0).toUpperCase()}
                </Text>
              </View>
              <View className='chat-page__thinking-bubble'>
                <View className='chat-page__thinking-dots'>
                  <View className='chat-page__thinking-dot chat-page__thinking-dot--1' />
                  <View className='chat-page__thinking-dot chat-page__thinking-dot--2' />
                  <View className='chat-page__thinking-dot chat-page__thinking-dot--3' />
                </View>
              </View>
            </View>
          </View>
        )}

        {/* 底部占位，确保最后一条消息不被输入框遮挡 */}
        <View className='chat-page__bottom-spacer' />
      </ScrollView>

      {/* 底部输入区域 */}
      <View className='chat-page__input-bar safe-area-bottom'>
        <View className='chat-page__input-wrapper'>
          <Input
            className='chat-page__input'
            value={inputValue}
            placeholder='输入你的问题...'
            placeholderClass='chat-page__input-placeholder'
            confirmType='send'
            onInput={handleInput}
            onConfirm={handleConfirm}
            adjustPosition
          />
          <View
            className={`chat-page__send-btn ${
              !inputValue.trim() || loading ? 'chat-page__send-btn--disabled' : ''
            }`}
            onClick={handleSend}
          >
            <Text className='chat-page__send-btn-text'>发送</Text>
          </View>
        </View>
      </View>
    </View>
  )
}
