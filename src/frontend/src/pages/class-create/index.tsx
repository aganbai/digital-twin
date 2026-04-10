import { useState } from 'react'
import { View, Text, Input, Textarea, Switch } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import { createClassV11, CreateClassV11Response } from '@/api/class'
import type { CurriculumConfigFormValue } from '@/components/CurriculumConfigForm'
import CurriculumConfigForm from '@/components/CurriculumConfigForm'
import './index.scss'

export default function ClassCreate() {
  const router = useRouter()
  const isFromRegister = router.params.from === 'register'

  // 班级信息
  const [className, setClassName] = useState('')
  const [description, setDescription] = useState('')
  const [isPublic, setIsPublic] = useState(true)

  // 分身信息
  const [personaNickname, setPersonaNickname] = useState('')
  const [personaSchool, setPersonaSchool] = useState('')
  const [personaDescription, setPersonaDescription] = useState('')

  // 教材配置
  const [curriculumConfig, setCurriculumConfig] = useState<CurriculumConfigFormValue>({})
  const [curriculumExpanded, setCurriculumExpanded] = useState(false)

  const [submitting, setSubmitting] = useState(false)

  // 创建成功后的班级信息
  const [classInfo, setClassInfo] = useState<CreateClassV11Response | null>(null)
  const [createSuccess, setCreateSuccess] = useState(false)

  /** 表单是否可提交 */
  const canSubmit =
    className.trim().length >= 1 &&
    className.trim().length <= 50 &&
    personaNickname.trim().length >= 1 &&
    personaSchool.trim().length >= 1 &&
    personaDescription.trim().length >= 1 &&
    !submitting

  /** 提交创建 */
  const handleSubmit = async () => {
    if (!canSubmit) return

    if (!className.trim()) {
      Taro.showToast({ title: '请输入班级名称', icon: 'none' })
      return
    }
    if (!personaNickname.trim()) {
      Taro.showToast({ title: '请输入分身昵称', icon: 'none' })
      return
    }
    if (!personaSchool.trim()) {
      Taro.showToast({ title: '请输入学校名称', icon: 'none' })
      return
    }
    if (!personaDescription.trim()) {
      Taro.showToast({ title: '请输入分身描述', icon: 'none' })
      return
    }

    setSubmitting(true)
    try {
      const res = await createClassV11({
        name: className.trim(),
        description: description.trim() || undefined,
        persona_nickname: personaNickname.trim(),
        persona_school: personaSchool.trim(),
        persona_description: personaDescription.trim(),
        is_public: isPublic,
        // IT13: 仅当展开并填写了学段时才传递教材配置
        ...(curriculumExpanded && curriculumConfig.grade_level
          ? { curriculum_config: curriculumConfig }
          : {}),
      })
      Taro.showToast({ title: '创建成功', icon: 'success' })
      setClassInfo(res.data)
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
    if (!classInfo?.share_url) return
    Taro.setClipboardData({
      data: classInfo.share_url,
      success: () => {
        Taro.showToast({ title: '链接已复制', icon: 'success' })
      },
    })
  }

  /** 复制分享码 */
  const handleCopyCode = () => {
    if (!classInfo?.share_code) return
    Taro.setClipboardData({
      data: classInfo.share_code,
      success: () => {
        Taro.showToast({ title: '分享码已复制', icon: 'success' })
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

  // 创建成功后显示分身信息和分享信息
  if (createSuccess && classInfo) {
    return (
      <View className='class-create'>
        <View className='class-create__success'>
          <Text className='class-create__success-icon'>🎉</Text>
          <Text className='class-create__success-title'>班级创建成功！</Text>
          <Text className='class-create__success-subtitle'>已同步创建班级专属分身</Text>

          {/* 分身信息卡片 */}
          <View className='class-create__persona-card'>
            <Text className='class-create__persona-title'>班级分身信息</Text>
            <View className='class-create__persona-info'>
              <View className='class-create__persona-row'>
                <Text className='class-create__persona-label'>分身昵称</Text>
                <Text className='class-create__persona-value'>{classInfo.persona_nickname}</Text>
              </View>
              <View className='class-create__persona-row'>
                <Text className='class-create__persona-label'>分身ID</Text>
                <Text className='class-create__persona-value'>{classInfo.persona_id}</Text>
              </View>
              <View className='class-create__persona-row'>
                <Text className='class-create__persona-label'>所属学校</Text>
                <Text className='class-create__persona-value'>{classInfo.persona_school}</Text>
              </View>
            </View>
          </View>

          {/* 分享信息 */}
          <View className='class-create__share-info'>
            <View className='class-create__share-item'>
              <Text className='class-create__share-label'>分享链接</Text>
              <Text className='class-create__share-link'>{classInfo.share_url}</Text>
              <View className='class-create__copy-btn' onClick={handleCopyLink}>
                <Text className='class-create__copy-btn-text'>复制链接</Text>
              </View>
            </View>

            <View className='class-create__share-item'>
              <Text className='class-create__share-label'>分享码</Text>
              <Text className='class-create__invite-code'>{classInfo.share_code}</Text>
              <View className='class-create__copy-btn' onClick={handleCopyCode}>
                <Text className='class-create__copy-btn-text'>复制分享码</Text>
              </View>
            </View>
          </View>

          {/* 引导提示 */}
          <View className='class-create__guide'>
            <Text className='class-create__guide-text'>💡 将分享链接或分享码发给学生，即可邀请他们加入班级</Text>
          </View>

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

      {/* 分身信息区域 */}
      <View className='class-create__section-header'>
        <Text className='class-create__section-title'>分身信息</Text>
        <Text className='class-create__section-desc'>创建班级时会同步创建该班级专属的分身</Text>
      </View>

      {/* 分身昵称 */}
      <View className='class-create__section'>
        <Text className='class-create__label'>分身昵称 <Text className='class-create__required'>*</Text></Text>
        <Input
          className='class-create__input'
          placeholder='请输入分身昵称（如：王老师）'
          placeholderClass='class-create__input-placeholder'
          maxlength={30}
          value={personaNickname}
          onInput={(e) => setPersonaNickname(e.detail.value)}
        />
      </View>

      {/* 分身学校 */}
      <View className='class-create__section'>
        <Text className='class-create__label'>学校名称 <Text className='class-create__required'>*</Text></Text>
        <Input
          className='class-create__input'
          placeholder='请输入学校名称'
          placeholderClass='class-create__input-placeholder'
          maxlength={50}
          value={personaSchool}
          onInput={(e) => setPersonaSchool(e.detail.value)}
        />
      </View>

      {/* 分身描述 */}
      <View className='class-create__section'>
        <Text className='class-create__label'>分身描述 <Text className='class-create__required'>*</Text></Text>
        <Textarea
          className='class-create__textarea'
          placeholder='请输入分身描述（教学风格、擅长领域等）'
          maxlength={200}
          value={personaDescription}
          onInput={(e) => setPersonaDescription(e.detail.value)}
        />
        <Text className='class-create__char-count'>{personaDescription.length}/200</Text>
      </View>

      {/* 班级信息区域 */}
      <View className='class-create__section-header'>
        <Text className='class-create__section-title'>班级信息</Text>
      </View>

      {/* 班级名称 */}
      <View className='class-create__section'>
        <Text className='class-create__label'>班级名称 <Text className='class-create__required'>*</Text></Text>
        <Input
          className='class-create__input'
          placeholder='请输入班级名称'
          placeholderClass='class-create__input-placeholder'
          maxlength={50}
          value={className}
          onInput={(e) => setClassName(e.detail.value)}
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

      {/* 公开设置 */}
      <View className='class-create__section class-create__section--switch'>
        <View className='class-create__switch-row'>
          <View className='class-create__switch-info'>
            <Text className='class-create__label'>公开班级</Text>
            <Text className='class-create__switch-desc'>公开班级对所有学生可见，非公开班级仅限受邀学生加入</Text>
          </View>
          <Switch
            checked={isPublic}
            color='#1890ff'
            onChange={(e) => setIsPublic(e.detail.value)}
          />
        </View>
        <View className={`class-create__status-hint ${isPublic ? 'class-create__status-hint--public' : 'class-create__status-hint--private'}`}>
          <Text className='class-create__status-hint-text'>
            {isPublic ? '当前班级公开，所有学生可见' : '当前班级私密，仅受邀学生可加入'}
          </Text>
        </View>
      </View>

      {/* 教材配置区域（IT13新增） */}
      <CurriculumConfigForm
        expanded={curriculumExpanded}
        onExpandedChange={setCurriculumExpanded}
        onChange={setCurriculumConfig}
      />

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
