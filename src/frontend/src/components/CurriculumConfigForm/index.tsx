import { useState, useEffect } from 'react'
import { View, Text, Picker, Input } from '@tarojs/components'
import Taro from '@tarojs/taro'
import type { GradeLevel } from '@/constants/curriculum'
import {
  GRADE_LEVELS,
  K12_SUBJECTS,
  ADULT_LIFE_CATEGORIES,
  ADULT_PROFESSIONAL_CATEGORIES,
  TEXTBOOK_VERSIONS,
} from '@/api/curriculum'
import './index.scss'

/** 教材配置表单值 */
export interface CurriculumConfigFormValue {
  grade_level?: GradeLevel
  grade?: string
  subjects?: string[]
  textbook_versions?: string[]
  custom_textbooks?: string[]
  current_progress?: string
}

/** 组件Props */
interface CurriculumConfigFormProps {
  /** 初始值（用于编辑时回填） */
  initialValue?: CurriculumConfigFormValue | null
  /** 值变化回调 */
  onChange?: (value: CurriculumConfigFormValue | null) => void
  /** 是否展开显示 */
  expanded?: boolean
  /** 展开状态变化回调 */
  onExpandedChange?: (expanded: boolean) => void
  /** 是否已有配置（用于编辑时判断） */
  hasExistingConfig?: boolean
  /** 删除配置回调（仅编辑时有效） */
  onDeleteConfig?: () => void
}

