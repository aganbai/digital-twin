import { useState, useCallback, useEffect } from 'react'
import { View, Text } from '@tarojs/components'
import Taro, { usePullDownRefresh } from '@tarojs/taro'
import { getAssignments, AssignmentListItem } from '@/api/assignment'
import { formatTime } from '@/utils/format'
import Empty from '@/components/Empty'
import './index.scss'

export default function MyAssignments() {
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

  /** 查看详情 */
  const handleViewDetail = (id: number) => {
    Taro.navigateTo({ url: `/pages/assignment-detail/index?id=${id}` })
  }

  /** 提交新作业 */
  const handleSubmitNew = () => {
    Taro.navigateTo({ url: '/pages/submit-assignment/index' })
  }

  /** 获取点评分数显示 */
  const getScoreDisplay = (item: AssignmentListItem) => {
    if (item.status !== 'reviewed') return '待点评 ⏳'
    return `已点评 ✅ · ${item.review_count} 条点评`
  }

  return (
    <View className='my-assignments-page'>
      <View className='my-assignments-page__title'>
        <Text className='my-assignments-page__title-text'>我的作业</Text>
      </View>

      {loading ? (
        <View className='my-assignments-page__loading'>
          <Text>加载中...</Text>
        </View>
      ) : assignments.length === 0 ? (
        <Empty text='暂无作业' />
      ) : (
        <View className='my-assignments-page__list'>
          {assignments.map((item) => (
            <View
              key={item.id}
              className='my-assignments-page__card'
              onClick={() => handleViewDetail(item.id)}
            >
              <View className='my-assignments-page__card-info'>
                <Text className='my-assignments-page__card-title'>{item.title}</Text>
                <Text className='my-assignments-page__card-teacher'>
                  {item.teacher_nickname}
                </Text>
                <Text className='my-assignments-page__card-status'>
                  {getScoreDisplay(item)}
                </Text>
                <Text className='my-assignments-page__card-date'>
                  {formatTime(item.created_at)}
                </Text>
              </View>
              <View className='my-assignments-page__card-arrow'>
                <Text className='my-assignments-page__card-arrow-text'>→</Text>
              </View>
            </View>
          ))}
        </View>
      )}

      {/* 提交新作业按钮 */}
      <View className='my-assignments-page__fab' onClick={handleSubmitNew}>
        <Text className='my-assignments-page__fab-text'>+ 提交新作业</Text>
      </View>
    </View>
  )
}
