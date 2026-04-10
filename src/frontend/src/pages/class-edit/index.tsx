import { useState, useEffect } from 'react'
import { View, Text, Input, Textarea, Switch } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { updateClassV11, getClassDetail, type CurriculumConfig } from '@/api/class'
import type { CurriculumConfigFormValue } from '@/components/CurriculumConfigForm'
import CurriculumConfigForm from '@/components/CurriculumConfigForm'
import './index.scss'

export default function ClassEdit() {
  const params = Taro.getCurrentInstance().router?.params || {}
  const classId = Number(params.class_id) || 0

  // 班级信息
  const [className, setClassName] = useState('')
  const [description, setDescription] = useState('')
  const [isPublic, setIsPublic] = useState(true)

  // 教材配置
  const [curriculumConfig, setCurriculumConfig] = useState<CurriculumConfigFormValue | null>(null)
  const [curriculumExpanded, setCurriculumExpanded] = useState(false)
  const [initialCurriculumConfig, setInitialCurriculumConfig] = useState<CurriculumConfig | null>(null)
  // 标记是否需要删除配置（用户点击删除按钮时设置）
  const [shouldDeleteConfig, setShouldDeleteConfig] = useState(false)

  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  // 加载班级详情
  useEffect(() => {
    if (!classId) {
      Taro.showToast({ title: '班级ID不存在', icon: 'none' })
      setTimeout(() => Taro.navigateBack(), 1500)
      return
    }
    loadClassDetail()
  }, [classId])

  /** 加载班级详情 */
  const loadClassDetail = async () => {
    setLoading(true)
    try {
      const res = await getClassDetail(classId)
      const data = res.data
      setClassName(data.name || '')
      setDescription(data.description || '')
      setIsPublic(data.is_public !== undefined ? data.is_public : true)

      // IT13: 加载教材配置，如有配置则展开显示并回填表单
      if (data.curriculum_config) {
        setInitialCurriculumConfig(data.curriculum_config)
        setCurriculumConfig(data.curriculum_config)
        setCurriculumExpanded(true)
        setShouldDeleteConfig(false)
      } else {
        // 无配置时，确保状态被正确重置
        setInitialCurriculumConfig(null)
        setCurriculumConfig(null)
        setCurriculumExpanded(false)
        setShouldDeleteConfig(false)
      }
    } catch (error) {
      console.error('加载班级详情失败:', error)
      Taro.showToast({ title: '加载失败', icon: 'none' })
    } finally {
      setLoading(false)
    }
  }

  /** 表单是否可提交 */
  const canSubmit =
    className.trim().length >= 1 &&
    className.trim().length <= 50 &&
    !submitting

  /**
   * 处理删除教材配置
   * 用户确认删除后调用，标记需要删除配置
   */
  const handleDeleteConfig = () => {
    setShouldDeleteConfig(true)
    setCurriculumConfig(null)
  }

  /** 提交保存 */
  const handleSubmit = async () => {
    if (!canSubmit) return

    if (!className.trim()) {
      Taro.showToast({ title: '请输入班级名称', icon: 'none' })
      return
    }

    setSubmitting(true)
    try {
      // IT13: 构建教材配置参数
      // 1. 如果用户点击了删除配置，传递空对象表示删除
      // 2. 如果用户展开了区域并填写了配置，传递配置
      // 3. 如果用户未展开区域，不传递配置字段（保持原配置不变）
      let curriculumConfigParam: CurriculumConfigFormValue | undefined
      if (shouldDeleteConfig) {
        // 用户选择删除配置，传递空对象通知后端删除
        curriculumConfigParam = {}
      } else if (curriculumExpanded && curriculumConfig?.grade_level) {
        // 用户展开了区域并填写了配置，传递配置
        curriculumConfigParam = curriculumConfig
      }
      // 否则未展开区域，不传递该字段

      await updateClassV11(classId, {
        name: className.trim(),
        description: description.trim() || undefined,
        is_public: isPublic,
        ...(curriculumConfigParam !== undefined
          ? { curriculum_config: curriculumConfigParam }
          : {}),
      })
      Taro.showToast({ title: '保存成功', icon: 'success' })
      setTimeout(() => Taro.navigateBack(), 1500)
    } catch (error) {
      console.error('保存班级失败:', error)
      const errorMsg = error?.message || '保存失败'
      Taro.showToast({ title: errorMsg, icon: 'none', duration: 3000 })
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) {
    return (
      <View className='class-edit'>
        <View className='class-edit__loading'>
          <Text className='class-edit__loading-text'>加载中...</Text>
        </View>
      </View>
    )
  }

  return (
    <View className='class-edit'>
      <View className='class-edit__header'>
        <Text className='class-edit__title'>编辑班级信息</Text>
        <Text className='class-edit__subtitle'>修改班级的基本信息</Text>
      </View>

      {/* 班级名称 */}
      <View className='class-edit__section'>
        <Text className='class-edit__label'>班级名称 <Text className='class-edit__required'>*</Text></Text>
        <Input
          className='class-edit__input'
          placeholder='请输入班级名称'
          placeholderClass='class-edit__input-placeholder'
          maxlength={50}
          value={className}
          onInput={(e) => setClassName(e.detail.value)}
        />
      </View>

      {/* 班级描述 */}
      <View className='class-edit__section'>
        <Text className='class-edit__label'>班级描述（可选）</Text>
        <Textarea
          className='class-edit__textarea'
          placeholder='请输入班级描述'
          maxlength={200}
          value={description}
          onInput={(e) => setDescription(e.detail.value)}
        />
        <Text className='class-edit__char-count'>{description.length}/200</Text>
      </View>

      {/* 公开设置 */}
      <View className='class-edit__section class-edit__section--switch'>
        <View className='class-edit__switch-row'>
          <View className='class-edit__switch-info'>
            <Text className='class-edit__label'>公开班级</Text>
            <Text className='class-edit__switch-desc'>公开班级对所有学生可见，非公开班级仅限受邀学生加入</Text>
          </View>
          <Switch
            checked={isPublic}
            color='#1890ff'
            onChange={(e) => setIsPublic(e.detail.value)}
          />
        </View>
        <View className={`class-edit__status-hint ${isPublic ? 'class-edit__status-hint--public' : 'class-edit__status-hint--private'}`}>
          <Text className='class-edit__status-hint-text'>
            {isPublic ? '当前班级公开，所有学生可见' : '当前班级私密，仅受邀学生可加入'}
          </Text>
        </View>
      </View>

      {/* 教材配置区域（IT13新增） */}
      <CurriculumConfigForm
        expanded={curriculumExpanded}
        onExpandedChange={(expanded) => {
          setCurriculumExpanded(expanded)
          // 用户手动展开/折叠时重置删除标记
          if (!expanded) {
            setShouldDeleteConfig(false)
          }
        }}
        initialValue={initialCurriculumConfig || undefined}
        onChange={(value) => {
          setCurriculumConfig(value)
          // 如果 onChange 返回了值，说明用户修改了配置，重置删除标记
          if (value?.grade_level) {
            setShouldDeleteConfig(false)
          }
        }}
        hasExistingConfig={!!initialCurriculumConfig}
        onDeleteConfig={handleDeleteConfig}
      />

      {/* 保存按钮 */}
      <View
        className={`class-edit__submit ${!canSubmit ? 'class-edit__submit--disabled' : ''}`}
        onClick={canSubmit ? handleSubmit : undefined}
      >
        <Text className='class-edit__submit-text'>
          {submitting ? '保存中...' : '保存'}
        </Text>
      </View>
    </View>
  )
}
