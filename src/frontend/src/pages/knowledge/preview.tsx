import { useState, useEffect } from 'react'
import { View, Text, Input, Textarea } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { confirmDocument } from '@/api/knowledge'
import TagInput from '@/components/TagInput'
import './preview.scss'

/** 切片数据 */
interface ChunkItem {
  index: number
  content: string
  char_count: number
}

/** 预览缓存数据 */
interface PreviewData {
  preview_id: string
  title: string
  llm_title: string
  llm_summary: string
  tags: string
  chunks: ChunkItem[]
  chunk_count: number
  total_chars: number
  scope: string
  scopeIds: number[]
  doc_type?: string
  source_url?: string
}

export default function KnowledgePreview() {
  const [previewData, setPreviewData] = useState<PreviewData | null>(null)
  const [title, setTitle] = useState('')
  const [summary, setSummary] = useState('')
  const [tags, setTags] = useState<string[]>([])
  const [editingTitle, setEditingTitle] = useState(false)
  const [editingSummary, setEditingSummary] = useState(false)
  const [editingTags, setEditingTags] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    try {
      const raw = Taro.getStorageSync('previewData')
      if (raw) {
        const data: PreviewData = typeof raw === 'string' ? JSON.parse(raw) : raw
        setPreviewData(data)
        // 如果用户未提供 title，自动使用 llm_title
        const finalTitle = data.title || data.llm_title || ''
        setTitle(finalTitle)
        // 使用 llm_summary 作为初始摘要
        setSummary(data.llm_summary || '')
        setTags(data.tags ? data.tags.split(',').filter(Boolean) : [])
      } else {
        Taro.showToast({ title: '预览数据不存在', icon: 'none' })
      }
    } catch (error) {
      console.error('读取预览数据失败:', error)
      Taro.showToast({ title: '预览数据读取失败', icon: 'none' })
    }
  }, [])

  /** 获取 scope 展示文案 */
  const getScopeLabel = () => {
    if (!previewData) return '-'
    const { scope, scopeIds } = previewData
    if (scope === 'global') return '全部学生'
    if (scope === 'class') return `指定班级 (${scopeIds?.length || 0} 个)`
    if (scope === 'student') return '指定学生'
    return scope || '全部学生'
  }

  /** 取消 */
  const handleCancel = () => {
    Taro.removeStorageSync('previewData')
    Taro.navigateBack()
  }

  /** 确认入库 */
  const handleConfirm = async () => {
    if (!previewData || submitting) return

    setSubmitting(true)
    try {
      const tagsStr = tags.length > 0 ? tags.join(',') : undefined
      const scope = previewData.scope || undefined
      const scopeIds = previewData.scopeIds?.length > 0 ? previewData.scopeIds : undefined

      await confirmDocument(
        previewData.preview_id,
        title.trim() || undefined,
        tagsStr,
        scope,
        scopeIds,
        summary.trim() || undefined,
      )

      Taro.showToast({ title: '入库成功', icon: 'success' })
      // 清理缓存
      Taro.removeStorageSync('previewData')
      // 返回两次（回到知识库列表）
      setTimeout(() => {
        const pages = Taro.getCurrentPages()
        if (pages.length >= 3) {
          Taro.navigateBack({ delta: 2 })
        } else {
          Taro.navigateBack()
        }
      }, 1500)
    } catch (error) {
      console.error('确认入库失败:', error)
    } finally {
      setSubmitting(false)
    }
  }

  if (!previewData) {
    return (
      <View className='knowledge-preview-page'>
        <View className='knowledge-preview-page__loading'>
          <Text className='knowledge-preview-page__loading-text'>加载预览数据中...</Text>
        </View>
      </View>
    )
  }

  return (
    <View className='knowledge-preview-page'>
      {/* 文档信息卡片 */}
      <View className='knowledge-preview-page__info-card'>
        {/* 标题 */}
        <View className='knowledge-preview-page__field'>
          <Text className='knowledge-preview-page__field-label'>标题</Text>
          {editingTitle ? (
            <View className='knowledge-preview-page__field-edit'>
              <Input
                className='knowledge-preview-page__field-input'
                value={title}
                maxlength={200}
                onInput={(e) => setTitle(e.detail.value)}
                onBlur={() => setEditingTitle(false)}
                focus
              />
            </View>
          ) : (
            <View className='knowledge-preview-page__field-value-row' onClick={() => setEditingTitle(true)}>
              <Text className='knowledge-preview-page__field-value'>{title || '未命名文档'}</Text>
              <Text className='knowledge-preview-page__field-edit-icon'>✏️</Text>
            </View>
          )}
          {/* LLM 标题提示 */}
          {previewData.llm_title && previewData.llm_title !== title && (
            <View className='knowledge-preview-page__llm-hint'>
              <Text className='knowledge-preview-page__llm-hint-text'>
                🤖 AI 建议标题：{previewData.llm_title}
              </Text>
              <View
                className='knowledge-preview-page__llm-hint-use'
                onClick={() => setTitle(previewData.llm_title)}
              >
                <Text className='knowledge-preview-page__llm-hint-use-text'>使用</Text>
              </View>
            </View>
          )}
        </View>

        {/* 摘要 */}
        <View className='knowledge-preview-page__field'>
          <Text className='knowledge-preview-page__field-label'>摘要</Text>
          {editingSummary ? (
            <View className='knowledge-preview-page__field-edit'>
              <Textarea
                className='knowledge-preview-page__field-textarea'
                value={summary}
                maxlength={500}
                onInput={(e) => setSummary(e.detail.value)}
                onBlur={() => setEditingSummary(false)}
                autoFocus
              />
            </View>
          ) : (
            <View className='knowledge-preview-page__field-value-row' onClick={() => setEditingSummary(true)}>
              <Text className='knowledge-preview-page__field-value'>
                {summary || '暂无摘要，点击编辑'}
              </Text>
              <Text className='knowledge-preview-page__field-edit-icon'>✏️</Text>
            </View>
          )}
          {/* LLM 摘要提示 */}
          {previewData.llm_summary && previewData.llm_summary !== summary && (
            <View className='knowledge-preview-page__llm-hint'>
              <Text className='knowledge-preview-page__llm-hint-text'>
                🤖 AI 摘要：{previewData.llm_summary.length > 80
                  ? previewData.llm_summary.substring(0, 80) + '...'
                  : previewData.llm_summary}
              </Text>
              <View
                className='knowledge-preview-page__llm-hint-use'
                onClick={() => setSummary(previewData.llm_summary)}
              >
                <Text className='knowledge-preview-page__llm-hint-use-text'>使用</Text>
              </View>
            </View>
          )}
        </View>

        {/* 标签 */}
        <View className='knowledge-preview-page__field'>
          <View className='knowledge-preview-page__field-label-row'>
            <Text className='knowledge-preview-page__field-label'>标签</Text>
            <Text
              className='knowledge-preview-page__field-edit-icon'
              onClick={() => setEditingTags(!editingTags)}
            >
              {editingTags ? '完成' : '✏️'}
            </Text>
          </View>
          {editingTags ? (
            <TagInput tags={tags} onChange={setTags} maxTags={5} />
          ) : (
            <View className='knowledge-preview-page__tags'>
              {tags.length > 0 ? (
                tags.map((tag, idx) => (
                  <View key={idx} className='knowledge-preview-page__tag'>
                    <Text className='knowledge-preview-page__tag-text'>{tag}</Text>
                  </View>
                ))
              ) : (
                <Text className='knowledge-preview-page__field-value--muted'>暂无标签</Text>
              )}
            </View>
          )}
        </View>

        {/* 统计信息 */}
        <View className='knowledge-preview-page__stats'>
          <View className='knowledge-preview-page__stat-item'>
            <Text className='knowledge-preview-page__stat-label'>总字数</Text>
            <Text className='knowledge-preview-page__stat-value'>
              {previewData.total_chars?.toLocaleString() || 0}
            </Text>
          </View>
          <View className='knowledge-preview-page__stat-item'>
            <Text className='knowledge-preview-page__stat-label'>切片数</Text>
            <Text className='knowledge-preview-page__stat-value'>
              {previewData.chunk_count || previewData.chunks?.length || 0}
            </Text>
          </View>
          <View className='knowledge-preview-page__stat-item'>
            <Text className='knowledge-preview-page__stat-label'>生效范围</Text>
            <Text className='knowledge-preview-page__stat-value'>{getScopeLabel()}</Text>
          </View>
        </View>
      </View>

      {/* 切片预览 */}
      <View className='knowledge-preview-page__chunks-section'>
        <Text className='knowledge-preview-page__chunks-title'>切片预览</Text>
        {(previewData.chunks || []).map((chunk, idx) => (
          <View key={idx} className='knowledge-preview-page__chunk-card'>
            <View className='knowledge-preview-page__chunk-header'>
              <Text className='knowledge-preview-page__chunk-index'>
                切片 {chunk.index ?? idx + 1}
              </Text>
              <Text className='knowledge-preview-page__chunk-count'>
                {chunk.char_count} 字
              </Text>
            </View>
            <View className='knowledge-preview-page__chunk-content'>
              <Text className='knowledge-preview-page__chunk-text'>
                {chunk.content.length > 200
                  ? chunk.content.substring(0, 200) + '...'
                  : chunk.content}
              </Text>
            </View>
          </View>
        ))}
      </View>

      {/* 底部操作栏 */}
      <View className='knowledge-preview-page__bottom-bar'>
        <View className='knowledge-preview-page__bottom-btn knowledge-preview-page__bottom-btn--cancel' onClick={handleCancel}>
          <Text className='knowledge-preview-page__bottom-btn-text--cancel'>取消</Text>
        </View>
        <View
          className={`knowledge-preview-page__bottom-btn knowledge-preview-page__bottom-btn--confirm ${submitting ? 'knowledge-preview-page__bottom-btn--disabled' : ''}`}
          onClick={submitting ? undefined : handleConfirm}
        >
          <Text className='knowledge-preview-page__bottom-btn-text--confirm'>
            {submitting ? '入库中...' : '确认入库'}
          </Text>
        </View>
      </View>
    </View>
  )
}
