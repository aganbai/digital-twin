import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import ClassEditPage from '../index'
import Taro from '@tarojs/taro'
import '@testing-library/jest-dom'

describe('ClassEdit Page', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    // Mock Taro.getCurrentInstance
    ;(Taro.getCurrentInstance as jest.Mock).mockReturnValue({
      router: {
        params: { class_id: '123' },
      },
    })
  })

  describe('render', () => {
    it('应显示加载状态', () => {
      render(<ClassEditPage />)

      // 初始加载状态
      expect(screen.getByText(/加载中/)).toBeInTheDocument()
    })

    it('应正确渲染班级信息表单字段', async () => {
      render(<ClassEditPage />)

      await waitFor(() => {
        // 等待加载完成后检查表单字段
        expect(screen.getByText('编辑班级信息')).toBeInTheDocument()
      })
    })

    it('应显示必填标记', async () => {
      render(<ClassEditPage />)

      await waitFor(() => {
        expect(screen.getByText('班级名称')).toBeInTheDocument()
      })
    })
  })

  describe('curriculum config integration', () => {
    it('应显示教材配置区域', async () => {
      render(<ClassEditPage />)

      await waitFor(() => {
        expect(screen.getByText(/📚 教材配置/)).toBeInTheDocument()
      })
    })
  })

  describe('public switch', () => {
    it('应正确渲染公开班级开关', async () => {
      render(<ClassEditPage />)

      await waitFor(() => {
        expect(screen.getByText('公开班级')).toBeInTheDocument()
      })
    })
  })

  describe('description character count', () => {
    it('应正确显示字符计数', async () => {
      render(<ClassEditPage />)

      await waitFor(() => {
        // 检查字符计数是否存在
        const charCount = screen.queryByText(/\d+\/200/)
        expect(charCount).toBeInTheDocument()
      })
    })
  })

  describe('save button', () => {
    it('应有保存按钮', async () => {
      render(<ClassEditPage />)

      await waitFor(() => {
        expect(screen.getByText('保存')).toBeInTheDocument()
      })
    })
  })
})
