import { useState, useEffect } from 'react'
import { View, Text } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import { getShareInfo, joinShare } from '@/api/share'
import { getPersonas } from '@/api/persona'
import { applyTeacher } from '@/api/relation'
import type { Persona } from '@/api/persona'
import { useUserStore } from '@/store'
import './index.scss'

/** 扩展的分享详情（含 join_status） */
interface ShareDetailExtended {
  teacher_persona_id: number
  teacher_nickname: string
  teacher_school?: string
  teacher_description?: string
  class_name?: string
  target_student_persona_id?: number
  target_student_nickname?: string
  is_valid: boolean
  /** 迭代6新增：加入状态 */
  join_status?: 'can_join' | 'already_joined' | 'not_target' | 'need_login' | 'need_persona'
}

export default function ShareJoin() {
  const router = useRouter()
  const code = router.params.code || ''
  const { token } = useUserStore()

  const [shareInfo, setShareInfo] = useState<ShareDetailExtended | null>(null)
  const [studentPersonas, setStudentPersonas] = useState<Persona[]>([])
  const [selectedPersonaId, setSelectedPersonaId] = useState<number | undefined>(undefined)
  const [loading, setLoading] = useState(true)
  const [joining, setJoining] = useState(false)
  const [applying, setApplying] = useState(false)

  /** 获取分享码信息和学生分身列表 */
  useEffect(() => {
    const fetchData = async () => {
      setLoading(true)
      try {
        const shareRes = await getShareInfo(code)
        const info = shareRes.data as ShareDetailExtended
        setShareInfo(info)

        // 如果需要选择分身，获取分身列表
        if (info.join_status === 'can_join' || !info.join_status) {
          try {
            const personaRes = await getPersonas()
            const students = (personaRes.data.personas || []).filter(
              (p) => p.role === 'student' && p.is_active,
            )
            setStudentPersonas(students)
            if (students.length === 1) {
              setSelectedPersonaId(students[0].id)
            }
          } catch {
            // 未登录时获取分身会失败，忽略
          }
        }
      } catch (error) {
        console.error('获取分享信息失败:', error)
      } finally {
        setLoading(false)
      }
    }

    if (code) {
      fetchData()
    }
  }, [code])

  /** 确认加入 */
  const handleJoin = async () => {
    if (joining || !shareInfo?.is_valid) return
    setJoining(true)
    try {
      const res = await joinShare(code, selectedPersonaId)
      const data = res.data as any

      // 检查是否为非目标学生引导
      if (data?.join_status === 'not_target') {
        setShareInfo((prev) => prev ? { ...prev, join_status: 'not_target' } : prev)
        return
      }

      Taro.showToast({ title: '加入成功', icon: 'success' })
      setTimeout(() => {
        Taro.switchTab({ url: '/pages/home/index' })
      }, 1500)
    } catch (error) {
      console.error('加入失败:', error)
    } finally {
      setJoining(false)
    }
  }

  /** 向老师申请 */
  const handleApply = async () => {
    if (applying || !shareInfo) return
    setApplying(true)
    try {
      await applyTeacher(0, shareInfo.teacher_persona_id)
      Taro.showToast({ title: '申请已发送', icon: 'success' })
    } catch (error) {
      console.error('申请失败:', error)
    } finally {
      setApplying(false)
    }
  }

  /** 跳转登录 */
  const handleGoLogin = () => {
    Taro.redirectTo({ url: '/pages/login/index' })
  }

  /** 跳转创建分身 */
  const handleGoCreatePersona = () => {
    Taro.redirectTo({ url: '/pages/role-select/index' })
  }

  if (loading) {
    return (
      <View className='share-join'>
        <View className='share-join__loading'>
          <Text className='share-join__loading-text'>加载中...</Text>
        </View>
      </View>
    )
  }

  if (!shareInfo || !shareInfo.is_valid) {
    return (
      <View className='share-join'>
        <View className='share-join__invalid'>
          <Text className='share-join__invalid-icon'>😕</Text>
          <Text className='share-join__invalid-text'>分享码无效或已过期</Text>
          <View
            className='share-join__back-btn'
            onClick={() => Taro.navigateBack()}
          >
            <Text className='share-join__back-btn-text'>返回</Text>
          </View>
        </View>
      </View>
    )
  }

  const joinStatus = shareInfo.join_status || 'can_join'

  return (
    <View className='share-join'>
      {/* 教师信息 */}
      <View className='share-join__teacher'>
        <View className='share-join__teacher-avatar'>👨‍🏫</View>
        <Text className='share-join__teacher-name'>{shareInfo.teacher_nickname}</Text>
        {shareInfo.teacher_school && (
          <Text className='share-join__teacher-school'>{shareInfo.teacher_school}</Text>
        )}
        {shareInfo.teacher_description && (
          <Text className='share-join__teacher-desc'>{shareInfo.teacher_description}</Text>
        )}
        {shareInfo.class_name && (
          <View className='share-join__class-tag'>
            <Text className='share-join__class-tag-text'>班级：{shareInfo.class_name}</Text>
          </View>
        )}
      </View>

      {/* 状态：需要登录 */}
      {joinStatus === 'need_login' && (
        <View className='share-join__status-card'>
          <Text className='share-join__status-icon'>🔐</Text>
          <Text className='share-join__status-text'>请先登录后再加入</Text>
          <View className='share-join__action-btn' onClick={handleGoLogin}>
            <Text className='share-join__action-btn-text'>去登录</Text>
          </View>
        </View>
      )}

      {/* 状态：需要创建分身 */}
      {joinStatus === 'need_persona' && (
        <View className='share-join__status-card'>
          <Text className='share-join__status-icon'>👤</Text>
          <Text className='share-join__status-text'>请先创建学生分身</Text>
          <View className='share-join__action-btn' onClick={handleGoCreatePersona}>
            <Text className='share-join__action-btn-text'>去创建</Text>
          </View>
        </View>
      )}

      {/* 状态：已加入 */}
      {joinStatus === 'already_joined' && (
        <View className='share-join__status-card'>
          <Text className='share-join__status-icon'>✅</Text>
          <Text className='share-join__status-text'>你已经加入了该老师</Text>
          <View
            className='share-join__action-btn'
            onClick={() => Taro.switchTab({ url: '/pages/home/index' })}
          >
            <Text className='share-join__action-btn-text'>去对话</Text>
          </View>
        </View>
      )}

      {/* 状态：非目标学生 */}
      {joinStatus === 'not_target' && (
        <View className='share-join__status-card share-join__status-card--warning'>
          <Text className='share-join__status-icon'>🎯</Text>
          <Text className='share-join__status-text'>
            该邀请码是老师专门发给特定同学的
          </Text>
          <Text className='share-join__status-hint'>
            你可以向老师发起申请，等待老师同意后即可加入
          </Text>
          <View
            className={`share-join__action-btn share-join__action-btn--apply ${applying ? 'share-join__action-btn--disabled' : ''}`}
            onClick={applying ? undefined : handleApply}
          >
            <Text className='share-join__action-btn-text'>
              {applying ? '申请中...' : '向老师申请'}
            </Text>
          </View>
        </View>
      )}

      {/* 状态：可加入 */}
      {joinStatus === 'can_join' && (
        <>
          {/* 选择学生分身 */}
          {studentPersonas.length > 1 && (
            <View className='share-join__persona-section'>
              <Text className='share-join__persona-title'>选择学生分身</Text>
              {studentPersonas.map((p) => (
                <View
                  key={p.id}
                  className={`share-join__persona-item ${selectedPersonaId === p.id ? 'share-join__persona-item--active' : ''}`}
                  onClick={() => setSelectedPersonaId(p.id)}
                >
                  <View className='share-join__persona-icon'>👨‍🎓</View>
                  <Text className='share-join__persona-name'>{p.nickname}</Text>
                  {selectedPersonaId === p.id && (
                    <View className='share-join__persona-check'>✓</View>
                  )}
                </View>
              ))}
            </View>
          )}

          {/* 加入按钮 */}
          <View
            className={`share-join__join-btn ${joining ? 'share-join__join-btn--disabled' : ''}`}
            onClick={handleJoin}
          >
            <Text className='share-join__join-btn-text'>
              {joining ? '加入中...' : '确认加入'}
            </Text>
          </View>
        </>
      )}
    </View>
  )
}
