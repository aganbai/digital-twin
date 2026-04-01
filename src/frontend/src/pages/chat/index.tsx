import { useState, useEffect, useCallback, useRef } from 'react'
import { View, Text, ScrollView, Input } from '@tarojs/components'
import Taro, { useRouter, useDidShow, useUnload } from '@tarojs/taro'
import { sendMessage, getConversations, chatStream, getTakeoverStatus } from '@/api/chat'
import type { Conversation, TakeoverStatusResponse } from '@/api/chat'
import { uploadFile } from '@/api/upload'
import { useChatStore } from '@/store'
import { formatTime } from '@/utils/format'
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
  /** M4: 附件相关状态 */
  const [attachmentInfo, setAttachmentInfo] = useState<{
    url: string
    type: string
    name: string
  } | null>(null)
  const [uploading, setUploading] = useState(false)
  /** 流式回复内容累积 */
  const [streamingContent, setStreamingContent] = useState('')
  /** 是否正在流式接收 */
  const [isStreaming, setIsStreaming] = useState(false)
  /** 当前流式请求任务 */
  const streamTaskRef = useRef<any>(null)
  /** 接管状态 */
  const [takeoverInfo, setTakeoverInfo] = useState<TakeoverStatusResponse | null>(null)

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
        sender_type: item.sender_type,
        reply_to_id: item.reply_to_id,
        reply_to_content: item.reply_to_content,
        created_at: item.created_at,
      }))
      setMessages(historyMessages)

      // 从历史记录中获取最新的 session_id
      if (sorted.length > 0) {
        const lastItem = items[0] // 倒序中第一条是最新的
        if ((lastItem as any).session_id) {
          const sid = (lastItem as any).session_id
          setSessionId(sid)
          // 检查接管状态
          checkTakeover(sid)
        }
      }
    } catch (error) {
      console.error('加载对话历史失败:', error)
    }
  }, [teacherId, setMessages, setSessionId])

  /** 检查接管状态 */
  const checkTakeover = useCallback(async (sid: string) => {
    if (!sid) return
    try {
      const res = await getTakeoverStatus(sid)
      setTakeoverInfo(res.data)
    } catch {
      // 忽略
    }
  }, [])

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
      setTimeout(() => {
        setScrollIntoId(`msg-${messages.length - 1}`)
      }, 100)
    }
  }, [messages.length])

  /** 流式内容变化时滚动到底部 */
  useEffect(() => {
    if (isStreaming) {
      setTimeout(() => {
        setScrollIntoId('msg-streaming')
      }, 50)
    }
  }, [streamingContent, isStreaming])

  /** 页面卸载时清空 chatStore 和中断流式请求 */
  useUnload(() => {
    if (streamTaskRef.current) {
      streamTaskRef.current.abort?.()
    }
    clearMessages()
  })

  /** 使用 SSE 流式发送消息 */
  const handleStreamSend = useCallback(
    async (text: string, attachment?: { url: string; type: string; name: string }) => {
      // 乐观更新：立即添加用户消息
      const userMessage = {
        role: 'user' as const,
        content: text,
        sender_type: 'student' as const,
        created_at: new Date().toISOString(),
      }
      addMessage(userMessage)
      const userMsgIndex = useChatStore.getState().messages.length - 1

      // 开始流式接收
      setIsStreaming(true)
      setStreamingContent('')
      setLoading(true)

      let accumulatedContent = ''

      try {
        const task = chatStream(
          text,
          teacherId,
          {
            onStart: (newSessionId) => {
              if (newSessionId) {
                setSessionId(newSessionId)
              }
            },
            onDelta: (content) => {
              accumulatedContent += content
              setStreamingContent(accumulatedContent)
            },
            onDone: () => {
              // 流式完成，将累积内容添加为完整消息
              addMessage({
                role: 'assistant',
                content: accumulatedContent,
                sender_type: 'ai',
                created_at: new Date().toISOString(),
              })
              setIsStreaming(false)
              setStreamingContent('')
              setLoading(false)
              streamTaskRef.current = null
            },
            onError: (code, message) => {
              console.error('流式对话错误:', code, message)

              // 接管状态：code=40030
              if (code === 40030) {
                setTakeoverInfo({
                  is_taken_over: true,
                  teacher_persona_id: 0,
                  teacher_nickname: '',
                  started_at: '',
                })
                setIsStreaming(false)
                setStreamingContent('')
                setLoading(false)
                streamTaskRef.current = null
                Taro.showToast({ title: message || '老师正在亲自回复中', icon: 'none' })
                return
              }

              // 降级到普通接口
              if (accumulatedContent) {
                addMessage({
                  role: 'assistant',
                  content: accumulatedContent,
                  sender_type: 'ai',
                  created_at: new Date().toISOString(),
                })
              }
              setIsStreaming(false)
              setStreamingContent('')
              setLoading(false)
              streamTaskRef.current = null

              // 如果没有任何内容，尝试普通接口
              if (!accumulatedContent) {
                handleFallbackSend(text, userMsgIndex, attachment)
              }
            },
          },
          sessionId || undefined,
          attachment,
        )
        streamTaskRef.current = task
      } catch {
        // 流式请求失败，降级到普通接口
        handleFallbackSend(text, userMsgIndex)
      }
    },
    [teacherId, sessionId, addMessage, setLoading, setSessionId],
  )

  /** 降级到普通接口发送 */
  const handleFallbackSend = useCallback(
    async (text: string, userMsgIndex: number, attachment?: { url: string; type: string; name: string }) => {
      setIsStreaming(false)
      setStreamingContent('')

      try {
        const res = await sendMessage(text, teacherId, sessionId || undefined, attachment)
        const { reply, session_id } = res.data

        if (session_id) {
          setSessionId(session_id)
        }

        addMessage({
          role: 'assistant',
          content: reply,
          sender_type: 'ai',
          created_at: new Date().toISOString(),
        })
      } catch (error) {
        console.error('发送消息失败:', error)
        setFailedIndexes((prev) => new Set(prev).add(userMsgIndex))
      } finally {
        setLoading(false)
      }
    },
    [teacherId, sessionId, addMessage, setLoading, setSessionId],
  )

  /** 发送消息 */
  const handleSend = useCallback(async () => {
    const text = inputValue.trim()
    if ((!text && !attachmentInfo) || loading) return

    // 清空输入框和附件
    setInputValue('')
    const currentAttachment = attachmentInfo
    setAttachmentInfo(null)

    // 优先使用流式接口
    handleStreamSend(text || `[附件] ${currentAttachment?.name || ''}`, currentAttachment || undefined)
  }, [inputValue, loading, handleStreamSend, attachmentInfo])

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

      // 使用流式重新发送
      setLoading(true)
      handleStreamSend(msg.content)
    },
    [messages, handleStreamSend, setLoading],
  )

  /** 输入框内容变化 */
  const handleInput = useCallback((e: any) => {
    setInputValue(e.detail.value)
  }, [])

  /** 键盘确认发送 */
  const handleConfirm = useCallback(() => {
    handleSend()
  }, [handleSend])

  /** 获取消息气泡样式类 */
  const getBubbleClass = (msg: any) => {
    const senderType = msg.sender_type || (msg.role === 'user' ? 'student' : 'ai')
    if (senderType === 'teacher') return 'chat-bubble--teacher'
    if (senderType === 'student' || msg.role === 'user') return 'chat-bubble--user'
    return 'chat-bubble--assistant'
  }

  /** 获取发送者标签 */
  const getSenderLabel = (msg: any) => {
    const senderType = msg.sender_type || (msg.role === 'user' ? 'student' : 'ai')
    if (senderType === 'teacher') return `👨‍🏫 ${teacherName}(真人)`
    if (senderType === 'ai') return `🤖 ${teacherName}`
    return ''
  }

  /** 是否为空状态（无历史消息） */
  const isEmpty = messages.length === 0 && !loading && !isStreaming

  return (
    <View className='chat-page'>
      {/* 接管状态提示条 */}
      {takeoverInfo?.is_taken_over && (
        <View className='chat-page__takeover-bar'>
          <Text className='chat-page__takeover-text'>
            ⚡ 老师正在亲自回复中
          </Text>
        </View>
      )}

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
          const senderType = msg.sender_type || (msg.role === 'user' ? 'student' : 'ai')
          const isUser = senderType === 'student' || msg.role === 'user'
          const isTeacher = senderType === 'teacher'
          const senderLabel = getSenderLabel(msg)

          return (
            <View key={`msg-${index}`} id={`msg-${index}`}>
              {/* 时间戳 */}
              {showTime && msg.created_at && (
                <View className='chat-bubble__timestamp'>
                  <Text className='chat-bubble__timestamp-text'>{formatTime(msg.created_at)}</Text>
                </View>
              )}

              {/* 消息行 */}
              <View className={`chat-bubble ${getBubbleClass(msg)}`}>
                {/* 非用户消息：左侧头像 */}
                {!isUser && (
                  <View className={`chat-bubble__avatar ${isTeacher ? 'chat-bubble__avatar--teacher' : ''}`}>
                    <Text className='chat-bubble__avatar-text'>
                      {isTeacher ? '师' : teacherName.charAt(0).toUpperCase()}
                    </Text>
                  </View>
                )}

                <View className='chat-bubble__body'>
                  {/* 发送者标签 */}
                  {senderLabel && (
                    <Text className={`chat-bubble__sender-label ${isTeacher ? 'chat-bubble__sender-label--teacher' : ''}`}>
                      {senderLabel}
                    </Text>
                  )}

                  {/* 引用消息 */}
                  {msg.reply_to_id && msg.reply_to_id > 0 && msg.reply_to_content && (
                    <View className='chat-bubble__quote'>
                      <Text className='chat-bubble__quote-text'>
                        {msg.reply_to_content.length > 50
                          ? msg.reply_to_content.substring(0, 50) + '...'
                          : msg.reply_to_content}
                      </Text>
                    </View>
                  )}

                  {/* 气泡内容 */}
                  <View className={`chat-bubble__content chat-bubble__content--${isUser ? 'user' : isTeacher ? 'teacher' : 'assistant'}`}>
                    <Text className={`chat-bubble__text chat-bubble__text--${isUser ? 'user' : 'dark'}`}>
                      {msg.content}
                    </Text>
                  </View>
                </View>
              </View>

              {/* 发送失败提示 */}
              {failedIndexes.has(index) && (
                <View className='chat-bubble__failed' onClick={() => handleRetry(index)}>
                  <Text className='chat-bubble__failed-text'>发送失败，点击重试</Text>
                </View>
              )}
            </View>
          )
        })}

        {/* 流式回复气泡 */}
        {isStreaming && (
          <View id='msg-streaming'>
            <View className='chat-bubble chat-bubble--assistant'>
              <View className='chat-bubble__avatar'>
                <Text className='chat-bubble__avatar-text'>
                  {teacherName.charAt(0).toUpperCase()}
                </Text>
              </View>
              <View className='chat-bubble__body'>
                <Text className='chat-bubble__sender-label'>🤖 {teacherName}</Text>
                <View className='chat-bubble__content chat-bubble__content--assistant'>
                  <Text className='chat-bubble__text chat-bubble__text--dark'>
                    {streamingContent || ''}
                  </Text>
                </View>
              </View>
            </View>
            {/* 光标闪烁动画 */}
            {streamingContent && (
              <View className='chat-page__cursor-wrap'>
                <View className='chat-page__cursor' />
              </View>
            )}
          </View>
        )}

        {/* AI 思考中动画（仅在流式未开始时显示） */}
        {loading && !isStreaming && (
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
        {/* M4: 附件预览 */}
        {attachmentInfo && (
          <View className='chat-page__attachment-preview'>
            <Text className='chat-page__attachment-name'>📎 {attachmentInfo.name}</Text>
            <View
              className='chat-page__attachment-remove'
              onClick={() => setAttachmentInfo(null)}
            >
              <Text className='chat-page__attachment-remove-text'>✕</Text>
            </View>
          </View>
        )}
        <View className='chat-page__input-wrapper'>
          {/* M4: 附件按钮 */}
          <View
            className={`chat-page__attach-btn ${uploading ? 'chat-page__attach-btn--disabled' : ''}`}
            onClick={() => {
              if (uploading) return
              Taro.chooseMessageFile({
                count: 1,
                type: 'file',
                extension: ['pdf', 'docx', 'txt', 'md', 'jpg', 'jpeg', 'png'],
                success: async (res) => {
                  const file = res.tempFiles[0]
                  if (file.size > 10 * 1024 * 1024) {
                    Taro.showToast({ title: '文件不能超过10MB', icon: 'none' })
                    return
                  }
                  setUploading(true)
                  try {
                    const uploadRes = await uploadFile(file.path, 'assignment')
                    const ext = file.name.split('.').pop()?.toLowerCase() || ''
                    const typeMap: Record<string, string> = {
                      pdf: 'pdf', docx: 'docx', txt: 'txt', md: 'txt',
                      jpg: 'image', jpeg: 'image', png: 'image',
                    }
                    setAttachmentInfo({
                      url: uploadRes.data.url,
                      type: typeMap[ext] || 'general',
                      name: file.name,
                    })
                  } catch (err) {
                    console.error('上传文件失败:', err)
                    Taro.showToast({ title: '上传失败', icon: 'none' })
                  } finally {
                    setUploading(false)
                  }
                },
              })
            }}
          >
            <Text className='chat-page__attach-btn-text'>{uploading ? '...' : '+'}</Text>
          </View>
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
              (!inputValue.trim() && !attachmentInfo) || loading ? 'chat-page__send-btn--disabled' : ''
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
