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

export default function StudentChatHistory() {
  const router = useRouter()
  const studentPersonaId = Number(router.params.student_persona_id) || 0
  const studentName = decodeURIComponent(router.params.student_name || '学生')

  const [messages, setMessages] = useState<Conversation[]>([])
  const [sessionId, setSessionId] = useState('')
  const [takeoverStatus, setTakeoverStatus] = useState<'active' | 'ended' | 'none'>('none')
  const [loading, setLoading] = useState(true)
  const [sending, setSending] = useState(false)
  const [inputValue, setInputValue] = useState('')
  const [replyToMsg, setReplyToMsg] = useState<Conversation | null>(null)
  const [scrollIntoId, setScrollIntoId] = useState('')
  const initialized = useRef(false)

  /** 设置导航栏标题 */
  useDidShow(() => {
    Taro.setNavigationBarTitle({ title: `${studentName}的对话记录` })
  })

  /** 加载对话记录 */
  const loadConversations = useCallback(async () => {
    if (!studentPersonaId) return
    setLoading(true)
    try {
      const res = await getStudentConversations(studentPersonaId)
      const data = res.data
      setMessages(data.messages || [])
      setSessionId(data.session_id || '')
      setTakeoverStatus(data.takeover_status || 'none')
    } catch (error) {
      console.error('加载对话记录失败:', error)
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

  /** 查询接管状态 */
  const checkTakeoverStatus = useCallback(async () => {
    if (!sessionId) return
    try {
      const res = await getTakeoverStatus(sessionId)
      setTakeoverStatus(res.data.is_taken_over ? 'active' : 'none')
    } catch {
      // 忽略
    }
  }, [sessionId])

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

      // 添加新消息到列表
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

  /** 选择引用回复 */
  const handleQuoteReply = (msg: Conversation) => {
    setReplyToMsg(msg)
  }

  /** 取消引用 */
  const handleCancelQuote = () => {
    setReplyToMsg(null)
  }

  /** 获取发送者标签 */
  const getSenderLabel = (msg: Conversation) => {
    switch (msg.sender_type) {
      case 'student':
        return `👨‍🎓 ${studentName}`
      case 'ai':
        return '🤖 AI分身'
      case 'teacher':
        return '👨‍🏫 我(真人)'
      default:
        return msg.role === 'user' ? `👨‍🎓 ${studentName}` : '🤖 AI分身'
    }
  }

  /** 获取消息样式类 */
  const getMsgClass = (msg: Conversation) => {
    const senderType = msg.sender_type || (msg.role === 'user' ? 'student' : 'ai')
    return `student-chat__msg student-chat__msg--${senderType}`
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
        {messages.length === 0 && (
          <View className='student-chat__empty'>
            <Text className='student-chat__empty-text'>暂无对话记录</Text>
          </View>
        )}

        {messages.map((msg, index) => (
          <View key={msg.id || index} id={`msg-${index}`} className={getMsgClass(msg)}>
            {/* 发送者标签 */}
            <Text className='student-chat__sender'>{getSenderLabel(msg)}</Text>

            {/* 引用消息 */}
            {msg.reply_to_id && msg.reply_to_id > 0 && msg.reply_to_content && (
              <View className='student-chat__quote'>
                <Text className='student-chat__quote-text'>
                  引用: {msg.reply_to_content.length > 50
                    ? msg.reply_to_content.substring(0, 50) + '...'
                    : msg.reply_to_content}
                </Text>
              </View>
            )}

            {/* 消息内容 */}
            <Text className='student-chat__content'>{msg.content}</Text>

            {/* 时间和操作 */}
            <View className='student-chat__meta'>
              <Text className='student-chat__time'>
                {msg.created_at ? formatTime(msg.created_at) : ''}
              </Text>
              {/* 学生和 AI 消息可以引用回复 */}
              {(msg.sender_type === 'student' || msg.sender_type === 'ai' || msg.role === 'user') && (
                <View className='student-chat__reply-btn' onClick={() => handleQuoteReply(msg)}>
                  <Text className='student-chat__reply-btn-text'>引用回复</Text>
                </View>
              )}
            </View>
          </View>
        ))}

        <View className='student-chat__bottom-spacer' />
      </ScrollView>

      {/* 底部输入区域 */}
      <View className='student-chat__input-bar safe-area-bottom'>
        {/* 引用提示 */}
        {replyToMsg && (
          <View className='student-chat__quote-bar'>
            <Text className='student-chat__quote-bar-text'>
              引用: {replyToMsg.content.length > 30
                ? replyToMsg.content.substring(0, 30) + '...'
                : replyToMsg.content}
            </Text>
            <View className='student-chat__quote-cancel' onClick={handleCancelQuote}>
              <Text className='student-chat__quote-cancel-text'>✕</Text>
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
