import { useState, useCallback } from 'react'
import { View, Text } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { getPersonas, setVisibility } from '@/api/persona'
import type { Persona } from '@/api/persona'
import { getPersonaDashboard } from '@/api/persona'
import type { DashboardData } from '@/api/persona'
import { usePersonaStore } from '@/store'
import './index.scss'

/** 分身概览项 */
interface PersonaOverviewItem extends Persona {
  student_count?: number
  document_count?: number
  class_count?: number
}

export default function PersonaOverview() {
  const { setCurrentPersona } = usePersonaStore()
  const [personas, setPersonas] = useState<PersonaOverviewItem[]>([])
  const [loading, setLoading] = useState(true)
  const [toggling, setToggling] = useState<number | null>(null)

  /** 汇总统计 */
  const totalStudents = personas.reduce((sum, p) => sum + (p.student_count || 0), 0)
  const totalClasses = personas.reduce((sum, p) => sum + (p.class_count || 0), 0)

  /** 获取分身列表及统计数据 */
  const fetchPersonas = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getPersonas()
      const teacherPersonas = (res.data.personas || []).filter((p) => p.role === 'teacher')

      // 并行获取每个分身的仪表盘数据
      const dashboardPromises = teacherPersonas.map(async (p) => {
        try {
          const dashRes = await getPersonaDashboard(p.id)
          return {
            ...p,
            student_count: dashRes.data.stats?.total_students || 0,
            document_count: dashRes.data.stats?.total_documents || 0,
            class_count: dashRes.data.stats?.total_classes || 0,
          }
        } catch {
          return { ...p, student_count: 0, document_count: 0, class_count: 0 }
        }
      })

      const enrichedPersonas = await Promise.all(dashboardPromises)
      setPersonas(enrichedPersonas)
    } catch (error) {
      console.error('获取分身列表失败:', error)
    } finally {
      setLoading(false)
    }
  }, [])

  /** 页面显示时刷新 */
  useDidShow(() => {
    fetchPersonas()
  })

  /** 切换公开/私有 */
  const handleToggleVisibility = async (persona: PersonaOverviewItem) => {
    if (toggling) return
    setToggling(persona.id)
    try {
      const newIsPublic = !persona.is_public
      await setVisibility(persona.id, newIsPublic)
      setPersonas((prev) =>
        prev.map((p) => (p.id === persona.id ? { ...p, is_public: newIsPublic } : p)),
      )
      Taro.showToast({
        title: newIsPublic ? '已公开到广场' : '已设为私有',
        icon: 'success',
      })
    } catch (error) {
      console.error('设置公开状态失败:', error)
    } finally {
      setToggling(null)
    }
  }

  /** 进入分身仪表盘 */
  const handleEnterDashboard = (persona: PersonaOverviewItem) => {
    setCurrentPersona(persona)
    Taro.switchTab({ url: '/pages/home/index' })
  }

  /** 创建新分身 */
  const handleCreatePersona = () => {
    Taro.navigateTo({ url: '/pages/role-select/index' })
  }

  if (loading) {
    return (
      <View className='persona-overview'>
        <View className='persona-overview__loading'>
          <Text className='persona-overview__loading-text'>加载中...</Text>
        </View>
      </View>
    )
  }

  return (
    <View className='persona-overview'>
      {/* 顶部统计 */}
      <View className='persona-overview__header'>
        <Text className='persona-overview__title'>👨‍🏫 我的分身</Text>
        <Text className='persona-overview__summary'>
          共 {personas.length} 个分身 · {totalStudents} 名学生 · {totalClasses} 个班级
        </Text>
      </View>

      {/* 分身列表 */}
      <View className='persona-overview__list'>
        {personas.map((persona) => (
          <View key={persona.id} className='persona-overview__card'>
            <View className='persona-overview__card-header'>
              <View className='persona-overview__card-info'>
                <Text className='persona-overview__card-name'>{persona.nickname}</Text>
                {persona.school && (
                  <Text className='persona-overview__card-school'> · {persona.school}</Text>
                )}
              </View>
            </View>

            {persona.description && (
              <Text className='persona-overview__card-desc'>{persona.description}</Text>
            )}

            {/* 状态标签 */}
            <View className='persona-overview__card-badges'>
              <View
                className={`persona-overview__badge ${persona.is_active ? 'persona-overview__badge--active' : 'persona-overview__badge--inactive'}`}
              >
                <Text className='persona-overview__badge-text'>
                  {persona.is_active ? '🟢 启用中' : '🔴 已停用'}
                </Text>
              </View>
              <View
                className={`persona-overview__badge ${persona.is_public ? 'persona-overview__badge--public' : 'persona-overview__badge--private'}`}
                onClick={() => handleToggleVisibility(persona)}
              >
                <Text className='persona-overview__badge-text'>
                  {toggling === persona.id
                    ? '切换中...'
                    : persona.is_public
                      ? '🌐 已公开'
                      : '🔒 未公开'}
                </Text>
              </View>
            </View>

            {/* 统计数据 */}
            <View className='persona-overview__card-stats'>
              <View className='persona-overview__stat'>
                <Text className='persona-overview__stat-value'>{persona.student_count || 0}</Text>
                <Text className='persona-overview__stat-label'>学生</Text>
              </View>
              <View className='persona-overview__stat'>
                <Text className='persona-overview__stat-value'>{persona.class_count || 0}</Text>
                <Text className='persona-overview__stat-label'>班级</Text>
              </View>
              <View className='persona-overview__stat'>
                <Text className='persona-overview__stat-value'>{persona.document_count || 0}</Text>
                <Text className='persona-overview__stat-label'>文档</Text>
              </View>
            </View>

            {/* 操作按钮 */}
            <View className='persona-overview__card-actions-row'>
              <View
                className='persona-overview__card-btn'
                onClick={() => handleEnterDashboard(persona)}
              >
                <Text className='persona-overview__card-btn-text'>进入管理</Text>
              </View>
              {/* R6: 分享码管理入口 */}
              <View
                className='persona-overview__card-btn persona-overview__card-btn--share'
                onClick={() => {
                  // 进入该分身后跳转分享码管理
                  setCurrentPersona(persona)
                  Taro.navigateTo({ url: '/pages/share-manage/index' })
                }}
              >
                <Text className='persona-overview__card-btn-text--share'>🔗 分享码</Text>
              </View>
            </View>
          </View>
        ))}
      </View>

      {/* 创建新分身按钮 */}
      <View className='persona-overview__create-btn' onClick={handleCreatePersona}>
        <Text className='persona-overview__create-btn-text'>+ 创建新分身</Text>
      </View>
    </View>
  )
}
