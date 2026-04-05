import { useState, useEffect } from 'react'
import { View, Text, Image } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { usePersonaStore } from '@/store'
import './index.scss'

/** Tab 配置项 */
interface TabItem {
  /** 页面路径（不含前缀） */
  pagePath: string
  /** Tab 文字 */
  text: string
  /** 未选中图标（emoji 占位，实际可替换为图片路径） */
  icon: string
  /** 选中图标 */
  selectedIcon: string
  /** 角标数量（0 表示不显示） */
  badge?: number
  /** 是否显示红点 */
  showDot?: boolean
}

/** 教师端 Tab 列表 */
const TEACHER_TABS: TabItem[] = [
  {
    pagePath: 'pages/chat-list/index',
    text: '聊天',
    icon: '💬',
    selectedIcon: '💬',
  },
  {
    pagePath: 'pages/teacher-students/index',
    text: '学生管理',
    icon: '👥',
    selectedIcon: '👥',
  },
  {
    pagePath: 'pages/knowledge/index',
    text: '知识库',
    icon: '📚',
    selectedIcon: '📚',
  },
  {
    pagePath: 'pages/profile/index',
    text: '我的',
    icon: '👤',
    selectedIcon: '👤',
  },
]

/** 学生端 Tab 列表 */
const STUDENT_TABS: TabItem[] = [
  {
    pagePath: 'pages/home/index',
    text: '对话',
    icon: '💬',
    selectedIcon: '💬',
  },
  {
    pagePath: 'pages/discover/index',
    text: '发现',
    icon: '🌐',
    selectedIcon: '🌐',
  },
  {
    pagePath: 'pages/profile/index',
    text: '我的',
    icon: '👤',
    selectedIcon: '👤',
  },
]

/** 自定义 TabBar 组件 Props */
interface CustomTabBarProps {
  /** 待审批数量（教师端"学生"Tab 角标） */
  pendingCount?: number
}

/**
 * 自定义 TabBar 组件
 * 根据用户角色（teacher/student）展示不同 Tab 列表
 * 教师端：聊天列表 / 学生管理 / 知识库 / 我的
 * 学生端：对话 / 发现 / 我的
 */
export default function CustomTabBar(props: CustomTabBarProps) {
  const { pendingCount = 0 } = props
  const { currentPersona } = usePersonaStore()
  const [selectedIndex, setSelectedIndex] = useState(0)

  const isTeacher = currentPersona?.role === 'teacher'
  const tabs = isTeacher ? TEACHER_TABS : STUDENT_TABS

  /** 页面显示时同步选中状态 */
  useDidShow(() => {
    try {
      const pages = Taro.getCurrentPages()
      if (pages.length > 0) {
        const currentPage = pages[pages.length - 1]
        const route = currentPage.route || ''
        const index = tabs.findIndex((tab) => route.includes(tab.pagePath))
        if (index >= 0) {
          setSelectedIndex(index)
        }
      }
    } catch (e) {
      // 忽略获取页面栈异常
    }
  })

  /** 点击 Tab 切换页面 */
  const handleTabClick = (index: number) => {
    if (index === selectedIndex) return
    const tab = tabs[index]
    setSelectedIndex(index)
    Taro.switchTab({ url: `/${tab.pagePath}` })
  }

  /** 获取 Tab 的角标数量 */
  const getBadge = (tab: TabItem, index: number): number => {
    // 教师端"学生"Tab 显示待审批数量
    if (isTeacher && index === 1) {
      return pendingCount
    }
    return tab.badge || 0
  }

  return (
    <View className='custom-tabbar'>
      <View className='custom-tabbar__border' />
      <View className='custom-tabbar__content'>
        {tabs.map((tab, index) => {
          const isSelected = index === selectedIndex
          const badge = getBadge(tab, index)
          return (
            <View
              key={tab.pagePath}
              className='custom-tabbar__item'
              onClick={() => handleTabClick(index)}
            >
              <View className='custom-tabbar__icon-wrap'>
                <Text className='custom-tabbar__icon'>
                  {isSelected ? tab.selectedIcon : tab.icon}
                </Text>
                {/* 角标：数字 */}
                {badge > 0 && (
                  <View className='custom-tabbar__badge'>
                    <Text className='custom-tabbar__badge-text'>
                      {badge > 99 ? '99+' : badge}
                    </Text>
                  </View>
                )}
                {/* 红点 */}
                {tab.showDot && badge === 0 && (
                  <View className='custom-tabbar__dot' />
                )}
              </View>
              <Text
                className={`custom-tabbar__text ${isSelected ? 'custom-tabbar__text--active' : ''}`}
              >
                {tab.text}
              </Text>
            </View>
          )
        })}
      </View>
    </View>
  )
}
