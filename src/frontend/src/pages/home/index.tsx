import { View, Text } from '@tarojs/components'
import { useUserStore, usePersonaStore } from '@/store'
import TeacherDashboard from '@/components/TeacherDashboard'
import StudentHome from '@/components/StudentHome'
import Taro from '@tarojs/taro'
import './index.scss'

export default function Home() {
  const { userInfo } = useUserStore()
  const { currentPersona } = usePersonaStore()

  const isTeacher = currentPersona?.role === 'teacher'

  /** 切换分身 */
  const handleSwitchPersona = () => {
    Taro.navigateTo({ url: '/pages/persona-select/index' })
  }

  return (
    <View className='home-page'>
      {/* 顶部分身信息 */}
      <View className='home-page__header'>
        <View className='home-page__header-top'>
          <View className='home-page__persona-info'>
            <Text className='home-page__greeting'>
              {currentPersona?.nickname || userInfo?.nickname || '用户'}
            </Text>
            {isTeacher && currentPersona?.school && (
              <Text className='home-page__school'>{currentPersona.school}</Text>
            )}
            {currentPersona?.description && (
              <Text className='home-page__description'>{currentPersona.description}</Text>
            )}
          </View>
          <View className='home-page__persona-switch' onClick={handleSwitchPersona}>
            <Text className='home-page__persona-switch-text'>
              {isTeacher ? '切换分身' : '切换身份'}
            </Text>
          </View>
        </View>
      </View>

      {/* 根据角色渲染不同组件 */}
      {isTeacher ? (
        <TeacherDashboard personaId={currentPersona?.id || 0} />
      ) : (
        <StudentHome />
      )}

    </View>
  )
}
