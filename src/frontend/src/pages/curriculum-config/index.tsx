import { useState, useEffect, useCallback } from 'react'
import { View, Text, Picker, Input } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import {
  createCurriculumConfig,
  getCurriculumConfigs,
  updateCurriculumConfig,
  deleteCurriculumConfig,
  CurriculumConfig,
  GRADE_LEVELS,
  K12_SUBJECTS,
  ADULT_LIFE_CATEGORIES,
  ADULT_PROFESSIONAL_CATEGORIES,
  TEXTBOOK_VERSIONS,
} from '@/api/curriculum'
import './index.scss'

/** 引导步骤编号 */
const STEPS = ['学段与年级', '教材版本', '教学学科', '教学进度']
const ADULT_STEPS = ['学段', '课程类别', '教学进度']

export default function CurriculumConfigPage() {
  const router = useRouter()
  const personaId = Number(router.params.persona_id) || 0

  const [configs, setConfigs] = useState<CurriculumConfig[]>([])
  const [loading, setLoading] = useState(false)

  // 表单状态
  const [selectedLevelIdx, setSelectedLevelIdx] = useState(-1)
  const [selectedGradeIdx, setSelectedGradeIdx] = useState(-1)
  const [selectedSubjects, setSelectedSubjects] = useState<string[]>([])
  const [selectedTextbooks, setSelectedTextbooks] = useState<string[]>([])
  /** 手动填写的教材名称（大学及以上、辅导教材） */
  const [customTextbooks, setCustomTextbooks] = useState<string[]>([])
  const [customTextbookInput, setCustomTextbookInput] = useState('')
  /** 当前教学进度 */
  const [currentProgress, setCurrentProgress] = useState('')
  /** 编辑中的配置ID，为null表示新建 */
  const [editingId, setEditingId] = useState<number | null>(null)

  const selectedLevel = selectedLevelIdx >= 0 ? GRADE_LEVELS[selectedLevelIdx] : null
  const isAdult = selectedLevel?.value === 'adult_life' || selectedLevel?.value === 'adult_professional'
  const isUniversity = selectedLevel?.value === 'university'

  // 当前引导步骤
  const stepLabels = isAdult ? ADULT_STEPS : STEPS
  const currentStep = (() => {
    if (selectedLevelIdx < 0) return 0
    if (isAdult) {
      if (selectedSubjects.length === 0) return 1
      return 2
    }
    // 非成人学段
    if (selectedLevel && selectedLevel.grades.length > 0 && selectedGradeIdx < 0) return 0
    if (selectedTextbooks.length === 0 && customTextbooks.length === 0 && !isUniversity) return 1
    if (selectedSubjects.length === 0) return 2
    return 3
  })()

  // 获取学科选项
  const getSubjectOptions = () => {
    if (!selectedLevel) return K12_SUBJECTS
    if (selectedLevel.value === 'adult_life') return ADULT_LIFE_CATEGORIES
    if (selectedLevel.value === 'adult_professional') return ADULT_PROFESSIONAL_CATEGORIES
    return K12_SUBJECTS
  }

  // 加载配置列表
  const loadConfigs = useCallback(async () => {
    if (!personaId) return
    try {
      const res = await getCurriculumConfigs(personaId)
      setConfigs(res.data?.items || [])
    } catch (e) {
      console.error('加载教材配置失败:', e)
    }
  }, [personaId])

  useEffect(() => {
    loadConfigs()
  }, [loadConfigs])

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

  /** 添加手动填写的教材 */
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

  /** 移除手动填写的教材 */
  const removeCustomTextbook = (name: string) => {
    setCustomTextbooks(prev => prev.filter(t => t !== name))
  }

  /** 重置表单 */
  const resetForm = () => {
    setEditingId(null)
    setSelectedLevelIdx(-1)
    setSelectedGradeIdx(-1)
    setSelectedSubjects([])
    setSelectedTextbooks([])
    setCustomTextbooks([])
    setCustomTextbookInput('')
    setCurrentProgress('')
  }

  /** 进入编辑模式：将已有配置填充到表单 */
  const handleEdit = (cfg: CurriculumConfig) => {
    const levelIdx = GRADE_LEVELS.findIndex(l => l.value === cfg.grade_level)
    setSelectedLevelIdx(levelIdx)
    const level = levelIdx >= 0 ? GRADE_LEVELS[levelIdx] : null
    if (level && cfg.grade) {
      const gradeIdx = level.grades.indexOf(cfg.grade)
      setSelectedGradeIdx(gradeIdx >= 0 ? gradeIdx : -1)
    } else {
      setSelectedGradeIdx(-1)
    }
    setSelectedSubjects(cfg.subjects || [])
    setSelectedTextbooks(cfg.textbook_versions || [])
    setCustomTextbooks((cfg as any).custom_textbooks || [])
    setCurrentProgress(cfg.current_progress || '')
    setEditingId(cfg.id)
  }

  /** 取消编辑 */
  const handleCancelEdit = () => {
    resetForm()
  }

  // 保存配置（新建或更新）
  const handleSave = async () => {
    if (!selectedLevel) {
      Taro.showToast({ title: '请选择学段', icon: 'none' })
      return
    }

    const grade = selectedLevel.grades.length > 0 && selectedGradeIdx >= 0
      ? selectedLevel.grades[selectedGradeIdx]
      : ''

    setLoading(true)
    try {
      const payload = {
        persona_id: personaId,
        grade_level: selectedLevel.value,
        grade,
        textbook_versions: selectedTextbooks,
        custom_textbooks: customTextbooks,
        subjects: selectedSubjects,
        current_progress: currentProgress,
      }

      if (editingId) {
        await updateCurriculumConfig(editingId, payload)
        Taro.showToast({ title: '配置更新成功', icon: 'success' })
      } else {
        await createCurriculumConfig(payload)
        Taro.showToast({ title: '配置保存成功', icon: 'success' })
      }
      resetForm()
      loadConfigs()
    } catch (e: any) {
      Taro.showToast({ title: e?.message || '保存失败', icon: 'none' })
    } finally {
      setLoading(false)
    }
  }

  // 删除配置
  const handleDelete = async (id: number) => {
    const res = await Taro.showModal({ title: '确认删除', content: '确定要删除这个教材配置吗？' })
    if (!res.confirm) return
    try {
      await deleteCurriculumConfig(id)
      Taro.showToast({ title: '删除成功', icon: 'success' })
      loadConfigs()
    } catch (e: any) {
      Taro.showToast({ title: e?.message || '删除失败', icon: 'none' })
    }
  }

  return (
    <View className='curriculum-page'>
      <View className='curriculum-page__header'>
        <Text className='curriculum-page__title'>📚 教材配置</Text>
        <Text className='curriculum-page__subtitle'>配置年级和教材，让AI更精准地辅导</Text>
      </View>

      {/* 引导步骤指示器 */}
      <View className='curriculum-page__steps'>
        {stepLabels.map((label, idx) => (
          <View key={label} className='curriculum-page__step'>
            <View className={`curriculum-page__step-dot ${idx <= currentStep ? 'curriculum-page__step-dot--active' : ''}`}>
              <Text className='curriculum-page__step-dot-text'>{idx + 1}</Text>
            </View>
            <Text className={`curriculum-page__step-label ${idx <= currentStep ? 'curriculum-page__step-label--active' : ''}`}>{label}</Text>
            {idx < stepLabels.length - 1 && (
              <View className={`curriculum-page__step-line ${idx < currentStep ? 'curriculum-page__step-line--active' : ''}`} />
            )}
          </View>
        ))}
      </View>

      {/* 配置表单 */}
      <View className='curriculum-page__form'>
        {/* 学段选择 */}
        <View className='curriculum-page__field'>
          <Text className='curriculum-page__label'>
            <Text className='curriculum-page__label-num'>①</Text> 选择学段
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
            <View className='curriculum-page__picker'>
              <Text className={`curriculum-page__picker-text ${selectedLevel ? 'curriculum-page__picker-text--selected' : ''}`}>
                {selectedLevel ? selectedLevel.label : '请选择学段'}
              </Text>
              <Text className='curriculum-page__picker-arrow'>▸</Text>
            </View>
          </Picker>
        </View>

        {/* 年级选择（非成人学段且有年级列表） */}
        {selectedLevel && selectedLevel.grades.length > 0 && (
          <View className='curriculum-page__field'>
            <Text className='curriculum-page__label'>选择年级</Text>
            <Picker
              mode='selector'
              range={selectedLevel.grades}
              onChange={(e) => setSelectedGradeIdx(Number(e.detail.value))}
            >
              <View className='curriculum-page__picker'>
                <Text className={`curriculum-page__picker-text ${selectedGradeIdx >= 0 ? 'curriculum-page__picker-text--selected' : ''}`}>
                  {selectedGradeIdx >= 0 ? selectedLevel.grades[selectedGradeIdx] : '请选择年级'}
                </Text>
                <Text className='curriculum-page__picker-arrow'>▸</Text>
              </View>
            </Picker>
          </View>
        )}

        {/* 教材版本（非成人学段） */}
        {selectedLevel && !isAdult && (
          <View className='curriculum-page__field'>
            <Text className='curriculum-page__label'>
              <Text className='curriculum-page__label-num'>②</Text>
              {isUniversity ? '教材版本（手动填写）' : '教材版本（可多选）'}
            </Text>
            {/* 大学及以上：仅手动填写 */}
            {isUniversity ? (
              <View className='curriculum-page__custom-input-area'>
                <Text className='curriculum-page__hint'>大学及以上请手动填写使用的教材名称</Text>
                <View className='curriculum-page__custom-input-row'>
                  <Input
                    className='curriculum-page__custom-input'
                    placeholder='输入教材名称，如《高等数学》'
                    value={customTextbookInput}
                    onInput={(e) => setCustomTextbookInput(e.detail.value)}
                    onConfirm={addCustomTextbook}
                  />
                  <View className='curriculum-page__custom-add-btn' onClick={addCustomTextbook}>
                    <Text className='curriculum-page__custom-add-btn-text'>添加</Text>
                  </View>
                </View>
                {customTextbooks.length > 0 && (
                  <View className='curriculum-page__custom-list'>
                    {customTextbooks.map(t => (
                      <View key={t} className='curriculum-page__custom-item'>
                        <Text className='curriculum-page__custom-item-text'>{t}</Text>
                        <Text className='curriculum-page__custom-item-remove' onClick={() => removeCustomTextbook(t)}>✕</Text>
                      </View>
                    ))}
                  </View>
                )}
              </View>
            ) : (
              <>
                {/* K12学段：预设版本多选 */}
                <View className='curriculum-page__tags'>
                  {TEXTBOOK_VERSIONS.map(v => (
                    <View
                      key={v}
                      className={`curriculum-page__tag ${selectedTextbooks.includes(v) ? 'curriculum-page__tag--active' : ''}`}
                      onClick={() => toggleTextbook(v)}
                    >
                      <Text className='curriculum-page__tag-text'>{v}</Text>
                    </View>
                  ))}
                </View>
                {/* 辅导教材手动填写 */}
                <View className='curriculum-page__custom-input-area'>
                  <Text className='curriculum-page__hint'>📖 还可以手动添加辅导教材</Text>
                  <View className='curriculum-page__custom-input-row'>
                    <Input
                      className='curriculum-page__custom-input'
                      placeholder='输入辅导教材名称'
                      value={customTextbookInput}
                      onInput={(e) => setCustomTextbookInput(e.detail.value)}
                      onConfirm={addCustomTextbook}
                    />
                    <View className='curriculum-page__custom-add-btn' onClick={addCustomTextbook}>
                      <Text className='curriculum-page__custom-add-btn-text'>添加</Text>
                    </View>
                  </View>
                  {customTextbooks.length > 0 && (
                    <View className='curriculum-page__custom-list'>
                      {customTextbooks.map(t => (
                        <View key={t} className='curriculum-page__custom-item'>
                          <Text className='curriculum-page__custom-item-text'>{t}</Text>
                          <Text className='curriculum-page__custom-item-remove' onClick={() => removeCustomTextbook(t)}>✕</Text>
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
          <View className='curriculum-page__field'>
            <Text className='curriculum-page__label'>
              <Text className='curriculum-page__label-num'>{isAdult ? '②' : '③'}</Text>
              {isAdult ? '课程类别（可多选）' : '教学学科（可多选）'}
            </Text>
            <View className='curriculum-page__tags'>
              {getSubjectOptions().map(s => (
                <View
                  key={s}
                  className={`curriculum-page__tag ${selectedSubjects.includes(s) ? 'curriculum-page__tag--active' : ''}`}
                  onClick={() => toggleSubject(s)}
                >
                  <Text className='curriculum-page__tag-text'>{s}</Text>
                </View>
              ))}
            </View>
          </View>
        )}

        {/* 教学进度（可选） */}
        {selectedLevel && (
          <View className='curriculum-page__field'>
            <Text className='curriculum-page__label'>
              <Text className='curriculum-page__label-num'>{isAdult ? '③' : '④'}</Text>
              当前教学进度（可选）
            </Text>
            <Input
              className='curriculum-page__progress-input'
              placeholder='如：第三单元 二次函数'
              value={currentProgress}
              onInput={(e) => setCurrentProgress(e.detail.value)}
            />
          </View>
        )}

        {/* 保存/取消按钮 */}
        <View className='curriculum-page__btn-group'>
          {editingId && (
            <View
              className='curriculum-page__cancel-btn'
              onClick={handleCancelEdit}
            >
              <Text className='curriculum-page__cancel-btn-text'>取消编辑</Text>
            </View>
          )}
          <View
            className={`curriculum-page__save-btn ${(!selectedLevel || loading) ? 'curriculum-page__save-btn--disabled' : ''}`}
            onClick={handleSave}
          >
            <Text className='curriculum-page__save-btn-text'>
              {loading ? '保存中...' : (editingId ? '更新配置' : '保存配置')}
            </Text>
          </View>
        </View>
      </View>

      {/* 已有配置列表 */}
      {configs.length > 0 && (
        <View className='curriculum-page__list'>
          <Text className='curriculum-page__list-title'>已有配置</Text>
          {configs.map(cfg => {
            const allTextbooks = [
              ...(cfg.textbook_versions || []),
              ...((cfg as any).custom_textbooks || []),
            ]
            return (
              <View key={cfg.id} className='curriculum-page__list-item'>
                <View className='curriculum-page__list-item-header'>
                  <Text className='curriculum-page__list-item-level'>
                    {GRADE_LEVELS.find(l => l.value === cfg.grade_level)?.label || cfg.grade_level}
                    {cfg.grade ? ` · ${cfg.grade}` : ''}
                  </Text>
                  <View className='curriculum-page__list-item-actions'>
                    <View className='curriculum-page__list-item-edit' onClick={() => handleEdit(cfg)}>
                      <Text className='curriculum-page__list-item-edit-text'>编辑</Text>
                    </View>
                    <View className='curriculum-page__list-item-delete' onClick={() => handleDelete(cfg.id)}>
                      <Text className='curriculum-page__list-item-delete-text'>删除</Text>
                    </View>
                  </View>
                </View>
                {cfg.subjects?.length > 0 && (
                  <Text className='curriculum-page__list-item-info'>
                    {cfg.grade_level?.startsWith('adult') ? '课程' : '学科'}：{cfg.subjects.join('、')}
                  </Text>
                )}
                {allTextbooks.length > 0 && (
                  <Text className='curriculum-page__list-item-info'>教材：{allTextbooks.join('、')}</Text>
                )}
                {cfg.current_progress && (
                  <Text className='curriculum-page__list-item-info'>进度：{cfg.current_progress}</Text>
                )}
              </View>
            )
          })}
        </View>
      )}
    </View>
  )
}
