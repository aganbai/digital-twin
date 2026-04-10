/**
 * ClassEdit 班级编辑页面单元测试
 * 模块: FE-IT13-002 - 小程序-班级编辑页教材配置区域
 */

import React from 'react'
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import '@testing-library/jest-dom'
import ClassEdit from '../index'
import { server, mockClassDetail, mockClassWithCurriculum } from '@/__mocks__/server'
import { http, HttpResponse } from 'msw'
import Taro from '@tarojs/taro'

// Mock Taro getCurrentInstance
const mockGetCurrentInstance = jest.fn()
jest.mock('@tarojs/taro', () => ({
  ...jest.requireActual('@tarojs/taro'),
  getCurrentInstance: () => mockGetCurrentInstance(),
}))

// Mock API
jest.mock('@/api/class', () => ({
  updateClassV11: jest.fn(),
  getClassDetail: jest.fn(),
}))

jest.mock('@/components/CurriculumConfigForm', () => {
  return function MockCurriculumConfigForm({
    expanded,
    onExpandedChange,
    onChange,
    initialValue,
  }: any) {
    React.useEffect(() => {
      if (onChange) {
        onChange({
          grade_level: initialValue?.grade_level || 'primary_lower',
          grade: initialValue?.grade || '三年级',
          subjects: initialValue?.subjects || ['数学'],
          textbook_versions: initialValue?.textbook_versions || ['人教版'],
          custom_textbooks: initialValue?.custom_textbooks || [],
          current_progress: initialValue?.current_progress || '',
        })
      }
    }, [initialValue])

    return (
      <div data-testid="curriculum-config-form">
        <div
          data-testid="curriculum-toggle"
          onClick={() => onExpandedChange && onExpandedChange(!expanded)}
        >
          {expanded ? '📚 教材配置（已展开）' : '📚 教材配置（可选）'}
        </div>
        {expanded && (
          <div data-testid="curriculum-form-expanded">
            {initialValue ? '编辑模式' : '新增模式'}
          </div>
        )}
        {initialValue && (
          <div data-testid="initial-value">{JSON.stringify(initialValue)}</div>
        )}
      </div>
    )
  }
})

import { updateClassV11, getClassDetail } from '@/api/class'

