import { useState } from 'react'
import { View, Text, Input, Textarea, Button } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { updateStudentProfile } from '@/api/user'
import './index.scss'

/** 性别类型 */
type GenderType = 'male' | 'female' | ''

export default function StudentProfile() {
  const [age, setAge] = useState('')
  const [gender, setGender] = useState<GenderType>('')
  const [familyInfo, setFamilyInfo] = useState('')
  const [loading, setLoading] = useState(false)

  /** 处理年龄输入，仅允许数字 */
  const handleAgeInput = (value: string) => {
    const numStr = value.replace(/\D/g, '')
    if (numStr === '' || (Number(numStr) >= 1 && Number(numStr) <= 99)) {
      setAge(numStr)
    }
  }

  /** 跳过，直接进入首页 */
  const handleSkip = () => {
    Taro.switchTab({ url: '/pages/home/index' })
  }

  /** 提交保存 */
  const handleSubmit = async () => {
    if (loading) return

    setLoading(true)

    try {
      await updateStudentProfile({
        age: age ? Number(age) : undefined,
        gender: gender || undefined,
        family_info: familyInfo.trim() || undefined,
      })

      Taro.showToast({ title: '保存成功', icon: 'success' })
      setTimeout(() => {
        Taro.switchTab({ url: '/pages/home/index' })
      }, 800)
    } catch (error: any) {
      Taro.showToast({
        title: error?.message || '保存失败，请重试',
        icon: 'none',
        duration: 2000,
      })
    } finally {
      setLoading(false)
    }
  }

  return (
    <View className='student-profile'>
      {/* 标题区域 */}
      <View className='student-profile__header'>
        <Text className='student-profile__title'>完善个人信息</Text>
        <Text className='student-profile__subtitle'>
          填写基础信息，帮助老师更好地了解你{'\n'}你也可以选择跳过，稍后再填
        </Text>
      </View>

      {/* 表单区域 */}
      <View className='student-profile__form'>
        {/* 年龄 */}
        <View className='student-profile__field'>
          <Text className='student-profile__label'>年龄</Text>
          <Input
            className='student-profile__input'
            type='number'
            placeholder='请输入你的年龄'
            placeholderClass='student-profile__input-placeholder'
            maxlength={2}
            value={age}
            onInput={(e) => handleAgeInput(e.detail.value)}
          />
        </View>

        {/* 性别 */}
        <View className='student-profile__field'>
          <Text className='student-profile__label'>性别</Text>
          <View className='student-profile__gender-group'>
            <View
              className={`student-profile__gender-item ${gender === 'male' ? 'student-profile__gender-item--active' : ''}`}
              onClick={() => setGender('male')}
            >
              <Text className='student-profile__gender-icon'>👦</Text>
              <Text>男</Text>
            </View>
            <View
              className={`student-profile__gender-item ${gender === 'female' ? 'student-profile__gender-item--active' : ''}`}
              onClick={() => setGender('female')}
            >
              <Text className='student-profile__gender-icon'>👧</Text>
              <Text>女</Text>
            </View>
          </View>
        </View>

        {/* 家庭情况 */}
        <View className='student-profile__field'>
          <Text className='student-profile__label'>
            家庭情况
            <Text className='student-profile__label-optional'>（选填）</Text>
          </Text>
          <Textarea
            className='student-profile__textarea'
            placeholder='简要描述家庭情况，如家庭成员、兴趣爱好等'
            maxlength={200}
            value={familyInfo}
            onInput={(e) => setFamilyInfo(e.detail.value)}
          />
        </View>
      </View>

      {/* 底部按钮 */}
      <View className='student-profile__actions'>
        <Button
          className='student-profile__btn-skip'
          onClick={handleSkip}
        >
          跳过
        </Button>
        <Button
          className={`student-profile__btn-submit ${loading ? 'student-profile__btn-submit--disabled' : ''}`}
          onClick={handleSubmit}
          disabled={loading}
        >
          {loading ? '保存中...' : '完成'}
        </Button>
      </View>
    </View>
  )
}
