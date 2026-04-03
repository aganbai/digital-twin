import { useState, useCallback } from 'react'
import { View, Text } from '@tarojs/components'
import Taro, { useDidShow, usePullDownRefresh } from '@tarojs/taro'
import { getSessions } from '@/api/chat'
import type { Session } from '@/api/chat'
import Empty from '@/components/Empty'
import { formatTime, truncateText } from '@/utils/format'
import './index.scss'

export default function History() {
  const [sessions, setSessions] = useState<Session[]>([])
  const [loading, setLoading] = useState(false)

  /** 获取会话列表 */
  const fetchSessions = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getSessions(1, 20)
      setSessions(res.data.items || [])
    } catch (error) {
      console.error('获取会话列表失败:', error)
    } finally {
      setLoading(false)
    }
  }, [])

  /** 每次页面显示时刷新数据 */
  useDidShow(() => {
    fetchSessions()
  })

  /** 下拉刷新 */
  usePullDownRefresh(async () => {
    await fetchSessions()
    Taro.stopPullDownRefresh()
  })

  /** 点击会话 → 跳转对话页 */
  const handleSessionClick = (session: Session) => {
    Taro.navigateTo({
      url: `/pages/chat/index?teacher_id=${session.teacher_id}&teacher_name=${session.teacher_nickname}&session_id=${session.session_id}`,
    })
  }

  /** 获取头像文字（取昵称首字） */
  const getAvatarText = (nickname: string) => {
    return nickname ? nickname.charAt(0) : '师'
  }

  return (
    <View className='history-page'>
      {/* 顶部标题区 */}
      <View className='history-page__header'>
        <Text className='history-page__title'>对话历史</Text>
      </View>

      {/* 会话列表 / 空状态 / 加载状态 */}
      <View className='history-page__content'>
        {loading ? (
          <View className='history-page__loading'>
            <Text className='history-page__loading-text'>加载中...</Text>
          </View>
        ) : sessions.length > 0 ? (
          <View className='history-page__list'>
            {sessions.map((session) => (
              <View
                key={session.session_id}
                className='history-page__item'
                onClick={() => handleSessionClick(session)}
              >
                {/* 左侧头像 */}
                <View className='history-page__avatar'>
                  <Text className='history-page__avatar-text'>
                    {getAvatarText(session.teacher_nickname)}
                  </Text>
                </View>

                {/* 中间信息 */}
                <View className='history-page__info'>
                  <Text className='history-page__nickname'>
                    {session.teacher_nickname}
                  </Text>
                  <Text className='history-page__message'>
                    {truncateText(session.last_message, 30)}
                  </Text>
                </View>

                {/* 右侧时间和消息数 */}
                <View className='history-page__meta'>
                  <Text className='history-page__time'>
                    {formatTime(session.updated_at)}
                  </Text>
                  <View className='history-page__count-badge'>
                    <Text className='history-page__count-text'>
                      {session.message_count}条
                    </Text>
                  </View>
                </View>
              </View>
            ))}
          </View>
        ) : (
          <Empty text='暂无对话记录' />
        )}
      </View>

    </View>
  )
}