/** 教材配置表单组件（可与班级创建/编辑集成） */
export default function CurriculumConfigForm({
  initialValue,
  onChange,
  expanded: controlledExpanded,
  onExpandedChange,
  hasExistingConfig,
  onDeleteConfig,
}: CurriculumConfigFormProps) {
  // 内部展开状态（非受控模式）
  const [internalExpanded, setInternalExpanded] = useState(false)
  const isExpanded = controlledExpanded !== undefined ? controlledExpanded : internalExpanded

  // 表单状态
  const [selectedLevelIdx, setSelectedLevelIdx] = useState(-1)
  const [selectedGradeIdx, setSelectedGradeIdx] = useState(-1)
  const [selectedSubjects, setSelectedSubjects] = useState<string[]>([])
  const [selectedTextbooks, setSelectedTextbooks] = useState<string[]>([])
  const [customTextbooks, setCustomTextbooks] = useState<string[]>([])
  const [customTextbookInput, setCustomTextbookInput] = useState('')
  const [currentProgress, setCurrentProgress] = useState('')

  const selectedLevel = selectedLevelIdx >= 0 ? GRADE_LEVELS[selectedLevelIdx] : null
  const isAdult = selectedLevel?.value === 'adult_life' || selectedLevel?.value === 'adult_professional'
  const isUniversity = selectedLevel?.value === 'university'

  // 获取学科选项
  const getSubjectOptions = () => {
    if (!selectedLevel) return K12_SUBJECTS
    if (selectedLevel.value === 'adult_life') return ADULT_LIFE_CATEGORIES
    if (selectedLevel.value === 'adult_professional') return ADULT_PROFESSIONAL_CATEGORIES
    return K12_SUBJECTS
  }

  // 初始值回填（编辑时）
  useEffect(() => {
    if (initialValue && initialValue.grade_level) {
      const levelIdx = GRADE_LEVELS.findIndex(l => l.value === initialValue.grade_level)
      if (levelIdx >= 0) {
        setSelectedLevelIdx(levelIdx)
        const level = GRADE_LEVELS[levelIdx]
        if (level && initialValue.grade) {
          const gradeIdx = level.grades.indexOf(initialValue.grade)
          setSelectedGradeIdx(gradeIdx >= 0 ? gradeIdx : -1)
        }
      }
      setSelectedSubjects(initialValue.subjects || [])
      setSelectedTextbooks(initialValue.textbook_versions || [])
      setCustomTextbooks(initialValue.custom_textbooks || [])
      setCurrentProgress(initialValue.current_progress || '')
    }
  }, [initialValue])

  // 表单值变化时回调
  useEffect(() => {
    // 如果展开但没有选择学段，表示用户想删除配置
    if (isExpanded && !selectedLevel) {
      onChange?.(hasExistingConfig ? {} : undefined)
      return
    }

    const value: CurriculumConfigFormValue = {}
    if (selectedLevel) {
      value.grade_level = selectedLevel.value
      if (selectedLevel.grades.length > 0 && selectedGradeIdx >= 0) {
        value.grade = selectedLevel.grades[selectedGradeIdx]
      }
    }
    if (selectedSubjects.length > 0) {
      value.subjects = selectedSubjects
    }
    if (selectedTextbooks.length > 0) {
      value.textbook_versions = selectedTextbooks
    }
    if (customTextbooks.length > 0) {
      value.custom_textbooks = customTextbooks
    }
    if (currentProgress) {
      value.current_progress = currentProgress
    }
    onChange?.(value)
  }, [selectedLevelIdx, selectedGradeIdx, selectedSubjects, selectedTextbooks, customTextbooks, currentProgress, isExpanded, hasExistingConfig])

  // 切换展开状态
  const toggleExpanded = () => {
    const newExpanded = !isExpanded
    if (controlledExpanded === undefined) {
      setInternalExpanded(newExpanded)
    }
    onExpandedChange?.(newExpanded)
  }

  // 切换学科选择
  const toggleSubject = (subject: string) => {
    setSelectedSubjects(prev =>
      prev.includes(subject) ? prev.filter(s => s !== subject) : [...prev, subject]
    )
  }

  // 切换教材版本选择
  const toggleTextbook = (version: string) => {
    setSelectedTextbooks(prev =>
      prev.includes(version) ? prev.filter(v => v !== version) : [...prev, version]
    )
  }

  // 添加手动填写的教材
  const addCustomTextbook = () => {
    const name = customTextbookInput.trim()
    if (!name) {
      Taro.showToast({ title: '请输入教材名称', icon: 'none' })
      return
    }
    if (customTextbooks.includes(name)) {
      Taro.showToast({ title: '该教材已添加', icon: 'none' })
      return
    }
    setCustomTextbooks(prev => [...prev, name])
    setCustomTextbookInput('')
  }

  // 移除手动填写的教材
  const removeCustomTextbook = (name: string) => {
    setCustomTextbooks(prev => prev.filter(t => t !== name))
  }

  // 处理删除配置
  const handleDeleteConfig = () => {
    Taro.showModal({
      title: '确认删除',
      content: '确定要删除教材配置吗？删除后可重新添加。',
      success: (res) => {
        if (res.confirm) {
          // 重置表单状态
          setSelectedLevelIdx(-1)
          setSelectedGradeIdx(-1)
          setSelectedSubjects([])
          setSelectedTextbooks([])
          setCustomTextbooks([])
          setCustomTextbookInput('')
          setCurrentProgress('')
          // 通知父组件删除配置
          onDeleteConfig?.()
        }
      }
    })
  }

  return (
    <View className='curriculum-form'>
      {/* 折叠面板头部 */}
      <View className='curriculum-form__header' onClick={toggleExpanded}>
        <View className='curriculum-form__title-row'>
          <Text className='curriculum-form__title'>
            {hasExistingConfig ? '📚 教材配置' : '📚 教材配置（可选）'}
          </Text>
          <Text className={`curriculum-form__arrow ${isExpanded ? 'curriculum-form__arrow--expanded' : ''}`}>▸</Text>
        </View>
        <Text className='curriculum-form__subtitle'>
          {isExpanded
            ? '配置年级和教材，让AI更精准地辅导'
            : hasExistingConfig
              ? '已配置教材信息，点击展开查看'
              : '点击添加教材配置信息'}
        </Text>
      </View>

      {/* 展开的配置表单 */}
      {isExpanded && (
        <View className='curriculum-form__content'>
          {/* 已有配置时显示删除按钮 */}
          {hasExistingConfig && (
            <View className='curriculum-form__delete-row'>
              <Text className='curriculum-form__delete-hint'>已保存教材配置</Text>
              <Text className='curriculum-form__delete-btn' onClick={handleDeleteConfig}>删除配置</Text>
            </View>
          )}

          {/* 学段选择 */}
          <View className='curriculum-form__field'>
            <Text className='curriculum-form__label'>
              选择学段 <Text className='curriculum-form__required'>*</Text>
            </Text>
            <Picker
              mode='selector'
              range={GRADE_LEVELS.map(l => l.label)}
              onChange={(e) => {
                const idx = Number(e.detail.value)
                setSelectedLevelIdx(idx)
                setSelectedGradeIdx(-1)
                setSelectedSubjects([])
                setSelectedTextbooks([])
                setCustomTextbooks([])
                setCustomTextbookInput('')
                setCurrentProgress('')
              }}
            >
              <View className='curriculum-form__picker'>
                <Text className={`curriculum-form__picker-text ${selectedLevel ? 'curriculum-form__picker-text--selected' : ''}`}>
                  {selectedLevel ? selectedLevel.label : '请选择学段'}
                </Text>
                <Text className='curriculum-form__picker-arrow'>▸</Text>
              </View>
            </Picker>
          </View>

          {/* 年级选择（非成人学段且有年级列表） */}
          {selectedLevel && selectedLevel.grades.length > 0 && (
            <View className='curriculum-form__field'>
              <Text className='curriculum-form__label'>
                选择年级 <Text className='curriculum-form__required'>*</Text>
              </Text>
              <Picker
                mode='selector'
                range={selectedLevel.grades}
                onChange={(e) => setSelectedGradeIdx(Number(e.detail.value))}
              >
                <View className='curriculum-form__picker'>
                  <Text className={`curriculum-form__picker-text ${selectedGradeIdx >= 0 ? 'curriculum-form__picker-text--selected' : ''}`}>
                    {selectedGradeIdx >= 0 ? selectedLevel.grades[selectedGradeIdx] : '请选择年级'}
                  </Text>
                  <Text className='curriculum-form__picker-arrow'>▸</Text>
                </View>
              </Picker>
            </View>
          )}

          {/* 教材版本（非成人学段） */}
          {selectedLevel && !isAdult && (
            <View className='curriculum-form__field'>
              <Text className='curriculum-form__label'>
                {isUniversity ? '教材版本（手动填写）' : '教材版本（可多选）'}
              </Text>
              {/* 大学及以上：仅手动填写 */}
              {isUniversity ? (
                <View className='curriculum-form__custom-area'>
                  <Text className='curriculum-form__hint'>大学及以上请手动填写使用的教材名称</Text>
                  <View className='curriculum-form__custom-row'>
                    <Input
                      className='curriculum-form__custom-input'
                      placeholder='输入教材名称，如《高等数学》'
                      value={customTextbookInput}
                      onInput={(e) => setCustomTextbookInput(e.detail.value)}
                      onConfirm={addCustomTextbook}
                    />
                    <View className='curriculum-form__custom-add' onClick={addCustomTextbook}>
                      <Text className='curriculum-form__custom-add-text'>添加</Text>
                    </View>
                  </View>
                  {customTextbooks.length > 0 && (
                    <View className='curriculum-form__custom-list'>
                      {customTextbooks.map(t => (
                        <View key={t} className='curriculum-form__custom-tag'>
                          <Text className='curriculum-form__custom-tag-text'>{t}</Text>
                          <Text className='curriculum-form__custom-tag-remove' onClick={() => removeCustomTextbook(t)}>✕</Text>
                        </View>
                      ))}
                    </View>
                  )}
                </View>
              ) : (
                <>
                  {/* K12学段：预设版本多选 */}
                  <View className='curriculum-form__tags'>
                    {TEXTBOOK_VERSIONS.map(v => (
                      <View
                        key={v}
                        className={`curriculum-form__tag ${selectedTextbooks.includes(v) ? 'curriculum-form__tag--active' : ''}`}
                        onClick={() => toggleTextbook(v)}
                      >
                        <Text className='curriculum-form__tag-text'>{v}</Text>
                      </View>
                    ))}
                  </View>
                  {/* 辅导教材手动填写 */}
                  <View className='curriculum-form__custom-area'>
                    <Text className='curriculum-form__hint'>📖 还可以手动添加辅导教材</Text>
                    <View className='curriculum-form__custom-row'>
                      <Input
                        className='curriculum-form__custom-input'
                        placeholder='输入辅导教材名称'
                        value={customTextbookInput}
                        onInput={(e) => setCustomTextbookInput(e.detail.value)}
                        onConfirm={addCustomTextbook}
                      />
                      <View className='curriculum-form__custom-add' onClick={addCustomTextbook}>
                        <Text className='curriculum-form__custom-add-text'>添加</Text>
                      </View>
                    </View>
                    {customTextbooks.length > 0 && (
                      <View className='curriculum-form__custom-list'>
                        {customTextbooks.map(t => (
                          <View key={t} className='curriculum-form__custom-tag'>
                            <Text className='curriculum-form__custom-tag-text'>{t}</Text>
                            <Text className='curriculum-form__custom-tag-remove' onClick={() => removeCustomTextbook(t)}>✕</Text>
                          </View>
                        ))}
                      </View>
                    )}
                  </View>
                </>
              )}
            </View>
          )}

          {/* 学科/课程类别选择 */}
          {selectedLevel && (
            <View className='curriculum-form__field'>
              <Text className='curriculum-form__label'>
                {isAdult ? '课程类别（可多选）' : '教学学科（可多选）'}
                <Text className='curriculum-form__required'>*</Text>
              </Text>
              <View className='curriculum-form__tags'>
                {getSubjectOptions().map(s => (
                  <View
                    key={s}
                    className={`curriculum-form__tag ${selectedSubjects.includes(s) ? 'curriculum-form__tag--active' : ''}`}
                    onClick={() => toggleSubject(s)}
                  >
                    <Text className='curriculum-form__tag-text'>{s}</Text>
                  </View>
                ))}
              </View>
            </View>
          )}

          {/* 教学进度（可选） */}
          {selectedLevel && (
            <View className='curriculum-form__field'>
              <Text className='curriculum-form__label'>当前教学进度（可选）</Text>
              <Input
                className='curriculum-form__input'
                placeholder='如：第三单元 二次函数'
                value={currentProgress}
                onInput={(e) => setCurrentProgress(e.detail.value)}
              />
            </View>
          )}
        </View>
      )}
    </View>
  )
}