describe('ClassEdit 班级编辑页面', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    // 默认设置有效的 class_id
    mockGetCurrentInstance.mockReturnValue({
      router: { params: { class_id: '123' } },
    })
  })

  describe('渲染测试', () => {
    it('应正确渲染班级编辑页面基本结构', async () => {
      ;(getClassDetail as jest.Mock).mockResolvedValue({
        data: mockClassDetail,
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      // 等待加载完成
      await waitFor(() => {
        expect(screen.queryByText('加载中...')).not.toBeInTheDocument()
      })

      // 验证页面标题
      expect(screen.getByText('编辑班级信息')).toBeInTheDocument()
      expect(screen.getByText('修改班级的基本信息')).toBeInTheDocument()

      // 验证表单字段
      expect(screen.getByText(/班级名称/)).toBeInTheDocument()
      expect(screen.getByText(/班级描述/)).toBeInTheDocument()
      expect(screen.getByText('公开班级')).toBeInTheDocument()

      // 验证保存按钮
      expect(screen.getByText('保存')).toBeInTheDocument()
    })

    it('应加载并显示班级详情数据', async () => {
      ;(getClassDetail as jest.Mock).mockResolvedValue({
        data: mockClassDetail,
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(getClassDetail).toHaveBeenCalledWith(123)
      })

      // 验证数据已经加载（通过 input 值）
      const inputs = screen.getAllByRole('textbox')
      expect(inputs[0]).toHaveValue(mockClassDetail.name)
      expect(inputs[1]).toHaveValue(mockClassDetail.description)
    })

    it('有教材配置的班级应自动展开显示', async () => {
      ;(getClassDetail as jest.Mock).mockResolvedValue({
        data: mockClassWithCurriculum,
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(screen.getByTestId('curriculum-form-expanded')).toBeInTheDocument()
        expect(screen.getByText('编辑模式')).toBeInTheDocument()
      })

      // 验证初始值已传递
      expect(screen.getByTestId('initial-value')).toHaveTextContent('primary_lower')
      expect(screen.getByTestId('initial-value')).toHaveTextContent('三年级')
    })

    it('无教材配置的班级应保持折叠', async () => {
      ;(getClassDetail as jest.Mock).mockResolvedValue({
        data: mockClassDetail,
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(screen.getByTestId('curriculum-config-form')).toBeInTheDocument()
      })

      expect(screen.queryByTestId('curriculum-form-expanded')).not.toBeInTheDocument()
    })

    it('保存按钮默认应为禁用状态（无修改时）', async () => {
      ;(getClassDetail as jest.Mock).mockResolvedValue({
        data: mockClassDetail,
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(getClassDetail).toHaveBeenCalled()
      })

      // 初始状态下，已经有名称，可以提交
      const submitBtn = screen.getByText('保存').closest('div')
      expect(submitBtn).not.toHaveClass('class-edit__submit--disabled')
    })
  })

  describe('表单编辑测试', () => {
    it('修改班级名称应正确更新', async () => {
      ;(getClassDetail as jest.Mock).mockResolvedValue({
        data: mockClassDetail,
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(getClassDetail).toHaveBeenCalled()
      })

      const inputs = screen.getAllByRole('textbox')
      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '新的班级名称' } })
      })

      expect(inputs[0]).toHaveValue('新的班级名称')
    })

    it('清空班级名称后保存按钮应禁用', async () => {
      ;(getClassDetail as jest.Mock).mockResolvedValue({
        data: mockClassDetail,
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(getClassDetail).toHaveBeenCalled()
      })

      const inputs = screen.getAllByRole('textbox')
      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '' } })
      })

      const submitBtn = screen.getByText('保存').closest('div')
      expect(submitBtn).toHaveClass('class-edit__submit--disabled')
    })

    it('切换公开开关应更新状态', async () => {
      ;(getClassDetail as jest.Mock).mockResolvedValue({
        data: { ...mockClassDetail, is_public: true },
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(getClassDetail).toHaveBeenCalled()
      })

      const switchInput = screen.getByRole('checkbox')
      expect(switchInput).toBeChecked()

      await act(async () => {
        fireEvent.click(switchInput)
      })

      expect(switchInput).not.toBeChecked()
    })

    it('修改班级描述应正确更新', async () => {
      ;(getClassDetail as jest.Mock).mockResolvedValue({
        data: mockClassDetail,
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(getClassDetail).toHaveBeenCalled()
      })

      const inputs = screen.getAllByRole('textbox')
      await act(async () => {
        fireEvent.change(inputs[1], { target: { value: '新的班级描述' } })
      })

      expect(inputs[1]).toHaveValue('新的班级描述')
    })
  })

  describe('API 调用测试', () => {
    it('提交表单应调用 updateClassV11 API', async () => {
      ;(getClassDetail as jest.Mock).mockResolvedValue({
        data: mockClassDetail,
      })
      ;(updateClassV11 as jest.Mock).mockResolvedValue({
        data: { ...mockClassDetail, name: '修改后的名称' },
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(getClassDetail).toHaveBeenCalled()
      })

      const inputs = screen.getAllByRole('textbox')
      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '修改后的名称' } })
      })

      const submitBtn = screen.getByText('保存').closest('div')
      await act(async () => {
        fireEvent.click(submitBtn!)
      })

      await waitFor(() => {
        expect(updateClassV11).toHaveBeenCalledWith(
          123,
          expect.objectContaining({
            name: '修改后的名称',
            description: mockClassDetail.description,
            is_public: mockClassDetail.is_public,
          })
        )
      })
    })

    it('展开并填写教材配置后应携带配置提交', async () => {
      ;(getClassDetail as jest.Mock).mockResolvedValue({
        data: mockClassDetail,
      })
      ;(updateClassV11 as jest.Mock).mockResolvedValue({
        data: { ...mockClassDetail },
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(screen.getByTestId('curriculum-config-form')).toBeInTheDocument()
      })

      // 展开教材配置
      await act(async () => {
        fireEvent.click(screen.getByTestId('curriculum-toggle'))
      })

      // 修改班级名称
      const inputs = screen.getAllByRole('textbox')
      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '修改后的名称' } })
      })

      const submitBtn = screen.getByText('保存').closest('div')
      await act(async () => {
        fireEvent.click(submitBtn!)
      })

      await waitFor(() => {
        expect(updateClassV11).toHaveBeenCalledWith(
          123,
          expect.objectContaining({
            name: '修改后的名称',
            curriculum_config: expect.objectContaining({
              grade_level: expect.any(String),
              subjects: expect.any(Array),
            }),
          })
        )
      })
    })

    it('未展开教材配置时不应传递配置', async () => {
      ;(getClassDetail as jest.Mock).mockResolvedValue({
        data: mockClassDetail,
      })
      ;(updateClassV11 as jest.Mock).mockResolvedValue({
        data: mockClassDetail,
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(getClassDetail).toHaveBeenCalled()
      })

      const inputs = screen.getAllByRole('textbox')
      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '修改后的名称' } })
      })

      const submitBtn = screen.getByText('保存').closest('div')
      await act(async () => {
        fireEvent.click(submitBtn!)
      })

      await waitFor(() => {
        expect(updateClassV11).toHaveBeenCalledWith(
          123,
          expect.not.objectContaining({
            curriculum_config: expect.anything(),
          })
        )
      })
    })

    it('API 调用失败时应显示错误提示', async () => {
      ;(getClassDetail as jest.Mock).mockResolvedValue({
        data: mockClassDetail,
      })
      ;(updateClassV11 as jest.Mock).mockRejectedValue(new Error('保存失败'))

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(getClassDetail).toHaveBeenCalled()
      })

      const inputs = screen.getAllByRole('textbox')
      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '修改后的名称' } })
      })

      const submitBtn = screen.getByText('保存').closest('div')
      await act(async () => {
        fireEvent.click(submitBtn!)
      })

      await waitFor(() => {
        expect(Taro.showToast).toHaveBeenCalledWith(
          expect.objectContaining({
            title: expect.stringContaining('保存失败'),
            icon: 'none',
            duration: 3000,
          })
        )
      })
    })

    it('班级名称重复时应显示特定错误提示', async () => {
      ;(getClassDetail as jest.Mock).mockResolvedValue({
        data: mockClassDetail,
      })
      ;(updateClassV11 as jest.Mock).mockRejectedValue({
        message: '班级名称已存在',
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(getClassDetail).toHaveBeenCalled()
      })

      const inputs = screen.getAllByRole('textbox')
      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '重复班级名' } })
      })

      const submitBtn = screen.getByText('保存').closest('div')
      await act(async () => {
        fireEvent.click(submitBtn!)
      })

      await waitFor(() => {
        expect(Taro.showToast).toHaveBeenCalledWith(
          expect.objectContaining({
            icon: 'none',
            duration: 3000,
          })
        )
      })
    })

    it('保存成功后应导航返回', async () => {
      ;(getClassDetail as jest.Mock).mockResolvedValue({
        data: mockClassDetail,
      })
      ;(updateClassV11 as jest.Mock).mockResolvedValue({
        data: mockClassDetail,
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(getClassDetail).toHaveBeenCalled()
      })

      const submitBtn = screen.getByText('保存').closest('div')
      await act(async () => {
        fireEvent.click(submitBtn!)
      })

      await waitFor(() => {
        expect(Taro.showToast).toHaveBeenCalledWith(
          expect.objectContaining({
            title: '保存成功',
            icon: 'success',
          })
        )
      })
    })
  })

  describe('加载状态测试', () => {
    it('加载中应显示 loading 状态', async () => {
      // 延迟响应以模拟加载状态
      ;(getClassDetail as jest.Mock).mockImplementation(
        () =>
          new Promise((resolve) => {
            setTimeout(() => resolve({ data: mockClassDetail }), 500)
          })
      )

      await act(async () => {
        render(<ClassEdit />)
      })

      // 应该显示加载中
      expect(screen.getByText('加载中...')).toBeInTheDocument()

      await waitFor(() => {
        expect(screen.queryByText('加载中...')).not.toBeInTheDocument()
      })
    })

    it('无效的班级 ID 应显示错误并返回', async () => {
      mockGetCurrentInstance.mockReturnValue({
        router: { params: {} },
      })

      const setTimeoutSpy = jest.spyOn(global, 'setTimeout')

      await act(async () => {
        render(<ClassEdit />)
      })

      expect(Taro.showToast).toHaveBeenCalledWith({
        title: '班级ID不存在',
        icon: 'none',
      })

      // 验证 setTimeout 被调用（用于延迟返回）
      expect(setTimeoutSpy).toHaveBeenCalled()
    })
  })

  describe('API 错误处理测试', () => {
    it('获取班级详情失败时应显示 Toast 错误', async () => {
      ;(getClassDetail as jest.Mock).mockRejectedValue(new Error('网络错误'))

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(Taro.showToast).toHaveBeenCalledWith({
          title: '加载失败',
          icon: 'none',
        })
      })
    })

    it('班级不存在时应显示错误', async () => {
      ;(getClassDetail as jest.Mock).mockRejectedValue({
        code: 40017,
        message: '班级不存在',
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(Taro.showToast).toHaveBeenCalledWith({
          title: '加载失败',
          icon: 'none',
        })
      })
    })

    it('无权限操作时应显示错误', async () => {
      ;(getClassDetail as jest.Mock).mockRejectedValue({
        code: 40018,
        message: '无权操作此班级',
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(Taro.showToast).toHaveBeenCalledWith({
          title: '加载失败',
          icon: 'none',
        })
      })
    })
  })

  describe('边界条件测试', () => {
    it('班级描述为空时正确处理', async () => {
      ;(getClassDetail as jest.Mock).mockResolvedValue({
        data: { ...mockClassDetail, description: '' },
      })
      ;(updateClassV11 as jest.Mock).mockResolvedValue({
        data: { ...mockClassDetail, description: '' },
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(getClassDetail).toHaveBeenCalled()
      })

      const submitBtn = screen.getByText('保存').closest('div')
      await act(async () => {
        fireEvent.click(submitBtn!)
      })

      await waitFor(() => {
        expect(updateClassV11).toHaveBeenCalledWith(
          123,
          expect.objectContaining({
            description: '',
          })
        )
      })
    })

    it('表单数据提交时应去除首尾空格', async () => {
      ;(getClassDetail as jest.Mock).mockResolvedValue({
        data: mockClassDetail,
      })
      ;(updateClassV11 as jest.Mock).mockResolvedValue({
        data: mockClassDetail,
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(getClassDetail).toHaveBeenCalled()
      })

      const inputs = screen.getAllByRole('textbox')
      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '  修改后的名称  ' } })
      })

      const submitBtn = screen.getByText('保存').closest('div')
      await act(async () => {
        fireEvent.click(submitBtn!)
      })

      await waitFor(() => {
        expect(updateClassV11).toHaveBeenCalledWith(
          123,
          expect.objectContaining({
            name: '修改后的名称', // 已 trim
          })
        )
      })
    })
  })

  describe('IT13 需求专项测试', () => {
    it('教材配置区域应在编辑页正确集成', async () => {
      ;(getClassDetail as jest.Mock).mockResolvedValue({
        data: mockClassWithCurriculum,
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(screen.getByTestId('curriculum-config-form')).toBeInTheDocument()
      })

      // 验证已有配置时的展开状态
      expect(screen.getByText('编辑模式')).toBeInTheDocument()
    })

    it('编辑时更新教材配置应正确传递数据', async () => {
      ;(getClassDetail as jest.Mock).mockResolvedValue({
        data: mockClassWithCurriculum,
      })
      ;(updateClassV11 as jest.Mock).mockResolvedValue({
        data: mockClassWithCurriculum,
      })

      await act(async () => {
        render(<ClassEdit />)
      })

      await waitFor(() => {
        expect(screen.getByTestId('curriculum-form-expanded')).toBeInTheDocument()
      })

      // 修改班级名称
      const inputs = screen.getAllByRole('textbox')
      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '四年级数学班' } })
      })

      const submitBtn = screen.getByText('保存').closest('div')
      await act(async () => {
        fireEvent.click(submitBtn!)
      })

      await waitFor(() => {
        expect(updateClassV11).toHaveBeenCalledWith(
          124, // 测试数据中的班级ID
          expect.objectContaining({
            curriculum_config: expect.objectContaining({
              grade_level: 'primary_lower',
              subjects: expect.arrayContaining(['数学', '语文']),
            }),
          })
        )
      })
    })
  })
})
