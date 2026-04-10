import { useState, useMemo, useCallback, useEffect } from 'react'
import Taro from '@tarojs/taro'
import type { GradeLevel } from '@/constants/curriculum'
import {
  GRADE_LEVEL_OPTIONS,
  GRADE_MAP,
  getGradesByLevel,
  needsGradeSelection,
  needsCustomTextbook,
  getSubjectsByLevel,
  K12_SUBJECTS,
  EMPTY_CURRICULUM_CONFIG,
  type CurriculumConfigForm,
} from '@/constants/curriculum'
import type { CurriculumConfig } from '@/api/class'

/** 表单验证错误 */
export interface CurriculumConfigErrors {
  grade_level?: string
  grade?: string
  subjects?: string
  textbook_versions?: string
  custom_textbooks?: string
  current_progress?: string
}

/** useCurriculumConfig Hook 配置选项 */
export interface UseCurriculumConfigOptions {
  /** 初始值（用于编辑场景） */
  initialValue?: CurriculumConfig | null
  /** 默认展开状态 */
  defaultExpanded?: boolean
  /** 值变化回调 */
  onChange?: (value: CurriculumConfigForm) => void
  /** 展开状态变化回调 */
  onExpandedChange?: (expanded: boolean) => void
}

/** useCurriculumConfig Hook 返回值 */
export interface UseCurriculumConfigReturn {
  // === 状态 ===
  /** 当前表单值 */
  formValue: CurriculumConfigForm
  /** 是否展开 */
  expanded: boolean
  /** 当前选中的学段索引 */
  selectedLevelIdx: number
  /** 当前选中的年级索引 */
  selectedGradeIdx: number
  /** 已选学科 */
  selectedSubjects: string[]
  /** 已选教材版本 */
  selectedTextbooks: string[]
  /** 自定义教材列表 */
  customTextbooks: string[]
  /** 教学进度 */
  currentProgress: string
  /** 当前学段 */
  currentLevel: GradeLevel | undefined
  /** 表单错误 */
  errors: CurriculumConfigErrors

  // === 计算属性 ===
  /** 是否显示年级选择 */
  showGradeSelector: boolean
  /** 是否使用自定义教材（大学及以上） */
  useCustomTextbookMode: boolean
  /** 是否成人学段 */
  isAdultLevel: boolean
  /** 年级选项列表 */
  gradeOptions: string[]
  /** 学科选项列表 */
  subjectOptions: string[]
  /** 学段选项列表 */
  levelOptions: typeof GRADE_LEVEL_OPTIONS
  /** 是否已填写有效配置 */
  hasValidConfig: boolean
  /** 表单是否有效 */
  isValid: boolean

  // === 操作函数 ===
  /** 切换展开/折叠 */
  toggleExpanded: () => void
  /** 设置展开状态 */
  setExpanded: (expanded: boolean) => void
  /** 选择学段 */
  selectLevel: (index: number) => void
  /** 选择年级 */
  selectGrade: (index: number) => void
  /** 切换学科选择 */
  toggleSubject: (subject: string) => void
  /** 切换教材版本选择 */
  toggleTextbook: (version: string) => void
  /** 添加自定义教材 */
  addCustomTextbook: (name: string) => boolean
  /** 移除自定义教材 */
  removeCustomTextbook: (name: string) => void
  /** 设置教学进度 */
  setCurrentProgress: (progress: string) => void
  /** 验证表单 */
  validate: () => boolean
  /** 清空表单 */
  reset: () => void
  /** 重置为初始值 */
  resetToInitial: () => void
  /** 获取提交用的配置值 */
  getSubmitValue: () => CurriculumConfigForm | null
}

/**
 * 教材配置状态管理 Hook
 *
 * 封装教材配置的状态管理、学段年级联动逻辑和表单验证
 *
 * @example
 * // 创建班级场景
 * const config = useCurriculumConfig()
 *
 * // 编辑班级场景
 * const config = useCurriculumConfig({
 *   initialValue: classDetail.curriculum_config,
 *   defaultExpanded: !!classDetail.curriculum_config
 * })
 *
 * // 提交时获取配置值
 * const submitValue = config.getSubmitValue()
 * if (config.expanded && submitValue) {
 *   params.curriculum_config = submitValue
 * }
 */
