import { useState, useCallback } from 'react'
import { View, Text } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { getPersonas, switchPersona } from '@/api/persona'
import type { Persona } from '@/api/persona'
import { useUserStore, usePersonaStore } from '@/store'
import './index.scss'

export default function PersonaSelect() {
  const [personas, setPersonas] = useState<Persona[]>([])
  const [loading, setLoading] = useState(false)
  const [switching, setSwitching] = useState(false)
  const { setToken, setUserInfo } = useUserStore()
  const { setCurrentPersona, setPersonas: setStorePersonas } = usePersonaStore()

  /** 获取分身列表 */
  const fetchPersonas = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getPersonas()
      const list = res.data.personas || []
      setPersonas(list)
      setStorePersonas(list, res.data.default_persona_id)
    } catch (error) {
      console.error('获取分身列表失败:', error)
    } finally {
      setLoading(false)
    }
  }, [setStorePersonas])

  useDidShow(() => {
    fetchPersonas()
  })

  /** 教师分身列表 */
  const teacherPersonas = personas.filter((p) => p.role === 'teacher' && p.is_active)
  /** 学生分身列表 */
  const studentPersonas = personas.filter((p) => p.role === 'student' && p.is_active)

  /** 点击分身 → 切换 */
  const handleSwitch = async (persona: Persona) => {
    if (switching) return
    setSwitching(true)
    try {
      const res = await switchPersona(persona.id)
      const { token, persona: switchedPersona } = res.data

      // 更新 token 和用户信息
      setToken(token)
      setUserInfo({
        id: persona.id,
        nickname: switchedPersona.nickname,
        role: switchedPersona.role,
      })
      setCurrentPersona(switchedPersona)

      // 根据角色跳转
      if (switchedPersona.role === 'student') {
        Taro.switchTab({ url: '/pages/home/index' })
      } else if (switchedPersona.role === 'teacher') {
        Taro.redirectTo({ url: '/pages/knowledge/index' })
      }
    } catch (error) {
      console.error('切换分身失败:', error)
    } finally {
      setSwitching(false)
    }
  }

  /** 跳转创建新分身 */
  const handleCreate = () => {
    Taro.navigateTo({ url: '/pages/role-select/index' })
  }

  return (
    <View className='persona-select'>
      <View className='persona-select__header'>
        <Text className='persona-select__title'>选择分身</Text>
        <Text className='persona-select__subtitle'>请选择一个分身进入</Text>
      </View>

      {loading ? (
        <View className='persona-select__loading'>
          <Text className='persona-select__loading-text'>加载中...</Text>
        </View>
      ) : (
        <View className='persona-select__content'>
          {/* 教师分身 */}
          {teacherPersonas.length > 0 && (
            <View className='persona-select__section'>
              <Text className='persona-select__section-title'>教师分身</Text>
              {teacherPersonas.map((p) => (
                <View
                  key={p.id}
                  className='persona-select__card'
                  onClick={() => handleSwitch(p)}
                >
                  <View className='persona-select__card-icon'>👨‍🏫</View>
                  <View className='persona-select__card-info'>
                    <Text className='persona-select__card-name'>{p.nickname}</Text>
                    {p.school && (
                      <Text className='persona-select__card-school'>{p.school}</Text>
                    )}
                    {p.description && (
                      <Text className='persona-select__card-desc'>{p.description}</Text>
                    )}
                  </View>
                  <View className='persona-select__card-tag persona-select__card-tag--teacher'>
                    <Text className='persona-select__card-tag-text'>教师</Text>
                  </View>
                </View>
              ))}
            </View>
          )}

          {/* 学生分身 */}
          {studentPersonas.length > 0 && (
            <View className='persona-select__section'>
              <Text className='persona-select__section-title'>学生分身</Text>
              {studentPersonas.map((p) => (
                <View
                  key={p.id}
                  className='persona-select__card'
                  onClick={() => handleSwitch(p)}
                >
                  <View className='persona-select__card-icon'>👨‍🎓</View>
                  <View className='persona-select__card-info'>
                    <Text className='persona-select__card-name'>{p.nickname}</Text>
                  </View>
                  <View className='persona-select__card-tag persona-select__card-tag--student'>
                    <Text className='persona-select__card-tag-text'>学生</Text>
                  </View>
                </View>
              ))}
            </View>
          )}

          {/* 空状态 */}
          {teacherPersonas.length === 0 && studentPersonas.length === 0 && (
            <View className='persona-select__empty'>
              <Text className='persona-select__empty-text'>暂无可用分身</Text>
            </View>
          )}
        </View>
      )}

      {/* 创建新分身按钮 */}
      <View className='persona-select__footer'>
        <View
          className={`persona-select__create-btn ${switching ? 'persona-select__create-btn--disabled' : ''}`}
          onClick={switching ? undefined : handleCreate}
        >
          <Text className='persona-select__create-btn-text'>+ 创建新分身</Text>
        </View>
      </View>
    </View>
  )
}
