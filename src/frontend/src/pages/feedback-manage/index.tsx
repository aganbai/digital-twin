import { useState, useCallback } from 'react'
import { View, Text } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import {
  getFeedbacks,
  updateFeedbackStatus,
  FEEDBACK_TYPES,
  FEEDBACK_STATUSES,
  FeedbackItem,
  FeedbackStatus,
} from '@/api/feedback'
import './index.scss'

/** 状态标签颜色映射 */
const STATUS_COLORS: Record<FeedbackStatus, string> = {
  pending: '#f59e0b',
  reviewed: '#3b82f6',
  resolved: '#10b981',
}

/** 状态中文映射 */
const STATUS_LABELS: Record<FeedbackStatus, string> = {
  pending: '待处理',
  reviewed: '已查看',
  resolved: '已解决',
}

export default function FeedbackManagePage() {
  const [feedbacks, setFeedbacks] = useState<FeedbackItem[]>([])
  const [loading, setLoading] = useState(false)
  const [currentStatus, setCurrentStatus] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const pageSize = 10

  /** 加载反馈列表 */
  const fetchFeedbacks = useCallback(async (status: string, pageNum: number) => {
    setLoading(true)
    try {
      const params: Record<string, any> = { page: pageNum, page_size: pageSize }
      if (status) params.status = status
      const res = await getFeedbacks(params)
      if (pageNum === 1) {
        setFeedbacks(res.data.items || [])
      } else {
        setFeedbacks(prev => [...prev, ...(res.data.items || [])])
      }
      setTotal(res.data.total || 0)
    } catch (e) {
      console.error('[FeedbackManage] 获取反馈列表失败:', e)
      Taro.showToast({ title: '加载失败', icon: 'none' })
    } finally {
      setLoading(false)
    }
  }, [])

  /** 页面显示时刷新 */
  useDidShow(() => {
    setPage(1)
    fetchFeedbacks(currentStatus, 1)
  })

  /** 切换状态筛选 */
  const handleStatusFilter = (status: string) => {
    setCurrentStatus(status)
    setPage(1)
    fetchFeedbacks(status, 1)
  }

  /** 加载更多 */
  const handleLoadMore = () => {
    if (feedbacks.length >= total || loading) return
    const nextPage = page + 1
    setPage(nextPage)
    fetchFeedbacks(currentStatus, nextPage)
  }

  /** 更新反馈状态 */
  const handleUpdateStatus = (item: FeedbackItem) => {
    const nextStatusMap: Record<FeedbackStatus, FeedbackStatus> = {
      pending: 'reviewed',
      reviewed: 'resolved',
      resolved: 'resolved',
    }
    const nextStatus = nextStatusMap[item.status]
    if (nextStatus === item.status) {
      Taro.showToast({ title: '该反馈已解决', icon: 'none' })
      return
    }

    Taro.showModal({
      title: '更新状态',
      content: `确定将此反馈标记为"${STATUS_LABELS[nextStatus]}"吗？`,
      confirmText: '确定',
      success: async (res) => {
        if (res.confirm) {
          try {
            await updateFeedbackStatus(item.id, nextStatus)
            Taro.showToast({ title: '更新成功', icon: 'success' })
            // 局部更新列表
            setFeedbacks(prev =>
              prev.map(f => f.id === item.id ? { ...f, status: nextStatus } : f)
            )
          } catch (e) {
            Taro.showToast({ title: '更新失败', icon: 'none' })
          }
        }
      },
    })
  }

  /** 获取反馈类型标签 */
  const getTypeLabel = (type: string) => {
    return FEEDBACK_TYPES.find(t => t.value === type)?.label || type
  }

  /** 格式化时间 */
  const formatTime = (time: string) => {
    if (!time) return ''
    const d = new Date(time)
    const month = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    const hour = String(d.getHours()).padStart(2, '0')
    const min = String(d.getMinutes()).padStart(2, '0')
    return `${month}-${day} ${hour}:${min}`
  }

  return (
    <View className='fb-manage'>
      {/* 状态筛选 */}
      <View className='fb-manage__filters'>
        {FEEDBACK_STATUSES.map(s => (
          <View
            key={s.value}
            className={`fb-manage__filter-tag ${currentStatus === s.value ? 'fb-manage__filter-tag--active' : ''}`}
            onClick={() => handleStatusFilter(s.value)}
          >
            <Text className='fb-manage__filter-text'>{s.label}</Text>
          </View>
        ))}
      </View>

      {/* 统计信息 */}
      <View className='fb-manage__count'>
        <Text className='fb-manage__count-text'>共 {total} 条反馈</Text>
      </View>

      {/* 反馈列表 */}
      <View className='fb-manage__list'>
        {feedbacks.length === 0 && !loading && (
          <View className='fb-manage__empty'>
            <Text className='fb-manage__empty-text'>暂无反馈数据</Text>
          </View>
        )}

        {feedbacks.map(item => (
          <View key={item.id} className='fb-manage__card' onClick={() => handleUpdateStatus(item)}>
            <View className='fb-manage__card-header'>
              <Text className='fb-manage__card-type'>{getTypeLabel(item.feedback_type)}</Text>
              <View
                className='fb-manage__card-status'
                style={{ backgroundColor: STATUS_COLORS[item.status] + '20', color: STATUS_COLORS[item.status] }}
              >
                <Text className='fb-manage__card-status-text' style={{ color: STATUS_COLORS[item.status] }}>
                  {STATUS_LABELS[item.status]}
                </Text>
              </View>
            </View>
            <Text className='fb-manage__card-content'>{item.content}</Text>
            <View className='fb-manage__card-footer'>
              <Text className='fb-manage__card-user'>{item.user_nickname || '匿名用户'}</Text>
              <Text className='fb-manage__card-time'>{formatTime(item.created_at)}</Text>
            </View>
            {item.status !== 'resolved' && (
              <View className='fb-manage__card-action'>
                <Text className='fb-manage__card-action-text'>
                  点击标记为"{STATUS_LABELS[item.status === 'pending' ? 'reviewed' : 'resolved']}"
                </Text>
              </View>
            )}
          </View>
        ))}

        {/* 加载更多 */}
        {feedbacks.length > 0 && feedbacks.length < total && (
          <View className='fb-manage__load-more' onClick={handleLoadMore}>
            <Text className='fb-manage__load-more-text'>
              {loading ? '加载中...' : '加载更多'}
            </Text>
          </View>
        )}

        {feedbacks.length > 0 && feedbacks.length >= total && (
          <View className='fb-manage__no-more'>
            <Text className='fb-manage__no-more-text'>— 没有更多了 —</Text>
          </View>
        )}
      </View>
    </View>
  )
}
