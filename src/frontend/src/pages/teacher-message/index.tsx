import { useState, useEffect } from 'react'
import { View, Text, Textarea, Picker, Input } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { pushTeacherMessage, getTeacherMessageHistory, TeacherMessage } from '@/api/teacher-message'
import { getClasses } from '@/api/class'
import { useUserStore } from '@/store'
import { formatTime } from '@/utils/format'
import './index.scss'

export default function TeacherMessagePage() {
  const { userInfo } = useUserStore()
  const [targetType, setTargetType] = useState<'class' | 'student'>('class')
  const [targetId, setTargetId] = useState(0)
  const [content, setContent] = useState('')
  const [sending, setSending] = useState(false)
  const [history, setHistory] = useState<TeacherMessage[]>([])
  const [classes, setClasses] = useState<any[]>([])
  const [selectedClassName, setSelectedClassName] = useState('')

  // 加载班级列表
  useEffect(() => {
    loadClasses()
    loadHistory()
  }, [])

  const loadClasses = async () => {
    try {
      const res = await getClasses()
      setClasses(res.data || [])
    } catch (e) {
      console.error('加载班级列表失败:', e)
    }
  }

  const loadHistory = async () => {
    try {
      const res = await getTeacherMessageHistory(1, 20)
      setHistory(res.data.items || [])
    } catch (e) {
      console.error('加载推送历史失败:', e)
    }
  }

  const handleSend = async () => {
    if (!content.trim()) {
      Taro.showToast({ title: '请输入消息内容', icon: 'none' })
      return
    }
    if (!targetId) {
      Taro.showToast({ title: '请选择推送目标', icon: 'none' })
      return
    }
    setSending(true)
    try {
      const res = await pushTeacherMessage({
        target_type: targetType,
        target_id: targetId,
        content: content.trim(),
        persona_id: (userInfo as any)?.default_persona_id || 0,
      })
      Taro.showToast({ title: `推送成功，送达${res.data.success_count}人`, icon: 'success' })
      setContent('')
      loadHistory()
    } catch (e: any) {
      Taro.showToast({ title: e?.message || '推送失败', icon: 'none' })
    } finally {
      setSending(false)
    }
  }

  return (
    <View className='teacher-msg-page'>
      <View className='teacher-msg-page__header'>
        <Text className='teacher-msg-page__title'>📢 消息推送</Text>
        <Text className='teacher-msg-page__subtitle'>向班级或指定学生推送消息</Text>
      </View>

      {/* 推送表单 */}
      <View className='teacher-msg-page__form'>
        {/* 目标类型选择 */}
        <View className='teacher-msg-page__field'>
          <Text className='teacher-msg-page__label'>推送目标</Text>
          <View className='teacher-msg-page__type-tabs'>
            <View
              className={`teacher-msg-page__type-tab ${targetType === 'class' ? 'teacher-msg-page__type-tab--active' : ''}`}
              onClick={() => { setTargetType('class'); setTargetId(0); setSelectedClassName('') }}
            >
              <Text className='teacher-msg-page__type-tab-text'>班级</Text>
            </View>
            <View
              className={`teacher-msg-page__type-tab ${targetType === 'student' ? 'teacher-msg-page__type-tab--active' : ''}`}
              onClick={() => { setTargetType('student'); setTargetId(0) }}
            >
              <Text className='teacher-msg-page__type-tab-text'>指定学生</Text>
            </View>
          </View>
        </View>

        {/* 班级选择 */}
        {targetType === 'class' && (
          <View className='teacher-msg-page__field'>
            <Text className='teacher-msg-page__label'>选择班级</Text>
            <Picker
              mode='selector'
              range={classes.map(c => c.name)}
              onChange={(e) => {
                const idx = Number(e.detail.value)
                setTargetId(classes[idx]?.id || 0)
                setSelectedClassName(classes[idx]?.name || '')
              }}
            >
              <View className='teacher-msg-page__picker'>
                <Text className='teacher-msg-page__picker-text'>
                  {selectedClassName || '请选择班级'}
                </Text>
                <Text className='teacher-msg-page__picker-arrow'>▸</Text>
              </View>
            </Picker>
          </View>
        )}

        {/* 学生ID输入 */}
        {targetType === 'student' && (
          <View className='teacher-msg-page__field'>
            <Text className='teacher-msg-page__label'>学生分身ID</Text>
            <View className='teacher-msg-page__input-wrap'>
              <Input
                className='teacher-msg-page__input'
                type='number'
                placeholder='输入学生分身ID'
                value={String(targetId || '')}
                onInput={(e) => setTargetId(Number(e.detail.value) || 0)}
              />
            </View>
          </View>
        )}

        {/* 消息内容 */}
        <View className='teacher-msg-page__field'>
          <Text className='teacher-msg-page__label'>消息内容</Text>
          <Textarea
            className='teacher-msg-page__textarea'
            value={content}
            placeholder='输入要推送的消息内容（最多1000字）'
            maxlength={1000}
            onInput={(e) => setContent(e.detail.value)}
          />
          <Text className='teacher-msg-page__char-count'>{content.length}/1000</Text>
        </View>

        {/* 发送按钮 */}
        <View
          className={`teacher-msg-page__send-btn ${(!content.trim() || !targetId || sending) ? 'teacher-msg-page__send-btn--disabled' : ''}`}
          onClick={handleSend}
        >
          <Text className='teacher-msg-page__send-btn-text'>
            {sending ? '推送中...' : '发送推送'}
          </Text>
        </View>
      </View>

      {/* 推送历史 */}
      <View className='teacher-msg-page__history'>
        <Text className='teacher-msg-page__history-title'>推送历史</Text>
        {history.length === 0 ? (
          <Text className='teacher-msg-page__history-empty'>暂无推送记录</Text>
        ) : (
          history.map((item) => (
            <View key={item.id} className='teacher-msg-page__history-item'>
              <View className='teacher-msg-page__history-header'>
                <Text className='teacher-msg-page__history-target'>
                  {item.target_type === 'class' ? '📚 班级' : '👤 学生'} #{item.target_id}
                </Text>
                <Text className='teacher-msg-page__history-time'>{formatTime(item.created_at)}</Text>
              </View>
              <Text className='teacher-msg-page__history-content'>{item.content}</Text>
              <Text className={`teacher-msg-page__history-status teacher-msg-page__history-status--${item.status}`}>
                {item.status === 'sent' ? '✅ 已送达' : item.status === 'failed' ? '❌ 失败' : '⏳ 待发送'}
              </Text>
            </View>
          ))
        )}
      </View>
    </View>
  )
}
