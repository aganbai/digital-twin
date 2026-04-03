import { useState } from 'react'
import { View, Text, Input, Textarea, Image } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import { createClassV8, getClassShareInfo, ClassShareInfo } from '@/api/class'
import './index.scss'

/** 学科选项 */
const SUBJECT_OPTIONS = ['语文', '数学', '英语', '物理', '化学', '生物', '其他']

/** 学员年龄范畴选项 */
const AGE_GROUP_OPTIONS = ['学前', '小学低年级', '小学高年级', '初中', '高中', '成人']

export default function ClassCreate() {
  const router = useRouter()
  const isFromRegister = router.params.from === 'register'

  const [teacherDisplayName, setTeacherDisplayName] = useState('')
  const [subject, setSubject] = useState('')
  const [ageGroups, setAgeGroups] = useState<string[]>([])
  const [className, setClassName] = useState('')
  const [description, setDescription] = useState('')
  const [submitting, setSubmitting] = useState(false)

  // 创建成功后的分享信息
  const [shareInfo, setShareInfo] = useState<ClassShareInfo | null>(null)
  const [createSuccess, setCreateSuccess] = useState(false)

  /** 表单是否可提交 */
  const canSubmit =
    teacherDisplayName.trim().length >= 1 &&
    subject !== '' &&
    ageGroups.length > 0 &&
    className.trim().length >= 1 &&
    className.trim().length <= 50 &&
    !submitting

  /** 切换年龄范畴选中状态 */
  const toggleAgeGroup = (group: string) => {
    setAgeGroups(prev =>
      prev.includes(group) ? prev.filter(g => g !== group) : [...prev, group]
    )
  }

  /** 提交创建 */
  const handleSubmit = async () => {
    if (!canSubmit) return

    if (!teacherDisplayName.trim()) {
      Taro.showToast({ title: '请输入教师昵称', icon: 'none' })
      return
    }
    if (!subject) {
      Taro.showToast({ title: '请选择学科', icon: 'none' })
      return
    }
    if (ageGroups.length === 0) {
      Taro.showToast({ title: '请选择学员年龄范畴', icon: 'none' })
      return
    }
    if (!className.trim()) {
      Taro.showToast({ title: '请输入班级名称', icon: 'none' })
      return
    }

    setSubmitting(true)
    try {
      const res = await createClassV8({
        teacher_display_name: teacherDisplayName.trim(),
        subject,
        age_group: ageGroups,
        name: className.trim(),
        description: description.trim() || undefined,
      })
      Taro.showToast({ title: '创建成功', icon: 'success' })

      // 获取分享信息
      const classId = res.data?.id
      if (classId) {
        try {
          const shareRes = await getClassShareInfo(classId)
          setShareInfo(shareRes.data)
        } catch (e) {
          console.error('获取分享信息失败:', e)
        }
      }
      setCreateSuccess(true)
    } catch (error) {
      console.error('创建班级失败:', error)
      const errorMsg = error?.message || '创建班级失败'
      Taro.showToast({ title: errorMsg, icon: 'none', duration: 3000 })
    } finally {
      setSubmitting(false)
    }
  }

  /** 复制分享链接 */
  const handleCopyLink = () => {
    if (!shareInfo?.share_link) return
    Taro.setClipboardData({
      data: shareInfo.share_link,
      success: () => {
        Taro.showToast({ title: '链接已复制', icon: 'success' })
      },
    })
  }

  /** 返回首页 */
  const handleGoHome = () => {
    if (isFromRegister) {
      Taro.switchTab({ url: '/pages/home/index' })
    } else {
      Taro.navigateBack()
    }
  }

  // 创建成功后显示分享信息
  if (createSuccess) {
    return (
      <View className='class-create'>
        <View className='class-create__success'>
          <Text className='class-create__success-icon'>🎉</Text>
          <Text className='class-create__success-title'>班级创建成功！</Text>
          <Text className='class-create__success-subtitle'>分享以下信息邀请学生加入</Text>

          {shareInfo && (
            <View className='class-create__share-info'>
              {/* 分享链接 */}
              <View className='class-create__share-item'>
                <Text className='class-create__share-label'>分享链接</Text>
                <Text className='class-create__share-link'>{shareInfo.share_link}</Text>
                <View className='class-create__copy-btn' onClick={handleCopyLink}>
                  <Text className='class-create__copy-btn-text'>复制链接</Text>
                </View>
              </View>

              {/* 邀请码 */}
              {shareInfo.invite_code && (
                <View className='class-create__share-item'>
                  <Text className='class-create__share-label'>邀请码</Text>
                  <Text className='class-create__invite-code'>{shareInfo.invite_code}</Text>
                </View>
              )}

              {/* 二维码 */}
              {shareInfo.qr_code_url && (
                <View className='class-create__share-item class-create__share-item--center'>
                  <Text className='class-create__share-label'>二维码</Text>
                  <Image
                    className='class-create__qrcode'
                    src={shareInfo.qr_code_url}
                    mode='aspectFit'
                  />
                </View>
              )}
            </View>
          )}

          <View className='class-create__submit' onClick={handleGoHome}>
            <Text className='class-create__submit-text'>
              {isFromRegister ? '进入首页' : '返回'}
            </Text>
          </View>
        </View>
      </View>
    )
  }

  return (
    <View className='class-create'>
      <View className='class-create__header'>
        <Text className='class-create__title'>
          {isFromRegister ? '创建你的第一个班级' : '创建班级'}
        </Text>
        <Text className='class-create__subtitle'>
          {isFromRegister ? '创建班级后即可邀请学生加入' : '创建一个班级来管理你的学生'}
        </Text>
      </View>

      {/* 教师昵称 */}
      <View className='class-create__section'>
        <Text className='class-create__label'>教师昵称</Text>
        <Input
          className='class-create__input'
          placeholder='请输入教师昵称'
          placeholderClass='class-create__input-placeholder'
          maxlength={30}
          value={teacherDisplayName}
          onInput={(e) => setTeacherDisplayName(e.detail.value)}
        />
      </View>

      {/* 学科选择 */}
      <View className='class-create__section'>
        <Text className='class-create__label'>学科</Text>
        <View className='class-create__tag-grid'>
          {SUBJECT_OPTIONS.map((item) => (
            <View
              key={item}
              className={`class-create__tag ${subject === item ? 'class-create__tag--active' : ''}`}
              onClick={() => setSubject(item)}
            >
              <Text
                className={`class-create__tag-text ${subject === item ? 'class-create__tag-text--active' : ''}`}
              >
                {item}
              </Text>
            </View>
          ))}
        </View>
      </View>

      {/* 学员年龄范畴（多选） */}
      <View className='class-create__section'>
        <Text className='class-create__label'>学员年龄范畴（可多选）</Text>
        <View className='class-create__tag-grid'>
          {AGE_GROUP_OPTIONS.map((item) => (
            <View
              key={item}
              className={`class-create__tag ${ageGroups.includes(item) ? 'class-create__tag--active' : ''}`}
              onClick={() => toggleAgeGroup(item)}
            >
              <Text
                className={`class-create__tag-text ${ageGroups.includes(item) ? 'class-create__tag-text--active' : ''}`}
              >
                {item}
              </Text>
            </View>
          ))}
        </View>
      </View>

      {/* 班级名称 */}
      <View className='class-create__section'>
        <Text className='class-create__label'>班级名称</Text>
        <Input
          className='class-create__input'
          placeholder='请输入班级名称'
          placeholderClass='class-create__input-placeholder'
          maxlength={50}
          value={className}
          onInput={(e) => setClassName(e.detail.value)}
        />
      </View>

      {/* 简介 */}
      <View className='class-create__section'>
        <Text className='class-create__label'>简介（可选）</Text>
        <Textarea
          className='class-create__textarea'
          placeholder='请输入班级简介'
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
