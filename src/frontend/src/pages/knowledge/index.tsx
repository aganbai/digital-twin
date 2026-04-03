import { useState, useCallback } from 'react'
import { View, Text, Input } from '@tarojs/components'
import Taro, { useDidShow, usePullDownRefresh } from '@tarojs/taro'
import {
  getKnowledgeList,
  deleteKnowledgeItem,
  updateKnowledge,
  knowledgeUpload,
  uploadKnowledgeFile,
} from '@/api/knowledge'
import type { KnowledgeItem } from '@/api/knowledge'
import { usePersonaStore } from '@/store'
import Empty from '@/components/Empty'
import { formatTime } from '@/utils/format'
import './index.scss'

/** 类型筛选 */
type TypeFilter = 'all' | 'url' | 'text' | 'file'

/** 获取类型图标 */
const getTypeIcon = (type: string) => {
  switch (type) {
    case 'url': return '🔗'
    case 'text': return '📝'
    case 'file': return '📄'
    default: return '📋'
  }
}

/** 获取类型标签 */
const getTypeLabel = (type: string) => {
  switch (type) {
    case 'url': return '链接'
    case 'text': return '文字'
    case 'file': return '文件'
    default: return '未知'
  }
}

export default function Knowledge() {
  const { currentPersona } = usePersonaStore()
  const [items, setItems] = useState<KnowledgeItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [keyword, setKeyword] = useState('')
  const [typeFilter, setTypeFilter] = useState<TypeFilter>('all')
  const [page, setPage] = useState(1)

  // 统一输入框状态
  const [inputValue, setInputValue] = useState('')
  const [uploading, setUploading] = useState(false)

  // 左滑删除状态
  const [slidingId, setSlidingId] = useState<number | null>(null)

  // 重命名状态
  const [renamingId, setRenamingId] = useState<number | null>(null)
  const [renameValue, setRenameValue] = useState('')

  /** 获取知识库列表 */
  const fetchList = useCallback(async (p = 1, kw = keyword, tf = typeFilter) => {
    setLoading(true)
    try {
      const typeParam = tf === 'all' ? undefined : tf
      const res = await getKnowledgeList(p, 20, kw || undefined, typeParam)
      if (p === 1) {
        setItems(res.data.items || [])
      } else {
        setItems((prev) => [...prev, ...(res.data.items || [])])
      }
      setTotal(res.data.total || 0)
      setPage(p)
    } catch (error) {
      console.error('获取知识库列表失败:', error)
    } finally {
      setLoading(false)
    }
  }, [keyword, typeFilter])

  /** 页面显示时刷新 */
  useDidShow(() => {
    fetchList(1)
  })

  /** 下拉刷新 */
  usePullDownRefresh(async () => {
    await fetchList(1)
    Taro.stopPullDownRefresh()
  })

  /** 搜索 */
  const handleSearch = () => {
    fetchList(1, keyword, typeFilter)
  }

  /** 切换类型筛选 */
  const handleTypeChange = (type: TypeFilter) => {
    setTypeFilter(type)
    fetchList(1, keyword, type)
  }

  /** 加载更多 */
  const handleLoadMore = () => {
    if (items.length < total && !loading) {
      fetchList(page + 1)
    }
  }

  /** 智能识别输入类型并提交 */
  const handleSubmitInput = async () => {
    const trimmed = inputValue.trim()
    if (!trimmed || uploading) return
    if (!currentPersona?.id) {
      Taro.showToast({ title: '请先选择分身', icon: 'none' })
      return
    }

    setUploading(true)
    try {
      // 智能识别：URL 以 http:// 或 https:// 开头
      const isUrl = /^https?:\/\//i.test(trimmed)
      if (isUrl) {
        await knowledgeUpload({
          type: 'url',
          url: trimmed,
          persona_id: currentPersona.id,
        })
        Taro.showToast({ title: '链接添加成功', icon: 'success' })
      } else {
        await knowledgeUpload({
          type: 'text',
          content: trimmed,
          title: trimmed.length > 30 ? trimmed.substring(0, 30) + '...' : trimmed,
          persona_id: currentPersona.id,
        })
        Taro.showToast({ title: '文字添加成功', icon: 'success' })
      }
      setInputValue('')
      fetchList(1)
    } catch (error) {
      console.error('添加知识失败:', error)
    } finally {
      setUploading(false)
    }
  }

  /** 选择文件上传 */
  const handleChooseFile = () => {
    if (uploading || !currentPersona?.id) {
      if (!currentPersona?.id) {
        Taro.showToast({ title: '请先选择分身', icon: 'none' })
      }
      return
    }

    Taro.chooseMessageFile({
      count: 1,
      type: 'file',
      extension: ['pdf', 'docx', 'txt', 'md', 'jpg', 'png'],
      success: async (res) => {
        if (res.tempFiles && res.tempFiles.length > 0) {
          const file = res.tempFiles[0]
          // 检查文件大小 ≤ 20MB
          if (file.size > 20 * 1024 * 1024) {
            Taro.showToast({ title: '文件大小不能超过20MB', icon: 'none' })
            return
          }

          setUploading(true)
          try {
            // 先上传文件获取 URL
            const uploadRes = await uploadKnowledgeFile(file.path)
            // 再调用统一上传接口
            const nameWithoutExt = file.name.replace(/\.[^.]+$/, '')
            await knowledgeUpload({
              type: 'file',
              file_urls: [uploadRes.url],
              title: nameWithoutExt,
              persona_id: currentPersona!.id,
            })
            Taro.showToast({ title: '文件上传成功', icon: 'success' })
            fetchList(1)
          } catch (error) {
            console.error('文件上传失败:', error)
          } finally {
            setUploading(false)
          }
        }
      },
      fail: () => {
        // 用户取消选择
      },
    })
  }

  /** 左滑删除 */
  const handleDelete = (item: KnowledgeItem) => {
    Taro.showModal({
      title: '删除确认',
      content: `确定要删除「${item.title}」吗？`,
      confirmColor: '#FF4D4F',
      success: async (result) => {
        if (result.confirm) {
          try {
            await deleteKnowledgeItem(item.id)
            Taro.showToast({ title: '删除成功', icon: 'success' })
            setSlidingId(null)
            fetchList(1)
          } catch (error) {
            console.error('删除失败:', error)
          }
        }
      },
    })
  }

  /** 开始重命名 */
  const handleStartRename = (item: KnowledgeItem) => {
    setRenamingId(item.id)
    setRenameValue(item.title)
    setSlidingId(null)
  }

  /** 确认重命名 */
  const handleConfirmRename = async () => {
    if (!renamingId || !renameValue.trim()) return
    try {
      await updateKnowledge(renamingId, renameValue.trim())
      Taro.showToast({ title: '重命名成功', icon: 'success' })
      setRenamingId(null)
      setRenameValue('')
      fetchList(1)
    } catch (error) {
      console.error('重命名失败:', error)
    }
  }

  /** 取消重命名 */
  const handleCancelRename = () => {
    setRenamingId(null)
    setRenameValue('')
  }

  /** 点击预览 */
  const handlePreview = (item: KnowledgeItem) => {
    if (slidingId === item.id) return
    Taro.setStorageSync('knowledgePreviewItem', JSON.stringify(item))
    Taro.navigateTo({ url: '/pages/knowledge/preview' })
  }

  /** 左滑操作 */
  const handleSlideChange = (id: number, e: any) => {
    const x = e.detail.x
    // 当向左滑动超过一定距离时，显示操作按钮
    if (x < -60) {
      setSlidingId(id)
    } else {
      if (slidingId === id) {
        setSlidingId(null)
      }
    }
  }

  return (
    <View className='knowledge-page'>
      {/* 顶部标题区 */}
      <View className='knowledge-page__header'>
        <Text className='knowledge-page__title'>我的知识库</Text>
        <Text className='knowledge-page__count'>共 {total} 条知识</Text>
      </View>

      {/* 搜索框 */}
      <View className='knowledge-page__search'>
        <Input
          className='knowledge-page__search-input'
          placeholder='搜索知识标题...'
          placeholderClass='knowledge-page__search-placeholder'
          value={keyword}
          onInput={(e) => setKeyword(e.detail.value)}
          onConfirm={handleSearch}
        />
        <View className='knowledge-page__search-btn' onClick={handleSearch}>
          <Text className='knowledge-page__search-btn-text'>搜索</Text>
        </View>
      </View>

      {/* 类型筛选 Tab */}
      <View className='knowledge-page__type-tabs'>
        {(['all', 'url', 'text', 'file'] as TypeFilter[]).map((type) => (
          <View
            key={type}
            className={`knowledge-page__type-tab ${typeFilter === type ? 'knowledge-page__type-tab--active' : ''}`}
            onClick={() => handleTypeChange(type)}
          >
            <Text className={`knowledge-page__type-tab-text ${typeFilter === type ? 'knowledge-page__type-tab-text--active' : ''}`}>
              {type === 'all' ? '全部' : type === 'url' ? '🔗 链接' : type === 'text' ? '📝 文字' : '📄 文件'}
            </Text>
          </View>
        ))}
      </View>

      {/* 知识库列表 */}
      <View className='knowledge-page__content'>
        {loading && items.length === 0 ? (
          <View className='knowledge-page__loading'>
            <Text className='knowledge-page__loading-text'>加载中...</Text>
          </View>
        ) : items.length > 0 ? (
          <View className='knowledge-page__list'>
            {items.map((item) => (
              <View key={item.id} className='knowledge-page__item-wrapper'>
                <View
                  className={`knowledge-page__item-container ${slidingId === item.id ? 'knowledge-page__item-container--slid' : ''}`}
                >
                  {/* 主内容区域 */}
                  <View
                    className='knowledge-page__item'
                    onClick={() => handlePreview(item)}
                    onLongPress={() => handleStartRename(item)}
                  >
                    {/* 重命名模式 */}
                    {renamingId === item.id ? (
                      <View className='knowledge-page__rename-row'>
                        <Input
                          className='knowledge-page__rename-input'
                          value={renameValue}
                          maxlength={200}
                          focus
                          onInput={(e) => setRenameValue(e.detail.value)}
                          onConfirm={handleConfirmRename}
                        />
                        <View className='knowledge-page__rename-actions'>
                          <View className='knowledge-page__rename-btn knowledge-page__rename-btn--confirm' onClick={handleConfirmRename}>
                            <Text className='knowledge-page__rename-btn-text'>✓</Text>
                          </View>
                          <View className='knowledge-page__rename-btn knowledge-page__rename-btn--cancel' onClick={handleCancelRename}>
                            <Text className='knowledge-page__rename-btn-text'>✕</Text>
                          </View>
                        </View>
                      </View>
                    ) : (
                      <>
                        <View className='knowledge-page__item-header'>
                          <Text className='knowledge-page__item-type-icon'>{getTypeIcon(item.type)}</Text>
                          <Text className='knowledge-page__item-title'>{item.title}</Text>
                          <View className='knowledge-page__item-type-badge'>
                            <Text className='knowledge-page__item-type-badge-text'>{getTypeLabel(item.type)}</Text>
                          </View>
                        </View>
                        {item.url && (
                          <Text className='knowledge-page__item-url'>{item.url}</Text>
                        )}
                        <Text className='knowledge-page__item-time'>{formatTime(item.created_at)}</Text>
                      </>
                    )}
                  </View>

                  {/* 左滑操作按钮 */}
                  <View className='knowledge-page__item-actions'>
                    <View
                      className='knowledge-page__action-btn knowledge-page__action-btn--rename'
                      onClick={() => handleStartRename(item)}
                    >
                      <Text className='knowledge-page__action-btn-text'>重命名</Text>
                    </View>
                    <View
                      className='knowledge-page__action-btn knowledge-page__action-btn--delete'
                      onClick={() => handleDelete(item)}
                    >
                      <Text className='knowledge-page__action-btn-text'>删除</Text>
                    </View>
                  </View>
                </View>

                {/* 点击遮罩关闭滑动 */}
                {slidingId === item.id && (
                  <View className='knowledge-page__item-mask' onClick={() => setSlidingId(null)} />
                )}
              </View>
            ))}

            {/* 加载更多 */}
            {items.length < total && (
              <View className='knowledge-page__loadmore' onClick={handleLoadMore}>
                <Text className='knowledge-page__loadmore-text'>
                  {loading ? '加载中...' : '加载更多'}
                </Text>
              </View>
            )}
          </View>
        ) : (
          <Empty text='还没有知识，在下方输入框添加吧' />
        )}
      </View>

      {/* 底部统一输入框 */}
      <View className='knowledge-page__input-bar'>
        <View className='knowledge-page__input-row'>
          <View className='knowledge-page__file-btn' onClick={handleChooseFile}>
            <Text className='knowledge-page__file-btn-icon'>📎</Text>
          </View>
          <Input
            className='knowledge-page__unified-input'
            placeholder='输入链接或文字内容...'
            placeholderClass='knowledge-page__input-placeholder'
            value={inputValue}
            onInput={(e) => setInputValue(e.detail.value)}
            onConfirm={handleSubmitInput}
            disabled={uploading}
          />
          <View
            className={`knowledge-page__send-btn ${(!inputValue.trim() || uploading) ? 'knowledge-page__send-btn--disabled' : ''}`}
            onClick={handleSubmitInput}
          >
            <Text className='knowledge-page__send-btn-text'>
              {uploading ? '...' : '发送'}
            </Text>
          </View>
        </View>
        {/* 输入提示 */}
        <View className='knowledge-page__input-hint'>
          <Text className='knowledge-page__input-hint-text'>
            {inputValue.trim() && /^https?:\/\//i.test(inputValue.trim())
              ? '🔗 识别为链接，将自动抓取网页内容'
              : inputValue.trim()
                ? '📝 识别为文字内容'
                : '输入链接自动识别，或点击 📎 上传文件'}
          </Text>
        </View>
      </View>
    </View>
  )
}
