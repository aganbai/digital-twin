import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import ClassCreatePage from '../index'
import Taro from '@tarojs/taro'
import '@testing-library/jest-dom'

// Mock Taro
jest.mock('@tarojs/taro', () => ({
  ...jest.requireActual('@tarojs/taro'),
  useRouter: () => ({
    params: {},
  }),
  showToast: jest.fn(),
  navigateBack: jest.fn(),
  switchTab: jest.fn(),
  setClipboardData: jest.fn(),
}))

describe('ClassCreate Page', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('render', () => {
    it('应正确渲染班级创建页面', () => {
      render(<ClassCreatePage />)

      expect(screen.getByText('创建班级')).toBeInTheDocument()
      expect(screen.getByText('分身信息')).toBeInTheDocument()
      expect(screen.getByText('班级信息')).toBeInTheDocument()
    })

    it('应显示必填标记', () => {
      render(<ClassCreatePage />)

      expect(screen.getByText('分身昵称')).toBeInTheDocument()
      expect(screen.getByText('学校名称')).toBeInTheDocument()
      expect(screen.getByText('分身描述')).toBeInTheDocument()
      expect(screen.getByText('班级名称')).toBeInTheDocument()
    })

    it('应显示教材配置区域', () => {
      render(<ClassCreatePage />)

      expect(screen.getByText(/📚 教材配置/)).toBeInTheDocument()
    })
  })

  describe('input validation', () => {
    it('初始状态提交按钮应禁用', () => {
      render(<ClassCreatePage />)

      const submitBtn = screen.getByText('创建班级').closest('.class-create__submit')
      expect(submitBtn).toHaveClass('class-create__submit--disabled')
    })

    it('填写必填项后提交按钮应启用', () => {
      render(<ClassCreatePage />)

      // 获取输入框并填写值
      const inputs = document.querySelectorAll('input, textarea')
      const inputs_array = Array.from(inputs) as HTMLInputElement[]

      // 填写分身昵称
      fireEvent.change(inputs_array.find(i => i.placeholder?.includes('分身昵称'))!, {
        target: { value: '王老师' },
      })

      // 填写学校名称
      fireEvent.change(inputs_array.find(i => i.placeholder?.includes('学校名称'))!, {
        target: { value: '实验小学' },
      })

      // 填写分身描述
      fireEvent.change(inputs_array.find(i => i.placeholder?.includes('分身描述'))!, {
        target: { value: '10年教学经验' },
      })

      // 填写班级名称
      fireEvent.change(inputs_array.find(i => i.placeholder?.includes('班级名称'))!, {
        target: { value: '三年级数学班' },
      })

      const submitBtn = screen.getByText('创建班级').closest('.class-create__submit')
      expect(submitBtn).not.toHaveClass('class-create__submit--disabled')
    })
  })

  describe('curriculum config integration', () => {
    it('点击教材配置区域应展开', () => {
      render(<ClassCreatePage />)

      const curriculumHeader = screen.getByText(/📚 教材配置/).closest('.curriculum-form__header')
      fireEvent.click(curriculumHeader!)

      expect(screen.getByText('选择学段')).toBeInTheDocument()
    })

    it('展开教材配置后应显示学段选择', () => {
      render(<ClassCreatePage />)

      const curriculumHeader = screen.getByText(/📚 教材配置/).closest('.curriculum-form__header')
      fireEvent.click(curriculumHeader!)

      expect(screen.getByText(/教材配置/)).toBeInTheDocument()
      expect(document.querySelector('.curriculum-form__content')).toBeInTheDocument()
    })
  })

  describe('public switch', () => {
    it('应正确渲染公开班级开关', () => {
      render(<ClassCreatePage />)

      expect(screen.getByText('公开班级')).toBeInTheDocument()
      expect(screen.getByText(/公开班级对所有学生可见/)).toBeInTheDocument()
    })
  })

  describe('description character count', () => {
    it('应显示字符计数', () => {
      render(<ClassCreatePage />)

      const textareas = document.querySelectorAll('textarea')
      const descArea = Array.from(textareas).find(t =>
        t.placeholder?.includes('分身描述')
      )

      // 初始应为 0/200
      expect(screen.getByText(/0\/200/)).toBeInTheDocument()

      // 输入后计数应更新
      fireEvent.change(descArea!, { target: { value: '测试描述' } })

      expect(screen.getByText(/4\/200/)).toBeInTheDocument()
    })
  })
})
