/**
 * ClassCreate 班级创建页面单元测试
 * 模块: FE-IT13-002 - 小程序-班级编辑页教材配置区域
 */

import React from 'react'
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import '@testing-library/jest-dom'
import ClassCreate from '../index'
import { server, mockCreateClassV11Response } from '@/__mocks__/server'
import { http, HttpResponse } from 'msw'
import Taro from '@tarojs/taro'

// Mock Taro useRouter
const mockRouter = { params: {} }
jest.mock('@tarojs/taro', () => ({
  ...jest.requireActual('@tarojs/taro'),
  useRouter: () => mockRouter,
  getCurrentInstance: () => ({
    router: { params: {} },
  }),
}))

// Mock API
jest.mock('@/api/class', () => ({
  createClassV11: jest.fn(),
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
          grade_level: 'primary_lower',
          grade: '三年级',
          subjects: ['数学'],
          textbook_versions: ['人教版'],
          custom_textbooks: [],
          current_progress: '第三单元',
        })
      }
    }, [])

    return (
      <div data-testid="curriculum-config-form">
        <div
          data-testid="curriculum-toggle"
          onClick={() => onExpandedChange && onExpandedChange(!expanded)}
        >
          {expanded ? '📚 教材配置（已展开）' : '📚 教材配置（可选）'}
        </div>
        {expanded && <div data-testid="curriculum-form-expanded">教材配置表单</div>}
      </div>
    )
  }
})

import { createClassV11 } from '@/api/class'

