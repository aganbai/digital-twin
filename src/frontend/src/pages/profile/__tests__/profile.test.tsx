/**
 * Profile 个人中心页面单元测试
 * 模块: FE-IT13-003 - 小程序-个人中心移除教材配置入口
 */

import React from 'react'
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import '@testing-library/jest-dom'
import Profile from '../index'
import Taro from '@tarojs/taro'

// Mock user store
const mockLogout = jest.fn()
const mockSetUserInfo = jest.fn()

jest.mock('@/store', () => ({
  useUserStore: () => ({
    userInfo: {
      id: 1,
      nickname: '测试用户',
      role: 'teacher',
    },
    logout: mockLogout,
    setUserInfo: mockSetUserInfo,
  }),
}))

// Mock user API
jest.mock('@/api/user', () => ({
  getUserProfile: jest.fn(),
}))

import { getUserProfile } from '@/api/user'

// Mock console methods to reduce noise
const originalConsoleError = console.error
beforeAll(() => {
  console.error = jest.fn()
})

afterAll(() => {
  console.error = originalConsoleError
})

describe('Profile 个人中心页面', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('渲染测试', () => {
    it('应正确渲染页面基本结构', async () => {
      ;(getUserProfile as jest.Mock).mockResolvedValue({
        data: mockUserProfile,
      })

      await act(async () => {
        render(<Profile />)
      })

      // 验证头像区域
      expect(screen.getByText('测试用户')).toBeInTheDocument()

      // 验证角色标签
      expect(screen.getByText('教师')).toBeInTheDocument()

      // 验证切换角色按钮
      expect(screen.getByText('切换角色 ›')).toBeInTheDocument()
    })

    it('应正确渲染教师角色的菜单项', async () => {
      ;(getUserProfile as jest.Mock).mockResolvedValue({
        data: mockUserProfile,
      })

      await act(async () => {
        render(<Profile />)
      })

      // 教师特有菜单
      expect(screen.getByText('分身概览')).toBeInTheDocument()
      expect(screen.getByText('自测学生')).toBeInTheDocument()
      expect(screen.getByText('分享管理')).toBeInTheDocument()
      expect(screen.getByText('反馈管理')).toBeInTheDocument()

      // 通用菜单
      expect(screen.getByText('意见反馈')).toBeInTheDocument()
      expect(screen.getByText('关于系统')).toBeInTheDocument()
      expect(screen.getByText('退出登录')).toBeInTheDocument()
    })

    it('应正确渲染学生角色的菜单项', async () => {
      ;(getUserProfile as jest.Mock).mockResolvedValue({
        data: mockStudentProfile,
      })

      await act(async () => {
        render(<Profile />)
      })

      // 等待数据加载
      await waitFor(() => {
        expect(screen.getByText('学生')).toBeInTheDocument()
      })

      // 学生特有菜单
      expect(screen.getByText('我的记忆')).toBeInTheDocument()

      // 学生不应看到的教师菜单
      expect(screen.queryByText('分身概览')).not.toBeInTheDocument()
      expect(screen.queryByText('自测学生')).not.toBeInTheDocument()
      expect(screen.queryByText('分享管理')).not.toBeInTheDocument()
      expect(screen.queryByText('反馈管理')).not.toBeInTheDocument()

      // 通用菜单
      expect(screen.getByText('意见反馈')).toBeInTheDocument()
      expect(screen.getByText('关于系统')).toBeInTheDocument()
      expect(screen.getByText('退出登录')).toBeInTheDocument()
    })

    it('不应显示教材配置菜单项（IT13需求验证）', async () => {
      ;(getUserProfile as jest.Mock).mockResolvedValue({
        data: mockUserProfile,
      })

      await act(async () => {
        render(<Profile />)
      })

      // 关键断言：教材配置菜单项应该被移除
      expect(screen.queryByText('教材配置')).not.toBeInTheDocument()
      expect(screen.queryByText(/教材/)).not.toBeInTheDocument()
    })

    it('应正确显示教师的统计信息', async () => {
      ;(getUserProfile as jest.Mock).mockResolvedValue({
        data: mockUserProfile,
      })

      await act(async () => {
        render(<Profile />)
      })

      // 教师统计项
      expect(screen.getByText('文档数')).toBeInTheDocument()
      expect(screen.getByText('被提问数')).toBeInTheDocument()
      expect(screen.getByText('20')).toBeInTheDocument() // document_count
      expect(screen.getByText('10')).toBeInTheDocument() // conversation_count
    })

    it('应正确显示学生的统计信息', async () => {
      ;(getUserProfile as jest.Mock).mockResolvedValue({
        data: mockStudentProfile,
      })

      await act(async () => {
        render(<Profile />)
      })

      // 等待学生角色标签显示
      await waitFor(() => {
        expect(screen.getByText('学生')).toBeInTheDocument()
      })

      // 学生统计项
      expect(screen.getByText('对话数')).toBeInTheDocument()
      expect(screen.getByText('记忆数')).toBeInTheDocument()
      expect(screen.getByText('15')).toBeInTheDocument() // conversation_count
      expect(screen.getByText('8')).toBeInTheDocument() // memory_count
    })
  })

  describe('交互测试', () => {
    it('点击切换角色应跳转到角色选择页面', async () => {
      ;(getUserProfile as jest.Mock).mockResolvedValue({
        data: mockUserProfile,
      })

      await act(async () => {
        render(<Profile />)
      })

      const switchRoleBtn = screen.getByText('切换角色 ›')
      await act(async () => {
        fireEvent.click(switchRoleBtn)
      })

      expect(Taro.navigateTo).toHaveBeenCalledWith({
        url: '/pages/persona-select/index',
      })
    })

    it('点击分身概览应跳转到对应页面', async () => {
      ;(getUserProfile as jest.Mock).mockResolvedValue({
        data: mockUserProfile,
      })

      await act(async () => {
        render(<Profile />)
      })

      const personaOverviewBtn = screen.getByText('分身概览')
      await act(async () => {
        fireEvent.click(personaOverviewBtn)
      })

      expect(Taro.navigateTo).toHaveBeenCalledWith({
        url: '/pages/persona-overview/index',
      })
    })

    it('点击分享管理应跳转到对应页面', async () => {
      ;(getUserProfile as jest.Mock).mockResolvedValue({
        data: mockUserProfile,
      })

      await act(async () => {
        render(<Profile />)
      })

      const shareManageBtn = screen.getByText('分享管理')
      await act(async () => {
        fireEvent.click(shareManageBtn)
      })

      expect(Taro.navigateTo).toHaveBeenCalledWith({
        url: '/pages/share-manage/index',
      })
    })

    it('点击意见反馈应跳转到对应页面', async () => {
      ;(getUserProfile as jest.Mock).mockResolvedValue({
        data: mockUserProfile,
      })

      await act(async () => {
        render(<Profile />)
      })

      const feedbackBtn = screen.getByText('意见反馈')
      await act(async () => {
        fireEvent.click(feedbackBtn)
      })

      expect(Taro.navigateTo).toHaveBeenCalledWith({
        url: '/pages/feedback/index',
      })
    })

    it('点击关于系统应显示模态框', async () => {
      ;(getUserProfile as jest.Mock).mockResolvedValue({
        data: mockUserProfile,
      })

      await act(async () => {
        render(<Profile />)
      })

      const aboutBtn = screen.getByText('关于系统')
      await act(async () => {
        fireEvent.click(aboutBtn)
      })

      expect(Taro.showModal).toHaveBeenCalledWith({
        title: '关于系统',
        content: 'AI 数字分身教学系统\n版本：v2.0.0\n基于大语言模型的智能教学辅助平台',
        showCancel: false,
        confirmText: '知道了',
      })
    })

    it('点击退出登录应显示确认模态框', async () => {
      ;(getUserProfile as jest.Mock).mockResolvedValue({
        data: mockUserProfile,
      })

      await act(async () => {
        render(<Profile />)
      })

      const logoutBtn = screen.getByText('退出登录')
      await act(async () => {
        fireEvent.click(logoutBtn)
      })

      expect(Taro.showModal).toHaveBeenCalledWith({
        title: '提示',
        content: '确定要退出登录吗？',
        confirmText: '退出',
        confirmColor: '#EF4444',
        success: expect.any(Function),
      })
    })

    it('确认退出登录应执行登出操作并跳转', async () => {
      // 模拟用户点击确认
      ;(Taro.showModal as jest.Mock).mockImplementation(({ success }) => {
        success({ confirm: true, cancel: false })
      })

      ;(getUserProfile as jest.Mock).mockResolvedValue({
        data: mockUserProfile,
      })

      await act(async () => {
        render(<Profile />)
      })

      const logoutBtn = screen.getByText('退出登录')
      await act(async () => {
        fireEvent.click(logoutBtn)
      })

      expect(mockLogout).toHaveBeenCalled()
      expect(Taro.redirectTo).toHaveBeenCalledWith({
        url: '/pages/login/index',
      })
    })
  })

  describe('API 错误处理测试', () => {
    it('API 请求失败时应显示错误提示', async () => {
      ;(getUserProfile as jest.Mock).mockRejectedValue(new Error('Network error'))

      await act(async () => {
        render(<Profile />)
      })

      // 等待错误状态更新
      await waitFor(() => {
        expect(screen.getByText('加载失败，请点击重试')).toBeInTheDocument()
      })

      expect(screen.getByText('重试')).toBeInTheDocument()
    })

    it('点击重试应重新加载用户数据', async () => {
      ;(getUserProfile as jest.Mock)
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce({ data: mockUserProfile })

      await act(async () => {
        render(<Profile />)
      })

      // 等待错误状态
      await waitFor(() => {
        expect(screen.getByText('加载失败，请点击重试')).toBeInTheDocument()
      })

      // 点击重试
      const retryBtn = screen.getByText('重试')
      await act(async () => {
        fireEvent.click(retryBtn)
      })

      // 验证重新加载成功
      await waitFor(() => {
        expect(screen.getByText('测试用户')).toBeInTheDocument()
      })
    })
  })

  describe('数据同步测试', () => {
    it('获取用户资料后应同步更新 store', async () => {
      const profileData = {
        id: 1,
        nickname: '张三',
        role: 'teacher',
        stats: {
          conversation_count: 5,
          document_count: 10,
        },
      }

      ;(getUserProfile as jest.Mock).mockResolvedValue({
        data: profileData,
      })

      await act(async () => {
        render(<Profile />)
      })

      await waitFor(() => {
        expect(mockSetUserInfo).toHaveBeenCalledWith({
          id: 1,
          nickname: '张三',
          role: 'teacher',
        })
      })
    })

    it('当 API 返回的角色为空时应使用 store 中的角色', async () => {
      const profileData = {
        id: 1,
        nickname: '李四',
        role: '',
        stats: {},
      }

      ;(getUserProfile as jest.Mock).mockResolvedValue({
        data: profileData,
      })

      await act(async () => {
        render(<Profile />)
      })

      await waitFor(() => {
        expect(mockSetUserInfo).toHaveBeenCalledWith({
          id: 1,
          nickname: '李四',
          role: 'teacher', // 来自 store 的默认值
        })
      })
    })
  })

  describe('IT13 需求专项测试', () => {
    it('教材配置入口已从菜单配置中移除', async () => {
      ;(getUserProfile as jest.Mock).mockResolvedValue({
        data: mockUserProfile,
      })

      await act(async () => {
        render(<Profile />)
      })

      // 获取所有菜单项文本
      const menuItems = screen.getAllByText(
        /分身概览|自测学生|分享管理|我的记忆|意见反馈|反馈管理|关于系统|退出登录/
      )

      // 验证菜单项数量（教师应有 7 个菜单项）
      expect(menuItems.length).toBe(7)

      // 确保教材配置相关文本不存在
      const allText = document.body.textContent
      expect(allText).not.toMatch(/curriculum|教材配置/)
    })

    it('代码注释应包含 IT13 需求说明', () => {
      // 读取源代码文件验证注释
      const fs = require('fs')
      const path = require('path')
      const sourceFile = path.join(__dirname, '../index.tsx')
      const content = fs.readFileSync(sourceFile, 'utf-8')

      // 验证存在 IT13 相关注释
      expect(content).toContain('IT13')
      expect(content).toContain('教材配置')
      expect(content).toContain('curriculum-config')
    })

    it('菜单配置中应包含说明注释', () => {
      const fs = require('fs')
      const path = require('path')
      const sourceFile = path.join(__dirname, '../index.tsx')
      const content = fs.readFileSync(sourceFile, 'utf-8')

      // 验证注释说明教材配置入口已移除
      expect(content).toContain('教材配置入口已移除')
    })
  })
})

// Mock 数据
const mockUserProfile = {
  id: 1,
  username: 'test_user',
  nickname: '测试用户',
  role: 'teacher',
  email: 'test@example.com',
  created_at: '2026-01-01T00:00:00Z',
  stats: {
    conversation_count: 10,
    memory_count: 5,
    document_count: 20,
  },
}

const mockStudentProfile = {
  id: 2,
  username: 'student_user',
  nickname: '学生用户',
  role: 'student',
  email: 'student@example.com',
  created_at: '2026-01-01T00:00:00Z',
  stats: {
    conversation_count: 15,
    memory_count: 8,
  },
}
