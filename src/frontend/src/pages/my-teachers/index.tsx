import { useState, useCallback, useEffect } from 'react'
import { View, Text } from '@tarojs/components'
import Taro, { usePullDownRefresh } from '@tarojs/taro'
import { getRelations, RelationItemStudent } from '@/api/relation'
import Empty from '@/components/Empty'
import './index.scss'

export default function MyTeachers() {
  const [relations, setRelations] = useState<RelationItemStudent[]>([])
  const [loading, setLoading] = useState(false)

  /** 获取关系列表 */
  const fetchRelations = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getRelations({ page: 1, page_size: 100 })
      setRelations((res.data.items || []) as RelationItemStudent[])
    } catch (error) {
      console.error('获取教师列表失败:', error)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchRelations()
  }, [fetchRelations])

  usePullDownRefresh(async () => {
    await fetchRelations()
    Taro.stopPullDownRefresh()
  })

  /** 已授权教师 */
  const approvedList = relations.filter((r) => r.status === 'approved')
  /** 审批中教师 */
  const pendingList = relations.filter((r) => r.status === 'pending')

  /** 进入对话 */
  const handleChat = (teacherId: number, teacherName: string) => {
    Taro.navigateTo({
      url: `/pages/chat/index?teacher_id=${teacherId}&teacher_name=${encodeURIComponent(teacherName)}`,
    })
  }

  return (
    <View className='my-teachers-page'>
      <View className='my-teachers-page__title'>
        <Text className='my-teachers-page__title-text'>我的教师</Text>
      </View>

      {loading ? (
        <View className='my-teachers-page__loading'>
          <Text>加载中...</Text>
        </View>
      ) : relations.length === 0 ? (
        <Empty text='暂无教师' />
      ) : (
        <>
          {/* 已授权 */}
          {approvedList.length > 0 && (
            <View className='my-teachers-page__section'>
              <Text className='my-teachers-page__section-title'>已授权</Text>
              {approvedList.map((item) => (
                <View key={item.id} className='my-teachers-page__card'>
                  <View className='my-teachers-page__card-avatar'>
                    <Text className='my-teachers-page__card-avatar-text'>
                      {item.teacher_nickname?.charAt(0) || 'T'}
                    </Text>
                  </View>
                  <View className='my-teachers-page__card-info'>
                    <Text className='my-teachers-page__card-name'>{item.teacher_nickname}</Text>
                    <Text className='my-teachers-page__card-school'>{item.teacher_school}</Text>
                    {item.teacher_description && (
                      <Text className='my-teachers-page__card-desc'>{item.teacher_description}</Text>
                    )}
                  </View>
                  <View
                    className='my-teachers-page__card-btn'
                    onClick={() => handleChat(item.teacher_id, item.teacher_nickname)}
                  >
                    <Text className='my-teachers-page__card-btn-text'>进入对话</Text>
                  </View>
                </View>
              ))}
            </View>
          )}

          {/* 审批中 */}
          {pendingList.length > 0 && (
            <View className='my-teachers-page__section'>
              <Text className='my-teachers-page__section-title'>审批中</Text>
              {pendingList.map((item) => (
                <View key={item.id} className='my-teachers-page__card'>
                  <View className='my-teachers-page__card-avatar'>
                    <Text className='my-teachers-page__card-avatar-text'>
                      {item.teacher_nickname?.charAt(0) || 'T'}
                    </Text>
                  </View>
                  <View className='my-teachers-page__card-info'>
                    <Text className='my-teachers-page__card-name'>{item.teacher_nickname}</Text>
                    <Text className='my-teachers-page__card-school'>{item.teacher_school}</Text>
                  </View>
                  <View className='my-teachers-page__card-status'>
                    <Text className='my-teachers-page__card-status-text'>等待审批中 ⏳</Text>
                  </View>
                </View>
              ))}
            </View>
          )}
        </>
      )}
    </View>
  )
}