describe('ClassCreate 班级创建页面', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockRouter.params = {}
  })

  describe('渲染测试', () => {
    it('应正确渲染班级创建页面基本结构', async () => {
      await act(async () => {
        render(<ClassCreate />)
      })

      // 验证页面标题
      expect(screen.getByText('创建班级')).toBeInTheDocument()
      expect(screen.getByText('创建一个班级来管理你的学生')).toBeInTheDocument()

      // 验证表单区域标题
      expect(screen.getByText('分身信息')).toBeInTheDocument()
      expect(screen.getByText('班级信息')).toBeInTheDocument()

      // 验证必填字段标签
      expect(screen.getByText(/分身昵称/)).toBeInTheDocument()
      expect(screen.getByText(/学校名称/)).toBeInTheDocument()
      expect(screen.getByText(/分身描述/)).toBeInTheDocument()
      expect(screen.getByText(/班级名称/)).toBeInTheDocument()

      // 验证公开设置
      expect(screen.getByText('公开班级')).toBeInTheDocument()
      expect(screen.getByText(/公开班级对所有学生可见/)).toBeInTheDocument()
    })

    it('应显示来自注册流程的标题', async () => {
      mockRouter.params = { from: 'register' }

      await act(async () => {
        render(<ClassCreate />)
      })

      expect(screen.getByText('创建你的第一个班级')).toBeInTheDocument()
      expect(screen.getByText('创建班级后即可邀请学生加入')).toBeInTheDocument()
    })

    it('应渲染教材配置表单组件', async () => {
      await act(async () => {
        render(<ClassCreate />)
      })

      expect(screen.getByTestId('curriculum-config-form')).toBeInTheDocument()
      expect(screen.getByText('📚 教材配置（可选）')).toBeInTheDocument()
    })

    it('创建按钮默认应为禁用状态', async () => {
      await act(async () => {
        render(<ClassCreate />)
      })

      // 检查创建按钮
      const submitBtn = screen.getByText('创建班级').closest('div')
      expect(submitBtn).toHaveClass('class-create__submit--disabled')
    })

    it('应显示角色计数统计', async () => {
      await act(async () => {
        render(<ClassCreate />)
      })

      // 验证字符计数器
      expect(screen.getAllByText(/\/200/)[0]).toBeInTheDocument()
      expect(screen.getAllByText(/\/200/)[1]).toBeInTheDocument()
    })
  })

  describe('表单输入测试', () => {
    it('输入班级名称应正确更新', async () => {
      await act(async () => {
        render(<ClassCreate />)
      })

      const inputs = screen.getAllByRole('textbox')
      // 班级名称是第4个输入框（前面有3个分身相关输入）
      await act(async () => {
        fireEvent.change(inputs[3], { target: { value: '测试班级' } })
      })

      expect(inputs[3]).toHaveValue('测试班级')
    })

    it('填写所有必填字段后提交按钮应启用', async () => {
      await act(async () => {
        render(<ClassCreate />)
      })

      const inputs = screen.getAllByRole('textbox')

      // 填写分身昵称
      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '王老师' } })
      })

      // 填写学校名称
      await act(async () => {
        fireEvent.change(inputs[1], { target: { value: '实验小学' } })
      })

      // 填写分身描述
      await act(async () => {
        fireEvent.change(inputs[2], { target: { value: '10年教学经验' } })
      })

      // 填写班级名称
      await act(async () => {
        fireEvent.change(inputs[3], { target: { value: '三年级数学班' } })
      })

      // 检查提交按钮已启用
      const submitBtn = screen.getByText('创建班级').closest('div')
      expect(submitBtn).not.toHaveClass('class-create__submit--disabled')
    })

    it('班级名称超过50个字符应截断', async () => {
      await act(async () => {
        render(<ClassCreate />)
      })

      const inputs = screen.getAllByRole('textbox')
      const longName = 'a'.repeat(60)

      await act(async () => {
        fireEvent.change(inputs[3], { target: { value: longName } })
      })

      // maxlength=50 应该限制输入
      expect(inputs[3]).toHaveAttribute('maxlength', '50')
    })

    it('切换公开开关应更新状态', async () => {
      await act(async () => {
        render(<ClassCreate />)
      })

      const switchInput = screen.getByRole('checkbox')
      expect(switchInput).toBeChecked() // 默认公开

      await act(async () => {
        fireEvent.click(switchInput)
      })

      expect(switchInput).not.toBeChecked()
    })
  })

  describe('教材配置交互测试', () => {
    it('点击教材配置区域应展开表单', async () => {
      await act(async () => {
        render(<ClassCreate />)
      })

      expect(screen.queryByTestId('curriculum-form-expanded')).not.toBeInTheDocument()

      await act(async () => {
        fireEvent.click(screen.getByTestId('curriculum-toggle'))
      })

      expect(screen.getByTestId('curriculum-form-expanded')).toBeInTheDocument()
    })
  })

  describe('API 调用测试', () => {
    it('提交表单应调用 createClassV11 API', async () => {
      ;(createClassV11 as jest.Mock).mockResolvedValue({
        data: mockCreateClassV11Response,
      })

      await act(async () => {
        render(<ClassCreate />)
      })

      const inputs = screen.getAllByRole('textbox')

      // 填写所有必填字段
      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '王老师' } })
        fireEvent.change(inputs[1], { target: { value: '实验小学' } })
        fireEvent.change(inputs[2], { target: { value: '10年教学经验' } })
        fireEvent.change(inputs[3], { target: { value: '三年级数学班' } })
      })

      // 点击提交按钮
      const submitBtn = screen.getByText('创建班级').closest('div')
      await act(async () => {
        fireEvent.click(submitBtn!)
      })

      await waitFor(() => {
        expect(createClassV11).toHaveBeenCalledWith(
          expect.objectContaining({
            name: '三年级数学班',
            persona_nickname: '王老师',
            persona_school: '实验小学',
            persona_description: '10年教学经验',
            is_public: true,
          })
        )
      })
    })

    it('展开教材配置并填写后端应包含教材配置', async () => {
      ;(createClassV11 as jest.Mock).mockResolvedValue({
        data: mockCreateClassV11Response,
      })

      await act(async () => {
        render(<ClassCreate />)
      })

      // 展开教材配置
      await act(async () => {
        fireEvent.click(screen.getByTestId('curriculum-toggle'))
      })

      const inputs = screen.getAllByRole('textbox')

      // 填写所有必填字段
      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '王老师' } })
        fireEvent.change(inputs[1], { target: { value: '实验小学' } })
        fireEvent.change(inputs[2], { target: { value: '10年教学经验' } })
        fireEvent.change(inputs[3], { target: { value: '三年级数学班' } })
      })

      // 点击提交
      const submitBtn = screen.getByText('创建班级').closest('div')
      await act(async () => {
        fireEvent.click(submitBtn!)
      })

      await waitFor(() => {
        expect(createClassV11).toHaveBeenCalledWith(
          expect.objectContaining({
            curriculum_config: expect.objectContaining({
              grade_level: 'primary_lower',
              subjects: ['数学'],
              textbook_versions: ['人教版'],
            }),
          })
        )
      })
    })

    it('API 调用失败时应显示错误提示', async () => {
      const errorMessage = '创建班级失败: 网络错误'
      ;(createClassV11 as jest.Mock).mockRejectedValue(new Error(errorMessage))

      await act(async () => {
        render(<ClassCreate />)
      })

      const inputs = screen.getAllByRole('textbox')

      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '王老师' } })
        fireEvent.change(inputs[1], { target: { value: '实验小学' } })
        fireEvent.change(inputs[2], { target: { value: '10年教学经验' } })
        fireEvent.change(inputs[3], { target: { value: '三年级数学班' } })
      })

      const submitBtn = screen.getByText('创建班级').closest('div')
      await act(async () => {
        fireEvent.click(submitBtn!)
      })

      await waitFor(() => {
        expect(Taro.showToast).toHaveBeenCalledWith(
          expect.objectContaining({
            title: expect.stringContaining(errorMessage),
            icon: 'none',
            duration: 3000,
          })
        )
      })
    })

    it('班级名称重复时应显示特定错误提示', async () => {
      ;(createClassV11 as jest.Mock).mockRejectedValue({
        message: '该班级名称已存在',
      })

      await act(async () => {
        render(<ClassCreate />)
      })

      const inputs = screen.getAllByRole('textbox')

      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '王老师' } })
        fireEvent.change(inputs[1], { target: { value: '实验小学' } })
        fireEvent.change(inputs[2], { target: { value: '10年教学经验' } })
        fireEvent.change(inputs[3], { target: { value: '重复班级名' } })
      })

      const submitBtn = screen.getByText('创建班级').closest('div')
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
  })

  describe('创建成功页面测试', () => {
    it('创建成功应显示成功页面', async () => {
      ;(createClassV11 as jest.Mock).mockResolvedValue({
        data: mockCreateClassV11Response,
      })

      await act(async () => {
        render(<ClassCreate />)
      })

      const inputs = screen.getAllByRole('textbox')

      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '王老师' } })
        fireEvent.change(inputs[1], { target: { value: '实验小学' } })
        fireEvent.change(inputs[2], { target: { value: '10年教学经验' } })
        fireEvent.change(inputs[3], { target: { value: '三年级数学班' } })
      })

      const submitBtn = screen.getByText('创建班级').closest('div')
      await act(async () => {
        fireEvent.click(submitBtn!)
      })

      await waitFor(() => {
        expect(screen.getByText('班级创建成功！')).toBeInTheDocument()
        expect(screen.getByText('已同步创建班级专属分身')).toBeInTheDocument()
      })

      // 验证分身信息卡片
      expect(screen.getByText('班级分身信息')).toBeInTheDocument()
      expect(screen.getByText('王老师')).toBeInTheDocument()
      expect(screen.getByText('实验小学')).toBeInTheDocument()
    })

    it('点击复制链接按钮应复制分享链接', async () => {
      ;(createClassV11 as jest.Mock).mockResolvedValue({
        data: mockCreateClassV11Response,
      })

      await act(async () => {
        render(<ClassCreate />)
      })

      const inputs = screen.getAllByRole('textbox')

      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '王老师' } })
        fireEvent.change(inputs[1], { target: { value: '实验小学' } })
        fireEvent.change(inputs[2], { target: { value: '10年教学经验' } })
        fireEvent.change(inputs[3], { target: { value: '三年级数学班' } })
      })

      const submitBtn = screen.getByText('创建班级').closest('div')
      await act(async () => {
        fireEvent.click(submitBtn!)
      })

      await waitFor(() => {
        expect(screen.getByText('复制链接')).toBeInTheDocument()
      })

      const copyLinkBtn = screen.getByText('复制链接')
      await act(async () => {
        fireEvent.click(copyLinkBtn)
      })

      expect(Taro.setClipboardData).toHaveBeenCalledWith(
        expect.objectContaining({
          data: mockCreateClassV11Response.share_url,
        })
      )
    })

    it('点击返回按钮应导航到首页', async () => {
      ;(createClassV11 as jest.Mock).mockResolvedValue({
        data: mockCreateClassV11Response,
      })

      await act(async () => {
        render(<ClassCreate />)
      })

      const inputs = screen.getAllByRole('textbox')

      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '王老师' } })
        fireEvent.change(inputs[1], { target: { value: '实验小学' } })
        fireEvent.change(inputs[2], { target: { value: '10年教学经验' } })
        fireEvent.change(inputs[3], { target: { value: '三年级数学班' } })
      })

      const submitBtn = screen.getByText('创建班级').closest('div')
      await act(async () => {
        fireEvent.click(submitBtn!)
      })

      await waitFor(() => {
        expect(screen.getByText('返回')).toBeInTheDocument()
      })

      const backBtn = screen.getByText('返回')
      await act(async () => {
        fireEvent.click(backBtn)
      })

      expect(Taro.navigateBack).toHaveBeenCalled()
    })

    it('从注册流程进入创建成功后应跳转到首页', async () => {
      mockRouter.params = { from: 'register' }
      ;(createClassV11 as jest.Mock).mockResolvedValue({
        data: mockCreateClassV11Response,
      })

      await act(async () => {
        render(<ClassCreate />)
      })

      const inputs = screen.getAllByRole('textbox')

      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '王老师' } })
        fireEvent.change(inputs[1], { target: { value: '实验小学' } })
        fireEvent.change(inputs[2], { target: { value: '10年教学经验' } })
        fireEvent.change(inputs[3], { target: { value: '三年级数学班' } })
      })

      const submitBtn = screen.getByText('创建班级').closest('div')
      await act(async () => {
        fireEvent.click(submitBtn!)
      })

      await waitFor(() => {
        expect(screen.getByText('进入首页')).toBeInTheDocument()
      })

      const homeBtn = screen.getByText('进入首页')
      await act(async () => {
        fireEvent.click(homeBtn)
      })

      expect(Taro.switchTab).toHaveBeenCalledWith({
        url: '/pages/home/index',
      })
    })
  })

  describe('边界条件测试', () => {
    it('输入空白字符应正常处理', async () => {
      await act(async () => {
        render(<ClassCreate />)
      })

      const inputs = screen.getAllByRole('textbox')

      // 填写带空格的数据
      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '  王老师  ' } })
        fireEvent.change(inputs[1], { target: { value: '  实验小学  ' } })
        fireEvent.change(inputs[2], { target: { value: '  10年教学经验  ' } })
        fireEvent.change(inputs[3], { target: { value: '  三年级数学班  ' } })
      })

      ;(createClassV11 as jest.Mock).mockResolvedValue({
        data: mockCreateClassV11Response,
      })

      const submitBtn = screen.getByText('创建班级').closest('div')
      await act(async () => {
        fireEvent.click(submitBtn!)
      })

      await waitFor(() => {
        expect(createClassV11).toHaveBeenCalledWith(
          expect.objectContaining({
            name: '三年级数学班', // 已trim
            persona_nickname: '王老师',
            persona_school: '实验小学',
            persona_description: '10年教学经验',
          })
        )
      })
    })

    it('部分填写表单不应允许提交', async () => {
      await act(async () => {
        render(<ClassCreate />)
      })

      const inputs = screen.getAllByRole('textbox')

      // 只填写部分字段
      await act(async () => {
        fireEvent.change(inputs[0], { target: { value: '王老师' } })
        fireEvent.change(inputs[3], { target: { value: '三年级数学班' } })
      })

      // 提交按钮仍应为禁用状态
      const submitBtn = screen.getByText('创建班级').closest('div')
      expect(submitBtn).toHaveClass('class-create__submit--disabled')
    })
  })
})
