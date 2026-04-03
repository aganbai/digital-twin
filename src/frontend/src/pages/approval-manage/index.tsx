import { useState, useCallback } from 'react'
import { View, Text, Image } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { getPendingJoinRequests, JoinRequestItem } from '@/api/class'
import './index.scss'

export default function ApprovalManage() {
  const [requests, setRequests] = useState<JoinRequestItem[]>([])
  const [loading, setLoading] = useState(false)

  /** 加载待审批列表 */
  const fetchRequests = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getPendingJoinRequests()
      setRequests(res.data || [])
    } catch (e) {
      console.error('[ApprovalManage] 获取待审批列表失败:', e)
      Taro.showToast({ title: '加载失败', icon: 'none' })
    } finally {
      setLoading(false)
    }
  }, [])

  /** 页面显示时刷新 */
  useDidShow(() => {
    fetchRequests()
  })

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

  /** 跳转审批详情 */
  const handleGoDetail = (item: JoinRequestItem) => {
    Taro.navigateTo({
      url: `/pages/approval-detail/index?id=${item.id}&nickname=${encodeURIComponent(item.student_nickname)}&avatar=${encodeURIComponent(item.student_avatar || '')}`,
    })
  }

  return (
    <View className='approval-manage'>
      {/* 统计 */}
      <View className='approval-manage__count'>
        <Text className='approval-manage__count-text'>
          待审批 {requests.length} 人
        </Text>
      </View>

      {/* 列表 */}
      <View className='approval-manage__list'>
        {requests.length === 0 && !loading && (
          <View className='approval-manage__empty'>
            <Text className='approval-manage__empty-icon'>📋</Text>
            <Text className='approval-manage__empty-text'>暂无待审批申请</Text>
          </View>
        )}

        {requests.map((item) => (
          <View
            key={item.id}
            className='approval-manage__card'
            onClick={() => handleGoDetail(item)}
          >
            <View className='approval-manage__card-left'>
              {item.student_avatar ? (
                <Image
                  className='approval-manage__avatar'
                  src={item.student_avatar}
                  mode='aspectFill'
                />
              ) : (
                <View className='approval-manage__avatar approval-manage__avatar--default'>
                  <Text className='approval-manage__avatar-text'>
                    {item.student_nickname?.charAt(0) || '?'}
                  </Text>
                </View>
              )}
            </View>
            <View className='approval-manage__card-center'>
              <Text className='approval-manage__card-name'>
                {item.student_nickname}
              </Text>
              <Text className='approval-manage__card-time'>
                申请时间：{formatTime(item.request_time)}
              </Text>
            </View>
            <View className='approval-manage__card-right'>
              <Text className='approval-manage__card-arrow'>›</Text>
            </View>
          </View>
        ))}
      </View>

      {/* 加载状态 */}
      {loading && (
        <View className='approval-manage__loading'>
          <Text className='approval-manage__loading-text'>加载中...</Text>
        </View>
      )}
    </View>
  )
}
