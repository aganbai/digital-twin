import { useState } from 'react'
import { View, Text, Input, Button } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { completeProfile } from '@/api/auth'
import { useUserStore } from '@/store'
import './index.scss'

/** 角色类型 */
type RoleType = 'teacher' | 'student' | ''

export default function RoleSelect() {
  const [selectedRole, setSelectedRole] = useState<RoleType>('')
  const [nickname, setNickname] = useState('')
  const [loading, setLoading] = useState(false)
  const { setUserInfo } = useUserStore()

  /** 昵称是否合法（1-20 字符） */
  const isNicknameValid = nickname.trim().length >= 1 && nickname.trim().length <= 20

  /** 按钮是否可用 */
  const canSubmit = !!selectedRole && isNicknameValid && !loading

  /** 提交角色选择 */
  const handleSubmit = async () => {
    if (!canSubmit) return

    // 校验昵称
    if (!nickname.trim()) {
      Taro.showToast({ title: '请输入昵称', icon: 'none' })
      return
    }
    if (nickname.trim().length > 20) {
      Taro.showToast({ title: '昵称不能超过20个字符', icon: 'none' })
      return
    }

    setLoading(true)

    try {
      const res = await completeProfile(selectedRole, nickname.trim())
      const { user_id, role, nickname: savedNickname } = res.data

      // 更新用户信息
      setUserInfo({ id: user_id, nickname: savedNickname, role })

      // 根据角色跳转对应首页
      if (role === 'student') {
        Taro.switchTab({ url: '/pages/home/index' })
      } else if (role === 'teacher') {
        Taro.redirectTo({ url: '/pages/knowledge/index' })
      }
    } catch (error) {
      Taro.showToast({
        title: '提交失败，请重试',
        icon: 'none',
        duration: 2000,
      })
    } finally {
      setLoading(false)
    }
  }

  return (
    <View className='role-select'>
      {/* 标题区域 */}
      <View className='role-select__header'>
        <Text className='role-select__title'>欢迎加入 AI 数字分身</Text>
        <Text className='role-select__subtitle'>请选择你的身份</Text>
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
            <Text className='role-select__card-title'>我是教师</Text>
            <Text className='role-select__card-desc'>
              创建知识库，打造你的数字分身
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
              与教师的数字分身对话学习
            </Text>
          </View>
          {selectedRole === 'student' && (
            <View className='role-select__card-check'>✓</View>
          )}
        </View>
      </View>

      {/* 昵称输入 */}
      <View className='role-select__input-wrap'>
        <Input
          className='role-select__input'
          placeholder='请输入你的昵称'
          placeholderClass='role-select__input-placeholder'
          maxlength={20}
          value={nickname}
          onInput={(e) => setNickname(e.detail.value)}
        />
      </View>

      {/* 确认按钮 */}
      <Button
        className={`role-select__btn ${!canSubmit ? 'role-select__btn--disabled' : ''}`}
        onClick={handleSubmit}
        disabled={!canSubmit}
      >
        {loading ? '提交中...' : '开始使用'}
      </Button>
    </View>
  )
}
