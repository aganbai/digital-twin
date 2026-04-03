import { useState, useCallback, useRef, useEffect } from 'react'
import { View, Text, ScrollView, Input } from '@tarojs/components'
import Taro, { useRouter, useDidShow } from '@tarojs/taro'
import {
  getStudentConversations,
  teacherReply,
  getTakeoverStatus,
  endTakeover,
} from '@/api/chat'
import type { Conversation, TakeoverStatusResponse } from '@/api/chat'
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

export default function StudentChatHistory() {
  const router = useRouter()
  const studentPersonaId = Number(router.params.student_persona_id) || 0
  const studentName = decodeURIComponent(router.params.student_name || '学生')

  const [statusBarHeight, setStatusBarHeight] = useState(44)
  const [messages, setMessages] = useState<Conversation[]>([])
  const [sessionId, setSessionId] = useState('')
  const [takeoverStatus, setTakeoverStatus] = useState<'active' | 'ended' | 'none'>('none')
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(false)
  const [sending, setSending] = useState(false)
  const [inputValue, setInputValue] = useState('')
  const [replyToMsg, setReplyToMsg] = useState<Conversation | null>(null)
  const [scrollIntoId, setScrollIntoId] = useState('')
  const initialized = useRef(false)

  /** 设置导航栏标题 */
  useDidShow(() => {
    Taro.setNavigationBarTitle({ title: `${studentName}的对话` })
  })

  /** 动态获取状态栏高度 */
  useEffect(() => {
    const sysInfo = Taro.getSystemInfoSync()
    setStatusBarHeight(sysInfo.statusBarHeight || 44)
  }, [])

  /** 加载对话记录 */
  const loadConversations = useCallback(async () => {
    if (!studentPersonaId) return
    setLoading(true)
    setLoadError(false)
    try {
      const res = await getStudentConversations(studentPersonaId)
      const data = res.data
      const msgList = data.messages || []
      console.log('[StudentChatHistory] 加载到', msgList.length, '条消息')
      setMessages(msgList)
      setSessionId(data.session_id || '')
      setTakeoverStatus(data.takeover_status || 'none')
    } catch (error) {
      console.error('加载对话记录失败:', error)
      setLoadError(true)
    } finally {
      setLoading(false)
    }
  }, [studentPersonaId])

  /** 初始化 */
  useEffect(() => {
    if (!initialized.current && studentPersonaId) {
      initialized.current = true
      loadConversations()
    }
  }, [studentPersonaId, loadConversations])

  /** 消息列表变化时滚动到底部 */
  useEffect(() => {
    if (messages.length > 0) {
      setTimeout(() => {
        setScrollIntoId(`msg-${messages.length - 1}`)
      }, 100)
    }
  }, [messages.length])

  /** 发送教师回复 */
  const handleSend = useCallback(async () => {
    const text = inputValue.trim()
    if (!text || sending || !sessionId) return

    setSending(true)
    try {
      const res = await teacherReply({
        student_persona_id: studentPersonaId,
        session_id: sessionId,
        content: text,
        reply_to_id: replyToMsg?.id || 0,
      })

      const newMsg: Conversation = {
        id: res.data.conversation_id,
        session_id: sessionId,
        role: 'assistant',
        content: text,
        sender_type: 'teacher',
        reply_to_id: replyToMsg?.id || 0,
        reply_to_content: replyToMsg?.content?.substring(0, 100) || '',
        created_at: res.data.created_at,
      }
      setMessages((prev) => [...prev, newMsg])
      setInputValue('')
      setReplyToMsg(null)
      setTakeoverStatus('active')
    } catch (error) {
      console.error('发送回复失败:', error)
      Taro.showToast({ title: '发送失败', icon: 'none' })
    } finally {
      setSending(false)
    }
  }, [inputValue, sending, sessionId, studentPersonaId, replyToMsg])

  /** 退出接管 */
  const handleEndTakeover = useCallback(async () => {
    if (!sessionId) return
    try {
      await endTakeover(sessionId)
      setTakeoverStatus('none')
      Taro.showToast({ title: '已退出接管', icon: 'success' })
    } catch (error) {
      console.error('退出接管失败:', error)
    }
  }, [sessionId])

  /** 长按消息弹出操作菜单 */
  const handleLongPress = (msg: Conversation) => {
    Taro.showActionSheet({
      itemList: ['引用回复', '复制'],
      success: (res) => {
        if (res.tapIndex === 0) {
          setReplyToMsg(msg)
        } else if (res.tapIndex === 1) {
          Taro.setClipboardData({ data: msg.content })
        }
      },
    })
  }

  /** 获取消息气泡样式类（老师视角：教师消息靠右，学生/AI消息靠左） */
  const getBubbleClass = (msg: Conversation) => {
    const senderType = msg.sender_type || (msg.role === 'user' ? 'student' : 'ai')
    if (senderType === 'teacher') return 'chat-bubble--teacher'
    if (senderType === 'student') return 'chat-bubble--student'
    return 'chat-bubble--ai'
  }

  /** 获取发送者标签 */
  const getSenderLabel = (msg: Conversation) => {
    const senderType = msg.sender_type || (msg.role === 'user' ? 'student' : 'ai')
    if (senderType === 'student') return `👨‍🎓 ${studentName}`
    if (senderType === 'teacher') return '👨‍🏫 我(真人)'
    return '🤖 AI分身'
  }

  /** 获取头像文字 */
  const getAvatarText = (msg: Conversation) => {
    const senderType = msg.sender_type || (msg.role === 'user' ? 'student' : 'ai')
    if (senderType === 'student') return studentName.charAt(0).toUpperCase()
    if (senderType === 'teacher') return '师'
    return 'AI'
  }

  if (loading) {
    return (
      <View className='student-chat'>
        <View className='student-chat__loading'>
          <Text className='student-chat__loading-text'>加载中...</Text>
        </View>
      </View>
    )
  }

  return (
    <View className='student-chat'>
      {/* 自定义导航栏 */}
      <View className='student-chat__navbar' style={{ paddingTop: `${statusBarHeight}px` }}>
        <View className='student-chat__navbar-back' onClick={() => {
          const pages = Taro.getCurrentPages()
          if (pages.length > 1) {
            Taro.navigateBack()
          } else {
            Taro.switchTab({ url: '/pages/home/index' })
          }
        }}>
          <Text className='student-chat__navbar-back-icon'>←</Text>
        </View>
        <Text className='student-chat__navbar-title'>{studentName}的对话</Text>
        <View className='student-chat__navbar-placeholder' />
      </View>

      {/* 接管状态栏 */}
      {takeoverStatus === 'active' && (
        <View className='student-chat__takeover-bar'>
          <Text className='student-chat__takeover-text'>🟢 接管中 — AI 已暂停</Text>
          <View className='student-chat__takeover-btn' onClick={handleEndTakeover}>
            <Text className='student-chat__takeover-btn-text'>退出接管</Text>
          </View>
        </View>
      )}

      {/* 消息列表 */}
      <ScrollView
        className='student-chat__messages'
        scrollY
        scrollIntoView={scrollIntoId}
        scrollWithAnimation
        enhanced
        showScrollbar={false}
      >
        {messages.length === 0 && !loadError && (
          <View className='student-chat__empty'>
            <Text className='student-chat__empty-text'>暂无对话记录</Text>
          </View>
        )}

        {loadError && (
          <View className='student-chat__empty'>
            <Text className='student-chat__empty-text'>加载失败</Text>
            <View
              className='student-chat__retry-btn'
              onClick={loadConversations}
            >
              <Text className='student-chat__retry-btn-text'>点击重试</Text>
            </View>
          </View>
        )}

        {messages.map((msg, index) => {
          const prevMsg = index > 0 ? messages[index - 1] : undefined
          const showTime = shouldShowTimestamp(prevMsg?.created_at, msg.created_at)
          const senderType = msg.sender_type || (msg.role === 'user' ? 'student' : 'ai')
          const isStudent = senderType === 'student'
          const isTeacher = senderType === 'teacher'
          const isAI = !isStudent && !isTeacher
          const senderLabel = getSenderLabel(msg)

          return (
            <View key={msg.id || index} id={`msg-${index}`} onLongPress={() => handleLongPress(msg)}>
              {/* 时间戳 */}
              {showTime && msg.created_at && (
                <View className='chat-bubble__timestamp'>
                  <Text className='chat-bubble__timestamp-text'>{formatTime(msg.created_at)}</Text>
                </View>
              )}

              {/* 消息行（老师视角：教师消息靠右，学生/AI消息靠左） */}
              <View className={`chat-bubble ${getBubbleClass(msg)}`}>
                {/* 非教师消息：左侧头像（学生/AI） */}
                {!isTeacher && (
                  <View className={`chat-bubble__avatar ${isStudent ? 'chat-bubble__avatar--student' : 'chat-bubble__avatar--ai'}`}>
                    <Text className='chat-bubble__avatar-text'>
                      {getAvatarText(msg)}
                    </Text>
                  </View>
                )}

                <View className='chat-bubble__body'>
                  {/* 发送者标签 */}
                  <Text className={`chat-bubble__sender-label ${isTeacher ? 'chat-bubble__sender-label--teacher' : isAI ? 'chat-bubble__sender-label--ai' : ''}`}>
                    {senderLabel}
                  </Text>

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
                  <View className={`chat-bubble__content chat-bubble__content--${isStudent ? 'student' : isTeacher ? 'teacher' : 'ai'}`}>
                    <Text className={`chat-bubble__text chat-bubble__text--dark`}>
                      {msg.content}
                    </Text>
                  </View>
                </View>

                {/* 教师消息：右侧头像 */}
                {isTeacher && (
                  <View className='chat-bubble__avatar chat-bubble__avatar--teacher'>
                    <Text className='chat-bubble__avatar-text'>
                      {getAvatarText(msg)}
                    </Text>
                  </View>
                )}
              </View>
            </View>
          )
        })}

        <View className='student-chat__bottom-spacer' />
      </ScrollView>

      {/* 底部输入区域 */}
      <View className='student-chat__input-bar safe-area-bottom'>
        {/* 引用回复预览 */}
        {replyToMsg && (
          <View className='student-chat__quote-preview'>
            <View className='student-chat__quote-preview-content'>
              <Text className='student-chat__quote-preview-text'>
                {replyToMsg.content.length > 40
                  ? replyToMsg.content.substring(0, 40) + '...'
                  : replyToMsg.content}
              </Text>
            </View>
            <View className='student-chat__quote-preview-close' onClick={() => setReplyToMsg(null)}>
              <Text className='student-chat__quote-preview-close-text'>✕</Text>
            </View>
          </View>
        )}
        <View className='student-chat__input-wrapper'>
          <Input
            className='student-chat__input'
            value={inputValue}
            placeholder='输入回复内容...'
            placeholderClass='student-chat__input-placeholder'
            confirmType='send'
            onInput={(e) => setInputValue(e.detail.value)}
            onConfirm={handleSend}
            adjustPosition
          />
          <View
            className={`student-chat__send-btn ${
              !inputValue.trim() || sending ? 'student-chat__send-btn--disabled' : ''
            }`}
            onClick={handleSend}
          >
            <Text className='student-chat__send-btn-text'>
              {sending ? '...' : '发送'}
            </Text>
          </View>
        </View>
      </View>
    </View>
  )
}
