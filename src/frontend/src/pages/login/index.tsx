import { useState } from 'react'
import { View, Text, Button } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { wxLogin } from '@/api/auth'
import { useUserStore } from '@/store'
import './index.scss'

export default function Login() {
  const [loading, setLoading] = useState(false)
  const { setToken, setUserInfo } = useUserStore()

  /** 处理微信登录 */
  const handleLogin = async () => {
    if (loading) return
    setLoading(true)

    try {
      const res = await wxLogin()
      const { token, is_new_user, role, nickname, user_id } = res.data

      // 存储 token
      setToken(token)

      if (is_new_user) {
        // 新用户 → 跳转角色选择页
        Taro.redirectTo({ url: '/pages/role-select/index' })
      } else {
        // 老用户 → 存储用户信息并根据角色跳转
        setUserInfo({ id: user_id, nickname, role })

        if (role === 'student') {
          Taro.switchTab({ url: '/pages/home/index' })
        } else if (role === 'teacher') {
          Taro.redirectTo({ url: '/pages/knowledge/index' })
        }
      }
    } catch (error) {
      Taro.showToast({
        title: '登录失败，请重试',
        icon: 'none',
        duration: 2000,
      })
    } finally {
      setLoading(false)
    }
  }

  return (
    <View className='login'>
      <View className='login__content'>
        {/* Logo 区域 */}
        <View className='login__logo'>
          <View className='login__logo-icon'>AI</View>
        </View>

        {/* 应用名称 */}
        <Text className='login__title'>AI 数字分身</Text>

        {/* 应用简介 */}
        <Text className='login__subtitle'>
          基于苏格拉底式教学的 AI 数字分身
        </Text>
      </View>

      {/* 底部操作区域 */}
      <View className='login__footer'>
        {/* 微信登录按钮 */}
        <Button
          className='login__btn'
          onClick={handleLogin}
          disabled={loading}
        >
          {loading ? '登录中...' : '微信登录'}
        </Button>

        {/* 用户协议 */}
        <View className='login__agreement'>
          <Text className='login__agreement-text'>
            登录即表示同意
          </Text>
          <Text className='login__agreement-link'>《用户协议》</Text>
          <Text className='login__agreement-text'>和</Text>
          <Text className='login__agreement-link'>《隐私政策》</Text>
        </View>
      </View>
    </View>
  )
}
