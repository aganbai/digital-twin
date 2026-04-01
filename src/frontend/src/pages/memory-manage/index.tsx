import { useState, useEffect, useCallback } from 'react'
import { View, Text, Input, Textarea } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import {
  getMemories, updateMemory, deleteMemory, summarizeMemories,
  Memory, MemoryLayer, MEMORY_LAYER_LABELS,
} from '@/api/memory'
import { usePersonaStore } from '@/store'
import { MEMORY_TYPE_LABELS } from '@/utils/constants'
import { formatTime } from '@/utils/format'
import Empty from '@/components/Empty'
import './index.scss'

export default function MemoryManage() {
  const router = useRouter()
  const studentPersonaId = Number(router.params.student_persona_id) || 0
  const studentName = decodeURIComponent(router.params.student_name || '学生')
  const { currentPersona } = usePersonaStore()

  const [memories, setMemories] = useState<Memory[]>([])
  const [loading, setLoading] = useState(true)
  const [activeLayer, setActiveLayer] = useState<MemoryLayer | ''>('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [summarizing, setSummarizing] = useState(false)

  // 编辑状态
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editContent, setEditContent] = useState('')
  const [saving, setSaving] = useState(false)

  /** 设置导航栏标题 */
  useEffect(() => {
    Taro.setNavigationBarTitle({ title: `${studentName} 的记忆` })
  }, [studentName])

  /** 获取记忆列表 */
  const fetchMemories = useCallback(async () => {
    if (!currentPersona?.id || !studentPersonaId) return
    setLoading(true)
    try {
      const res = await getMemories(
        currentPersona.id,
        studentPersonaId,
        activeLayer || undefined,
        page,
        20,
      )
      setMemories(res.data.items || [])
      setTotal(res.data.total || 0)
    } catch (error) {
      console.error('获取记忆失败:', error)
    } finally {
      setLoading(false)
    }
  }, [currentPersona?.id, studentPersonaId, activeLayer, page])

  useEffect(() => {
    fetchMemories()
  }, [fetchMemories])

  /** 切换层级筛选 */
  const handleLayerChange = (layer: MemoryLayer | '') => {
    setActiveLayer(layer)
    setPage(1)
  }

  /** 开始编辑 */
  const handleStartEdit = (memory: Memory) => {
    setEditingId(memory.id)
    setEditContent(memory.content)
  }

  /** 取消编辑 */
  const handleCancelEdit = () => {
    setEditingId(null)
    setEditContent('')
  }

  /** 保存编辑 */
  const handleSaveEdit = async (memory: Memory) => {
    if (!editContent.trim()) {
      Taro.showToast({ title: '内容不能为空', icon: 'none' })
      return
    }
    setSaving(true)
    try {
      await updateMemory(memory.id, { content: editContent.trim() })
      Taro.showToast({ title: '保存成功', icon: 'success' })
      setEditingId(null)
      setEditContent('')
      fetchMemories()
    } catch (error) {
      console.error('保存记忆失败:', error)
    } finally {
      setSaving(false)
    }
  }

  /** 删除记忆 */
  const handleDelete = (memoryId: number) => {
    Taro.showModal({
      title: '确认删除',
      content: '删除后无法恢复，确定要删除这条记忆吗？',
      confirmColor: '#EF4444',
      success: async (res) => {
        if (res.confirm) {
          try {
            await deleteMemory(memoryId)
            Taro.showToast({ title: '已删除', icon: 'success' })
            fetchMemories()
          } catch (error) {
            console.error('删除记忆失败:', error)
          }
        }
      },
    })
  }

  /** 触发摘要合并 */
  const handleSummarize = async () => {
    if (!currentPersona?.id || !studentPersonaId) return
    Taro.showModal({
      title: '摘要合并',
      content: '将情景记忆压缩合并为核心记忆，原始记忆会被归档。确定执行？',
      success: async (res) => {
        if (res.confirm) {
          setSummarizing(true)
          try {
            const result = await summarizeMemories(currentPersona.id, studentPersonaId)
            Taro.showToast({
              title: `合并完成：${result.data.summarized_count} 条 → ${result.data.new_core_memories.length} 条核心记忆`,
              icon: 'none',
              duration: 3000,
            })
            fetchMemories()
          } catch (error) {
            console.error('摘要合并失败:', error)
          } finally {
            setSummarizing(false)
          }
        }
      },
    })
  }

  /** 获取层级标签颜色 */
  const getLayerColor = (layer: MemoryLayer): string => {
    switch (layer) {
      case 'core': return '#4F46E5'
      case 'episodic': return '#10B981'
      case 'archived': return '#9CA3AF'
      default: return '#6B7280'
    }
  }

  const layers: Array<{ key: MemoryLayer | ''; label: string }> = [
    { key: '', label: '全部' },
    { key: 'core', label: '核心' },
    { key: 'episodic', label: '情景' },
    { key: 'archived', label: '已归档' },
  ]

  return (
    <View className='memory-manage'>
      {/* 层级筛选 */}
      <View className='memory-manage__filter'>
        {layers.map((item) => (
          <View
            key={item.key}
            className={`memory-manage__filter-item ${activeLayer === item.key ? 'memory-manage__filter-item--active' : ''}`}
            onClick={() => handleLayerChange(item.key)}
          >
            <Text className={`memory-manage__filter-text ${activeLayer === item.key ? 'memory-manage__filter-text--active' : ''}`}>
              {item.label}
            </Text>
          </View>
        ))}
      </View>

      {/* 操作按钮 */}
      <View className='memory-manage__actions'>
        <Text className='memory-manage__count'>共 {total} 条记忆</Text>
        <View
          className={`memory-manage__summarize-btn ${summarizing ? 'memory-manage__summarize-btn--disabled' : ''}`}
          onClick={summarizing ? undefined : handleSummarize}
        >
          <Text className='memory-manage__summarize-text'>
            {summarizing ? '合并中...' : '🧠 摘要合并'}
          </Text>
        </View>
      </View>

      {/* 记忆列表 */}
      {loading ? (
        <View className='memory-manage__loading'>
          <Text className='memory-manage__loading-text'>加载中...</Text>
        </View>
      ) : memories.length > 0 ? (
        <View className='memory-manage__list'>
          {memories.map((memory) => (
            <View key={memory.id} className='memory-manage__card'>
              {/* 标签行 */}
              <View className='memory-manage__card-tags'>
                <View
                  className='memory-manage__layer-tag'
                  style={{ backgroundColor: `${getLayerColor(memory.memory_layer)}20`, borderColor: getLayerColor(memory.memory_layer) }}
                >
                  <Text
                    className='memory-manage__layer-tag-text'
                    style={{ color: getLayerColor(memory.memory_layer) }}
                  >
                    {MEMORY_LAYER_LABELS[memory.memory_layer]}
                  </Text>
                </View>
                <Text className='memory-manage__type-tag'>
                  {MEMORY_TYPE_LABELS[memory.memory_type] || memory.memory_type}
                </Text>
                <Text className='memory-manage__importance'>
                  ⭐ {(memory.importance * 100).toFixed(0)}%
                </Text>
              </View>

              {/* 内容区域 */}
              {editingId === memory.id ? (
                <View className='memory-manage__edit-area'>
                  <Textarea
                    className='memory-manage__edit-textarea'
                    value={editContent}
                    maxlength={2000}
                    onInput={(e) => setEditContent(e.detail.value)}
                    autoFocus
                  />
                  <View className='memory-manage__edit-actions'>
                    <View className='memory-manage__edit-btn memory-manage__edit-btn--cancel' onClick={handleCancelEdit}>
                      <Text className='memory-manage__edit-btn-text'>取消</Text>
                    </View>
                    <View
                      className={`memory-manage__edit-btn memory-manage__edit-btn--save ${saving ? 'memory-manage__edit-btn--disabled' : ''}`}
                      onClick={saving ? undefined : () => handleSaveEdit(memory)}
                    >
                      <Text className='memory-manage__edit-btn-text--save'>
                        {saving ? '保存中...' : '保存'}
                      </Text>
                    </View>
                  </View>
                </View>
              ) : (
                <Text className='memory-manage__card-content'>{memory.content}</Text>
              )}

              {/* 底部操作 */}
              <View className='memory-manage__card-footer'>
                <Text className='memory-manage__card-time'>
                  {formatTime(memory.updated_at || memory.created_at)}
                </Text>
                {memory.memory_layer !== 'archived' && editingId !== memory.id && (
                  <View className='memory-manage__card-ops'>
                    <Text
                      className='memory-manage__op-btn memory-manage__op-btn--edit'
                      onClick={() => handleStartEdit(memory)}
                    >
                      编辑
                    </Text>
                    <Text
                      className='memory-manage__op-btn memory-manage__op-btn--delete'
                      onClick={() => handleDelete(memory.id)}
                    >
                      删除
                    </Text>
                  </View>
                )}
              </View>
            </View>
          ))}
        </View>
      ) : (
        <Empty text='暂无记忆数据' />
      )}

      {/* 分页 */}
      {total > 20 && (
        <View className='memory-manage__pagination'>
          {page > 1 && (
            <View className='memory-manage__page-btn' onClick={() => setPage(page - 1)}>
              <Text className='memory-manage__page-btn-text'>上一页</Text>
            </View>
          )}
          <Text className='memory-manage__page-info'>
            第 {page} 页 / 共 {Math.ceil(total / 20)} 页
          </Text>
          {page * 20 < total && (
            <View className='memory-manage__page-btn' onClick={() => setPage(page + 1)}>
              <Text className='memory-manage__page-btn-text'>下一页</Text>
            </View>
          )}
        </View>
      )}
    </View>
  )
}