export function useCurriculumConfig(
  options: UseCurriculumConfigOptions = {}
): UseCurriculumConfigReturn {
  const {
    initialValue,
    defaultExpanded = false,
    onChange,
    onExpandedChange,
  } = options

  // === 基础状态 ===
  const [expanded, setInternalExpanded] = useState(defaultExpanded)
  const [selectedLevelIdx, setSelectedLevelIdx] = useState(-1)
  const [selectedGradeIdx, setSelectedGradeIdx] = useState(-1)
  const [selectedSubjects, setSelectedSubjects] = useState<string[]>([])
  const [selectedTextbooks, setSelectedTextbooks] = useState<string[]>([])
  const [customTextbooks, setCustomTextbooks] = useState<string[]>([])
  const [currentProgress, setCurrentProgressState] = useState('')
  const [errors, setErrors] = useState<CurriculumConfigErrors>({})

  // === 计算属性 ===
  const currentLevel = useMemo<GradeLevel | undefined>(() => {
    return selectedLevelIdx >= 0 ? GRADE_LEVEL_OPTIONS[selectedLevelIdx].value : undefined
  }, [selectedLevelIdx])

  const showGradeSelector = useMemo(() => {
    return currentLevel ? needsGradeSelection(currentLevel) : false
  }, [currentLevel])

  const useCustomTextbookMode = useMemo(() => {
    return currentLevel ? needsCustomTextbook(currentLevel) : false
  }, [currentLevel])

  const isAdultLevel = useMemo(() => {
    return currentLevel === 'adult_life' || currentLevel === 'adult_professional'
  }, [currentLevel])

  const gradeOptions = useMemo(() => {
    return currentLevel ? getGradesByLevel(currentLevel) : []
  }, [currentLevel])

  const subjectOptions = useMemo(() => {
    return currentLevel ? getSubjectsByLevel(currentLevel) : K12_SUBJECTS
  }, [currentLevel])

  const levelOptions = GRADE_LEVEL_OPTIONS

  const hasValidConfig = useMemo(() => {
    return !!currentLevel && selectedSubjects.length > 0
  }, [currentLevel, selectedSubjects])

  const isValid = useMemo(() => {
    if (!expanded) return true
    if (!currentLevel) return false
    if (showGradeSelector && selectedGradeIdx < 0) return false
    if (selectedSubjects.length === 0) return false
    return Object.keys(errors).length === 0
  }, [expanded, currentLevel, showGradeSelector, selectedGradeIdx, selectedSubjects, errors])

  const formValue = useMemo<CurriculumConfigForm>(() => ({
    grade_level: currentLevel,
    grade: showGradeSelector && selectedGradeIdx >= 0 ? gradeOptions[selectedGradeIdx] : undefined,
    subjects: selectedSubjects,
    textbook_versions: selectedTextbooks,
    custom_textbooks: customTextbooks,
    current_progress: currentProgress || undefined,
  }), [currentLevel, showGradeSelector, selectedGradeIdx, gradeOptions, selectedSubjects,
      selectedTextbooks, customTextbooks, currentProgress])

  // === 副作用 ===
  // 初始值回填
  useEffect(() => {
    if (initialValue && initialValue.grade_level) {
      const levelIdx = GRADE_LEVEL_OPTIONS.findIndex(l => l.value === initialValue.grade_level)
      if (levelIdx >= 0) {
        setSelectedLevelIdx(levelIdx)
        const level = GRADE_LEVEL_OPTIONS[levelIdx].value
        const grades = GRADE_MAP[level]
        if (grades && initialValue.grade) {
          const gradeIdx = grades.indexOf(initialValue.grade)
          setSelectedGradeIdx(gradeIdx >= 0 ? gradeIdx : -1)
        }
      }
      setSelectedSubjects(initialValue.subjects || [])
      setSelectedTextbooks(initialValue.textbook_versions || [])
      setCustomTextbooks(initialValue.custom_textbooks || [])
      setCurrentProgressState(initialValue.current_progress || '')
    }
  }, [initialValue])

  // 值变化回调
  useEffect(() => {
    onChange?.(formValue)
  }, [formValue, onChange])

  // === 操作函数 ===
  const setExpanded = useCallback((newExpanded: boolean) => {
    setInternalExpanded(newExpanded)
    onExpandedChange?.(newExpanded)
  }, [onExpandedChange])

  const toggleExpanded = useCallback(() => {
    const newExpanded = !expanded
    setInternalExpanded(newExpanded)
    onExpandedChange?.(newExpanded)
  }, [expanded, onExpandedChange])

  const selectLevel = useCallback((index: number) => {
    setSelectedLevelIdx(index)
    setSelectedGradeIdx(-1)
    setSelectedSubjects([])
    setSelectedTextbooks([])
    setCustomTextbooks([])
    setErrors(prev => ({ ...prev, grade_level: undefined, grade: undefined }))
  }, [])

  const selectGrade = useCallback((index: number) => {
    setSelectedGradeIdx(index)
    setErrors(prev => ({ ...prev, grade: undefined }))
  }, [])

  const toggleSubject = useCallback((subject: string) => {
    setSelectedSubjects(prev => {
      const newSubjects = prev.includes(subject)
        ? prev.filter(s => s !== subject)
        : [...prev, subject]
      return newSubjects
    })
    setErrors(prev => ({ ...prev, subjects: undefined }))
  }, [])

  const toggleTextbook = useCallback((version: string) => {
    setSelectedTextbooks(prev =>
      prev.includes(version)
        ? prev.filter(v => v !== version)
        : [...prev, version]
    )
  }, [])

  const addCustomTextbook = useCallback((name: string): boolean => {
    const trimmedName = name.trim()
    if (!trimmedName) {
      Taro.showToast({ title: '请输入教材名称', icon: 'none' })
      return false
    }
    if (customTextbooks.includes(trimmedName)) {
      Taro.showToast({ title: '该教材已添加', icon: 'none' })
      return false
    }
    setCustomTextbooks(prev => [...prev, trimmedName])
    return true
  }, [customTextbooks])

  const removeCustomTextbook = useCallback((name: string) => {
    setCustomTextbooks(prev => prev.filter(t => t !== name))
  }, [])

  const setCurrentProgress = useCallback((progress: string) => {
    setCurrentProgressState(progress)
  }, [])

  const validate = useCallback((): boolean => {
    if (!expanded) return true

    const newErrors: CurriculumConfigErrors = {}

    // 学段必填
    if (!currentLevel) {
      newErrors.grade_level = '请选择学段'
    }

    // 年级必填（有年级选项的学段）
    if (showGradeSelector && selectedGradeIdx < 0) {
      newErrors.grade = '请选择年级'
    }

    // 学科必填
    if (selectedSubjects.length === 0) {
      newErrors.subjects = isAdultLevel ? '请选择课程类别' : '请选择教学学科'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }, [expanded, currentLevel, showGradeSelector, selectedGradeIdx, selectedSubjects, isAdultLevel])

  const reset = useCallback(() => {
    setSelectedLevelIdx(-1)
    setSelectedGradeIdx(-1)
    setSelectedSubjects([])
    setSelectedTextbooks([])
    setCustomTextbooks([])
    setCurrentProgressState('')
    setErrors({})
  }, [])

  const resetToInitial = useCallback(() => {
    if (initialValue && initialValue.grade_level) {
      const levelIdx = GRADE_LEVEL_OPTIONS.findIndex(l => l.value === initialValue.grade_level)
      if (levelIdx >= 0) {
        setSelectedLevelIdx(levelIdx)
        const level = GRADE_LEVEL_OPTIONS[levelIdx].value
        const grades = GRADE_MAP[level]
        if (grades && initialValue.grade) {
          const gradeIdx = grades.indexOf(initialValue.grade)
          setSelectedGradeIdx(gradeIdx >= 0 ? gradeIdx : -1)
        } else {
          setSelectedGradeIdx(-1)
        }
      }
      setSelectedSubjects(initialValue.subjects || [])
      setSelectedTextbooks(initialValue.textbook_versions || [])
      setCustomTextbooks(initialValue.custom_textbooks || [])
      setCurrentProgressState(initialValue.current_progress || '')
    } else {
      reset()
    }
    setErrors({})
  }, [initialValue, reset])

  const getSubmitValue = useCallback((): CurriculumConfigForm | null => {
    if (!expanded || !currentLevel) return null

    // 验证不通过返回null
    if (!validate()) return null

    return {
      grade_level: currentLevel,
      grade: showGradeSelector && selectedGradeIdx >= 0
        ? gradeOptions[selectedGradeIdx]
        : undefined,
      subjects: selectedSubjects,
      textbook_versions: selectedTextbooks.length > 0 ? selectedTextbooks : undefined,
      custom_textbooks: customTextbooks.length > 0 ? customTextbooks : undefined,
      current_progress: currentProgress || undefined,
    }
  }, [
    expanded,
    currentLevel,
    showGradeSelector,
    selectedGradeIdx,
    gradeOptions,
    selectedSubjects,
    selectedTextbooks,
    customTextbooks,
    currentProgress,
    validate,
  ])

  return {
    // 状态
    formValue,
    expanded,
    selectedLevelIdx,
    selectedGradeIdx,
    selectedSubjects,
    selectedTextbooks,
    customTextbooks,
    currentProgress,
    currentLevel,
    errors,

    // 计算属性
    showGradeSelector,
    useCustomTextbookMode,
    isAdultLevel,
    gradeOptions,
    subjectOptions,
    levelOptions,
    hasValidConfig,
    isValid,

    // 操作函数
    toggleExpanded,
    setExpanded,
    selectLevel,
    selectGrade,
    toggleSubject,
    toggleTextbook,
    addCustomTextbook,
    removeCustomTextbook,
    setCurrentProgress,
    validate,
    reset,
    resetToInitial,
    getSubmitValue,
  }
}

export default useCurriculumConfig
