import { useState, useEffect } from 'react'
import { View, Text, Input, Textarea } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { previewDocument, previewUpload, previewUrl, importChat } from '@/api/knowledge'
import { getClasses } from '@/api/class'
import type { ClassInfo } from '@/api/class'
import { batchUploadDocuments, getBatchTaskStatus } from '@/api/batch-upload'
import type { BatchTask } from '@/api/batch-upload'
import { usePersonaStore } from '@/store'
import TagInput from '@/components/TagInput'
import './add.scss'

/** Tab 类型 */
type TabType = 'text' | 'file' | 'url' | 'chat' | 'batch'

/** Scope 类型 */
type ScopeType = 'global' | 'class' | 'student'

export default function KnowledgeAdd() {
  const [activeTab, setActiveTab] = useState<TabType>('url')
  const { currentPersona } = usePersonaStore()
  const isTeacher = currentPersona?.role === 'teacher'

  // 文本录入
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [tags, setTags] = useState<string[]>([])

  // 文件上传
  const [fileTitle, setFileTitle] = useState('')
  const [fileTags, setFileTags] = useState<string[]>([])
  const [selectedFile, setSelectedFile] = useState<string>('')
  const [fileName, setFileName] = useState('')

  // URL 导入
  const [urlValue, setUrlValue] = useState('')
  const [urlTitle, setUrlTitle] = useState('')
  const [urlTags, setUrlTags] = useState<string[]>([])

  // 聊天记录导入
  const [chatFile, setChatFile] = useState<string>('')
  const [chatFileName, setChatFileName] = useState('')
  const [chatTitle, setChatTitle] = useState('')
  const [chatTags, setChatTags] = useState<string[]>([])

  // 批量上传
  const [batchFiles, setBatchFiles] = useState<{ path: string; name: string; size: number }[]>([])
  const [, setBatchTaskId] = useState('')
  const [batchTaskStatus, setBatchTaskStatus] = useState<BatchTask | null>(null)
  const [batchUploading, setBatchUploading] = useState(false)
  const [pollTimer, setPollTimer] = useState<ReturnType<typeof setInterval> | null>(null)

  // Scope 选择
  const [scope, setScope] = useState<ScopeType>('global')
  const [classes, setClasses] = useState<ClassInfo[]>([])
  const [selectedClassIds, setSelectedClassIds] = useState<number[]>([])

  const [submitting, setSubmitting] = useState(false)

  /** 获取班级列表 */

  useEffect(() => {
    const fetchClasses = async () => {
      try {
        const res = await getClasses()
        setClasses(res.data || [])
      } catch (error) {
        console.error('获取班级列表失败:', error)
      }
    }
    fetchClasses()
  }, [])

  /** 切换班级选中状态 */
  const handleToggleClass = (classId: number) => {
    setSelectedClassIds((prev) => {
      if (prev.includes(classId)) {
        return prev.filter((id) => id !== classId)
      }
      return [...prev, classId]
    })
  }

  /** 选择文件 */
  const handleChooseFile = () => {
    Taro.chooseMessageFile({
      count: 1,
      type: 'file',
      extension: ['pdf', 'docx', 'txt', 'md', 'pptx'],
      success: (res) => {
        if (res.tempFiles && res.tempFiles.length > 0) {
          const file = res.tempFiles[0]
          setSelectedFile(file.path)
          setFileName(file.name)
          if (!fileTitle) {
            const nameWithoutExt = file.name.replace(/\.[^.]+$/, '')
            setFileTitle(nameWithoutExt)
          }
        }
      },
      fail: () => {
        // 用户取消选择
      },
    })
  }

  /** 预览文本文档 */
  const handlePreviewText = async () => {
    const trimmedTitle = title.trim()
    if (!trimmedTitle || trimmedTitle.length > 200) {
      Taro.showToast({ title: '标题为必填项，长度 1-200 字符', icon: 'none' })
      return
    }
    const trimmedContent = content.trim()
    if (!trimmedContent || trimmedContent.length > 100000) {
      Taro.showToast({ title: '内容为必填项，长度 1-100000 字符', icon: 'none' })
      return
    }

    setSubmitting(true)
    try {
      const tagsStr = tags.length > 0 ? tags.join(',') : undefined
      const res = await previewDocument(trimmedTitle, trimmedContent, tagsStr)
      const previewResult = res.data
      // 将预览数据存入缓存供 preview 页面使用
      Taro.setStorageSync('previewData', JSON.stringify({
        preview_id: previewResult.preview_id,
        title: previewResult.title,
        llm_title: previewResult.llm_title || '',
        llm_summary: previewResult.llm_summary || '',
        tags: previewResult.tags || tagsStr || '',
        chunks: previewResult.chunks,
        chunk_count: previewResult.chunk_count,
        total_chars: previewResult.total_chars,
        scope: scope,
        scopeIds: scope === 'class' ? selectedClassIds : [],
        doc_type: previewResult.doc_type || 'text',
        source_url: previewResult.source_url,
      }))
      Taro.navigateTo({ url: '/pages/knowledge/preview' })
    } catch (error) {
      console.error('预览文档失败:', error)
    } finally {
      setSubmitting(false)
    }
  }

  /** 预览文件上传 */
  const handlePreviewFile = async () => {
    if (!selectedFile) {
      Taro.showToast({ title: '请选择文件', icon: 'none' })
      return
    }

    setSubmitting(true)
    try {
      const tagsStr = fileTags.length > 0 ? fileTags.join(',') : undefined
      const previewResult = await previewUpload(
        selectedFile,
        fileTitle.trim() || undefined,
        tagsStr,
      )
      // 将预览数据存入缓存供 preview 页面使用
      Taro.setStorageSync('previewData', JSON.stringify({
        preview_id: previewResult.preview_id,
        title: previewResult.title,
        llm_title: previewResult.llm_title || '',
        llm_summary: previewResult.llm_summary || '',
        tags: previewResult.tags || tagsStr || '',
        chunks: previewResult.chunks,
        chunk_count: previewResult.chunk_count,
        total_chars: previewResult.total_chars,
        scope: scope,
        scopeIds: scope === 'class' ? selectedClassIds : [],
        doc_type: previewResult.doc_type || 'file',
        source_url: previewResult.source_url,
      }))
      Taro.navigateTo({ url: '/pages/knowledge/preview' })
    } catch (error) {
      console.error('预览文件失败:', error)
    } finally {
      setSubmitting(false)
    }
  }

  /** 预览 URL 导入 */
  const handlePreviewUrl = async () => {
    const trimmedUrl = urlValue.trim()
    if (!trimmedUrl) {
      Taro.showToast({ title: '请输入 URL', icon: 'none' })
      return
    }
    if (!trimmedUrl.startsWith('http://') && !trimmedUrl.startsWith('https://')) {
      Taro.showToast({ title: 'URL 必须以 http:// 或 https:// 开头', icon: 'none' })
      return
    }

    setSubmitting(true)
    try {
      const tagsStr = urlTags.length > 0 ? urlTags.join(',') : undefined
      const res = await previewUrl(
        trimmedUrl,
        urlTitle.trim() || undefined,
        tagsStr,
      )
      const previewResult = res.data
      // 将预览数据存入缓存供 preview 页面使用
      Taro.setStorageSync('previewData', JSON.stringify({
        preview_id: previewResult.preview_id,
        title: previewResult.title,
        llm_title: previewResult.llm_title || '',
        llm_summary: previewResult.llm_summary || '',
        tags: previewResult.tags || tagsStr || '',
        chunks: previewResult.chunks,
        chunk_count: previewResult.chunk_count,
        total_chars: previewResult.total_chars,
        scope: scope,
        scopeIds: scope === 'class' ? selectedClassIds : [],
        doc_type: previewResult.doc_type || 'url',
        source_url: previewResult.source_url || trimmedUrl,
      }))
      Taro.navigateTo({ url: '/pages/knowledge/preview' })
    } catch (error) {
      console.error('URL 预览失败:', error)
    } finally {
      setSubmitting(false)
    }
  }

  /** 选择聊天记录 JSON 文件 */
  const handleChooseChatFile = () => {
    Taro.chooseMessageFile({
      count: 1,
      type: 'file',
      extension: ['json'],
      success: (res) => {
        if (res.tempFiles && res.tempFiles.length > 0) {
          const file = res.tempFiles[0]
          setChatFile(file.path)
          setChatFileName(file.name)
          if (!chatTitle) {
            const nameWithoutExt = file.name.replace(/\.[^.]+$/, '')
            setChatTitle(nameWithoutExt)
          }
        }
      },
      fail: () => {
        // 用户取消选择
      },
    })
  }

  /** 选择批量上传文件 */
  const handleChooseBatchFiles = () => {
    Taro.chooseMessageFile({
      count: 20,
      type: 'file',
      extension: ['pdf', 'docx', 'txt', 'md'],
      success: (res) => {
        if (res.tempFiles && res.tempFiles.length > 0) {
          const newFiles = res.tempFiles.map((f) => ({
            path: f.path,
            name: f.name,
            size: f.size,
          }))
          // 合并已选文件，去重并限制20个
          const merged = [...batchFiles]
          for (const nf of newFiles) {
            if (merged.length >= 20) break
            if (!merged.find((mf) => mf.name === nf.name && mf.size === nf.size)) {
              merged.push(nf)
            }
          }
          // 检查总大小 ≤ 100MB
          const totalSize = merged.reduce((sum, f) => sum + f.size, 0)
          if (totalSize > 100 * 1024 * 1024) {
            Taro.showToast({ title: '文件总大小不能超过100MB', icon: 'none' })
            return
          }
          setBatchFiles(merged)
        }
      },
      fail: () => {
        // 用户取消选择
      },
    })
  }

  /** 移除批量上传中的某个文件 */
  const handleRemoveBatchFile = (index: number) => {
    setBatchFiles((prev) => prev.filter((_, i) => i !== index))
  }

  /** 格式化文件大小 */
  const formatFileSize = (bytes: number): string => {
    if (bytes < 1024) return `${bytes}B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)}MB`
  }

  /** 获取状态标签文本 */
  const getStatusLabel = (status: string): string => {
    const map: Record<string, string> = {
      pending: '⏳ 等待处理',
      processing: '🔄 处理中...',
      success: '✅ 全部成功',
      partial_success: '⚠️ 部分成功',
      failed: '❌ 上传失败',
    }
    return map[status] || status
  }

  /** 开始批量上传 */
  const handleBatchUpload = async () => {
    if (batchFiles.length === 0) {
      Taro.showToast({ title: '请先选择文件', icon: 'none' })
      return
    }
    if (!currentPersona?.id) {
      Taro.showToast({ title: '请先选择分身', icon: 'none' })
      return
    }

    setBatchUploading(true)
    setBatchTaskStatus(null)
    try {
      const filePaths = batchFiles.map((f) => f.path)
      const result = await batchUploadDocuments(filePaths, currentPersona.id)
      setBatchTaskId(result.task_id)

      // 指数退避轮询配置
      const INITIAL_INTERVAL = 2000 // 初始间隔2秒
      const MAX_INTERVAL = 10000    // 最大间隔10秒
      const BACKOFF_FACTOR = 1.5    // 指数因子1.5
      const maxPollCount = 60       // 最多轮询60次

      let pollCount = 0
      let currentInterval = INITIAL_INTERVAL

      // 使用递归setTimeout实现指数退避
      const pollTaskStatus = () => {
        pollCount++
        if (pollCount > maxPollCount) {
          setPollTimer(null)
          setBatchUploading(false)
          Taro.showToast({ title: '任务超时，请稍后查看结果', icon: 'none' })
          return
        }

        getBatchTaskStatus(result.task_id)
          .then((statusRes) => {
            const task = statusRes.data
            setBatchTaskStatus(task)

            // 任务完成，停止轮询
            if (task.status === 'success' || task.status === 'partial_success' || task.status === 'failed') {
              setPollTimer(null)
              setBatchUploading(false)
              return
            }

            // 根据任务状态调整轮询间隔
            // pending/processing 状态变化快，使用短间隔
            // 接近完成时逐渐增加间隔
            if (task.status === 'pending' || task.status === 'processing') {
              // 计算下一次轮询间隔（指数退避）
              currentInterval = Math.min(
                currentInterval * BACKOFF_FACTOR,
                MAX_INTERVAL
              )
            }

            // 设置下一次轮询
            const timer = setTimeout(pollTaskStatus, currentInterval)
            setPollTimer(timer as unknown as ReturnType<typeof setInterval>)
          })
          .catch(() => {
            // 轮询出错不中断，继续重试（保持当前间隔）
            const timer = setTimeout(pollTaskStatus, currentInterval)
            setPollTimer(timer as unknown as ReturnType<typeof setInterval>)
          })
      }

      // 启动首次轮询
      const initialTimer = setTimeout(pollTaskStatus, INITIAL_INTERVAL)
      setPollTimer(initialTimer as unknown as ReturnType<typeof setInterval>)
    } catch (error) {
      console.error('批量上传失败:', error)
      setBatchUploading(false)
    }
  }

  // 组件卸载时清理轮询
  useEffect(() => {
    return () => {
      if (pollTimer) {
        // 同时清除interval和timeout
        clearInterval(pollTimer)
        clearTimeout(pollTimer)
      }
    }
  }, [pollTimer])

  /** 导入聊天记录 */
  const handleImportChat = async () => {
    if (!chatFile) {
      Taro.showToast({ title: '请选择 JSON 文件', icon: 'none' })
      return
    }
    if (!currentPersona?.id) {
      Taro.showToast({ title: '请先选择分身', icon: 'none' })
      return
    }

    setSubmitting(true)
    try {
      const tagsStr = chatTags.length > 0 ? JSON.stringify(chatTags) : undefined
      const scopeIdsArr = scope === 'class' ? selectedClassIds : undefined
      const result = await importChat(
        chatFile,
        currentPersona.id,
        chatTitle.trim() || undefined,
        tagsStr,
        scope,
        scopeIdsArr,
      )
      Taro.showToast({
        title: `导入成功，共 ${result.conversation_count} 轮对话`,
        icon: 'success',
        duration: 2000,
      })
      setTimeout(() => {
        Taro.navigateBack()
      }, 2000)
    } catch (error) {
      console.error('导入聊天记录失败:', error)
    } finally {
      setSubmitting(false)
    }
  }

  /** 提交（改为预览） */
  const handleSubmit = () => {
    if (submitting) return
    switch (activeTab) {
      case 'text':
        handlePreviewText()
        break
      case 'file':
        handlePreviewFile()
        break
      case 'url':
        handlePreviewUrl()
        break
      case 'chat':
        handleImportChat()
        break
      case 'batch':
        handleBatchUpload()
        break
    }
  }

  return (
    <View className='knowledge-add-page'>
      {/* Tab 切换 */}
      <View className='knowledge-add-page__tabs'>
        <View
          className={`knowledge-add-page__tab ${activeTab === 'text' ? 'knowledge-add-page__tab--active' : ''}`}
          onClick={() => setActiveTab('text')}
        >
          <Text className={`knowledge-add-page__tab-text ${activeTab === 'text' ? 'knowledge-add-page__tab-text--active' : ''}`}>
            文本录入
          </Text>
        </View>
        <View
          className={`knowledge-add-page__tab ${activeTab === 'file' ? 'knowledge-add-page__tab--active' : ''}`}
          onClick={() => setActiveTab('file')}
        >
          <Text className={`knowledge-add-page__tab-text ${activeTab === 'file' ? 'knowledge-add-page__tab-text--active' : ''}`}>
            文件上传
          </Text>
        </View>
        <View
          className={`knowledge-add-page__tab ${activeTab === 'url' ? 'knowledge-add-page__tab--active' : ''}`}
          onClick={() => setActiveTab('url')}
        >
          <Text className={`knowledge-add-page__tab-text ${activeTab === 'url' ? 'knowledge-add-page__tab-text--active' : ''}`}>
            URL导入
          </Text>
        </View>
        {/* 聊天记录导入（仅教师端显示） */}
        {isTeacher && (
          <View
            className={`knowledge-add-page__tab ${activeTab === 'chat' ? 'knowledge-add-page__tab--active' : ''}`}
            onClick={() => setActiveTab('chat')}
          >
            <Text className={`knowledge-add-page__tab-text ${activeTab === 'chat' ? 'knowledge-add-page__tab-text--active' : ''}`}>
              💬 聊天记录
            </Text>
          </View>
        )}
        {/* 批量上传（仅教师端显示） */}
        {isTeacher && (
          <View
            className={`knowledge-add-page__tab ${activeTab === 'batch' ? 'knowledge-add-page__tab--active' : ''}`}
            onClick={() => setActiveTab('batch')}
          >
            <Text className={`knowledge-add-page__tab-text ${activeTab === 'batch' ? 'knowledge-add-page__tab-text--active' : ''}`}>
              📦 批量上传
            </Text>
          </View>
        )}
      </View>

      {/* Scope 选择 */}
      <View className='knowledge-add-page__section'>
        <Text className='knowledge-add-page__label'>文档范围</Text>
        <View className='knowledge-add-page__scope-select'>
          <View
            className={`knowledge-add-page__scope-option ${scope === 'global' ? 'knowledge-add-page__scope-option--active' : ''}`}
            onClick={() => setScope('global')}
          >
            <Text className={`knowledge-add-page__scope-option-text ${scope === 'global' ? 'knowledge-add-page__scope-option-text--active' : ''}`}>
              全部学生
            </Text>
          </View>
          <View
            className={`knowledge-add-page__scope-option ${scope === 'class' ? 'knowledge-add-page__scope-option--active' : ''}`}
            onClick={() => setScope('class')}
          >
            <Text className={`knowledge-add-page__scope-option-text ${scope === 'class' ? 'knowledge-add-page__scope-option-text--active' : ''}`}>
              指定班级
            </Text>
          </View>
          <View
            className={`knowledge-add-page__scope-option ${scope === 'student' ? 'knowledge-add-page__scope-option--active' : ''}`}
            onClick={() => setScope('student')}
          >
            <Text className={`knowledge-add-page__scope-option-text ${scope === 'student' ? 'knowledge-add-page__scope-option-text--active' : ''}`}>
              指定学生
            </Text>
          </View>
        </View>
        {/* 班级多选列表 */}
        {scope === 'class' && classes.length > 0 && (
          <View className='knowledge-add-page__class-list'>
            {classes.map((cls) => (
              <View
                key={cls.id}
                className={`knowledge-add-page__class-item ${selectedClassIds.includes(cls.id) ? 'knowledge-add-page__class-item--selected' : ''}`}
                onClick={() => handleToggleClass(cls.id)}
              >
                <View className='knowledge-add-page__class-checkbox'>
                  <Text className='knowledge-add-page__class-checkbox-icon'>
                    {selectedClassIds.includes(cls.id) ? '✓' : ''}
                  </Text>
                </View>
                <Text className='knowledge-add-page__class-item-name'>{cls.name}</Text>
                <Text className='knowledge-add-page__class-item-count'>{cls.member_count} 人</Text>
              </View>
            ))}
          </View>
        )}
        {scope === 'class' && classes.length === 0 && (
          <Text className='knowledge-add-page__no-class'>暂无班级，请先创建班级</Text>
        )}
      </View>

      {/* 文本录入 Tab */}
      {activeTab === 'text' && (
        <>
          <View className='knowledge-add-page__section'>
            <Text className='knowledge-add-page__label'>标题</Text>
            <Input
              className='knowledge-add-page__title-input'
              placeholder='请输入文档标题'
              value={title}
              maxlength={200}
              onInput={(e) => setTitle(e.detail.value)}
            />
          </View>
          <View className='knowledge-add-page__section'>
            <Text className='knowledge-add-page__label'>内容</Text>
            <View className='knowledge-add-page__textarea-wrap'>
              <Textarea
                className='knowledge-add-page__textarea'
                placeholder='请输入文档内容...'
                value={content}
                maxlength={100000}
                onInput={(e) => setContent(e.detail.value)}
              />
              <Text className='knowledge-add-page__word-count'>{content.length} 字</Text>
            </View>
          </View>
          <View className='knowledge-add-page__section'>
            <Text className='knowledge-add-page__label'>标签</Text>
            <TagInput tags={tags} onChange={setTags} maxTags={5} />
          </View>
        </>
      )}

      {/* 文件上传 Tab */}
      {activeTab === 'file' && (
        <>
          <View className='knowledge-add-page__section'>
            <Text className='knowledge-add-page__label'>选择文件</Text>
            <View className='knowledge-add-page__file-picker' onClick={handleChooseFile}>
              {selectedFile ? (
                <View className='knowledge-add-page__file-info'>
                  <Text className='knowledge-add-page__file-icon'>📄</Text>
                  <Text className='knowledge-add-page__file-name'>{fileName}</Text>
                </View>
              ) : (
                <View className='knowledge-add-page__file-placeholder'>
                  <Text className='knowledge-add-page__file-icon'>📁</Text>
                  <Text className='knowledge-add-page__file-hint'>点击上传文件</Text>
                  <Text className='knowledge-add-page__file-desc'>
                    支持 PDF/DOCX/TXT/MD/PPTX，最大 50MB
                  </Text>
                </View>
              )}
            </View>
          </View>
          <View className='knowledge-add-page__section'>
            <Text className='knowledge-add-page__label'>标题（可选）</Text>
            <Input
              className='knowledge-add-page__title-input'
              placeholder='不填则使用文件名'
              value={fileTitle}
              maxlength={200}
              onInput={(e) => setFileTitle(e.detail.value)}
            />
          </View>
          <View className='knowledge-add-page__section'>
            <Text className='knowledge-add-page__label'>标签</Text>
            <TagInput tags={fileTags} onChange={setFileTags} maxTags={5} />
          </View>
        </>
      )}

      {/* URL 导入 Tab */}
      {activeTab === 'url' && (
        <>
          <View className='knowledge-add-page__section'>
            <Text className='knowledge-add-page__label'>URL 地址</Text>
            <Input
              className='knowledge-add-page__title-input knowledge-add-page__url-input'
              placeholder='https://example.com/article'
              value={urlValue}
              onInput={(e) => setUrlValue(e.detail.value)}
            />
          </View>
          <View className='knowledge-add-page__section'>
            <Text className='knowledge-add-page__label'>标题（可选）</Text>
            <Input
              className='knowledge-add-page__title-input'
              placeholder='不填则自动提取'
              value={urlTitle}
              maxlength={200}
              onInput={(e) => setUrlTitle(e.detail.value)}
            />
          </View>
          <View className='knowledge-add-page__section'>
            <Text className='knowledge-add-page__label'>标签</Text>
            <TagInput tags={urlTags} onChange={setUrlTags} maxTags={5} />
          </View>
        </>
      )}

      {/* 聊天记录导入 Tab */}
      {activeTab === 'chat' && (
        <>
          <View className='knowledge-add-page__section'>
            <Text className='knowledge-add-page__label'>选择 JSON 文件</Text>
            <View className='knowledge-add-page__file-picker' onClick={handleChooseChatFile}>
              {chatFile ? (
                <View className='knowledge-add-page__file-info'>
                  <Text className='knowledge-add-page__file-icon'>💬</Text>
                  <Text className='knowledge-add-page__file-name'>{chatFileName}</Text>
                </View>
              ) : (
                <View className='knowledge-add-page__file-placeholder'>
                  <Text className='knowledge-add-page__file-icon'>💬</Text>
                  <Text className='knowledge-add-page__file-hint'>点击选择聊天记录文件</Text>
                  <Text className='knowledge-add-page__file-desc'>
                    支持 JSON 格式（OpenAI 风格 / 通用格式），最大 5MB
                  </Text>
                </View>
              )}
            </View>
          </View>
          <View className='knowledge-add-page__section'>
            <Text className='knowledge-add-page__label'>标题（可选）</Text>
            <Input
              className='knowledge-add-page__title-input'
              placeholder='不填则自动生成'
              value={chatTitle}
              maxlength={200}
              onInput={(e) => setChatTitle(e.detail.value)}
            />
          </View>
          <View className='knowledge-add-page__section'>
            <Text className='knowledge-add-page__label'>标签</Text>
            <TagInput tags={chatTags} onChange={setChatTags} maxTags={5} />
          </View>
        </>
      )}

      {/* 批量上传 Tab */}
      {activeTab === 'batch' && (
        <>
          <View className='knowledge-add-page__section'>
            <Text className='knowledge-add-page__label'>选择文件</Text>
            <View className='knowledge-add-page__batch-picker' onClick={handleChooseBatchFiles}>
              <View className='knowledge-add-page__file-placeholder'>
                <Text className='knowledge-add-page__file-icon'>📦</Text>
                <Text className='knowledge-add-page__file-hint'>点击选择文件（可多选）</Text>
                <Text className='knowledge-add-page__file-desc'>
                  支持 PDF/DOCX/TXT/MD，最多20个文件，总大小≤100MB
                </Text>
              </View>
            </View>
          </View>

          {/* 已选文件列表 */}
          {batchFiles.length > 0 && (
            <View className='knowledge-add-page__section'>
              <Text className='knowledge-add-page__label'>
                已选文件（{batchFiles.length}/20）
              </Text>
              <View className='knowledge-add-page__batch-file-list'>
                {batchFiles.map((file, index) => (
                  <View key={`${file.name}-${index}`} className='knowledge-add-page__batch-file-item'>
                    <Text className='knowledge-add-page__batch-file-icon'>📄</Text>
                    <View className='knowledge-add-page__batch-file-info'>
                      <Text className='knowledge-add-page__batch-file-name'>{file.name}</Text>
                      <Text className='knowledge-add-page__batch-file-size'>{formatFileSize(file.size)}</Text>
                    </View>
                    <View
                      className='knowledge-add-page__batch-file-delete'
                      onClick={(e) => { e.stopPropagation(); handleRemoveBatchFile(index) }}
                    >
                      <Text className='knowledge-add-page__batch-file-delete-text'>✕</Text>
                    </View>
                  </View>
                ))}
              </View>
              <Text className='knowledge-add-page__batch-total-size'>
                总大小：{formatFileSize(batchFiles.reduce((sum, f) => sum + f.size, 0))}
              </Text>
            </View>
          )}

          {/* 任务状态卡片 */}
          {batchTaskStatus && (
            <View className='knowledge-add-page__section knowledge-add-page__batch-status-card'>
              <Text className='knowledge-add-page__label'>任务状态</Text>
              <View className='knowledge-add-page__batch-status-row'>
                <Text className='knowledge-add-page__batch-status-label'>
                  {getStatusLabel(batchTaskStatus.status)}
                </Text>
              </View>
              {/* 进度条 */}
              <View className='knowledge-add-page__batch-progress-bar'>
                <View
                  className='knowledge-add-page__batch-progress-fill'
                  style={{
                    width: batchTaskStatus.total_files > 0
                      ? `${((batchTaskStatus.success_files + batchTaskStatus.failed_files) / batchTaskStatus.total_files) * 100}%`
                      : '0%',
                  }}
                />
              </View>
              <View className='knowledge-add-page__batch-status-detail'>
                <Text className='knowledge-add-page__batch-status-text'>
                  总计：{batchTaskStatus.total_files} 个文件
                </Text>
                <Text className='knowledge-add-page__batch-status-text knowledge-add-page__batch-status-text--success'>
                  成功：{batchTaskStatus.success_files}
                </Text>
                {batchTaskStatus.failed_files > 0 && (
                  <Text className='knowledge-add-page__batch-status-text knowledge-add-page__batch-status-text--fail'>
                    失败：{batchTaskStatus.failed_files}
                  </Text>
                )}
              </View>
            </View>
          )}
        </>
      )}

      {/* 预览/导入按钮 */}
      <View
        className={`knowledge-add-page__submit ${submitting || batchUploading ? 'knowledge-add-page__submit--disabled' : ''}`}
        onClick={submitting || batchUploading ? undefined : handleSubmit}
      >
        <Text className='knowledge-add-page__submit-text'>
          {submitting || batchUploading
            ? '处理中...'
            : activeTab === 'chat'
              ? '确认导入'
              : activeTab === 'batch'
                ? '开始上传'
                : '预览'}
        </Text>
      </View>
    </View>
  )
}
