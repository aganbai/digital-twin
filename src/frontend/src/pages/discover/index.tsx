import { useState, useCallback } from 'react'
import { View, Text, Input } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { getMarketplace } from '@/api/persona'
import type { MarketplacePersona } from '@/api/persona'
import { applyTeacher } from '@/api/relation'
import Empty from '@/components/Empty'
import './index.scss'

export default function Discover() {
  // 广场数据
  const [personas, setPersonas] = useState<MarketplacePersona[]>([])
  const [loading, setLoading] = useState(false)
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [applyingId, setApplyingId] = useState<number | null>(null)

  /** 获取广场数据 */
  const fetchMarketplace = useCallback(async (searchKeyword = '', pageNum = 1) => {
    setLoading(true)
    try {
      const res = await getMarketplace({
        keyword: searchKeyword || undefined,
        page: pageNum,
        page_size: 20,
      })
      if (pageNum === 1) {
        setPersonas(res.data.items || [])
      } else {
        setPersonas((prev) => [...prev, ...(res.data.items || [])])
      }
      setTotal(res.data.total || 0)
      setPage(pageNum)
    } catch (error) {
      console.error('获取广场数据失败:', error)
    } finally {
      setLoading(false)
    }
  }, [])

  /** 页面显示时刷新 */
  useDidShow(() => {
    fetchMarketplace(keyword, 1)
  })

  /** 搜索 */
  const handleSearch = () => {
    fetchMarketplace(keyword, 1)
  }

  /** 加载更多 */
  const handleLoadMore = () => {
    if (personas.length < total && !loading) {
      fetchMarketplace(keyword, page + 1)
    }
  }

  /** 申请使用教师分身 */
  const handleApply = async (persona: MarketplacePersona) => {
    if (applyingId || persona.application_status === 'pending') return
    setApplyingId(persona.id)
    try {
      await applyTeacher(persona.id, persona.id)
      setPersonas((prev) =>
        prev.map((p) =>
          p.id === persona.id ? { ...p, application_status: 'pending' as const } : p,
        ),
      )
      Taro.showToast({ title: '申请已发送', icon: 'success' })
    } catch (error) {
      console.error('申请失败:', error)
      Taro.showToast({ title: '申请失败', icon: 'none' })
    } finally {
      setApplyingId(null)
    }
  }

  return (
    <View className='discover-page'>
      <View className='discover-page__header'>
        <Text className='discover-page__title'>🌐 老师分身广场</Text>
        <Text className='discover-page__subtitle'>发现优秀的 AI 教师分身</Text>
      </View>

      {/* 搜索框 */}
      <View className='discover-page__search'>
        <Input
          className='discover-page__search-input'
          placeholder='搜索老师（昵称/学校）...'
          placeholderClass='discover-page__search-placeholder'
          value={keyword}
          onInput={(e) => setKeyword(e.detail.value)}
          onConfirm={handleSearch}
        />
        <View className='discover-page__search-btn' onClick={handleSearch}>
          <Text className='discover-page__search-btn-text'>搜索</Text>
        </View>
      </View>

      {/* 广场列表 */}
      {loading && personas.length === 0 ? (
        <View className='discover-page__loading'>
          <Text className='discover-page__loading-text'>加载中...</Text>
        </View>
      ) : personas.length > 0 ? (
        <View className='discover-page__list'>
          {personas.map((persona) => (
            <View key={persona.id} className='discover-page__card'>
              <View className='discover-page__card-info'>
                <Text className='discover-page__card-name'>
                  👨‍🏫 {persona.nickname}
                </Text>
                {persona.school && (
                  <Text className='discover-page__card-school'>
                    🏫 {persona.school}
                  </Text>
                )}
                {persona.description && (
                  <Text className='discover-page__card-desc'>
                    {persona.description}
                  </Text>
                )}
                <Text className='discover-page__card-stats'>
                  {persona.student_count} 名学生 · {persona.document_count} 篇文档
                </Text>
              </View>
              <View
                className={`discover-page__apply-btn ${
                  persona.application_status === 'pending'
                    ? 'discover-page__apply-btn--pending'
                    : ''
                }`}
                onClick={() => handleApply(persona)}
              >
                <Text className='discover-page__apply-text'>
                  {applyingId === persona.id
                    ? '申请中...'
                    : persona.application_status === 'pending'
                      ? '审核中'
                      : '申请使用'}
                </Text>
              </View>
            </View>
          ))}

          {/* 加载更多 */}
          {personas.length < total && (
            <View className='discover-page__loadmore' onClick={handleLoadMore}>
              <Text className='discover-page__loadmore-text'>
                {loading ? '加载中...' : '加载更多'}
              </Text>
            </View>
          )}
        </View>
      ) : (
        <View className='discover-page__empty'>
          <Empty text='暂无公开的教师分身' />
        </View>
      )}
    </View>
  )
}
