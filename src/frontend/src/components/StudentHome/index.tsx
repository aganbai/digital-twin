import { useCallback, useEffect, useState, useRef } from 'react'
import { View, Text } from '@tarojs/components'
import Taro, { usePullDownRefresh, useDidShow } from '@tarojs/taro'
import { getTeachers, Teacher } from '@/api/teacher'
import { useTeacherStore } from '@/store'
import Empty from '@/components/Empty'
import './index.scss'

export default function StudentHome() {
  const { teachers, loading: teacherLoading, setTeachers, setLoading: setTeacherLoading } = useTeacherStore()
  const autoRedirected = useRef(false)

  /** 获取教师列表（学生视角） */
  const fetchTeachers = useCallback(async () => {
    setTeacherLoading(true)
    try {
      const res = await getTeachers(1, 50)
      const items = res.data.items || []
      setTeachers(items)
      return items
    } catch (error) {
      console.error('获取教师列表失败:', error)
      return []
    } finally {
      setTeacherLoading(false)
    }
  }, [setTeachers, setTeacherLoading])

  /** 页面加载时获取数据并判断跳转逻辑 */
  useEffect(() => {
    const init = async () => {
      const items = await fetchTeachers()
      // R3: 学生只有1个已授权老师时，直接跳转对话页
      if (items.length === 1 && !autoRedirected.current) {
        autoRedirected.current = true
        const teacher = items[0]
        Taro.redirectTo({
          url: `/pages/chat/index?teacher_id=${teacher.id}&teacher_name=${encodeURIComponent(teacher.nickname)}`,
        })
      }
    }
    init()
  }, [fetchTeachers])

  /** 页面显示时刷新数据 */
  useDidShow(() => {
    // 如果已经自动跳转过，重新进入首页时重置标记
    autoRedirected.current = false
    fetchTeachers()
  })

  /** 下拉刷新 */
  usePullDownRefresh(async () => {
    await fetchTeachers()
    Taro.stopPullDownRefresh()
  })

  /** 跳转快捷操作 */
  const handleQuickAction = (path: string) => {
    Taro.navigateTo({ url: path })
  }

  /** 学生点击开始对话 */
  const handleChat = (teacher: Teacher) => {
    Taro.navigateTo({
      url: `/pages/chat/index?teacher_id=${teacher.id}&teacher_name=${encodeURIComponent(teacher.nickname)}`,
    })
  }

  // 0个老师 → 引导页
  if (!teacherLoading && teachers.length === 0) {
    return (
      <View className='student-home'>
        <View className='student-home__guide'>
          <View className='student-home__guide-icon'>
            <Text className='student-home__guide-icon-text'>🎓</Text>
          </View>
          <Text className='student-home__guide-title'>还没有老师</Text>
          <Text className='student-home__guide-desc'>
            输入老师的分享码加入班级，或去发现页找到心仪的老师
          </Text>
          <View className='student-home__guide-actions'>
            <View
              className='student-home__guide-btn student-home__guide-btn--primary'
              onClick={() => handleQuickAction('/pages/share-join/index')}
            >
              <Text className='student-home__guide-btn-text'>🔗 输入分享码</Text>
            </View>
            <View
              className='student-home__guide-btn student-home__guide-btn--secondary'
              onClick={() => handleQuickAction('/pages/discover/index')}
            >
              <Text className='student-home__guide-btn-text--secondary'>🌐 去发现页</Text>
            </View>
          </View>
        </View>
      </View>
    )
  }

  return (
    <View className='student-home'>
      {/* 快捷操作 */}
      <View className='student-home__card student-home__actions-card'>
        <View className='student-home__actions'>
          <View
            className='student-home__action-item'
            onClick={() => handleQuickAction('/pages/my-assignments/index')}
          >
            <Text className='student-home__action-icon'>📝</Text>
            <Text className='student-home__action-label'>我的作业</Text>
          </View>
          {/* M2: 移除"我的评语"入口，评语改为教师私有备注，学生不可见 */}
          <View
            className='student-home__action-item'
            onClick={() => handleQuickAction('/pages/discover/index')}
          >
            <Text className='student-home__action-icon'>🌐</Text>
            <Text className='student-home__action-label'>发现</Text>
          </View>
        </View>
      </View>

      {/* 分享码加入入口 */}
      <View className='student-home__card student-home__join-card'>
        <Text className='student-home__card-title'>加入班级</Text>
        <View className='student-home__join-row'>
          <View
            className='student-home__join-btn'
            onClick={() => handleQuickAction('/pages/share-join/index')}
          >
            <Text className='student-home__join-btn-text'>输入分享码加入</Text>
          </View>
        </View>
      </View>

      {/* 我的老师列表 */}
      <View className='student-home__card'>
        <Text className='student-home__card-title'>我的老师</Text>
        {teacherLoading ? (
          <View className='student-home__loading'>
            <Text className='student-home__loading-text'>加载中...</Text>
          </View>
        ) : teachers.length > 0 ? (
          <View className='student-home__teacher-list'>
            {teachers.map((teacher) => (
              <View key={teacher.id} className='student-home__teacher-item'>
                <View className='student-home__teacher-info'>
                  <Text className='student-home__teacher-name'>{teacher.nickname}</Text>
                  {teacher.school && (
                    <Text className='student-home__teacher-school'>{teacher.school}</Text>
                  )}
                  {teacher.description && (
                    <Text className='student-home__teacher-desc'>{teacher.description}</Text>
                  )}
                </View>
                <View
                  className='student-home__teacher-chat-btn'
                  onClick={() => handleChat(teacher)}
                >
                  <Text className='student-home__teacher-chat-text'>开始对话</Text>
                </View>
              </View>
            ))}
          </View>
        ) : (
          <Empty text='暂无已授权教师' />
        )}
      </View>
    </View>
  )
}
