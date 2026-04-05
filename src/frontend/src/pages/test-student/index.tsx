import { useState, useEffect } from 'react'
import { View, Text } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { getTestStudent, resetTestStudent, loginTestStudent, TestStudentInfo } from '@/api/test-student'
import './index.scss'

export default function TestStudentPage() {
  const [loading, setLoading] = useState(false)
  const [testStudent, setTestStudent] = useState<TestStudentInfo | null>(null)
  const [error, setError] = useState(false)

  /** 获取自测学生信息 */
  const fetchTestStudent = async () => {
    setLoading(true)
    setError(false)
    try {
      const res = await getTestStudent()
      setTestStudent(res.data)
    } catch (err) {
      console.error('[TestStudent] 获取自测学生信息失败:', err)
      setError(true)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchTestStudent()
  }, [])

  /** 格式化创建时间 */
  const formatTime = (time: string): string => {
    if (!time) return '-'
    const date = new Date(time)
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
  }

  /** 重置数据 */
  const handleReset = () => {
    Taro.showModal({
      title: '确认重置',
      content: '确定要重置自测学生的所有对话和记忆吗？此操作不可恢复。',
      confirmText: '确定重置',
      confirmColor: '#EF4444',
      success: async (res) => {
        if (res.confirm) {
          Taro.showLoading({ title: '重置中...' })
          try {
            const result = await resetTestStudent()
            Taro.hideLoading()
            Taro.showModal({
              title: '重置成功',
              content: `已清空 ${result.data.cleared_conversations} 条对话记录，${result.data.cleared_memories} 条记忆。`,
              showCancel: false,
            })
          } catch (err: any) {
            Taro.hideLoading()
            Taro.showToast({ title: err?.message || '重置失败', icon: 'none' })
          }
        }
      },
    })
  }

  /** 模拟登录 */
  const handleLogin = () => {
    Taro.showModal({
      title: '模拟登录',
      content: '将以自测学生身份登录，当前账号将被退出。确定继续吗？',
      confirmText: '确定登录',
      success: async (res) => {
        if (res.confirm) {
          Taro.showLoading({ title: '登录中...' })
          try {
            const result = await loginTestStudent()
            Taro.hideLoading()
            // 保存token和用户信息
            Taro.setStorageSync('token', result.data.token)
            Taro.setStorageSync('userInfo', {
              id: result.data.user_id,
              username: result.data.username,
              nickname: result.data.nickname,
              role: 'student',
            })
            Taro.showToast({ title: '登录成功', icon: 'success' })
            // 跳转到首页
            setTimeout(() => {
              Taro.reLaunch({ url: '/pages/home/index' })
            }, 1000)
          } catch (err: any) {
            Taro.hideLoading()
            Taro.showToast({ title: err?.message || '登录失败', icon: 'none' })
          }
        }
      },
    })
  }

  // 加载状态
  if (loading && !testStudent) {
    return (
      <View className='test-student-page'>
        <View className='test-student-page__loading'>
          <Text className='test-student-page__loading-text'>加载中...</Text>
        </View>
      </View>
    )
  }

  // 错误状态
  if (error && !testStudent) {
    return (
      <View className='test-student-page'>
        <View className='test-student-page__error'>
          <Text className='test-student-page__error-text'>加载失败</Text>
          <View className='test-student-page__error-btn' onClick={fetchTestStudent}>
            <Text className='test-student-page__error-btn-text'>重试</Text>
          </View>
        </View>
      </View>
    )
  }

  return (
    <View className='test-student-page'>
      {/* 学生信息卡片 */}
      {testStudent && (
        <View className='test-student-page__card'>
          <View className='test-student-page__card-header'>
            <Text className='test-student-page__card-title'>测试学生信息</Text>
          </View>
          
          <View className='test-student-page__info-item'>
            <Text className='test-student-page__info-label'>用户名</Text>
            <Text className='test-student-page__info-value'>{testStudent.username}</Text>
          </View>
          
          <View className='test-student-page__info-item'>
            <Text className='test-student-page__info-label'>昵称</Text>
            <Text className='test-student-page__info-value'>{testStudent.nickname}</Text>
          </View>
          
          <View className='test-student-page__info-item'>
            <Text className='test-student-page__info-label'>密码提示</Text>
            <Text className='test-student-page__info-value test-student-page__info-value--hint'>
              {testStudent.password_hint || '-'}
            </Text>
          </View>
          
          <View className='test-student-page__info-item'>
            <Text className='test-student-page__info-label'>状态</Text>
            <View className={`test-student-page__status ${testStudent.is_active ? 'test-student-page__status--active' : 'test-student-page__status--inactive'}`}>
              <Text className='test-student-page__status-text'>
                {testStudent.is_active ? '正常' : '已禁用'}
              </Text>
            </View>
          </View>
          
          <View className='test-student-page__info-item'>
            <Text className='test-student-page__info-label'>创建时间</Text>
            <Text className='test-student-page__info-value'>{formatTime(testStudent.created_at)}</Text>
          </View>
        </View>
      )}

      {/* 已加入班级 */}
      {testStudent && testStudent.joined_classes && testStudent.joined_classes.length > 0 && (
        <View className='test-student-page__card'>
          <View className='test-student-page__card-header'>
            <Text className='test-student-page__card-title'>已加入班级</Text>
          </View>
          
          {testStudent.joined_classes.map((cls) => (
            <View key={cls.class_id} className='test-student-page__class-item'>
              <View className='test-student-page__class-icon'>
                <Text className='test-student-page__class-icon-text'>📚</Text>
              </View>
              <View className='test-student-page__class-info'>
                <Text className='test-student-page__class-name'>{cls.class_name}</Text>
                <Text className='test-student-page__class-id'>ID: {cls.class_id}</Text>
              </View>
            </View>
          ))}
        </View>
      )}

      {/* 无班级提示 */}
      {testStudent && (!testStudent.joined_classes || testStudent.joined_classes.length === 0) && (
        <View className='test-student-page__empty'>
          <Text className='test-student-page__empty-text'>暂未加入任何班级</Text>
        </View>
      )}

      {/* 操作按钮 */}
      <View className='test-student-page__actions'>
        <View className='test-student-page__btn test-student-page__btn--primary' onClick={handleLogin}>
          <Text className='test-student-page__btn-text'>模拟登录</Text>
        </View>
        <View className='test-student-page__btn test-student-page__btn--danger' onClick={handleReset}>
          <Text className='test-student-page__btn-text'>重置数据</Text>
        </View>
      </View>

      {/* 说明文字 */}
      <View className='test-student-page__tips'>
        <Text className='test-student-page__tips-title'>📌 说明</Text>
        <Text className='test-student-page__tips-text'>• 自测学生是系统为您自动创建的学生角色</Text>
        <Text className='test-student-page__tips-text'>• 使用模拟登录功能可以以学生身份体验系统</Text>
        <Text className='test-student-page__tips-text'>• 重置数据将清空所有对话记录和记忆</Text>
      </View>
    </View>
  )
}
