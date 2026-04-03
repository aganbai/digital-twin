import { useState } from 'react'
import { View, Text, Button } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { createPersona } from '@/api/persona'
import { getClasses } from '@/api/class'
import { useUserStore, usePersonaStore } from '@/store'
import './index.scss'

/** 角色类型 */
type RoleType = 'teacher' | 'student' | ''

export default function RoleSelect() {
  const [selectedRole, setSelectedRole] = useState<RoleType>('')
  const [loading, setLoading] = useState(false)
  const { setUserInfo, setToken } = useUserStore()
  const { setCurrentPersona } = usePersonaStore()

  /** 按钮是否可用 */
  const canSubmit = !!selectedRole && !loading

  /** 获取微信用户昵称 */
  const getWxNickname = (): string => {
    // 微信小程序中，昵称通过后端从微信接口获取，前端传默认值
    // 后端 createPersona 会自动使用微信昵称
    return '微信用户'
  }

  /** 提交选择 */
  const handleSubmit = async () => {
    if (!canSubmit) return

    setLoading(true)

    try {
      const nickname = getWxNickname()
      const res = await createPersona(selectedRole, nickname)

      const rawData = res.data as any
      const personaId = rawData.persona_id || rawData.id
      const role = rawData.role || selectedRole
      const personaNickname = rawData.nickname || nickname

      // 使用返回的新 token 更新本地存储
      if (rawData.token) {
        setToken(rawData.token)
      }

      // 更新用户信息
      setUserInfo({ id: personaId, nickname: personaNickname, role })

      // 更新当前分身
      setCurrentPersona({
        id: personaId,
        role,
        nickname: personaNickname,
        school: rawData.school || '',
        description: rawData.description || '',
        is_active: true,
        created_at: new Date().toISOString(),
      } as any)

      Taro.showToast({ title: '创建成功', icon: 'success' })

      // 根据角色跳转
      setTimeout(async () => {
        if (role === 'student') {
          // 学生 → 跳转基础信息填写页（可跳过）
          Taro.redirectTo({ url: '/pages/student-profile/index' })
        } else if (role === 'teacher') {
          // 教师 → 检查是否有班级
          try {
            const classRes = await getClasses()
            const classes = classRes.data as any
            const classList = Array.isArray(classes) ? classes : (classes?.items || [])

            if (classList.length > 0) {
              // 已有班级 → 直接进首页
              Taro.switchTab({ url: '/pages/home/index' })
            } else {
              // 无班级 → 跳转班级创建引导
              Taro.redirectTo({ url: '/pages/class-create/index?from=register' })
            }
          } catch {
            // 获取班级失败，跳转班级创建引导
            Taro.redirectTo({ url: '/pages/class-create/index?from=register' })
          }
        }
      }, 800)
    } catch (error: any) {
      const msg = error?.message || '提交失败，请重试'
      if (!msg.includes('同名')) {
        Taro.showToast({ title: msg, icon: 'none', duration: 2000 })
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <View className='role-select'>
      {/* 标题区域 */}
      <View className='role-select__header'>
        <Text className='role-select__title'>选择你的身份</Text>
        <Text className='role-select__subtitle'>请选择你的角色，开始使用</Text>
      </View>

      {/* 角色卡片区域 */}
      <View className='role-select__cards'>
        {/* 教师卡片 */}
        <View
          className={`role-select__card ${selectedRole === 'teacher' ? 'role-select__card--active' : ''}`}
          onClick={() => setSelectedRole('teacher')}
        >
          <View className='role-select__card-icon'>👨‍🏫</View>
          <View className='role-select__card-info'>
            <Text className='role-select__card-title'>我是老师</Text>
            <Text className='role-select__card-desc'>
              创建班级，打造你的数字分身
            </Text>
          </View>
          {selectedRole === 'teacher' && (
            <View className='role-select__card-check'>✓</View>
          )}
        </View>

        {/* 学生卡片 */}
        <View
          className={`role-select__card ${selectedRole === 'student' ? 'role-select__card--active' : ''}`}
          onClick={() => setSelectedRole('student')}
        >
          <View className='role-select__card-icon'>👨‍🎓</View>
          <View className='role-select__card-info'>
            <Text className='role-select__card-title'>我是学生</Text>
            <Text className='role-select__card-desc'>
              与老师的数字分身对话学习
            </Text>
          </View>
          {selectedRole === 'student' && (
            <View className='role-select__card-check'>✓</View>
          )}
        </View>
      </View>

      {/* 确认按钮 */}
      <Button
        className={`role-select__btn ${!canSubmit ? 'role-select__btn--disabled' : ''}`}
        onClick={handleSubmit}
        disabled={!canSubmit}
      >
        {loading ? '处理中...' : '确认选择'}
      </Button>
    </View>
  )
}
