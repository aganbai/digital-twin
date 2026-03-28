import { useState, useCallback } from 'react'
import { View, Text } from '@tarojs/components'
import Taro, { useDidShow, usePullDownRefresh } from '@tarojs/taro'
import { getDocuments, deleteDocument } from '@/api/knowledge'
import type { Document } from '@/api/knowledge'
import Empty from '@/components/Empty'
import { formatTime } from '@/utils/format'
import './index.scss'

export default function Knowledge() {
  const [documents, setDocuments] = useState<Document[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)

  /** 获取文档列表 */
  const fetchDocuments = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getDocuments(1, 20)
      setDocuments(res.data.items || [])
      setTotal(res.data.total || 0)
    } catch (error) {
      console.error('获取文档列表失败:', error)
    } finally {
      setLoading(false)
    }
  }, [])

  /** 每次页面显示时刷新列表（从添加页返回后也会触发） */
  useDidShow(() => {
    fetchDocuments()
  })

  /** 下拉刷新 */
  usePullDownRefresh(async () => {
    await fetchDocuments()
    Taro.stopPullDownRefresh()
  })

  /** 长按文档 → 删除确认 */
  const handleLongPress = (doc: Document) => {
    Taro.showModal({
      title: '删除确认',
      content: `确定要删除「${doc.title}」吗？`,
      confirmColor: '#FF4D4F',
      success: async (result) => {
        if (result.confirm) {
          try {
            await deleteDocument(doc.id)
            Taro.showToast({ title: '删除成功', icon: 'success' })
            fetchDocuments()
          } catch (error) {
            console.error('删除文档失败:', error)
          }
        }
      },
    })
  }

  /** 跳转添加文档页 */
  const handleAdd = () => {
    Taro.navigateTo({ url: '/pages/knowledge/add' })
  }

  return (
    <View className='knowledge-page'>
      {/* 顶部标题区 */}
      <View className='knowledge-page__header'>
        <Text className='knowledge-page__title'>我的知识库</Text>
        <Text className='knowledge-page__count'>共 {total} 篇文档</Text>
      </View>

      {/* 文档列表 / 空状态 / 加载状态 */}
      <View className='knowledge-page__content'>
        {loading ? (
          <View className='knowledge-page__loading'>
            <Text className='knowledge-page__loading-text'>加载中...</Text>
          </View>
        ) : documents.length > 0 ? (
          <View className='knowledge-page__list'>
            {documents.map((doc) => (
              <View
                key={doc.id}
                className='knowledge-page__item'
                onLongPress={() => handleLongPress(doc)}
              >
                <View className='knowledge-page__item-header'>
                  <Text className='knowledge-page__item-title'>{doc.title}</Text>
                </View>
                {doc.tags && doc.tags.length > 0 && (
                  <View className='knowledge-page__item-tags'>
                    {doc.tags.map((tag) => (
                      <Text key={tag} className='knowledge-page__item-tag'>{tag}</Text>
                    ))}
                  </View>
                )}
                <Text className='knowledge-page__item-time'>{formatTime(doc.created_at)}</Text>
              </View>
            ))}
          </View>
        ) : (
          <Empty text='还没有文档，点击右下角添加' />
        )}
      </View>

      {/* FAB 添加按钮 */}
      <View className='knowledge-page__fab' onClick={handleAdd}>
        <Text className='knowledge-page__fab-icon'>+</Text>
      </View>
    </View>
  )
}
