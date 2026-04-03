import { useState, useCallback, useEffect } from 'react'
import { View, Text, Picker } from '@tarojs/components'
import Taro, { useDidShow, usePullDownRefresh } from '@tarojs/taro'
import { getMemories } from '@/api/memory'
import type { Memory, MemoryLayer, MEMORY_LAYER_LABELS } from '@/api/memory'
import { getTeachers } from '@/api/teacher'
import type { Teacher } from '@/api/teacher'
import Empty from '@/components/Empty'
import { formatTime } from '@/utils/format'
import './index.scss'

/** 记忆层级选项 */
const MEMORY_LAYERS = [
  { key: '', label: '全部' },
  { key: 'core', label: '核心记忆' },
  { key: 'episodic', label: '情景记忆' },
  { key: 'archived', label: '已归档' },
]

/** 记忆类型中文映射 */
const MEMORY_TYPE_MAP: Record<string, string> = {
  conversation: '对话记忆',
  learning_progress: '学习进度',
  personality_traits: '个性特征',
  preference: '偏好',
  ability: '能力',
  interaction: '互动',
  general: '通用',
}

/** 重要性等级配置 */
const IMPORTANCE_CONFIG: Record<string, { label: string; className: string }> = {
  high: { label: '高', className: 'high' },
  medium: { label: '中', className: 'medium' },
  low: { label: '低', className: 'low' },
}

/** 根据数值获取重要性等级 */
function getImportanceLevel(importance: number): string {
  if (importance >= 0.7) return 'high'
  if (importance >= 0.4) return 'medium'
  return 'low'
}

/** 记忆层级标签映射 */
const LAYER_LABEL_MAP: Record<string, string> = {
  core: '核心',
  episodic: '情景',
  archived: '归档',
}

export default function Memories() {
  const [teachers, setTeachers] = useState<Teacher[]>([])
  const [selectedTeacherIndex, setSelectedTeacherIndex] = useState(0)
  const [activeLayer, setActiveLayer] = useState('')
  const [memories, setMemories] = useState<Memory[]>([])
  const [loading, setLoading] = useState(false)

  /** 获取教师列表 */
  const fetchTeachers = useCallback(async () => {
    try {
      const res = await getTeachers(1, 100)
      setTeachers(res.data.items || [])
    } catch (error) {
      console.error('获取教师列表失败:', error)
    }
  }, [])

  /** 获取记忆列表 */
  const fetchMemories = useCallback(async (teacherPersonaId: number, layer?: string) => {
    setLoading(true)
    try {
      // 传入 teacher_persona_id，student_persona_id 传 0（后端会从 JWT 自动补全）
      const res = await getMemories(teacherPersonaId, 0, (layer || undefined) as any, 1, 20)
      setMemories(res.data.items || [])
    } catch (error) {
      console.error('获取记忆列表失败:', error)
      setMemories([])
    } finally {
      setLoading(false)
    }
  }, [])

  /** 页面显示时获取教师列表 */
  useDidShow(() => {
    fetchTeachers()
  })

  /** 教师列表加载后，自动获取第一个教师的记忆 */
  useEffect(() => {
    if (teachers.length > 0) {
      const teacher = teachers[selectedTeacherIndex]
      if (teacher && teacher.persona_id) {
        fetchMemories(teacher.persona_id, activeLayer)
      }
    }
  }, [teachers, selectedTeacherIndex, activeLayer, fetchMemories])

  /** 下拉刷新 */
  usePullDownRefresh(async () => {
    if (teachers.length > 0) {
      const teacher = teachers[selectedTeacherIndex]
      if (teacher && teacher.persona_id) {
        await fetchMemories(teacher.persona_id, activeLayer)
      }
    }
    Taro.stopPullDownRefresh()
  })

  /** 切换教师 */
  const handleTeacherChange = (e: any) => {
    const index = Number(e.detail.value)
    setSelectedTeacherIndex(index)
  }

  /** 切换记忆层级 */
  const handleLayerChange = (layerKey: string) => {
    setActiveLayer(layerKey)
  }

  /** 获取重要性标签配置 */
  const getImportanceBadge = (importance: number) => {
    const level = getImportanceLevel(importance)
    return IMPORTANCE_CONFIG[level] || IMPORTANCE_CONFIG.low
  }

  /** 当前选中的教师昵称 */
  const selectedTeacherName = teachers.length > 0
    ? teachers[selectedTeacherIndex]?.nickname || '请选择教师'
    : '暂无教师'

  return (
    <View className='memories-page'>
      {/* 教师筛选 */}
      <View className='memories-page__filter'>
        <Text className='memories-page__filter-label'>选择教师：</Text>
        <Picker
          mode='selector'
          range={teachers.map((t) => t.nickname)}
          value={selectedTeacherIndex}
          onChange={handleTeacherChange}
        >
          <View className='memories-page__picker'>
            <Text className='memories-page__picker-text'>{selectedTeacherName}</Text>
            <Text className='memories-page__picker-arrow'>▼</Text>
          </View>
        </Picker>
      </View>

      {/* 记忆层级 Tab */}
      <View className='memories-page__tabs'>
        {MEMORY_LAYERS.map((item) => (
          <View
            key={item.key}
            className={`memories-page__tab ${activeLayer === item.key ? 'memories-page__tab--active' : ''}`}
            onClick={() => handleLayerChange(item.key)}
          >
            <Text className={`memories-page__tab-text ${activeLayer === item.key ? 'memories-page__tab-text--active' : ''}`}>
              {item.label}
            </Text>
          </View>
        ))}
      </View>

      {/* 记忆列表 / 空状态 / 加载状态 */}
      <View className='memories-page__content'>
        {loading ? (
          <View className='memories-page__loading'>
            <Text className='memories-page__loading-text'>加载中...</Text>
          </View>
        ) : memories.length > 0 ? (
          <View className='memories-page__list'>
            {memories.map((memory) => {
              const badge = getImportanceBadge(memory.importance)
              const layerLabel = LAYER_LABEL_MAP[memory.memory_layer] || memory.memory_layer
              return (
                <View key={memory.id} className='memories-page__item'>
                  <View className='memories-page__item-header'>
                    {memory.memory_type && (
                      <Text className='memories-page__item-type'>
                        {MEMORY_TYPE_MAP[memory.memory_type] || memory.memory_type}
                      </Text>
                    )}
                    {memory.memory_layer && (
                      <Text className={`memories-page__item-layer memories-page__item-layer--${memory.memory_layer}`}>
                        {layerLabel}
                      </Text>
                    )}
                    <View className={`memories-page__importance memories-page__importance--${badge.className}`}>
                      <Text className={`memories-page__importance-text memories-page__importance-text--${badge.className}`}>
                        {badge.label}
                      </Text>
                    </View>
                  </View>
                  <Text className='memories-page__item-content'>{memory.content}</Text>
                  <Text className='memories-page__item-time'>{formatTime(memory.created_at)}</Text>
                </View>
              )
            })}
          </View>
        ) : (
          <Empty text='暂无记忆记录' />
        )}
      </View>
    </View>
  )
}
