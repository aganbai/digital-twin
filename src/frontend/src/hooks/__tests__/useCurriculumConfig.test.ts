/**
 * useCurriculumConfig Hook 单元测试
 * 模块: FE-IT13-006 - 小程序-教材配置Hook封装
 *
 * 测试覆盖:
 * 1. 基础状态管理
 * 2. 学段年级联动逻辑
 * 3. 表单验证
 * 4. 初始值回填（编辑场景）
 * 5. 计算属性
 * 6. 操作函数
 */

import { renderHook, act, waitFor } from '@testing-library/react'
import { useCurriculumConfig } from '../useCurriculumConfig'
import Taro from '@tarojs/taro'
import { GRADE_LEVEL_OPTIONS, K12_SUBJECTS, EMPTY_CURRICULUM_CONFIG } from '@/constants/curriculum'
import { mockCurriculumConfigPrimary, mockCurriculumConfigUniversity } from '@/__tests__/fixtures/curriculum-fixtures'

describe('useCurriculumConfig Hook', () => {
  // 清理 mock
  beforeEach(() => {
    jest.clearAllMocks()
  })

  // ==================== 基础状态测试 ====================
  describe('基础状态管理', () => {
    it('应使用默认初始值创建 hook', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      // 验证初始状态
      expect(result.current.expanded).toBe(false)
      expect(result.current.selectedLevelIdx).toBe(-1)
      expect(result.current.selectedGradeIdx).toBe(-1)
      expect(result.current.selectedSubjects).toEqual([])
      expect(result.current.selectedTextbooks).toEqual([])
      expect(result.current.customTextbooks).toEqual([])
      expect(result.current.currentProgress).toBe('')
      expect(result.current.errors).toEqual({})
    })

    it('应接受自定义 defaultExpanded 选项', () => {
      const { result } = renderHook(() => useCurriculumConfig({
        defaultExpanded: true,
      }))

      expect(result.current.expanded).toBe(true)
    })

    it('toggleExpanded 应切换展开状态', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      // 初始为折叠
      expect(result.current.expanded).toBe(false)

      // 切换为展开
      act(() => {
        result.current.toggleExpanded()
      })
      expect(result.current.expanded).toBe(true)

      // 切换回折叠
      act(() => {
        result.current.toggleExpanded()
      })
      expect(result.current.expanded).toBe(false)
    })

    it('setExpanded 应直接设置展开状态', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      act(() => {
        result.current.setExpanded(true)
      })
      expect(result.current.expanded).toBe(true)

      act(() => {
        result.current.setExpanded(false)
      })
      expect(result.current.expanded).toBe(false)
    })
  })

  // ==================== 学段年级联动测试 ====================
  describe('学段年级联动逻辑', () => {
    it('selectLevel 应选择学段并重置相关状态', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      act(() => {
        result.current.selectLevel(1) // 选择小学低年级
      })

      expect(result.current.selectedLevelIdx).toBe(1)
      expect(result.current.currentLevel).toBe('primary_lower')
      expect(result.current.gradeOptions).toEqual(['一年级', '二年级', '三年级'])
      expect(result.current.subjectOptions).toEqual(K12_SUBJECTS)
    })

    it('选择学段应清空已选学科和教材版本', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      // 先设置一些状态
      act(() => {
        result.current.selectLevel(1)
        result.current.toggleSubject('数学')
        result.current.toggleTextbook('人教版')
      })

      expect(result.current.selectedSubjects).toContain('数学')
      expect(result.current.selectedTextbooks).toContain('人教版')

      // 切换学段
      act(() => {
        result.current.selectLevel(2)
      })

      // 学科和教材版本应被清空
      expect(result.current.selectedSubjects).toEqual([])
      expect(result.current.selectedTextbooks).toEqual([])
    })

    it('selectGrade 应选择年级', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      act(() => {
        result.current.selectLevel(1) // 小学低年级
      })

      act(() => {
        result.current.selectGrade(1) // 选择二年级
      })

      expect(result.current.selectedGradeIdx).toBe(1)
    })

    it('选择年级应清除年级错误', () => {
      const { result } = renderHook(() => useCurriculumConfig({
        defaultExpanded: true,
      }))

      act(() => {
        result.current.selectLevel(1)
      })

      act(() => {
        result.current.validate() // 触发验证错误
      })

      expect(result.current.errors.grade).toBe('请选择年级')

      act(() => {
        result.current.selectGrade(0)
      })

      expect(result.current.errors.grade).toBeUndefined()
    })
  })

  // ==================== 学科选择测试 ====================
  describe('学科选择', () => {
    it('toggleSubject 应添加未选学科', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      act(() => {
        result.current.selectLevel(1)
        result.current.toggleSubject('数学')
      })

      expect(result.current.selectedSubjects).toContain('数学')
    })

    it('toggleSubject 应移除已选学科', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      act(() => {
        result.current.selectLevel(1)
        result.current.toggleSubject('数学')
        result.current.toggleSubject('数学')
      })

      expect(result.current.selectedSubjects).not.toContain('数学')
    })

    it('toggleSubject 应清除学科错误', () => {
      const { result } = renderHook(() => useCurriculumConfig({
        defaultExpanded: true,
      }))

      act(() => {
        result.current.selectLevel(1)
      })

      act(() => {
        result.current.validate()
      })

      expect(result.current.errors.subjects).toBe('请选择教学学科')

      act(() => {
        result.current.toggleSubject('数学')
      })

      expect(result.current.errors.subjects).toBeUndefined()
    })
  })

  // ==================== 教材版本选择测试 ====================
  describe('教材版本选择', () => {
    it('toggleTextbook 应添加未选版本', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      act(() => {
        result.current.toggleTextbook('人教版')
      })

      expect(result.current.selectedTextbooks).toContain('人教版')
    })

    it('toggleTextbook 应移除已选版本', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      act(() => {
        result.current.toggleTextbook('人教版')
        result.current.toggleTextbook('人教版')
      })

      expect(result.current.selectedTextbooks).not.toContain('人教版')
    })

    it('toggleTextbook 应支持多选', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      act(() => {
        result.current.toggleTextbook('人教版')
        result.current.toggleTextbook('北师大版')
      })

      expect(result.current.selectedTextbooks).toContain('人教版')
      expect(result.current.selectedTextbooks).toContain('北师大版')
    })
  })

  // ==================== 自定义教材测试 ====================
  describe('自定义教材', () => {
    it('addCustomTextbook 应添加有效教材', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      let success: boolean = false
      act(() => {
        success = result.current.addCustomTextbook('《高等数学》')
      })

      expect(success).toBe(true)
      expect(result.current.customTextbooks).toContain('《高等数学》')
    })

    it('addCustomTextbook 应拒绝空字符串', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      let success: boolean = true
      act(() => {
        success = result.current.addCustomTextbook('   ')
      })

      expect(success).toBe(false)
      expect(result.current.customTextbooks).toEqual([])
      expect(Taro.showToast).toHaveBeenCalledWith({
        title: '请输入教材名称',
        icon: 'none',
      })
    })

    it('addCustomTextbook 应拒绝重复教材', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      act(() => {
        result.current.addCustomTextbook('《高等数学》')
      })

      let success: boolean = true
      act(() => {
        success = result.current.addCustomTextbook('《高等数学》')
      })

      expect(success).toBe(false)
      expect(result.current.customTextbooks).toHaveLength(1)
      expect(Taro.showToast).toHaveBeenCalledWith({
        title: '该教材已添加',
        icon: 'none',
      })
    })

    it('removeCustomTextbook 应移除指定教材', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      act(() => {
        result.current.addCustomTextbook('教材1')
        result.current.addCustomTextbook('教材2')
      })

      act(() => {
        result.current.removeCustomTextbook('教材1')
      })

      expect(result.current.customTextbooks).not.toContain('教材1')
      expect(result.current.customTextbooks).toContain('教材2')
    })
  })

  // ==================== 教学进度设置测试 ====================
  describe('教学进度设置', () => {
    it('setCurrentProgress 应设置进度', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      act(() => {
        result.current.setCurrentProgress('第三单元')
      })

      expect(result.current.currentProgress).toBe('第三单元')
    })
  })

  // ==================== 表单验证测试 ====================
  describe('表单验证', () => {
    it('折叠状态下验证应通过', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      let isValid: boolean = false
      act(() => {
        isValid = result.current.validate()
      })

      expect(isValid).toBe(true)
    })

    it('展开但未填写应验证失败', () => {
      const { result } = renderHook(() => useCurriculumConfig({
        defaultExpanded: true,
      }))

      let isValid: boolean = true

      act(() => {
        isValid = result.current.validate()
      })

      expect(isValid).toBe(false)
      expect(result.current.errors.grade_level).toBe('请选择学段')
    })

    it('选择学段后应清除学段错误', () => {
      const { result } = renderHook(() => useCurriculumConfig({
        defaultExpanded: true,
      }))

      act(() => {
        result.current.validate()
      })

      expect(result.current.errors.grade_level).toBeDefined()

      act(() => {
        result.current.selectLevel(1)
      })

      expect(result.current.errors.grade_level).toBeUndefined()
    })

    it('未选年级应显示年级错误', () => {
      const onChange = jest.fn()
      const { result } = renderHook(() => useCurriculumConfig({
        defaultExpanded: true,
        onChange,
      }))

      act(() => {
        result.current.selectLevel(1)
      })

      // 等待 effect 执行完成
      waitFor(() => {
        expect(onChange).toHaveBeenCalled()
      })

      act(() => {
        result.current.validate()
      })

      expect(result.current.errors.grade).toBe('请选择年级')
    })

    it('未选学科应显示学科错误', () => {
      const { result } = renderHook(() => useCurriculumConfig({
        defaultExpanded: true,
      }))

      act(() => {
        result.current.selectLevel(1)
        result.current.selectGrade(0)
        result.current.validate()
      })

      expect(result.current.errors.subjects).toBe('请选择教学学科')
    })

    it('成人学段应显示正确错误提示', () => {
      const onChange = jest.fn()
      const { result } = renderHook(() => useCurriculumConfig({
        defaultExpanded: true,
        onChange,
      }))

      act(() => {
        result.current.selectLevel(6) // adult_life
      })

      // 等待 useEffect 更新 isAdultLevel
      waitFor(() => {
        expect(result.current.isAdultLevel).toBe(true)
      })

      act(() => {
        result.current.validate()
      })

      expect(result.current.errors.subjects).toBe('请选择课程类别')
    })

    it('完整填写后验证应通过', () => {
      const { result } = renderHook(() => useCurriculumConfig({
        defaultExpanded: true,
      }))

      act(() => {
        result.current.selectLevel(1)
        result.current.selectGrade(0)
        result.current.toggleSubject('数学')
      })

      let isValid: boolean = false
      act(() => {
        isValid = result.current.validate()
      })

      expect(isValid).toBe(true)
      expect(result.current.errors).toEqual({})
    })
  })

  // ==================== 计算属性测试 ====================
  describe('计算属性', () => {
    it('showGradeSelector 应根据学段正确显示', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      // K12学段需要年级选择
      act(() => { result.current.selectLevel(0) })
      expect(result.current.showGradeSelector).toBe(true)

      act(() => { result.current.selectLevel(1) })
      expect(result.current.showGradeSelector).toBe(true)

      // 成人学段不需要年级选择
      act(() => { result.current.selectLevel(6) })
      expect(result.current.showGradeSelector).toBe(false)

      act(() => { result.current.selectLevel(7) })
      expect(result.current.showGradeSelector).toBe(false)
    })

    it('useCustomTextbookMode 应正确识别需要自定义教材的学段', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      // 小学不需要自定义教材
      act(() => { result.current.selectLevel(1) })
      expect(result.current.useCustomTextbookMode).toBe(false)

      // 大学需要自定义教材
      act(() => { result.current.selectLevel(5) })
      expect(result.current.useCustomTextbookMode).toBe(true)

      // 成人培训需要自定义教材
      act(() => { result.current.selectLevel(6) })
      expect(result.current.useCustomTextbookMode).toBe(true)
    })

    it('isAdultLevel 应正确识别成人学段', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      act(() => { result.current.selectLevel(6) })
      expect(result.current.isAdultLevel).toBe(true)

      act(() => { result.current.selectLevel(7) })
      expect(result.current.isAdultLevel).toBe(true)

      act(() => { result.current.selectLevel(1) })
      expect(result.current.isAdultLevel).toBe(false)
    })

    it('hasValidConfig 应在填充有效配置后为 true', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      expect(result.current.hasValidConfig).toBe(false)

      act(() => {
        result.current.selectLevel(1)
        result.current.toggleSubject('数学')
      })

      expect(result.current.hasValidConfig).toBe(true)
    })

    it('isValid 应综合判断表单状态', async () => {
      const { result } = renderHook(() => useCurriculumConfig())

      // 未展开，应视为有效
      expect(result.current.isValid).toBe(true)

      act(() => {
        result.current.setExpanded(true)
      })

      // 展开但未填写
      expect(result.current.isValid).toBe(false)

      // 选择学段
      act(() => {
        result.current.selectLevel(1)
      })

      // 等待状态更新
      await waitFor(() => {
        expect(result.current.currentLevel).toBe('primary_lower')
      })

      // 选择年级
      act(() => {
        result.current.selectGrade(0)
      })

      // 选择学科
      act(() => {
        result.current.toggleSubject('数学')
      })

      // 等待 useMemo 重新计算（errors 可能还包含旧错误）
      // 清除错误
      act(() => {
        result.current.validate()
      })

      // 等待状态稳定
      await waitFor(() => {
        expect(result.current.errors).toEqual({})
      })

      // 完全填写后应为有效
      expect(result.current.isValid).toBe(true)
    })
  })

  // ==================== 重置功能测试 ====================
  describe('重置功能', () => {
    it('reset 应清空所有状态', () => {
      const { result } = renderHook(() => useCurriculumConfig({
        defaultExpanded: true,
      }))

      act(() => {
        result.current.selectLevel(1)
        result.current.selectGrade(0)
        result.current.toggleSubject('数学')
        result.current.toggleTextbook('人教版')
        result.current.addCustomTextbook('书籍')
        result.current.setCurrentProgress('第一单元')
      })

      act(() => {
        result.current.reset()
      })

      expect(result.current.selectedLevelIdx).toBe(-1)
      expect(result.current.selectedGradeIdx).toBe(-1)
      expect(result.current.selectedSubjects).toEqual([])
      expect(result.current.selectedTextbooks).toEqual([])
      expect(result.current.customTextbooks).toEqual([])
      expect(result.current.currentProgress).toBe('')
      expect(result.current.errors).toEqual({})
    })

    it('resetToInitial 在没有初始值时应调用 reset', () => {
      const { result } = renderHook(() => useCurriculumConfig({
        defaultExpanded: true,
      }))

      act(() => {
        result.current.selectLevel(1)
        result.current.toggleSubject('数学')
      })

      act(() => {
        result.current.resetToInitial()
      })

      expect(result.current.selectedLevelIdx).toBe(-1)
      expect(result.current.selectedSubjects).toEqual([])
    })
  })

  // ==================== 初始值回填测试（编辑场景） ====================
  describe('初始值回填（编辑场景）', () => {
    it('应正确回填完整配置数据', () => {
      const { result } = renderHook(() => useCurriculumConfig({
        initialValue: mockCurriculumConfigPrimary,
      }))

      expect(result.current.selectedLevelIdx).toBe(1) // primary_lower
      expect(result.current.currentLevel).toBe('primary_lower')
      expect(result.current.selectedGradeIdx).toBe(2) // 三年级
      expect(result.current.selectedSubjects).toEqual(['数学', '语文'])
      expect(result.current.selectedTextbooks).toEqual(['人教版', '北师大版'])
      expect(result.current.customTextbooks).toEqual(['《小学奥数启蒙》'])
      expect(result.current.currentProgress).toBe('第三单元 乘法初步')
    })

    it('应正确回填大学学段配置', () => {
      const { result } = renderHook(() => useCurriculumConfig({
        initialValue: mockCurriculumConfigUniversity,
      }))

      expect(result.current.selectedLevelIdx).toBe(5) // university
      expect(result.current.currentLevel).toBe('university')
      expect(result.current.selectedGradeIdx).toBe(1) // 大二
    })

    it('应处理 null 初始值', () => {
      const { result } = renderHook(() => useCurriculumConfig({
        initialValue: null,
      }))

      expect(result.current.selectedLevelIdx).toBe(-1)
      expect(result.current.selectedSubjects).toEqual([])
    })

    it('resetToInitial 应恢复为初始值', () => {
      const { result } = renderHook(() => useCurriculumConfig({
        initialValue: mockCurriculumConfigPrimary,
      }))

      // 修改一些值
      act(() => {
        result.current.toggleSubject('英语')
        result.current.selectGrade(0)
      })

      // 恢复初始值
      act(() => {
        result.current.resetToInitial()
      })

      expect(result.current.selectedSubjects).toEqual(['数学', '语文'])
      expect(result.current.selectedGradeIdx).toBe(2)
    })

    it('应处理无 grade_level 的配置对象', () => {
      const partialConfig = {
        grade: '三年级',
        subjects: ['数学'],
      }

      const { result } = renderHook(() => useCurriculumConfig({
        // @ts-expect-error 测试部分配置对象
        initialValue: partialConfig,
      }))

      expect(result.current.selectedLevelIdx).toBe(-1)
    })
  })

  // ==================== 获取提交值测试 ====================
  describe('getSubmitValue', () => {
    it('折叠状态应返回 null', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      let value: any = undefined
      act(() => {
        value = result.current.getSubmitValue()
      })

      expect(value).toBeNull()
    })

    it('展开但未选学段应返回 null', () => {
      const { result } = renderHook(() => useCurriculumConfig({
        defaultExpanded: true,
      }))

      let value: any = undefined
      act(() => {
        value = result.current.getSubmitValue()
      })

      expect(value).toBeNull()
    })

    it('完整配置应返回正确数据结构', () => {
      const { result } = renderHook(() => useCurriculumConfig({
        defaultExpanded: true,
      }))

      act(() => {
        result.current.selectLevel(1)
        result.current.selectGrade(0)
        result.current.toggleSubject('数学')
        result.current.toggleTextbook('人教版')
        result.current.setCurrentProgress('第一单元')
      })

      let value: any
      act(() => {
        value = result.current.getSubmitValue()
      })

      expect(value).toEqual({
        grade_level: 'primary_lower',
        grade: '一年级',
        subjects: ['数学'],
        textbook_versions: ['人教版'],
        custom_textbooks: undefined,
        current_progress: '第一单元',
      })
    })

    it('无教材版本时应返回 undefined', () => {
      const { result } = renderHook(() => useCurriculumConfig({
        defaultExpanded: true,
      }))

      act(() => {
        result.current.selectLevel(6) // adult_life，不需要年级
        result.current.toggleSubject('中餐')
      })

      let value: any
      act(() => {
        value = result.current.getSubmitValue()
      })

      expect(value?.textbook_versions).toBeUndefined()
    })

    it('验证失败时应返回 null', () => {
      const { result } = renderHook(() => useCurriculumConfig({
        defaultExpanded: true,
      }))

      act(() => {
        result.current.selectLevel(1)
        // 不选择年级和学科，验证应失败
      })

      let value: any
      act(() => {
        value = result.current.getSubmitValue()
      })

      expect(value).toBeNull()
    })
  })

  // ==================== 回调函数测试 ====================
  describe('回调函数', () => {
    it('onChange 应在值变化时触发', () => {
      const onChange = jest.fn()

      const { result } = renderHook(() => useCurriculumConfig({
        onChange,
      }))

      act(() => {
        result.current.selectLevel(1)
      })

      expect(onChange).toHaveBeenCalled()
    })

    it('onExpandedChange 应在展开状态变化时触发', () => {
      const onExpandedChange = jest.fn()

      const { result } = renderHook(() => useCurriculumConfig({
        onExpandedChange,
      }))

      act(() => {
        result.current.toggleExpanded()
      })

      expect(onExpandedChange).toHaveBeenCalledWith(true)

      act(() => {
        result.current.toggleExpanded()
      })

      expect(onExpandedChange).toHaveBeenCalledWith(false)
    })

    it('setExpanded 应触发 onExpandedChange', () => {
      const onExpandedChange = jest.fn()

      const { result } = renderHook(() => useCurriculumConfig({
        onExpandedChange,
      }))

      act(() => {
        result.current.setExpanded(true)
      })

      expect(onExpandedChange).toHaveBeenCalledWith(true)
    })
  })

  // ==================== 边界条件测试 ====================
  describe('边界条件', () => {
    it('学段索引超出范围应处理正确', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      act(() => {
        result.current.selectLevel(-1)
      })

      expect(result.current.currentLevel).toBeUndefined()
      expect(result.current.gradeOptions).toEqual([])
    })

    it('年级索引超出范围应处理正确', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      act(() => {
        result.current.selectLevel(1)
        result.current.selectGrade(99)
      })

      // 选择超出的年级索引
      expect(result.current.selectedGradeIdx).toBe(99)
      // formValue 中不应包含无效的年级
      expect(result.current.formValue.grade).toBeUndefined()
    })

    it('formValue 应在折叠且未选择时包含空值', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      expect(result.current.formValue).toEqual({
        grade_level: undefined,
        grade: undefined,
        subjects: [],
        textbook_versions: [],
        custom_textbooks: [],
        current_progress: undefined,
      })
    })

    it('自定义教材去重应正确处理边界空白字符', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      let success: boolean = false
      act(() => {
        success = result.current.addCustomTextbook('  教材1  ')
      })

      expect(success).toBe(true)
      expect(result.current.customTextbooks).toContain('教材1')
    })
  })

  // ==================== 学段年级映射测试 ====================
  describe('学段年级映射', () => {
    it('所有学段应返回正确的年级选项', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      const testCases = [
        { idx: 0, expected: ['幼儿园大班', '学前'] },
        { idx: 1, expected: ['一年级', '二年级', '三年级'] },
        { idx: 2, expected: ['四年级', '五年级', '六年级'] },
        { idx: 3, expected: ['七年级', '八年级', '九年级'] },
        { idx: 4, expected: ['高一', '高二', '高三'] },
        { idx: 5, expected: ['大一', '大二', '大三', '大四', '研究生', '博士'] },
        { idx: 6, expected: [] },
        { idx: 7, expected: [] },
      ]

      testCases.forEach(({ idx, expected }) => {
        act(() => {
          result.current.selectLevel(idx)
        })
        expect(result.current.gradeOptions).toEqual(expected)
      })
    })

    it('所有学段应返回正确的学科选项', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      // K12 学段
      for (let i = 0; i <= 4; i++) {
        act(() => {
          result.current.selectLevel(i)
        })
        expect(result.current.subjectOptions).toEqual(K12_SUBJECTS)
      }

      // 大学
      act(() => { result.current.selectLevel(5) })
      expect(result.current.subjectOptions).toEqual(K12_SUBJECTS)

      // 成人生活
      act(() => { result.current.selectLevel(6) })
      expect(result.current.subjectOptions).not.toEqual(K12_SUBJECTS)

      // 成人职业
      act(() => { result.current.selectLevel(7) })
      expect(result.current.subjectOptions).not.toEqual(K12_SUBJECTS)
    })

    it('levelOptions 应返回所有学段选项', () => {
      const { result } = renderHook(() => useCurriculumConfig())

      expect(result.current.levelOptions).toBe(GRADE_LEVEL_OPTIONS)
      expect(result.current.levelOptions).toHaveLength(8)
    })
  })
})
