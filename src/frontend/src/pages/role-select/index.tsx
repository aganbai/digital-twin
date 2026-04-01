import { useState } from 'react'
import { View, Text, Input, Textarea, Button } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { createPersona } from '@/api/persona'
import { useUserStore, usePersonaStore } from '@/store'
import './index.scss'

/** 角色类型 */
type RoleType = 'teacher' | 'student' | ''

export default function RoleSelect() {
  const [selectedRole, setSelectedRole] = useState<RoleType>('')
  const [nickname, setNickname] = useState('')
  const [school, setSchool] = useState('')
  const [description, setDescription] = useState('')
  const [loading, setLoading] = useState(false)
  const { setUserInfo, setToken } = useUserStore()
  const { setCurrentPersona } = usePersonaStore()

  /** 昵称是否合法（1-20 字符） */
  const isNicknameValid = nickname.trim().length >= 1 && nickname.trim().length <= 20

  /** 教师额外字段是否合法 */
  const isTeacherFieldsValid =
    selectedRole !== 'teacher' ||
    (school.trim().length >= 1 && school.trim().length <= 128 &&
      description.trim().length >= 1 && description.trim().length <= 500)

  /** 按钮是否可用 */
  const canSubmit = !!selectedRole && isNicknameValid && isTeacherFieldsValid && !loading

  /** 提交创建分身 */
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

    // 教师额外校验
    if (selectedRole === 'teacher') {
      if (!school.trim()) {
        Taro.showToast({ title: '请输入学校名称', icon: 'none' })
        return
      }
      if (!description.trim()) {
        Taro.showToast({ title: '请输入分身描述', icon: 'none' })
        return
      }
    }

    setLoading(true)

    try {
      const res = await createPersona(
        selectedRole,
        nickname.trim(),
        selectedRole === 'teacher' ? school.trim() : undefined,
        selectedRole === 'teacher' ? description.trim() : undefined,
      )
      const persona = res.data

      // 更新用户信息
      setUserInfo({ id: persona.id, nickname: persona.nickname, role: persona.role })
      setCurrentPersona(persona)

      Taro.showToast({ title: '创建成功', icon: 'success' })

      // 根据角色跳转对应首页
      setTimeout(() => {
        if (persona.role === 'student') {
          Taro.switchTab({ url: '/pages/home/index' })
        } else if (persona.role === 'teacher') {
          Taro.redirectTo({ url: '/pages/knowledge/index' })
        }
      }, 1000)
    } catch (error: any) {
      // 处理 40008 错误码（同名+同校教师已存在）
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
        <Text className='role-select__title'>创建新分身</Text>
        <Text className='role-select__subtitle'>请选择分身角色并填写信息</Text>
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
            <Text className='role-select__card-title'>教师分身</Text>
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
            <Text className='role-select__card-title'>学生分身</Text>
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
        <Text className='role-select__field-label'>昵称</Text>
        <Input
          className='role-select__input'
          placeholder='请输入你的昵称'
          placeholderClass='role-select__input-placeholder'
          maxlength={20}
          value={nickname}
          onInput={(e) => setNickname(e.detail.value)}
        />
      </View>

      {/* 教师额外字段 */}
      {selectedRole === 'teacher' && (
        <>
          <View className='role-select__input-wrap'>
            <Text className='role-select__field-label'>学校</Text>
            <Input
              className='role-select__input'
              placeholder='请输入学校名称'
              placeholderClass='role-select__input-placeholder'
              maxlength={128}
              value={school}
              onInput={(e) => setSchool(e.detail.value)}
            />
          </View>
          <View className='role-select__input-wrap'>
            <Text className='role-select__field-label'>分身描述</Text>
            <Textarea
              className='role-select__textarea'
              placeholder='请简要描述你的数字分身（如：物理学教授，专注力学教学）'
              maxlength={500}
              value={description}
              onInput={(e) => setDescription(e.detail.value)}
            />
            <Text className='role-select__char-count'>{description.length}/500</Text>
          </View>
        </>
      )}

      {/* 确认按钮 */}
      <Button
        className={`role-select__btn ${!canSubmit ? 'role-select__btn--disabled' : ''}`}
        onClick={handleSubmit}
        disabled={!canSubmit}
      >
        {loading ? '创建中...' : '创建分身'}
      </Button>
    </View>
  )
}
