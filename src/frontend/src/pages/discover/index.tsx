import { useState, useCallback } from 'react'
import { View, Text, Input, ScrollView } from '@tarojs/components'
import Taro, { useDidShow, usePullDownRefresh } from '@tarojs/taro'
import { getDiscoverList, searchDiscover } from '@/api/discover'
import type {
  HotClass,
  RecommendedTeacher,
  SubjectItem,
} from '@/api/discover'
import { getMarketplace } from '@/api/persona'
import type { MarketplacePersona } from '@/api/persona'
import { applyTeacher } from '@/api/relation'
import Empty from '@/components/Empty'
import './index.scss'

export default function Discover() {
  // 推荐数据
  const [hotClasses, setHotClasses] = useState<HotClass[]>([])
  const [recommendedTeachers, setRecommendedTeachers] = useState<RecommendedTeacher[]>([])
  const [subjects, setSubjects] = useState<SubjectItem[]>([])
  const [loading, setLoading] = useState(false)

  // 搜索相关
  const [keyword, setKeyword] = useState('')
  const [searchMode, setSearchMode] = useState(false)
  const [searchClasses, setSearchClasses] = useState<HotClass[]>([])
  const [searchTeachers, setSearchTeachers] = useState<RecommendedTeacher[]>([])
  const [searching, setSearching] = useState(false)

  // 广场数据（兼容旧逻辑）
  const [personas, setPersonas] = useState<MarketplacePersona[]>([])
  const [applyingId, setApplyingId] = useState<number | null>(null)

  // 选中的学科
  const [selectedSubject, setSelectedSubject] = useState<string>('')

  /** 获取发现页推荐数据 */
  const fetchDiscoverData = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getDiscoverList()
      setHotClasses(res.data.hot_classes || [])
      setRecommendedTeachers(res.data.recommended_teachers || [])
      setSubjects(res.data.subjects || [])
    } catch (error) {
      console.error('获取发现页数据失败:', error)
      // 如果新接口失败，回退到旧的广场接口
      try {
        const fallbackRes = await getMarketplace({ page: 1, page_size: 20 })
        setPersonas(fallbackRes.data.items || [])
      } catch (e) {
        console.error('回退广场接口也失败:', e)
      }
    } finally {
      setLoading(false)
    }
  }, [])

  /** 页面显示时刷新 */
  useDidShow(() => {
    fetchDiscoverData()
  })

  /** 下拉刷新 */
  usePullDownRefresh(async () => {
    if (searchMode) {
      if (keyword.trim()) {
        await handleSearch()
      }
    } else {
      await fetchDiscoverData()
    }
    Taro.stopPullDownRefresh()
  })

  /** 搜索 */
  const handleSearch = async () => {
    const trimmed = keyword.trim()
    if (!trimmed) {
      setSearchMode(false)
      return
    }

    setSearching(true)
    setSearchMode(true)
    try {
      const res = await searchDiscover(trimmed)
      setSearchClasses(res.data.classes || [])
      setSearchTeachers(res.data.teachers || [])
    } catch (error) {
      console.error('搜索失败:', error)
      // 回退到广场搜索
      try {
        const fallbackRes = await getMarketplace({ keyword: trimmed, page: 1, page_size: 20 })
        setPersonas(fallbackRes.data.items || [])
      } catch (e) {
        console.error('回退搜索也失败:', e)
      }
    } finally {
      setSearching(false)
    }
  }

  /** 清除搜索 */
  const handleClearSearch = () => {
    setKeyword('')
    setSearchMode(false)
    setSearchClasses([])
    setSearchTeachers([])
  }

  /** 点击学科筛选 */
  const handleSubjectClick = (subjectName: string) => {
    if (selectedSubject === subjectName) {
      setSelectedSubject('')
    } else {
      setSelectedSubject(subjectName)
    }
  }

  /** 查看班级详情 */
  const handleClassDetail = (classItem: HotClass) => {
    Taro.navigateTo({
      url: `/pages/class-detail/index?id=${classItem.id}`,
    })
  }

  /** 申请使用教师分身 */
  const handleApply = async (teacher: RecommendedTeacher) => {
    if (applyingId) return
    setApplyingId(teacher.id)
    try {
      await applyTeacher(teacher.id, teacher.id)
      Taro.showToast({ title: '申请已发送', icon: 'success' })
    } catch (error) {
      console.error('申请失败:', error)
      Taro.showToast({ title: '申请失败', icon: 'none' })
    } finally {
      setApplyingId(null)
    }
  }

  /** 申请使用旧广场分身 */
  const handleApplyPersona = async (persona: MarketplacePersona) => {
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

  /** 按学科过滤热门班级 */
  const filteredClasses = selectedSubject
    ? hotClasses.filter((c) => c.subject === selectedSubject)
    : hotClasses

  return (
    <View className='discover-page'>
      {/* 顶部标题区 */}
      <View className='discover-page__header'>
        <Text className='discover-page__title'>🌐 发现</Text>
        <Text className='discover-page__subtitle'>探索优秀的班级和老师</Text>
      </View>

      {/* 搜索框 */}
      <View className='discover-page__search'>
        <Input
          className='discover-page__search-input'
          placeholder='搜索班级或老师...'
          placeholderClass='discover-page__search-placeholder'
          value={keyword}
          onInput={(e) => setKeyword(e.detail.value)}
          onConfirm={handleSearch}
        />
        {searchMode ? (
          <View className='discover-page__search-btn discover-page__search-btn--cancel' onClick={handleClearSearch}>
            <Text className='discover-page__search-btn-text'>取消</Text>
          </View>
        ) : (
          <View className='discover-page__search-btn' onClick={handleSearch}>
            <Text className='discover-page__search-btn-text'>搜索</Text>
          </View>
        )}
      </View>

      {loading ? (
        <View className='discover-page__loading'>
          <Text className='discover-page__loading-text'>加载中...</Text>
        </View>
      ) : searchMode ? (
        /* ========== 搜索结果模式 ========== */
        <View className='discover-page__search-results'>
          {searching ? (
            <View className='discover-page__loading'>
              <Text className='discover-page__loading-text'>搜索中...</Text>
            </View>
          ) : (
            <>
              {/* 搜索到的班级 */}
              {searchClasses.length > 0 && (
                <View className='discover-page__section'>
                  <Text className='discover-page__section-title'>📚 班级</Text>
                  {searchClasses.map((cls) => (
                    <View key={cls.id} className='discover-page__class-card' onClick={() => handleClassDetail(cls)}>
                      <View className='discover-page__class-card-info'>
                        <Text className='discover-page__class-card-name'>{cls.name}</Text>
                        <Text className='discover-page__class-card-teacher'>👨‍🏫 {cls.teacher_name}</Text>
                        {cls.description && (
                          <Text className='discover-page__class-card-desc'>{cls.description}</Text>
                        )}
                        <View className='discover-page__class-card-meta'>
                          <Text className='discover-page__class-card-subject'>{cls.subject}</Text>
                          <Text className='discover-page__class-card-count'>{cls.member_count} 名成员</Text>
                        </View>
                      </View>
                    </View>
                  ))}
                </View>
              )}

              {/* 搜索到的老师 */}
              {searchTeachers.length > 0 && (
                <View className='discover-page__section'>
                  <Text className='discover-page__section-title'>👨‍🏫 老师</Text>
                  {searchTeachers.map((teacher) => (
                    <View key={teacher.id} className='discover-page__teacher-card'>
                      <View className='discover-page__teacher-card-info'>
                        <Text className='discover-page__teacher-card-name'>{teacher.nickname}</Text>
                        {teacher.school && (
                          <Text className='discover-page__teacher-card-school'>🏫 {teacher.school}</Text>
                        )}
                        {teacher.description && (
                          <Text className='discover-page__teacher-card-desc'>{teacher.description}</Text>
                        )}
                        <View className='discover-page__teacher-card-meta'>
                          <Text className='discover-page__teacher-card-subject'>{teacher.subject}</Text>
                          <Text className='discover-page__teacher-card-stats'>
                            {teacher.student_count} 名学生 · {teacher.document_count} 篇文档
                          </Text>
                        </View>
                      </View>
                      <View className='discover-page__apply-btn' onClick={() => handleApply(teacher)}>
                        <Text className='discover-page__apply-text'>
                          {applyingId === teacher.id ? '申请中...' : '申请使用'}
                        </Text>
                      </View>
                    </View>
                  ))}
                </View>
              )}

              {searchClasses.length === 0 && searchTeachers.length === 0 && (
                <View className='discover-page__empty'>
                  <Empty text='未找到相关结果' />
                </View>
              )}
            </>
          )}
        </View>
      ) : (
        /* ========== 推荐模式 ========== */
        <View className='discover-page__recommend'>
          {/* 学科浏览 */}
          {subjects.length > 0 && (
            <View className='discover-page__section'>
              <Text className='discover-page__section-title'>📖 按学科浏览</Text>
              <ScrollView scrollX className='discover-page__subjects-scroll'>
                <View className='discover-page__subjects'>
                  {subjects.map((subject) => (
                    <View
                      key={subject.name}
                      className={`discover-page__subject-tag ${selectedSubject === subject.name ? 'discover-page__subject-tag--active' : ''}`}
                      onClick={() => handleSubjectClick(subject.name)}
                    >
                      <Text className='discover-page__subject-icon'>{subject.icon}</Text>
                      <Text className={`discover-page__subject-name ${selectedSubject === subject.name ? 'discover-page__subject-name--active' : ''}`}>
                        {subject.name}
                      </Text>
                      <Text className={`discover-page__subject-count ${selectedSubject === subject.name ? 'discover-page__subject-count--active' : ''}`}>
                        {subject.count}
                      </Text>
                    </View>
                  ))}
                </View>
              </ScrollView>
            </View>
          )}

          {/* 热门推荐（按成员数排序） */}
          {filteredClasses.length > 0 && (
            <View className='discover-page__section'>
              <Text className='discover-page__section-title'>
                🔥 {selectedSubject ? `${selectedSubject} 班级` : '热门班级'}
              </Text>
              {filteredClasses.map((cls) => (
                <View key={cls.id} className='discover-page__class-card' onClick={() => handleClassDetail(cls)}>
                  <View className='discover-page__class-card-info'>
                    <View className='discover-page__class-card-header'>
                      <Text className='discover-page__class-card-name'>{cls.name}</Text>
                      <View className='discover-page__class-card-badge'>
                        <Text className='discover-page__class-card-badge-text'>{cls.subject}</Text>
                      </View>
                    </View>
                    <Text className='discover-page__class-card-teacher'>👨‍🏫 {cls.teacher_name}</Text>
                    {cls.description && (
                      <Text className='discover-page__class-card-desc'>{cls.description}</Text>
                    )}
                    <View className='discover-page__class-card-meta'>
                      <Text className='discover-page__class-card-count'>👥 {cls.member_count} 名成员</Text>
                    </View>
                  </View>
                  <View className='discover-page__class-card-arrow'>
                    <Text className='discover-page__class-card-arrow-text'>›</Text>
                  </View>
                </View>
              ))}
            </View>
          )}

          {/* 推荐老师 */}
          {recommendedTeachers.length > 0 && (
            <View className='discover-page__section'>
              <Text className='discover-page__section-title'>⭐ 推荐老师</Text>
              {recommendedTeachers.map((teacher) => (
                <View key={teacher.id} className='discover-page__teacher-card'>
                  <View className='discover-page__teacher-card-avatar'>
                    <Text className='discover-page__teacher-card-avatar-text'>
                      {teacher.nickname?.charAt(0) || '师'}
                    </Text>
                  </View>
                  <View className='discover-page__teacher-card-info'>
                    <Text className='discover-page__teacher-card-name'>{teacher.nickname}</Text>
                    {teacher.school && (
                      <Text className='discover-page__teacher-card-school'>🏫 {teacher.school}</Text>
                    )}
                    {teacher.description && (
                      <Text className='discover-page__teacher-card-desc'>{teacher.description}</Text>
                    )}
                    <View className='discover-page__teacher-card-meta'>
                      <Text className='discover-page__teacher-card-subject'>{teacher.subject}</Text>
                      <Text className='discover-page__teacher-card-stats'>
                        {teacher.student_count} 名学生 · {teacher.document_count} 篇文档
                      </Text>
                    </View>
                    {teacher.rating > 0 && (
                      <Text className='discover-page__teacher-card-rating'>
                        {'⭐'.repeat(Math.min(Math.round(teacher.rating), 5))} {teacher.rating.toFixed(1)}
                      </Text>
                    )}
                  </View>
                  <View className='discover-page__apply-btn' onClick={() => handleApply(teacher)}>
                    <Text className='discover-page__apply-text'>
                      {applyingId === teacher.id ? '申请中...' : '申请使用'}
                    </Text>
                  </View>
                </View>
              ))}
            </View>
          )}

          {/* 兼容旧广场数据（当新接口无数据但旧接口有数据时展示） */}
          {hotClasses.length === 0 && recommendedTeachers.length === 0 && personas.length > 0 && (
            <View className='discover-page__section'>
              <Text className='discover-page__section-title'>👨‍🏫 老师分身广场</Text>
              {personas.map((persona) => (
                <View key={persona.id} className='discover-page__teacher-card'>
                  <View className='discover-page__teacher-card-info'>
                    <Text className='discover-page__teacher-card-name'>👨‍🏫 {persona.nickname}</Text>
                    {persona.school && (
                      <Text className='discover-page__teacher-card-school'>🏫 {persona.school}</Text>
                    )}
                    {persona.description && (
                      <Text className='discover-page__teacher-card-desc'>{persona.description}</Text>
                    )}
                    <Text className='discover-page__teacher-card-stats'>
                      {persona.student_count} 名学生 · {persona.document_count} 篇文档
                    </Text>
                  </View>
                  <View
                    className={`discover-page__apply-btn ${
                      persona.application_status === 'pending' ? 'discover-page__apply-btn--pending' : ''
                    }`}
                    onClick={() => handleApplyPersona(persona)}
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
            </View>
          )}

          {/* 完全空状态 */}
          {hotClasses.length === 0 && recommendedTeachers.length === 0 && personas.length === 0 && subjects.length === 0 && (
            <View className='discover-page__empty'>
              <Empty text='暂无推荐内容' />
            </View>
          )}
        </View>
      )}
    </View>
  )
}
