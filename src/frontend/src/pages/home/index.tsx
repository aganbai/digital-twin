import { useCallback, useEffect } from 'react'
import { View, Text, Input } from '@tarojs/components'
import Taro, { usePullDownRefresh } from '@tarojs/taro'
import { getTeachers } from '@/api/teacher'
import { useUserStore, useTeacherStore } from '@/store'
import TeacherCard from '@/components/TeacherCard'
import Empty from '@/components/Empty'
import './index.scss'

export default function Home() {
  const { userInfo } = useUserStore()
  const { teachers, loading, setTeachers, setLoading } = useTeacherStore()

  /** 获取教师列表 */
  const fetchTeachers = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getTeachers(1, 50)
      setTeachers(res.data.items || [])
    } catch (error) {
      console.error('获取教师列表失败:', error)
    } finally {
      setLoading(false)
    }
  }, [setTeachers, setLoading])

  /** 页面加载时获取教师列表 */
  useEffect(() => {
    fetchTeachers()
  }, [fetchTeachers])

  /** 下拉刷新 */
  usePullDownRefresh(async () => {
    await fetchTeachers()
    Taro.stopPullDownRefresh()
  })

  /** 点击"开始对话" → 跳转聊天页 */
  const handleChat = (teacherId: number) => {
    const teacher = teachers.find((t) => t.id === teacherId)
    const teacherName = teacher?.nickname || ''
    Taro.navigateTo({
      url: `/pages/chat/index?teacher_id=${teacherId}&teacher_name=${encodeURIComponent(teacherName)}`,
    })
  }

  return (
    <View className='home-page'>
      {/* 顶部问候语 */}
      <View className='home-page__header'>
        <Text className='home-page__greeting'>
          你好，{userInfo?.nickname || '同学'}
        </Text>
        <Text className='home-page__subtitle'>选择一位教师开始学习吧</Text>
      </View>

      {/* 搜索框（UI 占位，本迭代不实现搜索功能） */}
      <View className='home-page__search'>
        <Input
          className='home-page__search-input'
          placeholder='搜索教师...'
          disabled
        />
      </View>

      {/* 教师列表 / 空状态 / 加载状态 */}
      <View className='home-page__content'>
        {loading ? (
          <View className='home-page__loading'>
            <Text className='home-page__loading-text'>加载中...</Text>
          </View>
        ) : teachers.length > 0 ? (
          <View className='home-page__list'>
            {teachers.map((teacher) => (
              <TeacherCard
                key={teacher.id}
                id={teacher.id}
                nickname={teacher.nickname}
                username={teacher.username}
                documentCount={teacher.document_count}
                onChat={handleChat}
              />
            ))}
          </View>
        ) : (
          <Empty text='暂无教师' />
        )}
      </View>
    </View>
  )
}
