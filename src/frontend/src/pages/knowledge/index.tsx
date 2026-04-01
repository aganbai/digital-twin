import { useState, useCallback } from 'react'
import { View, Text } from '@tarojs/components'
import Taro, { useDidShow, usePullDownRefresh } from '@tarojs/taro'
import { getDocuments, deleteDocument } from '@/api/knowledge'
import type { Document } from '@/api/knowledge'
import { getClasses } from '@/api/class'
import type { ClassInfo } from '@/api/class'
import Empty from '@/components/Empty'
import { formatTime } from '@/utils/format'
import './index.scss'

/** scope 筛选类型 */
type ScopeFilter = 'all' | 'global' | 'class'

export default function Knowledge() {
  const [documents, setDocuments] = useState<Document[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [scopeFilter, setScopeFilter] = useState<ScopeFilter>('all')
  const [classes, setClasses] = useState<ClassInfo[]>([])
  const [selectedClassId, setSelectedClassId] = useState<number | undefined>(undefined)

  /** 获取文档列表 */
  const fetchDocuments = useCallback(async () => {
    setLoading(true)
    try {
      let scope: string | undefined
      let scopeId: number | undefined

      if (scopeFilter === 'global') {
        scope = 'global'
      } else if (scopeFilter === 'class' && selectedClassId) {
        scope = 'class'
        scopeId = selectedClassId
      }

      const res = await getDocuments(1, 20, scope, scopeId)
      setDocuments(res.data.items || [])
      setTotal(res.data.total || 0)
    } catch (error) {
      console.error('获取文档列表失败:', error)
    } finally {
      setLoading(false)
    }
  }, [scopeFilter, selectedClassId])

  /** 获取班级列表 */
  const fetchClasses = useCallback(async () => {
    try {
      const res = await getClasses()
      setClasses(res.data || [])
    } catch (error) {
      console.error('获取班级列表失败:', error)
    }
  }, [])

  /** 每次页面显示时刷新列表 */
  useDidShow(() => {
    fetchDocuments()
    fetchClasses()
  })

  /** 下拉刷新 */
  usePullDownRefresh(async () => {
    await fetchDocuments()
    Taro.stopPullDownRefresh()
  })

  /** 切换 scope 筛选 */
  const handleScopeChange = (scope: ScopeFilter) => {
    setScopeFilter(scope)
    if (scope !== 'class') {
      setSelectedClassId(undefined)
    }
    // 延迟触发刷新（等待状态更新）
    setTimeout(() => fetchDocuments(), 0)
  }

  /** 选择班级 */
  const handleClassSelect = (classId: number) => {
    setSelectedClassId(classId)
    setTimeout(() => fetchDocuments(), 0)
  }

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

  /** 跳转添加文档页（使用 redirectTo 避免 E2E 测试中页面栈溢出） */
  const handleAdd = () => {
    Taro.redirectTo({ url: '/pages/knowledge/add' })
  }

  return (
    <View className='knowledge-page'>
      {/* 顶部标题区 */}
      <View className='knowledge-page__header'>
        <Text className='knowledge-page__title'>我的知识库</Text>
        <Text className='knowledge-page__count'>共 {total} 篇文档</Text>
      </View>

      {/* Scope 筛选 */}
      <View className='knowledge-page__scope-tabs'>
        <View
          className={`knowledge-page__scope-tab ${scopeFilter === 'all' ? 'knowledge-page__scope-tab--active' : ''}`}
          onClick={() => handleScopeChange('all')}
        >
          <Text className={`knowledge-page__scope-tab-text ${scopeFilter === 'all' ? 'knowledge-page__scope-tab-text--active' : ''}`}>
            全部
          </Text>
        </View>
        <View
          className={`knowledge-page__scope-tab ${scopeFilter === 'global' ? 'knowledge-page__scope-tab--active' : ''}`}
          onClick={() => handleScopeChange('global')}
        >
          <Text className={`knowledge-page__scope-tab-text ${scopeFilter === 'global' ? 'knowledge-page__scope-tab-text--active' : ''}`}>
            全局
          </Text>
        </View>
        <View
          className={`knowledge-page__scope-tab ${scopeFilter === 'class' ? 'knowledge-page__scope-tab--active' : ''}`}
          onClick={() => handleScopeChange('class')}
        >
          <Text className={`knowledge-page__scope-tab-text ${scopeFilter === 'class' ? 'knowledge-page__scope-tab-text--active' : ''}`}>
            班级
          </Text>
        </View>
      </View>

      {/* 班级选择（仅在 scope=class 时显示） */}
      {scopeFilter === 'class' && classes.length > 0 && (
        <View className='knowledge-page__class-filter'>
          {classes.map((cls) => (
            <View
              key={cls.id}
              className={`knowledge-page__class-chip ${selectedClassId === cls.id ? 'knowledge-page__class-chip--active' : ''}`}
              onClick={() => handleClassSelect(cls.id)}
            >
              <Text className={`knowledge-page__class-chip-text ${selectedClassId === cls.id ? 'knowledge-page__class-chip-text--active' : ''}`}>
                {cls.name}
              </Text>
            </View>
          ))}
        </View>
      )}

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
