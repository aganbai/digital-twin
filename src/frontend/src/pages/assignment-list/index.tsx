import { useState, useCallback, useEffect } from 'react'
import { View, Text } from '@tarojs/components'
import Taro, { usePullDownRefresh } from '@tarojs/taro'
import { getAssignments, AssignmentListItem } from '@/api/assignment'
import { formatTime } from '@/utils/format'
import Empty from '@/components/Empty'
import './index.scss'

export default function AssignmentList() {
  const [assignments, setAssignments] = useState<AssignmentListItem[]>([])
  const [loading, setLoading] = useState(false)

  /** 获取作业列表 */
  const fetchAssignments = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getAssignments({ page: 1, page_size: 50 })
      setAssignments(res.data.items || [])
    } catch (error) {
      console.error('获取作业列表失败:', error)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchAssignments()
  }, [fetchAssignments])

  usePullDownRefresh(async () => {
    await fetchAssignments()
    Taro.stopPullDownRefresh()
  })

  /** 待点评 */
  const pendingList = assignments.filter((a) => a.status === 'submitted')
  /** 已点评 */
  const reviewedList = assignments.filter((a) => a.status === 'reviewed')

  /** 查看详情 */
  const handleViewDetail = (id: number) => {
    Taro.navigateTo({ url: `/pages/assignment-detail/index?id=${id}` })
  }

  return (
    <View className='assignment-list-page'>
      <View className='assignment-list-page__title'>
        <Text className='assignment-list-page__title-text'>学生作业</Text>
      </View>

      {loading ? (
        <View className='assignment-list-page__loading'>
          <Text>加载中...</Text>
        </View>
      ) : assignments.length === 0 ? (
        <Empty text='暂无作业' />
      ) : (
        <>
          {/* 待点评 */}
          <View className='assignment-list-page__section'>
            <Text className='assignment-list-page__section-title'>
              待点评 ({pendingList.length})
            </Text>
            {pendingList.map((item) => (
              <View
                key={item.id}
                className='assignment-list-page__card'
                onClick={() => handleViewDetail(item.id)}
              >
                <View className='assignment-list-page__card-info'>
                  <Text className='assignment-list-page__card-student'>
                    {item.student_nickname}
                  </Text>
                  <Text className='assignment-list-page__card-title'>{item.title}</Text>
                  <Text className='assignment-list-page__card-date'>
                    {formatTime(item.created_at)} 提交
                  </Text>
                </View>
                <View className='assignment-list-page__card-arrow'>
                  <Text className='assignment-list-page__card-arrow-text'>→</Text>
                </View>
              </View>
            ))}
            {pendingList.length === 0 && (
              <Text className='assignment-list-page__empty-text'>暂无待点评作业</Text>
            )}
          </View>

          {/* 已点评 */}
          <View className='assignment-list-page__section'>
            <Text className='assignment-list-page__section-title'>
              已点评 ({reviewedList.length})
            </Text>
            {reviewedList.map((item) => (
              <View
                key={item.id}
                className='assignment-list-page__card'
                onClick={() => handleViewDetail(item.id)}
              >
                <View className='assignment-list-page__card-info'>
                  <Text className='assignment-list-page__card-student'>
                    {item.student_nickname}
                  </Text>
                  <Text className='assignment-list-page__card-title'>{item.title}</Text>
                  <Text className='assignment-list-page__card-status'>已点评 ✅</Text>
                </View>
                <View className='assignment-list-page__card-arrow'>
                  <Text className='assignment-list-page__card-arrow-text'>→</Text>
                </View>
              </View>
            ))}
            {reviewedList.length === 0 && (
              <Text className='assignment-list-page__empty-text'>暂无已点评作业</Text>
            )}
          </View>
        </>
      )}
    </View>
  )
}
