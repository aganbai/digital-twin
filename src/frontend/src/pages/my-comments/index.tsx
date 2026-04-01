import { useState, useCallback, useEffect } from 'react'
import { View, Text } from '@tarojs/components'
import Taro, { usePullDownRefresh } from '@tarojs/taro'
import { getComments, CommentItem } from '@/api/comment'
import { formatTime } from '@/utils/format'
import Empty from '@/components/Empty'
import './index.scss'

export default function MyComments() {
  const [comments, setComments] = useState<CommentItem[]>([])
  const [loading, setLoading] = useState(false)

  /** 获取评语列表 */
  const fetchComments = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getComments({ page: 1, page_size: 50 })
      setComments(res.data.items || [])
    } catch (error) {
      console.error('获取评语失败:', error)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchComments()
  }, [fetchComments])

  usePullDownRefresh(async () => {
    await fetchComments()
    Taro.stopPullDownRefresh()
  })

  return (
    <View className='my-comments-page'>
      <View className='my-comments-page__title'>
        <Text className='my-comments-page__title-text'>我的评语</Text>
      </View>

      {loading ? (
        <View className='my-comments-page__loading'>
          <Text>加载中...</Text>
        </View>
      ) : comments.length === 0 ? (
        <Empty text='暂无评语' />
      ) : (
        <View className='my-comments-page__list'>
          {comments.map((item) => (
            <View key={item.id} className='my-comments-page__card'>
              <View className='my-comments-page__card-header'>
                <Text className='my-comments-page__card-teacher'>
                  {item.teacher_nickname}
                </Text>
                <Text className='my-comments-page__card-date'>
                  {formatTime(item.created_at)}
                </Text>
              </View>
              <Text className='my-comments-page__card-content'>{item.content}</Text>
              {item.progress_summary && (
                <View className='my-comments-page__card-progress'>
                  <Text className='my-comments-page__card-progress-text'>
                    进度：{item.progress_summary}
                  </Text>
                </View>
              )}
            </View>
          ))}
        </View>
      )}
    </View>
  )
}
