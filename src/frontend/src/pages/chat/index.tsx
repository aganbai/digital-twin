import { useState, useEffect, useCallback, useRef } from 'react'
import { View, Text, ScrollView, Input } from '@tarojs/components'
import Taro, { useRouter, useDidShow, useUnload } from '@tarojs/taro'
import { sendMessage, getConversations, chatStream, getTakeoverStatus } from '@/api/chat'
import type { Conversation, TakeoverStatusResponse, ThinkingStepEvent } from '@/api/chat'
import { uploadFile } from '@/api/upload'
import { useChatStore } from '@/store'
import { formatTime } from '@/utils/format'
import { getUserInfo } from '@/utils/storage'
import EmojiPanel from '@/components/EmojiPanel'
import ThinkingPanel, { ThinkingStep } from '@/components/ThinkingPanel'
import VoiceInput, { VoiceButton } from '@/components/VoiceInput'
import PlusPanel from '@/components/PlusPanel'
import AvatarPopup from '@/components/AvatarPopup'
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
  const teacherId = Number(router.params.teacher_id) || Number(router.params.teacher_persona_id) || 0
  const teacherName = decodeURIComponent(router.params.teacher_name || '教师')
  // V2.0 中 teacher_id 实际就是 teacher_persona_id
  const teacherPersonaId = teacherId
  // 老师视角：学生分身 ID
  const studentPersonaId = Number(router.params.student_persona_id) || 0
  // 班级 ID（用于学生查看班级信息）
  const classId = Number(router.params.class_id) || 0

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

  /** 状态栏高度 */
  const [statusBarHeight, setStatusBarHeight] = useState(44)
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
  /** 引用回复的消息 */
  const [replyToMsg, setReplyToMsg] = useState<{id?: number, content: string} | null>(null)
  /** 流式回复内容累积 */
  const [streamingContent, setStreamingContent] = useState('')
  /** 是否正在流式接收 */
  const [isStreaming, setIsStreaming] = useState(false)
  /** 当前流式请求任务 */
  const streamTaskRef = useRef<any>(null)
  /** 接管状态 */
  const [takeoverInfo, setTakeoverInfo] = useState<TakeoverStatusResponse | null>(null)
  /** 是否显示 Emoji 面板 */
  const [showEmoji, setShowEmoji] = useState(false)
  /** 是否处于语音输入模式 */
  const [voiceMode, setVoiceMode] = useState(false)
  /** 是否显示+号面板 */
  const [showPlusPanel, setShowPlusPanel] = useState(false)
  /** 思考步骤列表 */
  const [thinkingSteps, setThinkingSteps] = useState<ThinkingStep[]>([])
  /** 是否显示头像弹窗 */
  const [showAvatarPopup, setShowAvatarPopup] = useState(false)
  /** 当前用户角色 */
  const currentUserRole = getUserInfo()?.role || 'student'
  /** 头像弹窗参数 */
  const [avatarPopupTargetId, setAvatarPopupTargetId] = useState<number>(0)
  const [avatarPopupClassId, setAvatarPopupClassId] = useState<number | undefined>(undefined)

  /** 语音识别结果填充到输入框 */
  const handleVoiceResult = useCallback((text: string) => {
    setInputValue((prev) => prev + text)
    setVoiceMode(false) // 识别完成后切回文字模式
  }, [])

  /** 设置导航栏标题 */
  useDidShow(() => {
    Taro.setNavigationBarTitle({ title: teacherName })
  })

  /** 动态获取状态栏高度 */
  useEffect(() => {
    const sysInfo = Taro.getSystemInfoSync()
    setStatusBarHeight(sysInfo.statusBarHeight || 44)
  }, [])

  /** 加载历史消息 */
  const loadHistory = useCallback(async () => {
    try {
      const res = await getConversations({
        teacher_persona_id: teacherPersonaId,
        page: 1,
        page_size: 50,
      })
      const items = res.data.items || []
      // 接口返回按时间正序（ASC），直接使用
      const sorted = items
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
        const lastItem = sorted[sorted.length - 1] // 正序中最后一条是最新的
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
  }, [teacherPersonaId, setMessages, setSessionId])

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
    if (!initialized.current && teacherPersonaId) {
      initialized.current = true
      loadHistory()
    }
  }, [teacherPersonaId, loadHistory])

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
      // 清空思考步骤
      setThinkingSteps([])

      let accumulatedContent = ''

      try {
        const task = chatStream(
          text,
          teacherPersonaId,
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
            onThinkingStep: (event: ThinkingStepEvent) => {
              setThinkingSteps((prev) => [...prev, {
                step: event.step,
                status: event.status,
                message: event.message,
                detail: event.detail,
                duration_ms: event.duration_ms,
                timestamp: event.timestamp,
              }])
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
    [teacherPersonaId, sessionId, addMessage, setLoading, setSessionId],
  )

  /** 降级到普通接口发送 */
  const handleFallbackSend = useCallback(
    async (text: string, userMsgIndex: number, attachment?: { url: string; type: string; name: string }) => {
      setIsStreaming(false)
      setStreamingContent('')

      try {
        const res = await sendMessage(text, teacherPersonaId, sessionId || undefined, attachment)
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
        Taro.showToast({ title: '发送失败，请点击重试', icon: 'none' })
      } finally {
        setLoading(false)
      }
    },
    [teacherPersonaId, sessionId, addMessage, setLoading, setSessionId],
  )

  /** 发送消息 */
  const handleSend = useCallback(async () => {
    const text = inputValue.trim()
    if ((!text && !attachmentInfo) || loading) return

    // 清空输入框和附件
    setInputValue('')
    const currentAttachment = attachmentInfo
    setAttachmentInfo(null)
    const currentReply = replyToMsg
    setReplyToMsg(null)

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

  /** 长按消息弹出操作菜单 */
  const handleLongPress = useCallback((msg: any) => {
    Taro.showActionSheet({
      itemList: ['引用回复', '复制'],
      success: (res) => {
        if (res.tapIndex === 0) {
          setReplyToMsg({ id: msg.id, content: msg.content })
        } else if (res.tapIndex === 1) {
          Taro.setClipboardData({ data: msg.content })
        }
      },
    })
  }, [])

  /** 头像点击处理 */
  const handleAvatarClick = useCallback((msg: any) => {
    const senderType = msg.sender_type || (msg.role === 'user' ? 'student' : 'ai')
    const isCurrentUser = senderType === 'student' || msg.role === 'user'

    // 如果当前用户是学生
    if (currentUserRole === 'student') {
      // 学生点击自己的头像不触发
      if (isCurrentUser) return
      // 学生点击老师头像 → 显示班级信息
      // targetId 为 teacher_persona_id
      setAvatarPopupTargetId(teacherPersonaId)
      // classId 从路由参数获取
      setAvatarPopupClassId(classId || undefined)
      setShowAvatarPopup(true)
    } else if (currentUserRole === 'teacher') {
      // 老师点击学生头像 → 显示学生信息
      // 只有学生消息才有学生信息
      if (!isCurrentUser) return
      // targetId 为 student_persona_id（从路由参数获取）
      if (studentPersonaId === 0) {
        // 如果路由参数没有 student_persona_id，尝试从消息中获取
        const msgStudentId = msg.student_persona_id || msg.sender_id || 0
        if (msgStudentId === 0) {
          Taro.showToast({ title: '无法获取学生信息', icon: 'none' })
          return
        }
        setAvatarPopupTargetId(msgStudentId)
      } else {
        setAvatarPopupTargetId(studentPersonaId)
      }
      setAvatarPopupClassId(undefined)
      setShowAvatarPopup(true)
    }
  }, [currentUserRole, teacherPersonaId, studentPersonaId, classId])

  /** 获取消息气泡样式类 */
  const getBubbleClass = (msg: any) => {
    const senderType = msg.sender_type || (msg.role === 'user' ? 'student' : 'ai')
    if (senderType === 'teacher') return 'chat-bubble--teacher'
    if (senderType === 'teacher_push') return 'chat-bubble--teacher-push'
    if (senderType === 'student' || msg.role === 'user') return 'chat-bubble--user'
    return 'chat-bubble--assistant'
  }

  /** 获取发送者标签 */
  const getSenderLabel = (msg: any) => {
    const senderType = msg.sender_type || (msg.role === 'user' ? 'student' : 'ai')
    if (senderType === 'teacher') return `👨‍🏫 ${teacherName}(真人)`
    if (senderType === 'teacher_push') return `📢 ${teacherName}(通知)`
    if (senderType === 'ai') return `🤖 ${teacherName}`
    return ''
  }

  /** Emoji 选择回调 */
  const handleEmojiSelect = useCallback((emoji: string) => {
    setInputValue((prev) => prev + emoji)
  }, [])

  /** 是否为空状态（无历史消息） */
  const isEmpty = messages.length === 0 && !loading && !isStreaming

  return (
    <View className='chat-page'>
      {/* 自定义导航栏 */}
      <View className='chat-page__navbar' style={{ paddingTop: `${statusBarHeight}px` }}>
        <View className='chat-page__navbar-back' onClick={() => {
          const pages = Taro.getCurrentPages()
          if (pages.length > 1) {
            Taro.navigateBack()
          } else {
            // 没有父页时进入发现页
            Taro.switchTab({ url: '/pages/discover/index' })
          }
        }}>
          <Text className='chat-page__navbar-back-icon'>←</Text>
        </View>
        {/* 老师头像（学生点击可查看班级信息） */}
        {currentUserRole === 'student' && (
          <View 
            className='chat-page__teacher-avatar'
            onClick={() => {
              setAvatarPopupTargetId(teacherPersonaId)
              setAvatarPopupClassId(classId || undefined)
              setShowAvatarPopup(true)
            }}
          >
            <Text className='chat-page__teacher-avatar-text'>
              {teacherName.charAt(0).toUpperCase()}
            </Text>
          </View>
        )}
        <Text className='chat-page__navbar-title'>{teacherName}</Text>
        <View className='chat-page__navbar-action' onClick={() => {
          Taro.navigateTo({ url: '/pages/history/index' })
        }}>
          <Text className='chat-page__navbar-action-text'>历史</Text>
        </View>
      </View>

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
            <View key={`msg-${index}`} id={`msg-${index}`} onLongPress={() => handleLongPress(msg)}>
              {/* 时间戳 */}
              {showTime && msg.created_at && (
                <View className='chat-bubble__timestamp'>
                  <Text className='chat-bubble__timestamp-text'>{formatTime(msg.created_at)}</Text>
                </View>
              )}

              {/* 教师推送消息 - 居中通知卡片 */}
              {senderType === 'teacher_push' ? (
                <View className='chat-push-card'>
                  <View className='chat-push-card__header'>
                    <Text className='chat-push-card__label'>📢 老师通知</Text>
                    {msg.created_at && (
                      <Text className='chat-push-card__time'>{formatTime(msg.created_at)}</Text>
                    )}
                  </View>
                  <Text className='chat-push-card__content'>{msg.content}</Text>
                </View>
              ) : (
              <View className={`chat-bubble ${getBubbleClass(msg)}`}>
                {/* 非用户消息：左侧头像 */}
                {!isUser && (
                  <View 
                    className={`chat-bubble__avatar ${isTeacher ? 'chat-bubble__avatar--teacher' : ''}`}
                    onClick={() => handleAvatarClick(msg)}
                  >
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

                {/* 用户消息：右侧头像（仅老师查看学生消息时显示） */}
                {isUser && currentUserRole === 'teacher' && (
                  <View 
                    className='chat-bubble__avatar chat-bubble__avatar--student'
                    onClick={() => handleAvatarClick(msg)}
                  >
                    <Text className='chat-bubble__avatar-text'>生</Text>
                  </View>
                )}
              </View>
              )}

              {/* 发送失败提示 */}
              {failedIndexes.has(index) && (
                <View className='chat-bubble__failed' onClick={() => handleRetry(index)}>
                  <Text className='chat-bubble__failed-text'>发送失败，点击重试</Text>
                </View>
              )}
            </View>
          )
        })}

        {/* 思考过程展示 */}
        {isStreaming && thinkingSteps.length > 0 && (
          <ThinkingPanel steps={thinkingSteps} teacherName={teacherName} />
        )}

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
        {/* 引用回复预览 */}
        {replyToMsg && (
          <View className='chat-page__quote-preview'>
            <View className='chat-page__quote-preview-content'>
              <Text className='chat-page__quote-preview-text'>
                {replyToMsg.content.length > 40
                  ? replyToMsg.content.substring(0, 40) + '...'
                  : replyToMsg.content}
              </Text>
            </View>
            <View className='chat-page__quote-preview-close' onClick={() => setReplyToMsg(null)}>
              <Text className='chat-page__quote-preview-close-text'>✕</Text>
            </View>
          </View>
        )}
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
          {/* 语音按钮 */}
          <View
            className='chat-page__voice-btn'
            onClick={() => {
              setVoiceMode(!voiceMode)
              setShowEmoji(false)
              setShowPlusPanel(false)
            }}
          >
            <Text className='chat-page__voice-btn-text'>{voiceMode ? '⌨️' : '🔊'}</Text>
          </View>

          {/* 语音模式：显示按住说话按钮 */}
          {voiceMode ? (
            <VoiceButton
              onPress={() => {
                Taro.showToast({ title: '开始录音', icon: 'none' })
              }}
              onRelease={() => {
                Taro.showToast({ title: '结束录音', icon: 'none' })
              }}
              onCancel={() => {
                Taro.showToast({ title: '取消录音', icon: 'none' })
              }}
            />
          ) : (
            <>
              {/* 文字输入框 */}
              <Input
                className='chat-page__input'
                value={inputValue}
                placeholder='输入你的问题...'
                placeholderClass='chat-page__input-placeholder'
                confirmType='send'
                onInput={handleInput}
                onConfirm={handleConfirm}
                onFocus={() => {
                  setShowEmoji(false)
                  setShowPlusPanel(false)
                }}
                adjustPosition={!showEmoji}
              />

              {/* Emoji 按钮 */}
              <View
                className='chat-page__emoji-btn'
                onClick={() => {
                  setShowEmoji(!showEmoji)
                  setShowPlusPanel(false)
                }}
              >
                <Text className='chat-page__emoji-btn-text'>{showEmoji ? '⌨️' : '😊'}</Text>
              </View>

              {/* +号按钮 */}
              <View
                className='chat-page__plus-btn'
                onClick={() => {
                  setShowPlusPanel(!showPlusPanel)
                  setShowEmoji(false)
                }}
              >
                <Text className='chat-page__plus-btn-text'>{showPlusPanel ? '✕' : '+'}</Text>
              </View>
            </>
          )}

          {/* 发送按钮（仅非语音模式显示） */}
          {!voiceMode && (
            <View
              className={`chat-page__send-btn ${
                (!inputValue.trim() && !attachmentInfo) || loading ? 'chat-page__send-btn--disabled' : ''
              }`}
              onClick={handleSend}
            >
              <Text className='chat-page__send-btn-text'>发送</Text>
            </View>
          )}
        </View>

        {/* Emoji 面板 */}
        <EmojiPanel visible={showEmoji} onSelect={handleEmojiSelect} />
      </View>

      {/* +号多功能面板 */}
      <PlusPanel
        visible={showPlusPanel}
        onClose={() => setShowPlusPanel(false)}
        onFileSelect={(files) => {
          console.log('选择文件:', files)
          setShowPlusPanel(false)
        }}
        onImageSelect={(images) => {
          console.log('选择图片:', images)
          setShowPlusPanel(false)
        }}
        onCameraCapture={(image) => {
          console.log('拍摄图片:', image)
          setShowPlusPanel(false)
        }}
      />

      {/* 头像点击弹窗 */}
      <AvatarPopup
        visible={showAvatarPopup}
        onClose={() => setShowAvatarPopup(false)}
        userRole={currentUserRole as 'student' | 'teacher'}
        targetId={avatarPopupTargetId}
        classId={avatarPopupClassId}
      />
    </View>
  )
}
