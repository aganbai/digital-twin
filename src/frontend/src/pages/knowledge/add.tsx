import { useState, useEffect } from 'react'
import { View, Text, Input, Textarea } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { previewDocument, previewUpload, previewUrl, importChat } from '@/api/knowledge'
import { getClasses } from '@/api/class'
import type { ClassInfo } from '@/api/class'
import { usePersonaStore } from '@/store'
import TagInput from '@/components/TagInput'
import './add.scss'

/** Tab 类型 */
type TabType = 'text' | 'file' | 'url' | 'chat'

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
      extension: ['pdf', 'docx', 'txt', 'md'],
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
                    支持 PDF/DOCX/TXT/MD，最大 50MB
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

      {/* 预览/导入按钮 */}
      <View
        className={`knowledge-add-page__submit ${submitting ? 'knowledge-add-page__submit--disabled' : ''}`}
        onClick={submitting ? undefined : handleSubmit}
      >
        <Text className='knowledge-add-page__submit-text'>
          {submitting ? '处理中...' : (activeTab === 'chat' ? '确认导入' : '预览')}
        </Text>
      </View>
    </View>
  )
}
