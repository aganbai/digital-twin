import { useState } from 'react'
import { View, Text, Input, Textarea } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { createClass } from '@/api/class'
import './index.scss'

export default function ClassCreate() {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [submitting, setSubmitting] = useState(false)

  /** 名称是否合法 */
  const canSubmit = name.trim().length >= 1 && name.trim().length <= 50 && !submitting

  /** 提交创建 */
  const handleSubmit = async () => {
    if (!canSubmit) return

    if (!name.trim()) {
      Taro.showToast({ title: '请输入班级名称', icon: 'none' })
      return
    }

    setSubmitting(true)
    try {
      await createClass(name.trim(), description.trim() || undefined)
      Taro.showToast({ title: '创建成功', icon: 'success' })
      setTimeout(() => Taro.navigateBack(), 1500)
    } catch (error) {
      console.error('创建班级失败:', error)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <View className='class-create'>
      <View className='class-create__header'>
        <Text className='class-create__title'>创建班级</Text>
        <Text className='class-create__subtitle'>创建一个班级来管理你的学生</Text>
      </View>

      {/* 班级名称 */}
      <View className='class-create__section'>
        <Text className='class-create__label'>班级名称</Text>
        <Input
          className='class-create__input'
          placeholder='请输入班级名称'
          placeholderClass='class-create__input-placeholder'
          maxlength={50}
          value={name}
          onInput={(e) => setName(e.detail.value)}
        />
      </View>

      {/* 班级描述 */}
      <View className='class-create__section'>
        <Text className='class-create__label'>班级描述（可选）</Text>
        <Textarea
          className='class-create__textarea'
          placeholder='请输入班级描述'
          maxlength={200}
          value={description}
          onInput={(e) => setDescription(e.detail.value)}
        />
        <Text className='class-create__char-count'>{description.length}/200</Text>
      </View>

      {/* 创建按钮 */}
      <View
        className={`class-create__submit ${!canSubmit ? 'class-create__submit--disabled' : ''}`}
        onClick={canSubmit ? handleSubmit : undefined}
      >
        <Text className='class-create__submit-text'>
          {submitting ? '创建中...' : '创建班级'}
        </Text>
      </View>
    </View>
  )
}
