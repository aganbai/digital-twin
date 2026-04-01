import { useCallback, useEffect, useState } from 'react'
import { View, Text } from '@tarojs/components'
import Taro, { usePullDownRefresh, useDidShow } from '@tarojs/taro'
import { getPersonaDashboard, DashboardData } from '@/api/persona'
import { usePersonaStore } from '@/store'
import './index.scss'

interface TeacherDashboardProps {
  personaId: number
}

export default function TeacherDashboard({ personaId }: TeacherDashboardProps) {
  const [dashboardData, setDashboardData] = useState<DashboardData | null>(null)
  const [dashboardLoading, setDashboardLoading] = useState(false)
  const [shareCodeCopied, setShareCodeCopied] = useState(false)

  /** 获取教师仪表盘数据 */
  const fetchDashboard = useCallback(async () => {
    if (!personaId) return
    setDashboardLoading(true)
    try {
      const res = await getPersonaDashboard(personaId)
      setDashboardData(res.data)
    } catch (error) {
      console.error('获取仪表盘数据失败:', error)
    } finally {
      setDashboardLoading(false)
    }
  }, [personaId])

  /** 页面加载时获取数据 */
  useEffect(() => {
    fetchDashboard()
  }, [fetchDashboard])

  /** 页面显示时刷新数据 */
  useDidShow(() => {
    fetchDashboard()
  })

  /** 下拉刷新 */
  usePullDownRefresh(async () => {
    await fetchDashboard()
    Taro.stopPullDownRefresh()
  })

  /** 复制分享码 */
  const handleCopyShareCode = (code: string) => {
    Taro.setClipboardData({
      data: code,
      success: () => {
        setShareCodeCopied(true)
        setTimeout(() => setShareCodeCopied(false), 2000)
      },
    })
  }

  /** 跳转班级详情 */
  const handleGoClassDetail = (classId: number, className: string) => {
    Taro.navigateTo({
      url: `/pages/class-detail/index?class_id=${classId}&class_name=${encodeURIComponent(className)}`,
    })
  }

  /** 跳转快捷操作 */
  const handleQuickAction = (path: string) => {
    Taro.navigateTo({ url: path })
  }

  if (dashboardLoading && !dashboardData) {
    return (
      <View className='teacher-dashboard__loading'>
        <Text className='teacher-dashboard__loading-text'>加载中...</Text>
      </View>
    )
  }

  return (
    <View className='teacher-dashboard'>
      {/* 待审批提醒 */}
      {dashboardData && dashboardData.pending_count > 0 && (
        <View
          className='teacher-dashboard__card teacher-dashboard__pending-card'
          onClick={() => handleQuickAction('/pages/teacher-students/index')}
        >
          <View className='teacher-dashboard__pending-info'>
            <Text className='teacher-dashboard__pending-icon'>🔔</Text>
            <Text className='teacher-dashboard__pending-text'>
              有 {dashboardData.pending_count} 条待审批申请
            </Text>
          </View>
          <Text className='teacher-dashboard__pending-arrow'>→</Text>
        </View>
      )}

      {/* 数据统计 */}
      {dashboardData?.stats && (
        <View className='teacher-dashboard__card teacher-dashboard__stats-card'>
          <View className='teacher-dashboard__stats-item'>
            <Text className='teacher-dashboard__stats-value'>{dashboardData.stats.total_students}</Text>
            <Text className='teacher-dashboard__stats-label'>学生</Text>
          </View>
          <View className='teacher-dashboard__stats-divider' />
          <View className='teacher-dashboard__stats-item'>
            <Text className='teacher-dashboard__stats-value'>{dashboardData.stats.total_documents}</Text>
            <Text className='teacher-dashboard__stats-label'>文档</Text>
          </View>
          <View className='teacher-dashboard__stats-divider' />
          <View className='teacher-dashboard__stats-item'>
            <Text className='teacher-dashboard__stats-value'>{dashboardData.stats.total_classes}</Text>
            <Text className='teacher-dashboard__stats-label'>班级</Text>
          </View>
          {/* TODO: 后续迭代扩展 Dashboard 接口后，新增"对话数"统计卡片 */}
          {/* TODO: 后续迭代扩展 Dashboard 接口后，新增"今日活跃"统计卡片 */}
        </View>
      )}

      {/* 快捷操作 */}
      <View className='teacher-dashboard__card teacher-dashboard__actions-card'>
        <Text className='teacher-dashboard__card-title'>快捷操作</Text>
        <View className='teacher-dashboard__actions'>
          <View
            className='teacher-dashboard__action-item'
            onClick={() => handleQuickAction('/pages/persona-overview/index')}
          >
            <Text className='teacher-dashboard__action-icon'>🧑‍🏫</Text>
            <Text className='teacher-dashboard__action-label'>分身概览</Text>
          </View>
          <View
            className='teacher-dashboard__action-item'
            onClick={() => handleQuickAction('/pages/knowledge/index')}
          >
            <Text className='teacher-dashboard__action-icon'>📚</Text>
            <Text className='teacher-dashboard__action-label'>知识库管理</Text>
          </View>
          <View
            className='teacher-dashboard__action-item'
            onClick={() => handleQuickAction('/pages/assignment-list/index')}
          >
            <Text className='teacher-dashboard__action-icon'>📝</Text>
            <Text className='teacher-dashboard__action-label'>作业管理</Text>
          </View>
          <View
            className='teacher-dashboard__action-item'
            onClick={() => handleQuickAction('/pages/teacher-students/index')}
          >
            <Text className='teacher-dashboard__action-icon'>👥</Text>
            <Text className='teacher-dashboard__action-label'>学生管理</Text>
          </View>
        </View>
      </View>

      {/* 我的班级 */}
      <View className='teacher-dashboard__card'>
        <View className='teacher-dashboard__card-header'>
          <Text className='teacher-dashboard__card-title'>我的班级</Text>
          <View
            className='teacher-dashboard__card-action'
            onClick={() => handleQuickAction('/pages/class-create/index')}
          >
            <Text className='teacher-dashboard__card-action-text'>+ 创建班级</Text>
          </View>
        </View>
        {dashboardData?.classes && dashboardData.classes.length > 0 ? (
          <View className='teacher-dashboard__class-list'>
            {dashboardData.classes.map((cls) => (
              <View
                key={cls.id}
                className='teacher-dashboard__class-item'
                onClick={() => handleGoClassDetail(cls.id, cls.name)}
              >
                <View className='teacher-dashboard__class-info'>
                  <Text className='teacher-dashboard__class-name'>{cls.name}</Text>
                  <Text className='teacher-dashboard__class-count'>{cls.member_count} 名学生</Text>
                </View>
                <View className='teacher-dashboard__class-status'>
                  <Text className={`teacher-dashboard__class-badge ${cls.is_active ? 'teacher-dashboard__class-badge--active' : 'teacher-dashboard__class-badge--inactive'}`}>
                    {cls.is_active ? '启用' : '停用'}
                  </Text>
                  <Text className='teacher-dashboard__class-arrow'>›</Text>
                </View>
              </View>
            ))}
          </View>
        ) : (
          <View className='teacher-dashboard__empty-hint'>
            <Text className='teacher-dashboard__empty-hint-text'>暂无班级，点击上方创建</Text>
          </View>
        )}
      </View>

      {/* 分享码展示 */}
      {dashboardData?.latest_share && (
        <View className='teacher-dashboard__card teacher-dashboard__share-card'>
          <Text className='teacher-dashboard__card-title'>最新分享码</Text>
          <View className='teacher-dashboard__share-info'>
            <View className='teacher-dashboard__share-code-wrap'>
              <Text className='teacher-dashboard__share-code'>{dashboardData.latest_share.share_code}</Text>
              <View
                className='teacher-dashboard__share-copy'
                onClick={() => handleCopyShareCode(dashboardData.latest_share!.share_code)}
              >
                <Text className='teacher-dashboard__share-copy-text'>
                  {shareCodeCopied ? '已复制' : '复制'}
                </Text>
              </View>
            </View>
            <Text className='teacher-dashboard__share-meta'>
              班级：{dashboardData.latest_share.class_name} · 已使用 {dashboardData.latest_share.used_count}/{dashboardData.latest_share.max_uses}
            </Text>
          </View>
        </View>
      )}
    </View>
  )
}
