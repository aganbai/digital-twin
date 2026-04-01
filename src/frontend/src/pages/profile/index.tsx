import { useState } from 'react'
import { View, Text } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { getUserProfile, UserProfile } from '@/api/user'
import { useUserStore } from '@/store'
import './index.scss'

/** 功能列表项类型 */
interface MenuItem {
  key: string
  label: string
  /** 可见角色，不传表示所有角色可见 */
  roles?: string[]
  /** 点击处理 */
  action: () => void
  /** 是否为危险操作（红色文字） */
  danger?: boolean
}

export default function Profile() {
  const { userInfo, logout, setUserInfo } = useUserStore()
  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [loading, setLoading] = useState(false)

  /** 获取用户详情 */
  const fetchProfile = async () => {
    setLoading(true)
    try {
      const res = await getUserProfile()
      setProfile(res.data)
      // 同步更新 Zustand store 中的 userInfo，确保角色等信息与 API 返回一致
      if (res.data) {
        setUserInfo({
          id: res.data.id,
          nickname: res.data.nickname,
          role: res.data.role,
        })
      }
    } catch (error) {
      console.error('获取用户信息失败:', error)
    } finally {
      setLoading(false)
    }
  }

  /** 每次页面显示时刷新用户信息 */
  useDidShow(() => {
    fetchProfile()
  })

  /** 获取用户昵称首字母（用于头像占位） */
  const getInitial = (): string => {
    const name = profile?.nickname || userInfo?.nickname || '?'
    return name.charAt(0).toUpperCase()
  }

  /** 判断是否为教师角色（优先使用 API 返回的 profile 数据） */
  const isTeacher = (): boolean => {
    const role = profile?.role ?? userInfo?.role
    return role === 'teacher'
  }

  /** 判断是否为学生角色（优先使用 API 返回的 profile 数据） */
  const isStudent = (): boolean => {
    const role = profile?.role ?? userInfo?.role
    return role === 'student'
  }

  /** 获取角色显示文本 */
  const getRoleLabel = (): string => {
    return isTeacher() ? '教师' : '学生'
  }

  /** 处理退出登录 */
  const handleLogout = () => {
    Taro.showModal({
      title: '提示',
      content: '确定要退出登录吗？',
      confirmText: '退出',
      confirmColor: '#EF4444',
      success: (res) => {
        if (res.confirm) {
          logout()
          Taro.redirectTo({ url: '/pages/login/index' })
        }
      },
    })
  }

  /** 显示关于系统信息 */
  const handleAbout = () => {
    Taro.showModal({
      title: '关于系统',
      content: 'AI 数字分身教学系统\n版本：v2.0.0\n基于大语言模型的智能教学辅助平台',
      showCancel: false,
      confirmText: '知道了',
    })
  }

  /** 功能列表配置 */
  const menuItems: MenuItem[] = [
    {
      key: 'persona-overview',
      label: '分身概览',
      roles: ['teacher'],
      action: () => Taro.navigateTo({ url: '/pages/persona-overview/index' }),
    },
    {
      key: 'assignments',
      label: '作业管理',
      roles: ['teacher'],
      action: () => Taro.navigateTo({ url: '/pages/assignment-list/index' }),
    },
    {
      key: 'share-manage',
      label: '分享管理',
      roles: ['teacher'],
      action: () => Taro.navigateTo({ url: '/pages/share-manage/index' }),
    },
    {
      key: 'memories',
      label: '我的记忆',
      roles: ['student'],
      action: () => Taro.navigateTo({ url: '/pages/memories/index' }),
    },
    {
      key: 'my-assignments',
      label: '我的作业',
      roles: ['student'],
      action: () => Taro.navigateTo({ url: '/pages/my-assignments/index' }),
    },
    {
      key: 'about',
      label: '关于系统',
      action: handleAbout,
    },
    {
      key: 'logout',
      label: '退出登录',
      action: handleLogout,
      danger: true,
    },
  ]

  /** 根据角色过滤可见的菜单项（优先使用 API 返回的 profile 数据） */
  const visibleMenuItems = menuItems.filter((item) => {
    if (!item.roles) return true
    const role = profile?.role ?? userInfo?.role ?? ''
    return item.roles.includes(role)
  })

  return (
    <View className='profile-page'>
      {/* 用户信息区域 */}
      <View className='profile-page__header'>
        <View className='profile-page__avatar'>
          <Text className='profile-page__avatar-text'>{getInitial()}</Text>
        </View>
        <View className='profile-page__info'>
          <Text className='profile-page__nickname'>
            {profile?.nickname || userInfo?.nickname || '加载中...'}
          </Text>
          <View
            className={`profile-page__role-tag ${isTeacher() ? 'profile-page__role-tag--teacher' : 'profile-page__role-tag--student'}`}
          >
            <Text className='profile-page__role-text'>{getRoleLabel()}</Text>
          </View>
        </View>
      </View>

      {/* 统计信息 */}
      <View className='profile-page__stats'>
        {isStudent() && (
          <>
            <View className='profile-page__stat-item'>
              <Text className='profile-page__stat-num'>
                {profile?.stats?.conversation_count ?? '-'}
              </Text>
              <Text className='profile-page__stat-label'>对话数</Text>
            </View>
            <View className='profile-page__stat-divider' />
            <View className='profile-page__stat-item'>
              <Text className='profile-page__stat-num'>
                {profile?.stats?.memory_count ?? '-'}
              </Text>
              <Text className='profile-page__stat-label'>记忆数</Text>
            </View>
          </>
        )}
        {isTeacher() && (
          <>
            <View className='profile-page__stat-item'>
              <Text className='profile-page__stat-num'>
                {profile?.stats?.document_count ?? '-'}
              </Text>
              <Text className='profile-page__stat-label'>文档数</Text>
            </View>
            <View className='profile-page__stat-divider' />
            <View className='profile-page__stat-item'>
              <Text className='profile-page__stat-num'>
                {profile?.stats?.conversation_count ?? '-'}
              </Text>
              <Text className='profile-page__stat-label'>被提问数</Text>
            </View>
          </>
        )}
      </View>

      {/* 功能列表 */}
      <View className='profile-page__menu'>
        {visibleMenuItems.map((item, index) => (
          <View
            key={item.key}
            className={`profile-page__menu-item ${item.danger ? 'profile-page__menu-item--danger' : ''}`}
            onClick={item.action}
          >
            <Text
              className={`profile-page__menu-label ${item.danger ? 'profile-page__menu-label--danger' : ''}`}
            >
              {item.label}
            </Text>
            {!item.danger && (
              <Text className='profile-page__menu-arrow'>›</Text>
            )}
          </View>
        ))}
      </View>
    </View>
  )
}
