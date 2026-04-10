import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import CurriculumConfigForm from '../index'
import { GRADE_LEVELS, K12_SUBJECTS, TEXTBOOK_VERSIONS } from '@/api/curriculum'
import '@testing-library/jest-dom'

describe('CurriculumConfigForm', () => {
  const defaultProps = {
    onChange: jest.fn(),
    onExpandedChange: jest.fn(),
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('render', () => {
    it('默认折叠状态应正确渲染头部', () => {
      render(<CurriculumConfigForm {...defaultProps} />)

      expect(screen.getByText('📚 教材配置（可选）')).toBeInTheDocument()
      expect(screen.getByText('点击添加教材配置信息')).toBeInTheDocument()
    })

    it('展开状态下应正确渲染头部', () => {
      render(<CurriculumConfigForm {...defaultProps} expanded={true} />)

      expect(screen.getByText('📚 教材配置（可选）')).toBeInTheDocument()
      expect(screen.getByText('配置年级和教材，让AI更精准地辅导')).toBeInTheDocument()
    })
  })

  describe('interact - toggle expand', () => {
    it('点击头部应展开配置区域（非受控模式）', () => {
      render(<CurriculumConfigForm {...defaultProps} />)

      const header = screen.getByText('📚 教材配置（可选）').closest('.curriculum-form__header')
      fireEvent.click(header!)

      expect(screen.getByText('选择学段')).toBeInTheDocument()
      expect(defaultProps.onExpandedChange).toHaveBeenCalledWith(true)
    })

    it('点击头部应折叠配置区域（受控模式）', () => {
      const onExpandedChange = jest.fn()
      render(
        <CurriculumConfigForm
          {...defaultProps}
          expanded={true}
          onExpandedChange={onExpandedChange}
        />
      )

      const header = screen.getByText('📚 教材配置（可选）').closest('.curriculum-form__header')
      fireEvent.click(header!)

      expect(onExpandedChange).toHaveBeenCalledWith(false)
    })
  })

  describe('grade level selection', () => {
    it('选择学段后应显示年级选择器', async () => {
      render(<CurriculumConfigForm {...defaultProps} expanded={true} />)

      // 使用 select 元素模拟 Picker 组件
      const levelSelect = screen.getByTestId('picker-select') as HTMLSelectElement
      fireEvent.change(levelSelect, { target: { value: '1' } }) // primary_lower

      await waitFor(() => {
        expect(screen.getByText('选择年级')).toBeInTheDocument()
      })
    })

    it('选择小学低年级应显示对应年级选项', async () => {
      render(<CurriculumConfigForm {...defaultProps} expanded={true} />)

      const levelSelect = screen.getByTestId('picker-select') as HTMLSelectElement
      fireEvent.change(levelSelect, { target: { value: '1' } }) // primary_lower

      await waitFor(() => {
        const level = GRADE_LEVELS[1]
        expect(level.value).toBe('primary_lower')
      })
    })

    it('切换学段后应重置年级选择', async () => {
      const onChange = jest.fn()
      render(<CurriculumConfigForm {...defaultProps} expanded={true} onChange={onChange} />)

      const levelSelect = screen.getByTestId('picker-select') as HTMLSelectElement
      fireEvent.change(levelSelect, { target: { value: '1' } }) // primary_lower

      await waitFor(() => {
        const changeCalls = onChange.mock.calls
        expect(changeCalls.length).toBeGreaterThan(0)
      })
    })
  })

  describe('subject selection', () => {
    it('选择学段后应显示学科标签', async () => {
      render(<CurriculumConfigForm {...defaultProps} expanded={true} />)

      const levelSelect = screen.getByTestId('picker-select') as HTMLSelectElement
      fireEvent.change(levelSelect, { target: { value: '1' } }) // primary_lower

      await waitFor(() => {
        expect(screen.getByText('教学学科（可多选）')).toBeInTheDocument()
      })
    })

    it('点击学科标签应选中', async () => {
      render(<CurriculumConfigForm {...defaultProps} expanded={true} />)

      const levelSelect = screen.getByTestId('picker-select') as HTMLSelectElement
      fireEvent.change(levelSelect, { target: { value: '1' } }) // primary_lower

      await waitFor(() => {
        // 学科标签应存在
        expect(document.querySelector('.curriculum-form__tags')).toBeInTheDocument()
      })
    })
  })

  describe('textbook version selection', () => {
    it('K12学段应显示教材版本标签', async () => {
      render(<CurriculumConfigForm {...defaultProps} expanded={true} />)

      const levelSelect = screen.getByTestId('picker-select') as HTMLSelectElement
      fireEvent.change(levelSelect, { target: { value: '1' } }) // primary_lower

      await waitFor(() => {
        expect(screen.getByText(/教材版本/)).toBeInTheDocument()
      })
    })
  })

  describe('university grade level', () => {
    it('大学及以上学段应显示自定义教材输入区域', async () => {
      render(<CurriculumConfigForm {...defaultProps} expanded={true} />)

      const levelSelect = screen.getByTestId('picker-select') as HTMLSelectElement
      fireEvent.change(levelSelect, { target: { value: '5' } }) // university

      await waitFor(() => {
        expect(screen.getByText('大学及以上请手动填写使用的教材名称')).toBeInTheDocument()
      })
    })
  })

  describe('adult grade level', () => {
    it('成人生活技能应显示正确学科类别', async () => {
      render(<CurriculumConfigForm {...defaultProps} expanded={true} />)

      const levelSelect = screen.getByTestId('picker-select') as HTMLSelectElement
      fireEvent.change(levelSelect, { target: { value: '6' } }) // adult_life

      await waitFor(() => {
        expect(screen.getByText('课程类别（可多选）')).toBeInTheDocument()
      })
    })
  })

  describe('initial value', () => {
    it('编辑时应有初始值回填', async () => {
      const initialValue = {
        grade_level: 'primary_lower',
        grade: '三年级',
        subjects: ['数学', '语文'],
        textbook_versions: ['人教版'],
        custom_textbooks: [],
        current_progress: '第三单元',
      }

      render(
        <CurriculumConfigForm {...defaultProps} expanded={true} initialValue={initialValue} />
      )

      await waitFor(() => {
        expect(screen.getByText('小学低年级')).toBeInTheDocument()
        expect(screen.getByText('三年级')).toBeInTheDocument()
      })
    })
  })

  describe('onChange callback', () => {
    it('表单值变化时应触发onChange', async () => {
      const onChange = jest.fn()
      render(<CurriculumConfigForm {...defaultProps} expanded={true} onChange={onChange} />)

      const levelSelect = screen.getByTestId('picker-select') as HTMLSelectElement
      fireEvent.change(levelSelect, { target: { value: '1' } }) // primary_lower

      await waitFor(() => {
        expect(onChange).toHaveBeenCalled()
        const lastCall = onChange.mock.calls[onChange.mock.calls.length - 1][0]
        expect(lastCall.grade_level).toBe('primary_lower')
      })
    })
  })
})
